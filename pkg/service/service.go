// Package service is the shared business-logic layer for PRISM Control.
// Both the CLI (cmd/prismctl) and MCP server call these methods;
// protocol rules (validation, transitions, lease checks) live here
// so they cannot diverge between adapters.
package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Service wraps the store with business logic shared by all adapters.
type Service struct {
	Store store.Store
}

// New creates a Service backed by the given store implementation.
func New(s store.Store) *Service {
	return &Service{Store: s}
}

// DBInit bootstraps a Dolt database at the given directory.
// It runs `dolt init` if the directory doesn't already contain a Dolt repo,
// then returns the directory path for use as a sql-server root.
func DBInit(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	doltDir := filepath.Join(dir, ".dolt")
	if _, err := os.Stat(doltDir); err == nil {
		return nil
	}

	cmd := exec.Command("dolt", "init", "--name", "prismctl", "--email", "prismctl@local")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dolt init: %w", err)
	}
	return nil
}

// DBServe starts a Dolt SQL server in the given directory.
// It blocks until the server exits or the context is cancelled.
func DBServe(ctx context.Context, dir string, port int) error {
	args := []string{"sql-server", "--host", "127.0.0.1", "--port", fmt.Sprintf("%d", port)}
	cmd := exec.CommandContext(ctx, "dolt", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("dolt sql-server: %w", err)
	}
	return nil
}

// RegisterRepository adds a repository to the catalog.
func (s *Service) RegisterRepository(ctx context.Context, org, name, defaultBranch, localPath, domain string) (*store.Repository, error) {
	id := fmt.Sprintf("github.com/%s/%s", org, name)

	if defaultBranch == "" {
		defaultBranch = "main"
	}

	orgRow, _, err := s.EnsureOrganization(ctx, org, "")
	if err != nil {
		return nil, fmt.Errorf("ensure organization %s: %w", org, err)
	}

	repo := &store.Repository{
		ID:             id,
		Organization:   org,
		RepositoryName: name,
		DefaultBranch:  defaultBranch,
		LocalPath:      localPath,
		Domain:         domain,
		Status:         "active",
		OrganizationID: orgRow.ID,
	}
	if err := s.Store.CreateRepository(ctx, repo); err != nil {
		return nil, err
	}
	return repo, nil
}

// ListRepositories returns all registered repositories.
func (s *Service) ListRepositories(ctx context.Context) ([]*store.Repository, error) {
	return s.Store.ListRepositories(ctx)
}

// ListRepositoriesByOrg returns repositories filtered by organization.
func (s *Service) ListRepositoriesByOrg(ctx context.Context, org string) ([]*store.Repository, error) {
	return s.Store.ListRepositoriesByOrg(ctx, org)
}

// GetRepository returns a single repository by ID.
func (s *Service) GetRepository(ctx context.Context, id string) (*store.Repository, error) {
	return s.Store.GetRepository(ctx, id)
}

// ImportRepository upserts a repository — creates if new, updates if existing.
// Used by the scan flow for bulk import.
func (s *Service) ImportRepository(ctx context.Context, repo *store.Repository) error {
	existing, err := s.Store.GetRepository(ctx, repo.ID)
	if err != nil {
		repo.Status = "active"
		return s.Store.CreateRepository(ctx, repo)
	}
	existing.LocalPath = repo.LocalPath
	existing.GoModule = repo.GoModule
	existing.Status = "active"
	if repo.DefaultBranch != "" {
		existing.DefaultBranch = repo.DefaultBranch
	}
	return s.Store.UpdateRepository(ctx, existing)
}

// ImportRepoDependencies replaces repo-level dependencies for a source repo.
// It creates new dependencies that don't exist yet. Only dependencies where
// both source and target are in the registry are stored.
func (s *Service) ImportRepoDependencies(ctx context.Context, sourceRepoID string, targetRepoIDs []string) (int, error) {
	var created int
	for _, targetID := range targetRepoIDs {
		if _, err := s.Store.GetRepository(ctx, targetID); err != nil {
			continue
		}
		dep := &store.RepositoryDependency{
			SourceRepositoryID: sourceRepoID,
			TargetRepositoryID: targetID,
			DependencyType:     "go_module",
		}
		if err := s.Store.CreateRepoDependency(ctx, dep); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return created, err
		}
		created++
	}
	return created, nil
}

// RepoDependencyGraph returns all repo dependencies for display.
func (s *Service) RepoDependencyGraph(ctx context.Context) ([]*store.RepositoryDependency, error) {
	return s.Store.ListAllRepoDependencies(ctx)
}

// RepoInfo is a summary used by the unpushed command.
type RepoInfo struct {
	ID        string
	LocalPath string
}

// ListReposWithLocalPath returns repos that have a local_path set.
func (s *Service) ListReposWithLocalPath(ctx context.Context) ([]RepoInfo, error) {
	repos, err := s.Store.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}
	var result []RepoInfo
	for _, r := range repos {
		if r.LocalPath != "" {
			result = append(result, RepoInfo{ID: r.ID, LocalPath: r.LocalPath})
		}
	}
	return result, nil
}

// Migrate runs database schema migration via the provided migrator function.
// This indirection keeps the service layer independent of Ent.
func (s *Service) Migrate(ctx context.Context, migrator func(ctx context.Context) error) error {
	return migrator(ctx)
}

// Now returns the current time. Centralized for testability.
func Now() time.Time {
	return time.Now()
}
