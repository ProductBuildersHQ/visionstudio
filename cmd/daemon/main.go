package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/grokify/mogo/os/osutil"

	specconfig "github.com/ProductBuildersHQ/visionspec/pkg/config"
	"github.com/ProductBuildersHQ/visionspec/pkg/lint"
	"github.com/ProductBuildersHQ/visionspec/pkg/profiles"
	"github.com/ProductBuildersHQ/visionspec/pkg/types"
	"github.com/ProductBuildersHQ/visionspec/pkg/workflow"
	"github.com/ProductBuildersHQ/visionspec/pkg/workflow/specworkflow"
	"github.com/ProductBuildersHQ/visionstudio/pkg/api"
	"github.com/ProductBuildersHQ/visionstudio/pkg/config"
)

func main() {
	port := flag.Int("port", 8765, "Port to listen on")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	srv := NewServer(logger)

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("Starting VisionStudio daemon", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Shutdown error", "error", err)
	}
}

// Server handles HTTP requests
type Server struct {
	logger *slog.Logger

	// File watching
	watchers   map[string]*fsnotify.Watcher // project name -> watcher
	watchersMu sync.RWMutex

	// SSE clients per project
	sseClients   map[string]map[chan api.FileEvent]struct{} // project -> set of channels
	sseClientsMu sync.RWMutex
}

// NewServer creates a new server
func NewServer(logger *slog.Logger) *Server {
	return &Server{
		logger:     logger,
		watchers:   make(map[string]*fsnotify.Watcher),
		sseClients: make(map[string]map[chan api.FileEvent]struct{}),
	}
}

// Router returns the HTTP router
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)

		// Organization
		r.Get("/organization", s.handleGetOrganization)
		r.Post("/organization", s.handleCreateOrganization)
		r.Put("/organization", s.handleUpdateOrganization)
		r.Delete("/organization", s.handleDeleteOrganization)
		r.Get("/organization/v2moms", s.handleListOrganizationV2MOMs)
		r.Get("/organization/v2moms/{v2momId}", s.handleGetOrganizationV2MOM)
		r.Get("/organization/cascade", s.handleGetOrganizationCascade)

		// Profiles (requirements methodologies)
		r.Get("/profiles", s.handleListProfiles)

		// Methodologies
		r.Get("/methodologies/requirements", s.handleListRequirementsMethodologies)
		r.Get("/methodologies/implementation", s.handleListImplementationMethodologies)
		r.Get("/projects/{project}/methodology", s.handleGetProjectMethodology)
		r.Put("/projects/{project}/methodology", s.handleUpdateProjectMethodology)

		// Samples
		r.Get("/samples", s.handleListSamples)
		r.Get("/samples/{sampleId}", s.handleGetSample)

		// Projects
		r.Get("/projects", s.handleListProjects)
		r.Post("/projects", s.handleAddProject)
		r.Get("/projects/{project}", s.handleGetProject)
		r.Delete("/projects/{project}", s.handleRemoveProject)
		r.Get("/projects/{project}/lint", s.handleLintProject)

		// Specs
		r.Get("/projects/{project}/specs/{specType}", s.handleGetSpec)
		r.Put("/projects/{project}/specs/{specType}", s.handleSaveSpec)
		r.Post("/projects/{project}/specs/{specType}/evaluate", s.handleEvaluateSpec)

		// Workflow
		r.Get("/projects/{project}/workflow", s.handleGetWorkflow)
		r.Get("/projects/{project}/workflow/status", s.handleGetWorkflowStatus)

		// Real-time events (SSE)
		r.Get("/projects/{project}/events", s.handleSSE)

		// Chat
		r.Post("/chat", s.handleChat)

		// V2MOM
		r.Get("/projects/{project}/v2moms", s.handleListV2MOMs)
		r.Get("/projects/{project}/v2moms/cascade", s.handleGetV2MOMCascade)
		r.Get("/projects/{project}/v2moms/{v2momId}", s.handleGetV2MOM)

		// Capabilities
		r.Get("/projects/{project}/capabilities", s.handleListCapabilities)
		r.Get("/projects/{project}/capabilities/{capabilityId}", s.handleGetCapability)

		// Roadmap
		r.Get("/projects/{project}/roadmap", s.handleGetRoadmap)

		// Maturity Model
		r.Get("/projects/{project}/maturity/models", s.handleListMaturityModels)
		r.Get("/projects/{project}/maturity/models/{modelId}", s.handleGetMaturityModel)
		r.Get("/projects/{project}/maturity/dashboard", s.handleMaturityDashboard)

		// AIDLC (AWS AI DLC Workflow)
		r.Get("/projects/{project}/aidlc/state", s.handleGetAIDLCState)
		r.Get("/projects/{project}/aidlc/workflow", s.handleGetAIDLCWorkflow)
		r.Get("/projects/{project}/aidlc/documents", s.handleListAIDLCDocuments)
		r.Get("/projects/{project}/aidlc/documents/{docId}", s.handleGetAIDLCDocument)
		r.Post("/projects/{project}/aidlc/documents/create", s.handleCreateAIDLCDocument)
		r.Get("/projects/{project}/aidlc/sync/diff", s.handleGetAIDLCSyncDiff)
		r.Post("/projects/{project}/aidlc/sync", s.handleAIDLCSync)
		r.Get("/projects/{project}/aidlc/phase/requirements", s.handleGetAIDLCPhaseRequirements)
		r.Post("/projects/{project}/aidlc/phase/transition", s.handleAIDLCPhaseTransition)
		r.Get("/projects/{project}/aidlc/templates", s.handleListAIDLCTemplates)
		r.Get("/projects/{project}/aidlc/templates/{docType}", s.handleGetAIDLCTemplate)
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// profileLoader is the visionspec profile loader (cached)
var profileLoader = profiles.DefaultLoader()

// loadAPIProfile loads a profile from visionspec and converts to API type
func loadAPIProfile(name string) api.Profile {
	p, err := profileLoader.Load(name)
	if err != nil {
		// Fall back to a minimal profile on error
		return api.Profile{
			Name:        name,
			Description: "Unknown profile",
			Workflow:    []string{},
		}
	}

	// Get required specs as workflow
	workflow := p.RequiredSpecs()

	return api.Profile{
		Name:        p.Name,
		Description: p.Description,
		Workflow:    workflow,
	}
}

