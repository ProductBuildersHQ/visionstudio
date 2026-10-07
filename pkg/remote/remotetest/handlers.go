package remotetest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/assignment"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Handlers mirror the cloud tenant API's (one per route, same validation,
// same status codes and error codes). Each delegates to the same
// pkg/service and store methods the cloud does.

// --- programs ---

func (s *Server) listPrograms(w http.ResponseWriter, r *http.Request) {
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		progs, err := svc.ListPrograms(ctx)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"programs": nonNil(progs)}, http.StatusOK, nil
	})
}

func (s *Server) getProgram(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		p, err := svc.GetProgram(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return p, http.StatusOK, nil
	})
}

func (s *Server) createProgram(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Organization string `json:"organization"`
		Description  string `json:"description,omitempty"`
		Hidden       bool   `json:"hidden,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		switch {
		case req.ID == "":
			return nil, 0, errMissing("id")
		case req.Name == "":
			return nil, 0, errMissing("name")
		}
		p, err := svc.CreateProgram(ctx, req.ID, req.Name, req.Organization, req.Description)
		if err != nil {
			return nil, 0, err
		}
		if req.Hidden {
			p.Hidden = true
			if err := svc.UpdateProgram(ctx, p); err != nil {
				return nil, 0, err
			}
		}
		return p, http.StatusCreated, nil
	})
}

func (s *Server) updateProgram(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data, ok := readBody(w, r)
	if !ok {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		p, err := buildUpdate(ctx, r.Method, data, id, svc.GetProgram, func(v *store.Program) *string { return &v.ID })
		if err != nil {
			return nil, 0, err
		}
		if err := svc.UpdateProgram(ctx, p); err != nil {
			return nil, 0, err
		}
		return p, http.StatusOK, nil
	})
}

// --- initiatives ---

func (s *Server) listInitiatives(w http.ResponseWriter, r *http.Request) {
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		inits, err := svc.ListInitiatives(ctx)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"initiatives": nonNil(inits)}, http.StatusOK, nil
	})
}

func (s *Server) getInitiative(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		in, err := svc.GetInitiative(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return in, http.StatusOK, nil
	})
}

func (s *Server) createInitiative(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID           string `json:"id"`
		Organization string `json:"organization"`
		Title        string `json:"title"`
		Description  string `json:"description,omitempty"`
		Priority     string `json:"priority,omitempty"`
		InitType     string `json:"init_type,omitempty"`
		WorkflowID   string `json:"workflow_id,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		in, err := svc.CreateInitiative(ctx, req.ID, req.Organization, req.Title, req.Description, req.Priority, req.InitType, req.WorkflowID)
		if err != nil {
			return nil, 0, asReference(err)
		}
		return in, http.StatusCreated, nil
	})
}

func (s *Server) updateInitiative(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data, ok := readBody(w, r)
	if !ok {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if _, err := svc.GetInitiative(ctx, id); err != nil {
			return nil, 0, err
		}
		in, err := buildUpdate(ctx, r.Method, data, id, svc.GetInitiative, func(v *store.Initiative) *string { return &v.ID })
		if err != nil {
			return nil, 0, err
		}
		if err := svc.UpdateInitiative(ctx, in); err != nil {
			return nil, 0, err
		}
		return in, http.StatusOK, nil
	})
}

func (s *Server) transitionInitiative(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if req.Status == "" {
			return nil, 0, errMissing("status")
		}
		in, err := svc.TransitionInitiative(ctx, id, req.Status)
		if err != nil {
			return nil, 0, err
		}
		return in, http.StatusOK, nil
	})
}

func (s *Server) getInitiativeWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		iw, err := svc.Store.GetWorkflowForInitiative(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"workflow": iw}, http.StatusOK, nil
	})
}

