package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func fakeLookup(releases []GHRelease) GitHubReleasesLookup {
	return func(_ context.Context, _, _ string) ([]GHRelease, error) {
		return releases, nil
	}
}

func TestGitHubReleasesGapFillsAndSkipsDrafts(t *testing.T) {
	s := store.NewMemStore()
	svc := service.New(s)
	ctx := context.Background()
	repoID := "github.com/plexusone/omniskill"
	if err := s.CreateRepository(ctx, &store.Repository{ID: repoID, Organization: "plexusone", RepositoryName: "omniskill", Status: "active"}); err != nil {
		t.Fatal(err)
	}

	pub := time.Date(2026, 1, 11, 0, 10, 53, 0, time.UTC)
	lookup := fakeLookup([]GHRelease{
		{TagName: "v0.1.0", PublishedAt: pub, HTMLURL: "https://github.com/plexusone/omniskill/releases/tag/v0.1.0", Body: "Release Notes: v0.1.0"},
		{TagName: "v0.13.0-draft", Draft: true, HTMLURL: "https://example/draft"},
	})

	res, err := GitHubReleases(ctx, svc, repoID, lookup)
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if res.Fetched != 2 || res.GapFilled != 1 || res.DraftsSkipped != 1 {
		t.Fatalf("result: %+v", res)
	}

	rel, err := s.GetRelease(ctx, service.ReleaseID(repoID, "v0.1.0"))
	if err != nil {
		t.Fatalf("gap-filled release not created: %v", err)
	}
	if rel.URL == "" || rel.Body == "" {
		t.Fatalf("gap-filled release missing evidence: %+v", rel)
	}
	if len(rel.InitiativeIDs) != 0 || len(rel.RMIIDs) != 0 {
		t.Fatalf("gap-filled release must have NO associations (no trailer evidence): %+v", rel)
	}

	// Draft never created.
	if _, err := s.GetRelease(ctx, service.ReleaseID(repoID, "v0.13.0-draft")); err == nil {
		t.Fatal("draft release must not be created")
	}
}

func TestGitHubReleasesEnrichesWithoutClobbering(t *testing.T) {
	s := store.NewMemStore()
	svc := service.New(s)
	ctx := context.Background()
	repoID := "github.com/x/r"
	if err := s.CreateRepository(ctx, &store.Repository{ID: repoID, Organization: "x", RepositoryName: "r", Status: "active"}); err != nil {
		t.Fatal(err)
	}

	// Existing release: trusted CHANGELOG.json date + trailer-derived
	// associations, no URL/body yet.
	trustedDate := time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)
	if _, err := svc.RecordRelease(ctx, repoID, "v1.0.0", trustedDate, "", "CHANGELOG.json#v1.0.0", []string{"INIT-X-001"}, []string{"RMI-X-001"}); err != nil {
		t.Fatal(err)
	}

	ghDate := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC) // a different timestamp GitHub reports
	lookup := fakeLookup([]GHRelease{
		{TagName: "v1.0.0", PublishedAt: ghDate, HTMLURL: "https://github.com/x/r/releases/tag/v1.0.0", Body: "notes"},
	})

	res, err := GitHubReleases(ctx, svc, repoID, lookup)
	if err != nil {
		t.Fatal(err)
	}
	if res.Enriched != 1 || res.GapFilled != 0 {
		t.Fatalf("result: %+v", res)
	}

	rel, err := s.GetRelease(ctx, service.ReleaseID(repoID, "v1.0.0"))
	if err != nil {
		t.Fatal(err)
	}
	if rel.URL == "" || rel.Body == "" {
		t.Fatalf("expected URL/Body backfilled: %+v", rel)
	}
	// The trusted CHANGELOG.json date must survive untouched.
	if !rel.ReleasedAt.Equal(trustedDate) {
		t.Fatalf("ReleasedAt clobbered: got %v, want trusted %v", rel.ReleasedAt, trustedDate)
	}
	// Associations untouched.
	if len(rel.InitiativeIDs) != 1 || rel.InitiativeIDs[0] != "INIT-X-001" {
		t.Fatalf("associations disturbed: %+v", rel.InitiativeIDs)
	}

	// Second run: everything already filled, must report unchanged.
	res2, err := GitHubReleases(ctx, svc, repoID, lookup)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Enriched != 0 || res2.Unchanged != 1 {
		t.Fatalf("second run not idempotent: %+v", res2)
	}
}

func TestGitHubReleasesUnregisteredRepo(t *testing.T) {
	svc := service.New(store.NewMemStore())
	if _, err := GitHubReleases(context.Background(), svc, "github.com/nope/nope", fakeLookup(nil)); err == nil {
		t.Fatal("expected error for unregistered repo")
	}
}

func TestSplitRepoID(t *testing.T) {
	owner, name, err := splitRepoID("github.com/plexusone/omniskill")
	if err != nil || owner != "plexusone" || name != "omniskill" {
		t.Fatalf("split = %q, %q, %v", owner, name, err)
	}
	if _, _, err := splitRepoID("malformed"); err == nil {
		t.Fatal("expected error for malformed repo ID")
	}
}

func TestTruncateBody(t *testing.T) {
	long := make([]byte, maxReleaseBodyLen+500)
	for i := range long {
		long[i] = 'x'
	}
	got := truncateBody(string(long))
	if len(got) != maxReleaseBodyLen {
		t.Fatalf("truncated len = %d, want %d", len(got), maxReleaseBodyLen)
	}
	if got2 := truncateBody("short"); got2 != "short" {
		t.Fatalf("short body altered: %q", got2)
	}
}
