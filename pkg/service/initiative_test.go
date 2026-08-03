package service

import (
	"context"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestCreateInitiative(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	init, err := svc.CreateInitiative(ctx, "INIT-TEST-001", "test", "Test Initiative", "A test", "high", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if init.Status != initiative.StatusProposed {
		t.Fatalf("expected status proposed, got %s", init.Status)
	}
	if init.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}

	// duplicate
	_, err = svc.CreateInitiative(ctx, "INIT-TEST-001", "test", "Dup", "", "", "", "")
	if err == nil {
		t.Fatal("expected error on duplicate")
	}
}

func TestListInitiatives(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-A-001", "test", "A", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateInitiative(ctx, "INIT-B-001", "test", "B", "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	inits, err := svc.ListInitiatives(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(inits) != 2 {
		t.Fatalf("expected 2 initiatives, got %d", len(inits))
	}
}

func TestTransitionInitiative(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-T-001", "test", "Transition Test", "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	// proposed -> planned
	init, err := svc.TransitionInitiative(ctx, "INIT-T-001", "planned")
	if err != nil {
		t.Fatal(err)
	}
	if init.Status != "planned" {
		t.Fatalf("expected planned, got %s", init.Status)
	}
	if init.PlannedAt == nil {
		t.Fatal("expected PlannedAt to be set")
	}

	// planned -> executing
	init, err = svc.TransitionInitiative(ctx, "INIT-T-001", "executing")
	if err != nil {
		t.Fatal(err)
	}
	if init.ExecutingAt == nil {
		t.Fatal("expected ExecutingAt to be set")
	}

	// invalid: executing -> proposed
	_, err = svc.TransitionInitiative(ctx, "INIT-T-001", "proposed")
	if err == nil {
		t.Fatal("expected error on invalid transition")
	}
}

func TestTransitionInitiativeNotFound(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, err := svc.TransitionInitiative(ctx, "INIT-NOPE-001", "planned")
	if err == nil {
		t.Fatal("expected error for nonexistent initiative")
	}
}

func TestGetInitiativeDetail(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-D-001", "test", "Detail Test", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-D-001/phase-1", "INIT-D-001", 1, "Foundation", "Build the base"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-D-001/phase-2", "INIT-D-001", 2, "Features", "Add features"); err != nil {
		t.Fatal(err)
	}

	// Add an RMI to phase 1
	if err := svc.Store.CreateRMI(ctx, &store.RoadmapItem{
		ID:           "RMI-TEST-001",
		RepositoryID: "github.com/test/repo",
		InitiativeID: "INIT-D-001",
		PhaseID:      "INIT-D-001/phase-1",
		Title:        "First item",
		ItemType:     "capability",
		Status:       "completed",
		Required:     true,
	}); err != nil {
		t.Fatal(err)
	}

	detail, err := svc.GetInitiativeDetail(ctx, "INIT-D-001")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Initiative.ID != "INIT-D-001" {
		t.Fatalf("unexpected initiative ID: %s", detail.Initiative.ID)
	}
	if len(detail.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(detail.Phases))
	}

	// Phase 1 should be completed (1 required RMI completed)
	var phase1 *PhaseDetail
	for i := range detail.Phases {
		if detail.Phases[i].Phase.SequenceNumber == 1 {
			phase1 = &detail.Phases[i]
			break
		}
	}
	if phase1 == nil {
		t.Fatal("phase 1 not found")
	}
	if phase1.Status != initiative.PhaseCompleted {
		t.Fatalf("expected phase 1 status completed, got %s", phase1.Status)
	}
	if len(phase1.RMIs) != 1 {
		t.Fatalf("expected 1 RMI in phase 1, got %d", len(phase1.RMIs))
	}

	// Phase 2 should be planned (no RMIs)
	var phase2 *PhaseDetail
	for i := range detail.Phases {
		if detail.Phases[i].Phase.SequenceNumber == 2 {
			phase2 = &detail.Phases[i]
			break
		}
	}
	if phase2 == nil {
		t.Fatal("phase 2 not found")
	}
	if phase2.Status != initiative.PhasePlanned {
		t.Fatalf("expected phase 2 status planned, got %s", phase2.Status)
	}
}

func TestCreatePhase(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-P-001", "test", "Phase Test", "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	p, err := svc.CreatePhase(ctx, "INIT-P-001/phase-1", "INIT-P-001", 1, "Foundation", "Build the base")
	if err != nil {
		t.Fatal(err)
	}
	if p.InitiativeID != "INIT-P-001" {
		t.Fatalf("expected initiative ID INIT-P-001, got %s", p.InitiativeID)
	}

	// nonexistent initiative
	_, err = svc.CreatePhase(ctx, "INIT-NOPE-001/phase-1", "INIT-NOPE-001", 1, "Bad", "")
	if err == nil {
		t.Fatal("expected error for nonexistent initiative")
	}
}

func TestListPhases(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateInitiative(ctx, "INIT-LP-001", "test", "List Phases", "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-LP-001/phase-1", "INIT-LP-001", 1, "One", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePhase(ctx, "INIT-LP-001/phase-2", "INIT-LP-001", 2, "Two", ""); err != nil {
		t.Fatal(err)
	}

	phases, err := svc.ListPhases(ctx, "INIT-LP-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(phases))
	}
}
