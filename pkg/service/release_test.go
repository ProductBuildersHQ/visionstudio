package service

import (
	"context"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func seedReleaseFixtures(t *testing.T, s *store.MemStore) {
	t.Helper()
	ctx := context.Background()
	repos := []*store.Repository{
		{ID: "github.com/plexusone/omnidevx", Organization: "plexusone", RepositoryName: "omnidevx", Status: "active"},
	}
	for _, r := range repos {
		if err := s.CreateRepository(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.CreateInitiative(ctx, &store.Initiative{ID: "INIT-X-001", Title: "X", Status: "executing", Organization: "default"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRMI(ctx, &store.RoadmapItem{ID: "RMI-X-001", RepositoryID: "github.com/plexusone/omnidevx", Title: "x", ItemType: "capability", Status: "completed"}); err != nil {
		t.Fatal(err)
	}
}

func TestRecordReleaseUpsert(t *testing.T) {
	s := store.NewMemStore()
	seedReleaseFixtures(t, s)
	svc := New(s)
	ctx := context.Background()
	day := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	rel, err := svc.RecordRelease(ctx, "github.com/plexusone/omnidevx", "v1.2.0", day, "", "", []string{"INIT-X-001"}, nil)
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if rel.ID != "github.com/plexusone/omnidevx@v1.2.0" {
		t.Fatalf("ID = %s", rel.ID)
	}

	// Upsert merges associations, never drops them.
	rel2, err := svc.RecordRelease(ctx, "github.com/plexusone/omnidevx", "v1.2.0", day, "https://example/rel", "", nil, []string{"RMI-X-001"})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if len(rel2.InitiativeIDs) != 1 || len(rel2.RMIIDs) != 1 {
		t.Fatalf("associations lost on upsert: %+v", rel2)
	}
	if rel2.URL == "" {
		t.Fatal("URL not applied on upsert")
	}

	// Unregistered repo refused.
	if _, err := svc.RecordRelease(ctx, "github.com/nope/nope", "v1", day, "", "", nil, nil); err == nil {
		t.Fatal("expected unregistered-repo error")
	}
}

func TestAttachDetachRelease(t *testing.T) {
	s := store.NewMemStore()
	seedReleaseFixtures(t, s)
	svc := New(s)
	ctx := context.Background()

	rel, err := svc.RecordRelease(ctx, "github.com/plexusone/omnidevx", "v1.3.0", time.Time{}, "", "", nil, nil)
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	if _, err := svc.AttachRelease(ctx, rel.ID, []string{"INIT-MISSING"}, nil); err == nil {
		t.Fatal("expected error for unknown initiative")
	}

	got, err := svc.AttachRelease(ctx, rel.ID, []string{"INIT-X-001"}, []string{"RMI-X-001"})
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	if len(got.InitiativeIDs) != 1 || len(got.RMIIDs) != 1 {
		t.Fatalf("attach result: %+v", got)
	}

	got, err = svc.DetachRelease(ctx, rel.ID, []string{"INIT-X-001"}, nil)
	if err != nil {
		t.Fatalf("detach: %v", err)
	}
	if len(got.InitiativeIDs) != 0 || len(got.RMIIDs) != 1 {
		t.Fatalf("detach result: %+v", got)
	}
}

func TestListReleasesFilters(t *testing.T) {
	s := store.NewMemStore()
	seedReleaseFixtures(t, s)
	svc := New(s)
	ctx := context.Background()

	if _, err := svc.RecordRelease(ctx, "github.com/plexusone/omnidevx", "v1.0.0", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "", "", []string{"INIT-X-001"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RecordRelease(ctx, "github.com/plexusone/omnidevx", "v1.1.0", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), "", "", nil, nil); err != nil {
		t.Fatal(err)
	}

	all, err := svc.ListReleases(ctx, "", "")
	if err != nil || len(all) != 2 {
		t.Fatalf("all = %d (%v), want 2", len(all), err)
	}
	byInit, err := svc.ListReleases(ctx, "", "INIT-X-001")
	if err != nil || len(byInit) != 1 {
		t.Fatalf("byInit = %d (%v), want 1", len(byInit), err)
	}
	both, err := svc.ListReleases(ctx, "github.com/plexusone/omnidevx", "INIT-X-001")
	if err != nil || len(both) != 1 {
		t.Fatalf("both = %d (%v), want 1", len(both), err)
	}
	none, err := svc.ListReleases(ctx, "github.com/other/repo", "INIT-X-001")
	if err != nil || len(none) != 0 {
		t.Fatalf("none = %d (%v), want 0", len(none), err)
	}
}
