package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// API response types for the JSON API.
// These mirror the dashboard data structures but are JSON-clean for frontend consumption.

type APIProgram struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Hidden      bool   `json:"hidden,omitempty"`
}

type APIInitiative struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Status      string  `json:"status"`
	Type        string  `json:"type,omitempty"`
	ProgramID   string  `json:"programId,omitempty"`
	ProgramName string  `json:"programName,omitempty"`
	HomeRepo    string  `json:"homeRepo,omitempty"`
	Progress    float64 `json:"progress"`
}

type APIPhase struct {
	ID             string  `json:"id"`
	InitiativeID   string  `json:"initiativeId"`
	Title          string  `json:"title"`
	SequenceNumber int     `json:"sequenceNumber"`
	Progress       float64 `json:"progress"`
}

type APIRMI struct {
	ID             string  `json:"id"`
	InitiativeID   string  `json:"initiativeId"`
	PhaseID        string  `json:"phaseId"`
	Title          string  `json:"title"`
	Status         string  `json:"status"`
	Type           string  `json:"type,omitempty"`
	RepositoryID   string  `json:"repositoryId,omitempty"`
	SequenceNumber int     `json:"sequenceNumber"`
	ClaimedBy      string  `json:"claimedBy,omitempty"`
	ClaimedAt      string  `json:"claimedAt,omitempty"`
	CompletedAt    string  `json:"completedAt,omitempty"`
	TokensTotal    int64   `json:"tokensTotal,omitempty"`
	CostUSD        float64 `json:"costUsd,omitempty"`
}

type APITokens struct {
	InputTokens         int64   `json:"inputTokens"`
	OutputTokens        int64   `json:"outputTokens"`
	CacheReadTokens     int64   `json:"cacheReadTokens"`
	CacheCreationTokens int64   `json:"cacheCreationTokens"`
	TotalTokens         int64   `json:"totalTokens"`
	CostUSD             float64 `json:"costUsd"`
}

type APIStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type APIRMIDependency struct {
	SourceRMIID  string `json:"sourceRmiId"`
	TargetRMIID  string `json:"targetRmiId"`
	Relationship string `json:"relationship"`
}

type APIInitiativeDependency struct {
	SourceInitiativeID string `json:"sourceInitiativeId"`
	TargetInitiativeID string `json:"targetInitiativeId"`
	Relationship       string `json:"relationship"`
}

// ExecutionResponse is the response for /api/execution.
type ExecutionResponse struct {
	Programs               []APIProgram              `json:"programs"`
	Initiatives            []APIInitiative           `json:"initiatives"`
	Phases                 []APIPhase                `json:"phases"`
	RMIs                   []APIRMI                  `json:"rmis"`
	StatusDist             []APIStatusCount          `json:"statusDistribution"`
	RMIDependencies        []APIRMIDependency        `json:"rmiDependencies"`
	InitiativeDependencies []APIInitiativeDependency `json:"initiativeDependencies"`
}

// SpendResponse is the response for /api/spend.
type SpendResponse struct {
	Total        *APITokens            `json:"total,omitempty"`
	ByInitiative map[string]*APITokens `json:"byInitiative,omitempty"`
	ByPhase      map[string]*APITokens `json:"byPhase,omitempty"`
	ByRMI        map[string]*APITokens `json:"byRmi,omitempty"`
	HasData      bool                  `json:"hasData"`
	DataNote     string                `json:"dataNote,omitempty"`
}

// MaturityResponse is the response for /api/maturity.
type MaturityResponse struct {
	Models      []*store.CapabilityModel    `json:"models"`
	Assessments []*store.MaturityAssessment `json:"assessments"`
}

// SpecsResponse is the response for /api/specs.
type SpecsResponse struct {
	Workflows    []*store.SpecWorkflow `json:"workflows"`
	JudgeResults []*store.JudgeResult  `json:"judgeResults"`
}

