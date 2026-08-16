package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// GHRelease is a GitHub Release fetched via the API.
type GHRelease struct {
	TagName     string
	Name        string
	PublishedAt time.Time
	HTMLURL     string
	Body        string
	Draft       bool
	Prerelease  bool
}

// GitHubReleasesLookup fetches all releases for a repository. Injectable
// for tests; GHReleasesLookup is the production implementation.
type GitHubReleasesLookup func(ctx context.Context, owner, name string) ([]GHRelease, error)

type ghReleaseJSON struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
}

// GHReleasesLookup fetches releases via the gh CLI (same tool and auth as
// visibility.go's GHVisibilityLookup). --paginate follows all pages.
func GHReleasesLookup(ctx context.Context, owner, name string) ([]GHRelease, error) {
	// #nosec G204 -- owner/name come from the local registry catalog
	// (already-validated repo IDs), not untrusted input.
	out, err := exec.CommandContext(ctx, "gh", "api",
		fmt.Sprintf("repos/%s/%s/releases", owner, name), "--paginate").Output()
	if err != nil {
		return nil, fmt.Errorf("gh api releases for %s/%s: %w", owner, name, err)
	}
	// --paginate concatenates JSON arrays back to back for multi-page
	// results; decode as a stream rather than a single array.
	dec := json.NewDecoder(strings.NewReader(string(out)))
	var result []GHRelease
	for {
		var page []ghReleaseJSON
		if err := dec.Decode(&page); err != nil {
			break // end of stream
		}
		for _, r := range page {
			gr := GHRelease{
				TagName:    r.TagName,
				Name:       r.Name,
				HTMLURL:    r.HTMLURL,
				Body:       r.Body,
				Draft:      r.Draft,
				Prerelease: r.Prerelease,
			}
			if t, err := time.Parse(time.RFC3339, r.PublishedAt); err == nil {
				gr.PublishedAt = t
			}
			result = append(result, gr)
		}
	}
	return result, nil
}

// GitHubReleaseIngestResult summarizes a GitHub Releases API ingest run.
// Distinguishes gap-filled releases (existed on GitHub, absent locally —
// the completeness gap this RMI closes) from enrichment of releases
// ingest already knew about.
type GitHubReleaseIngestResult struct {
	RepoID        string
	Fetched       int
	GapFilled     int // new Release rows: existed on GitHub, no local tag/CHANGELOG entry found them
	Enriched      int // existing rows: URL and/or Body backfilled (only where previously empty)
	Unchanged     int
	DraftsSkipped int
	Err           error
}

const maxReleaseBodyLen = 4000

// GitHubReleases fetches a repository's GitHub Releases and reconciles
// them against the store. Merge semantics are deliberately conservative:
//   - A release GitHub knows about but the store doesn't is created
//     (gap-filled) with no initiative/RMI associations — there is no
//     local commit range to walk for trailers, so none are claimed.
//   - A release already in the store gets URL/Body filled in only where
//     currently empty; ReleasedAt, Tag, and every association are left
//     untouched — a CHANGELOG.json-derived date is not something a
//     GitHub API timestamp gets to overwrite.
//   - Drafts are skipped (never actually shipped).
func GitHubReleases(ctx context.Context, svc *service.Service, repoID string, lookup GitHubReleasesLookup) (*GitHubReleaseIngestResult, error) {
	repo, err := svc.Store.GetRepository(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("get repository %s: %w", repoID, err)
	}
	owner, name, err := splitRepoID(repo.ID)
	if err != nil {
		return nil, err
	}

	ghReleases, err := lookup(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	res := &GitHubReleaseIngestResult{RepoID: repoID, Fetched: len(ghReleases)}
	for _, gr := range ghReleases {
		if gr.Draft {
			res.DraftsSkipped++
			continue
		}
		id := service.ReleaseID(repoID, gr.TagName)
		body := truncateBody(gr.Body)

		existing, getErr := svc.Store.GetRelease(ctx, id)
		if getErr != nil {
			// Gap-filled: GitHub has it, we don't.
			rel := &store.Release{
				ID:           id,
				RepositoryID: repoID,
				Tag:          gr.TagName,
				ReleasedAt:   gr.PublishedAt,
				URL:          gr.HTMLURL,
				Body:         body,
				CreatedAt:    time.Now().UTC(),
				UpdatedAt:    time.Now().UTC(),
			}
			if err := svc.Store.CreateRelease(ctx, rel); err != nil {
				return res, fmt.Errorf("create gap-filled release %s: %w", id, err)
			}
			res.GapFilled++
			continue
		}

		changed := false
		if existing.URL == "" && gr.HTMLURL != "" {
			existing.URL = gr.HTMLURL
			changed = true
		}
		if existing.Body == "" && body != "" {
			existing.Body = body
			changed = true
		}
		if !changed {
			res.Unchanged++
			continue
		}
		existing.UpdatedAt = time.Now().UTC()
		if err := svc.Store.UpdateRelease(ctx, existing); err != nil {
			return res, fmt.Errorf("enrich release %s: %w", id, err)
		}
		res.Enriched++
	}
	return res, nil
}

// GitHubReleasesAll runs GitHubReleases for every repository with a local
// path, sequentially (GitHub API rate limits make heavy parallelism
// counterproductive here, unlike the git-only ingest paths).
func GitHubReleasesAll(ctx context.Context, svc *service.Service, lookup GitHubReleasesLookup) ([]*GitHubReleaseIngestResult, error) {
	repos, err := svc.ListReposWithLocalPath(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	var out []*GitHubReleaseIngestResult
	for _, r := range repos {
		res, err := GitHubReleases(ctx, svc, r.ID, lookup)
		if err != nil {
			out = append(out, &GitHubReleaseIngestResult{RepoID: r.ID, Err: err})
			continue
		}
		out = append(out, res)
	}
	return out, nil
}

func truncateBody(s string) string {
	if len(s) <= maxReleaseBodyLen {
		return s
	}
	return s[:maxReleaseBodyLen]
}

// splitRepoID splits "github.com/org/name" into ("org", "name").
func splitRepoID(id string) (owner, name string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("malformed repository ID %q", id)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}
