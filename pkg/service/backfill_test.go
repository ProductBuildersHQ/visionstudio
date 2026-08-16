package service

import (
	"context"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func seedBackfillFixtures(t *testing.T, s *store.MemStore) {
	t.Helper()
	ctx := context.Background()
	if err := s.CreateRepository(ctx, &store.Repository{ID: "github.com/x/r", Organization: "x", RepositoryName: "r", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateInitiative(ctx, &store.Initiative{ID: "INIT-HOME", Title: "Home Repo Initiative", Status: "executing", Organization: "default", HomeRepo: "github.com/x/r"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateInitiative(ctx, &store.Initiative{ID: "INIT-RMI-ONLY", Title: "RMI-Referenced Initiative", Status: "executing", Organization: "default"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateInitiative(ctx, &store.Initiative{ID: "INIT-UNRELATED", Title: "Different Repo", Status: "executing", Organization: "default", HomeRepo: "github.com/x/other"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRMI(ctx, &store.RoadmapItem{ID: "RMI-X-001", RepositoryID: "github.com/x/r", InitiativeID: "INIT-RMI-ONLY", Title: "Do the thing", ItemType: "capability", Status: "completed"}); err != nil {
		t.Fatal(err)
	}
}

func TestUnmatchedReleases(t *testing.T) {
	s := store.NewMemStore()
	seedBackfillFixtures(t, s)
	svc := New(s)
	ctx := context.Background()

	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := s.CreateRelease(ctx, &store.Release{ID: "github.com/x/r@v2", RepositoryID: "github.com/x/r", Tag: "v2", ReleasedAt: recent}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRelease(ctx, &store.Release{ID: "github.com/x/r@v1", RepositoryID: "github.com/x/r", Tag: "v1", ReleasedAt: old}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRelease(ctx, &store.Release{ID: "github.com/x/r@v3-matched", RepositoryID: "github.com/x/r", Tag: "v3", ReleasedAt: recent, InitiativeIDs: []string{"INIT-HOME"}}); err != nil {
		t.Fatal(err)
	}

	unmatched, err := svc.UnmatchedReleases(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(unmatched) != 2 {
		t.Fatalf("unmatched = %d, want 2 (v3 is matched, excluded)", len(unmatched))
	}
	// Oldest first.
	if unmatched[0].Tag != "v1" {
		t.Fatalf("first = %s, want v1 (oldest)", unmatched[0].Tag)
	}
}

func TestGetBackfillCandidatesIncludesHomeAndRMIReferenced(t *testing.T) {
	s := store.NewMemStore()
	seedBackfillFixtures(t, s)
	svc := New(s)
	ctx := context.Background()

	rel := &store.Release{ID: "github.com/x/r@v1", RepositoryID: "github.com/x/r", Tag: "v1", ReleasedAt: time.Now(), Body: "notes"}
	if err := s.CreateRelease(ctx, rel); err != nil {
		t.Fatal(err)
	}

	bc, err := svc.GetBackfillCandidates(ctx, rel.ID)
	if err != nil {
		t.Fatal(err)
	}
	if bc.Release.ID != rel.ID {
		t.Fatalf("release = %s", bc.Release.ID)
	}
	if len(bc.Candidates) != 2 {
		t.Fatalf("candidates = %d, want 2 (home-repo + RMI-referenced, NOT the unrelated-repo one)", len(bc.Candidates))
	}
	ids := map[string]bool{}
	for _, c := range bc.Candidates {
		ids[c.Initiative.ID] = true
	}
	if !ids["INIT-HOME"] || !ids["INIT-RMI-ONLY"] {
		t.Fatalf("candidates = %v, want INIT-HOME and INIT-RMI-ONLY", ids)
	}
	if ids["INIT-UNRELATED"] {
		t.Fatal("unrelated-repo initiative must not appear as a candidate")
	}

	// RMI titles attached for the RMI-referenced candidate.
	for _, c := range bc.Candidates {
		if c.Initiative.ID == "INIT-RMI-ONLY" && len(c.RMITitles) != 1 {
			t.Fatalf("RMI titles = %v, want 1", c.RMITitles)
		}
	}
}

func TestGetBackfillCandidatesNoCandidatesErrors(t *testing.T) {
	s := store.NewMemStore()
	ctx := context.Background()
	if err := s.CreateRepository(ctx, &store.Repository{ID: "github.com/x/lonely", Organization: "x", RepositoryName: "lonely", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRelease(ctx, &store.Release{ID: "github.com/x/lonely@v1", RepositoryID: "github.com/x/lonely", Tag: "v1", ReleasedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	svc := New(s)
	if _, err := svc.GetBackfillCandidates(ctx, "github.com/x/lonely@v1"); err == nil {
		t.Fatal("expected error when no candidate initiatives exist")
	}
}

func TestGetBackfillCandidatesUnknownRelease(t *testing.T) {
	svc := New(store.NewMemStore())
	if _, err := svc.GetBackfillCandidates(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("expected error for unknown release")
	}
}
