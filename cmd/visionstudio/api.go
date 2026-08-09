package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/scale"
	"github.com/ProductBuildersHQ/scale/catalog"
	scalereport "github.com/ProductBuildersHQ/scale/report"
	"github.com/grokify/gogit/scanner"
	"github.com/plexusone/structured-evaluation/rubric"

	"github.com/ProductBuildersHQ/visionstudio/pkg/apitypes"
	"github.com/ProductBuildersHQ/visionstudio/pkg/ir"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
	"github.com/ProductBuildersHQ/visionstudio/pkg/tokens"
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

type APIRepository struct {
	ID             string `json:"id"`
	Organization   string `json:"organization"`
	RepositoryName string `json:"repositoryName"`
	DefaultBranch  string `json:"defaultBranch,omitempty"`
	LocalPath      string `json:"localPath,omitempty"`
	GoModule       string `json:"goModule,omitempty"`
	Domain         string `json:"domain,omitempty"`
	Status         string `json:"status"`
}

// ExecutionResponse is the response for /api/execution.
type ExecutionResponse struct {
	Programs               []APIProgram              `json:"programs"`
	Initiatives            []APIInitiative           `json:"initiatives"`
	Phases                 []APIPhase                `json:"phases"`
	RMIs                   []APIRMI                  `json:"rmis"`
	Repositories           []APIRepository           `json:"repositories"`
	StatusDist             []APIStatusCount          `json:"statusDistribution"`
	RMIDependencies        []APIRMIDependency        `json:"rmiDependencies"`
	InitiativeDependencies []APIInitiativeDependency `json:"initiativeDependencies"`
}

// APITimeBucket represents token spend for a time bucket (week/month).
type APITimeBucket struct {
	Period  string                `json:"period"`  // e.g., "2026-W31" or "2026-07"
	Start   string                `json:"start"`   // ISO date
	End     string                `json:"end"`     // ISO date
	Totals  *APITokens            `json:"totals"`
	ByModel map[string]*APITokens `json:"byModel,omitempty"`
}

// SpendResponse is the response for /api/spend.
type SpendResponse struct {
	Total        *APITokens            `json:"total,omitempty"`
	ByModel      map[string]*APITokens `json:"byModel,omitempty"`
	ByInitiative map[string]*APITokens `json:"byInitiative,omitempty"`
	ByPhase      map[string]*APITokens `json:"byPhase,omitempty"`
	ByRMI        map[string]*APITokens `json:"byRmi,omitempty"`
	ByWeek       []APITimeBucket       `json:"byWeek,omitempty"`
	ByMonth      []APITimeBucket       `json:"byMonth,omitempty"`
	HasData      bool                  `json:"hasData"`
	DataNote     string                `json:"dataNote,omitempty"`
}

// MaturityResponse is the response for /api/maturity.
type MaturityResponse struct {
	Models      []*store.CapabilityModel    `json:"models"`
	Assessments []*store.MaturityAssessment `json:"assessments"`
}

// ScaleResponse is the response for /api/scale.
type ScaleResponse struct {
	Framework  *ScaleFramework   `json:"framework,omitempty"`
	Assessment *ScaleAssessment  `json:"assessment,omitempty"`
	Rollup     *ScaleRollup      `json:"rollup,omitempty"`
	HasData    bool              `json:"hasData"`
	DataNote   string            `json:"dataNote,omitempty"`
}

// ScaleFramework is a simplified view of a SCALE framework for the UI.
type ScaleFramework struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Domains     []ScaleDomain  `json:"domains"`
}

// ScaleDomain represents a SCALE domain with its capabilities and metrics.
type ScaleDomain struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Status       string            `json:"status,omitempty"`
	Capabilities []ScaleCapability `json:"capabilities"`
}

// ScaleCapability represents a SCALE capability with its metrics.
type ScaleCapability struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Metrics     []ScaleMetric `json:"metrics"`
}

// ScaleMetric represents a SCALE metric with its current observation.
type ScaleMetric struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	Aspect          string   `json:"aspect"`
	ConsumptionKind string   `json:"consumptionKind,omitempty"`
	Unit            string   `json:"unit,omitempty"`
	Direction       string   `json:"direction,omitempty"`
	TargetValue     *float64 `json:"targetValue,omitempty"`
	TargetBy        string   `json:"targetBy,omitempty"`
	Owner           string   `json:"owner,omitempty"`
	// Observation data
	Value       *float64 `json:"value,omitempty"`
	Numerator   *int     `json:"numerator,omitempty"`
	Denominator *int     `json:"denominator,omitempty"`
	Attainment  *float64 `json:"attainment,omitempty"`
	Note        string   `json:"note,omitempty"`
}

