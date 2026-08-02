package ingest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func initRepo(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "init", "-q", "-b", "main")
	run(t, dir, "config", "user.name", "Test User")
	run(t, dir, "config", "user.email", "test@example.com")
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...) //nolint:gosec // test helper with controlled inputs
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_DATE=2026-07-01T10:00:00Z",
		"GIT_COMMITTER_DATE=2026-07-01T10:00:00Z",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func commitFile(t *testing.T, dir, name, content, message string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "add", name)
	run(t, dir, "commit", "-q", "-m", message)
}

func setupIngest(t *testing.T) (*service.Service, string) {
	t.Helper()
	dir := t.TempDir()
	initRepo(t, dir)

	ms := store.NewMemStore()
	svc := service.New(ms)
	ctx := context.Background()

	now := time.Now()
	if err := ms.CreateRepository(ctx, &store.Repository{
		ID: "github.com/test/repo", Organization: "test",
		RepositoryName: "repo", DefaultBranch: "main",
		LocalPath: dir, Status: "active",
	}); err != nil {
		t.Fatal(err)
	}
	if err := ms.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := ms.CreateRMI(ctx, &store.RoadmapItem{
		ID: "RMI-TEST-001", RepositoryID: "github.com/test/repo",
		InitiativeID: "INIT-TEST-001", Title: "First",
		ItemType: "capability", Status: "in_progress",
		Required: true, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := ms.CreateRMI(ctx, &store.RoadmapItem{
		ID: "RMI-TEST-002", RepositoryID: "github.com/test/repo",
		InitiativeID: "INIT-TEST-001", Title: "Second",
		ItemType: "capability", Status: "in_progress",
		Required: true, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	return svc, dir
}

func TestGitIngest(t *testing.T) {
	svc, dir := setupIngest(t)
	ctx := context.Background()

	commitFile(t, dir, "a.txt", "one\n", "feat: add feature\n\nRefs: RMI-TEST-001")
	commitFile(t, dir, "b.txt", "two\n", "fix: bug fix\n\nRefs: RMI-TEST-002")
	commitFile(t, dir, "c.txt", "three\n", "chore: no trailer")

	result, err := Git(ctx, svc, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}

	if result.CommitsWalked != 3 {
		t.Fatalf("commits walked: got %d, want 3", result.CommitsWalked)
	}
	if result.EvidenceAdded != 2 {
		t.Fatalf("evidence added: got %d, want 2", result.EvidenceAdded)
	}
	if result.Unattributed != 1 {
		t.Fatalf("unattributed: got %d, want 1", result.Unattributed)
	}

	// High-water mark should be set.
	repo, err := svc.Store.GetRepository(ctx, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo.IngestHighWater == "" {
		t.Fatal("expected non-empty high-water mark")
	}
}

func TestGitIngestIdempotent(t *testing.T) {
	svc, dir := setupIngest(t)
	ctx := context.Background()

	commitFile(t, dir, "a.txt", "one\n", "feat: add feature\n\nRefs: RMI-TEST-001")

	// First ingest
	r1, err := Git(ctx, svc, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}
	if r1.EvidenceAdded != 1 {
		t.Fatalf("first: evidence added = %d, want 1", r1.EvidenceAdded)
	}

	// Second ingest with no new commits
	r2, err := Git(ctx, svc, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}
	if r2.CommitsWalked != 0 {
		t.Fatalf("second: commits walked = %d, want 0", r2.CommitsWalked)
	}
	if r2.EvidenceAdded != 0 {
		t.Fatalf("second: evidence added = %d, want 0", r2.EvidenceAdded)
	}
}

func TestGitIngestIncremental(t *testing.T) {
	svc, dir := setupIngest(t)
	ctx := context.Background()

	commitFile(t, dir, "a.txt", "one\n", "feat: first\n\nRefs: RMI-TEST-001")

	// First ingest
	r1, err := Git(ctx, svc, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}
	if r1.EvidenceAdded != 1 {
		t.Fatalf("first: evidence = %d", r1.EvidenceAdded)
	}

	// Add more commits
	commitFile(t, dir, "b.txt", "two\n", "fix: second\n\nRefs: RMI-TEST-002")

	// Second ingest picks up only new commit
	r2, err := Git(ctx, svc, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}
	if r2.CommitsWalked != 1 {
		t.Fatalf("second: walked = %d, want 1", r2.CommitsWalked)
	}
	if r2.EvidenceAdded != 1 {
		t.Fatalf("second: evidence = %d, want 1", r2.EvidenceAdded)
	}
}

func TestGitIngestMultiRef(t *testing.T) {
	svc, dir := setupIngest(t)
	ctx := context.Background()

	commitFile(t, dir, "a.txt", "one\n", "feat: spans two RMIs\n\nRefs: RMI-TEST-001, RMI-TEST-002")

	result, err := Git(ctx, svc, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}
	if result.EvidenceAdded != 2 {
		t.Fatalf("evidence added: got %d, want 2 (one per RMI)", result.EvidenceAdded)
	}
}
