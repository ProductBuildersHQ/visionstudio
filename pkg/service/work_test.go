package service

import (
	"context"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func setupWorkReadyTest(t *testing.T) (*Service, context.Context) {
	t.Helper()
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.CreateRMI(ctx, "RMI-W-001", "github.com/org/repo-a", "INIT-W-001", "",
		"Ready item", "", "capability", "", true, 1, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateRMIStatus(ctx, "RMI-W-001", RMIStatusReady); err != nil {
		t.Fatal(err)
	}

	return svc, ctx
}

func TestWorkReadyBasic(t *testing.T) {
	svc, ctx := setupWorkReadyTest(t)

	ready, err := svc.WorkReady(ctx, WorkReadyFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready RMI, got %d", len(ready))
	}
	if ready[0].ID != "RMI-W-001" {
		t.Fatalf("expected RMI-W-001, got %s", ready[0].ID)
	}
}

func TestWorkReadyExcludesNonReady(t *testing.T) {
	svc, ctx := setupWorkReadyTest(t)

	// Add a proposed RMI — should not appear
	if _, err := svc.CreateRMI(ctx, "RMI-W-002", "github.com/org/repo-a", "INIT-W-001", "",
		"Proposed item", "", "capability", "", true, 2, nil); err != nil {
		t.Fatal(err)
	}

	ready, err := svc.WorkReady(ctx, WorkReadyFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready RMI (excluding proposed), got %d", len(ready))
	}
}

func TestWorkReadyBlockedByDependency(t *testing.T) {
	svc, ctx := setupWorkReadyTest(t)

	// Add a dependency target that is NOT completed
	if _, err := svc.CreateRMI(ctx, "RMI-W-010", "github.com/org/repo-a", "INIT-W-001", "",
		"Dependency target", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateDependency(ctx, "RMI-W-001", "RMI-W-010", "requires"); err != nil {
		t.Fatal(err)
	}

	ready, err := svc.WorkReady(ctx, WorkReadyFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 0 {
		t.Fatalf("expected 0 ready (blocked by dep), got %d", len(ready))
	}

	// Complete the dependency target — now it should be ready
	if _, err := svc.UpdateRMIStatus(ctx, "RMI-W-010", RMIStatusCompleted); err != nil {
		t.Fatal(err)
	}

	ready, err = svc.WorkReady(ctx, WorkReadyFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready (dep completed), got %d", len(ready))
	}
}

func TestWorkReadyRelatesDoesNotBlock(t *testing.T) {
	svc, ctx := setupWorkReadyTest(t)

	// "relates" dependency should NOT block
	if _, err := svc.CreateRMI(ctx, "RMI-W-020", "github.com/org/repo-a", "INIT-W-001", "",
		"Related item", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateDependency(ctx, "RMI-W-001", "RMI-W-020", "relates"); err != nil {
		t.Fatal(err)
	}

	ready, err := svc.WorkReady(ctx, WorkReadyFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready (relates doesn't block), got %d", len(ready))
	}
}

func TestWorkReadyExcludesClaimed(t *testing.T) {
	svc, ctx := setupWorkReadyTest(t)

	// Create an active assignment
	if err := svc.Store.CreateAssignment(ctx, &store.Assignment{
		ID:     "ASSIGN-001",
		RMIID:  "RMI-W-001",
		Worker: "session-123",
		Status: "active",
	}); err != nil {
		t.Fatal(err)
	}

	ready, err := svc.WorkReady(ctx, WorkReadyFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 0 {
		t.Fatalf("expected 0 ready (claimed), got %d", len(ready))
	}
}

func TestWorkReadyFilterByInitiative(t *testing.T) {
	svc, ctx := setupWorkReadyTest(t)

	// Add a ready RMI in a different initiative
	if _, err := svc.CreateRMI(ctx, "RMI-W-030", "github.com/org/repo-b", "INIT-OTHER-001", "",
		"Other initiative", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateRMIStatus(ctx, "RMI-W-030", RMIStatusReady); err != nil {
		t.Fatal(err)
	}

	// Filter by INIT-W-001
	ready, err := svc.WorkReady(ctx, WorkReadyFilters{InitiativeID: "INIT-W-001"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready for INIT-W-001, got %d", len(ready))
	}
	if ready[0].ID != "RMI-W-001" {
		t.Fatalf("expected RMI-W-001, got %s", ready[0].ID)
	}

	// No filter should return both
	all, err := svc.WorkReady(ctx, WorkReadyFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 ready with no filter, got %d", len(all))
	}
}

func TestWorkReadyFilterByRepo(t *testing.T) {
	svc, ctx := setupWorkReadyTest(t)

	// Add a ready RMI in a different repo
	if _, err := svc.CreateRMI(ctx, "RMI-W-040", "github.com/org/repo-b", "INIT-W-001", "",
		"Other repo", "", "capability", "", true, 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateRMIStatus(ctx, "RMI-W-040", RMIStatusReady); err != nil {
		t.Fatal(err)
	}

	ready, err := svc.WorkReady(ctx, WorkReadyFilters{RepoID: "github.com/org/repo-a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 1 {
		t.Fatalf("expected 1 ready for repo-a, got %d", len(ready))
	}
	if ready[0].ID != "RMI-W-001" {
		t.Fatalf("expected RMI-W-001, got %s", ready[0].ID)
	}
}

func TestWorkReadyEmpty(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	ready, err := svc.WorkReady(ctx, WorkReadyFilters{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ready) != 0 {
		t.Fatalf("expected 0 ready on empty store, got %d", len(ready))
	}
}
