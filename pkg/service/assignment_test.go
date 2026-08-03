package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/pcerr"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func setupClaimTest(t *testing.T) (*Service, context.Context, string) {
	t.Helper()
	ctx := context.Background()
	svc := newTestService()

	rmi, err := svc.CreateRMI(ctx, "RMI-C-001", "github.com/org/repo", "INIT-C-001", "",
		"Claimable item", "", "capability", "", true, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateRMIStatus(ctx, rmi.ID, RMIStatusReady); err != nil {
		t.Fatal(err)
	}
	return svc, ctx, rmi.ID
}

func TestClaimRMI(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	result, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "/workspace", 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Assignment.Status != "active" {
		t.Fatalf("expected active, got %s", result.Assignment.Status)
	}
	if result.Assignment.Worker != "session-abc" {
		t.Fatalf("expected worker session-abc, got %s", result.Assignment.Worker)
	}
	if result.TrailerLine != "Refs: RMI-C-001" {
		t.Fatalf("unexpected trailer: %s", result.TrailerLine)
	}

	// RMI should now be in_progress
	rmi, err := svc.GetRMI(ctx, rmiID)
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != RMIStatusInProgress {
		t.Fatalf("expected RMI in_progress after claim, got %s", rmi.Status)
	}
}

func TestClaimRMIAlreadyClaimed(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	if _, err := svc.ClaimRMI(ctx, rmiID, "session-1", "", 0); err != nil {
		t.Fatal(err)
	}

	// Second claim should fail
	_, err := svc.ClaimRMI(ctx, rmiID, "session-2", "", 0)
	if err == nil {
		t.Fatal("expected error on double claim")
	}
}

func TestClaimRMIWrongStatus(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateRMI(ctx, "RMI-CS-001", "github.com/org/repo", "", "", "Item", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateRMIStatus(ctx, "RMI-CS-001", RMIStatusInProgress); err != nil {
		t.Fatal(err)
	}

	_, err := svc.ClaimRMI(ctx, "RMI-CS-001", "session-x", "", 0)
	if err == nil {
		t.Fatal("expected error claiming in_progress RMI")
	}
}

func TestRenewLease(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	result, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	originalExpiry := result.Assignment.LeaseExpiresAt

	renewed, err := svc.RenewLease(ctx, result.Assignment.ID, 8*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !renewed.LeaseExpiresAt.After(originalExpiry) {
		t.Fatal("expected lease to be extended")
	}
}

func TestRenewLeaseNotActive(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	result, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 0)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.ReleaseWork(ctx, result.Assignment.ID, nil); err != nil {
		t.Fatal(err)
	}

	_, err = svc.RenewLease(ctx, result.Assignment.ID, 0)
	if err == nil {
		t.Fatal("expected error renewing released assignment")
	}
}

func TestReleaseWork(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	result, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 0)
	if err != nil {
		t.Fatal(err)
	}

	handoff := &store.Handoff{
		Completed:  []string{"step 1"},
		Remaining:  []string{"step 2"},
		Decisions:  []string{"chose approach A"},
		NextAction: "continue with step 2",
	}

	released, err := svc.ReleaseWork(ctx, result.Assignment.ID, handoff)
	if err != nil {
		t.Fatal(err)
	}
	if released.Status != "released" {
		t.Fatalf("expected released, got %s", released.Status)
	}
	if released.Handoff == nil {
		t.Fatal("expected handoff to be set")
	}
	if released.Handoff.NextAction != "continue with step 2" {
		t.Fatalf("unexpected next_action: %s", released.Handoff.NextAction)
	}

	// RMI should be back to ready
	rmi, err := svc.GetRMI(ctx, rmiID)
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != RMIStatusReady {
		t.Fatalf("expected RMI ready after release, got %s", rmi.Status)
	}
}