// availableProfiles returns the list of available workflow profiles from visionspec
func availableProfiles() []api.Profile {
	result := make([]api.Profile, 0, len(profiles.DefaultProfileNames))
	for _, name := range profiles.DefaultProfileNames {
		result = append(result, loadAPIProfile(name))
	}
	return result
}

// getProfileByName returns a profile by name
func getProfileByName(name string) api.Profile {
	if profiles.IsDefaultProfile(name) {
		return loadAPIProfile(name)
	}
	// Default to startup if not found
	return loadAPIProfile("startup")
}

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, api.ListProfilesResponse{Profiles: availableProfiles()})
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadProjects()
	if err != nil {
		s.logger.Error("Failed to load projects config", "error", err)
		s.writeJSON(w, http.StatusInternalServerError, api.ListProjectsResponse{})
		return
	}

	projects := make([]api.Project, 0, len(cfg.Projects))
	for _, p := range cfg.Projects {
		profile := getProfileByName(p.Profile)
		gitRemote := getGitRemote(p.Path)

		// Build spec list from profile workflow
		specs := buildSpecsFromWorkflow(profile.Workflow, p.Path)

		// Determine methodologies
		reqMethodology := p.RequirementsMethodology
		if reqMethodology == "" {
			reqMethodology = p.Profile
		}
		implMethodology := p.ImplementationMethodology
		if implMethodology == "" {
			implMethodology = "none"
		}

		projects = append(projects, api.Project{
			Name:                      p.Name,
			Path:                      p.Path,
			Profile:                   profile,
			RequirementsMethodology:   reqMethodology,
			ImplementationMethodology: implMethodology,
			GitRemote:                 gitRemote,
			Specs:                     specs,
		})
	}

	s.writeJSON(w, http.StatusOK, api.ListProjectsResponse{Projects: projects})
}

func (s *Server) handleAddProject(w http.ResponseWriter, r *http.Request) {
	var req api.AddProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.AddProjectResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Name == "" {
		s.writeJSON(w, http.StatusBadRequest, api.AddProjectResponse{
			Error: "Name is required",
		})
		return
	}
	if req.Path == "" {
		s.writeJSON(w, http.StatusBadRequest, api.AddProjectResponse{
			Error: "Path is required",
		})
		return
	}
	if req.Profile == "" {
		req.Profile = "startup" // Default to lightweight profile
	}

	absPath, _ := filepath.Abs(req.Path)

	// Initialize project directory if requested
	if req.Initialize {
		if err := s.initializeProject(absPath, req.Name, req.Profile); err != nil {
			s.writeJSON(w, http.StatusBadRequest, api.AddProjectResponse{
				Error: fmt.Sprintf("failed to initialize project: %v", err),
			})
			return
		}
	}

	// Add project to config
	if err := config.AddProject(req.Name, absPath, req.Profile); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.AddProjectResponse{
			Error: err.Error(),
		})
		return
	}

	// Build response
	profile := getProfileByName(req.Profile)
	gitRemote := getGitRemote(absPath)
	specs := buildSpecsFromWorkflow(profile.Workflow, absPath)

	project := api.Project{
		Name:      req.Name,
		Path:      absPath,
		Profile:   profile,
		GitRemote: gitRemote,
		Specs:     specs,
	}

	s.logger.Info("Added project", "name", req.Name, "path", absPath, "profile", req.Profile, "initialized", req.Initialize)
	s.writeJSON(w, http.StatusCreated, api.AddProjectResponse{Project: project})
}

// initializeProject creates the project directory structure and scaffolds specs
func (s *Server) initializeProject(projectPath, projectName, profileName string) error {
	// Check if directory exists
	info, err := os.Stat(projectPath)
	if err == nil {
		// Directory exists - check if it's empty or already a visionspec project
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", projectPath)
		}

		// Check if visionspec.yaml already exists
		configPath := filepath.Join(projectPath, specconfig.ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("project already initialized (visionspec.yaml exists)")
		}

		// Check if directory is empty (allow .git, .gitignore, README.md)
		entries, err := os.ReadDir(projectPath)
		if err != nil {
			return fmt.Errorf("reading directory: %w", err)
		}

		allowedFiles := map[string]bool{
			".git": true, ".gitignore": true, "README.md": true, ".DS_Store": true,
		}
		for _, entry := range entries {
			if !allowedFiles[entry.Name()] {
				return fmt.Errorf("directory not empty: contains %s (use an empty directory or remove existing files)", entry.Name())
			}
		}
	} else if os.IsNotExist(err) {
		// Create directory
		if err := os.MkdirAll(projectPath, 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
	} else {
		return fmt.Errorf("checking path: %w", err)
	}

	// Create subdirectories
	subdirs := []string{
		specconfig.SourceDir,
		specconfig.GTMDir,
		specconfig.TechnicalDir,
		specconfig.EvalDir,
	}
	for _, subdir := range subdirs {
		path := filepath.Join(projectPath, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("creating %s directory: %w", subdir, err)
		}
	}

	// Load profile to get templates and spec config
	profile, err := profileLoader.Load(profileName)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	// Create visionspec.yaml
	visionspecConfig := &types.Project{
		Name:     projectName,
		Path:     projectPath,
		Workflow: profileName,
	}
	if err := specconfig.Save(visionspecConfig); err != nil {
		return fmt.Errorf("creating visionspec.yaml: %w", err)
	}

	// Scaffold spec files from templates
	templateLoader := profile.GetTemplateLoader()
	if templateLoader == nil {
		s.logger.Warn("No template loader for profile", "profile", profileName)
		return nil
	}

	specConfig := profile.GetSpecConfig()
	if specConfig == nil {
		return nil
	}

	for specName, req := range specConfig.Specs {
		if req == nil || !req.Required {
			continue
		}

		// Load template
		specType := types.SpecType(specName)
		tmpl, err := templateLoader.Load(specType)
		if err != nil {
			s.logger.Debug("No template for spec type", "type", specName, "error", err)
			continue
		}

		// Determine target directory based on category
		var targetDir string
		switch req.Category {
		case types.CategorySource:
			targetDir = specconfig.SourceDir
		case types.CategoryGTM:
			targetDir = specconfig.GTMDir
		case types.CategoryTechnical:
			targetDir = specconfig.TechnicalDir
		default:
			targetDir = specconfig.SourceDir
		}

		// Write template to file
		specPath := filepath.Join(projectPath, targetDir, specType.Filename())
		if err := os.WriteFile(specPath, []byte(tmpl.Content), 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", specPath, err)
		}

		s.logger.Debug("Created spec file", "path", specPath)
	}

	s.logger.Info("Initialized project", "path", projectPath, "profile", profileName)
	return nil
}

