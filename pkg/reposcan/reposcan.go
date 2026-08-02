// Package reposcan wraps gogit/scanner to bulk-import repositories
// into the PRISM Control registry and populate dependency edges.
package reposcan

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/grokify/gogit/scanner"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// ScanResult summarizes what a scan imported.
type ScanResult struct {
	ScannedDir   string
	TotalScanned int
	GitRepos     int
	Imported     int
	DepsCreated  int
}

// ScanAndImport scans an organization directory, registers all git repos
// found, and populates repository dependency edges from go.mod data.
// The org is inferred from the directory name (e.g., "plexusone" from
// ~/go/src/github.com/plexusone/).
func ScanAndImport(ctx context.Context, svc *service.Service, dirPath string, opts ScanOptions) (*ScanResult, error) {
	org := filepath.Base(dirPath)

	scanOpts := scanner.ScanOptions{
		CheckUnpushed: opts.CheckUnpushed,
	}

	results, err := scanner.ScanDirectoryWithProgress(dirPath, opts.Progress, scanOpts)
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", dirPath, err)
	}

	res := &ScanResult{
		ScannedDir:   dirPath,
		TotalScanned: len(results),
	}

	for _, r := range results {
		if !r.IsGitRepo {
			continue
		}
		res.GitRepos++

		repo := resultToRepository(org, r, dirPath)
		if err := svc.ImportRepository(ctx, repo); err != nil {
			return res, fmt.Errorf("import %s: %w", r.Name, err)
		}
		res.Imported++
	}

	res.DepsCreated, err = importDependencies(ctx, svc, org, results)
	if err != nil {
		return res, fmt.Errorf("import dependencies: %w", err)
	}

	return res, nil
}

// ScanOptions configures the scan behavior.
type ScanOptions struct {
	CheckUnpushed bool
	Progress      scanner.ProgressFunc
}

func resultToRepository(org string, r scanner.RepoResult, dirPath string) *store.Repository {
	id := fmt.Sprintf("github.com/%s/%s", org, r.Name)
	repo := &store.Repository{
		ID:             id,
		Organization:   org,
		RepositoryName: r.Name,
		DefaultBranch:  "main",
		LocalPath:      filepath.Join(dirPath, r.Name),
		Status:         "active",
	}
	if r.ModuleName != "" {
		repo.GoModule = r.ModuleName
	}
	return repo
}

func importDependencies(ctx context.Context, svc *service.Service, org string, results []scanner.RepoResult) (int, error) {
	moduleToRepoID := make(map[string]string)
	for _, r := range results {
		if r.ModuleName != "" {
			moduleToRepoID[r.ModuleName] = fmt.Sprintf("github.com/%s/%s", org, r.Name)
		}
	}

	totalCreated := 0
	for _, r := range results {
		if !r.IsGitRepo || r.ModuleName == "" || len(r.Dependencies) == 0 {
			continue
		}

		sourceID := fmt.Sprintf("github.com/%s/%s", org, r.Name)
		var targetIDs []string
		for _, dep := range r.Dependencies {
			if targetID, ok := moduleToRepoID[dep]; ok {
				targetIDs = append(targetIDs, targetID)
			}
		}
		if len(targetIDs) == 0 {
			continue
		}

		created, err := svc.ImportRepoDependencies(ctx, sourceID, targetIDs)
		if err != nil {
			return totalCreated, err
		}
		totalCreated += created
	}
	return totalCreated, nil
}

// UnpushedRepo is a repo with uncommitted or unpushed work.
type UnpushedRepo struct {
	ID                    string
	Name                  string
	Path                  string
	HasUncommittedChanges bool
	HasUnpushedCommits    bool
}

// FindUnpushed scans an organization directory for repos with unpushed work.
func FindUnpushed(dirPath string) ([]UnpushedRepo, error) {
	org := filepath.Base(dirPath)

	results, err := scanner.ScanDirectoryWithProgress(dirPath, nil, scanner.ScanOptions{
		CheckUnpushed: true,
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", dirPath, err)
	}

	var unpushed []UnpushedRepo
	for _, r := range results {
		if !r.IsGitRepo || !r.NeedsPush() {
			continue
		}
		unpushed = append(unpushed, UnpushedRepo{
			ID:                    fmt.Sprintf("github.com/%s/%s", org, r.Name),
			Name:                  r.Name,
			Path:                  filepath.Join(dirPath, r.Name),
			HasUncommittedChanges: r.HasUncommittedChanges,
			HasUnpushedCommits:    r.HasUnpushedCommits,
		})
	}
	return unpushed, nil
}

// DependencyOrder returns repos in topological order (dependencies first).
func DependencyOrder(dirPath string) ([]OrderedRepo, error) {
	org := filepath.Base(dirPath)

	results, err := scanner.ScanDirectory(dirPath)
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", dirPath, err)
	}

	sorted, _ := scanner.TopologicalSort(results)

	var ordered []OrderedRepo
	for i, r := range sorted {
		deps := internalDeps(r, results, org)
		ordered = append(ordered, OrderedRepo{
			Position: i + 1,
			ID:       fmt.Sprintf("github.com/%s/%s", org, r.Name),
			Name:     r.Name,
			Deps:     deps,
		})
	}
	return ordered, nil
}

// OrderedRepo is a repo with its position in the dependency order.
type OrderedRepo struct {
	Position int
	ID       string
	Name     string
	Deps     []string
}

func internalDeps(r scanner.RepoResult, all []scanner.RepoResult, _ string) []string {
	moduleToName := make(map[string]string)
	for _, a := range all {
		if a.ModuleName != "" {
			moduleToName[a.ModuleName] = a.Name
		}
	}

	var deps []string
	for _, d := range r.Dependencies {
		if name, ok := moduleToName[d]; ok {
			deps = append(deps, name)
		}
	}
	return deps
}

// MultiOrgScanAndImport scans multiple org directories under a common parent.
// Example: parentDir = ~/go/src/github.com, orgs = ["plexusone", "grokify"]
func MultiOrgScanAndImport(ctx context.Context, svc *service.Service, parentDir string, orgs []string, opts ScanOptions) ([]*ScanResult, error) {
	var results []*ScanResult
	for _, org := range orgs {
		orgDir := filepath.Join(parentDir, org)
		res, err := ScanAndImport(ctx, svc, orgDir, opts)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				continue
			}
			return results, fmt.Errorf("scan %s: %w", org, err)
		}
		results = append(results, res)
	}
	return results, nil
}
