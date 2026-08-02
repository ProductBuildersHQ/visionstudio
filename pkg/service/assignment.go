package service

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/assignment"
	"github.com/ProductBuildersHQ/visionstudio/pkg/pcerr"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

var evidenceSeq atomic.Int64

// ClaimResult is returned by ClaimRMI with the assignment and trailer.
type ClaimResult struct {
	Assignment  *store.Assignment
	TrailerLine string
}

// ClaimRMI claims an RMI for a worker session, creating a lease-based assignment.
// It transitions the RMI to in_progress and returns the git trailer line.
func (s *Service) ClaimRMI(ctx context.Context, rmiID, worker, workspace string, lease time.Duration) (*ClaimResult, error) {
	rmi, err := s.Store.GetRMI(ctx, rmiID)
	if err != nil {
		return nil, pcerr.Wrap(pcerr.NotFound,
			fmt.Sprintf("RMI %s not found", rmiID),
			"verify the ID with: prismctl rmi list --initiative <INIT-ID>", err)
	}

	if rmi.Status != RMIStatusReady && rmi.Status != RMIStatusProposed && rmi.Status != RMIStatusPlanned {
		switch rmi.Status {
		case RMIStatusInProgress:
			return nil, pcerr.New(pcerr.StateConflict,
				fmt.Sprintf("RMI %s is already in_progress", rmiID),
				"check the current owner with: prismctl work status")
		case RMIStatusCompleted:
			return nil, pcerr.New(pcerr.StateAlreadyDone,
				fmt.Sprintf("RMI %s is already completed", rmiID),
				"nothing to claim — this work is finished")
		default:
			return nil, pcerr.New(pcerr.StateWrongStatus,
				fmt.Sprintf("RMI %s has status %q — must be ready, proposed, or planned to claim", rmiID, rmi.Status),
				fmt.Sprintf("transition it first with: prismctl rmi update %s --status ready", rmiID))
		}
	}

	existing, err := s.Store.GetActiveAssignment(ctx, rmiID)
	if err != nil {
		return nil, pcerr.Wrap(pcerr.InternalStore, "failed to check existing assignment", "", err)
	}

	now := time.Now()
	a, err := assignment.Claim(rmiID, worker, lease, now, existing)
	if err != nil {
		return nil, err
	}

	a.ID = fmt.Sprintf("assign-%s-%d", rmiID, now.UnixMilli())
	a.Workspace = workspace

	if err := s.Store.CreateAssignment(ctx, a); err != nil {
		return nil, fmt.Errorf("create assignment: %w", err)
	}

	rmi.Status = RMIStatusInProgress
	rmi.UpdatedAt = now
	if err := s.Store.UpdateRMI(ctx, rmi); err != nil {
		return nil, fmt.Errorf("transition RMI to in_progress: %w", err)
	}

	return &ClaimResult{
		Assignment:  a,
		TrailerLine: assignment.TrailerLine(rmiID),
	}, nil
}

// RenewLease extends a lease on an active assignment.
func (s *Service) RenewLease(ctx context.Context, assignmentID string, lease time.Duration) (*store.Assignment, error) {
	a, err := s.Store.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment %s: %w", assignmentID, err)
	}

	now := time.Now()
	if err := assignment.Renew(a, lease, now); err != nil {
		return nil, err
	}

	if err := s.Store.UpdateAssignment(ctx, a); err != nil {
		return nil, fmt.Errorf("update assignment: %w", err)
	}
	return a, nil
}

