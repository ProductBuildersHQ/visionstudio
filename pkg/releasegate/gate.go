// Package releasegate is the changelog release gate: CHANGELOG.json
// (structured-changelog) is our convention, not a market standard, so
// instead of presuming the habit, recording a release checks for it and
// offers to manufacture it — scaffolding changelog entries from
// conventional commits. Habit as tooling output, not entry fee.
package releasegate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/grokify/gogit"
)

// CheckResult reports whether a repository carries the changelog
// convention and whether a specific version is recorded in it.
type CheckResult struct {
	ChangelogPath   string
	ChangelogExists bool
	VersionPresent  bool
}

// Check inspects CHANGELOG.json in a repository working tree for the
// given version. A missing or unparseable file is reported, never fatal —
// bare-tag repositories are a supported fallback.
func Check(localPath, version string) CheckResult {
	res := CheckResult{ChangelogPath: filepath.Join(localPath, "CHANGELOG.json")}
	data, err := os.ReadFile(res.ChangelogPath)
	if err != nil {
		return res
	}
	res.ChangelogExists = true

	var doc struct {
		Releases []struct {
			Version string `json:"version"`
		} `json:"releases"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return res
	}
	for _, rel := range doc.Releases {
		if rel.Version == version {
			res.VersionPresent = true
			return res
		}
	}
	return res
}

// ScaffoldResult summarizes a generated changelog release entry.
type ScaffoldResult struct {
	ChangelogPath string
	Created       bool // file created (vs entry prepended to existing)
	Version       string
	SinceRef      string
	Entries       int
	Categories    []string
	RMIRefs       int
}

type scaffoldEntry struct {
	Description string `json:"description"`
	Commit      string `json:"commit,omitempty"`
	RMI         string `json:"rmi,omitempty"`
}

// categoryFor maps a conventional-commit type to a structured-changelog
// category. Unrecognized and non-conventional commits land in "changed" —
// the deterministic fallback (AI classification is a future refinement).
func categoryFor(commitType string) string {
	switch commitType {
	case "feat":
		return "added"
	case "fix":
		return "fixed"
	case "docs":
		return "docs"
	case "test":
		return "tests"
	case "build", "deps":
		return "dependencies"
	case "security":
		return "security"
	case "revert":
		return "removed"
	default:
		return "changed"
	}
}

// Scaffold generates a CHANGELOG.json release entry for version from the
// conventional commits since the previous semver-like tag (or the whole
// history when none exists), preserving every unknown field in an
// existing file. It never overwrites an entry that already exists.
func Scaffold(ctx context.Context, localPath, version string) (*ScaffoldResult, error) {
	check := Check(localPath, version)
	if check.VersionPresent {
		return nil, fmt.Errorf("CHANGELOG.json already contains %s", version)
	}

	r, err := gogit.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("open git repo: %w", err)
	}
	sinceRef := latestSemverishTag(ctx, r)
	commits, err := r.Log(ctx, gogit.LogOptions{SinceCommit: sinceRef, Reverse: true})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits since %s to scaffold from", refOrStart(sinceRef))
	}

	res := &ScaffoldResult{
		ChangelogPath: check.ChangelogPath,
		Version:       version,
		SinceRef:      refOrStart(sinceRef),
	}

	byCategory := map[string][]scaffoldEntry{}
	for _, c := range commits {
		subject := c.Subject
		category := "changed"
		if cc := c.ParseConventional(); cc != nil {
			category = categoryFor(cc.Type)
			subject = cc.Subject
		}
		entry := scaffoldEntry{Description: subject, Commit: shortHash(c.Hash)}
		if refs := c.TrailerValue("Refs"); refs != "" {
			entry.RMI = strings.TrimSpace(refs)
			res.RMIRefs++
		}
		byCategory[category] = append(byCategory[category], entry)
		res.Entries++
	}

	release := map[string]any{
		"version": version,
		"date":    time.Now().UTC().Format("2006-01-02"),
	}
	for cat, entries := range byCategory {
		release[cat] = entries
		res.Categories = append(res.Categories, cat)
	}
	sort.Strings(res.Categories)

	doc, created, err := loadOrInitChangelog(check.ChangelogPath, localPath)
	if err != nil {
		return nil, err
	}
	res.Created = created

	existing, _ := doc["releases"].([]any)
	doc["releases"] = append([]any{release}, existing...)

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal changelog: %w", err)
	}
	if err := os.WriteFile(check.ChangelogPath, append(out, '\n'), 0o644); err != nil { // #nosec G306 -- changelog is a public repo artifact
		return nil, fmt.Errorf("write %s: %w", check.ChangelogPath, err)
	}
	return res, nil
}

// loadOrInitChangelog reads CHANGELOG.json as a generic document so
// unknown fields survive the round-trip, or initializes a minimal
// structured-changelog skeleton when the file is absent.
func loadOrInitChangelog(path, localPath string) (map[string]any, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		name := filepath.Base(localPath)
		return map[string]any{
			"project":  name,
			"releases": []any{},
		}, true, nil
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, false, fmt.Errorf("parse %s (fix it before scaffolding): %w", path, err)
	}
	return doc, false, nil
}

func latestSemverishTag(ctx context.Context, r *gogit.Repo) string {
	tagDates, err := r.TagsWithDates(ctx)
	if err != nil {
		return ""
	}
	var latest string
	var latestTime time.Time
	for name, date := range tagDates {
		if !semverishTag(name) {
			continue
		}
		if latest == "" || date.After(latestTime) {
			latest = name
			latestTime = date
		}
	}
	return latest
}

// semverishTag mirrors the ingest filter: v-prefixed or bare semver-like.
func semverishTag(name string) bool {
	rest := strings.TrimPrefix(name, "v")
	parts := strings.SplitN(rest, ".", 3)
	if len(parts) < 2 {
		return false
	}
	for i, p := range parts[:2] {
		_ = i
		if p == "" || strings.IndexFunc(p, func(r rune) bool { return r < '0' || r > '9' }) != -1 {
			return false
		}
	}
	return true
}

func shortHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}

func refOrStart(ref string) string {
	if ref == "" {
		return "start of history"
	}
	return ref
}
