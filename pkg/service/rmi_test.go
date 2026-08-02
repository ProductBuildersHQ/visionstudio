package service

import (
	"context"
	"testing"
)

func TestCreateRMI(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	rmi, err := svc.CreateRMI(ctx, "RMI-TEST-001", "github.com/test/repo", "INIT-A-001", "INIT-A-001/phase-1",
		"Add feature X", "Detailed description", "capability", "high", true, 1,
		[]string{"unit tests pass", "docs updated"})
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != RMIStatusProposed {
		t.Fatalf("expected status proposed, got %s", rmi.Status)
	}
	if rmi.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if len(rmi.AcceptanceCriteria) != 2 {
		t.Fatalf("expected 2 acceptance criteria, got %d", len(rmi.AcceptanceCriteria))
	}

	// duplicate
	_, err = svc.CreateRMI(ctx, "RMI-TEST-001", "github.com/test/repo", "", "", "Dup", "", "capability", "", true, 0, nil)
	if err == nil {
		t.Fatal("expected error on duplicate")
	}
}

func TestCreateRMIValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	// missing required fields
	_, err := svc.CreateRMI(ctx, "", "repo", "", "", "title", "", "capability", "", true, 0, nil)
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
	_, err = svc.CreateRMI(ctx, "RMI-X-001", "", "", "", "title", "", "capability", "", true, 0, nil)
	if err == nil {
		t.Fatal("expected error for empty repo")
	}
}

func TestGetRMI(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateRMI(ctx, "RMI-G-001", "github.com/test/repo", "", "", "Get test", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}

	rmi, err := svc.GetRMI(ctx, "RMI-G-001")
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Title != "Get test" {
		t.Fatalf("expected title 'Get test', got %s", rmi.Title)
	}

	// not found
	_, err = svc.GetRMI(ctx, "RMI-NOPE-001")
	if err == nil {
		t.Fatal("expected error for nonexistent RMI")
	}
}

func TestListRMIs(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateRMI(ctx, "RMI-L-001", "github.com/test/a", "INIT-A-001", "", "One", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-L-002", "github.com/test/b", "INIT-A-001", "", "Two", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-L-003", "github.com/test/c", "INIT-B-001", "", "Three", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}

	rmis, err := svc.ListRMIs(ctx, "INIT-A-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(rmis) != 2 {
		t.Fatalf("expected 2 RMIs for INIT-A-001, got %d", len(rmis))
	}
}

func TestListRMIsByRepo(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateRMI(ctx, "RMI-R-001", "github.com/test/repo", "", "", "One", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-R-002", "github.com/test/repo", "", "", "Two", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-R-003", "github.com/test/other", "", "", "Three", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}

	rmis, err := svc.ListRMIsByRepo(ctx, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}
	if len(rmis) != 2 {
		t.Fatalf("expected 2 RMIs for test/repo, got %d", len(rmis))
	}
}

func TestUpdateRMIStatus(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateRMI(ctx, "RMI-S-001", "github.com/test/repo", "", "", "Status test", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}

	// transition to in_progress
	rmi, err := svc.UpdateRMIStatus(ctx, "RMI-S-001", RMIStatusInProgress)
	if err != nil {
		t.Fatal(err)
	}
	if rmi.Status != RMIStatusInProgress {
		t.Fatalf("expected in_progress, got %s", rmi.Status)
	}
	if rmi.CompletedAt != nil {
		t.Fatal("CompletedAt should be nil for in_progress")
	}

	// transition to completed
	rmi, err = svc.UpdateRMIStatus(ctx, "RMI-S-001", RMIStatusCompleted)
	if err != nil {
		t.Fatal(err)
	}
	if rmi.CompletedAt == nil {
		t.Fatal("CompletedAt should be set for completed")
	}

	// invalid status
	_, err = svc.UpdateRMIStatus(ctx, "RMI-S-001", "bogus")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}

	// not found
	_, err = svc.UpdateRMIStatus(ctx, "RMI-NOPE-001", RMIStatusReady)
	if err == nil {
		t.Fatal("expected error for nonexistent RMI")
	}
}

func TestCreateDependency(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateRMI(ctx, "RMI-D-001", "github.com/test/repo", "", "", "Source", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-D-002", "github.com/test/repo", "", "", "Target", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}

	if err := svc.CreateDependency(ctx, "RMI-D-001", "RMI-D-002", "requires"); err != nil {
		t.Fatal(err)
	}

	// idempotent
	if err := svc.CreateDependency(ctx, "RMI-D-001", "RMI-D-002", "requires"); err != nil {
		t.Fatalf("expected idempotent, got %v", err)
	}

	// invalid relationship
	if err := svc.CreateDependency(ctx, "RMI-D-001", "RMI-D-002", "bogus"); err == nil {
		t.Fatal("expected error for invalid relationship")
	}

	// default relationship
	if err := svc.CreateDependency(ctx, "RMI-D-002", "RMI-D-001", ""); err != nil {
		t.Fatal(err)
	}

	deps, err := svc.ListDependencies(ctx, "RMI-D-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps))
	}
}

func TestGetRMIDetail(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateRMI(ctx, "RMI-DT-001", "github.com/test/repo", "", "", "Detail test", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-DT-002", "github.com/test/repo", "", "", "Dep target", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateDependency(ctx, "RMI-DT-001", "RMI-DT-002", "requires"); err != nil {
		t.Fatal(err)
	}

	detail, err := svc.GetRMIDetail(ctx, "RMI-DT-001")
	if err != nil {
		t.Fatal(err)
	}
	if detail.RMI.ID != "RMI-DT-001" {
		t.Fatalf("expected RMI ID RMI-DT-001, got %s", detail.RMI.ID)
	}
	if len(detail.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(detail.Dependencies))
	}
	if detail.Dependencies[0].TargetRMIID != "RMI-DT-002" {
		t.Fatalf("expected target RMI-DT-002, got %s", detail.Dependencies[0].TargetRMIID)
	}
}