// SpecFile represents a spec document read from disk.
type SpecFile struct {
	InitiativeID string `json:"initiativeId"`
	SpecType     string `json:"specType"`
	Path         string `json:"path"`
	Content      string `json:"content"`
	ModTime      string `json:"modTime,omitempty"`
	EvalJSON     string `json:"evalJson,omitempty"`
}

// SpecFilesResponse is the response for /api/spec-files.
type SpecFilesResponse struct {
	Files []SpecFile `json:"files"`
}

// registerAPIRoutes adds JSON API endpoints to the dashboard mux.
func registerAPIRoutes(mux *http.ServeMux, connectSvc func() (*service.Service, func(), error), dataDir string) {
	// CORS middleware for local development
	cors := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h(w, r)
		}
	}

	writeJSON := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	mux.HandleFunc("/api/execution", cors(func(w http.ResponseWriter, r *http.Request) {
		svc, cleanup, err := connectSvc()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cleanup()

		resp, err := buildExecutionResponse(r.Context(), svc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp)
	}))

	mux.HandleFunc("/api/spend", cors(func(w http.ResponseWriter, r *http.Request) {
		svc, cleanup, err := connectSvc()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cleanup()

		resp, err := buildSpendResponse(r.Context(), svc, dataDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp)
	}))

	mux.HandleFunc("/api/maturity", cors(func(w http.ResponseWriter, r *http.Request) {
		svc, cleanup, err := connectSvc()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cleanup()

		resp, err := buildMaturityResponse(r.Context(), svc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp)
	}))

	mux.HandleFunc("/api/specs", cors(func(w http.ResponseWriter, r *http.Request) {
		svc, cleanup, err := connectSvc()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cleanup()

		resp, err := buildSpecsResponse(r.Context(), svc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp)
	}))

	// Spec files endpoint - reads spec markdown and eval JSON from disk
	mux.HandleFunc("/api/spec-files/", cors(func(w http.ResponseWriter, r *http.Request) {
		initID := strings.TrimPrefix(r.URL.Path, "/api/spec-files/")
		if initID == "" {
			http.Error(w, "initiative ID required", http.StatusBadRequest)
			return
		}

		svc, cleanup, err := connectSvc()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cleanup()

		resp, err := buildSpecFilesResponse(r.Context(), svc, initID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, resp)
	}))
}

