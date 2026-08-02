package ingest

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func setupChangelog(t *testing.T) (*service.Service, string) {
	t.Helper()
	dir := t.TempDir()
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

	return svc, dir
}

func TestChangelogIngest(t *testing.T) {
	svc, dir := setupChangelog(t)
	ctx := context.Background()

	changelog := `{
		"irVersion": "1.0",
		"project": "test-repo",
		"releases": [
			{
				"version": "v1.0.0",
				"date": "2026-07-01",
				"commit": "abc123",
				"added": [
					{"description": "New feature", "commit": "def456", "rmi_ref": "RMI-TEST-001"},
					{"description": "Another feature", "commit": "ghi789"}
				],
				"fixed": [
					{"description": "Bug fix", "rmi_ref": "RMI-TEST-001"}
				]
			}
		]
	}`
	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.json"), []byte(changelog), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Changelog(ctx, svc, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}

	if result.ReleasesRead != 1 {
		t.Fatalf("releases: got %d, want 1", result.ReleasesRead)
	}
	if result.EntriesRead != 3 {
		t.Fatalf("entries: got %d, want 3", result.EntriesRead)
	}
	if result.EvidenceAdded != 2 {
		t.Fatalf("evidence: got %d, want 2", result.EvidenceAdded)
	}
}

func TestChangelogNoRefs(t *testing.T) {
	svc, dir := setupChangelog(t)
	ctx := context.Background()

	changelog := `{
		"releases": [
			{
				"version": "v0.1.0",
				"added": [
					{"description": "Something"}
				]
			}
		]
	}`
	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.json"), []byte(changelog), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Changelog(ctx, svc, "github.com/test/repo")
	if err != nil {
		t.Fatal(err)
	}

	if result.EvidenceAdded != 0 {
		t.Fatalf("evidence: got %d, want 0", result.EvidenceAdded)
	}
}