// ReleaseWork releases an active assignment and transitions the RMI back to ready.
func (s *Service) ReleaseWork(ctx context.Context, assignmentID string, handoff *store.Handoff) (*store.Assignment, error) {
	a, err := s.Store.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment %s: %w", assignmentID, err)
	}

	now := time.Now()
	if err := assignment.Release(a, now); err != nil {
		return nil, err
	}
	a.Handoff = handoff

	if err := s.Store.UpdateAssignment(ctx, a); err != nil {
		return nil, fmt.Errorf("update assignment: %w", err)
	}

	rmi, err := s.Store.GetRMI(ctx, a.RMIID)
	if err != nil {
		return nil, fmt.Errorf("get RMI for release: %w", err)
	}
	rmi.Status = RMIStatusReady
	rmi.UpdatedAt = now
	if err := s.Store.UpdateRMI(ctx, rmi); err != nil {
		return nil, fmt.Errorf("transition RMI back to ready: %w", err)
	}

	return a, nil
}

// CompleteWork marks an assignment as completed and the RMI as completed.
func (s *Service) CompleteWork(ctx context.Context, assignmentID string, handoff *store.Handoff) (*store.Assignment, error) {
	a, err := s.Store.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment %s: %w", assignmentID, err)
	}

	now := time.Now()
	if err := assignment.Complete(a, now); err != nil {
		return nil, err
	}
	a.Handoff = handoff

	if err := s.Store.UpdateAssignment(ctx, a); err != nil {
		return nil, fmt.Errorf("update assignment: %w", err)
	}

	rmi, err := s.Store.GetRMI(ctx, a.RMIID)
	if err != nil {
		return nil, fmt.Errorf("get RMI for completion: %w", err)
	}
	rmi.Status = RMIStatusCompleted
	rmi.CompletedAt = &now
	rmi.UpdatedAt = now
	if err := s.Store.UpdateRMI(ctx, rmi); err != nil {
		return nil, fmt.Errorf("transition RMI to completed: %w", err)
	}

	return a, nil
}

// UpdateHandoff updates the handoff state on an active assignment.
func (s *Service) UpdateHandoff(ctx context.Context, assignmentID string, handoff *store.Handoff) (*store.Assignment, error) {
	a, err := s.Store.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment %s: %w", assignmentID, err)
	}
	if a.Status != "active" {
		return nil, fmt.Errorf("assignment %s is %q, must be active", assignmentID, a.Status)
	}

	a.Handoff = handoff
	a.UpdatedAt = time.Now()

	if err := s.Store.UpdateAssignment(ctx, a); err != nil {
		return nil, fmt.Errorf("update assignment: %w", err)
	}
	return a, nil
}

// AddEvidence attaches delivery evidence to an RMI.
func (s *Service) AddEvidence(ctx context.Context, rmiID, evidenceType, reference string) (*store.DeliveryEvidence, error) {
	if _, err := s.Store.GetRMI(ctx, rmiID); err != nil {
		return nil, fmt.Errorf("RMI %s: %w", rmiID, err)
	}

	now := time.Now()
	seq := evidenceSeq.Add(1)
	ev := &store.DeliveryEvidence{
		ID:           fmt.Sprintf("ev-%s-%d-%d", rmiID, now.UnixMilli(), seq),
		RMIID:        rmiID,
		EvidenceType: evidenceType,
		Reference:    reference,
		CreatedAt:    now,
	}
	if err := s.Store.CreateEvidence(ctx, ev); err != nil {
		return nil, fmt.Errorf("create evidence: %w", err)
	}
	return ev, nil
}

// ListActiveAssignments returns all active assignments.
func (s *Service) ListActiveAssignments(ctx context.Context) ([]*store.Assignment, error) {
	return s.Store.ListActiveAssignments(ctx)
}

// GetAssignment returns an assignment by ID.
func (s *Service) GetAssignment(ctx context.Context, id string) (*store.Assignment, error) {
	return s.Store.GetAssignment(ctx, id)
}