func buildExecutionResponse(ctx context.Context, svc *service.Service) (*ExecutionResponse, error) {
	programs, err := svc.Store.ListPrograms(ctx)
	if err != nil {
		return nil, err
	}
	programByID := map[string]*store.Program{}
	var apiPrograms []APIProgram
	for _, p := range programs {
		programByID[p.ID] = p
		apiPrograms = append(apiPrograms, APIProgram{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Hidden:      p.Hidden,
		})
	}

	initiatives, err := svc.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}

	allAssignments, err := svc.Store.ListAllAssignments(ctx)
	if err != nil {
		return nil, err
	}
	assignmentsByRMI := map[string]*store.Assignment{}
	for _, a := range allAssignments {
		existing, ok := assignmentsByRMI[a.RMIID]
		if !ok || a.CreatedAt.After(existing.CreatedAt) {
			assignmentsByRMI[a.RMIID] = a
		}
	}

	statusCounts := map[string]int{}
	var apiInitiatives []APIInitiative
	var apiPhases []APIPhase
	var apiRMIs []APIRMI

	for _, init := range initiatives {
		phases, err := svc.Store.ListPhases(ctx, init.ID)
		if err != nil {
			return nil, err
		}

		rmis, err := svc.Store.ListRMIs(ctx, init.ID)
		if err != nil {
			return nil, err
		}

		totalRMIs := len(rmis)
		completedRMIs := 0
		for _, r := range rmis {
			statusCounts[strings.ToLower(r.Status)]++
			if strings.ToLower(r.Status) == "completed" {
				completedRMIs++
			}

			apiRMI := APIRMI{
				ID:             r.ID,
				InitiativeID:   init.ID,
				PhaseID:        r.PhaseID,
				Title:          r.Title,
				Status:         r.Status,
				Type:           r.ItemType,
				RepositoryID:   r.RepositoryID,
				SequenceNumber: r.SequenceNumber,
			}
			if r.CompletedAt != nil {
				apiRMI.CompletedAt = r.CompletedAt.Format("2006-01-02T15:04:05Z")
			}
			if a, ok := assignmentsByRMI[r.ID]; ok {
				apiRMI.ClaimedBy = a.Worker
				apiRMI.ClaimedAt = a.CreatedAt.Format("2006-01-02T15:04:05Z")
			}
			apiRMIs = append(apiRMIs, apiRMI)
		}

		for _, p := range phases {
			phaseRMIs := 0
			phaseCompleted := 0
			for _, r := range rmis {
				if r.PhaseID == p.ID {
					phaseRMIs++
					if strings.ToLower(r.Status) == "completed" {
						phaseCompleted++
					}
				}
			}
			progress := 0.0
			if phaseRMIs > 0 {
				progress = float64(phaseCompleted) / float64(phaseRMIs)
			}
			apiPhases = append(apiPhases, APIPhase{
				ID:             p.ID,
				InitiativeID:   init.ID,
				Title:          p.Title,
				SequenceNumber: p.SequenceNumber,
				Progress:       progress,
			})
		}

		progress := 0.0
		if totalRMIs > 0 {
			progress = float64(completedRMIs) / float64(totalRMIs)
		}

		programName := ""
		if p, ok := programByID[init.ProgramID]; ok {
			programName = p.Name
		}

		apiInitiatives = append(apiInitiatives, APIInitiative{
			ID:          init.ID,
			Title:       init.Title,
			Description: init.Description,
			Status:      init.Status,
			Type:        init.InitType,
			ProgramID:   init.ProgramID,
			ProgramName: programName,
			HomeRepo:    init.HomeRepo,
			Progress:    progress,
		})
	}

	var statusDist []APIStatusCount
	for status, count := range statusCounts {
		statusDist = append(statusDist, APIStatusCount{Status: status, Count: count})
	}

	// Load dependencies
	allDeps, err := svc.Store.ListAllDependencies(ctx)
	if err != nil {
		return nil, err
	}
	var rmiDeps []APIRMIDependency
	for _, d := range allDeps {
		rmiDeps = append(rmiDeps, APIRMIDependency{
			SourceRMIID:  d.SourceRMIID,
			TargetRMIID:  d.TargetRMIID,
			Relationship: d.Relationship,
		})
	}

	initDeps, err := svc.Store.ListAllInitiativeDependencies(ctx)
	if err != nil {
		return nil, err
	}
	var apiInitDeps []APIInitiativeDependency
	for _, d := range initDeps {
		apiInitDeps = append(apiInitDeps, APIInitiativeDependency{
			SourceInitiativeID: d.SourceInitiativeID,
			TargetInitiativeID: d.TargetInitiativeID,
			Relationship:       d.Relationship,
		})
	}

	return &ExecutionResponse{
		Programs:               apiPrograms,
		Initiatives:            apiInitiatives,
		Phases:                 apiPhases,
		RMIs:                   apiRMIs,
		StatusDist:             statusDist,
		RMIDependencies:        rmiDeps,
		InitiativeDependencies: apiInitDeps,
	}, nil
}

