package export

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func populateStore(t *testing.T, s store.Store) {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	if err := s.CreateInitiative(ctx, &store.Initiative{
		ID: "INIT-TEST-001", Organization: "test", Title: "Test Initiative",
		Status: "executing", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreatePhase(ctx, &store.Phase{
		ID: "INIT-TEST-001/phase-1", InitiativeID: "INIT-TEST-001",
		SequenceNumber: 1, Title: "Foundation",
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateRepository(ctx, &store.Repository{
		ID: "github.com/test/repo", Organization: "test",
		RepositoryName: "repo", DefaultBranch: "main", Status: "active",
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateRMI(ctx, &store.RoadmapItem{
		ID: "RMI-TEST-001", RepositoryID: "github.com/test/repo",
		InitiativeID: "INIT-TEST-001", PhaseID: "INIT-TEST-001/phase-1",
		Title: "First item", ItemType: "capability", Status: "completed",
		Required: true, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRMI(ctx, &store.RoadmapItem{
		ID: "RMI-TEST-002", RepositoryID: "github.com/test/repo",
		InitiativeID: "INIT-TEST-001", PhaseID: "INIT-TEST-001/phase-1",
		Title: "Second item", ItemType: "capability", Status: "ready",
		Required: true, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateDependency(ctx, &store.RMIDependency{
		SourceRMIID: "RMI-TEST-002", TargetRMIID: "RMI-TEST-001", Relationship: "requires",
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateAssignment(ctx, &store.Assignment{
		ID: "ASSIGN-001", RMIID: "RMI-TEST-001", Worker: "session-1",
		Status: "completed", LeaseExpiresAt: now.Add(4 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.CreateEvidence(ctx, &store.DeliveryEvidence{
		ID: "EV-001", RMIID: "RMI-TEST-001", EvidenceType: "commit",
		Reference: "abc123", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestExportRun(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemStore()
	populateStore(t, s)

	dir := t.TempDir()
	exportDir := filepath.Join(dir, "exports")

	result, err := Run(ctx, s, exportDir)
	if err != nil {
		t.Fatal(err)
	}

	// Check manifest
	if result.Manifest.Counts["initiatives"] != 1 {
		t.Fatalf("expected 1 initiative, got %d", result.Manifest.Counts["initiatives"])
	}
	if result.Manifest.Counts["phases"] != 1 {
		t.Fatalf("expected 1 phase, got %d", result.Manifest.Counts["phases"])
	}
	if result.Manifest.Counts["roadmap_items"] != 2 {
		t.Fatalf("expected 2 RMIs, got %d", result.Manifest.Counts["roadmap_items"])
	}
	if result.Manifest.Counts["rmi_dependencies"] != 1 {
		t.Fatalf("expected 1 dep, got %d", result.Manifest.Counts["rmi_dependencies"])
	}
	if result.Manifest.Counts["assignments"] != 1 {
		t.Fatalf("expected 1 assignment, got %d", result.Manifest.Counts["assignments"])
	}
	if result.Manifest.Counts["evidence"] != 1 {
		t.Fatalf("expected 1 evidence, got %d", result.Manifest.Counts["evidence"])
	}
	if result.Manifest.Counts["repositories"] != 1 {
		t.Fatalf("expected 1 repo, got %d", result.Manifest.Counts["repositories"])
	}
	if result.Manifest.Counts["repository_dependencies"] != 0 {
		t.Fatalf("expected 0 repo deps, got %d", result.Manifest.Counts["repository_dependencies"])
	}

	// Verify files exist
	expectedFiles := []string{
		"initiatives.jsonl", "phases.jsonl", "roadmap_items.jsonl",
		"rmi_dependencies.jsonl", "assignments.jsonl", "evidence.jsonl",
		"repositories.jsonl", "repository_dependencies.jsonl", "manifest.json",
	}
	for _, name := range expectedFiles {
		path := filepath.Join(exportDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("expected file %s to exist", name)
		}
	}

	// Verify manifest.json is valid JSON
	manifestBytes, err := os.ReadFile(filepath.Join(exportDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("invalid manifest JSON: %v", err)
	}
	if manifest.ExportedAt.IsZero() {
		t.Fatal("expected non-zero exported_at")
	}

	// Verify JSONL line count for roadmap_items
	rmiData, err := os.ReadFile(filepath.Join(exportDir, "roadmap_items.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := 0
	for _, b := range rmiData {
		if b == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Fatalf("expected 2 JSONL lines for roadmap_items, got %d", lines)
	}
}

func TestExportEmpty(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemStore()

	dir := t.TempDir()
	exportDir := filepath.Join(dir, "exports")

	result, err := Run(ctx, s, exportDir)
	if err != nil {
		t.Fatal(err)
	}

	for table, count := range result.Manifest.Counts {
		if count != 0 {
			t.Fatalf("expected 0 for %s, got %d", table, count)
		}
	}
}
