// Package storetest holds behaviour tests that every store.Store
// implementation must pass, so the in-memory fake and the Dolt store are held
// to the same contract.
package storetest

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

const (
	oldRepo   = "github.com/acme/old-name"
	newRepo   = "github.com/acme/new-name"
	otherRepo = "github.com/acme/other"
)

// seedRename registers three repositories and, on the one being renamed, one
// of everything that can reference a repository.
func seedRename(t *testing.T, s store.Store) {
	t.Helper()
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	must := func(what string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("seed %s: %v", what, err)
		}
	}
	must("organization", s.CreateOrganization(ctx, &store.Organization{ID: "github.com/acme", Login: "acme", Kind: "organization", CreatedAt: now, UpdatedAt: now}))
	for _, id := range []string{oldRepo, otherRepo} {
		must("repository", s.CreateRepository(ctx, &store.Repository{
			ID: id, Organization: "acme", RepositoryName: id[len("github.com/acme/"):], DefaultBranch: "main",
			Status: "active", OrganizationID: "github.com/acme", Visibility: "public",
			LocalPath: "/src/" + id, GoModule: id,
		}))
	}
	must("initiative", s.CreateInitiative(ctx, &store.Initiative{ID: "INIT-OLDNAME-001", Organization: "acme", Title: "Work", Status: "executing", HomeRepo: oldRepo, CreatedAt: now, UpdatedAt: now}))
	must("other initiative", s.CreateInitiative(ctx, &store.Initiative{ID: "INIT-OTHER-001", Organization: "acme", Title: "Other", Status: "executing", HomeRepo: otherRepo, CreatedAt: now, UpdatedAt: now}))
	for _, id := range []string{"RMI-OLDNAME-001", "RMI-OLDNAME-002"} {
		must("RMI", s.CreateRMI(ctx, &store.RoadmapItem{ID: id, RepositoryID: oldRepo, InitiativeID: "INIT-OLDNAME-001", Title: id, ItemType: "capability", Status: "completed", CreatedAt: now, UpdatedAt: now}))
	}
	must("other RMI", s.CreateRMI(ctx, &store.RoadmapItem{ID: "RMI-OTHER-001", RepositoryID: otherRepo, InitiativeID: "INIT-OTHER-001", Title: "x", ItemType: "capability", Status: "proposed", CreatedAt: now, UpdatedAt: now}))
	must("spec document", s.CreateSpecDocument(ctx, &store.SpecDocument{ID: "spec-prd", RepositoryID: oldRepo, InitiativeID: "INIT-OLDNAME-001", SpecType: "prd", FilePath: "docs/PRD.md", SyncedAt: now, CreatedAt: now, UpdatedAt: now}))
	must("release", s.CreateRelease(ctx, &store.Release{ID: oldRepo + "@v0.1.0", RepositoryID: oldRepo, Tag: "v0.1.0", ReleasedAt: now, CreatedAt: now, UpdatedAt: now}))
	// Dependencies: one out, one in, one that becomes a duplicate, and one that
	// becomes a self-reference once the old ID is replaced.
	must("dependency out", s.CreateRepoDependency(ctx, &store.RepositoryDependency{SourceRepositoryID: oldRepo, TargetRepositoryID: otherRepo, DependencyType: "go_module"}))
	must("dependency in", s.CreateRepoDependency(ctx, &store.RepositoryDependency{SourceRepositoryID: otherRepo, TargetRepositoryID: oldRepo, DependencyType: "go_module"}))
	must("PRISM roadmap", s.CreatePRISMRoadmap(ctx, &store.PRISMRoadmap{ID: "prm-1", RepositoryID: oldRepo, CreatedAt: now, UpdatedAt: now}))
	must("PRISM goal", s.CreatePRISMGoal(ctx, &store.PRISMGoal{ID: "pg-1", RepositoryID: oldRepo, GoalType: "okr", CreatedAt: now, UpdatedAt: now}))
	must("PRISM document", s.CreatePRISMDocument(ctx, &store.PRISMDocument{ID: "pd-1", RepositoryID: oldRepo, Name: "doc", CreatedAt: now, UpdatedAt: now}))
	must("person", s.CreatePerson(ctx, &store.Person{ID: "person:dev", GitHubLogin: "dev", CreatedAt: now, UpdatedAt: now}))
	must("DevX report", s.CreateDevXPeriodReport(ctx, &store.DevXPeriodReport{ID: "dx-1", RepositoryID: oldRepo, PersonID: "person:dev", PeriodType: "week", PeriodLabel: "2026-W44", PeriodStart: now, PeriodEnd: now.AddDate(0, 0, 7), CreatedAt: now}))
}