func buildSpendResponse(ctx context.Context, svc *service.Service, dataDir string) (*SpendResponse, error) {
	// Reuse loadDashboardData to get token data
	data, err := loadDashboardData(ctx, svc, dataDir)
	if err != nil {
		return nil, err
	}

	resp := &SpendResponse{
		HasData:      data.HasTokenData,
		DataNote:     data.TokenDataNote,
		ByInitiative: make(map[string]*APITokens),
		ByPhase:      make(map[string]*APITokens),
		ByRMI:        make(map[string]*APITokens),
	}

	if data.TotalTokens != nil {
		resp.Total = &APITokens{
			InputTokens:         data.TotalTokens.InputTokens,
			OutputTokens:        data.TotalTokens.OutputTokens,
			CacheReadTokens:     data.TotalTokens.CacheReadTokens,
			CacheCreationTokens: data.TotalTokens.CacheCreationTokens,
			TotalTokens:         data.TotalTokens.TotalTokens,
			CostUSD:             data.TotalTokens.CostUSD,
		}
	}

	for _, init := range data.Initiatives {
		if init.Tokens != nil {
			resp.ByInitiative[init.Initiative.ID] = &APITokens{
				InputTokens:         init.Tokens.InputTokens,
				OutputTokens:        init.Tokens.OutputTokens,
				CacheReadTokens:     init.Tokens.CacheReadTokens,
				CacheCreationTokens: init.Tokens.CacheCreationTokens,
				TotalTokens:         init.Tokens.TotalTokens,
				CostUSD:             init.Tokens.CostUSD,
			}
		}
		for _, phase := range init.Phases {
			if phase.Tokens != nil {
				resp.ByPhase[phase.Phase.ID] = &APITokens{
					InputTokens:         phase.Tokens.InputTokens,
					OutputTokens:        phase.Tokens.OutputTokens,
					CacheReadTokens:     phase.Tokens.CacheReadTokens,
					CacheCreationTokens: phase.Tokens.CacheCreationTokens,
					TotalTokens:         phase.Tokens.TotalTokens,
					CostUSD:             phase.Tokens.CostUSD,
				}
			}
			for _, rmi := range phase.RMIs {
				if rmi.Tokens != nil {
					resp.ByRMI[rmi.RMI.ID] = &APITokens{
						InputTokens:         rmi.Tokens.InputTokens,
						OutputTokens:        rmi.Tokens.OutputTokens,
						CacheReadTokens:     rmi.Tokens.CacheReadTokens,
						CacheCreationTokens: rmi.Tokens.CacheCreationTokens,
						TotalTokens:         rmi.Tokens.TotalTokens,
						CostUSD:             rmi.Tokens.CostUSD,
					}
				}
			}
		}
	}

	return resp, nil
}

func buildMaturityResponse(ctx context.Context, svc *service.Service) (*MaturityResponse, error) {
	models, err := svc.Store.ListCapabilityModels(ctx)
	if err != nil {
		return nil, err
	}

	// Collect assessments from all initiatives
	initiatives, err := svc.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}

	var allAssessments []*store.MaturityAssessment
	for _, init := range initiatives {
		assessments, err := svc.Store.ListMaturityAssessments(ctx, init.ID)
		if err != nil {
			return nil, err
		}
		allAssessments = append(allAssessments, assessments...)
	}

	return &MaturityResponse{
		Models:      models,
		Assessments: allAssessments,
	}, nil
}

func buildSpecsResponse(ctx context.Context, svc *service.Service) (*SpecsResponse, error) {
	workflows, err := svc.Store.ListSpecWorkflows(ctx)
	if err != nil {
		return nil, err
	}

	// Get all judge results across all initiatives
	initiatives, err := svc.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}

	var allResults []*store.JudgeResult
	seenIDs := make(map[string]bool)

	// First, load from store
	for _, init := range initiatives {
		results, err := svc.Store.ListJudgeResults(ctx, init.ID)
		if err != nil {
			return nil, err
		}
		for _, r := range results {
			seenIDs[r.ID] = true
			allResults = append(allResults, r)
		}
	}

	// Also load from disk (evaluations/ directories) for results not in store
	for _, init := range initiatives {
		diskResults := loadJudgeResultsFromDisk(init.ID, svc)
		for _, r := range diskResults {
			if !seenIDs[r.ID] {
				seenIDs[r.ID] = true
				allResults = append(allResults, r)
			}
		}
	}

	// Also scan for disk-only initiatives (not in database)
	diskInitIDs := scanDiskInitiatives()
	for _, initID := range diskInitIDs {
		// Skip if we already processed from store
		alreadyProcessed := false
		for _, init := range initiatives {
			if init.ID == initID {
				alreadyProcessed = true
				break
			}
		}
		if alreadyProcessed {
			continue
		}

		diskResults := loadJudgeResultsFromDisk(initID, svc)
		for _, r := range diskResults {
			if !seenIDs[r.ID] {
				seenIDs[r.ID] = true
				allResults = append(allResults, r)
			}
		}
	}

	return &SpecsResponse{
		Workflows:    workflows,
		JudgeResults: allResults,
	}, nil
}