func TestCompleteWork(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	result, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 0)
	if err != nil {
		t.Fatal(err)
	}

	completed, err := svc.CompleteWork(ctx, result.Assignment.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != "completed" {
		t.Fatalf("expected completed, got %s", completed.Status)
	}

	// RMI should be completed
	rmi, err := svc.GetRMI(ctx, rmiID)
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != RMIStatusCompleted {
		t.Fatalf("expected RMI completed, got %s", rmi.Status)
	}
	if rmi.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set")
	}
}

func TestUpdateHandoff(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	result, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 0)
	if err != nil {
		t.Fatal(err)
	}

	handoff := &store.Handoff{
		Completed:  []string{"implemented core"},
		Remaining:  []string{"add tests"},
		NextAction: "write unit tests",
	}

	updated, err := svc.UpdateHandoff(ctx, result.Assignment.ID, handoff)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Handoff == nil {
		t.Fatal("expected handoff to be set")
	}
	if updated.Handoff.NextAction != "write unit tests" {
		t.Fatalf("unexpected next_action: %s", updated.Handoff.NextAction)
	}
}

func TestAddEvidence(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	ev, err := svc.AddEvidence(ctx, rmiID, "commit", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if ev.EvidenceType != "commit" {
		t.Fatalf("expected commit, got %s", ev.EvidenceType)
	}
	if ev.Reference != "abc123" {
		t.Fatalf("expected abc123, got %s", ev.Reference)
	}

	// nonexistent RMI
	_, err = svc.AddEvidence(ctx, "RMI-NOPE-001", "commit", "def456")
	if err == nil {
		t.Fatal("expected error for nonexistent RMI")
	}
}

func TestListActiveAssignments(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	if _, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 0); err != nil {
		t.Fatal(err)
	}

	active, err := svc.ListActiveAssignments(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active assignment, got %d", len(active))
	}
}

func TestResolveAssignmentByRMIID(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	result, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 0)
	if err != nil {
		t.Fatal(err)
	}

	a, err := svc.ResolveAssignment(ctx, rmiID)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != result.Assignment.ID {
		t.Fatalf("expected %s, got %s", result.Assignment.ID, a.ID)
	}
}

func TestResolveAssignmentByAssignmentID(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	result, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 0)
	if err != nil {
		t.Fatal(err)
	}

	a, err := svc.ResolveAssignment(ctx, result.Assignment.ID)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != result.Assignment.ID {
		t.Fatalf("expected %s, got %s", result.Assignment.ID, a.ID)
	}
}

func TestResolveAssignmentNoActive(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	_, err := svc.ResolveAssignment(ctx, rmiID)
	if err == nil {
		t.Fatal("expected error for unclaimed RMI")
	}
}

func TestCompleteWorkByRef(t *testing.T) {
	svc, ctx, rmiID := setupClaimTest(t)

	if _, err := svc.ClaimRMI(ctx, rmiID, "session-abc", "", 0); err != nil {
		t.Fatal(err)
	}

	a, err := svc.CompleteWorkByRef(ctx, rmiID, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != "completed" {
		t.Fatalf("expected completed, got %s", a.Status)
	}

	rmi, err := svc.GetRMI(ctx, rmiID)
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != RMIStatusCompleted {
		t.Fatalf("expected RMI completed after transition, got %s", rmi.Status)
	}
}

func setupPhaseTest(t *testing.T) (*Service, context.Context, string) {
	t.Helper()
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-P-001", "org", "Phase test", "", "high", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-P-001/phase-1", "INIT-P-001", 1, "Phase 1", "test"); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("RMI-P-%03d", i)
		if _, err := svc.CreateRMI(ctx, id, "github.com/org/repo", "INIT-P-001", "INIT-P-001/phase-1",
			fmt.Sprintf("Item %d", i), "", "capability", "", true, i, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.UpdateRMIStatus(ctx, id, RMIStatusReady); err != nil {
			t.Fatal(err)
		}
	}
	return svc, ctx, "INIT-P-001/phase-1"
}

