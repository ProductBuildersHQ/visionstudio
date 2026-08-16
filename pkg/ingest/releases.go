package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/grokify/gogit"

	"github.com/ProductBuildersHQ/visionstudio/pkg/evidence"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
)

// ReleaseIngestResult summarizes a release ingest (backfill) run for one
// repository, including the trailer-coverage numbers the roadmap requires:
// silent truncation reads as "covered everything" when it didn't.
type ReleaseIngestResult struct {
	RepoID              string
	ReleasesUpserted    int
	WithChangelog       int // releases whose metadata came from CHANGELOG.json
	TagsSkipped         int // non-semverish tags filtered out
	ChangelogOnly       int // changelog versions with no matching git tag (not ingested)
	InitiativeLinks     int
	RMILinks            int
	CommitsWalked       int
	CommitsWithTrailers int
	UnresolvedRMIRefs   int // trailer RMI IDs not present in the store
	Err                 error
}

// Coverage returns the fraction of walked commits carrying Refs trailers.
func (r *ReleaseIngestResult) Coverage() float64 {
	if r.CommitsWalked == 0 {
		return 0
	}
	return float64(r.CommitsWithTrailers) / float64(r.CommitsWalked)
}

// semverishTag matches v-prefixed or bare semver-like tags (v1.2.3, 0.4.0,
// v2.0.0-rc.1). Everything else is skipped unless allTags is set.
var semverishTag = regexp.MustCompile(`^v?\d+\.\d+(\.\d+)?([-+.].*)?$`)