// ResolveAssignment accepts either an assignment ID ("assign-...") or an
// RMI ID ("RMI-...") and returns the corresponding assignment. For RMI IDs
// it returns the active assignment.
func (s *Service) ResolveAssignment(ctx context.Context, ref string) (*store.Assignment, error) {
	if strings.HasPrefix(ref, "RMI-") {
		rmi, err := s.Store.GetRMI(ctx, ref)
		if err != nil {
			return nil, pcerr.Wrap(pcerr.NotFound,
				fmt.Sprintf("RMI %s not found", ref),
				"verify the ID with: prismctl rmi list", err)
		}
		a, err := s.Store.GetActiveAssignment(ctx, ref)
		if err != nil {
			return nil, pcerr.Wrap(pcerr.InternalStore,
				fmt.Sprintf("failed to look up assignment for %s", ref), "", err)
		}
		if a == nil {
			return nil, pcerr.New(pcerr.StateWrongStatus,
				fmt.Sprintf("%s has no active assignment (status=%q)", ref, rmi.Status),
				fmt.Sprintf("claim it first with: prismctl work claim %s --worker <session-id>", ref))
		}
		return a, nil
	}
	a, err := s.Store.GetAssignment(ctx, ref)
	if err != nil {
		return nil, pcerr.Wrap(pcerr.NotFound,
			fmt.Sprintf("assignment %q not found", ref),
			"if you meant an RMI, pass the RMI ID (e.g. RMI-MYREPO-001) instead of the assignment ID", err)
	}
	return a, nil
}

// CompleteWorkByRef resolves a ref (RMI ID or assignment ID) and completes
// the assignment. If transition is true, the RMI is also marked completed.
func (s *Service) CompleteWorkByRef(ctx context.Context, ref string, handoff *store.Handoff, transition bool) (*store.Assignment, error) {
	a, err := s.ResolveAssignment(ctx, ref)
	if err != nil {
		return nil, err
	}

	a, err = s.CompleteWork(ctx, a.ID, handoff)
	if err != nil {
		return nil, err
	}

	if transition {
		if _, err := s.UpdateRMIStatus(ctx, a.RMIID, RMIStatusCompleted); err != nil {
			return nil, fmt.Errorf("transition RMI: %w", err)
		}
	}
	return a, nil
}

// ReleaseWorkByRef resolves a ref and releases the assignment.
func (s *Service) ReleaseWorkByRef(ctx context.Context, ref string, handoff *store.Handoff) (*store.Assignment, error) {
	a, err := s.ResolveAssignment(ctx, ref)
	if err != nil {
		return nil, err
	}
	return s.ReleaseWork(ctx, a.ID, handoff)
}

// RenewLeaseByRef resolves a ref and renews the lease.
func (s *Service) RenewLeaseByRef(ctx context.Context, ref string, lease time.Duration) (*store.Assignment, error) {
	a, err := s.ResolveAssignment(ctx, ref)
	if err != nil {
		return nil, err
	}
	return s.RenewLease(ctx, a.ID, lease)
}

// UpdateHandoffByRef resolves a ref and updates the handoff.
func (s *Service) UpdateHandoffByRef(ctx context.Context, ref string, handoff *store.Handoff) (*store.Assignment, error) {
	a, err := s.ResolveAssignment(ctx, ref)
	if err != nil {
		return nil, err
	}
	return s.UpdateHandoff(ctx, a.ID, handoff)
}

// ClaimPhaseResult holds the claimed assignments and diagnostic info about
// skipped RMIs so the caller can report why items weren't claimed.
type ClaimPhaseResult struct {
	Claimed      []*ClaimResult
	Blocked      []string // RMI IDs skipped because of unmet dependencies
	AlreadyOwned []string // RMI IDs skipped because already claimed
}