// ScaleAssessment is a simplified view of a SCALE assessment.
type ScaleAssessment struct {
	Period       string               `json:"period"`
	AsOf         string               `json:"asOf,omitempty"`
	Observations int                  `json:"observations"`
	Narratives   []ScaleNarrative     `json:"narratives,omitempty"`
}

// ScaleNarrative is a journey or outlook narrative.
type ScaleNarrative struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title,omitempty"`
	Body  string `json:"body"`
}

// ScaleRollup contains the computed aspect scores.
type ScaleRollup struct {
	Aspects []ScaleAspectScore `json:"aspects"`
}

// ScaleAspectScore is the score for a single SCALE aspect.
type ScaleAspectScore struct {
	Aspect      string  `json:"aspect"`
	Letter      string  `json:"letter"`
	DisplayName string  `json:"displayName"`
	Score       float64 `json:"score"`
	Eligible    int     `json:"eligible"`
	Observed    int     `json:"observed"`
}

// SpecsResponse is the response for /api/specs using canonical API types.
// Uses apitypes for camelCase JSON serialization.
type SpecsResponse = apitypes.SpecsResponse

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

	mux.HandleFunc("/api/scale", cors(func(w http.ResponseWriter, r *http.Request) {
		resp := buildScaleResponse()
		writeJSON(w, resp)
	}))

	mux.HandleFunc("/api/scale/report", cors(func(w http.ResponseWriter, r *http.Request) {
		ir := buildScaleReportIR()
		if ir == nil {
			http.Error(w, "failed to build report IR", http.StatusInternalServerError)
			return
		}
		writeJSON(w, ir)
	}))

	mux.HandleFunc("/api/leverage", cors(func(w http.ResponseWriter, r *http.Request) {
		graph := buildLeverageGraph()
		writeJSON(w, graph)
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

	// Load repositories
	repos, err := svc.Store.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}
	var apiRepos []APIRepository
	for _, r := range repos {
		apiRepos = append(apiRepos, APIRepository{
			ID:             r.ID,
			Organization:   r.Organization,
			RepositoryName: r.RepositoryName,
			DefaultBranch:  r.DefaultBranch,
			LocalPath:      r.LocalPath,
			GoModule:       r.GoModule,
			Domain:         r.Domain,
			Status:         r.Status,
		})
	}

	return &ExecutionResponse{
		Programs:               apiPrograms,
		Initiatives:            apiInitiatives,
		Phases:                 apiPhases,
		RMIs:                   apiRMIs,
		Repositories:           apiRepos,
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
		ByModel:      make(map[string]*APITokens),
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

		// Aggregate model breakdown from token report
		if init.TokenReport != nil {
			for _, mt := range init.TokenReport.ByModel {
				if resp.ByModel[mt.Model] == nil {
					resp.ByModel[mt.Model] = &APITokens{}
				}
				resp.ByModel[mt.Model].InputTokens += mt.Totals.InputTokens
				resp.ByModel[mt.Model].OutputTokens += mt.Totals.OutputTokens
				resp.ByModel[mt.Model].CacheReadTokens += mt.Totals.CacheReadTokens
				resp.ByModel[mt.Model].CacheCreationTokens += mt.Totals.CacheCreationTokens
				resp.ByModel[mt.Model].TotalTokens += mt.Totals.TotalTokens
				resp.ByModel[mt.Model].CostUSD += mt.Totals.CostUSD
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

	// Compute time-bucketed data from raw events
	resp.ByWeek, resp.ByMonth = computeTimeBuckets(ctx, svc, dataDir)

	return resp, nil
}

// computeTimeBuckets reads raw token events and buckets them by week and month.
func computeTimeBuckets(ctx context.Context, svc *service.Service, dataDir string) ([]APITimeBucket, []APITimeBucket) {
	tokenSource, err := tokens.NewJSONLSource(dataDir)
	if err != nil {
		return nil, nil
	}

	// Query events for the last 18 months (to cover all historical data)
	now := time.Now()
	queryStart := now.AddDate(0, -18, 0)
	result, err := tokenSource.Read(ctx, tokens.Query{
		Period: tokens.Period{Start: queryStart, End: now},
	})
	if err != nil || len(result.Events) == 0 {
		return nil, nil
	}

	// Build attributor for cost calculation
	attr, err := buildSpendAttributor(ctx, svc)
	if err != nil {
		return nil, nil
	}

	// Attribute events
	attributed := attr.AttributeAll(result.Events)

	// Bucket by week and month
	weekBuckets := make(map[string]*timeBucketAccum)
	monthBuckets := make(map[string]*timeBucketAccum)

	for _, ev := range attributed {
		weekKey := weekKey(ev.Event.Timestamp)
		monthKey := monthKey(ev.Event.Timestamp)

		if weekBuckets[weekKey] == nil {
			weekBuckets[weekKey] = newTimeBucketAccum(weekKey, weekStart(ev.Event.Timestamp), weekEnd(ev.Event.Timestamp))
		}
		if monthBuckets[monthKey] == nil {
			monthBuckets[monthKey] = newTimeBucketAccum(monthKey, monthStart(ev.Event.Timestamp), monthEnd(ev.Event.Timestamp))
		}

		weekBuckets[weekKey].add(ev)
		monthBuckets[monthKey].add(ev)
	}

	return sortAndConvertBuckets(weekBuckets), sortAndConvertBuckets(monthBuckets)
}

type timeBucketAccum struct {
	period  string
	start   time.Time
	end     time.Time
	totals  APITokens
	byModel map[string]*APITokens
}

func newTimeBucketAccum(period string, start, end time.Time) *timeBucketAccum {
	return &timeBucketAccum{
		period:  period,
		start:   start,
		end:     end,
		byModel: make(map[string]*APITokens),
	}
}

func (b *timeBucketAccum) add(ev tokens.AttributedEvent) {
	b.totals.InputTokens += ev.Event.InputTokens
	b.totals.OutputTokens += ev.Event.OutputTokens
	b.totals.CacheReadTokens += ev.Event.CacheReadTokens
	b.totals.CacheCreationTokens += ev.Event.CacheCreationTokens
	b.totals.TotalTokens += ev.Event.TotalTokens()
	b.totals.CostUSD += ev.CostUSD

	model := ev.Event.Model
	if model == "" {
		model = "unknown"
	}
	if b.byModel[model] == nil {
		b.byModel[model] = &APITokens{}
	}
	b.byModel[model].InputTokens += ev.Event.InputTokens
	b.byModel[model].OutputTokens += ev.Event.OutputTokens
	b.byModel[model].CacheReadTokens += ev.Event.CacheReadTokens
	b.byModel[model].CacheCreationTokens += ev.Event.CacheCreationTokens
	b.byModel[model].TotalTokens += ev.Event.TotalTokens()
	b.byModel[model].CostUSD += ev.CostUSD
}

func sortAndConvertBuckets(buckets map[string]*timeBucketAccum) []APITimeBucket {
	var result []APITimeBucket
	for _, b := range buckets {
		result = append(result, APITimeBucket{
			Period:  b.period,
			Start:   b.start.Format("2006-01-02"),
			End:     b.end.Format("2006-01-02"),
			Totals:  &b.totals,
			ByModel: b.byModel,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Period < result[j].Period
	})
	return result
}

func weekKey(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

func weekStart(t time.Time) time.Time {
	year, week := t.ISOWeek()
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, t.Location())
	offset := int(time.Monday - jan1.Weekday())
	if offset > 0 {
		offset -= 7
	}
	firstMonday := jan1.AddDate(0, 0, offset)
	return firstMonday.AddDate(0, 0, (week-1)*7)
}

func weekEnd(t time.Time) time.Time {
	return weekStart(t).AddDate(0, 0, 6)
}

func monthKey(t time.Time) string {
	return t.Format("2006-01")
}

func monthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func monthEnd(t time.Time) time.Time {
	return monthStart(t).AddDate(0, 1, -1)
}

// buildSpendAttributor creates an Attributor for cost calculation.
func buildSpendAttributor(ctx context.Context, svc *service.Service) (*tokens.Attributor, error) {
	assignments, err := svc.Store.ListAllAssignments(ctx)
	if err != nil {
		return nil, err
	}

	repos, err := svc.Store.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}

	allRMIs, err := svc.Store.ListAllRMIs(ctx)
	if err != nil {
		return nil, err
	}

	initiatives, err := svc.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}

	rmiLookup := make(map[string]*store.RoadmapItem)
	for _, rmi := range allRMIs {
		rmiLookup[rmi.ID] = rmi
	}

	initLookup := make(map[string]*store.Initiative)
	for _, init := range initiatives {
		initLookup[init.ID] = init
	}

	var assignmentInfos []tokens.AssignmentInfo
	for _, a := range assignments {
		rmi := rmiLookup[a.RMIID]
		if rmi == nil {
			continue
		}
		init := initLookup[rmi.InitiativeID]
		programID := ""
		if init != nil {
			programID = init.ProgramID
		}

		info := tokens.AssignmentInfo{
			ID:          a.ID,
			Worker:      a.Worker,
			RMI:         a.RMIID,
			Phase:       rmi.PhaseID,
			Initiative:  rmi.InitiativeID,
			Program:     programID,
			CreatedAt:   a.CreatedAt,
			CompletedAt: a.CompletedAt,
		}
		assignmentInfos = append(assignmentInfos, info)
	}

	var repoInfos []tokens.RepositoryInfo
	for _, r := range repos {
		if r.LocalPath == "" {
			continue
		}
		repoInfos = append(repoInfos, tokens.RepositoryInfo{
			ID:        r.ID,
			LocalPath: r.LocalPath,
		})
	}

	return tokens.NewAttributor(assignmentInfos, repoInfos), nil
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

	// Convert store types to API types (camelCase JSON)
	apiWorkflows := make([]apitypes.SpecWorkflow, 0, len(workflows))
	for _, w := range workflows {
		apiWorkflows = append(apiWorkflows, storeWorkflowToAPI(w))
	}

	apiResults := make([]apitypes.JudgeResult, 0, len(allResults))
	for _, r := range allResults {
		apiResults = append(apiResults, storeJudgeResultToAPI(r))
	}

	return &SpecsResponse{
		Workflows:    apiWorkflows,
		JudgeResults: apiResults,
	}, nil
}

// storeWorkflowToAPI converts store.SpecWorkflow to apitypes.SpecWorkflow.
func storeWorkflowToAPI(w *store.SpecWorkflow) apitypes.SpecWorkflow {
	return apitypes.SpecWorkflow{
		ID:            w.ID,
		Name:          w.Name,
		Description:   w.Description,
		SpecsRequired: w.SpecsRequired,
		SpecsOptional: w.SpecsOptional,
		InitTypes:     w.InitTypes,
	}
}

// storeJudgeResultToAPI converts store.JudgeResult to apitypes.JudgeResult.
// Both use rubric.Rubric directly, so this is a straightforward copy.
func storeJudgeResultToAPI(r *store.JudgeResult) apitypes.JudgeResult {
	return apitypes.JudgeResult{
		ID:           r.ID,
		InitiativeID: r.InitiativeID,
		SpecPath:     r.SpecPath,
		SpecType:     r.SpecType,
		RubricID:     r.RubricID,
		EvaluatedAt:  r.EvaluatedAt,
		Report:       r.Report,
	}
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
// Supports both legacy format (flat fields) and structured-evaluation rubric.Rubric format.
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

		result := parseEvalFile(data, initiativeID, entry.Name())
		if result != nil {
			results = append(results, result)
		}
	}

	return results
}

// parseEvalFile parses a rubric.Rubric format eval file.
func parseEvalFile(data []byte, initiativeID, filename string) *store.JudgeResult {
	var report rubric.Rubric
	if err := json.Unmarshal(data, &report); err != nil {
		return nil
	}

	// Validate it's a proper rubric report
	if report.ReviewType == "" {
		return nil
	}

	specType := strings.TrimSuffix(filename, ".eval.json")
	return &store.JudgeResult{
		ID:           fmt.Sprintf("eval-%s-%s", initiativeID, specType),
		InitiativeID: initiativeID,
		SpecPath:     report.Metadata.Document,
		SpecType:     specType,
		RubricID:     report.RubricID,
		EvaluatedAt:  report.Metadata.GeneratedAt,
		Report:       &report,
	}
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

// buildScaleResponse builds the SCALE platform adoption response with live dependency data.
func buildScaleResponse() *ScaleResponse {
	// Try to load SCALE catalog from local path first, then fall back to embedded
	var framework *scale.Framework
	var err error

	// Check for local SCALE catalog
	scaleCatalogPath := filepath.Join(os.Getenv("HOME"), "go/src/github.com/ProductBuildersHQ/scale/catalog")
	if _, statErr := os.Stat(scaleCatalogPath); statErr == nil {
		framework, err = scale.LoadFrameworkDir(scaleCatalogPath)
	} else {
		// Fall back to embedded catalog
		framework, err = catalog.Default()
	}

	if err != nil {
		return &ScaleResponse{
			HasData:  false,
			DataNote: fmt.Sprintf("Failed to load SCALE catalog: %v", err),
		}
	}

	// Find the platform domain
	platformDomain := framework.Domain("platform")
	if platformDomain == nil {
		return &ScaleResponse{
			HasData:  false,
			DataNote: "Platform domain not found in SCALE catalog",
		}
	}

	// Scan for live dependency data
	observations := computePlatformObservations()

	// Build the response
	resp := &ScaleResponse{
		HasData: true,
		Framework: &ScaleFramework{
			ID:          framework.ID,
			Name:        framework.Name,
			Description: framework.Description,
		},
	}

	// Convert the platform domain
	domain := ScaleDomain{
		ID:          platformDomain.ID,
		Name:        platformDomain.Name,
		Description: platformDomain.Description,
		Status:      platformDomain.Status,
	}

	// Convert capabilities and metrics
	for _, cap := range platformDomain.Capabilities {
		scaleCap := ScaleCapability{
			ID:          cap.ID,
			Name:        cap.Name,
			Description: cap.Description,
		}

		for _, m := range cap.Metrics {
			metric := ScaleMetric{
				ID:              m.ID,
				Name:            m.Name,
				Description:     m.Description,
				Aspect:          m.Aspect,
				ConsumptionKind: m.ConsumptionKind,
				Unit:            m.Unit,
				Direction:       m.Direction,
				Owner:           m.Owner,
			}
			if m.Target != nil {
				metric.TargetValue = &m.Target.Value
				metric.TargetBy = m.Target.By
			}

			// Add observation if we have one
			if obs, ok := observations[m.ID]; ok {
				metric.Value = &obs.Value
				metric.Numerator = obs.Numerator
				metric.Denominator = obs.Denominator
				metric.Note = obs.Note
				// Compute attainment
				if m.Target != nil && m.Target.Value > 0 {
					att := obs.Value / m.Target.Value
					if att > 1 {
						att = 1
					}
					metric.Attainment = &att
				}
			}

			scaleCap.Metrics = append(scaleCap.Metrics, metric)
		}

		domain.Capabilities = append(domain.Capabilities, scaleCap)
	}

	resp.Framework.Domains = []ScaleDomain{domain}

	// Compute aspect rollup
	resp.Rollup = computeScaleRollup(domain, observations)

	// Add assessment info
	resp.Assessment = &ScaleAssessment{
		Period:       time.Now().Format("2006-Q1"),
		AsOf:         time.Now().Format("2006-01-02"),
		Observations: len(observations),
	}

	return resp
}

// observationData holds computed observation values.
type observationData struct {
	Value       float64
	Numerator   *int
	Denominator *int
	Note        string
}

// computePlatformObservations scans local repos and computes live metrics.
func computePlatformObservations() map[string]observationData {
	obs := make(map[string]observationData)

	// Scan grokify org for dependency data
	grokifyPath := filepath.Join(os.Getenv("HOME"), "go/src/github.com/grokify")
	plexusonePath := filepath.Join(os.Getenv("HOME"), "go/src/github.com/plexusone")

	// Scan grokify repos
	grokifyResults, err := scanner.ScanDirectoryWithProgress(grokifyPath, nil, scanner.ScanOptions{})
	if err != nil {
		grokifyResults = nil
	}

	// Scan plexusone repos
	plexusoneResults, err := scanner.ScanDirectoryWithProgress(plexusonePath, nil, scanner.ScanOptions{})
	if err != nil {
		plexusoneResults = nil
	}

	// Count repos with Go modules
	grokifyTotal := 0
	plexusoneTotal := 0
	for _, r := range grokifyResults {
		if r.HasGoMod {
			grokifyTotal++
		}
	}
	for _, r := range plexusoneResults {
		if r.HasGoMod {
			plexusoneTotal++
		}
	}

	// Compute mogo adoption (grokify org)
	mogoCount := 0
	for _, r := range grokifyResults {
		if r.HasDependency("github.com/grokify/mogo") {
			mogoCount++
		}
	}
	if grokifyTotal > 0 {
		pct := float64(mogoCount) / float64(grokifyTotal) * 100
		num, denom := mogoCount, grokifyTotal
		obs["platform.consumption.mogo-adoption"] = observationData{
			Value:       pct,
			Numerator:   &num,
			Denominator: &denom,
			Note:        fmt.Sprintf("%d repos in grokify depend on mogo", mogoCount),
		}
	}

	// Compute gohttp adoption (grokify org)
	gohttpCount := 0
	for _, r := range grokifyResults {
		if r.HasDependency("github.com/grokify/gohttp") {
			gohttpCount++
		}
	}
	if grokifyTotal > 0 {
		pct := float64(gohttpCount) / float64(grokifyTotal) * 100
		num, denom := gohttpCount, grokifyTotal
		obs["platform.consumption.gohttp-adoption"] = observationData{
			Value:       pct,
			Numerator:   &num,
			Denominator: &denom,
			Note:        fmt.Sprintf("%d repos in grokify depend on gohttp", gohttpCount),
		}
	}

	// Compute omniagent adoption (plexusone org)
	omniagentCount := 0
	for _, r := range plexusoneResults {
		if r.HasDependency("github.com/plexusone/omniagent") {
			omniagentCount++
		}
	}
	obs["platform.consumption.omniagent-adoption"] = observationData{
		Value: float64(omniagentCount),
		Note:  fmt.Sprintf("%d repos in plexusone depend on omniagent", omniagentCount),
	}

	// Compute gogit adoption
	gogitCount := 0
	for _, r := range grokifyResults {
		if r.HasDependency("github.com/grokify/gogit") {
			gogitCount++
		}
	}
	obs["platform.consumption.gogit-adoption"] = observationData{
		Value: float64(gogitCount),
		Note:  fmt.Sprintf("%d repos depend on gogit", gogitCount),
	}

	// Compute dependency health metrics
	noReplaceCount := 0
	moduleMatchCount := 0
	for _, r := range grokifyResults {
		if r.HasGoMod && !r.HasReplaceDirectives {
			noReplaceCount++
		}
		if r.HasGoMod && !r.HasModuleMismatch {
			moduleMatchCount++
		}
	}
	if grokifyTotal > 0 {
		pct := float64(noReplaceCount) / float64(grokifyTotal) * 100
		num, denom := noReplaceCount, grokifyTotal
		obs["platform.consumption.repos-no-replace"] = observationData{
			Value:       pct,
			Numerator:   &num,
			Denominator: &denom,
			Note:        fmt.Sprintf("%d/%d repos have no replace directives", noReplaceCount, grokifyTotal),
		}

		pct = float64(moduleMatchCount) / float64(grokifyTotal) * 100
		num, denom = moduleMatchCount, grokifyTotal
		obs["platform.consumption.repos-module-match"] = observationData{
			Value:       pct,
			Numerator:   &num,
			Denominator: &denom,
			Note:        fmt.Sprintf("%d/%d repos have matching module names", moduleMatchCount, grokifyTotal),
		}
	}

	// Leverage metrics - compute internal dependency ratio
	// Internal platform prefixes we care about
	internalPrefixes := []string{
		"github.com/grokify/",
		"github.com/plexusone/",
		"github.com/ProductBuildersHQ/",
	}

	totalInternalDeps := 0
	totalExternalDeps := 0
	reposWithDeps := 0

	// Count across all scanned repos
	allResults := append(grokifyResults, plexusoneResults...)
	for _, r := range allResults {
		if !r.HasGoMod || len(r.Dependencies) == 0 {
			continue
		}
		reposWithDeps++
		for _, dep := range r.Dependencies {
			isInternal := false
			for _, prefix := range internalPrefixes {
				if strings.HasPrefix(dep, prefix) {
					isInternal = true
					break
				}
			}
			if isInternal {
				totalInternalDeps++
			} else {
				totalExternalDeps++
			}
		}
	}

	if totalInternalDeps+totalExternalDeps > 0 {
		ratio := float64(totalInternalDeps) / float64(totalInternalDeps+totalExternalDeps) * 100
		obs["platform.leverage.shared-deps-ratio"] = observationData{
			Value:       ratio,
			Numerator:   &totalInternalDeps,
			Denominator: func() *int { t := totalInternalDeps + totalExternalDeps; return &t }(),
			Note:        fmt.Sprintf("%d internal deps out of %d total across %d repos (%.1f%%)", totalInternalDeps, totalInternalDeps+totalExternalDeps, reposWithDeps, ratio),
		}
	}

	// Standards metrics - these are binary (library exists or not)
	obs["platform.standards.mogo-exists"] = observationData{Value: 1, Note: "mogo library exists"}
	obs["platform.standards.omniagent-exists"] = observationData{Value: 1, Note: "omniagent framework exists"}
	obs["platform.standards.visionstudio-exists"] = observationData{Value: 1, Note: "VisionStudio exists"}
	obs["platform.standards.gogit-exists"] = observationData{Value: 1, Note: "gogit library exists"}
	obs["platform.standards.scale-exists"] = observationData{Value: 1, Note: "SCALE framework exists"}

	return obs
}

// computeScaleRollup computes the aspect-level rollup scores.
func computeScaleRollup(domain ScaleDomain, observations map[string]observationData) *ScaleRollup {
	// Group metrics by aspect
	aspectMetrics := make(map[string][]ScaleMetric)
	for _, cap := range domain.Capabilities {
		for _, m := range cap.Metrics {
			aspectMetrics[m.Aspect] = append(aspectMetrics[m.Aspect], m)
		}
	}

	rollup := &ScaleRollup{}

	for _, aspect := range scale.AllAspects() {
		metrics := aspectMetrics[aspect]
		eligible := 0
		observed := 0
		totalAttainment := 0.0

		for _, m := range metrics {
			if m.TargetValue != nil && m.Owner != "" {
				eligible++
				if obs, ok := observations[m.ID]; ok {
					observed++
					att := obs.Value / *m.TargetValue
					if att > 1 {
						att = 1
					}
					totalAttainment += att
				}
			}
		}

		score := 0.0
		if observed > 0 {
			score = totalAttainment / float64(observed) * 100
		}

		rollup.Aspects = append(rollup.Aspects, ScaleAspectScore{
			Aspect:      aspect,
			Letter:      scale.AspectLetter(aspect),
			DisplayName: scale.AspectDisplayName(aspect),
			Score:       score,
			Eligible:    eligible,
			Observed:    observed,
		})
	}

	return rollup
}

// buildLeverageGraph builds a dependency leverage graph from local repos.
func buildLeverageGraph() *ir.LeverageGraph {
	adapter := ir.DefaultGoAdapter()

	dirs := []string{
		filepath.Join(os.Getenv("HOME"), "go/src/github.com/grokify"),
		filepath.Join(os.Getenv("HOME"), "go/src/github.com/plexusone"),
		filepath.Join(os.Getenv("HOME"), "go/src/github.com/ProductBuildersHQ"),
	}

	graph, err := adapter.ScanAndBuild(dirs)
	if err != nil {
		return &ir.LeverageGraph{
			Ecosystem: "go",
			Scope:     "error",
		}
	}
	return graph
}

// buildScaleReportIR builds a SCALE ReportIR with live dependency metrics.
func buildScaleReportIR() *scalereport.ReportIR {
	// Load framework
	scaleCatalogPath := filepath.Join(os.Getenv("HOME"), "go/src/github.com/ProductBuildersHQ/scale/catalog")
	var framework *scale.Framework
	var err error
	if _, statErr := os.Stat(scaleCatalogPath); statErr == nil {
		framework, err = scale.LoadFrameworkDir(scaleCatalogPath)
	} else {
		framework, err = catalog.Default()
	}
	if err != nil || framework == nil {
		return nil
	}

	// Build assessment from live observations
	observations := computePlatformObservations()

	// Create assessment
	assessment := &scale.Assessment{
		FrameworkID: "scale",
		Period:      time.Now().Format("2006-Q1"),
		AsOf:        time.Now().Format("2006-01-02"),
	}
	for metricID, obs := range observations {
		assessment.Observations = append(assessment.Observations, scale.Observation{
			MetricID:    metricID,
			Value:       obs.Value,
			Numerator:   obs.Numerator,
			Denominator: obs.Denominator,
			Note:        obs.Note,
		})
	}

	// Build IR
	reportIR, err := scalereport.BuildIR(framework, assessment, &scalereport.Options{
		GeneratedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return nil
	}

	return reportIR
}
