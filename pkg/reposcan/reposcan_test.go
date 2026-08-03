package reposcan

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/grokify/gogit/scanner"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func newTestService() *service.Service {
	return service.New(store.NewMemStore())
}

func TestResultToRepository(t *testing.T) {
	r := scanner.RepoResult{
		Name:       "omnidevx",
		IsGitRepo:  true,
		HasGoMod:   true,
		ModuleName: "github.com/plexusone/omnidevx",
	}
	// Use filepath.Join for cross-platform path construction
	orgDir := filepath.Join("home", "user", "go", "src", "github.com", "plexusone")
	repo := resultToRepository("plexusone", r, orgDir)

	if repo.ID != "github.com/plexusone/omnidevx" {
		t.Fatalf("expected ID github.com/plexusone/omnidevx, got %s", repo.ID)
	}
	if repo.Organization != "plexusone" {
		t.Fatalf("expected org plexusone, got %s", repo.Organization)
	}
	if repo.GoModule != "github.com/plexusone/omnidevx" {
		t.Fatalf("expected module github.com/plexusone/omnidevx, got %s", repo.GoModule)
	}
	expectedPath := filepath.Join(orgDir, "omnidevx")
	if repo.LocalPath != expectedPath {
		t.Fatalf("unexpected local path: got %s, want %s", repo.LocalPath, expectedPath)
	}
	if repo.Status != "active" {
		t.Fatalf("expected status active, got %s", repo.Status)
	}
}

func TestImportDependencies(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	// Pre-register repos so dependency import can find them
	for _, name := range []string{"omnidevx", "omnidevx-core", "omniskill"} {
		if _, err := svc.RegisterRepository(ctx, "plexusone", name, "main", "", ""); err != nil {
			t.Fatal(err)
		}
	}

	results := []scanner.RepoResult{
		{
			Name:         "omnidevx",
			IsGitRepo:    true,
			ModuleName:   "github.com/plexusone/omnidevx",
			Dependencies: []string{"github.com/plexusone/omnidevx-core", "github.com/plexusone/omniskill", "github.com/external/lib"},
		},
		{
			Name:         "omnidevx-core",
			IsGitRepo:    true,
			ModuleName:   "github.com/plexusone/omnidevx-core",
			Dependencies: []string{},
		},
		{
			Name:       "omniskill",
			IsGitRepo:  true,
			ModuleName: "github.com/plexusone/omniskill",
		},
	}

	created, err := importDependencies(ctx, svc, "plexusone", results)
	if err != nil {
		t.Fatal(err)
	}
	if created != 2 {
		t.Fatalf("expected 2 deps created (external skipped), got %d", created)
	}

	// Idempotent
	created, err = importDependencies(ctx, svc, "plexusone", results)
	if err != nil {
		t.Fatal(err)
	}
	if created != 0 {
		t.Fatalf("expected 0 on second import, got %d", created)
	}
}

func TestInternalDeps(t *testing.T) {
	all := []scanner.RepoResult{
		{Name: "omnidevx-core", ModuleName: "github.com/plexusone/omnidevx-core"},
		{Name: "omniskill", ModuleName: "github.com/plexusone/omniskill"},
	}

	r := scanner.RepoResult{
		Name:         "omnidevx",
		ModuleName:   "github.com/plexusone/omnidevx",
		Dependencies: []string{"github.com/plexusone/omnidevx-core", "github.com/plexusone/omniskill", "github.com/external/lib"},
	}

	deps := internalDeps(r, all, "plexusone")
	if len(deps) != 2 {
		t.Fatalf("expected 2 internal deps, got %d", len(deps))
	}
}
