package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ProductBuildersHQ/visionstudio/pkg/evidence"
	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
)

// ChangelogResult summarizes what a changelog ingest found.
type ChangelogResult struct {
	RepoID        string
	ReleasesRead  int
	EntriesRead   int
	EvidenceAdded int
}

// changelogFile is the top-level CHANGELOG.json structure.
type changelogFile struct {
	Releases []changelogRelease `json:"releases"`
}

type changelogRelease struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Commit  string `json:"commit"`
	// Category arrays — each entry may have an optional rmi_ref.
	Added        []changelogEntry `json:"added"`
	Fixed        []changelogEntry `json:"fixed"`
	Changed      []changelogEntry `json:"changed"`
	Deprecated   []changelogEntry `json:"deprecated"`
	Removed      []changelogEntry `json:"removed"`
	Security     []changelogEntry `json:"security"`
	Tests        []changelogEntry `json:"tests"`
	Docs         []changelogEntry `json:"docs"`
	Dependencies []changelogEntry `json:"dependencies"`
}

type changelogEntry struct {
	Description string `json:"description"`
	Commit      string `json:"commit"`
	RMIRef      string `json:"rmi_ref"`
}

// Changelog reads a CHANGELOG.json file from a repository and creates
// evidence rows for entries that reference RMIs (via the rmi_ref field)
// and for releases that have a commit SHA.
func Changelog(ctx context.Context, svc *service.Service, repoID string) (*ChangelogResult, error) {
	repo, err := svc.Store.GetRepository(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("get repository %s: %w", repoID, err)
	}
	if repo.LocalPath == "" {
		return nil, fmt.Errorf("repository %s has no local_path", repoID)
	}

	path := filepath.Join(repo.LocalPath, "CHANGELOG.json")
	return ChangelogFromFile(ctx, svc, repoID, path)
}

// ChangelogFromFile ingests a specific CHANGELOG.json file.
func ChangelogFromFile(ctx context.Context, svc *service.Service, repoID, path string) (*ChangelogResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var cl changelogFile
	if err := json.Unmarshal(data, &cl); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	result := &ChangelogResult{
		RepoID:       repoID,
		ReleasesRead: len(cl.Releases),
	}

	for _, rel := range cl.Releases {
		allEntries := collectEntries(rel)
		result.EntriesRead += len(allEntries)

		for _, entry := range allEntries {
			if entry.RMIRef == "" {
				continue
			}
			refs := evidence.ParseTrailer(entry.RMIRef)
			for _, rmiID := range refs {
				ref := entry.Commit
				if ref == "" {
					ref = fmt.Sprintf("changelog:%s:%s", rel.Version, truncate(entry.Description, 50))
				}
				if _, err := svc.AddEvidence(ctx, rmiID, "changelog", ref); err != nil {
					continue
				}
				result.EvidenceAdded++
			}
		}
	}

	return result, nil
}

func collectEntries(rel changelogRelease) []changelogEntry {
	var all []changelogEntry
	all = append(all, rel.Added...)
	all = append(all, rel.Fixed...)
	all = append(all, rel.Changed...)
	all = append(all, rel.Deprecated...)
	all = append(all, rel.Removed...)
	all = append(all, rel.Security...)
	all = append(all, rel.Tests...)
	all = append(all, rel.Docs...)
	all = append(all, rel.Dependencies...)
	return all
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
