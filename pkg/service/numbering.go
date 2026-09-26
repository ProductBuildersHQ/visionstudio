package service

import (
	"context"
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/pkg/initiative"
	"github.com/ProductBuildersHQ/visionstudio/pkg/rmi"
)

// NextRMIID returns the next unused RMI ID for a repository's slug. The number
// is one greater than the highest existing number for that slug across every
// initiative — numbering is per repo-slug, not per initiative, because one
// slug's RMIs (e.g. RMI-SYSTEMFORGE-NNN) can span several initiatives.
//
// Numbers are allocated as max+1 and never fill gaps below the maximum, so an
// ID stays stable once assigned even if the RMI is later deleted — a reused
// number could collide with an ID already referenced in a git 'Refs:' trailer.
// Allocation is not transactional against concurrent creates; the store's
// unique-ID constraint remains the backstop for that race.
func (s *Service) NextRMIID(ctx context.Context, repoID string) (string, error) {
	slug := rmi.SlugFromRepoID(repoID)
	if slug == "" {
		return "", fmt.Errorf("cannot derive an RMI slug from repository %q", repoID)
	}
	all, err := s.Store.ListAllRMIs(ctx)
	if err != nil {
		return "", fmt.Errorf("list RMIs: %w", err)
	}
	max := 0
	for _, r := range all {
		if sl, n, ok := rmi.ParseID(r.ID); ok && sl == slug && n > max {
			max = n
		}
	}
	return rmi.FormatID(slug, max+1), nil
}

// NextInitiativeID returns the next unused initiative ID for a slug, following
// the same max+1, never-reuse-gaps policy as NextRMIID. The slug is normalized
// (uppercased, non-alphanumerics stripped); a repository ID may be passed and
// its final path segment is used.
func (s *Service) NextInitiativeID(ctx context.Context, slug string) (string, error) {
	slug = initiative.NormalizeSlug(slug)
	if slug == "" {
		return "", fmt.Errorf("an initiative slug is required")
	}
	all, err := s.Store.ListInitiatives(ctx)
	if err != nil {
		return "", fmt.Errorf("list initiatives: %w", err)
	}
	max := 0
	for _, in := range all {
		if sl, n, ok := initiative.ParseID(in.ID); ok && sl == slug && n > max {
			max = n
		}
	}
	return initiative.FormatID(slug, max+1), nil
}
