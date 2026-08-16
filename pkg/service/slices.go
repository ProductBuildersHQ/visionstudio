package service

import (
	"context"
	"sort"
	"strings"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// OrgStats is a per-organization rollup over the registry and portfolio.
type OrgStats struct {
	Org          *store.Organization
	Repos        int
	Public       int
	Private      int
	Unknown      int
	Initiatives  int // initiatives whose home repo belongs to the org
	ActiveInits  int // of those, not closed/cancelled
	OpenRMIs     int // RMIs on the org's repos not completed/cancelled
	PeopleLogins []string
}

// FocusEntry pairs a private repository with its active initiatives —
// the "private repos we're focusing on" view.
type FocusEntry struct {
	Repo        *store.Repository
	Initiatives []*store.Initiative
}

func initiativeActive(status string) bool {
	switch status {
	case "closed", "cancelled":
		return false
	}
	return true
}

func rmiOpen(status string) bool {
	switch status {
	case "completed", "cancelled":
		return false
	}
	return true
}

// OrgRollup aggregates repos, visibility, initiatives, RMIs, and people
// per organization via the organization edge (not string matching).
func (s *Service) OrgRollup(ctx context.Context) ([]*OrgStats, error) {
	orgs, err := s.Store.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	repos, err := s.Store.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}
	inits, err := s.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}
	rmis, err := s.Store.ListAllRMIs(ctx)
	if err != nil {
		return nil, err
	}
	people, err := s.Store.ListPeople(ctx)
	if err != nil {
		return nil, err
	}

	byOrg := make(map[string]*OrgStats, len(orgs))
	var result []*OrgStats
	for _, o := range orgs {
		st := &OrgStats{Org: o}
		byOrg[o.ID] = st
		result = append(result, st)
	}

	repoOrg := make(map[string]string, len(repos)) // repo ID -> org entity ID
	for _, r := range repos {
		st, ok := byOrg[r.OrganizationID]
		if !ok {
			continue
		}
		repoOrg[r.ID] = r.OrganizationID
		st.Repos++
		switch r.Visibility {
		case "public":
			st.Public++
		case "private":
			st.Private++
		default:
			st.Unknown++
		}
	}

	for _, in := range inits {
		orgID, ok := repoOrg[in.HomeRepo]
		if !ok {
			continue
		}
		st := byOrg[orgID]
		st.Initiatives++
		if initiativeActive(in.Status) {
			st.ActiveInits++
		}
	}

	for _, rmi := range rmis {
		orgID, ok := repoOrg[rmi.RepositoryID]
		if !ok || !rmiOpen(rmi.Status) {
			continue
		}
		byOrg[orgID].OpenRMIs++
	}

	for _, p := range people {
		for _, orgID := range p.OrgIDs {
			if st, ok := byOrg[orgID]; ok {
				st.PeopleLogins = append(st.PeopleLogins, p.GitHubLogin)
			}
		}
	}
	return result, nil
}

// FocusList returns private repositories with their active initiatives.
// Repositories with unknown visibility are excluded — this is a view of
// confirmed-private work, not a guess.
func (s *Service) FocusList(ctx context.Context) ([]*FocusEntry, error) {
	repos, err := s.Store.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}
	inits, err := s.Store.ListInitiatives(ctx)
	if err != nil {
		return nil, err
	}

	initsByHome := make(map[string][]*store.Initiative)
	for _, in := range inits {
		if initiativeActive(in.Status) {
			initsByHome[in.HomeRepo] = append(initsByHome[in.HomeRepo], in)
		}
	}

	var result []*FocusEntry
	for _, r := range repos {
		if r.Visibility != "private" {
			continue
		}
		result = append(result, &FocusEntry{Repo: r, Initiatives: initsByHome[r.ID]})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Repo.ID < result[j].Repo.ID })
	return result, nil
}

// PersonRepositories resolves the practitioner lens: every repository in
// every organization the person is affiliated with. This is the query
// ACTS's portfolio-period rollup consumes (person -> orgs -> repos).
func (s *Service) PersonRepositories(ctx context.Context, personID string) (*store.Person, []*store.Repository, error) {
	if !strings.Contains(personID, ":") {
		personID = "person:" + personID
	}
	p, err := s.Store.GetPerson(ctx, personID)
	if err != nil {
		return nil, nil, err
	}
	member := make(map[string]bool, len(p.OrgIDs))
	for _, id := range p.OrgIDs {
		member[id] = true
	}
	repos, err := s.Store.ListRepositories(ctx)
	if err != nil {
		return nil, nil, err
	}
	var result []*store.Repository
	for _, r := range repos {
		if member[r.OrganizationID] {
			result = append(result, r)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return p, result, nil
}