func (s *Server) handleRemoveProject(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	if err := config.RemoveProject(projectName); err != nil {
		s.writeJSON(w, http.StatusNotFound, api.RemoveProjectResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	s.logger.Info("Removed project", "name", projectName)
	s.writeJSON(w, http.StatusOK, api.RemoveProjectResponse{Success: true})
}

func (s *Server) handleLintProject(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	// Get project from config
	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.LintProjectResponse{
			Error: err.Error(),
		})
		return
	}

	// Load the profile to get spec config
	profile, err := profileLoader.Load(tracked.Profile)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.LintProjectResponse{
			Error: fmt.Sprintf("failed to load profile: %v", err),
		})
		return
	}

	// Create linter with profile's spec config
	linter := lint.NewWithConfig(tracked.Path, profile.GetSpecConfig())
	result, err := linter.LintProject(projectName, tracked.Path)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.LintProjectResponse{
			Error: fmt.Sprintf("lint failed: %v", err),
		})
		return
	}

	// Convert to API types
	findings := make([]api.LintFinding, len(result.Findings))
	for i, f := range result.Findings {
		findings[i] = api.LintFinding{
			Path:     f.Path,
			Rule:     f.Rule,
			Message:  f.Message,
			Severity: string(f.Severity),
		}
	}

	s.writeJSON(w, http.StatusOK, api.LintProjectResponse{
		Result: api.LintResult{
			Project:  result.Project,
			Findings: findings,
			Errors:   result.Errors,
			Warnings: result.Warnings,
			Passed:   result.Passed,
		},
	})
}

// buildSpecsFromWorkflow creates spec entries based on workflow steps
func buildSpecsFromWorkflow(workflow []string, projectPath string) []api.Spec {
	specMeta := map[string]struct {
		name string
		path string
	}{
		"mrd":              {name: "Market Requirements", path: "source/mrd.md"},
		"opportunity-spec": {name: "Opportunity Spec", path: "source/opportunity-spec.md"},
		"press":            {name: "Press Release", path: "gtm/press.md"},
		"faq":              {name: "FAQ", path: "gtm/faq.md"},
		"narrative-1p":     {name: "1-Pager", path: "gtm/narrative-1p.md"},
		"narrative-6p":     {name: "6-Pager", path: "gtm/narrative-6p.md"},
		"prd":              {name: "Product Requirements", path: "source/prd.md"},
		"uxd":              {name: "User Experience", path: "source/uxd.md"},
		"trd":              {name: "Technical Design", path: "technical/trd.md"},
		"tpd":              {name: "Test Plan", path: "technical/tpd.md"},
		"ird":              {name: "Infrastructure", path: "technical/ird.md"},
		"spec":             {name: "Execution Spec", path: "spec.md"},
		"current-truth":    {name: "Current Truth", path: "current-truth.md"},
	}

	specs := make([]api.Spec, 0, len(workflow))
	for _, step := range workflow {
		meta, ok := specMeta[step]
		if !ok {
			meta = struct {
				name string
				path string
			}{name: step, path: fmt.Sprintf("specs/%s.md", step)}
		}

		// Check if spec file exists to determine status
		status := api.SpecStatusNotStarted
		var evalResult *api.EvalResult
		specPath := filepath.Join(projectPath, meta.path)
		if _, err := os.Stat(specPath); err == nil {
			status = api.SpecStatusDraft

			// Check for eval result
			evalPath := filepath.Join(projectPath, "eval", step+".json")
			if evalData, err := os.ReadFile(evalPath); err == nil {
				var result api.EvalResult
				if json.Unmarshal(evalData, &result) == nil {
					evalResult = &result
					// Update status based on eval decision
					if result.Decision == api.EvalDecisionPass {
						status = api.SpecStatusApproved
					} else {
						status = api.SpecStatusEvaluated
					}
				}
			}
		}

		specs = append(specs, api.Spec{
			Type:       step,
			Name:       meta.name,
			Path:       meta.path,
			Status:     status,
			EvalResult: evalResult,
		})
	}

	return specs
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	s.logger.Debug("Getting project", "name", projectName)

	// Get absolute path for project directory
	// The daemon binary is in bin/, so go up one level to repo root
	execPath, err := os.Executable()
	if err != nil {
		execPath = ""
	}
	repoRoot := filepath.Dir(filepath.Dir(execPath))
	projectPath := filepath.Join(repoRoot, "docs/specs", projectName)

	// TODO: Load actual project from filesystem
	s.writeJSON(w, http.StatusOK, api.GetProjectResponse{
		Project: api.Project{
			Name: projectName,
			Path: projectPath,
		},
	})
}