// scanDiskInitiatives looks for initiative directories in the current working directory.
func scanDiskInitiatives() []string {
	var initIDs []string

	cwd, err := os.Getwd()
	if err != nil {
		return initIDs
	}

	initDir := filepath.Join(cwd, "docs", "specs", "initiatives")
	entries, err := os.ReadDir(initDir)
	if err != nil {
		return initIDs
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "INIT-") {
			initIDs = append(initIDs, entry.Name())
		}
	}

	return initIDs
}

// loadJudgeResultsFromDisk reads *.eval.json files from disk and converts them to JudgeResults.
func loadJudgeResultsFromDisk(initiativeID string, svc *service.Service) []*store.JudgeResult {
	var results []*store.JudgeResult
	var evalDir string

	// Try to find the evaluations directory
	// First check if initiative has HomeRepo with LocalPath
	init, err := svc.Store.GetInitiative(context.Background(), initiativeID)
	if err == nil && init.HomeRepo != "" {
		repo, err := svc.Store.GetRepository(context.Background(), init.HomeRepo)
		if err == nil && repo.LocalPath != "" {
			evalDir = filepath.Join(repo.LocalPath, "docs", "specs", "initiatives", initiativeID, "evaluations")
		}
	}

	// Fallback: check current working directory
	if evalDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return results
		}
		evalDir = filepath.Join(cwd, "docs", "specs", "initiatives", initiativeID, "evaluations")
	}
	entries, err := os.ReadDir(evalDir)
	if err != nil {
		return results
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".eval.json") {
			continue
		}

		filePath := filepath.Join(evalDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var evalFile struct {
			ID           string    `json:"id"`
			InitiativeID string    `json:"initiative_id"`
			SpecPath     string    `json:"spec_path"`
			RubricID     string    `json:"rubric_id"`
			Score        float64   `json:"score"`
			Rationale    string    `json:"rationale"`
			Model        string    `json:"model"`
			EvaluatedAt  time.Time `json:"evaluated_at"`
		}
		if err := json.Unmarshal(data, &evalFile); err != nil {
			continue
		}

		results = append(results, &store.JudgeResult{
			ID:           evalFile.ID,
			InitiativeID: evalFile.InitiativeID,
			SpecPath:     evalFile.SpecPath,
			RubricID:     evalFile.RubricID,
			Score:        evalFile.Score,
			Rationale:    evalFile.Rationale,
			Model:        evalFile.Model,
			EvaluatedAt:  evalFile.EvaluatedAt,
		})
	}

	return results
}

