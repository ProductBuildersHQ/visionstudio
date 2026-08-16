package service

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// VisibilityLookup resolves a repository's visibility on GitHub.
// It returns "public" or "private"; errors leave the stored value untouched.
type VisibilityLookup func(ctx context.Context, org, name string) (string, error)

// GHVisibilityLookup queries GitHub via the gh CLI.
func GHVisibilityLookup(ctx context.Context, org, name string) (string, error) {
	// #nosec G204 -- org and name come from the local registry catalog
	// (validated repo IDs), not untrusted input; gh is invoked by name.
	out, err := exec.CommandContext(ctx, "gh", "repo", "view",
		fmt.Sprintf("%s/%s", org, name), "--json", "visibility", "--jq", ".visibility").Output()
	if err != nil {
		return "", fmt.Errorf("gh repo view %s/%s: %w", org, name, err)
	}
	return NormalizeVisibility(strings.TrimSpace(string(out))), nil
}

// NormalizeVisibility maps GitHub API values to the stored vocabulary.
// GitHub reports PUBLIC/PRIVATE/INTERNAL; internal is not public, so it
// maps to private. Anything unrecognized is unknown — and unknown must
// never be treated as public.
func NormalizeVisibility(v string) string {
	switch strings.ToLower(v) {
	case "public":
		return "public"
	case "private", "internal":
		return "private"
	default:
		return "unknown"
	}
}

// VisibilityRefreshResult summarizes a refresh run.
type VisibilityRefreshResult struct {
	Updated   int
	Unchanged int
	Errors    []string
}

// RefreshVisibility updates repository visibility from GitHub via lookup.
// repoID scopes the refresh to one repository; empty refreshes all.
// Lookup failures leave the stored value untouched (unknown stays unknown).
func (s *Service) RefreshVisibility(ctx context.Context, lookup VisibilityLookup, repoID string) (*VisibilityRefreshResult, error) {
	var repos []*store.Repository
	if repoID != "" {
		repo, err := s.Store.GetRepository(ctx, repoID)
		if err != nil {
			return nil, err
		}
		repos = []*store.Repository{repo}
	} else {
		var err error
		repos, err = s.Store.ListRepositories(ctx)
		if err != nil {
			return nil, err
		}
	}

	res := &VisibilityRefreshResult{}
	for _, repo := range repos {
		vis, err := lookup(ctx, repo.Organization, repo.RepositoryName)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", repo.ID, err))
			continue
		}
		vis = NormalizeVisibility(vis)
		if vis == "unknown" || vis == repo.Visibility {
			res.Unchanged++
			continue
		}
		repo.Visibility = vis
		if err := s.Store.UpdateRepository(ctx, repo); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: update: %v", repo.ID, err))
			continue
		}
		res.Updated++
	}
	return res, nil
}