func (s *Server) selectInitiativeWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		WorkflowID string `json:"workflow_id"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if req.WorkflowID == "" {
			return nil, 0, errMissing("workflow_id")
		}
		if _, err := svc.GetInitiative(ctx, id); err != nil {
			return nil, 0, err
		}
		if err := svc.SelectWorkflow(ctx, id, req.WorkflowID); err != nil {
			return nil, 0, asReference(err)
		}
		iw, err := svc.Store.GetWorkflowForInitiative(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"workflow": iw}, http.StatusOK, nil
	})
}

func (s *Server) listJudgeResults(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		results, err := svc.Store.ListJudgeResults(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"judge_results": nonNil(results)}, http.StatusOK, nil
	})
}

func (s *Server) createJudgeResult(w http.ResponseWriter, r *http.Request) {
	var jr store.JudgeResult
	if !decodeJSON(w, r, &jr) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		switch {
		case jr.ID == "":
			return nil, 0, errMissing("id")
		case jr.InitiativeID == "":
			return nil, 0, errMissing("initiative_id")
		}
		if _, err := svc.GetInitiative(ctx, jr.InitiativeID); err != nil {
			return nil, 0, asReference(err)
		}
		if err := svc.Store.CreateJudgeResult(ctx, &jr); err != nil {
			return nil, 0, err
		}
		return &jr, http.StatusCreated, nil
	})
}

// --- phases ---

func (s *Server) listPhases(w http.ResponseWriter, r *http.Request) {
	initiativeID := r.URL.Query().Get("initiative")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if initiativeID != "" {
			ph, err := svc.ListPhases(ctx, initiativeID)
			if err != nil {
				return nil, 0, err
			}
			return map[string]any{"phases": nonNil(ph)}, http.StatusOK, nil
		}
		inits, err := svc.ListInitiatives(ctx)
		if err != nil {
			return nil, 0, err
		}
		var all []*store.Phase
		for _, in := range inits {
			ph, err := svc.ListPhases(ctx, in.ID)
			if err != nil {
				return nil, 0, err
			}
			all = append(all, ph...)
		}
		return map[string]any{"phases": nonNil(all)}, http.StatusOK, nil
	})
}

func (s *Server) createPhase(w http.ResponseWriter, r *http.Request) {
	var ph store.Phase
	if !decodeJSON(w, r, &ph) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		switch {
		case ph.ID == "":
			return nil, 0, errMissing("id")
		case ph.InitiativeID == "":
			return nil, 0, errMissing("initiative_id")
		}
		p, err := svc.CreatePhase(ctx, ph.ID, ph.InitiativeID, ph.SequenceNumber, ph.Title, ph.Theme)
		if err != nil {
			return nil, 0, asReference(err)
		}
		return p, http.StatusCreated, nil
	})
}

func (s *Server) deletePhase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if err := svc.RemovePhase(ctx, id); err != nil {
			if strings.Contains(err.Error(), "still has member RMIs") {
				return nil, 0, conflict(CodeConflict, err.Error())
			}
			return nil, 0, err
		}
		return nil, http.StatusNoContent, nil
	})
}

// --- RMIs ---

func (s *Server) listRMIs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	initiativeID, repoID, status := q.Get("initiative"), q.Get("repository"), q.Get("status")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		var (
			rmis []*store.RoadmapItem
			err  error
		)
		switch {
		case initiativeID != "":
			rmis, err = svc.ListRMIs(ctx, initiativeID)
		case repoID != "":
			rmis, err = svc.ListRMIsByRepo(ctx, repoID)
		case status != "":
			rmis, err = svc.Store.ListRMIsByStatus(ctx, status)
		default:
			rmis, err = svc.ListAllRMIs(ctx)
		}
		if err != nil {
			return nil, 0, err
		}
		out := make([]*store.RoadmapItem, 0, len(rmis))
		for _, rmi := range rmis {
			if (repoID == "" || rmi.RepositoryID == repoID) && (status == "" || rmi.Status == status) {
				out = append(out, rmi)
			}
		}
		return map[string]any{"rmis": out}, http.StatusOK, nil
	})
}

func (s *Server) getRMI(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		rmi, err := svc.GetRMI(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return rmi, http.StatusOK, nil
	})
}

func (s *Server) createRMI(w http.ResponseWriter, r *http.Request) {
	var req struct {
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
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		rmi, err := svc.CreateRMI(ctx, req.ID, req.RepositoryID, req.InitiativeID, req.PhaseID, req.Title, req.Description, req.ItemType, req.Priority, req.Required, req.SequenceNumber, req.AcceptanceCriteria)
		if err != nil {
			return nil, 0, asReference(err)
		}
		return rmi, http.StatusCreated, nil
	})
}

func (s *Server) updateRMI(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data, ok := readBody(w, r)
	if !ok {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		rmi, err := buildUpdate(ctx, r.Method, data, id, svc.GetRMI, func(v *store.RoadmapItem) *string { return &v.ID })
		if err != nil {
			return nil, 0, err
		}
		if err := svc.UpdateRMI(ctx, rmi); err != nil {
			return nil, 0, err
		}
		return rmi, http.StatusOK, nil
	})
}

func (s *Server) updateRMIStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if req.Status == "" {
			return nil, 0, errMissing("status")
		}
		if _, err := svc.GetRMI(ctx, id); err != nil {
			return nil, 0, err
		}
		rmi, err := svc.UpdateRMIStatus(ctx, id, req.Status)
		if err != nil {
			return nil, 0, err
		}
		return rmi, http.StatusOK, nil
	})
}

func (s *Server) moveRMI(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		PhaseID        string `json:"phase_id"`
		SequenceNumber int    `json:"sequence_number,omitempty"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if req.PhaseID == "" {
			return nil, 0, errMissing("phase_id")
		}
		if _, err := svc.GetRMI(ctx, id); err != nil {
			return nil, 0, err
		}
		rmi, err := svc.MoveRMI(ctx, id, req.PhaseID, req.SequenceNumber)
		if err != nil {
			return nil, 0, asReference(err)
		}
		return rmi, http.StatusOK, nil
	})
}