func (s *Server) handleGetSpec(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	specType := chi.URLParam(r, "specType")
	s.logger.Debug("Getting spec", "project", projectName, "type", specType)

	// Validate specType for path traversal
	if err := osutil.ValidateNoTraversal(specType); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.GetSpecResponse{})
		return
	}

	// Get project from config
	project, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetSpecResponse{})
		return
	}

	// Map spec type to file path
	specMeta := map[string]struct {
		name string
		path string
	}{
		"mrd":              {name: "Market Requirements", path: "source/mrd.md"},
		"opportunity-spec": {name: "Opportunity Spec", path: "source/opportunity-spec.md"},
		"press":            {name: "Press Release", path: "gtm/press.md"},
		"faq":              {name: "FAQ", path: "gtm/faq.md"},
		"narrative-1p":     {name: "1-Pager", path: "gtm/narrative-1p.md"},
		"narrative-6p":     {name: "6-Pager", path: "gtm/narrative-6p.md"},
		"prd":              {name: "Product Requirements", path: "source/prd.md"},
		"uxd":              {name: "User Experience", path: "source/uxd.md"},
		"trd":              {name: "Technical Design", path: "technical/trd.md"},
		"tpd":              {name: "Test Plan", path: "technical/tpd.md"},
		"ird":              {name: "Infrastructure", path: "technical/ird.md"},
		"spec":             {name: "Execution Spec", path: "spec.md"},
	}

	meta, ok := specMeta[specType]
	if !ok {
		meta = struct {
			name string
			path string
		}{name: specType, path: fmt.Sprintf("specs/%s.md", specType)}
	}

	// Read spec content from filesystem (specType validated above)
	specPath := filepath.Join(project.Path, meta.path) //nolint:gosec // specType validated by osutil.ValidateNoTraversal
	content := ""
	status := api.SpecStatusNotStarted

	if data, err := os.ReadFile(specPath); err == nil { //nolint:gosec // specType validated above
		content = string(data)
		status = api.SpecStatusDraft
	}

	// Check for eval result (specType validated above)
	var evalResult *api.EvalResult
	evalPath := filepath.Join(project.Path, "eval", specType+".json") //nolint:gosec // specType validated above
	if evalData, err := os.ReadFile(evalPath); err == nil {           //nolint:gosec // specType validated above
		var result api.EvalResult
		if json.Unmarshal(evalData, &result) == nil {
			evalResult = &result
			if result.Decision == api.EvalDecisionPass {
				status = api.SpecStatusApproved
			} else {
				status = api.SpecStatusEvaluated
			}
		}
	}

	s.writeJSON(w, http.StatusOK, api.GetSpecResponse{
		Spec: api.Spec{
			Type:       specType,
			Name:       meta.name,
			Path:       meta.path,
			Status:     status,
			Content:    content,
			EvalResult: evalResult,
		},
	})
}

func (s *Server) handleSaveSpec(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	specType := chi.URLParam(r, "specType")

	// Validate specType for path traversal
	if err := osutil.ValidateNoTraversal(specType); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.SaveSpecResponse{
			Success: false,
			Error:   "Invalid spec type",
		})
		return
	}

	var req api.SaveSpecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.SaveSpecResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	s.logger.Debug("Saving spec", "project", projectName, "type", specType, "contentLen", len(req.Content))

	// Get project from config
	project, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.SaveSpecResponse{
			Success: false,
			Error:   fmt.Sprintf("Project not found: %s", projectName),
		})
		return
	}

	// Map spec type to file path
	specMeta := map[string]string{
		"mrd":              "source/mrd.md",
		"opportunity-spec": "source/opportunity-spec.md",
		"press":            "gtm/press.md",
		"faq":              "gtm/faq.md",
		"narrative-1p":     "gtm/narrative-1p.md",
		"narrative-6p":     "gtm/narrative-6p.md",
		"prd":              "source/prd.md",
		"uxd":              "source/uxd.md",
		"trd":              "technical/trd.md",
		"tpd":              "technical/tpd.md",
		"ird":              "technical/ird.md",
		"spec":             "spec.md",
	}

	relPath, ok := specMeta[specType]
	if !ok {
		relPath = fmt.Sprintf("specs/%s.md", specType)
	}

	// specType validated above, relPath is either from hardcoded map or validated specType
	specPath := filepath.Join(project.Path, relPath) //nolint:gosec // specType validated by osutil.ValidateNoTraversal

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(specPath), 0755); err != nil { //nolint:gosec // specType validated above
		s.writeJSON(w, http.StatusInternalServerError, api.SaveSpecResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		})
		return
	}

	// Write spec content
	if err := os.WriteFile(specPath, []byte(req.Content), 0o600); err != nil { //nolint:gosec // specType validated above
		s.writeJSON(w, http.StatusInternalServerError, api.SaveSpecResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write file: %v", err),
		})
		return
	}

	s.logger.Info("Saved spec", "project", projectName, "type", specType, "path", specPath)
	s.writeJSON(w, http.StatusOK, api.SaveSpecResponse{Success: true})
}

func (s *Server) handleEvaluateSpec(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	specType := chi.URLParam(r, "specType")
	s.logger.Debug("Evaluating spec", "project", projectName, "type", specType)

	// TODO: Integrate with visionspec evaluation
	s.writeJSON(w, http.StatusOK, api.EvaluateSpecResponse{
		Result: api.EvalResult{
			Score:    7.5,
			Decision: api.EvalDecisionPass,
			Findings: []api.Finding{
				{Category: "clarity", Severity: "medium", Message: "Consider adding more specific user personas"},
			},
		},
	})
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req api.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.ChatResponse{
			Error: "Invalid request body",
		})
		return
	}

	s.logger.Debug("Chat request", "message", req.Message)

	// TODO: Integrate with omniagent LLM
	s.writeJSON(w, http.StatusOK, api.ChatResponse{
		Response: fmt.Sprintf("I received your message: %q. LLM integration coming soon!", req.Message),
	})
}

// Maturity Model handlers

func (s *Server) handleListMaturityModels(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	s.logger.Debug("Listing maturity models", "project", projectName)

	// Get project from config
	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]any{
			"models": []any{},
			"error":  err.Error(),
		})
		return
	}

	// Look for maturity model files in .visionspec/maturity/ or maturity/
	models := []api.MaturityModelSummary{}

	maturityDirs := []string{
		filepath.Join(tracked.Path, ".visionspec", "maturity"),
		filepath.Join(tracked.Path, "maturity"),
	}

	for _, dir := range maturityDirs {
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
					continue
				}

				modelPath := filepath.Join(dir, entry.Name())
				if model, err := s.loadMaturityModelSummary(modelPath); err == nil {
					models = append(models, model)
				}
			}
			break // Only use the first directory that exists
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"models": models,
	})
}

