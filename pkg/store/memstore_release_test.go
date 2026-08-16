package store

import (
	"context"
	"testing"
	"time"
)

func TestMemStoreReleaseCRUD(t *testing.T) {
	s := NewMemStore()
	ctx := context.Background()
	t0 := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	rels := []*Release{
		{
			ID:           "github.com/grokify/releaselog@v0.1.0",
			RepositoryID: "github.com/grokify/releaselog",
			Tag:          "v0.1.0",
			ReleasedAt:   t0,
		},
		{
			ID:            "github.com/ProductBuildersHQ/visionstudio@v0.3.0",
			RepositoryID:  "github.com/ProductBuildersHQ/visionstudio",
			Tag:           "v0.3.0",
			ReleasedAt:    t1,
			InitiativeIDs: []string{"INIT-VISIONSTUDIO-007"},
			RMIIDs:        []string{"RMI-VISIONSTUDIO-401"},
		},
	}
	for _, r := range rels {
		if err := s.CreateRelease(ctx, r); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	if err := s.CreateRelease(ctx, rels[0]); err == nil {
		t.Fatal("expected duplicate error")
	}

	got, err := s.GetRelease(ctx, rels[1].ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Tag != "v0.3.0" || len(got.InitiativeIDs) != 1 {
		t.Fatalf("got = %+v", got)
	}

	all, err := s.ListReleases(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 2 || all[0].ReleasedAt.Before(all[1].ReleasedAt) {
		t.Fatalf("list not newest-first: %v, %v", all[0].ID, all[1].ID)
	}

	byRepo, err := s.ListReleasesByRepo(ctx, "github.com/grokify/releaselog")
	if err != nil {
		t.Fatalf("by repo: %v", err)
	}
	if len(byRepo) != 1 || byRepo[0].Tag != "v0.1.0" {
		t.Fatalf("by repo = %+v", byRepo)
	}

	byInit, err := s.ListReleasesByInitiative(ctx, "INIT-VISIONSTUDIO-007")
	if err != nil {
		t.Fatalf("by initiative: %v", err)
	}
	if len(byInit) != 1 || byInit[0].ID != rels[1].ID {
		t.Fatalf("by initiative = %+v", byInit)
	}
	if empty, _ := s.ListReleasesByInitiative(ctx, "INIT-NONE"); len(empty) != 0 {
		t.Fatalf("expected no releases for unknown initiative")
	}

	rels[1].URL = "https://github.com/ProductBuildersHQ/visionstudio/releases/tag/v0.3.0"
	if err := s.UpdateRelease(ctx, rels[1]); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.UpdateRelease(ctx, &Release{ID: "missing@v0"}); err == nil {
		t.Fatal("expected not-found on update")
	}
}
