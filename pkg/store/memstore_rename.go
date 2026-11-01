package store

import (
	"context"
	"fmt"
	"slices"
)

// RenameRepository moves a repository to a new ID. The in-memory store holds
// one lock for the whole operation, so a rename is all or nothing.
func (m *MemStore) RenameRepository(_ context.Context, from, to *Repository, opts RenameRepositoryOptions) (*RenameRepositoryResult, error) {
	if from == nil || to == nil || from.ID == "" || to.ID == "" || from.ID == to.ID {
		return nil, fmt.Errorf("rename repository: need two different repository IDs")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	old, ok := m.repositories[from.ID]
	if !ok {
		return nil, fmt.Errorf("repository %s not found", from.ID)
	}
	res := &RenameRepositoryResult{}
	_, exists := m.repositories[to.ID]
	res.Created = !exists
	apply := !opts.DryRun

	for _, r := range m.rmis {
		if r.RepositoryID == from.ID {
			res.RMIs++
			if apply {
				r.RepositoryID = to.ID
			}
		}
		if r.ContextSpec != nil && slices.Contains(r.ContextSpec.ExtraRepos, from.ID) {
			res.RMIContextSpecs++
			if apply {
				r.ContextSpec.ExtraRepos = renameInList(r.ContextSpec.ExtraRepos, from.ID, to.ID)
			}
		}
	}
	for _, d := range m.specDocuments {
		if d.RepositoryID == from.ID {
			res.SpecDocuments++
			if apply {
				d.RepositoryID = to.ID
			}
		}
	}
	if opts.MoveReleases {
		for _, r := range m.releases {
			if r.RepositoryID == from.ID {
				res.Releases++
				if apply {
					r.RepositoryID = to.ID
				}
			}
		}
	}
	for _, i := range m.initiatives {
		if i.HomeRepo == from.ID {
			res.Initiatives++
			if apply {
				i.HomeRepo = to.ID
			}
		}
	}
	for _, r := range m.prismRoadmaps {
		if r.RepositoryID == from.ID {
			res.PRISMRoadmaps++
			if apply {
				r.RepositoryID = to.ID
			}
		}
	}
	for _, g := range m.prismGoals {
		if g.RepositoryID == from.ID {
			res.PRISMGoals++
			if apply {
				g.RepositoryID = to.ID
			}
		}
	}
	for _, d := range m.prismDocuments {
		if d.RepositoryID == from.ID {
			res.PRISMDocuments++
			if apply {
				d.RepositoryID = to.ID
			}
		}
	}
	for _, r := range m.devxPeriodReports {
		if r.RepositoryID == from.ID {
			res.DevXReports++
			if apply {
				r.RepositoryID = to.ID
			}
		}
	}

	// Dependencies: repoint each edge, and drop one that would point a
	// repository at itself or repeat an edge that is already there.
	key := func(d *RepositoryDependency) string {
		return d.SourceRepositoryID + "\x00" + d.TargetRepositoryID + "\x00" + d.DependencyType
	}
	seen := map[string]bool{}
	for _, d := range m.repoDeps {
		if d.SourceRepositoryID != from.ID && d.TargetRepositoryID != from.ID {
			seen[key(d)] = true
		}
	}
	var kept []*RepositoryDependency
	for _, d := range m.repoDeps {
		if d.SourceRepositoryID != from.ID && d.TargetRepositoryID != from.ID {
			kept = append(kept, d)
			continue
		}
		moved := *d
		if moved.SourceRepositoryID == from.ID {
			moved.SourceRepositoryID = to.ID
		}
		if moved.TargetRepositoryID == from.ID {
			moved.TargetRepositoryID = to.ID
		}
		if moved.SourceRepositoryID == moved.TargetRepositoryID || seen[key(&moved)] {
			res.DependenciesDropped++
			continue
		}
		seen[key(&moved)] = true
		res.Dependencies++
		kept = append(kept, &moved)
	}

	if !apply {
		return res, nil
	}
	m.repoDeps = kept
	if !exists {
		created := *to
		m.repositories[to.ID] = &created
	}
	archived := *old
	archived.Status = "archived"
	archived.SupersededBy = to.ID
	m.repositories[from.ID] = &archived
	return res, nil
}

// renameInList returns list with every from replaced by to, and to listed once.
func renameInList(list []string, from, to string) []string {
	var out []string
	for _, v := range list {
		if v == from {
			v = to
		}
		if !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}
