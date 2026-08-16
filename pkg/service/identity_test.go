package service

import (
	"context"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func TestRegisterRepositoryEnsuresOrganization(t *testing.T) {
	svc := New(store.NewMemStore())
	ctx := context.Background()

	repo, err := svc.RegisterRepository(ctx, "plexusone", "omnidevx", "", "", "")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if repo.OrganizationID != "github.com/plexusone" {
		t.Fatalf("OrganizationID = %q, want github.com/plexusone", repo.OrganizationID)
	}
	org, err := svc.Store.GetOrganization(ctx, "github.com/plexusone")
	if err != nil {
		t.Fatalf("org row not created: %v", err)
	}
	if org.Kind != "organization" {
		t.Fatalf("kind = %q, want organization", org.Kind)
	}

	// Second repo in the same org reuses the row.
	if _, err := svc.RegisterRepository(ctx, "plexusone", "devfolio", "", "", ""); err != nil {
		t.Fatalf("register second: %v", err)
	}
	orgs, err := svc.ListOrganizations(ctx)
	if err != nil {
		t.Fatalf("list orgs: %v", err)
	}
	if len(orgs) != 1 {
		t.Fatalf("orgs = %d, want 1", len(orgs))
	}
}

func TestBackfillOrganizations(t *testing.T) {
	s := store.NewMemStore()
	svc := New(s)
	ctx := context.Background()

	// Simulate pre-entity rows: org string set, no OrganizationID.
	seed := []*store.Repository{
		{ID: "github.com/plexusone/omnidevx", Organization: "plexusone", RepositoryName: "omnidevx", Status: "active"},
		{ID: "github.com/grokify/gogit", Organization: "grokify", RepositoryName: "gogit", Status: "active"},
		{ID: "github.com/ProductBuildersHQ/acts", Organization: "ProductBuildersHQ", RepositoryName: "acts", Status: "active"},
	}
	for _, r := range seed {
		if err := s.CreateRepository(ctx, r); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	res, err := svc.BackfillOrganizations(ctx, map[string]bool{"grokify": true})
	if err != nil {
		t.Fatalf("backfill: %v", err)
	}
	if len(res.OrgsCreated) != 3 {
		t.Fatalf("orgs created = %d, want 3", len(res.OrgsCreated))
	}
	if res.ReposLinked != 3 {
		t.Fatalf("repos linked = %d, want 3", res.ReposLinked)
	}

	grok, err := s.GetOrganization(ctx, "github.com/grokify")
	if err != nil {
		t.Fatalf("get grokify org: %v", err)
	}
	if grok.Kind != "user" {
		t.Fatalf("grokify kind = %q, want user (user-account-as-org)", grok.Kind)
	}

	repo, err := s.GetRepository(ctx, "github.com/plexusone/omnidevx")
	if err != nil {
		t.Fatalf("get repo: %v", err)
	}
	if repo.OrganizationID != "github.com/plexusone" {
		t.Fatalf("repo OrganizationID = %q", repo.OrganizationID)
	}

	// Idempotent: second run creates and links nothing.
	res2, err := svc.BackfillOrganizations(ctx, nil)
	if err != nil {
		t.Fatalf("backfill 2: %v", err)
	}
	if len(res2.OrgsCreated) != 0 || res2.ReposLinked != 0 {
		t.Fatalf("second run not idempotent: %+v", res2)
	}
}

func TestRegisterPerson(t *testing.T) {
	svc := New(store.NewMemStore())
	ctx := context.Background()

	p, err := svc.RegisterPerson(ctx, "", "grokify", "John Wang",
		[]string{"johncwang@gmail.com"}, []string{"grokify", "plexusone", "ProductBuildersHQ"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if p.ID != "person:grokify" {
		t.Fatalf("ID = %q, want person:grokify", p.ID)
	}
	if len(p.OrgIDs) != 3 {
		t.Fatalf("org IDs = %d, want 3", len(p.OrgIDs))
	}

	// Upsert path: update display name, keep orgs.
	p2, err := svc.RegisterPerson(ctx, "", "grokify", "John", nil, nil)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if p2.DisplayName != "John" {
		t.Fatalf("display = %q", p2.DisplayName)
	}
	if len(p2.OrgIDs) != 3 {
		t.Fatalf("orgs lost on upsert: %d", len(p2.OrgIDs))
	}

	if _, err := svc.RegisterPerson(ctx, "", "", "", nil, nil); err == nil {
		t.Fatal("expected error for missing github login")
	}
}
