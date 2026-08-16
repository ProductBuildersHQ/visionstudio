package doltstore

import (
	"context"
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/ent"
	"github.com/ProductBuildersHQ/visionstudio/ent/organization"
	"github.com/ProductBuildersHQ/visionstudio/ent/person"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// ---------------------------------------------------------------------------
// Organization CRUD
// ---------------------------------------------------------------------------

func entOrgToStore(o *ent.Organization) *store.Organization {
	return &store.Organization{
		ID:             o.ID,
		Login:          o.Login,
		Kind:           o.Kind,
		DisplayName:    o.DisplayName,
		Website:        o.Website,
		ReleasePageURL: o.ReleasePageURL,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

func (d *DoltStore) CreateOrganization(ctx context.Context, org *store.Organization) error {
	b := d.client.Organization.Create().
		SetID(org.ID).
		SetLogin(org.Login).
		SetKind(org.Kind).
		SetCreatedAt(org.CreatedAt).
		SetUpdatedAt(org.UpdatedAt)
	if org.DisplayName != "" {
		b.SetDisplayName(org.DisplayName)
	}
	if org.Website != "" {
		b.SetWebsite(org.Website)
	}
	if org.ReleasePageURL != "" {
		b.SetReleasePageURL(org.ReleasePageURL)
	}
	if _, err := b.Save(ctx); err != nil {
		return fmt.Errorf("create organization %s: %w", org.ID, err)
	}
	return nil
}

func (d *DoltStore) GetOrganization(ctx context.Context, id string) (*store.Organization, error) {
	o, err := d.client.Organization.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get organization %s: %w", id, err)
	}
	return entOrgToStore(o), nil
}

func (d *DoltStore) GetOrganizationByLogin(ctx context.Context, login string) (*store.Organization, error) {
	o, err := d.client.Organization.Query().
		Where(organization.LoginEQ(login)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get organization by login %s: %w", login, err)
	}
	return entOrgToStore(o), nil
}

func (d *DoltStore) ListOrganizations(ctx context.Context) ([]*store.Organization, error) {
	rows, err := d.client.Organization.Query().
		Order(ent.Asc(organization.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	result := make([]*store.Organization, len(rows))
	for i, o := range rows {
		result[i] = entOrgToStore(o)
	}
	return result, nil
}

func (d *DoltStore) UpdateOrganization(ctx context.Context, org *store.Organization) error {
	b := d.client.Organization.UpdateOneID(org.ID).
		SetLogin(org.Login).
		SetKind(org.Kind).
		SetUpdatedAt(org.UpdatedAt)
	if org.DisplayName != "" {
		b.SetDisplayName(org.DisplayName)
	} else {
		b.ClearDisplayName()
	}
	if org.Website != "" {
		b.SetWebsite(org.Website)
	} else {
		b.ClearWebsite()
	}
	if org.ReleasePageURL != "" {
		b.SetReleasePageURL(org.ReleasePageURL)
	} else {
		b.ClearReleasePageURL()
	}
	if _, err := b.Save(ctx); err != nil {
		return fmt.Errorf("update organization %s: %w", org.ID, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Person CRUD
// ---------------------------------------------------------------------------

func entPersonToStore(p *ent.Person) *store.Person {
	sp := &store.Person{
		ID:              p.ID,
		GitHubLogin:     p.GithubLogin,
		DisplayName:     p.DisplayName,
		EmailIdentities: p.EmailIdentities,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
	for _, o := range p.Edges.Organizations {
		sp.OrgIDs = append(sp.OrgIDs, o.ID)
	}
	return sp
}

func (d *DoltStore) CreatePerson(ctx context.Context, p *store.Person) error {
	b := d.client.Person.Create().
		SetID(p.ID).
		SetGithubLogin(p.GitHubLogin).
		SetCreatedAt(p.CreatedAt).
		SetUpdatedAt(p.UpdatedAt)
	if p.DisplayName != "" {
		b.SetDisplayName(p.DisplayName)
	}
	if len(p.EmailIdentities) > 0 {
		b.SetEmailIdentities(p.EmailIdentities)
	}
	if len(p.OrgIDs) > 0 {
		b.AddOrganizationIDs(p.OrgIDs...)
	}
	if _, err := b.Save(ctx); err != nil {
		return fmt.Errorf("create person %s: %w", p.ID, err)
	}
	return nil
}

func (d *DoltStore) GetPerson(ctx context.Context, id string) (*store.Person, error) {
	p, err := d.client.Person.Query().
		Where(person.IDEQ(id)).
		WithOrganizations().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get person %s: %w", id, err)
	}
	return entPersonToStore(p), nil
}

func (d *DoltStore) ListPeople(ctx context.Context) ([]*store.Person, error) {
	rows, err := d.client.Person.Query().
		WithOrganizations().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list people: %w", err)
	}
	result := make([]*store.Person, len(rows))
	for i, p := range rows {
		result[i] = entPersonToStore(p)
	}
	return result, nil
}

func (d *DoltStore) UpdatePerson(ctx context.Context, p *store.Person) error {
	b := d.client.Person.UpdateOneID(p.ID).
		SetGithubLogin(p.GitHubLogin).
		SetUpdatedAt(p.UpdatedAt).
		ClearOrganizations()
	if p.DisplayName != "" {
		b.SetDisplayName(p.DisplayName)
	} else {
		b.ClearDisplayName()
	}
	if len(p.EmailIdentities) > 0 {
		b.SetEmailIdentities(p.EmailIdentities)
	} else {
		b.ClearEmailIdentities()
	}
	if len(p.OrgIDs) > 0 {
		b.AddOrganizationIDs(p.OrgIDs...)
	}
	if _, err := b.Save(ctx); err != nil {
		return fmt.Errorf("update person %s: %w", p.ID, err)
	}
	return nil
}