// --- RMI dependencies ---

func (s *Server) listRMIDependencies(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		deps, err := svc.ListDependencies(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"dependencies": nonNil(deps)}, http.StatusOK, nil
	})
}

func (s *Server) listDependencies(w http.ResponseWriter, r *http.Request) {
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		deps, err := svc.Store.ListAllDependencies(ctx)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"dependencies": nonNil(deps)}, http.StatusOK, nil
	})
}

func (s *Server) createDependency(w http.ResponseWriter, r *http.Request) {
	var dep store.RMIDependency
	if !decodeJSON(w, r, &dep) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		switch {
		case dep.SourceRMIID == "":
			return nil, 0, errMissing("source_rmi_id")
		case dep.TargetRMIID == "":
			return nil, 0, errMissing("target_rmi_id")
		}
		for _, ref := range []string{dep.SourceRMIID, dep.TargetRMIID} {
			if _, err := svc.GetRMI(ctx, ref); err != nil {
				return nil, 0, asReference(err)
			}
		}
		if err := svc.CreateDependency(ctx, dep.SourceRMIID, dep.TargetRMIID, dep.Relationship); err != nil {
			return nil, 0, err
		}
		if dep.Relationship == "" {
			dep.Relationship = "requires"
		}
		return &dep, http.StatusCreated, nil
	})
}

// --- assignments (leases) ---

var assignmentStatuses = map[string]bool{
	assignment.StatusActive:    true,
	assignment.StatusReleased:  true,
	assignment.StatusExpired:   true,
	assignment.StatusCompleted: true,
}

func validateAssignment(a *store.Assignment) error {
	switch {
	case a.ID == "":
		return errMissing("id")
	case a.RMIID == "":
		return errMissing("rmi_id")
	case !assignmentStatuses[a.Status]:
		return badRequest(fmt.Sprintf("invalid assignment status %q", a.Status))
	}
	return nil
}

func (s *Server) listAssignments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status, rmiID := q.Get("status"), q.Get("rmi")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		var (
			as  []*store.Assignment
			err error
		)
		switch status {
		case "":
			as, err = svc.Store.ListAllAssignments(ctx)
		case assignment.StatusActive:
			as, err = svc.ListActiveAssignments(ctx)
		default:
			return nil, 0, badRequest(fmt.Sprintf("unsupported status filter %q (only %q)", status, assignment.StatusActive))
		}
		if err != nil {
			return nil, 0, err
		}
		out := make([]*store.Assignment, 0, len(as))
		for _, a := range as {
			if rmiID == "" || a.RMIID == rmiID {
				out = append(out, a)
			}
		}
		return map[string]any{"assignments": out}, http.StatusOK, nil
	})
}

func (s *Server) getAssignment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		a, err := svc.GetAssignment(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return a, http.StatusOK, nil
	})
}

func (s *Server) getActiveAssignment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		a, err := svc.Store.GetActiveAssignment(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"assignment": a}, http.StatusOK, nil
	})
}

