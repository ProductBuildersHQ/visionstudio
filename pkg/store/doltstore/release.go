package doltstore

import (
	"context"
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/ent"
	"github.com/ProductBuildersHQ/visionstudio/ent/initiative"
	"github.com/ProductBuildersHQ/visionstudio/ent/release"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// ---------------------------------------------------------------------------
// Release CRUD
// ---------------------------------------------------------------------------

func entReleaseToStore(r *ent.Release) *store.Release {
	sr := &store.Release{
		ID:           r.ID,
		RepositoryID: r.RepositoryID,
		Tag:          r.Tag,
		ReleasedAt:   r.ReleasedAt,
		URL:          r.URL,
		NotesRef:     r.NotesRef,
		Body:         r.Body,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
	for _, in := range r.Edges.Initiatives {
		sr.InitiativeIDs = append(sr.InitiativeIDs, in.ID)
	}
	for _, rmi := range r.Edges.RoadmapItems {
		sr.RMIIDs = append(sr.RMIIDs, rmi.ID)
	}
	return sr
}

func (d *DoltStore) CreateRelease(ctx context.Context, rel *store.Release) error {
	b := d.client.Release.Create().
		SetID(rel.ID).
		SetRepositoryID(rel.RepositoryID).
		SetTag(rel.Tag).
		SetReleasedAt(rel.ReleasedAt).
		SetCreatedAt(rel.CreatedAt).
		SetUpdatedAt(rel.UpdatedAt)
	if rel.URL != "" {
		b.SetURL(rel.URL)
	}
	if rel.NotesRef != "" {
		b.SetNotesRef(rel.NotesRef)
	}
	if rel.Body != "" {
		b.SetBody(rel.Body)
	}
	if len(rel.InitiativeIDs) > 0 {
		b.AddInitiativeIDs(rel.InitiativeIDs...)
	}
	if len(rel.RMIIDs) > 0 {
		b.AddRoadmapItemIDs(rel.RMIIDs...)
	}
	if _, err := b.Save(ctx); err != nil {
		return fmt.Errorf("create release %s: %w", rel.ID, err)
	}
	return nil
}

func (d *DoltStore) GetRelease(ctx context.Context, id string) (*store.Release, error) {
	r, err := d.client.Release.Query().
		Where(release.IDEQ(id)).
		WithInitiatives().
		WithRoadmapItems().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get release %s: %w", id, err)
	}
	return entReleaseToStore(r), nil
}

func (d *DoltStore) ListReleases(ctx context.Context) ([]*store.Release, error) {
	rows, err := d.client.Release.Query().
		WithInitiatives().
		WithRoadmapItems().
		Order(ent.Desc(release.FieldReleasedAt), ent.Asc(release.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}
	result := make([]*store.Release, len(rows))
	for i, r := range rows {
		result[i] = entReleaseToStore(r)
	}
	return result, nil
}

func (d *DoltStore) ListReleasesByRepo(ctx context.Context, repoID string) ([]*store.Release, error) {
	rows, err := d.client.Release.Query().
		Where(release.RepositoryIDEQ(repoID)).
		WithInitiatives().
		WithRoadmapItems().
		Order(ent.Desc(release.FieldReleasedAt), ent.Asc(release.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list releases by repo %s: %w", repoID, err)
	}
	result := make([]*store.Release, len(rows))
	for i, r := range rows {
		result[i] = entReleaseToStore(r)
	}
	return result, nil
}

func (d *DoltStore) ListReleasesByInitiative(ctx context.Context, initiativeID string) ([]*store.Release, error) {
	rows, err := d.client.Release.Query().
		Where(release.HasInitiativesWith(initiative.IDEQ(initiativeID))).
		WithInitiatives().
		WithRoadmapItems().
		Order(ent.Desc(release.FieldReleasedAt), ent.Asc(release.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list releases by initiative %s: %w", initiativeID, err)
	}
	result := make([]*store.Release, len(rows))
	for i, r := range rows {
		result[i] = entReleaseToStore(r)
	}
	return result, nil
}

func (d *DoltStore) DeleteRelease(ctx context.Context, id string) error {
	if err := d.client.Release.DeleteOneID(id).Exec(ctx); err != nil {
		return fmt.Errorf("delete release %s: %w", id, err)
	}
	return nil
}

func (d *DoltStore) UpdateRelease(ctx context.Context, rel *store.Release) error {
	b := d.client.Release.UpdateOneID(rel.ID).
		SetTag(rel.Tag).
		SetReleasedAt(rel.ReleasedAt).
		SetUpdatedAt(rel.UpdatedAt).
		ClearInitiatives().
		ClearRoadmapItems()
	if rel.URL != "" {
		b.SetURL(rel.URL)
	} else {
		b.ClearURL()
	}
	if rel.NotesRef != "" {
		b.SetNotesRef(rel.NotesRef)
	} else {
		b.ClearNotesRef()
	}
	if rel.Body != "" {
		b.SetBody(rel.Body)
	} else {
		b.ClearBody()
	}
	if len(rel.InitiativeIDs) > 0 {
		b.AddInitiativeIDs(rel.InitiativeIDs...)
	}
	if len(rel.RMIIDs) > 0 {
		b.AddRoadmapItemIDs(rel.RMIIDs...)
	}
	if _, err := b.Save(ctx); err != nil {
		return fmt.Errorf("update release %s: %w", rel.ID, err)
	}
	return nil
}
