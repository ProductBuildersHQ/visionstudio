package store

import (
	"context"
	"testing"
)

func TestMemStoreOrganizationCRUD(t *testing.T) {
	s := NewMemStore()
	ctx := context.Background()

	org := &Organization{ID: "github.com/plexusone", Login: "plexusone", Kind: "organization"}
	if err := s.CreateOrganization(ctx, org); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreateOrganization(ctx, org); err == nil {
		t.Fatal("expected duplicate error")
	}

	got, err := s.GetOrganization(ctx, "github.com/plexusone")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Login != "plexusone" {
		t.Fatalf("login = %q, want plexusone", got.Login)
	}

	byLogin, err := s.GetOrganizationByLogin(ctx, "plexusone")
	if err != nil {
		t.Fatalf("get by login: %v", err)
	}
	if byLogin.ID != org.ID {
		t.Fatalf("by-login ID = %q, want %q", byLogin.ID, org.ID)
	}
	if _, err := s.GetOrganizationByLogin(ctx, "nope"); err == nil {
		t.Fatal("expected not-found for unknown login")
	}

	org.Kind = "user"
	if err := s.UpdateOrganization(ctx, org); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.UpdateOrganization(ctx, &Organization{ID: "github.com/missing"}); err == nil {
		t.Fatal("expected not-found on update")
	}

	if err := s.CreateOrganization(ctx, &Organization{ID: "github.com/grokify", Login: "grokify", Kind: "user"}); err != nil {
		t.Fatalf("create second: %v", err)
	}
	orgs, err := s.ListOrganizations(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(orgs) != 2 {
		t.Fatalf("list len = %d, want 2", len(orgs))
	}
	if orgs[0].ID != "github.com/grokify" {
		t.Fatalf("list not sorted by ID: first = %s", orgs[0].ID)
	}
}

func TestMemStorePersonCRUD(t *testing.T) {
	s := NewMemStore()
	ctx := context.Background()

	p := &Person{
		ID:              "person:grokify",
		GitHubLogin:     "grokify",
		EmailIdentities: []string{"johncwang@gmail.com"},
		OrgIDs:          []string{"github.com/grokify", "github.com/plexusone"},
	}
	if err := s.CreatePerson(ctx, p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.CreatePerson(ctx, p); err == nil {
		t.Fatal("expected duplicate error")
	}

	got, err := s.GetPerson(ctx, "person:grokify")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.OrgIDs) != 2 {
		t.Fatalf("org IDs len = %d, want 2", len(got.OrgIDs))
	}

	p.DisplayName = "John Wang"
	if err := s.UpdatePerson(ctx, p); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.UpdatePerson(ctx, &Person{ID: "person:missing"}); err == nil {
		t.Fatal("expected not-found on update")
	}

	people, err := s.ListPeople(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(people) != 1 || people[0].DisplayName != "John Wang" {
		t.Fatalf("list = %+v, want single updated person", people)
	}
}