func renameTarget(t *testing.T, s store.Store) (*store.Repository, *store.Repository) {
	t.Helper()
	from, err := s.GetRepository(context.Background(), oldRepo)
	if err != nil {
		t.Fatal(err)
	}
	to := *from
	to.ID, to.RepositoryName, to.Status, to.SupersededBy = newRepo, "new-name", "active", ""
	to.GoModule, to.LocalPath = newRepo, "/src/"+newRepo
	return from, &to
}

// RenameRepository checks that a rename moves every reference in one step,
// leaves releases unless asked, archives the old record, and that a dry run
// counts the same and changes nothing.
func RenameRepository(t *testing.T, s store.Store) {
	count := counter(t)
	ctx := context.Background()
	seedRename(t, s)
	want := store.RenameRepositoryResult{
		Created: true, RMIs: 2, SpecDocuments: 1, Initiatives: 1, Dependencies: 2,
		PRISMRoadmaps: 1, PRISMGoals: 1, PRISMDocuments: 1, DevXReports: 1,
	}

	from, to := renameTarget(t, s)
	dry, err := s.RenameRepository(ctx, from, to, store.RenameRepositoryOptions{DryRun: true})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if *dry != want {
		t.Errorf("dry run counts = %+v, want %+v", *dry, want)
	}
	if _, err := s.GetRepository(ctx, newRepo); err == nil {
		t.Error("a dry run registered the new repository")
	}
	if rmis := count(length(s.ListRMIsByRepo(ctx, oldRepo))); rmis != 2 {
		t.Errorf("a dry run moved RMIs: %d left on the old ID", rmis)
	}

	from, to = renameTarget(t, s)
	got, err := s.RenameRepository(ctx, from, to, store.RenameRepositoryOptions{})
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if *got != want {
		t.Errorf("counts = %+v, want %+v", *got, want)
	}

	old, err := s.GetRepository(ctx, oldRepo)
	if err != nil || old.Status != "archived" || old.SupersededBy != newRepo {
		t.Errorf("old record = %+v (%v), want archived and superseded by %s", old, err, newRepo)
	}
	nw, err := s.GetRepository(ctx, newRepo)
	if err != nil || nw.Status != "active" || nw.GoModule != newRepo || nw.LocalPath != "/src/"+newRepo || nw.Visibility != "public" {
		t.Errorf("new record = %+v (%v)", nw, err)
	}
	if rmis := count(length(s.ListRMIsByRepo(ctx, newRepo))); rmis != 2 {
		t.Errorf("%d RMIs on the new ID, want 2", rmis)
	}
	if rmis := count(length(s.ListRMIsByRepo(ctx, otherRepo))); rmis != 1 {
		t.Errorf("an unrelated RMI moved: %d left on %s", rmis, otherRepo)
	}
	if docs := count(length(s.ListSpecDocumentsByRepo(ctx, newRepo))); docs != 1 {
		t.Errorf("%d spec documents on the new ID, want 1", docs)
	}
	if init := initiative(t, s, "INIT-OLDNAME-001"); init.HomeRepo != newRepo {
		t.Errorf("initiative home repo = %+v", init)
	}
	if init := initiative(t, s, "INIT-OTHER-001"); init.HomeRepo != otherRepo {
		t.Errorf("an unrelated initiative moved: %+v", init)
	}
	if rels := count(length(s.ListReleasesByRepo(ctx, oldRepo))); rels != 1 {
		t.Errorf("releases should stay on the old ID by default, %d there", rels)
	}
	deps, err := s.ListRepoDependencies(ctx, newRepo)
	if err != nil {
		t.Fatal(err)
	}
	var pairs []string
	for _, d := range deps {
		pairs = append(pairs, d.SourceRepositoryID+">"+d.TargetRepositoryID)
	}
	slices.Sort(pairs)
	if !slices.Equal(pairs, []string{newRepo + ">" + otherRepo, otherRepo + ">" + newRepo}) {
		t.Errorf("dependencies = %v", pairs)
	}
	if left := count(length(s.ListRepoDependencies(ctx, oldRepo))); left != 0 {
		t.Errorf("%d dependencies still name the old ID", left)
	}
	if r := count(length(s.ListPRISMRoadmapsByRepo(ctx, newRepo))); r != 1 {
		t.Errorf("%d PRISM roadmaps on the new ID", r)
	}
	if g := count(length(s.ListPRISMGoals(ctx, newRepo))); g != 1 {
		t.Errorf("%d PRISM goals on the new ID", g)
	}
	if d := count(length(s.ListPRISMDocumentsByRepo(ctx, newRepo))); d != 1 {
		t.Errorf("%d PRISM documents on the new ID", d)
	}
	if r := count(length(s.ListDevXPeriodReportsByRepo(ctx, newRepo))); r != 1 {
		t.Errorf("%d DevX reports on the new ID", r)
	}
}

