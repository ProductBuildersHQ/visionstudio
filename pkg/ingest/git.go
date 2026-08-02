// Package ingest scans git repositories for commit trailers and
// conventional commit metadata, creating delivery evidence rows.
package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/grokify/gogit"

	"github.com/ProductBuildersHQ/visionstudio/pkg/evidence"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
)

// GitResult summarizes what a git ingest run found.
type GitResult struct {
	RepoID        string
	CommitsWalked int
	EvidenceAdded int
	Unattributed  int
	NewHighWater  string
	Err           error
}

// Git walks commits in a repository since its high-water mark, parses
// trailers and conventional commit type/scope, creates evidence rows,
// and advances the high-water mark. Branch is used as the fallback
// attribution source per TRD §8 precedence.
func Git(ctx context.Context, svc *service.Service, repoID string) (*GitResult, error) {
	repo, err := svc.Store.GetRepository(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("get repository %s: %w", repoID, err)
	}
	if repo.LocalPath == "" {
		return nil, fmt.Errorf("repository %s has no local_path", repoID)
	}

	r, err := gogit.Open(repo.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("open git repo at %s: %w", repo.LocalPath, err)
	}

	branch, err := r.Branch(ctx)
	if err != nil {
		return nil, fmt.Errorf("get branch: %w", err)
	}

	commits, err := r.Log(ctx, gogit.LogOptions{
		SinceCommit: repo.IngestHighWater,
		Reverse:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	result := &GitResult{
		RepoID:        repoID,
		CommitsWalked: len(commits),
	}

	for _, c := range commits {
		trailerValue := c.TrailerValue(evidence.TrailerKey)
		attr := evidence.Attribute(trailerValue, branch)

		if len(attr.RMIIDs) == 0 {
			result.Unattributed++
			result.NewHighWater = c.Hash
			continue
		}

		cc := c.ParseConventional()

		for _, rmiID := range attr.RMIIDs {
			ev, err := svc.AddEvidence(ctx, rmiID, "commit", c.Hash)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "Duplicate") {
					continue
				}
				return result, fmt.Errorf("add evidence for %s: %w", rmiID, err)
			}
			if cc != nil {
				ev.CommitType = cc.Type
				ev.CommitScope = cc.Scope
			}
			ev.OccurredAt = &c.AuthorDate
			result.EvidenceAdded++
		}

		result.NewHighWater = c.Hash
	}

	if result.NewHighWater != "" {
		repo.IngestHighWater = result.NewHighWater
		if err := svc.Store.UpdateRepository(ctx, repo); err != nil {
			return result, fmt.Errorf("update high-water mark: %w", err)
		}
	}

	return result, nil
}

// GitAll ingests all repositories with a local path in parallel, returning
// results for each. Errors on individual repos are captured in their result
// rather than aborting the run.
func GitAll(ctx context.Context, svc *service.Service, workers int) ([]*GitResult, error) {
	repos, err := svc.ListReposWithLocalPath(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}

	ids := make([]string, len(repos))
	for i, r := range repos {
		ids[i] = r.ID
	}

	results := gogit.RunAllPaths(ctx, ids, func(ctx context.Context, repoID string) (*GitResult, error) {
		return Git(ctx, svc, repoID)
	}, workers)

	out := make([]*GitResult, 0, len(results))
	for _, r := range results {
		if r.Err != nil {
			out = append(out, &GitResult{RepoID: r.Path, Err: r.Err})
			continue
		}
		out = append(out, r.Value)
	}
	return out, nil
}
