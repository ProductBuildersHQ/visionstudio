package ingest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestSemverishTag(t *testing.T) {
	pass := []string{"v0.1.0", "0.1.0", "v1.2", "v2.0.0-rc.1", "v0.13.0"}
	fail := []string{"smoke-test", "release", "v", "latest", "nightly-2026"}
	for _, s := range pass {
		if !semverishTag.MatchString(s) {
			t.Errorf("%q should match", s)
		}
	}
	for _, s := range fail {
		if semverishTag.MatchString(s) {
			t.Errorf("%q should not match", s)
		}
	}
}

func gitT(t *testing.T, dir string, args ...string) {
	t.Helper()
	// #nosec G204 -- test helper; args are literals from this test file.
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func commitT(t *testing.T, dir, msg string) {
	t.Helper()
	gitT(t, dir, "commit", "--allow-empty", "-m", msg)
}

func TestReleasesTrailerChain(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")

	// Range 1 (..v0.1.0): one trailered commit, one plain.
	commitT(t, dir, "feat: a\n\nRefs: RMI-T-001")
	commitT(t, dir, "chore: noise")
	gitT(t, dir, "tag", "v0.1.0")
	// Range 2 (v0.1.0..v0.2.0): trailer for a second RMI + one unknown RMI.
	commitT(t, dir, "feat: b\n\nRefs: RMI-T-002")
	commitT(t, dir, "fix: c\n\nRefs: RMI-UNKNOWN-999")
	gitT(t, dir, "tag", "v0.2.0")
	// A non-semver tag that must be skipped.
	gitT(t, dir, "tag", "not-a-release")

	// CHANGELOG.json supplies the date + notes ref for v0.2.0 only.
	cl := `{"releases":[{"version":"v0.2.0","date":"2026-03-04"},{"version":"v9.9.9","date":"2026-01-01"}]}`
	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.json"), []byte(cl), 0o600); err != nil {
		t.Fatal(err)
	}

	s := store.NewMemStore()
	svc := service.New(s)
	ctx := context.Background()
	repoID := "github.com/test/repo"
	if err := s.CreateRepository(ctx, &store.Repository{ID: repoID, Organization: "test", RepositoryName: "repo", Status: "active", LocalPath: dir}); err != nil {
		t.Fatal(err)
	}
	for _, rmi := range []*store.RoadmapItem{
		{ID: "RMI-T-001", RepositoryID: repoID, InitiativeID: "INIT-T-001", Title: "a", ItemType: "capability", Status: "completed"},
		{ID: "RMI-T-002", RepositoryID: repoID, InitiativeID: "INIT-T-002", Title: "b", ItemType: "capability", Status: "completed"},
	} {
		if err := s.CreateRMI(ctx, rmi); err != nil {
			t.Fatal(err)
		}
	}

	res, err := Releases(ctx, svc, repoID, false)
	if err != nil {
		t.Fatalf("releases: %v", err)
	}
	if res.ReleasesUpserted != 2 {
		t.Fatalf("upserted = %d, want 2", res.ReleasesUpserted)
	}
	if res.TagsSkipped != 1 {
		t.Fatalf("skipped = %d, want 1 (not-a-release)", res.TagsSkipped)
	}
	if res.ChangelogOnly != 1 {
		t.Fatalf("changelog-only = %d, want 1 (v9.9.9)", res.ChangelogOnly)
	}
	if res.CommitsWalked != 4 || res.CommitsWithTrailers != 3 {
		t.Fatalf("coverage = %d/%d, want 3/4", res.CommitsWithTrailers, res.CommitsWalked)
	}
	if res.UnresolvedRMIRefs != 1 {
		t.Fatalf("unresolved = %d, want 1 (RMI-UNKNOWN-999)", res.UnresolvedRMIRefs)
	}

	// Release 1: associated with INIT-T-001 via trailer chain, tag date.
	rel1, err := s.GetRelease(ctx, repoID+"@v0.1.0")
	if err != nil {
		t.Fatalf("get v0.1.0: %v", err)
	}
	if len(rel1.InitiativeIDs) != 1 || rel1.InitiativeIDs[0] != "INIT-T-001" {
		t.Fatalf("v0.1.0 initiatives = %v", rel1.InitiativeIDs)
	}
	if rel1.NotesRef != "" {
		t.Fatalf("v0.1.0 notes = %q, want empty (not in changelog)", rel1.NotesRef)
	}

	// Release 2: range isolation (only RMI-T-002), changelog date + notes.
	rel2, err := s.GetRelease(ctx, repoID+"@v0.2.0")
	if err != nil {
		t.Fatalf("get v0.2.0: %v", err)
	}
	if len(rel2.InitiativeIDs) != 1 || rel2.InitiativeIDs[0] != "INIT-T-002" {
		t.Fatalf("v0.2.0 initiatives = %v (range isolation broken?)", rel2.InitiativeIDs)
	}
	if rel2.NotesRef != "CHANGELOG.json#v0.2.0" {
		t.Fatalf("v0.2.0 notes = %q", rel2.NotesRef)
	}
	if rel2.ReleasedAt.Format("2006-01-02") != "2026-03-04" {
		t.Fatalf("v0.2.0 date = %s, want changelog date", rel2.ReleasedAt.Format("2006-01-02"))
	}

	// Idempotent re-run: same counts, no duplicates.
	res2, err := Releases(ctx, svc, repoID, false)
	if err != nil {
		t.Fatalf("re-run: %v", err)
	}
	if res2.ReleasesUpserted != 2 {
		t.Fatalf("re-run upserted = %d", res2.ReleasesUpserted)
	}
	all, _ := s.ListReleases(ctx)
	if len(all) != 2 {
		t.Fatalf("releases after re-run = %d, want 2", len(all))
	}
}

func TestReleasesMultiTrailerCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")
	// One commit naming two RMIs via two trailer lines.
	commitT(t, dir, "feat: combo\n\nRefs: RMI-M-001\nRefs: RMI-M-002")
	gitT(t, dir, "tag", "v1.0.0")

	s := store.NewMemStore()
	svc := service.New(s)
	ctx := context.Background()
	repoID := "github.com/test/multi"
	if err := s.CreateRepository(ctx, &store.Repository{ID: repoID, Organization: "test", RepositoryName: "multi", Status: "active", LocalPath: dir}); err != nil {
		t.Fatal(err)
	}
	for _, rmi := range []*store.RoadmapItem{
		{ID: "RMI-M-001", RepositoryID: repoID, InitiativeID: "INIT-M-001", Title: "a", ItemType: "capability", Status: "completed"},
		{ID: "RMI-M-002", RepositoryID: repoID, InitiativeID: "INIT-M-002", Title: "b", ItemType: "capability", Status: "completed"},
	} {
		if err := s.CreateRMI(ctx, rmi); err != nil {
			t.Fatal(err)
		}
	}

	res, err := Releases(ctx, svc, repoID, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.CommitsWithTrailers != 1 {
		t.Fatalf("trailered commits = %d, want 1 (one commit, two trailers)", res.CommitsWithTrailers)
	}
	rel, err := s.GetRelease(ctx, repoID+"@v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(rel.RMIIDs) != 2 || len(rel.InitiativeIDs) != 2 {
		t.Fatalf("associations = %v / %v, want both RMIs and both initiatives", rel.RMIIDs, rel.InitiativeIDs)
	}
}

func TestSummarizeReleaseResults(t *testing.T) {
	results := []*ReleaseIngestResult{
		{ReleasesUpserted: 2, InitiativeLinks: 1, RMILinks: 3, CommitsWalked: 10, CommitsWithTrailers: 4},
		{ReleasesUpserted: 1, CommitsWalked: 5, CommitsWithTrailers: 1, UnresolvedRMIRefs: 2},
		{Err: os.ErrNotExist},
	}
	sum := SummarizeReleaseResults(results)
	if sum.Releases != 3 || sum.InitiativeLinks != 1 || sum.RMILinks != 3 {
		t.Fatalf("totals = %+v", sum)
	}
	if sum.CommitsWalked != 15 || sum.CommitsWithTrailers != 5 {
		t.Fatalf("coverage inputs = %+v", sum)
	}
	if got := sum.Coverage(); got < 0.33 || got > 0.34 {
		t.Fatalf("coverage = %f, want ~1/3", got)
	}
	if sum.UnresolvedRMIRefs != 2 || sum.ReposWithErrors != 1 {
		t.Fatalf("errs = %+v", sum)
	}
	if (&ReleaseIngestSummary{}).Coverage() != 0 {
		t.Fatal("zero-walk coverage must be 0, not NaN")
	}
}

func TestReleasesTaglessRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")
	commitT(t, dir, "feat: unreleased work\n\nRefs: RMI-T-001")

	s := store.NewMemStore()
	svc := service.New(s)
	ctx := context.Background()
	repoID := "github.com/test/tagless"
	if err := s.CreateRepository(ctx, &store.Repository{ID: repoID, Organization: "test", RepositoryName: "tagless", Status: "active", LocalPath: dir}); err != nil {
		t.Fatal(err)
	}

	res, err := Releases(ctx, svc, repoID, false)
	if err != nil {
		t.Fatalf("tagless repo must not error: %v", err)
	}
	if res.ReleasesUpserted != 0 || res.CommitsWalked != 0 {
		t.Fatalf("tagless result = %+v, want all zeros (no ranges to walk)", res)
	}
}
