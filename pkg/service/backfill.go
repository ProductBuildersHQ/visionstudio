package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// UnmatchedReleases returns releases with zero initiative associations —
// the AI-assisted historical backfill queue (RMI-VISIONSTUDIO-315).
// repoFilter, if non-empty, restricts to one repository. Oldest first:
// the deepest pre-adoption history is the least likely to ever get
// trailer-derived evidence, so it is reviewed first while it's still
// findable.
func (s *Service) UnmatchedReleases(ctx context.Context, repoFilter string) ([]*store.Release, error) {
	var rels []*store.Release
	var err error
	if repoFilter != "" {
		rels, err = s.Store.ListReleasesByRepo(ctx, repoFilter)
	} else {
		rels, err = s.Store.ListReleases(ctx)
	}
	if err != nil {
		return nil, err
	}
	var out []*store.Release
	for _, r := range rels {
		if len(r.InitiativeIDs) == 0 {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ReleasedAt.Before(out[j].ReleasedAt) })
	return out, nil
}

// BackfillCandidateInitiative is one candidate for a human (or the
// reviewing agent session) to judge against a release's evidence.
type BackfillCandidateInitiative struct {
	Initiative *store.Initiative
	// RMITitles are a few of this initiative's RMI titles in the release's
	// repo, for topical context — not a match signal by themselves.
	RMITitles []string
}

// BackfillCandidates is everything needed to review one unmatched
// release: its own evidence (tag, date, URL, body text) plus every
// initiative with a footprint in the same repository (home repo, or any
// RMI homed there). This is the full candidate set — no automated
// pre-filtering or scoring happens here; ranking judgment belongs to
// whoever reviews it (per CLAUDE.md "AI-assisted historical backfill
// matching": always Analyst inference, always human-confirmed via
// release attach, never auto-attached).
type BackfillCandidates struct {
	Release    *store.Release
	Candidates []BackfillCandidateInitiative
}

// GetBackfillCandidates assembles the review payload for one release.
func (s *Service) GetBackfillCandidates(ctx context.Context, releaseID string) (*BackfillCandidates, error) {
	rel, err := s.Store.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}

	inits, err := s.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}
	rmis, err := s.Store.ListRMIsByRepo(ctx, rel.RepositoryID)
	if err != nil {
		return nil, err
	}

	rmiTitlesByInit := map[string][]string{}
	referencedInits := map[string]bool{}
	for _, rmi := range rmis {
		if rmi.InitiativeID == "" {
			continue
		}
		referencedInits[rmi.InitiativeID] = true
		if len(rmiTitlesByInit[rmi.InitiativeID]) < 3 {
			rmiTitlesByInit[rmi.InitiativeID] = append(rmiTitlesByInit[rmi.InitiativeID], rmi.Title)
		}
	}

	var out []BackfillCandidateInitiative
	for _, in := range inits {
		if in.HomeRepo != rel.RepositoryID && !referencedInits[in.ID] {
			continue
		}
		out = append(out, BackfillCandidateInitiative{
			Initiative: in,
			RMITitles:  rmiTitlesByInit[in.ID],
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Initiative.ID < out[j].Initiative.ID })

	if len(out) == 0 {
		return nil, fmt.Errorf("no candidate initiatives found for repository %s (nothing homed there, no RMIs reference it)", rel.RepositoryID)
	}
	return &BackfillCandidates{Release: rel, Candidates: out}, nil
}
