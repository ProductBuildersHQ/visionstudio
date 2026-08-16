// Package publicrail is the single implementation of the two-filter
// safety rail for external projections (the public roadmap export and the
// cloud serving layer). Every payload that leaves for a public surface
// must pass through these predicates — one implementation, every call
// site (INIT-VISIONSTUDIO-006 static export, INIT-VISIONSTUDIO-002
// serving).
//
// The two independent filters:
//
//  1. Repository level: only visibility == "public" passes. "unknown"
//     (not yet ingested, or GitHub lookup failed) NEVER passes.
//  2. Initiative level: only visibility == "public" passes; the default
//     is "internal".
//
// A public-flagged initiative whose repositories are all private exports
// nothing repo-identifying; naming a private repository publicly must be
// a deliberate act elsewhere, never a side effect here.
package publicrail

import (
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// Visibility vocabulary.
const (
	VisibilityPublic   = "public"
	VisibilityPrivate  = "private"
	VisibilityUnknown  = "unknown"
	VisibilityInternal = "internal"
)

// RepoPublic reports whether a repository may appear in public
// projections. Only an explicit "public" qualifies.
func RepoPublic(r *store.Repository) bool {
	return r != nil && r.Visibility == VisibilityPublic
}

// InitiativePublic reports whether an initiative is flagged for public
// projection. Only an explicit "public" qualifies; empty means the
// default, internal. Hidden initiatives never qualify.
func InitiativePublic(i *store.Initiative) bool {
	return i != nil && !i.Hidden && i.Visibility == VisibilityPublic
}

// FilterRepos returns only the repositories that pass the repo-level
// filter, preserving order.
func FilterRepos(repos []*store.Repository) []*store.Repository {
	var out []*store.Repository
	for _, r := range repos {
		if RepoPublic(r) {
			out = append(out, r)
		}
	}
	return out
}

// InitiativeAllowed applies both filters for an initiative and the
// repositories its public payload would reference. It passes only when
// the initiative is public AND, if any repositories are referenced, at
// least one of them is public (the payload then names only the public
// subset via FilterRepos). An initiative that references repositories
// none of which are public exports nothing.
func InitiativeAllowed(i *store.Initiative, repos []*store.Repository) bool {
	if !InitiativePublic(i) {
		return false
	}
	if len(repos) == 0 {
		return true
	}
	return len(FilterRepos(repos)) > 0
}
