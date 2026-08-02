// Package webapi serves the VisionStudio JSON API consumed by the unified
// web app. The contract matches the API introduced on the prism-build
// dashboard server (RMI-PRISMBUILD-101): /api/execution, /api/spend,
// /api/maturity, /api/specs.
package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProductBuildersHQ/visionstudio/pkg/report"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/tokens"
)

// API response types. These mirror the dashboard data structures but are
// JSON-clean for frontend consumption.

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

// ConnectFunc returns a connected Service plus a cleanup function.
type ConnectFunc func() (*service.Service, func(), error)

// RegisterRoutes adds the JSON API endpoints to mux. dataDir is the
// omnidevx data directory used for token spend attribution; pass ""
// to use the default (~/.plexusone/omnidevx/data).
func RegisterRoutes(mux *http.ServeMux, connectSvc ConnectFunc, dataDir string) {
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

	handle := func(path string, build func(ctx context.Context, svc *service.Service) (any, error)) {
		mux.HandleFunc(path, cors(func(w http.ResponseWriter, r *http.Request) {
			svc, cleanup, err := connectSvc()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cleanup()

			resp, err := build(r.Context(), svc)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, resp)
		}))
	}

	handle("/api/execution", func(ctx context.Context, svc *service.Service) (any, error) {
		return BuildExecutionResponse(ctx, svc)
	})
	handle("/api/spend", func(ctx context.Context, svc *service.Service) (any, error) {
		return BuildSpendResponse(ctx, svc, dataDir)
	})
	handle("/api/maturity", func(ctx context.Context, svc *service.Service) (any, error) {
		return BuildMaturityResponse(ctx, svc)
	})
	handle("/api/specs", func(ctx context.Context, svc *service.Service) (any, error) {
		return BuildSpecsResponse(ctx, svc)
	})
}

// BuildExecutionResponse assembles the /api/execution payload.
func BuildExecutionResponse(ctx context.Context, svc *service.Service) (*ExecutionResponse, error) {
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

func totalsToAPI(t *report.TokenTotals) *APITokens {
	return &APITokens{
		InputTokens:         t.InputTokens,
		OutputTokens:        t.OutputTokens,
		CacheReadTokens:     t.CacheReadTokens,
		CacheCreationTokens: t.CacheCreationTokens,
		TotalTokens:         t.TotalTokens,
		CostUSD:             t.CostUSD,
	}
}

func addTotals(dst *APITokens, t report.TokenTotals) {
	dst.InputTokens += t.InputTokens
	dst.OutputTokens += t.OutputTokens
	dst.CacheReadTokens += t.CacheReadTokens
	dst.CacheCreationTokens += t.CacheCreationTokens
	dst.TotalTokens += t.TotalTokens
	dst.CostUSD += t.CostUSD
}

// BuildSpendResponse assembles the /api/spend payload from omnidevx
// token events attributed to initiatives, phases, and RMIs.
func BuildSpendResponse(ctx context.Context, svc *service.Service, dataDir string) (*SpendResponse, error) {
	resp := &SpendResponse{
		ByInitiative: make(map[string]*APITokens),
		ByPhase:      make(map[string]*APITokens),
		ByRMI:        make(map[string]*APITokens),
	}

	tokenSource, err := tokens.NewJSONLSource(dataDir)
	if err != nil {
		resp.DataNote = "Token data unavailable"
		return resp, nil
	}
	eventsDir := filepath.Join(tokenSource.Dir, "events")
	if _, statErr := os.Stat(eventsDir); statErr != nil {
		resp.DataNote = "No omnidevx events directory found"
		return resp, nil
	}
	resp.HasData = true

	initiatives, err := svc.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}

	var total *APITokens
	for _, init := range initiatives {
		tr, trErr := report.GenerateInitiativeTokenReport(ctx, svc.Store, tokenSource, init.ID)
		if trErr != nil || tr.Totals.TotalTokens == 0 {
			continue
		}
		if total == nil {
			total = &APITokens{}
		}
		addTotals(total, tr.Totals)
		resp.ByInitiative[init.ID] = totalsToAPI(&tr.Totals)

		for _, rt := range tr.ByRMI {
			t := rt.Totals
			resp.ByRMI[rt.RMIID] = totalsToAPI(&t)
			if rt.PhaseID != "" {
				if resp.ByPhase[rt.PhaseID] == nil {
					resp.ByPhase[rt.PhaseID] = &APITokens{}
				}
				addTotals(resp.ByPhase[rt.PhaseID], rt.Totals)
			}
		}
	}
	resp.Total = total

	return resp, nil
}

// BuildMaturityResponse assembles the /api/maturity payload.
func BuildMaturityResponse(ctx context.Context, svc *service.Service) (*MaturityResponse, error) {
	models, err := svc.Store.ListCapabilityModels(ctx)
	if err != nil {
		return nil, err
	}

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

// BuildSpecsResponse assembles the /api/specs payload.
func BuildSpecsResponse(ctx context.Context, svc *service.Service) (*SpecsResponse, error) {
	workflows, err := svc.Store.ListSpecWorkflows(ctx)
	if err != nil {
		return nil, err
	}

	initiatives, err := svc.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}

	var allResults []*store.JudgeResult
	for _, init := range initiatives {
		results, err := svc.Store.ListJudgeResults(ctx, init.ID)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return &SpecsResponse{
		Workflows:    workflows,
		JudgeResults: allResults,
	}, nil
}
