

package doltstore

import (
	"context"
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/ent"
	"github.com/ProductBuildersHQ/visionstudio/ent/prismdocument"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func entPRISMDocumentToStore(e *ent.PRISMDocument) *store.PRISMDocument {
	return &store.PRISMDocument{
		ID:            e.ID,
		Organization:  e.Organization,
		RepositoryID:  e.RepositoryID,
		Name:          e.Name,
		Description:   e.Description,
		Version:       e.Version,
		Domains:       e.Domains,
		Layers:        e.Layers,
		Metrics:       e.Metrics,
		Maturity:      e.Maturity,
		SLIState:      e.SliState,
		MaturityState: e.MaturityState,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func (d *DoltStore) CreatePRISMDocument(ctx context.Context, doc *store.PRISMDocument) error {
	b := d.client.PRISMDocument.Create().
		SetID(doc.ID).
		SetName(doc.Name).
		SetCreatedAt(doc.CreatedAt).
		SetUpdatedAt(doc.UpdatedAt)
	if doc.Organization != "" {
		b.SetOrganization(doc.Organization)
	}
	if doc.RepositoryID != "" {
		b.SetRepositoryID(doc.RepositoryID)
	}
	if doc.Description != "" {
		b.SetDescription(doc.Description)
	}
	if doc.Version != "" {
		b.SetVersion(doc.Version)
	}
	if doc.Domains != nil {
		b.SetDomains(doc.Domains)
	}
	if doc.Layers != nil {
		b.SetLayers(doc.Layers)
	}
	if doc.Metrics != nil {
		b.SetMetrics(doc.Metrics)
	}
	if doc.Maturity != nil {
		b.SetMaturity(doc.Maturity)
	}
	if doc.SLIState != nil {
		b.SetSliState(doc.SLIState)
	}
	if doc.MaturityState != nil {
		b.SetMaturityState(doc.MaturityState)
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("create prism document: %w", err)
	}
	return nil
}

func (d *DoltStore) GetPRISMDocument(ctx context.Context, id string) (*store.PRISMDocument, error) {
	e, err := d.client.PRISMDocument.Query().
		Where(prismdocument.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get prism document %s: %w", id, err)
	}
	return entPRISMDocumentToStore(e), nil
}

func (d *DoltStore) ListPRISMDocuments(ctx context.Context) ([]*store.PRISMDocument, error) {
	rows, err := d.client.PRISMDocument.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list prism documents: %w", err)
	}
	result := make([]*store.PRISMDocument, len(rows))
	for i, e := range rows {
		result[i] = entPRISMDocumentToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListPRISMDocumentsByOrg(ctx context.Context, org string) ([]*store.PRISMDocument, error) {
	rows, err := d.client.PRISMDocument.Query().
		Where(prismdocument.OrganizationEQ(org)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list prism documents by org: %w", err)
	}
	result := make([]*store.PRISMDocument, len(rows))
	for i, e := range rows {
		result[i] = entPRISMDocumentToStore(e)
	}
	return result, nil
}

func (d *DoltStore) ListPRISMDocumentsByRepo(ctx context.Context, repoID string) ([]*store.PRISMDocument, error) {
	rows, err := d.client.PRISMDocument.Query().
		Where(prismdocument.RepositoryIDEQ(repoID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list prism documents by repo: %w", err)
	}
	result := make([]*store.PRISMDocument, len(rows))
	for i, e := range rows {
		result[i] = entPRISMDocumentToStore(e)
	}
	return result, nil
}

func (d *DoltStore) UpdatePRISMDocument(ctx context.Context, doc *store.PRISMDocument) error {
	b := d.client.PRISMDocument.UpdateOneID(doc.ID).
		SetName(doc.Name).
		SetUpdatedAt(doc.UpdatedAt)
	if doc.Organization != "" {
		b.SetOrganization(doc.Organization)
	} else {
		b.ClearOrganization()
	}
	if doc.RepositoryID != "" {
		b.SetRepositoryID(doc.RepositoryID)
	} else {
		b.ClearRepositoryID()
	}
	if doc.Description != "" {
		b.SetDescription(doc.Description)
	} else {
		b.ClearDescription()
	}
	if doc.Version != "" {
		b.SetVersion(doc.Version)
	} else {
		b.ClearVersion()
	}
	if doc.Domains != nil {
		b.SetDomains(doc.Domains)
	} else {
		b.ClearDomains()
	}
	if doc.Layers != nil {
		b.SetLayers(doc.Layers)
	} else {
		b.ClearLayers()
	}
	if doc.Metrics != nil {
		b.SetMetrics(doc.Metrics)
	} else {
		b.ClearMetrics()
	}
	if doc.Maturity != nil {
		b.SetMaturity(doc.Maturity)
	} else {
		b.ClearMaturity()
	}
	if doc.SLIState != nil {
		b.SetSliState(doc.SLIState)
	} else {
		b.ClearSliState()
	}
	if doc.MaturityState != nil {
		b.SetMaturityState(doc.MaturityState)
	} else {
		b.ClearMaturityState()
	}
	_, err := b.Save(ctx)
	if err != nil {
		return fmt.Errorf("update prism document %s: %w", doc.ID, err)
	}
	return nil
}