// createAssignment enforces one live lease per RMI: a live lease answers
// 409 LEASE_CONFLICT; an expired one is marked expired first.
func (s *Server) createAssignment(w http.ResponseWriter, r *http.Request) {
	var a store.Assignment
	if !decodeJSON(w, r, &a) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if err := validateAssignment(&a); err != nil {
			return nil, 0, err
		}
		s.leaseMu.Lock()
		defer s.leaseMu.Unlock()

		if _, err := svc.GetRMI(ctx, a.RMIID); err != nil {
			return nil, 0, asReference(err)
		}
		now := time.Now()
		if a.Status == assignment.StatusActive {
			if err := settleActiveLeases(ctx, svc, a.RMIID, a.Worker, now); err != nil {
				return nil, 0, err
			}
		}
		if a.CreatedAt.IsZero() {
			a.CreatedAt = now
		}
		if a.UpdatedAt.IsZero() {
			a.UpdatedAt = now
		}
		if err := svc.Store.CreateAssignment(ctx, &a); err != nil {
			return nil, 0, err
		}
		return &a, http.StatusCreated, nil
	})
}

func settleActiveLeases(ctx context.Context, svc *service.Service, rmiID, worker string, now time.Time) error {
	active, err := svc.ListActiveAssignments(ctx)
	if err != nil {
		return err
	}
	for _, ex := range active {
		if ex.RMIID != rmiID {
			continue
		}
		if _, err := assignment.Claim(rmiID, worker, assignment.DefaultLease, now, ex); err != nil {
			return conflict(CodeLeaseConflict, err.Error())
		}
		if assignment.ExpireStale(ex, now) {
			if err := svc.Store.UpdateAssignment(ctx, ex); err != nil {
				return fmt.Errorf("expire stale assignment %s: %w", ex.ID, err)
			}
		}
	}
	return nil
}

// updateAssignment only updates an active assignment (409 LEASE_CONFLICT
// otherwise) and never moves it to another RMI (400).
func (s *Server) updateAssignment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var a store.Assignment
	if !decodeJSON(w, r, &a) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		switch {
		case a.ID == "":
			a.ID = id
		case a.ID != id:
			return nil, 0, badRequest(fmt.Sprintf("body id %q does not match URL id %q", a.ID, id))
		}
		if err := validateAssignment(&a); err != nil {
			return nil, 0, err
		}
		s.leaseMu.Lock()
		defer s.leaseMu.Unlock()

		cur, err := svc.GetAssignment(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		if cur.RMIID != a.RMIID {
			return nil, 0, badRequest(fmt.Sprintf("assignment %s belongs to %s; rmi_id cannot change", id, cur.RMIID))
		}
		if cur.Status != assignment.StatusActive {
			return nil, 0, conflict(CodeLeaseConflict, fmt.Sprintf("assignment %s is %q; only an active assignment can be updated", id, cur.Status))
		}
		if a.UpdatedAt.IsZero() {
			a.UpdatedAt = time.Now()
		}
		if err := svc.Store.UpdateAssignment(ctx, &a); err != nil {
			return nil, 0, err
		}
		return &a, http.StatusOK, nil
	})
}

// --- delivery evidence ---

func (s *Server) listEvidence(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rmiID, initiativeID := q.Get("rmi"), q.Get("initiative")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		var (
			ev  []*store.DeliveryEvidence
			err error
		)
		switch {
		case rmiID != "" && initiativeID != "":
			return nil, 0, badRequest("pass either rmi or initiative, not both")
		case rmiID != "":
			ev, err = svc.Store.ListEvidenceByRMI(ctx, rmiID)
		case initiativeID != "":
			ev, err = svc.Store.ListEvidenceByInitiative(ctx, initiativeID)
		default:
			ev, err = svc.Store.ListAllEvidence(ctx)
		}
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"evidence": nonNil(ev)}, http.StatusOK, nil
	})
}

func (s *Server) createEvidence(w http.ResponseWriter, r *http.Request) {
	var ev store.DeliveryEvidence
	if !decodeJSON(w, r, &ev) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		switch {
		case ev.ID == "":
			return nil, 0, errMissing("id")
		case ev.RMIID == "":
			return nil, 0, errMissing("rmi_id")
		case ev.EvidenceType == "":
			return nil, 0, errMissing("evidence_type")
		case ev.Reference == "":
			return nil, 0, errMissing("reference")
		}
		if _, err := svc.GetRMI(ctx, ev.RMIID); err != nil {
			return nil, 0, asReference(err)
		}
		if ev.CreatedAt.IsZero() {
			ev.CreatedAt = time.Now()
		}
		if err := svc.Store.CreateEvidence(ctx, &ev); err != nil {
			return nil, 0, err
		}
		return &ev, http.StatusCreated, nil
	})
}