// RenameRepositoryFinishesAPartialRename checks the path where the new ID was
// registered and RMIs were moved by hand first: the rename reuses the record,
// moves what is left, and can move releases on request.
func RenameRepositoryFinishesAPartialRename(t *testing.T, s store.Store) {
	count := counter(t)
	ctx := context.Background()
	seedRename(t, s)
	from, to := renameTarget(t, s)
	to.LocalPath = "/src/registered-by-hand"
	if err := s.CreateRepository(ctx, to); err != nil {
		t.Fatal(err)
	}
	// A dependency that already exists on the new ID makes the moved one a
	// duplicate, and an edge between the two IDs becomes a self-reference.
	if err := s.CreateRepoDependency(ctx, &store.RepositoryDependency{SourceRepositoryID: newRepo, TargetRepositoryID: otherRepo, DependencyType: "go_module"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRepoDependency(ctx, &store.RepositoryDependency{SourceRepositoryID: oldRepo, TargetRepositoryID: newRepo, DependencyType: "go_module"}); err != nil {
		t.Fatal(err)
	}

	got, err := s.RenameRepository(ctx, from, to, store.RenameRepositoryOptions{MoveReleases: true})
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if got.Created || got.Releases != 1 || got.DependenciesDropped != 2 || got.Dependencies != 1 {
		t.Errorf("counts = %+v, want the record reused, 1 release moved, 2 dependencies dropped, 1 moved", *got)
	}
	nw, err := s.GetRepository(ctx, newRepo)
	if err != nil {
		t.Fatal(err)
	}
	if nw.LocalPath != "/src/registered-by-hand" {
		t.Errorf("the existing record was overwritten: %+v", nw)
	}
	if rels := count(length(s.ListReleasesByRepo(ctx, newRepo))); rels != 1 {
		t.Errorf("%d releases on the new ID, want 1", rels)
	}
	deps, err := s.ListRepoDependencies(ctx, newRepo)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range deps {
		if d.SourceRepositoryID == d.TargetRepositoryID {
			t.Errorf("self-reference survived: %+v", d)
		}
	}
	if len(deps) != 2 {
		t.Errorf("%d dependencies on the new ID, want 2: %+v", len(deps), deps)
	}
}

// length turns a list lookup into its length, keeping the error.
func length[T any](items []T, err error) (int, error) { return len(items), err }

// counter returns a function that fails the test on a lookup error and
// otherwise returns the count, so count(length(s.ListX(...))) reads as one call.
func counter(t *testing.T) func(int, error) int {
	return func(n int, err error) int {
		t.Helper()
		if err != nil {
			t.Fatalf("lookup: %v", err)
		}
		return n
	}
}

func initiative(t *testing.T, s store.Store, id string) *store.Initiative {
	t.Helper()
	init, err := s.GetInitiative(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return init
}
