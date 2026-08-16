package releasegate

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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

func TestCheck(t *testing.T) {
	dir := t.TempDir()

	res := Check(dir, "v0.1.0")
	if res.ChangelogExists || res.VersionPresent {
		t.Fatalf("missing file: %+v", res)
	}

	cl := `{"project":"x","releases":[{"version":"v0.1.0","date":"2026-01-01"}]}`
	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.json"), []byte(cl), 0o600); err != nil {
		t.Fatal(err)
	}
	res = Check(dir, "v0.1.0")
	if !res.ChangelogExists || !res.VersionPresent {
		t.Fatalf("present version: %+v", res)
	}
	res = Check(dir, "v0.2.0")
	if !res.ChangelogExists || res.VersionPresent {
		t.Fatalf("absent version: %+v", res)
	}
}

func TestScaffoldNewFile(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")
	commitT(t, dir, "feat: add widget\n\nRefs: RMI-G-001")
	commitT(t, dir, "fix: repair gadget")
	commitT(t, dir, "random non-conventional message")

	res, err := Scaffold(context.Background(), dir, "v0.1.0")
	if err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	if !res.Created || res.Entries != 3 || res.RMIRefs != 1 {
		t.Fatalf("result: %+v", res)
	}

	data, err := os.ReadFile(res.ChangelogPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("scaffolded file unparseable: %v", err)
	}
	rels := doc["releases"].([]any)
	if len(rels) != 1 {
		t.Fatalf("releases = %d", len(rels))
	}
	rel := rels[0].(map[string]any)
	if rel["version"] != "v0.1.0" {
		t.Fatalf("version = %v", rel["version"])
	}
	added := rel["added"].([]any)
	entry := added[0].(map[string]any)
	if entry["description"] != "add widget" {
		t.Fatalf("description = %v (type prefix should be stripped)", entry["description"])
	}
	if entry["rmi"] != "RMI-G-001" {
		t.Fatalf("rmi = %v", entry["rmi"])
	}
	if _, ok := rel["fixed"]; !ok {
		t.Fatal("fixed category missing")
	}
	if _, ok := rel["changed"]; !ok {
		t.Fatal("changed category missing (non-conventional fallback)")
	}

	// Version now present; re-scaffold must refuse.
	if _, err := Scaffold(context.Background(), dir, "v0.1.0"); err == nil {
		t.Fatal("expected already-present error")
	}
}

func TestScaffoldPreservesUnknownFieldsAndPrepends(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	gitT(t, dir, "init", "-q", "-b", "main")
	commitT(t, dir, "feat: first")
	gitT(t, dir, "tag", "v0.1.0")
	commitT(t, dir, "feat: second\n\nRefs: RMI-G-002")

	existing := `{
  "$schema": "https://example/schema.json",
  "irVersion": "1.0",
  "project": "demo",
  "repository": "github.com/test/demo",
  "customField": {"keep": true},
  "releases": [{"version": "v0.1.0", "date": "2026-01-01", "added": [{"description": "first"}]}]
}`
	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.json"), []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}

	res, err := Scaffold(context.Background(), dir, "v0.2.0")
	if err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	if res.Created {
		t.Fatal("should not report created for existing file")
	}
	if res.SinceRef != "v0.1.0" {
		t.Fatalf("since = %s, want v0.1.0 (only commits after the last tag)", res.SinceRef)
	}
	if res.Entries != 1 {
		t.Fatalf("entries = %d, want 1 (pre-tag commit excluded)", res.Entries)
	}

	data, _ := os.ReadFile(res.ChangelogPath)
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	// Unknown fields survive.
	if doc["$schema"] != "https://example/schema.json" || doc["irVersion"] != "1.0" {
		t.Fatalf("top-level fields lost: %v", doc)
	}
	if cf, ok := doc["customField"].(map[string]any); !ok || cf["keep"] != true {
		t.Fatalf("customField lost: %v", doc["customField"])
	}
	// New release prepended, old preserved.
	rels := doc["releases"].([]any)
	if len(rels) != 2 {
		t.Fatalf("releases = %d", len(rels))
	}
	if rels[0].(map[string]any)["version"] != "v0.2.0" {
		t.Fatalf("first release = %v, want v0.2.0 (prepend)", rels[0])
	}
	if rels[1].(map[string]any)["version"] != "v0.1.0" {
		t.Fatalf("second release = %v", rels[1])
	}
}