func buildSpecFilesResponse(ctx context.Context, svc *service.Service, initiativeID string) (*SpecFilesResponse, error) {
	init, err := svc.Store.GetInitiative(ctx, initiativeID)
	if err != nil {
		return nil, err
	}

	var specDir string

	// Try to find specs directory via HomeRepo -> Repository.LocalPath
	if init.HomeRepo != "" {
		repo, err := svc.Store.GetRepository(ctx, init.HomeRepo)
		if err == nil && repo.LocalPath != "" {
			specDir = filepath.Join(repo.LocalPath, "docs", "specs", "initiatives", initiativeID)
		}
	}

	// If no specDir from HomeRepo, try current working directory
	if specDir == "" {
		cwd, err := os.Getwd()
		if err == nil {
			candidate := filepath.Join(cwd, "docs", "specs", "initiatives", initiativeID)
			if _, statErr := os.Stat(candidate); statErr == nil {
				specDir = candidate
			}
		}
	}

	if specDir == "" {
		return &SpecFilesResponse{Files: []SpecFile{}}, nil
	}

	files, err := readSpecFiles(specDir, initiativeID)
	if err != nil {
		return &SpecFilesResponse{Files: []SpecFile{}}, nil
	}

	return &SpecFilesResponse{Files: files}, nil
}

// readSpecFiles reads spec files from a directory, supporting both:
// 1. Flat structure: PRD.md, TRD.md, etc. directly in the directory
// 2. VisionSpec structure: source/*.md + eval/*.json
func readSpecFiles(specDir, initiativeID string) ([]SpecFile, error) {
	var files []SpecFile

	// Check for VisionSpec structure (source/ subdirectory)
	sourceDir := filepath.Join(specDir, "source")
	evalDir := filepath.Join(specDir, "eval")

	if _, err := os.Stat(sourceDir); err == nil {
		// VisionSpec structure: read from source/ and eval/
		entries, err := os.ReadDir(sourceDir)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			specPath := filepath.Join(sourceDir, entry.Name())
			content, err := os.ReadFile(specPath)
			if err != nil {
				continue
			}

			info, _ := entry.Info()
			modTime := ""
			if info != nil {
				modTime = info.ModTime().Format("2006-01-02T15:04:05Z")
			}

			specType := deriveSpecType(entry.Name())

			sf := SpecFile{
				InitiativeID: initiativeID,
				SpecType:     specType,
				Path:         specPath,
				Content:      string(content),
				ModTime:      modTime,
			}

			// Check for corresponding eval JSON
			evalName := strings.TrimSuffix(entry.Name(), ".md") + ".json"
			evalPath := filepath.Join(evalDir, evalName)
			if evalContent, err := os.ReadFile(evalPath); err == nil {
				sf.EvalJSON = string(evalContent)
			}

			files = append(files, sf)
		}
	} else {
		// Flat structure: read .md files directly from specDir
		entries, err := os.ReadDir(specDir)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			specPath := filepath.Join(specDir, entry.Name())
			content, err := os.ReadFile(specPath)
			if err != nil {
				continue
			}

			info, _ := entry.Info()
			modTime := ""
			if info != nil {
				modTime = info.ModTime().Format("2006-01-02T15:04:05Z")
			}

			sf := SpecFile{
				InitiativeID: initiativeID,
				SpecType:     deriveSpecType(entry.Name()),
				Path:         specPath,
				Content:      string(content),
				ModTime:      modTime,
			}

			// Check for corresponding eval JSON in evaluations/ directory
			evalName := strings.ToLower(strings.TrimSuffix(entry.Name(), ".md")) + ".eval.json"
			evalPath := filepath.Join(specDir, "evaluations", evalName)
			if evalContent, err := os.ReadFile(evalPath); err == nil {
				sf.EvalJSON = string(evalContent)
			}

			files = append(files, sf)
		}
	}

	return files, nil
}

func deriveSpecType(filename string) string {
	name := strings.ToLower(filename)
	name = strings.TrimSuffix(name, ".md")

	switch {
	case strings.Contains(name, "prd"):
		return "PRD"
	case strings.Contains(name, "trd"):
		return "TRD"
	case strings.Contains(name, "plan"):
		return "PLAN"
	case strings.Contains(name, "roadmap"):
		return "ROADMAP"
	case strings.Contains(name, "mrd"):
		return "MRD"
	case strings.Contains(name, "uxd"):
		return "UXD"
	case strings.Contains(name, "opportunity"):
		return "OPPORTUNITY"
	default:
		return strings.ToUpper(name)
	}
}
