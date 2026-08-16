package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// ReleaseID composes the canonical release ID from a repository ID and tag.
func ReleaseID(repoID, tag string) string {
	return fmt.Sprintf("%s@%s", repoID, tag)
}

// RecordRelease upserts a release for a registered repository. The
// repository must exist in the registry; associations are merged (never
// silently dropped) on the upsert path.
func (s *Service) RecordRelease(ctx context.Context, repoID, tag string, releasedAt time.Time, url, notesRef string, initiativeIDs, rmiIDs []string) (*store.Release, error) {
	if repoID == "" || tag == "" {
		return nil, fmt.Errorf("repository and tag are required")
	}
	if _, err := s.Store.GetRepository(ctx, repoID); err != nil {
		return nil, fmt.Errorf("repository %s is not registered: %w", repoID, err)
	}
	now := time.Now().UTC()
	if releasedAt.IsZero() {
		releasedAt = now
	}

	id := ReleaseID(repoID, tag)
	if existing, err := s.Store.GetRelease(ctx, id); err == nil {
		existing.ReleasedAt = releasedAt
		if url != "" {
			existing.URL = url
		}
		if notesRef != "" {
			existing.NotesRef = notesRef
		}
		existing.InitiativeIDs = mergeIDs(existing.InitiativeIDs, initiativeIDs)
		existing.RMIIDs = mergeIDs(existing.RMIIDs, rmiIDs)
		existing.UpdatedAt = now
		if err := s.Store.UpdateRelease(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	rel := &store.Release{
		ID:            id,
		RepositoryID:  repoID,
		Tag:           tag,
		ReleasedAt:    releasedAt,
		URL:           url,
		NotesRef:      notesRef,
		InitiativeIDs: dedupeIDs(initiativeIDs),
		RMIIDs:        dedupeIDs(rmiIDs),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.Store.CreateRelease(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// AttachRelease adds initiative and RMI associations to a release.
func (s *Service) AttachRelease(ctx context.Context, releaseID string, initiativeIDs, rmiIDs []string) (*store.Release, error) {
	rel, err := s.Store.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	for _, id := range initiativeIDs {
		if _, err := s.Store.GetInitiative(ctx, id); err != nil {
			return nil, fmt.Errorf("initiative %s not found: %w", id, err)
		}
	}
	for _, id := range rmiIDs {
		if _, err := s.Store.GetRMI(ctx, id); err != nil {
			return nil, fmt.Errorf("RMI %s not found: %w", id, err)
		}
	}
	rel.InitiativeIDs = mergeIDs(rel.InitiativeIDs, initiativeIDs)
	rel.RMIIDs = mergeIDs(rel.RMIIDs, rmiIDs)
	rel.UpdatedAt = time.Now().UTC()
	if err := s.Store.UpdateRelease(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// DetachRelease removes initiative and RMI associations from a release.
func (s *Service) DetachRelease(ctx context.Context, releaseID string, initiativeIDs, rmiIDs []string) (*store.Release, error) {
	rel, err := s.Store.GetRelease(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	rel.InitiativeIDs = removeIDs(rel.InitiativeIDs, initiativeIDs)
	rel.RMIIDs = removeIDs(rel.RMIIDs, rmiIDs)
	rel.UpdatedAt = time.Now().UTC()
	if err := s.Store.UpdateRelease(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// GetRelease returns a release by ID.
func (s *Service) GetRelease(ctx context.Context, id string) (*store.Release, error) {
	return s.Store.GetRelease(ctx, id)
}

// ListReleases returns releases, optionally filtered by repository or
// initiative (both filters combine as AND when supplied).
func (s *Service) ListReleases(ctx context.Context, repoID, initiativeID string) ([]*store.Release, error) {
	switch {
	case initiativeID != "":
		rels, err := s.Store.ListReleasesByInitiative(ctx, initiativeID)
		if err != nil {
			return nil, err
		}
		if repoID == "" {
			return rels, nil
		}
		var out []*store.Release
		for _, r := range rels {
			if r.RepositoryID == repoID {
				out = append(out, r)
			}
		}
		return out, nil
	case repoID != "":
		return s.Store.ListReleasesByRepo(ctx, repoID)
	default:
		return s.Store.ListReleases(ctx)
	}
}

func dedupeIDs(ids []string) []string {
	return mergeIDs(nil, ids)
}

func mergeIDs(existing, add []string) []string {
	seen := make(map[string]bool, len(existing)+len(add))
	var out []string
	for _, id := range existing {
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	for _, id := range add {
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

func removeIDs(existing, remove []string) []string {
	drop := make(map[string]bool, len(remove))
	for _, id := range remove {
		drop[id] = true
	}
	var out []string
	for _, id := range existing {
		if !drop[id] {
			out = append(out, id)
		}
	}
	return out
}
