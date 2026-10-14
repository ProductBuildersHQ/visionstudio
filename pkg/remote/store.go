package remote

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// The methods in this file and work.go are backed by VisionStudio Cloud
// tenant API endpoints. Request bodies are the store JSON types (or, for the
// two create endpoints that predate them, the API's create request shapes);
// the server decodes strictly and rejects unknown fields. Responses are
// store types; a write replaces the caller's value with the server's copy
// (so server-stamped timestamps and defaults are visible).

// --- generic helpers ---

// write sends v and replaces it with the server's response.
func write[T any](ctx context.Context, s *Store, method, path string, v *T) error {
	var out T
	if err := s.do(ctx, method, path, nil, v, &out); err != nil {
		return err
	}
	*v = out
	return nil
}

// get fetches one entity.
func get[T any](ctx context.Context, s *Store, path string) (*T, error) {
	var out T
	if err := s.do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// list fetches a wrapped list ({"<key>": [...]}).
func list[T any](ctx context.Context, s *Store, path string, q url.Values, key string) ([]*T, error) {
	var out map[string][]*T
	if err := s.do(ctx, http.MethodGet, path, q, nil, &out); err != nil {
		return nil, err
	}
	return out[key], nil
}

// requireID reports a call with no ID without a round trip (an empty path
// segment would otherwise hit a list route or a bare 404).
func requireID(method, collection, kind, id string) error {
	if id != "" {
		return nil
	}
	return &APIError{Method: method, Path: collection + "/", StatusCode: http.StatusNotFound, Code: CodeNotFound, Message: kind + " ID is required"}
}

// --- ProgramStore ---

type createProgramRequest struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Description  string `json:"description,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`
}

// CreateProgram POSTs the program; prog is replaced with the server's copy.
func (s *Store) CreateProgram(ctx context.Context, prog *store.Program) error {
	if prog == nil {
		return fmt.Errorf("remote: CreateProgram: nil program")
	}
	req := createProgramRequest{ID: prog.ID, Name: prog.Name, Organization: prog.Organization, Description: prog.Description, Hidden: prog.Hidden}
	var out store.Program
	if err := s.do(ctx, http.MethodPost, "/programs", nil, req, &out); err != nil {
		return err
	}
	*prog = out
	return nil
}

// GetProgram fetches one program.
func (s *Store) GetProgram(ctx context.Context, id string) (*store.Program, error) {
	if err := requireID(http.MethodGet, "/programs", "program", id); err != nil {
		return nil, err
	}
	return get[store.Program](ctx, s, idPath("/programs", id))
}

// ListPrograms lists every program in the tenant.
func (s *Store) ListPrograms(ctx context.Context) ([]*store.Program, error) {
	return list[store.Program](ctx, s, "/programs", nil, "programs")
}

// UpdateProgram replaces the program (PUT).
func (s *Store) UpdateProgram(ctx context.Context, prog *store.Program) error {
	if prog == nil {
		return fmt.Errorf("remote: UpdateProgram: nil program")
	}
	if err := requireID(http.MethodPut, "/programs", "program", prog.ID); err != nil {
		return err
	}
	return write(ctx, s, http.MethodPut, idPath("/programs", prog.ID), prog)
}

// --- InitiativeStore ---

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

