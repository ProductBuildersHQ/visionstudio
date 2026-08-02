package store

import (
	"context"
	"testing"
	"time"
)

func TestMemStoreInitiativeCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()
	now := time.Now()

	init := &Initiative{
		ID:           "INIT-TEST-001",
		Organization: "test",
		Title:        "Test Initiative",
		Status:       "proposed",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.CreateInitiative(ctx, init); err != nil {
		t.Fatal(err)
	}

	// duplicate
	if err := s.CreateInitiative(ctx, init); err == nil {
		t.Fatal("expected error on duplicate")
	}

	got, err := s.GetInitiative(ctx, "INIT-TEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Test Initiative" {
		t.Fatalf("expected title 'Test Initiative', got %q", got.Title)
	}

	// not found
	_, err = s.GetInitiative(ctx, "INIT-NOPE-001")
	if err == nil {
		t.Fatal("expected not found error")
	}

	got.Status = "executing"
	if err := s.UpdateInitiative(ctx, got); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListInitiatives(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Status != "executing" {
		t.Fatalf("unexpected list result: %+v", list)
	}
}

func TestMemStoreProgramCRUD(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()
	now := time.Now()

	prog := &Program{
		ID:           "PROG-DELIVERY",
		Name:         "Product Delivery",
		Organization: "default",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.CreateProgram(ctx, prog); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateProgram(ctx, prog); err == nil {
		t.Fatal("expected error on duplicate")
	}

	got, err := s.GetProgram(ctx, "PROG-DELIVERY")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Product Delivery" {
		t.Fatalf("expected name 'Product Delivery', got %q", got.Name)
	}

	_, err = s.GetProgram(ctx, "PROG-NOPE")
	if err == nil {
		t.Fatal("expected not found error")
	}

	got.Description = "Updated description"
	if err := s.UpdateProgram(ctx, got); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListPrograms(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Description != "Updated description" {
		t.Fatalf("unexpected list result: %+v", list)
	}
}

func TestMemStoreAssignmentActiveQuery(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()
	now := time.Now()

	a := &Assignment{
		ID:             "ASSIGN-001",
		RMIID:          "RMI-A-001",
		Worker:         "session-1",
		Status:         "active",
		LeaseExpiresAt: now.Add(4 * time.Hour),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.CreateAssignment(ctx, a); err != nil {
		t.Fatal(err)
	}

	active, err := s.GetActiveAssignment(ctx, "RMI-A-001")
	if err != nil {
		t.Fatal(err)
	}
	if active == nil || active.ID != "ASSIGN-001" {
		t.Fatal("expected active assignment")
	}

	none, err := s.GetActiveAssignment(ctx, "RMI-A-999")
	if err != nil {
		t.Fatal(err)
	}
	if none != nil {
		t.Fatal("expected nil for no active assignment")
	}
}

func TestMemStoreUnitOfWork(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()
	uow := &MemUnitOfWork{Store: s}
	now := time.Now()

	err := uow.Execute(ctx, func(ctx context.Context, s Store) error {
		return s.CreateInitiative(ctx, &Initiative{
			ID:           "INIT-UOW-001",
			Organization: "test",
			Title:        "UoW Test",
			Status:       "proposed",
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.GetInitiative(ctx, "INIT-UOW-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "UoW Test" {
		t.Fatalf("expected 'UoW Test', got %q", got.Title)
	}
}

func TestMemStoreRMIDependencies(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()

	dep := &RMIDependency{
		SourceRMIID:  "RMI-A-002",
		TargetRMIID:  "RMI-A-001",
		Relationship: "requires",
	}
	if err := s.CreateDependency(ctx, dep); err != nil {
		t.Fatal(err)
	}

	// duplicate
	if err := s.CreateDependency(ctx, dep); err == nil {
		t.Fatal("expected error on duplicate dependency")
	}

	deps, err := s.ListDependencies(ctx, "RMI-A-002")
	if err != nil {
		t.Fatal(err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(deps))
	}
}

func TestMemStoreEvidenceByInitiative(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()
	now := time.Now()

	if err := s.CreateRMI(ctx, &RoadmapItem{
		ID:           "RMI-A-001",
		RepositoryID: "repo-a",
		InitiativeID: "INIT-TEST-001",
		Title:        "Item 1",
		ItemType:     "capability",
		Status:       "completed",
		Required:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateEvidence(ctx, &DeliveryEvidence{
		ID:           "EV-001",
		RMIID:        "RMI-A-001",
		EvidenceType: "commit",
		Reference:    "abc123",
		CommitType:   "feat",
		CreatedAt:    now,
	}); err != nil {
		t.Fatal(err)
	}

	evs, err := s.ListEvidenceByInitiative(ctx, "INIT-TEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 || evs[0].ID != "EV-001" {
		t.Fatalf("unexpected evidence result: %+v", evs)
	}

	evs, err = s.ListEvidenceByInitiative(ctx, "INIT-NOPE-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 0 {
		t.Fatal("expected empty for nonexistent initiative")
	}
}

func TestMemStoreListRepositoriesByOrg(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()

	repos := []*Repository{
		{ID: "github.com/plexusone/omnidevx", Organization: "plexusone", RepositoryName: "omnidevx", Status: "active"},
		{ID: "github.com/plexusone/devfolio", Organization: "plexusone", RepositoryName: "devfolio", Status: "active"},
		{ID: "github.com/grokify/gogit", Organization: "grokify", RepositoryName: "gogit", Status: "active"},
	}
	for _, r := range repos {
		if err := s.CreateRepository(ctx, r); err != nil {
			t.Fatal(err)
		}
	}

	plexus, err := s.ListRepositoriesByOrg(ctx, "plexusone")
	if err != nil {
		t.Fatal(err)
	}
	if len(plexus) != 2 {
		t.Fatalf("expected 2 plexusone repos, got %d", len(plexus))
	}

	grokify, err := s.ListRepositoriesByOrg(ctx, "grokify")
	if err != nil {
		t.Fatal(err)
	}
	if len(grokify) != 1 {
		t.Fatalf("expected 1 grokify repo, got %d", len(grokify))
	}
}

func TestMemStoreRepoDependencies(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore()

	dep := &RepositoryDependency{
		SourceRepositoryID: "github.com/plexusone/omnidevx",
		TargetRepositoryID: "github.com/plexusone/omnidevx-core",
		DependencyType:     "go_module",
	}
	if err := s.CreateRepoDependency(ctx, dep); err != nil {
		t.Fatal(err)
	}

	// duplicate
	if err := s.CreateRepoDependency(ctx, dep); err == nil {
		t.Fatal("expected error on duplicate repo dependency")
	}

	deps, err := s.ListRepoDependencies(ctx, "github.com/plexusone/omnidevx")
	if err != nil {
		t.Fatal(err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(deps))
	}

	// list by target too
	deps, err = s.ListRepoDependencies(ctx, "github.com/plexusone/omnidevx-core")
	if err != nil {
		t.Fatal(err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency for target, got %d", len(deps))
	}

	all, err := s.ListAllRepoDependencies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 total dependency, got %d", len(all))
	}
}
