package remote

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// The methods in this file are backed by VisionStudio Cloud tenant API
// endpoints. Request bodies mirror the API's create request shapes exactly
// (the server rejects unknown fields); response bodies are store types.

// statusProposed is the status the cloud API assigns on create.
const statusProposed = "proposed"

type createInitiativeRequest struct {
	ID           string `json:"id"`
	Organization string `json:"organization"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Priority     string `json:"priority,omitempty"`
	InitType     string `json:"init_type,omitempty"`
	WorkflowID   string `json:"workflow_id,omitempty"`
}

type createRMIRequest struct {
	ID                 string   `json:"id"`
	RepositoryID       string   `json:"repository_id"`
	InitiativeID       string   `json:"initiative_id,omitempty"`
	PhaseID            string   `json:"phase_id,omitempty"`
	Title              string   `json:"title"`
	Description        string   `json:"description,omitempty"`
	ItemType           string   `json:"item_type"`
	Priority           string   `json:"priority,omitempty"`
	Required           bool     `json:"required"`
	SequenceNumber     int      `json:"sequence_number,omitempty"`
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
}

// --- InitiativeStore ---

// CreateInitiative POSTs the initiative. The cloud API creates it in
// "proposed" status through the same service layer, so fields the create
// endpoint cannot carry are rejected rather than silently dropped. On
// success init is replaced with the server's copy (timestamps included).
func (s *Store) CreateInitiative(ctx context.Context, init *store.Initiative) error {
	if init == nil {
		return fmt.Errorf("remote: CreateInitiative: nil initiative")
	}
	if extra := initiativeExtraFields(init); len(extra) > 0 {
		return &NotSupportedError{
			Op:     "CreateInitiative",
			Reason: "the cloud create endpoint cannot set " + strings.Join(extra, ", ") + " and initiative updates have no endpoint yet",
		}
	}
	req := createInitiativeRequest{
		ID:           init.ID,
		Organization: init.Organization,
		Title:        init.Title,
		Description:  init.Description,
		Priority:     init.Priority,
		InitType:     init.InitType,
		WorkflowID:   init.WorkflowID,
	}
	var out store.Initiative
	if err := s.do(ctx, http.MethodPost, "/initiatives", nil, req, &out); err != nil {
		return err
	}
	*init = out
	return nil
}

func initiativeExtraFields(in *store.Initiative) []string {
	var f []string
	if in.Status != "" && in.Status != statusProposed {
		f = append(f, "status "+in.Status)
	}
	if in.HomeRepo != "" {
		f = append(f, "home_repo")
	}
	if in.Workspace != "" {
		f = append(f, "workspace")
	}
	if in.ProgramID != "" {
		f = append(f, "program_id")
	}
	if in.Hidden {
		f = append(f, "hidden")
	}
	if in.Visibility != "" {
		f = append(f, "visibility")
	}
	if len(in.Specs) > 0 {
		f = append(f, "specs")
	}
	if in.PlannedAt != nil || in.ExecutingAt != nil || in.DeliveryCompleteAt != nil || in.ReleasedAt != nil || in.ClosedAt != nil {
		f = append(f, "lifecycle timestamps")
	}
	return f
}

// GetInitiative fetches one initiative.
func (s *Store) GetInitiative(ctx context.Context, id string) (*store.Initiative, error) {
	var out store.Initiative
	if err := s.do(ctx, http.MethodGet, "/initiatives/"+url.PathEscape(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListInitiatives lists every initiative in the tenant.
func (s *Store) ListInitiatives(ctx context.Context) ([]*store.Initiative, error) {
	var out struct {
		Initiatives []*store.Initiative `json:"initiatives"`
	}
	if err := s.do(ctx, http.MethodGet, "/initiatives", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Initiatives, nil
}

// --- RMIStore ---

// CreateRMI POSTs the RMI. The cloud API creates it in "proposed" status;
// fields the create endpoint cannot carry are rejected rather than dropped.
// On success rmi is replaced with the server's copy.
func (s *Store) CreateRMI(ctx context.Context, rmi *store.RoadmapItem) error {
	if rmi == nil {
		return fmt.Errorf("remote: CreateRMI: nil RMI")
	}
	if extra := rmiExtraFields(rmi); len(extra) > 0 {
		return &NotSupportedError{
			Op:     "CreateRMI",
			Reason: "the cloud create endpoint cannot set " + strings.Join(extra, ", ") + " and RMI updates have no endpoint yet",
		}
	}
	req := createRMIRequest{
		ID:                 rmi.ID,
		RepositoryID:       rmi.RepositoryID,
		InitiativeID:       rmi.InitiativeID,
		PhaseID:            rmi.PhaseID,
		Title:              rmi.Title,
		Description:        rmi.Description,
		ItemType:           rmi.ItemType,
		Priority:           rmi.Priority,
		Required:           rmi.Required,
		SequenceNumber:     rmi.SequenceNumber,
		AcceptanceCriteria: rmi.AcceptanceCriteria,
	}
	var out store.RoadmapItem
	if err := s.do(ctx, http.MethodPost, "/rmis", nil, req, &out); err != nil {
		return err
	}
	*rmi = out
	return nil
}

func rmiExtraFields(r *store.RoadmapItem) []string {
	var f []string
	if r.Status != "" && r.Status != statusProposed {
		f = append(f, "status "+r.Status)
	}
	if r.Origin != "" && r.Origin != "spec" {
		f = append(f, "origin "+r.Origin)
	}
	if r.ContextSpec != nil {
		f = append(f, "context_spec")
	}
	if r.CompletedAt != nil {
		f = append(f, "completed_at")
	}
	return f
}

// GetRMI fetches one RMI.
func (s *Store) GetRMI(ctx context.Context, id string) (*store.RoadmapItem, error) {
	var out store.RoadmapItem
	if err := s.do(ctx, http.MethodGet, "/rmis/"+url.PathEscape(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRMIs lists the RMIs of one initiative.
func (s *Store) ListRMIs(ctx context.Context, initiativeID string) ([]*store.RoadmapItem, error) {
	if initiativeID == "" {
		// The API treats an empty filter as "all RMIs"; the store
		// contract is "RMIs of this initiative", so match locally.
		all, err := s.ListAllRMIs(ctx)
		if err != nil {
			return nil, err
		}
		return filterRMIs(all, func(r *store.RoadmapItem) bool { return r.InitiativeID == "" }), nil
	}
	return s.listRMIs(ctx, url.Values{"initiative": {initiativeID}})
}

// ListAllRMIs lists every RMI in the tenant.
func (s *Store) ListAllRMIs(ctx context.Context) ([]*store.RoadmapItem, error) {
	return s.listRMIs(ctx, nil)
}

// ListRMIsByRepo filters the tenant's RMIs by repository client-side (the
// API has no repository filter yet).
func (s *Store) ListRMIsByRepo(ctx context.Context, repoID string) ([]*store.RoadmapItem, error) {
	all, err := s.ListAllRMIs(ctx)
	if err != nil {
		return nil, err
	}
	return filterRMIs(all, func(r *store.RoadmapItem) bool { return r.RepositoryID == repoID }), nil
}

// ListRMIsByStatus filters the tenant's RMIs by status client-side (the API
// has no status filter yet).
func (s *Store) ListRMIsByStatus(ctx context.Context, status string) ([]*store.RoadmapItem, error) {
	all, err := s.ListAllRMIs(ctx)
	if err != nil {
		return nil, err
	}
	return filterRMIs(all, func(r *store.RoadmapItem) bool { return r.Status == status }), nil
}

func (s *Store) listRMIs(ctx context.Context, q url.Values) ([]*store.RoadmapItem, error) {
	var out struct {
		RMIs []*store.RoadmapItem `json:"rmis"`
	}
	if err := s.do(ctx, http.MethodGet, "/rmis", q, nil, &out); err != nil {
		return nil, err
	}
	return out.RMIs, nil
}

func filterRMIs(in []*store.RoadmapItem, keep func(*store.RoadmapItem) bool) []*store.RoadmapItem {
	var out []*store.RoadmapItem
	for _, r := range in {
		if keep(r) {
			out = append(out, r)
		}
	}
	return out
}

// --- PhaseStore (read-only) ---

// ListPhases lists the phases of one initiative.
func (s *Store) ListPhases(ctx context.Context, initiativeID string) ([]*store.Phase, error) {
	if initiativeID == "" {
		// The API treats an empty filter as "all phases"; phases always
		// belong to an initiative, so the store contract yields none.
		return nil, nil
	}
	var out struct {
		Phases []*store.Phase `json:"phases"`
	}
	if err := s.do(ctx, http.MethodGet, "/phases", url.Values{"initiative": {initiativeID}}, nil, &out); err != nil {
		return nil, err
	}
	return out.Phases, nil
}

// --- ProgramStore (read-only) ---

// ListPrograms lists every program in the tenant.
func (s *Store) ListPrograms(ctx context.Context) ([]*store.Program, error) {
	var out struct {
		Programs []*store.Program `json:"programs"`
	}
	if err := s.do(ctx, http.MethodGet, "/programs", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Programs, nil
}

// GetProgram finds one program in the tenant's program list (the API has
// no single-program endpoint yet).
func (s *Store) GetProgram(ctx context.Context, id string) (*store.Program, error) {
	progs, err := s.ListPrograms(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range progs {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, &APIError{Method: http.MethodGet, Path: "/programs", StatusCode: http.StatusNotFound, Message: fmt.Sprintf("program %s not found", id)}
}
