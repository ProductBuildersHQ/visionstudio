package store

import (
	"context"
	"fmt"
	"sort"
)

func (m *MemStore) CreateRelease(_ context.Context, rel *Release) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.releases[rel.ID]; exists {
		return fmt.Errorf("release %s already exists", rel.ID)
	}
	m.releases[rel.ID] = rel
	return nil
}

func (m *MemStore) GetRelease(_ context.Context, id string) (*Release, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rel, ok := m.releases[id]
	if !ok {
		return nil, fmt.Errorf("release %s not found", id)
	}
	return rel, nil
}

func (m *MemStore) ListReleases(_ context.Context) ([]*Release, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Release, 0, len(m.releases))
	for _, rel := range m.releases {
		result = append(result, rel)
	}
	sortReleases(result)
	return result, nil
}

func (m *MemStore) ListReleasesByRepo(_ context.Context, repoID string) ([]*Release, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Release
	for _, rel := range m.releases {
		if rel.RepositoryID == repoID {
			result = append(result, rel)
		}
	}
	sortReleases(result)
	return result, nil
}

func (m *MemStore) ListReleasesByInitiative(_ context.Context, initiativeID string) ([]*Release, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Release
	for _, rel := range m.releases {
		for _, id := range rel.InitiativeIDs {
			if id == initiativeID {
				result = append(result, rel)
				break
			}
		}
	}
	sortReleases(result)
	return result, nil
}

func (m *MemStore) UpdateRelease(_ context.Context, rel *Release) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.releases[rel.ID]; !ok {
		return fmt.Errorf("release %s not found", rel.ID)
	}
	m.releases[rel.ID] = rel
	return nil
}

func (m *MemStore) DeleteRelease(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.releases[id]; !ok {
		return fmt.Errorf("release %s not found", id)
	}
	delete(m.releases, id)
	return nil
}

// sortReleases orders newest-first by release time, then by ID for
// deterministic output.
func sortReleases(rels []*Release) {
	sort.Slice(rels, func(i, j int) bool {
		if !rels[i].ReleasedAt.Equal(rels[j].ReleasedAt) {
			return rels[i].ReleasedAt.After(rels[j].ReleasedAt)
		}
		return rels[i].ID < rels[j].ID
	})
}
