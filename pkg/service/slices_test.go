package service

import (
	"context"
	"testing"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

func seedSliceFixtures(t *testing.T, s *store.MemStore) {
	t.Helper()
	ctx := context.Background()

	orgs := []*store.Organization{
		{ID: "github.com/plexusone", Login: "plexusone", Kind: "organization"},
		{ID: "github.com/grokify", Login: "grokify", Kind: "user"},
	}
	for _, o := range orgs {
		if err := s.CreateOrganization(ctx, o); err != nil {
			t.Fatal(err)
		}
	}

	repos := []*store.Repository{
		{ID: "github.com/plexusone/omnidevx", Organization: "plexusone", RepositoryName: "omnidevx", Status: "active", OrganizationID: "github.com/plexusone", Visibility: "public"},
		{ID: "github.com/plexusone/secret", Organization: "plexusone", RepositoryName: "secret", Status: "active", OrganizationID: "github.com/plexusone", Visibility: "private"},
		{ID: "github.com/grokify/gogit", Organization: "grokify", RepositoryName: "gogit", Status: "active", OrganizationID: "github.com/grokify", Visibility: "unknown"},
	}
	for _, r := range repos {
		if err := s.CreateRepository(ctx, r); err != nil {
			t.Fatal(err)
		}
	}

	inits := []*store.Initiative{
		{ID: "INIT-A", Title: "A", Status: "executing", HomeRepo: "github.com/plexusone/secret", Organization: "default"},
		{ID: "INIT-B", Title: "B", Status: "closed", HomeRepo: "github.com/plexusone/omnidevx", Organization: "default"},
	}
	for _, in := range inits {
		if err := s.CreateInitiative(ctx, in); err != nil {
			t.Fatal(err)
		}
	}

	rmis := []*store.RoadmapItem{
		{ID: "RMI-X-001", RepositoryID: "github.com/plexusone/omnidevx", Title: "x", ItemType: "capability", Status: "proposed"},
		{ID: "RMI-X-002", RepositoryID: "github.com/plexusone/omnidevx", Title: "y", ItemType: "capability", Status: "completed"},
	}
	for _, r := range rmis {
		if err := s.CreateRMI(ctx, r); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := New(s).RegisterPerson(ctx, "", "grokify", "", nil, []string{"grokify", "plexusone"}); err != nil {
		t.Fatal(err)
	}
}

func TestOrgRollup(t *testing.T) {
	s := store.NewMemStore()
	seedSliceFixtures(t, s)
	svc := New(s)

	rollup, err := svc.OrgRollup(context.Background())
	if err != nil {
		t.Fatalf("rollup: %v", err)
	}
	byLogin := map[string]*OrgStats{}
	for _, st := range rollup {
		byLogin[st.Org.Login] = st
	}

	px := byLogin["plexusone"]
	if px.Repos != 2 || px.Public != 1 || px.Private != 1 {
		t.Fatalf("plexusone repos = %+v", px)
	}
	if px.Initiatives != 2 || px.ActiveInits != 1 {
		t.Fatalf("plexusone inits = %d active %d, want 2/1", px.Initiatives, px.ActiveInits)
	}
	if px.OpenRMIs != 1 {
		t.Fatalf("plexusone open RMIs = %d, want 1 (completed excluded)", px.OpenRMIs)
	}
	if len(px.PeopleLogins) != 1 || px.PeopleLogins[0] != "grokify" {
		t.Fatalf("plexusone people = %v", px.PeopleLogins)
	}

	gr := byLogin["grokify"]
	if gr.Repos != 1 || gr.Unknown != 1 {
		t.Fatalf("grokify = %+v", gr)
	}
}

func TestFocusList(t *testing.T) {
	s := store.NewMemStore()
	seedSliceFixtures(t, s)
	svc := New(s)

	entries, err := svc.FocusList(context.Background())
	if err != nil {
		t.Fatalf("focus: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1 (only confirmed-private; unknown excluded)", len(entries))
	}
	e := entries[0]
	if e.Repo.ID != "github.com/plexusone/secret" {
		t.Fatalf("repo = %s", e.Repo.ID)
	}
	if len(e.Initiatives) != 1 || e.Initiatives[0].ID != "INIT-A" {
		t.Fatalf("initiatives = %+v, want active INIT-A only", e.Initiatives)
	}
}

func TestOrgRollupIgnoresUnlinkedRepos(t *testing.T) {
	s := store.NewMemStore()
	seedSliceFixtures(t, s)
	svc := New(s)
	ctx := context.Background()

	// A pre-backfill row: organization string set, no OrganizationID edge.
	if err := s.CreateRepository(ctx, &store.Repository{
		ID: "github.com/plexusone/legacy", Organization: "plexusone",
		RepositoryName: "legacy", Status: "active",
	}); err != nil {
		t.Fatal(err)
	}

	rollup, err := svc.OrgRollup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range rollup {
		if st.Org.Login == "plexusone" && st.Repos != 2 {
			t.Fatalf("plexusone repos = %d, want 2 — unlinked repo must not count (run org backfill to link it)", st.Repos)
		}
	}
}

func TestPersonRepositories(t *testing.T) {
	s := store.NewMemStore()
	seedSliceFixtures(t, s)
	svc := New(s)
	ctx := context.Background()

	p, repos, err := svc.PersonRepositories(ctx, "grokify") // bare login accepted
	if err != nil {
		t.Fatalf("person repos: %v", err)
	}
	if p.ID != "person:grokify" {
		t.Fatalf("person = %s", p.ID)
	}
	if len(repos) != 3 {
		t.Fatalf("repos = %d, want 3 (both orgs)", len(repos))
	}

	if _, _, err := svc.PersonRepositories(ctx, "person:missing"); err == nil {
		t.Fatal("expected not-found")
	}
}