// Releases ingests per-repo releases from git tags with CHANGELOG.json as
// the primary metadata source, auto-associating each release to the RMIs
// and initiatives named by Refs trailers in its tag range (trailer-chain
// association — no time-proximity guessing).
func Releases(ctx context.Context, svc *service.Service, repoID string, allTags bool) (*ReleaseIngestResult, error) {
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
	tagDates, err := r.TagsWithDates(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	result := &ReleaseIngestResult{RepoID: repoID}

	type tagInfo struct {
		name string
		date time.Time
	}
	var tags []tagInfo
	for name, date := range tagDates {
		if !allTags && !semverishTag.MatchString(name) {
			result.TagsSkipped++
			continue
		}
		tags = append(tags, tagInfo{name: name, date: date})
	}
	sort.Slice(tags, func(i, j int) bool {
		if !tags[i].date.Equal(tags[j].date) {
			return tags[i].date.Before(tags[j].date)
		}
		return tags[i].name < tags[j].name
	})

	clReleases := readChangelogReleases(repo.LocalPath)
	tagged := make(map[string]bool, len(tags))
	for _, t := range tags {
		tagged[t.name] = true
	}
	for version := range clReleases {
		if !tagged[version] {
			result.ChangelogOnly++
		}
	}

	prev := ""
	for _, t := range tags {
		walked, withTrailers, rmiIDs, err := gitRangeTrailerRMIs(ctx, repo.LocalPath, prev, t.name)
		if err != nil {
			return result, fmt.Errorf("walk %s..%s: %w", prev, t.name, err)
		}
		result.CommitsWalked += walked
		result.CommitsWithTrailers += withTrailers

		var initIDs []string
		var knownRMIs []string
		seenInit := map[string]bool{}
		for _, rmiID := range rmiIDs {
			rmi, err := svc.Store.GetRMI(ctx, rmiID)
			if err != nil {
				result.UnresolvedRMIRefs++
				continue
			}
			knownRMIs = append(knownRMIs, rmiID)
			if rmi.InitiativeID != "" && !seenInit[rmi.InitiativeID] {
				seenInit[rmi.InitiativeID] = true
				initIDs = append(initIDs, rmi.InitiativeID)
			}
		}

		releasedAt := t.date
		notesRef := ""
		if cl, ok := clReleases[t.name]; ok {
			notesRef = "CHANGELOG.json#" + t.name
			if parsed, err := time.Parse("2006-01-02", cl.Date); err == nil {
				releasedAt = parsed
			}
			result.WithChangelog++
		}

		if _, err := svc.RecordRelease(ctx, repoID, t.name, releasedAt, "", notesRef, initIDs, knownRMIs); err != nil {
			return result, fmt.Errorf("record release %s@%s: %w", repoID, t.name, err)
		}
		result.ReleasesUpserted++
		result.InitiativeLinks += len(initIDs)
		result.RMILinks += len(knownRMIs)
		prev = t.name
	}

	return result, nil
}

// ReleasesAll ingests releases for every repository with a local path.
// Individual repo errors are captured in their result, not fatal.
func ReleasesAll(ctx context.Context, svc *service.Service, allTags bool, workers int) ([]*ReleaseIngestResult, error) {
	repos, err := svc.ListReposWithLocalPath(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	ids := make([]string, len(repos))
	for i, r := range repos {
		ids[i] = r.ID
	}
	results := gogit.RunAllPaths(ctx, ids, func(ctx context.Context, repoID string) (*ReleaseIngestResult, error) {
		return Releases(ctx, svc, repoID, allTags)
	}, workers)

	out := make([]*ReleaseIngestResult, 0, len(results))
	for _, r := range results {
		if r.Err != nil {
			out = append(out, &ReleaseIngestResult{RepoID: r.Path, Err: r.Err})
			continue
		}
		out = append(out, r.Value)
	}
	return out, nil
}

// ReleaseIngestSummary aggregates per-repo ingest results — the totals a
// CLI, MCP tool, or SDK caller reports after a backfill run.
type ReleaseIngestSummary struct {
	Releases            int
	InitiativeLinks     int
	RMILinks            int
	CommitsWalked       int
	CommitsWithTrailers int
	UnresolvedRMIRefs   int
	ReposWithErrors     int
}

// Coverage returns the aggregate fraction of walked commits with trailers.
func (s *ReleaseIngestSummary) Coverage() float64 {
	if s.CommitsWalked == 0 {
		return 0
	}
	return float64(s.CommitsWithTrailers) / float64(s.CommitsWalked)
}

// SummarizeReleaseResults folds per-repo results into run totals.
func SummarizeReleaseResults(results []*ReleaseIngestResult) *ReleaseIngestSummary {
	sum := &ReleaseIngestSummary{}
	for _, r := range results {
		if r.Err != nil {
			sum.ReposWithErrors++
			continue
		}
		sum.Releases += r.ReleasesUpserted
		sum.InitiativeLinks += r.InitiativeLinks
		sum.RMILinks += r.RMILinks
		sum.CommitsWalked += r.CommitsWalked
		sum.CommitsWithTrailers += r.CommitsWithTrailers
		sum.UnresolvedRMIRefs += r.UnresolvedRMIRefs
	}
	return sum
}

// readChangelogReleases returns CHANGELOG.json releases keyed by version,
// or an empty map when the file is absent or unparseable (tags remain the
// fallback source — a broken changelog never blocks ingest).
func readChangelogReleases(localPath string) map[string]changelogRelease {
	out := map[string]changelogRelease{}
	data, err := os.ReadFile(filepath.Join(localPath, "CHANGELOG.json"))
	if err != nil {
		return out
	}
	var cl changelogFile
	if err := json.Unmarshal(data, &cl); err != nil {
		return out
	}
	for _, rel := range cl.Releases {
		out[rel.Version] = rel
	}
	return out
}

// gitRangeTrailerRMIs walks the commit range (prevTag, tag] and returns
// the commit count, how many commits carried Refs trailers, and the
// deduplicated RMI IDs named by those trailers.
func gitRangeTrailerRMIs(ctx context.Context, path, prevTag, tag string) (walked, withTrailers int, rmiIDs []string, err error) {
	rangeSpec := tag
	if prevTag != "" {
		rangeSpec = prevTag + ".." + tag
	}
	// #nosec G204 -- path and tags come from the local registry and the
	// repository's own tag list, not untrusted input.
	cmd := exec.CommandContext(ctx, "git", "-C", path, "log",
		"--format=%H%x1f%(trailers:key=Refs,valueonly)%x1e", rangeSpec)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, nil, fmt.Errorf("git log %s: %w", rangeSpec, err)
	}

	seen := map[string]bool{}
	records := strings.Split(string(out), "\x1e")
	for _, rec := range records {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		walked++
		parts := strings.SplitN(rec, "\x1f", 2)
		if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
			continue
		}
		var found bool
		for _, line := range strings.Split(parts[1], "\n") {
			for _, id := range evidence.ParseTrailer(line) {
				found = true
				if !seen[id] {
					seen[id] = true
					rmiIDs = append(rmiIDs, id)
				}
			}
		}
		if found {
			withTrailers++
		}
	}
	sort.Strings(rmiIDs)
	return walked, withTrailers, rmiIDs, nil
}