func TestClaimPhase(t *testing.T) {
	svc, ctx, phaseID := setupPhaseTest(t)

	result, err := svc.ClaimPhase(ctx, phaseID, "session-1", "ws", 4*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Claimed) != 3 {
		t.Fatalf("expected 3 claims, got %d", len(result.Claimed))
	}

	for _, r := range result.Claimed {
		if r.Assignment.Status != "active" {
			t.Fatalf("expected active, got %s", r.Assignment.Status)
		}
		if r.Assignment.Worker != "session-1" {
			t.Fatalf("expected worker session-1, got %s", r.Assignment.Worker)
		}
	}

	for i := 1; i <= 3; i++ {
		rmi, err := svc.GetRMI(ctx, fmt.Sprintf("RMI-P-%03d", i))
		if err != nil {
			t.Fatal(err)
		}
		if rmi.Status != RMIStatusInProgress {
			t.Fatalf("expected in_progress, got %s", rmi.Status)
		}
	}
}

func TestClaimPhaseEmpty(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-E-001", "org", "Empty", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-E-001/phase-1", "INIT-E-001", 1, "Phase 1", "test"); err != nil {
		t.Fatal(err)
	}

	_, err := svc.ClaimPhase(ctx, "INIT-E-001/phase-1", "session-1", "", 0)
	if err == nil {
		t.Fatal("expected error for empty phase")
	}
	if !pcerr.IsNotFound(err) {
		t.Fatalf("expected NOT_FOUND error, got: %v", err)
	}
}

func TestClaimPhaseAllProposed(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-AP-001", "org", "All Proposed", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-AP-001/phase-1", "INIT-AP-001", 1, "Phase 1", "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-AP-001", "github.com/org/repo", "INIT-AP-001", "INIT-AP-001/phase-1",
		"Proposed item", "", "capability", "", true, 1, nil); err != nil {
		t.Fatal(err)
	}

	_, err := svc.ClaimPhase(ctx, "INIT-AP-001/phase-1", "session-1", "", 0)
	if err == nil {
		t.Fatal("expected error for all-proposed phase")
	}
	if !pcerr.IsBlocked(err) {
		t.Fatalf("expected BLOCKED error, got: %v", err)
	}
}

func TestClaimPhaseInvalidFormat(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, err := svc.ClaimPhase(ctx, "bad-format", "session-1", "", 0)
	if err == nil {
		t.Fatal("expected error for bad phase ID format")
	}
	if !pcerr.IsInput(err) {
		t.Fatalf("expected INPUT error, got: %v", err)
	}
}

func TestCompletePhase(t *testing.T) {
	svc, ctx, phaseID := setupPhaseTest(t)

	if _, err := svc.ClaimPhase(ctx, phaseID, "session-1", "", 0); err != nil {
		t.Fatal(err)
	}

	result, err := svc.CompletePhase(ctx, phaseID, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Completed) != 3 {
		t.Fatalf("expected 3 completions, got %d", len(result.Completed))
	}

	for i := 1; i <= 3; i++ {
		rmi, err := svc.GetRMI(ctx, fmt.Sprintf("RMI-P-%03d", i))
		if err != nil {
			t.Fatal(err)
		}
		if rmi.Status != RMIStatusCompleted {
			t.Fatalf("expected completed, got %s", rmi.Status)
		}
	}
}

func TestCompletePhasePartial(t *testing.T) {
	svc, ctx, phaseID := setupPhaseTest(t)

	if _, err := svc.ClaimPhase(ctx, phaseID, "session-1", "", 0); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.CompleteWorkByRef(ctx, "RMI-P-001", nil, true); err != nil {
		t.Fatal(err)
	}

	result, err := svc.CompletePhase(ctx, phaseID, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Completed) != 2 {
		t.Fatalf("expected 2 completions (1 already done), got %d", len(result.Completed))
	}
}

func TestCompletePhaseNoClaims(t *testing.T) {
	svc, ctx, phaseID := setupPhaseTest(t)

	_, err := svc.CompletePhase(ctx, phaseID, nil, true)
	if err == nil {
		t.Fatal("expected error for unclaimed phase")
	}
	if !pcerr.IsBlocked(err) {
		t.Fatalf("expected BLOCKED error, got: %v", err)
	}
}