func TestMoveRMI(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-A-001", "org", "Source", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-A-001/phase-1", "INIT-A-001", 1, "Phase 1", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateInitiative(ctx, "INIT-B-001", "org", "Target", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-B-001/phase-1", "INIT-B-001", 1, "Phase 1", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-MOVE-001", "github.com/test/repo", "INIT-A-001", "INIT-A-001/phase-1",
		"Movable", "", "capability", "", true, 3, nil); err != nil {
		t.Fatal(err)
	}

	moved, err := svc.MoveRMI(ctx, "RMI-MOVE-001", "INIT-B-001/phase-1", 5)
	if err != nil {
		t.Fatal(err)
	}
	if moved.InitiativeID != "INIT-B-001" {
		t.Fatalf("expected initiative INIT-B-001, got %s", moved.InitiativeID)
	}
	if moved.PhaseID != "INIT-B-001/phase-1" {
		t.Fatalf("expected phase INIT-B-001/phase-1, got %s", moved.PhaseID)
	}
	if moved.SequenceNumber != 5 {
		t.Fatalf("expected sequence 5, got %d", moved.SequenceNumber)
	}

	got, err := svc.GetRMI(ctx, "RMI-MOVE-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.InitiativeID != "INIT-B-001" || got.PhaseID != "INIT-B-001/phase-1" {
		t.Fatalf("move not persisted: initiative=%s phase=%s", got.InitiativeID, got.PhaseID)
	}

	// seq 0 leaves sequence unchanged
	if _, err := svc.MoveRMI(ctx, "RMI-MOVE-001", "INIT-A-001/phase-1", 0); err != nil {
		t.Fatal(err)
	}
	got, err = svc.GetRMI(ctx, "RMI-MOVE-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.SequenceNumber != 5 {
		t.Fatalf("expected sequence preserved at 5, got %d", got.SequenceNumber)
	}
}

func TestMoveRMIValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-A-001", "org", "Source", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-A-001/phase-1", "INIT-A-001", 1, "Phase 1", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-MOVE-002", "github.com/test/repo", "INIT-A-001", "INIT-A-001/phase-1",
		"Movable", "", "capability", "", true, 1, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.MoveRMI(ctx, "RMI-MOVE-002", "not-a-phase-id", 0); err == nil {
		t.Fatal("expected error for malformed phase ID")
	}
	if _, err := svc.MoveRMI(ctx, "RMI-MOVE-002", "INIT-MISSING-001/phase-1", 0); err == nil {
		t.Fatal("expected error for missing initiative")
	}
	if _, err := svc.MoveRMI(ctx, "RMI-MOVE-002", "INIT-A-001/phase-9", 0); err == nil {
		t.Fatal("expected error for missing phase")
	}
	if _, err := svc.MoveRMI(ctx, "RMI-MISSING-001", "INIT-A-001/phase-1", 0); err == nil {
		t.Fatal("expected error for missing RMI")
	}
}

func TestRemovePhase(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-A-001", "org", "Source", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-A-001/phase-1", "INIT-A-001", 1, "Phase 1", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-A-001/phase-2", "INIT-A-001", 2, "Phase 2", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRMI(ctx, "RMI-RP-001", "github.com/test/repo", "INIT-A-001", "INIT-A-001/phase-1",
		"Occupant", "", "capability", "", true, 1, nil); err != nil {
		t.Fatal(err)
	}

	// phase with members refuses removal
	if err := svc.RemovePhase(ctx, "INIT-A-001/phase-1"); err == nil {
		t.Fatal("expected error removing phase with member RMIs")
	}

	// empty phase removes cleanly
	if err := svc.RemovePhase(ctx, "INIT-A-001/phase-2"); err != nil {
		t.Fatal(err)
	}
	phases, err := svc.ListPhases(ctx, "INIT-A-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(phases) != 1 {
		t.Fatalf("expected 1 phase remaining, got %d", len(phases))
	}

	// malformed and missing IDs
	if err := svc.RemovePhase(ctx, "not-a-phase"); err == nil {
		t.Fatal("expected error for malformed phase ID")
	}
	if err := svc.RemovePhase(ctx, "INIT-A-001/phase-9"); err == nil {
		t.Fatal("expected error for missing phase")
	}
}
