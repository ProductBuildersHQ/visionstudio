package remote

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// --- AssignmentStore ---
//
// The service layer (ClaimRMI, RenewLease, ReleaseWork, CompleteWork) runs
// on the caller's side and persists through these methods. The server
// enforces what a shared database needs: creating an active assignment while
// another live lease exists, or updating an assignment that is no longer
// active, fails with a 409 LEASE_CONFLICT (errors.Is ErrLeaseConflict and
// ErrConflict).

// CreateAssignment POSTs the assignment (a claim when it is active); a is
// replaced with the server's copy.
func (s *Store) CreateAssignment(ctx context.Context, a *store.Assignment) error {
	if a == nil {
		return fmt.Errorf("remote: CreateAssignment: nil assignment")
	}
	return write(ctx, s, http.MethodPost, "/assignments", a)
}

// GetAssignment fetches one assignment.
func (s *Store) GetAssignment(ctx context.Context, id string) (*store.Assignment, error) {
	if err := requireID(http.MethodGet, "/assignments", "assignment", id); err != nil {
		return nil, err
	}
	return get[store.Assignment](ctx, s, idPath("/assignments", id))
}

// GetActiveAssignment returns the RMI's active assignment, or nil (and no
// error) when it has none.
func (s *Store) GetActiveAssignment(ctx context.Context, rmiID string) (*store.Assignment, error) {
	if rmiID == "" {
		return nil, nil
	}
	var out struct {
		Assignment *store.Assignment `json:"assignment"`
	}
	if err := s.do(ctx, http.MethodGet, idPath("/rmis", rmiID)+"/active-assignment", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Assignment, nil
}

// ListActiveAssignments lists every active assignment in the tenant.
func (s *Store) ListActiveAssignments(ctx context.Context) ([]*store.Assignment, error) {
	return list[store.Assignment](ctx, s, "/assignments", url.Values{"status": {"active"}}, "assignments")
}

// ListAllAssignments lists every assignment in the tenant.
func (s *Store) ListAllAssignments(ctx context.Context) ([]*store.Assignment, error) {
	return list[store.Assignment](ctx, s, "/assignments", nil, "assignments")
}

// UpdateAssignment replaces the assignment (PUT): a lease renewal, handoff,
// release, completion or expiry. a is replaced with the server's copy.
func (s *Store) UpdateAssignment(ctx context.Context, a *store.Assignment) error {
	if a == nil {
		return fmt.Errorf("remote: UpdateAssignment: nil assignment")
	}
	if err := requireID(http.MethodPut, "/assignments", "assignment", a.ID); err != nil {
		return err
	}
	return write(ctx, s, http.MethodPut, idPath("/assignments", a.ID), a)
}

// --- EvidenceStore ---

// CreateEvidence POSTs delivery evidence; ev is replaced with the server's
// copy.
func (s *Store) CreateEvidence(ctx context.Context, ev *store.DeliveryEvidence) error {
	if ev == nil {
		return fmt.Errorf("remote: CreateEvidence: nil evidence")
	}
	return write(ctx, s, http.MethodPost, "/evidence", ev)
}

// ListEvidenceByRMI lists the evidence attached to one RMI.
func (s *Store) ListEvidenceByRMI(ctx context.Context, rmiID string) ([]*store.DeliveryEvidence, error) {
	if rmiID == "" {
		return nil, nil
	}
	return s.listEvidence(ctx, url.Values{"rmi": {rmiID}})
}

// ListEvidenceByInitiative lists the evidence attached to an initiative's
// RMIs.
func (s *Store) ListEvidenceByInitiative(ctx context.Context, initiativeID string) ([]*store.DeliveryEvidence, error) {
	if initiativeID == "" {
		return nil, nil
	}
	return s.listEvidence(ctx, url.Values{"initiative": {initiativeID}})
}

// ListAllEvidence lists every evidence record in the tenant.
func (s *Store) ListAllEvidence(ctx context.Context) ([]*store.DeliveryEvidence, error) {
	return s.listEvidence(ctx, nil)
}

func (s *Store) listEvidence(ctx context.Context, q url.Values) ([]*store.DeliveryEvidence, error) {
	return list[store.DeliveryEvidence](ctx, s, "/evidence", q, "evidence")
}

// --- ReleaseStore (read-only) ---

// GetRelease fetches one release ("<repository-id>@<tag>").
func (s *Store) GetRelease(ctx context.Context, id string) (*store.Release, error) {
	if err := requireID(http.MethodGet, "/releases", "release", id); err != nil {
		return nil, err
	}
	return get[store.Release](ctx, s, idPath("/releases", id))
}

// ListReleases lists every release in the tenant.
func (s *Store) ListReleases(ctx context.Context) ([]*store.Release, error) {
	return s.listReleases(ctx, nil)
}

// ListReleasesByRepo lists one repository's releases.
func (s *Store) ListReleasesByRepo(ctx context.Context, repoID string) ([]*store.Release, error) {
	if repoID == "" {
		return nil, nil
	}
	return s.listReleases(ctx, url.Values{"repository": {repoID}})
}

// ListReleasesByInitiative lists the releases that shipped an initiative.
func (s *Store) ListReleasesByInitiative(ctx context.Context, initiativeID string) ([]*store.Release, error) {
	if initiativeID == "" {
		return nil, nil
	}
	return s.listReleases(ctx, url.Values{"initiative": {initiativeID}})
}

func (s *Store) listReleases(ctx context.Context, q url.Values) ([]*store.Release, error) {
	return list[store.Release](ctx, s, "/releases", q, "releases")
}

// --- RepositoryStore (read-only) ---

// GetRepository fetches one repository ("github.com/org/name").
func (s *Store) GetRepository(ctx context.Context, id string) (*store.Repository, error) {
	if err := requireID(http.MethodGet, "/repositories", "repository", id); err != nil {
		return nil, err
	}
	return get[store.Repository](ctx, s, idPath("/repositories", id))
}

// ListRepositories lists every repository registered in the tenant.
func (s *Store) ListRepositories(ctx context.Context) ([]*store.Repository, error) {
	return list[store.Repository](ctx, s, "/repositories", nil, "repositories")
}

// ListRepositoriesByOrg lists one organization's repositories.
func (s *Store) ListRepositoriesByOrg(ctx context.Context, org string) ([]*store.Repository, error) {
	if org == "" {
		return nil, nil
	}
	return list[store.Repository](ctx, s, "/repositories", url.Values{"org": {org}}, "repositories")
}

// --- SpecWorkflowStore (read and select) ---

// GetSpecWorkflow fetches one spec workflow.
func (s *Store) GetSpecWorkflow(ctx context.Context, id string) (*store.SpecWorkflow, error) {
	if err := requireID(http.MethodGet, "/workflows", "workflow", id); err != nil {
		return nil, err
	}
	return get[store.SpecWorkflow](ctx, s, idPath("/workflows", id))
}

// ListSpecWorkflows lists the tenant's spec workflows.
func (s *Store) ListSpecWorkflows(ctx context.Context) ([]*store.SpecWorkflow, error) {
	return list[store.SpecWorkflow](ctx, s, "/workflows", nil, "workflows")
}

// SelectWorkflowForInitiative selects the initiative's spec workflow. An
// unknown workflow fails with a 400 INVALID_REFERENCE.
func (s *Store) SelectWorkflowForInitiative(ctx context.Context, initiativeID, workflowID string) error {
	if err := requireID(http.MethodPut, "/initiatives", "initiative", initiativeID); err != nil {
		return err
	}
	req := struct {
		WorkflowID string `json:"workflow_id"`
	}{WorkflowID: workflowID}
	return s.do(ctx, http.MethodPut, idPath("/initiatives", initiativeID)+"/workflow", nil, req, nil)
}

// GetWorkflowForInitiative returns the initiative's selected workflow, or
// nil (and no error) when none is selected.
func (s *Store) GetWorkflowForInitiative(ctx context.Context, initiativeID string) (*store.InitiativeWorkflow, error) {
	if initiativeID == "" {
		return nil, nil
	}
	var out struct {
		Workflow *store.InitiativeWorkflow `json:"workflow"`
	}
	if err := s.do(ctx, http.MethodGet, idPath("/initiatives", initiativeID)+"/workflow", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Workflow, nil
}

// --- JudgeResultStore ---

// CreateJudgeResult POSTs a judge result; result is replaced with the
// server's copy.
func (s *Store) CreateJudgeResult(ctx context.Context, result *store.JudgeResult) error {
	if result == nil {
		return fmt.Errorf("remote: CreateJudgeResult: nil judge result")
	}
	return write(ctx, s, http.MethodPost, "/judge-results", result)
}

// ListJudgeResults lists an initiative's judge results.
func (s *Store) ListJudgeResults(ctx context.Context, initiativeID string) ([]*store.JudgeResult, error) {
	if initiativeID == "" {
		return nil, nil
	}
	return list[store.JudgeResult](ctx, s, idPath("/initiatives", initiativeID)+"/judge-results", nil, "judge_results")
}

// --- SpecDocumentStore ---

// CreateSpecDocument POSTs a spec document; doc is replaced with the
// server's copy.
func (s *Store) CreateSpecDocument(ctx context.Context, doc *store.SpecDocument) error {
	if doc == nil {
		return fmt.Errorf("remote: CreateSpecDocument: nil spec document")
	}
	return write(ctx, s, http.MethodPost, "/spec-documents", doc)
}

// GetSpecDocument fetches one spec document ("<initiative-id>/<spec-type>").
func (s *Store) GetSpecDocument(ctx context.Context, id string) (*store.SpecDocument, error) {
	if err := requireID(http.MethodGet, "/spec-documents", "spec document", id); err != nil {
		return nil, err
	}
	return get[store.SpecDocument](ctx, s, idPath("/spec-documents", id))
}

// ListSpecDocuments lists every spec document in the tenant.
func (s *Store) ListSpecDocuments(ctx context.Context) ([]*store.SpecDocument, error) {
	return s.listSpecDocuments(ctx, nil)
}

// ListSpecDocumentsByRepo lists one repository's spec documents.
func (s *Store) ListSpecDocumentsByRepo(ctx context.Context, repoID string) ([]*store.SpecDocument, error) {
	if repoID == "" {
		return nil, nil
	}
	return s.listSpecDocuments(ctx, url.Values{"repository": {repoID}})
}

// ListSpecDocumentsByInitiative lists one initiative's spec documents.
func (s *Store) ListSpecDocumentsByInitiative(ctx context.Context, initiativeID string) ([]*store.SpecDocument, error) {
	if initiativeID == "" {
		return nil, nil
	}
	return s.listSpecDocuments(ctx, url.Values{"initiative": {initiativeID}})
}

func (s *Store) listSpecDocuments(ctx context.Context, q url.Values) ([]*store.SpecDocument, error) {
	return list[store.SpecDocument](ctx, s, "/spec-documents", q, "spec_documents")
}

// UpdateSpecDocument replaces the spec document (PUT).
func (s *Store) UpdateSpecDocument(ctx context.Context, doc *store.SpecDocument) error {
	if doc == nil {
		return fmt.Errorf("remote: UpdateSpecDocument: nil spec document")
	}
	if err := requireID(http.MethodPut, "/spec-documents", "spec document", doc.ID); err != nil {
		return err
	}
	return write(ctx, s, http.MethodPut, idPath("/spec-documents", doc.ID), doc)
}

// DeleteSpecDocument deletes a spec document.
func (s *Store) DeleteSpecDocument(ctx context.Context, id string) error {
	if err := requireID(http.MethodDelete, "/spec-documents", "spec document", id); err != nil {
		return err
	}
	return s.do(ctx, http.MethodDelete, idPath("/spec-documents", id), nil, nil, nil)
}
