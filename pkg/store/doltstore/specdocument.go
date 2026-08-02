//go:build dolt

package doltstore

import (
	"context"
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/ent"
	"github.com/ProductBuildersHQ/visionstudio/ent/specdocument"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func entSpecDocumentToStore(e *ent.SpecDocument) *store.SpecDocument {
	return &store.SpecDocument{
		ID:           e.ID,
		Organization: e.Organization,
		RepositoryID: e.RepositoryID,
		InitiativeID: e.InitiativeID,
		SpecType:     e.SpecType,
		FilePath:     e.FilePath,
		Title:        e.Title,
		Status:       e.Status,
		ContentHash:  e.ContentHash,
		SyncedAt:     e.SyncedAt,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func (d *DoltStore) CreateSpecDocument(ctx context.Context, doc *store.SpecDocument) error {
	b := d.client.SpecDocument.Create().
		SetID(doc.ID).
		SetRepositoryID(doc.RepositoryID).
		SetSpecType(doc.SpecType).
		SetFilePath(doc.FilePath).
		SetSyncedAt(doc.SyncedAt).
		SetCreatedAt(doc.CreatedAt).
		SetUpdatedAt(doc.UpdatedAt)
	if doc.Organization != "" {
		b.SetOrganization(doc.Organization)
	}
	if doc.InitiativeID != "" {
		b.SetInitiativeID(doc.InitiativeID)
	}
	if doc.Title != "" {
		b.SetTitle(doc.Title)
	}
	if doc.Status != "" {
		b.SetStatus(doc.Status)
	}
	if doc.ContentHash != "" {
		b.SetContentHash(doc.ContentHash)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create spec document: %w", err)
	}
	return nil
}

func (d *DoltStore) GetSpecDocument(ctx context.Context, id string) (*store.SpecDocument, error) {
	e, err := d.client.SpecDocument.Query().
		Where(specdocument.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get spec document %s: %w", id, err)
	}
	return entSpecDocumentToStore(e), nil
}

func (d *DoltStore) ListSpecDocuments(ctx context.Context) ([]*store.SpecDocument, error) {
	rows, err := d.client.SpecDocument.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list spec documents: %w", err)
	}
	result := make([]*store.SpecDocument, len(rows))
	for i, e := range rows {
		result[i] = entSpecDocumentToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListSpecDocumentsByRepo(ctx context.Context, repoID string) ([]*store.SpecDocument, error) {
	rows, err := d.client.SpecDocument.Query().
		Where(specdocument.RepositoryIDEQ(repoID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list spec documents by repo: %w", err)
	}
	result := make([]*store.SpecDocument, len(rows))
	for i, e := range rows {
		result[i] = entSpecDocumentToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListSpecDocumentsByInitiative(ctx context.Context, initiativeID string) ([]*store.SpecDocument, error) {
	rows, err := d.client.SpecDocument.Query().
		Where(specdocument.InitiativeIDEQ(initiativeID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list spec documents by initiative: %w", err)
	}
	result := make([]*store.SpecDocument, len(rows))
	for i, e := range rows {
		result[i] = entSpecDocumentToStore(e)
	}
	return result, nil
}

func (d *DoltStore) UpdateSpecDocument(ctx context.Context, doc *store.SpecDocument) error {
	b := d.client.SpecDocument.UpdateOneID(doc.ID).
		SetRepositoryID(doc.RepositoryID).
		SetSpecType(doc.SpecType).
		SetFilePath(doc.FilePath).
		SetSyncedAt(doc.SyncedAt).
		SetUpdatedAt(doc.UpdatedAt)
	if doc.Organization != "" {
		b.SetOrganization(doc.Organization)
	} else {
		b.ClearOrganization()
	}
	if doc.InitiativeID != "" {
		b.SetInitiativeID(doc.InitiativeID)
	} else {
		b.ClearInitiativeID()
	}
	if doc.Title != "" {
		b.SetTitle(doc.Title)
	} else {
		b.ClearTitle()
	}
	if doc.Status != "" {
		b.SetStatus(doc.Status)
	} else {
		b.ClearStatus()
	}
	if doc.ContentHash != "" {
		b.SetContentHash(doc.ContentHash)
	} else {
		b.ClearContentHash()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update spec document %s: %w", doc.ID, err)
	}
	return nil
}

func (d *DoltStore) DeleteSpecDocument(ctx context.Context, id string) error {
	err := d.client.SpecDocument.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete spec document %s: %w", id, err)
	}
	return nil
}