// --- releases ---

func (s *Server) listReleases(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	initiativeID, repoID := q.Get("initiative"), q.Get("repository")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		rels, err := svc.ListReleases(ctx, repoID, initiativeID)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"releases": nonNil(rels)}, http.StatusOK, nil
	})
}

func (s *Server) getRelease(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		rel, err := svc.GetRelease(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return rel, http.StatusOK, nil
	})
}

// --- repositories ---

func (s *Server) listRepositories(w http.ResponseWriter, r *http.Request) {
	org := r.URL.Query().Get("org")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		var (
			repos []*store.Repository
			err   error
		)
		if org != "" {
			repos, err = svc.ListRepositoriesByOrg(ctx, org)
		} else {
			repos, err = svc.ListRepositories(ctx)
		}
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"repositories": nonNil(repos)}, http.StatusOK, nil
	})
}

func (s *Server) getRepository(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		repo, err := svc.GetRepository(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return repo, http.StatusOK, nil
	})
}

// --- spec workflows ---

func (s *Server) listWorkflows(w http.ResponseWriter, r *http.Request) {
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		wfs, err := svc.ListWorkflows(ctx)
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"workflows": nonNil(wfs)}, http.StatusOK, nil
	})
}

func (s *Server) getWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		wf, err := svc.GetWorkflow(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return wf, http.StatusOK, nil
	})
}

// --- spec documents ---

func (s *Server) listSpecDocuments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	initiativeID, repoID := q.Get("initiative"), q.Get("repository")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		var (
			docs []*store.SpecDocument
			err  error
		)
		switch {
		case initiativeID != "" && repoID != "":
			return nil, 0, badRequest("pass either initiative or repository, not both")
		case initiativeID != "":
			docs, err = svc.ListSpecDocumentsByInitiative(ctx, initiativeID)
		case repoID != "":
			docs, err = svc.ListSpecDocumentsByRepo(ctx, repoID)
		default:
			docs, err = svc.ListSpecDocuments(ctx)
		}
		if err != nil {
			return nil, 0, err
		}
		return map[string]any{"spec_documents": nonNil(docs)}, http.StatusOK, nil
	})
}

func (s *Server) getSpecDocument(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		doc, err := svc.GetSpecDocument(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		return doc, http.StatusOK, nil
	})
}

func (s *Server) createSpecDocument(w http.ResponseWriter, r *http.Request) {
	var doc store.SpecDocument
	if !decodeJSON(w, r, &doc) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		switch {
		case doc.ID == "":
			return nil, 0, errMissing("id")
		case doc.SpecType == "":
			return nil, 0, errMissing("spec_type")
		}
		if doc.InitiativeID != "" {
			if _, err := svc.GetInitiative(ctx, doc.InitiativeID); err != nil {
				return nil, 0, asReference(err)
			}
		}
		if err := svc.CreateSpecDocument(ctx, &doc); err != nil {
			return nil, 0, err
		}
		return &doc, http.StatusCreated, nil
	})
}

func (s *Server) replaceSpecDocument(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var doc store.SpecDocument
	if !decodeJSON(w, r, &doc) {
		return
	}
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		switch {
		case doc.ID == "":
			doc.ID = id
		case doc.ID != id:
			return nil, 0, badRequest(fmt.Sprintf("body id %q does not match URL id %q", doc.ID, id))
		}
		if _, err := svc.GetSpecDocument(ctx, id); err != nil {
			return nil, 0, err
		}
		if err := svc.UpdateSpecDocument(ctx, &doc); err != nil {
			return nil, 0, err
		}
		return &doc, http.StatusOK, nil
	})
}

func (s *Server) deleteSpecDocument(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.withService(w, r, func(ctx context.Context, svc *service.Service) (any, int, error) {
		if _, err := svc.GetSpecDocument(ctx, id); err != nil {
			return nil, 0, err
		}
		if err := svc.DeleteSpecDocument(ctx, id); err != nil {
			return nil, 0, err
		}
		return nil, http.StatusNoContent, nil
	})
}