// CreateInitiative POSTs the initiative. The cloud create endpoint creates
// it in "proposed" status and carries only the core fields; when init sets
// anything else (another status, home_repo, program_id, specs, lifecycle
// timestamps, ...) the full initiative is then written with a PUT. The two
// requests are not atomic: if the PUT fails, the initiative exists with
// only its core fields and the error is returned. On success init is
// replaced with the server's copy.
func (s *Store) CreateInitiative(ctx context.Context, init *store.Initiative) error {
	if init == nil {
		return fmt.Errorf("remote: CreateInitiative: nil initiative")
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
	if !initiativeHasExtraFields(init) {
		*init = out
		return nil
	}
	full := *init
	if full.Status == "" {
		full.Status = out.Status
	}
	if full.InitType == "" {
		full.InitType = out.InitType
	}
	if full.Visibility == "" {
		full.Visibility = out.Visibility
	}
	if full.CreatedAt.IsZero() {
		full.CreatedAt = out.CreatedAt
	}
	if full.UpdatedAt.IsZero() {
		full.UpdatedAt = out.UpdatedAt
	}
	if err := s.UpdateInitiative(ctx, &full); err != nil {
		return fmt.Errorf("remote: initiative %s was created, but setting its remaining fields failed: %w", init.ID, err)
	}
	*init = full
	return nil
}

// initiativeHasExtraFields reports fields the create endpoint cannot carry.
func initiativeHasExtraFields(in *store.Initiative) bool {
	return (in.Status != "" && in.Status != statusProposed) ||
		in.HomeRepo != "" || in.Workspace != "" || in.ProgramID != "" || in.Hidden ||
		in.Visibility != "" || len(in.Specs) > 0 ||
		in.PlannedAt != nil || in.ExecutingAt != nil || in.DeliveryCompleteAt != nil ||
		in.ReleasedAt != nil || in.ClosedAt != nil
}

// GetInitiative fetches one initiative.
func (s *Store) GetInitiative(ctx context.Context, id string) (*store.Initiative, error) {
	if err := requireID(http.MethodGet, "/initiatives", "initiative", id); err != nil {
		return nil, err
	}
	return get[store.Initiative](ctx, s, idPath("/initiatives", id))
}

// ListInitiatives lists every initiative in the tenant.
func (s *Store) ListInitiatives(ctx context.Context) ([]*store.Initiative, error) {
	return list[store.Initiative](ctx, s, "/initiatives", nil, "initiatives")
}

// UpdateInitiative replaces the initiative (PUT). Like the local store it
// does not validate a status change.
func (s *Store) UpdateInitiative(ctx context.Context, init *store.Initiative) error {
	if init == nil {
		return fmt.Errorf("remote: UpdateInitiative: nil initiative")
	}
	if err := requireID(http.MethodPut, "/initiatives", "initiative", init.ID); err != nil {
		return err
	}
	return write(ctx, s, http.MethodPut, idPath("/initiatives", init.ID), init)
}

// --- PhaseStore ---

// CreatePhase POSTs the phase; phase is replaced with the server's copy.
func (s *Store) CreatePhase(ctx context.Context, phase *store.Phase) error {
	if phase == nil {
		return fmt.Errorf("remote: CreatePhase: nil phase")
	}
	return write(ctx, s, http.MethodPost, "/phases", phase)
}

// ListPhases lists the phases of one initiative.
func (s *Store) ListPhases(ctx context.Context, initiativeID string) ([]*store.Phase, error) {
	if initiativeID == "" {
		// The API treats an empty filter as "all phases"; phases always
		// belong to an initiative, so the store contract yields none.
		return nil, nil
	}
	return list[store.Phase](ctx, s, "/phases", url.Values{"initiative": {initiativeID}}, "phases")
}

// DeletePhase deletes a phase. The server refuses (409) a phase that still
// has member RMIs.
func (s *Store) DeletePhase(ctx context.Context, id string) error {
	if err := requireID(http.MethodDelete, "/phases", "phase", id); err != nil {
		return err
	}
	return s.do(ctx, http.MethodDelete, idPath("/phases", id), nil, nil, nil)
}

// --- RMIStore ---

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

// CreateRMI POSTs the RMI. The cloud create endpoint creates it in
// "proposed" status; when rmi sets a different status, an origin other than
// "spec", a context_spec or completed_at, the full RMI is then written with
// a PUT. The two requests are not atomic: if the PUT fails, the RMI exists
// with only its create fields and the error is returned. On success rmi is
// replaced with the server's copy.
func (s *Store) CreateRMI(ctx context.Context, rmi *store.RoadmapItem) error {
	if rmi == nil {
		return fmt.Errorf("remote: CreateRMI: nil RMI")
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
	if !rmiHasExtraFields(rmi) {
		*rmi = out
		return nil
	}
	full := *rmi
	if full.Status == "" {
		full.Status = out.Status
	}
	if full.Origin == "" {
		full.Origin = out.Origin
	}
	if full.CreatedAt.IsZero() {
		full.CreatedAt = out.CreatedAt
	}
	if full.UpdatedAt.IsZero() {
		full.UpdatedAt = out.UpdatedAt
	}
	if err := s.UpdateRMI(ctx, &full); err != nil {
		return fmt.Errorf("remote: RMI %s was created, but setting its remaining fields failed: %w", rmi.ID, err)
	}
	*rmi = full
	return nil
}

// rmiHasExtraFields reports fields the create endpoint cannot carry.
func rmiHasExtraFields(r *store.RoadmapItem) bool {
	return (r.Status != "" && r.Status != statusProposed) ||
		(r.Origin != "" && r.Origin != "spec") ||
		r.ContextSpec != nil || r.CompletedAt != nil
}

// GetRMI fetches one RMI.
func (s *Store) GetRMI(ctx context.Context, id string) (*store.RoadmapItem, error) {
	if err := requireID(http.MethodGet, "/rmis", "RMI", id); err != nil {
		return nil, err
	}
	return get[store.RoadmapItem](ctx, s, idPath("/rmis", id))
}

// ListRMIs lists the RMIs of one initiative.
func (s *Store) ListRMIs(ctx context.Context, initiativeID string) ([]*store.RoadmapItem, error) {
	if initiativeID == "" {
		// The API treats an empty filter as "all RMIs"; the store
		// contract is "RMIs of this initiative", so match locally.
		return s.filterAllRMIs(ctx, func(r *store.RoadmapItem) bool { return r.InitiativeID == "" })
	}
	return s.listRMIs(ctx, url.Values{"initiative": {initiativeID}})
}

// ListAllRMIs lists every RMI in the tenant.
func (s *Store) ListAllRMIs(ctx context.Context) ([]*store.RoadmapItem, error) {
	return s.listRMIs(ctx, nil)
}

// ListRMIsByRepo lists the RMIs of one repository.
func (s *Store) ListRMIsByRepo(ctx context.Context, repoID string) ([]*store.RoadmapItem, error) {
	if repoID == "" {
		// An empty filter means "no filter" to the API; match locally.
		return s.filterAllRMIs(ctx, func(r *store.RoadmapItem) bool { return r.RepositoryID == "" })
	}
	return s.listRMIs(ctx, url.Values{"repository": {repoID}})
}

// ListRMIsByStatus lists the RMIs with one status.
func (s *Store) ListRMIsByStatus(ctx context.Context, status string) ([]*store.RoadmapItem, error) {
	if status == "" {
		return s.filterAllRMIs(ctx, func(r *store.RoadmapItem) bool { return r.Status == "" })
	}
	return s.listRMIs(ctx, url.Values{"status": {status}})
}

func (s *Store) filterAllRMIs(ctx context.Context, keep func(*store.RoadmapItem) bool) ([]*store.RoadmapItem, error) {
	all, err := s.ListAllRMIs(ctx)
	if err != nil {
		return nil, err
	}
	var out []*store.RoadmapItem
	for _, r := range all {
		if keep(r) {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *Store) listRMIs(ctx context.Context, q url.Values) ([]*store.RoadmapItem, error) {
	return list[store.RoadmapItem](ctx, s, "/rmis", q, "rmis")
}

// UpdateRMI replaces the RMI (PUT). Like the local store it does not
// validate a status change.
func (s *Store) UpdateRMI(ctx context.Context, rmi *store.RoadmapItem) error {
	if rmi == nil {
		return fmt.Errorf("remote: UpdateRMI: nil RMI")
	}
	if err := requireID(http.MethodPut, "/rmis", "RMI", rmi.ID); err != nil {
		return err
	}
	return write(ctx, s, http.MethodPut, idPath("/rmis", rmi.ID), rmi)
}

// CreateDependency creates an RMI dependency edge (idempotent server-side;
// the relationship defaults to "requires").
func (s *Store) CreateDependency(ctx context.Context, dep *store.RMIDependency) error {
	if dep == nil {
		return fmt.Errorf("remote: CreateDependency: nil dependency")
	}
	return write(ctx, s, http.MethodPost, "/dependencies", dep)
}

// ListDependencies lists the dependency edges of one RMI.
func (s *Store) ListDependencies(ctx context.Context, rmiID string) ([]*store.RMIDependency, error) {
	if err := requireID(http.MethodGet, "/rmis", "RMI", rmiID); err != nil {
		return nil, err
	}
	return list[store.RMIDependency](ctx, s, idPath("/rmis", rmiID)+"/dependencies", nil, "dependencies")
}

// DeleteDependency is not served by the VisionStudio Cloud API yet.
func (s *Store) DeleteDependency(_ context.Context, _, _ string) (int, error) {
	return 0, &NotSupportedError{Op: "DeleteDependency"}
}

// ListAllDependencies lists every RMI dependency edge in the tenant.
func (s *Store) ListAllDependencies(ctx context.Context) ([]*store.RMIDependency, error) {
	return list[store.RMIDependency](ctx, s, "/dependencies", nil, "dependencies")
}