func (s *Server) loadMaturityModelSummary(path string) (api.MaturityModelSummary, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G703: Path from tracked project config
	if err != nil {
		return api.MaturityModelSummary{}, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return api.MaturityModelSummary{}, err
	}

	// Extract summary fields
	id := filepath.Base(path)
	id = strings.TrimSuffix(id, ".json")

	name, _ := raw["name"].(string)
	if name == "" {
		name = id
	}

	description, _ := raw["description"].(string)

	var dimensionCount int
	if dims, ok := raw["dimensions"].([]any); ok {
		dimensionCount = len(dims)
	}

	var goalCount int
	if goals, ok := raw["goals"].([]any); ok {
		goalCount = len(goals)
	}

	var overallScore float64
	if score, ok := raw["overallScore"].(float64); ok {
		overallScore = score
	}

	info, _ := os.Stat(path) //nolint:gosec // G703: Path from tracked project config
	lastUpdated := time.Now()
	if info != nil {
		lastUpdated = info.ModTime()
	}

	return api.MaturityModelSummary{
		ID:             id,
		Name:           name,
		Description:    description,
		DimensionCount: dimensionCount,
		GoalCount:      goalCount,
		OverallScore:   overallScore,
		LastUpdated:    lastUpdated.UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) handleGetMaturityModel(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	modelID := chi.URLParam(r, "modelId")
	s.logger.Debug("Getting maturity model", "project", projectName, "model", modelID)

	// Get project from config
	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Look for the model file
	modelPaths := []string{
		filepath.Join(tracked.Path, ".visionspec", "maturity", modelID+".json"),
		filepath.Join(tracked.Path, "maturity", modelID+".json"),
	}

	var modelData []byte
	for _, path := range modelPaths {
		if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G703: Path from tracked project config
			modelData = data
			break
		}
	}

	if modelData == nil {
		s.writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "Model not found: " + modelID,
		})
		return
	}

	// Parse and return the model
	var model map[string]any
	if err := json.Unmarshal(modelData, &model); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to parse model: " + err.Error(),
		})
		return
	}

	// Ensure ID is set
	if model["id"] == nil {
		model["id"] = modelID
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"model": model,
	})
}

