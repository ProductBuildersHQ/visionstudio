package cliconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissing(t *testing.T) {
	cfg, err := LoadFrom(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DSN != "" {
		t.Fatalf("expected empty DSN, got %q", cfg.DSN)
	}
}

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	orig := &Config{DSN: "root:@tcp(127.0.0.1:13306)/prismcontrol"}
	if err := orig.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DSN != orig.DSN {
		t.Fatalf("DSN mismatch: got %q, want %q", loaded.DSN, orig.DSN)
	}
}

func TestDefaultsWorkflow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	orig := &Config{
		DSN:      "root:@tcp(127.0.0.1:13306)/prismcontrol",
		Defaults: Defaults{Workflow: "pbhq-lite"},
	}
	if err := orig.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Defaults.Workflow != "pbhq-lite" {
		t.Fatalf("workflow mismatch: got %q, want %q", loaded.Defaults.Workflow, "pbhq-lite")
	}
}

func TestSaveCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep")
	path := filepath.Join(dir, "config.json")
	cfg := &Config{DSN: "test"}
	if err := cfg.SaveTo(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestSaveEmptyDSN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := &Config{DSN: "something"}
	if err := cfg.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	cfg.DSN = ""
	if err := cfg.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DSN != "" {
		t.Fatalf("expected empty DSN after unset, got %q", loaded.DSN)
	}
}

func TestTenantAssignments(t *testing.T) {
	c := &Config{}

	if _, ok := c.TenantFor("github.com/x/r"); ok {
		t.Fatal("expected no assignment on empty config")
	}

	c.AssignTenant("github.com/x/r", "acme")
	slug, ok := c.TenantFor("github.com/x/r")
	if !ok || slug != "acme" {
		t.Fatalf("TenantFor = %q, %v", slug, ok)
	}

	// Reassign overwrites.
	c.AssignTenant("github.com/x/r", "acme-personal")
	slug, _ = c.TenantFor("github.com/x/r")
	if slug != "acme-personal" {
		t.Fatalf("reassign = %q, want acme-personal", slug)
	}

	// A second repo can map to a different tenant — one machine, many tenants.
	c.AssignTenant("github.com/x/other", "acme-work")
	if slug, _ := c.TenantFor("github.com/x/other"); slug != "acme-work" {
		t.Fatalf("second repo = %q", slug)
	}
	if slug, _ := c.TenantFor("github.com/x/r"); slug != "acme-personal" {
		t.Fatal("first assignment disturbed by second")
	}

	if !c.UnassignTenant("github.com/x/r") {
		t.Fatal("expected unassign to report existing")
	}
	if _, ok := c.TenantFor("github.com/x/r"); ok {
		t.Fatal("expected no assignment after unassign")
	}
	if c.UnassignTenant("github.com/x/r") {
		t.Fatal("expected unassign of already-gone entry to report false")
	}
}

func TestTenantAssignmentsPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	orig := &Config{}
	orig.AssignTenant("github.com/x/r", "acme")
	if err := orig.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if slug, ok := loaded.TenantFor("github.com/x/r"); !ok || slug != "acme" {
		t.Fatalf("loaded assignment = %q, %v", slug, ok)
	}
}