// ClaimPhase claims all ready, unblocked, unclaimed RMIs in a phase.
// Returns a ClaimPhaseResult with diagnostics about skipped items.
func (s *Service) ClaimPhase(ctx context.Context, phaseID, worker, workspace string, lease time.Duration) (*ClaimPhaseResult, error) {
	allRMIs, err := s.phaseAllRMIs(ctx, phaseID)
	if err != nil {
		return nil, err
	}

	if len(allRMIs) == 0 {
		return nil, pcerr.New(pcerr.NotFound,
			fmt.Sprintf("phase %s has no RMIs", phaseID),
			"verify the phase exists with: prismctl initiative get <INIT-ID>")
	}

	readyRMIs, err := s.phaseRMIsByStatus(ctx, phaseID, RMIStatusReady)
	if err != nil {
		return nil, err
	}

	if len(readyRMIs) == 0 {
		counts := s.countByStatus(allRMIs)
		return nil, pcerr.New(pcerr.BlockedEmpty,
			fmt.Sprintf("no ready RMIs in phase %s — found %s", phaseID, formatStatusCounts(counts)),
			s.suggestRecovery(counts, phaseID))
	}

	result := &ClaimPhaseResult{}
	for _, rmi := range readyRMIs {
		blocked, err := s.isBlockedByDeps(ctx, rmi.ID)
		if err != nil {
			return nil, pcerr.Wrap(pcerr.InternalStore, fmt.Sprintf("check deps for %s", rmi.ID), "", err)
		}
		if blocked {
			result.Blocked = append(result.Blocked, rmi.ID)
			continue
		}

		claimed, err := s.isClaimed(ctx, rmi.ID)
		if err != nil {
			return nil, pcerr.Wrap(pcerr.InternalStore, fmt.Sprintf("check assignment for %s", rmi.ID), "", err)
		}
		if claimed {
			result.AlreadyOwned = append(result.AlreadyOwned, rmi.ID)
			continue
		}

		cr, err := s.ClaimRMI(ctx, rmi.ID, worker, workspace, lease)
		if err != nil {
			return nil, fmt.Errorf("claim %s: %w", rmi.ID, err)
		}
		result.Claimed = append(result.Claimed, cr)
	}
	return result, nil
}

func (s *Service) countByStatus(rmis []*store.RoadmapItem) map[string]int {
	counts := map[string]int{}
	for _, r := range rmis {
		counts[r.Status]++
	}
	return counts
}

func formatStatusCounts(counts map[string]int) string {
	var parts []string
	for status, count := range counts {
		parts = append(parts, fmt.Sprintf("%d %s", count, status))
	}
	return strings.Join(parts, ", ")
}

func (s *Service) suggestRecovery(counts map[string]int, phaseID string) string {
	if n := counts[RMIStatusProposed] + counts[RMIStatusPlanned]; n > 0 {
		return fmt.Sprintf(
			"%d RMIs are proposed/planned — use: prismctl work claim-phase %s --ready --worker <session-id> (or bulk-transition with: prismctl rmi update-phase %s --status ready)",
			n, phaseID, phaseID)
	}
	if counts[RMIStatusInProgress] > 0 {
		return fmt.Sprintf(
			"%d RMIs are in_progress — check owners with: prismctl work status",
			counts[RMIStatusInProgress])
	}
	if counts[RMIStatusCompleted] > 0 {
		return "all RMIs in this phase are already completed"
	}
	return "check phase status with: prismctl initiative get " + strings.SplitN(phaseID, "/", 2)[0]
}

// CompletePhaseResult holds completed assignments and diagnostics.
type CompletePhaseResult struct {
	Completed             []*store.Assignment
	NoAssignment          []string // in_progress RMIs without an active assignment
	InitiativeAllComplete bool     // true if all required RMIs in the initiative are now completed
	InitiativeID          string   // the parent initiative ID
}

