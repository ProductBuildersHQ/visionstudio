package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/ProductBuildersHQ/visionstudio/pkg/api"
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
}

// NewServer creates a new server
func NewServer(logger *slog.Logger) *Server {
	return &Server{logger: logger}
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

		// Projects
		r.Get("/projects", s.handleListProjects)
		r.Get("/projects/{project}", s.handleGetProject)

		// Specs
		r.Get("/projects/{project}/specs/{specType}", s.handleGetSpec)
		r.Put("/projects/{project}/specs/{specType}", s.handleSaveSpec)
		r.Post("/projects/{project}/specs/{specType}/evaluate", s.handleEvaluateSpec)

		// Workflow
		r.Get("/projects/{project}/workflow/status", s.handleGetWorkflowStatus)

		// Chat
		r.Post("/chat", s.handleChat)
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	// TODO: Integrate with visionspec to list actual projects
	// For now, return mock data
	projects := []api.Project{
		{
			Name: "example-project",
			Path: "docs/specs/example-project",
			Profile: api.Profile{
				Name:        "big-tech-product",
				Description: "Big Tech methodology for product development",
				Workflow:    []string{"mrd", "press", "faq", "narrative-6p", "prd", "uxd", "trd", "tpd"},
			},
			Specs: []api.Spec{
				{Type: "mrd", Name: "Market Requirements", Path: "source/mrd.md", Status: api.SpecStatusEvaluated,
					EvalResult: &api.EvalResult{Score: 8.2, Decision: api.EvalDecisionPass}},
				{Type: "press", Name: "Press Release", Path: "gtm/press.md", Status: api.SpecStatusEvaluated,
					EvalResult: &api.EvalResult{Score: 7.5, Decision: api.EvalDecisionPass}},
				{Type: "faq", Name: "FAQ", Path: "gtm/faq.md", Status: api.SpecStatusEvaluated,
					EvalResult: &api.EvalResult{Score: 6.1, Decision: api.EvalDecisionConditional}},
				{Type: "narrative-6p", Name: "6-Pager", Path: "gtm/narrative.md", Status: api.SpecStatusNotStarted},
				{Type: "prd", Name: "Product Requirements", Path: "source/prd.md", Status: api.SpecStatusNotStarted},
				{Type: "uxd", Name: "User Experience", Path: "source/uxd.md", Status: api.SpecStatusNotStarted},
				{Type: "trd", Name: "Technical Design", Path: "technical/trd.md", Status: api.SpecStatusNotStarted},
				{Type: "tpd", Name: "Test Plan", Path: "technical/tpd.md", Status: api.SpecStatusNotStarted},
			},
		},
	}

	s.writeJSON(w, http.StatusOK, api.ListProjectsResponse{Projects: projects})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	s.logger.Debug("Getting project", "name", projectName)

	// TODO: Load actual project from filesystem
	s.writeJSON(w, http.StatusOK, api.GetProjectResponse{
		Project: api.Project{
			Name: projectName,
			Path: fmt.Sprintf("docs/specs/%s", projectName),
		},
	})
}

func (s *Server) handleGetSpec(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	specType := chi.URLParam(r, "specType")
	s.logger.Debug("Getting spec", "project", projectName, "type", specType)

	// TODO: Load actual spec content from filesystem
	content := `# Market Requirements Document: VisionStudio

## 1. Executive Summary

VisionStudio is a desktop application that provides an LLM-powered workspace for authoring, evaluating, and iterating on product specifications...

## 2. Problem Statement

Product teams creating specifications face several challenges:

1. **Fragmented tooling** - Specs live in Google Docs, Notion, Confluence
2. **Quality inconsistency** - No objective evaluation of spec completeness
3. **Manual iteration** - Teams manually review specs without structured feedback
`

	s.writeJSON(w, http.StatusOK, api.GetSpecResponse{
		Spec: api.Spec{
			Type:    specType,
			Name:    "Market Requirements",
			Path:    fmt.Sprintf("source/%s.md", specType),
			Status:  api.SpecStatusEvaluated,
			Content: content,
		},
	})
}

func (s *Server) handleSaveSpec(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "project")
	specType := chi.URLParam(r, "specType")

	var req api.SaveSpecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, api.SaveSpecResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	s.logger.Debug("Saving spec", "project", projectName, "type", specType, "contentLen", len(req.Content))

	// TODO: Save to filesystem
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
