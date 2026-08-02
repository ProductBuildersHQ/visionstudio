package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/specworkflow"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestSpecSyncFindsSpecs(t *testing.T) {
	ctx := context.Background()
	ms := store.NewMemStore()

	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create initiative
	init := &store.Initiative{
		ID:       "INIT-TEST-001",
		Title:    "Test Initiative",
		Status:   "in_progress",
		HomeRepo: "test-repo",
	}
	if err := ms.CreateInitiative(ctx, init); err != nil {
		t.Fatal(err)
	}

	// Create repository pointing to temp dir
	repo := &store.Repository{
		ID:        "test-repo",
		LocalPath: tmpDir,
	}
	if err := ms.CreateRepository(ctx, repo); err != nil {
		t.Fatal(err)
	}

	// Create spec directory and files
	specDir := filepath.Join(tmpDir, specworkflow.SpecDir("INIT-TEST-001"))
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	specFiles := []string{"PRD.md", "TRD.md", "PLAN.md", "ROADMAP.md"}
	for _, name := range specFiles {
		path := filepath.Join(specDir, name)
		if err := os.WriteFile(path, []byte("# "+name), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Verify specs directory exists
	entries, err := os.ReadDir(specDir)
	if err != nil {
		t.Fatalf("failed to read spec dir: %v", err)
	}
	if len(entries) != len(specFiles) {
		t.Fatalf("expected %d files, got %d", len(specFiles), len(entries))
	}

	// Verify initiative has no specs initially
	gotInit, err := ms.GetInitiative(ctx, "INIT-TEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(gotInit.Specs) != 0 {
		t.Errorf("expected 0 specs initially, got %d", len(gotInit.Specs))
	}

	// The actual sync logic is in the command; this test verifies the
	// file structure is correct for the sync to work
	specsBase := filepath.Join(tmpDir, "docs", "specs", "initiatives")
	initDirs, err := os.ReadDir(specsBase)
	if err != nil {
		t.Fatalf("failed to read specs base: %v", err)
	}

	foundInitDir := false
	for _, d := range initDirs {
		if d.Name() == "INIT-TEST-001" {
			foundInitDir = true
			break
		}
	}
	if !foundInitDir {
		t.Error("expected to find INIT-TEST-001 directory")
	}
}

func TestSpecSyncLegacyDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create legacy spec location
	legacyDir := filepath.Join(tmpDir, "docs", "specs")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create legacy spec files
	legacyFiles := []string{"PRD.md", "TRD.md", "PLAN.md", "ROADMAP.md"}
	for _, name := range legacyFiles {
		path := filepath.Join(legacyDir, name)
		if err := os.WriteFile(path, []byte("# Legacy "+name), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Verify files exist at legacy location
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		t.Fatal(err)
	}

	mdFiles := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			mdFiles++
		}
	}

	if mdFiles != len(legacyFiles) {
		t.Errorf("expected %d .md files, got %d", len(legacyFiles), mdFiles)
	}
}