// CompletePhase completes all in-progress RMIs with active assignments in a phase.
// If transition is true, each RMI is also transitioned to completed.
func (s *Service) CompletePhase(ctx context.Context, phaseID string, handoff *store.Handoff, transition bool) (*CompletePhaseResult, error) {
	allRMIs, err := s.phaseAllRMIs(ctx, phaseID)
	if err != nil {
		return nil, err
	}

	if len(allRMIs) == 0 {
		return nil, pcerr.New(pcerr.NotFound,
			fmt.Sprintf("phase %s has no RMIs", phaseID),
			"verify the phase exists with: prismctl initiative get <INIT-ID>")
	}

	inProgress, err := s.phaseRMIsByStatus(ctx, phaseID, RMIStatusInProgress)
	if err != nil {
		return nil, err
	}

	if len(inProgress) == 0 {
		counts := s.countByStatus(allRMIs)
		return nil, pcerr.New(pcerr.BlockedEmpty,
			fmt.Sprintf("no in-progress RMIs in phase %s — found %s", phaseID, formatStatusCounts(counts)),
			"claim the phase first with: prismctl work claim-phase "+phaseID+" --worker <session-id>")
	}

	parts := strings.SplitN(phaseID, "/", 2)
	initiativeID := parts[0]

	result := &CompletePhaseResult{InitiativeID: initiativeID}
	for _, rmi := range inProgress {
		a, err := s.Store.GetActiveAssignment(ctx, rmi.ID)
		if err != nil {
			return nil, pcerr.Wrap(pcerr.InternalStore, fmt.Sprintf("get assignment for %s", rmi.ID), "", err)
		}
		if a == nil {
			result.NoAssignment = append(result.NoAssignment, rmi.ID)
			continue
		}

		completed, err := s.CompleteWork(ctx, a.ID, handoff)
		if err != nil {
			return nil, fmt.Errorf("complete %s: %w", rmi.ID, err)
		}

		if transition {
			if _, err := s.UpdateRMIStatus(ctx, rmi.ID, RMIStatusCompleted); err != nil {
				return nil, fmt.Errorf("transition %s: %w", rmi.ID, err)
			}
		}
		result.Completed = append(result.Completed, completed)
	}

	if transition && len(result.Completed) > 0 {
		result.InitiativeAllComplete = s.CheckInitiativeAllComplete(ctx, initiativeID)
	}

	return result, nil
}

// CheckInitiativeAllComplete returns true if all required RMIs in the initiative are completed.
func (s *Service) CheckInitiativeAllComplete(ctx context.Context, initiativeID string) bool {
	rmis, err := s.Store.ListRMIs(ctx, initiativeID)
	if err != nil {
		return false
	}
	for _, r := range rmis {
		if r.Required && r.Status != RMIStatusCompleted && r.Status != RMIStatusCancelled {
			return false
		}
	}
	return len(rmis) > 0
}

func (s *Service) phaseAllRMIs(ctx context.Context, phaseID string) ([]*store.RoadmapItem, error) {
	parts := strings.SplitN(phaseID, "/", 2)
	if len(parts) != 2 {
		return nil, pcerr.New(pcerr.InputInvalid,
			fmt.Sprintf("invalid phase ID %q", phaseID),
			"expected format: INITIATIVE-ID/phase-N (e.g. INIT-PRISMCONTROL-001/phase-5)")
	}
	initiativeID := parts[0]

	if _, err := s.Store.GetInitiative(ctx, initiativeID); err != nil {
		return nil, pcerr.Wrap(pcerr.NotFound,
			fmt.Sprintf("initiative %s not found", initiativeID),
			"list initiatives with: prismctl initiative list", err)
	}

	allRMIs, err := s.Store.ListRMIs(ctx, initiativeID)
	if err != nil {
		return nil, pcerr.Wrap(pcerr.InternalStore, "failed to list RMIs", "", err)
	}

	var filtered []*store.RoadmapItem
	for _, r := range allRMIs {
		if r.PhaseID == phaseID {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

func (s *Service) phaseRMIsByStatus(ctx context.Context, phaseID, status string) ([]*store.RoadmapItem, error) {
	allRMIs, err := s.phaseAllRMIs(ctx, phaseID)
	if err != nil {
		return nil, err
	}

	var filtered []*store.RoadmapItem
	for _, r := range allRMIs {
		if r.Status == status {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}
