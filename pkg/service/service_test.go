package service

import (
	"context"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func newTestService() *Service {
	return New(store.NewMemStore())
}

func TestRegisterRepository(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	repo, err := svc.RegisterRepository(ctx, "plexusone", "omnidevx", "", "/path/to/omnidevx", "productivity")
	if err != nil {
		t.Fatal(err)
	}
	if repo.ID != "github.com/plexusone/omnidevx" {
		t.Fatalf("expected ID github.com/plexusone/omnidevx, got %s", repo.ID)
	}
	if repo.DefaultBranch != "main" {
		t.Fatalf("expected default branch main, got %s", repo.DefaultBranch)
	}
	if repo.Status != "active" {
		t.Fatalf("expected status active, got %s", repo.Status)
	}

	// duplicate should fail
	_, err = svc.RegisterRepository(ctx, "plexusone", "omnidevx", "", "", "")
	if err == nil {
		t.Fatal("expected error on duplicate registration")
	}
}

func TestListRepositories(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.RegisterRepository(ctx, "plexusone", "omnidevx", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RegisterRepository(ctx, "grokify", "gogit", "", "", ""); err != nil {
		t.Fatal(err)
	}

	repos, err := svc.ListRepositories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
}

func TestListRepositoriesByOrg(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.RegisterRepository(ctx, "plexusone", "omnidevx", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RegisterRepository(ctx, "plexusone", "devfolio", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RegisterRepository(ctx, "grokify", "gogit", "", "", ""); err != nil {
		t.Fatal(err)
	}

	repos, err := svc.ListRepositoriesByOrg(ctx, "plexusone")
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 plexusone repos, got %d", len(repos))
	}
}

func TestImportRepository(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	repo := &store.Repository{
		ID:             "github.com/plexusone/omnidevx",
		Organization:   "plexusone",
		RepositoryName: "omnidevx",
		DefaultBranch:  "main",
		LocalPath:      "/old/path",
		GoModule:       "github.com/plexusone/omnidevx",
	}
	if err := svc.ImportRepository(ctx, repo); err != nil {
		t.Fatal(err)
	}

	got, err := svc.GetRepository(ctx, "github.com/plexusone/omnidevx")
	if err != nil {
		t.Fatal(err)
	}
	if got.LocalPath != "/old/path" {
		t.Fatalf("expected /old/path, got %s", got.LocalPath)
	}

	// upsert with new path
	repo2 := &store.Repository{
		ID:             "github.com/plexusone/omnidevx",
		Organization:   "plexusone",
		RepositoryName: "omnidevx",
		LocalPath:      "/new/path",
		GoModule:       "github.com/plexusone/omnidevx",
	}
	if err := svc.ImportRepository(ctx, repo2); err != nil {
		t.Fatal(err)
	}

	got, err = svc.GetRepository(ctx, "github.com/plexusone/omnidevx")
	if err != nil {
		t.Fatal(err)
	}
	if got.LocalPath != "/new/path" {
		t.Fatalf("expected /new/path after upsert, got %s", got.LocalPath)
	}
}

func TestImportRepoDependencies(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	// register both repos
	if _, err := svc.RegisterRepository(ctx, "plexusone", "omnidevx", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RegisterRepository(ctx, "plexusone", "omnidevx-core", "", "", ""); err != nil {
		t.Fatal(err)
	}

	targets := []string{
		"github.com/plexusone/omnidevx-core",
		"github.com/external/not-registered",
	}
	created, err := svc.ImportRepoDependencies(ctx, "github.com/plexusone/omnidevx", targets)
	if err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("expected 1 created (skipping unregistered), got %d", created)
	}

	// idempotent — second call creates 0
	created, err = svc.ImportRepoDependencies(ctx, "github.com/plexusone/omnidevx", targets)
	if err != nil {
		t.Fatal(err)
	}
	if created != 0 {
		t.Fatalf("expected 0 on idempotent call, got %d", created)
	}

	deps, err := svc.RepoDependencyGraph(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency in graph, got %d", len(deps))
	}
}

func TestListReposWithLocalPath(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	if _, err := svc.RegisterRepository(ctx, "plexusone", "omnidevx", "", "/path/a", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RegisterRepository(ctx, "plexusone", "devfolio", "", "", ""); err != nil {
		t.Fatal(err)
	}

	repos, err := svc.ListReposWithLocalPath(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo with local path, got %d", len(repos))
	}
	if repos[0].ID != "github.com/plexusone/omnidevx" {
		t.Fatalf("unexpected repo ID: %s", repos[0].ID)
	}
}