func (s *Server) handleMaturityDashboard(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	theme := r.URL.Query().Get("theme")
	if theme == "" {
		theme = "dark"
	}
	modelID := r.URL.Query().Get("model")
	if modelID == "" {
		modelID = "default"
	}

	s.logger.Debug("Getting maturity dashboard", "project", projectName, "theme", theme, "model", modelID)

	// Get project from config
	tracked, err := config.GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Look for model file to get data
	modelPaths := []string{
		filepath.Join(tracked.Path, ".visionspec", "maturity", modelID+".json"),
		filepath.Join(tracked.Path, "maturity", modelID+".json"),
		filepath.Join(tracked.Path, ".visionspec", "maturity", "model.json"),
		filepath.Join(tracked.Path, "maturity", "model.json"),
	}

	var modelData map[string]any
	for _, path := range modelPaths {
		if data, err := os.ReadFile(path); err == nil { //nolint:gosec // G703: Path from tracked project config
			if err := json.Unmarshal(data, &modelData); err == nil {
				break
			}
		}
	}

	// Generate HTML dashboard
	html := s.generateMaturityDashboardHTML(modelData, theme, projectName)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// generateMaturityDashboardHTML creates an HTML dashboard for maturity model visualization
func (s *Server) generateMaturityDashboardHTML(model map[string]any, theme, projectName string) string {
	// Default maturity levels
	levels := []struct {
		Level int
		Name  string
		Color string
	}{
		{1, "Reactive", "#ef4444"},
		{2, "Basic", "#f59e0b"},
		{3, "Defined", "#eab308"},
		{4, "Managed", "#22c55e"},
		{5, "Optimizing", "#3b82f6"},
	}

	// Theme colors
	bgColor := "#1b2636"
	textColor := "#e0e6ed"
	mutedColor := "#8899a6"
	borderColor := "#2d3e50"
	panelColor := "#1e2a3a"
	if theme == "light" {
		bgColor = "#ffffff"
		textColor = "#1a1a1a"
		mutedColor = "#6b7280"
		borderColor = "#e5e7eb"
		panelColor = "#f9fafb"
	}

	// Build dimensions HTML
	dimensionsHTML := ""
	if dims, ok := model["dimensions"].([]any); ok {
		for _, d := range dims {
			dim, ok := d.(map[string]any)
			if !ok {
				continue
			}
			name, _ := dim["name"].(string)
			currentLevel := 1
			if cl, ok := dim["currentLevel"].(float64); ok {
				currentLevel = int(cl)
			}
			targetLevel := currentLevel
			if tl, ok := dim["targetLevel"].(float64); ok {
				targetLevel = int(tl)
			}

			// Get level color
			levelColor := levels[0].Color
			if currentLevel >= 1 && currentLevel <= 5 {
				levelColor = levels[currentLevel-1].Color
			}

			dimensionsHTML += fmt.Sprintf(`
				<div style="background: %s; border: 1px solid %s; border-radius: 8px; padding: 16px; margin-bottom: 12px;">
					<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
						<span style="font-weight: 600;">%s</span>
						<span style="background: %s; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px;">M%d</span>
					</div>
					<div style="display: flex; gap: 4px; margin-bottom: 8px;">
						%s
					</div>
					<div style="font-size: 12px; color: %s;">Target: M%d</div>
				</div>
			`, panelColor, borderColor, name, levelColor, currentLevel, s.generateLevelBars(currentLevel, targetLevel, levels), mutedColor, targetLevel)
		}
	} else {
		// No dimensions - show placeholder
		dimensionsHTML = fmt.Sprintf(`
			<div style="text-align: center; padding: 40px; color: %s;">
				<div style="font-size: 48px; margin-bottom: 16px;">📊</div>
				<div style="font-size: 18px; margin-bottom: 8px;">No Maturity Data</div>
				<div style="font-size: 14px;">Add a maturity model configuration to see the dashboard.</div>
			</div>
		`, mutedColor)
	}

	// Build goals HTML
	goalsHTML := ""
	if goals, ok := model["goals"].([]any); ok && len(goals) > 0 {
		goalsHTML = `<h3 style="margin-top: 24px; margin-bottom: 12px;">Goals</h3>`
		for _, g := range goals {
			goal, ok := g.(map[string]any)
			if !ok {
				continue
			}
			name, _ := goal["name"].(string)
			status, _ := goal["status"].(string)
			currentLevel := 1
			if cl, ok := goal["currentLevel"].(float64); ok {
				currentLevel = int(cl)
			}
			targetLevel := currentLevel
			if tl, ok := goal["targetLevel"].(float64); ok {
				targetLevel = int(tl)
			}

			statusColor := mutedColor
			if status == "on_track" {
				statusColor = "#22c55e"
			} else if status == "at_risk" {
				statusColor = "#f59e0b"
			} else if status == "blocked" {
				statusColor = "#ef4444"
			}

			goalsHTML += fmt.Sprintf(`
				<div style="background: %s; border: 1px solid %s; border-radius: 8px; padding: 12px; margin-bottom: 8px; display: flex; justify-content: space-between; align-items: center;">
					<span>%s</span>
					<div style="display: flex; align-items: center; gap: 12px;">
						<span style="font-size: 12px;">M%d → M%d</span>
						<span style="width: 8px; height: 8px; border-radius: 50%%; background: %s;"></span>
					</div>
				</div>
			`, panelColor, borderColor, name, currentLevel, targetLevel, statusColor)
		}
	}

	// Overall score
	overallScore := 0.0
	if score, ok := model["overallScore"].(float64); ok {
		overallScore = score
	}

	modelName := "Maturity Model"
	if name, ok := model["name"].(string); ok && name != "" {
		modelName = name
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>%s - %s</title>
	<style>
		* { box-sizing: border-box; margin: 0; padding: 0; }
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: %s;
			color: %s;
			padding: 24px;
			line-height: 1.5;
		}
		h1 { font-size: 24px; margin-bottom: 8px; }
		h2 { font-size: 18px; margin-bottom: 16px; color: %s; font-weight: normal; }
		h3 { font-size: 16px; font-weight: 600; }
		.summary {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
			gap: 16px;
			margin-bottom: 24px;
		}
		.summary-card {
			background: %s;
			border: 1px solid %s;
			border-radius: 8px;
			padding: 16px;
			text-align: center;
		}
		.summary-value {
			font-size: 32px;
			font-weight: 600;
			margin-bottom: 4px;
		}
		.summary-label {
			font-size: 12px;
			color: %s;
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}
		.legend {
			display: flex;
			gap: 16px;
			margin-bottom: 24px;
			flex-wrap: wrap;
		}
		.legend-item {
			display: flex;
			align-items: center;
			gap: 6px;
			font-size: 12px;
		}
		.legend-dot {
			width: 12px;
			height: 12px;
			border-radius: 2px;
		}
	</style>
</head>
<body>
	<h1>%s</h1>
	<h2>%s</h2>

	<div class="summary">
		<div class="summary-card">
			<div class="summary-value">%.1f</div>
			<div class="summary-label">Overall Score</div>
		</div>
	</div>

	<div class="legend">
		<div class="legend-item"><div class="legend-dot" style="background: %s"></div>M1 Reactive</div>
		<div class="legend-item"><div class="legend-dot" style="background: %s"></div>M2 Basic</div>
		<div class="legend-item"><div class="legend-dot" style="background: %s"></div>M3 Defined</div>
		<div class="legend-item"><div class="legend-dot" style="background: %s"></div>M4 Managed</div>
		<div class="legend-item"><div class="legend-dot" style="background: %s"></div>M5 Optimizing</div>
	</div>

	<h3 style="margin-bottom: 12px;">Dimensions</h3>
	%s

	%s
</body>
</html>`,
		modelName, projectName,
		bgColor, textColor, mutedColor,
		panelColor, borderColor, mutedColor,
		modelName, projectName,
		overallScore,
		levels[0].Color, levels[1].Color, levels[2].Color, levels[3].Color, levels[4].Color,
		dimensionsHTML,
		goalsHTML,
	)
}

// generateLevelBars creates the visual level indicator bars
func (s *Server) generateLevelBars(current, target int, levels []struct {
	Level int
	Name  string
	Color string
}) string {
	html := ""
	for i := 0; i < 5; i++ {
		level := i + 1
		color := "#374151" // Default gray
		if level <= current {
			color = levels[i].Color
		}
		opacity := "1"
		if level > current && level <= target {
			opacity = "0.3"
			color = levels[i].Color
		}
		html += fmt.Sprintf(`<div style="flex: 1; height: 8px; background: %s; opacity: %s; border-radius: 2px;"></div>`, color, opacity)
	}
	return html
}

func (s *Server) handleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	s.logger.Debug("Getting workflow", "project", projectName)

	// Get project from config
	tracked, err := config.GetProject(projectName)
	if err != nil {
		s.writeJSON(w, http.StatusNotFound, api.GetWorkflowResponse{
			Error: err.Error(),
		})
		return
	}

	// Load the profile
	profile, err := profileLoader.Load(tracked.Profile)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.GetWorkflowResponse{
			Error: fmt.Sprintf("failed to load profile: %v", err),
		})
		return
	}

	// Generate workflow from profile
	wf, err := specworkflow.FromProfile(profile)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, api.GetWorkflowResponse{
			Error: fmt.Sprintf("failed to generate workflow: %v", err),
		})
		return
	}

	// Update workflow statuses from project specs on disk
	project := s.loadProjectState(tracked.Path)
	if project != nil {
		specworkflow.UpdateFromProject(wf, project)
	}

	// Render Mermaid diagram
	renderer := workflow.NewMermaidRenderer()
	mermaid := renderer.Render(wf)

	// Convert to API types
	apiWorkflow := convertWorkflowToAPI(wf)

	s.writeJSON(w, http.StatusOK, api.GetWorkflowResponse{
		Workflow: apiWorkflow,
		Mermaid:  mermaid,
	})
}

// loadProjectState loads the project state from disk
func (s *Server) loadProjectState(projectPath string) *types.Project {
	project, err := specconfig.Load(projectPath)
	if err != nil {
		s.logger.Debug("Could not load project config", "path", projectPath, "error", err)
		return nil
	}

	// Scan for spec files and update their statuses
	if project.Specs == nil {
		project.Specs = make(map[types.SpecType]*types.Spec)
	}

	specDirs := map[types.SpecCategory]string{
		types.CategorySource:    specconfig.SourceDir,
		types.CategoryGTM:       specconfig.GTMDir,
		types.CategoryTechnical: specconfig.TechnicalDir,
	}

	for _, specType := range types.AllSpecTypes() {
		cat := specType.Category()
		dir := specDirs[cat]
		if dir == "" {
			continue
		}

		// projectPath from user config, dir and specType.Filename() are hardcoded constants
		specPath := filepath.Join(projectPath, dir, specType.Filename())
		if _, err := os.Stat(specPath); err == nil { //nolint:gosec // paths from user config + hardcoded constants
			project.Specs[specType] = &types.Spec{
				Type:   specType,
				Path:   specPath,
				Status: types.StatusDraft,
			}
		}
	}

	return project
}

// convertWorkflowToAPI converts a workflow.Workflow to api.Workflow
func convertWorkflowToAPI(wf *workflow.Workflow) api.Workflow {
	phases := make([]api.WorkflowPhase, len(wf.Phases))
	for i, p := range wf.Phases {
		phases[i] = api.WorkflowPhase{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Order:       p.Order,
			Nodes:       p.Nodes,
		}
	}

	nodes := make(map[string]api.WorkflowNode)
	for id, n := range wf.Nodes {
		nodes[id] = api.WorkflowNode{
			ID:          n.ID,
			Name:        n.Name,
			Description: n.Description,
			Type:        string(n.Type),
			Phase:       n.Phase,
			Status:      string(n.Status),
			DependsOn:   n.DependsOn,
			Automated:   n.Automated,
			Metadata:    n.Metadata,
		}
	}

	completed, total, percent := wf.Progress()

	return api.Workflow{
		Name:        wf.Name,
		Description: wf.Description,
		Phases:      phases,
		Nodes:       nodes,
		Progress: api.WorkflowProgress{
			Completed: completed,
			Total:     total,
			Percent:   percent,
		},
	}
}

func (s *Server) handleGetWorkflowStatus(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	s.logger.Debug("Getting workflow status", "project", projectName)

	// TODO: Compute actual workflow status from project specs
	// For now, return mock data based on the example project
	specStatuses := map[string]string{
		"mrd":          string(api.SpecStatusEvaluated),
		"press":        string(api.SpecStatusEvaluated),
		"faq":          string(api.SpecStatusEvaluated),
		"narrative-6p": string(api.SpecStatusNotStarted),
		"prd":          string(api.SpecStatusNotStarted),
		"uxd":          string(api.SpecStatusNotStarted),
		"trd":          string(api.SpecStatusNotStarted),
		"tpd":          string(api.SpecStatusNotStarted),
	}

	// Determine current phase based on spec statuses
	currentPhase := "source"
	completedPhases := []string{}
	blockedBy := []string{}

	// Simple logic: if MRD is done, we're in GTM phase
	if specStatuses["mrd"] == string(api.SpecStatusEvaluated) ||
		specStatuses["mrd"] == string(api.SpecStatusApproved) {
		completedPhases = append(completedPhases, "source")
		currentPhase = "gtm"
	}

	// If GTM specs are done, we're in product phase
	gtmDone := (specStatuses["press"] == string(api.SpecStatusEvaluated) ||
		specStatuses["press"] == string(api.SpecStatusApproved)) &&
		(specStatuses["faq"] == string(api.SpecStatusEvaluated) ||
			specStatuses["faq"] == string(api.SpecStatusApproved))

	if gtmDone {
		completedPhases = append(completedPhases, "gtm")
		currentPhase = "product"
		blockedBy = []string{"narrative-6p"}
	}

	// Calculate progress
	totalSpecs := len(specStatuses)
	completedSpecs := 0
	for _, status := range specStatuses {
		if status == string(api.SpecStatusEvaluated) || status == string(api.SpecStatusApproved) {
			completedSpecs++
		}
	}
	progress := float64(completedSpecs) / float64(totalSpecs)

	s.writeJSON(w, http.StatusOK, api.GetWorkflowStatusResponse{
		Status: api.WorkflowStatus{
			CurrentPhase:    currentPhase,
			CompletedPhases: completedPhases,
			Progress:        progress,
			SpecStatuses:    specStatuses,
			BlockedBy:       blockedBy,
			LastUpdated:     time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON", "error", err)
	}
}

// getGitRemote returns the git remote URL for the given directory
func getGitRemote(dir string) string {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// SSE and file watching

// handleSSE handles Server-Sent Events for real-time updates
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")

	// Get project from config
	tracked, err := config.GetProject(projectName)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create channel for this client
	eventChan := make(chan api.FileEvent, 10)

	// Register client
	s.sseClientsMu.Lock()
	if s.sseClients[projectName] == nil {
		s.sseClients[projectName] = make(map[chan api.FileEvent]struct{})
	}
	s.sseClients[projectName][eventChan] = struct{}{}
	s.sseClientsMu.Unlock()

	// Start file watcher for this project if not already running
	s.ensureWatcher(projectName, tracked.Path)

	// Cleanup on disconnect
	defer func() {
		s.sseClientsMu.Lock()
		delete(s.sseClients[projectName], eventChan)
		noMoreClients := len(s.sseClients[projectName]) == 0
		if noMoreClients {
			delete(s.sseClients, projectName)
		}
		s.sseClientsMu.Unlock()

		// Stop watcher outside the lock to avoid potential deadlock
		if noMoreClients {
			s.stopWatcher(projectName)
		}

		// Note: Don't close the channel here - broadcastEvent may still have
		// a reference to it. Let it be garbage collected instead.
	}()

	// Flush support
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial connection event
	s.sendSSEEvent(w, flusher, api.FileEvent{
		Type:      "connected",
		Project:   projectName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})

	s.logger.Info("SSE client connected", "project", projectName)

	// Stream events
	for {
		select {
		case <-r.Context().Done():
			s.logger.Info("SSE client disconnected", "project", projectName)
			return
		case event := <-eventChan:
			s.sendSSEEvent(w, flusher, event)
		}
	}
}

func (s *Server) sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event api.FileEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("Failed to marshal SSE event", "error", err)
		return
	}

	fmt.Fprintf(w, "event: %s\n", event.Type)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// ensureWatcher starts a file watcher for the project if not already running
func (s *Server) ensureWatcher(projectName, projectPath string) {
	s.watchersMu.Lock()
	defer s.watchersMu.Unlock()

	if _, exists := s.watchers[projectName]; exists {
		return // Already watching
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.logger.Error("Failed to create file watcher", "project", projectName, "error", err)
		return
	}

	s.watchers[projectName] = watcher

	// Watch project directories (projectPath from user config, subdirs are hardcoded constants)
	dirsToWatch := []string{
		projectPath,
		filepath.Join(projectPath, specconfig.SourceDir),
		filepath.Join(projectPath, specconfig.GTMDir),
		filepath.Join(projectPath, specconfig.TechnicalDir),
		filepath.Join(projectPath, specconfig.EvalDir),
	}

	for _, dir := range dirsToWatch {
		if _, err := os.Stat(dir); err == nil { //nolint:gosec // paths from user config + hardcoded constants
			if err := watcher.Add(dir); err != nil {
				s.logger.Warn("Failed to watch directory", "dir", dir, "error", err)
			} else {
				s.logger.Debug("Watching directory", "dir", dir)
			}
		}
	}

	// Start goroutine to process events
	go s.processWatchEvents(projectName, projectPath, watcher)

	s.logger.Info("Started file watcher", "project", projectName, "path", projectPath)
}

// stopWatcher stops the file watcher for a project
func (s *Server) stopWatcher(projectName string) {
	s.watchersMu.Lock()
	defer s.watchersMu.Unlock()

	if watcher, exists := s.watchers[projectName]; exists {
		watcher.Close()
		delete(s.watchers, projectName)
		s.logger.Info("Stopped file watcher", "project", projectName)
	}
}

// processWatchEvents processes file system events and broadcasts to SSE clients
func (s *Server) processWatchEvents(projectName, projectPath string, watcher *fsnotify.Watcher) {
	// Debounce map to prevent duplicate events
	debounce := make(map[string]time.Time)
	debounceDuration := 300 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Skip non-file events and debounce
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}

			// Debounce rapid events for the same file
			now := time.Now()
			if lastTime, exists := debounce[event.Name]; exists {
				if now.Sub(lastTime) < debounceDuration {
					continue
				}
			}
			debounce[event.Name] = now

			// Only process .md and .json files
			ext := filepath.Ext(event.Name)
			if ext != ".md" && ext != ".json" && ext != ".yaml" {
				continue
			}

			// Determine event type
			var eventType api.FileEventType
			switch {
			case event.Op&fsnotify.Create != 0:
				eventType = api.EventFileCreated
			case event.Op&fsnotify.Write != 0:
				eventType = api.EventFileModified
			case event.Op&fsnotify.Remove != 0:
				eventType = api.EventFileDeleted
			case event.Op&fsnotify.Rename != 0:
				eventType = api.EventFileRenamed
			default:
				continue
			}

			// Get relative path
			relPath, _ := filepath.Rel(projectPath, event.Name)

			// Determine if this is a spec file and which type
			specType := detectSpecType(relPath)

			// Create file event
			fileEvent := api.FileEvent{
				Type:      eventType,
				Project:   projectName,
				Path:      relPath,
				SpecType:  specType,
				Timestamp: now.UTC().Format(time.RFC3339),
			}

			// Add extra data for eval files
			if strings.Contains(relPath, ".eval.json") && eventType == api.EventFileModified {
				fileEvent.Type = api.EventEvalComplete
				// Try to read eval result
				if evalData, err := os.ReadFile(event.Name); err == nil {
					var evalResult map[string]any
					if json.Unmarshal(evalData, &evalResult) == nil {
						eventData := map[string]any{
							"score":    evalResult["score"],
							"decision": evalResult["decision"],
						}
						// Include v2 fields if present
						if v, ok := evalResult["schemaVersion"]; ok {
							eventData["schemaVersion"] = v
						}
						if v, ok := evalResult["scoreV2"]; ok {
							eventData["scoreV2"] = v
						}
						if v, ok := evalResult["pass"]; ok {
							eventData["pass"] = v
						}
						if v, ok := evalResult["confidence"]; ok {
							eventData["confidence"] = v
						}
						if v, ok := evalResult["blocking"]; ok {
							eventData["blocking"] = v
						}
						fileEvent.Data = eventData
					}
				}
			}

			// Broadcast to all clients watching this project
			s.broadcastEvent(projectName, fileEvent)

			// Also send workflow_changed event if a spec file was modified
			if specType != "" && (eventType == api.EventFileModified || eventType == api.EventFileCreated) {
				s.broadcastEvent(projectName, api.FileEvent{
					Type:      api.EventWorkflowChanged,
					Project:   projectName,
					Timestamp: now.UTC().Format(time.RFC3339),
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			s.logger.Error("File watcher error", "project", projectName, "error", err)
		}
	}
}

// detectSpecType determines the spec type from a file path
func detectSpecType(relPath string) string {
	base := filepath.Base(relPath)

	// Remove extension
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Check known spec types
	specTypes := []string{"mrd", "prd", "uxd", "press", "faq", "narrative-1p", "narrative-6p", "trd", "tpd", "ird", "spec"}
	for _, st := range specTypes {
		if name == st {
			return st
		}
	}

	return ""
}

// broadcastEvent sends an event to all SSE clients watching a project
func (s *Server) broadcastEvent(projectName string, event api.FileEvent) {
	s.sseClientsMu.RLock()
	// Copy the channels to a slice to avoid holding the lock while sending
	clients := make([]chan api.FileEvent, 0, len(s.sseClients[projectName]))
	for clientChan := range s.sseClients[projectName] {
		clients = append(clients, clientChan)
	}
	s.sseClientsMu.RUnlock()

	for _, clientChan := range clients {
		// Use a select with default to avoid blocking if channel is full
		// The channel might have been removed between getting the list and now,
		// but since we don't close channels, this is safe
		select {
		case clientChan <- event:
		default:
			// Channel full, skip this event for this client
			s.logger.Warn("SSE client channel full, dropping event", "project", projectName)
		}
	}
}
