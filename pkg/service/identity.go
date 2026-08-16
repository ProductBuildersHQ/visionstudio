package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// OrgID composes the canonical organization entity ID from a GitHub login.
func OrgID(login string) string {
	return fmt.Sprintf("github.com/%s", login)
}

// EnsureOrganization returns the organization row for a login, creating it
// if absent. kind is only applied on creation ("" defaults to
// "organization"); existing rows are never reclassified here.
func (s *Service) EnsureOrganization(ctx context.Context, login, kind string) (*store.Organization, bool, error) {
	id := OrgID(login)
	if existing, err := s.Store.GetOrganization(ctx, id); err == nil {
		return existing, false, nil
	}
	if kind == "" {
		kind = "organization"
	}
	now := time.Now().UTC()
	org := &store.Organization{
		ID:        id,
		Login:     login,
		Kind:      kind,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.Store.CreateOrganization(ctx, org); err != nil {
		return nil, false, err
	}
	return org, true, nil
}

// BackfillResult summarizes an organization backfill run.
type BackfillResult struct {
	OrgsCreated  []string
	ReposLinked  int
	ReposSkipped int
}

// BackfillOrganizations creates Organization rows for every distinct
// Repository.Organization string and links each repository to its org via
// OrganizationID. userLogins marks logins that are GitHub user accounts
// (kind "user") rather than organizations. Idempotent.
func (s *Service) BackfillOrganizations(ctx context.Context, userLogins map[string]bool) (*BackfillResult, error) {
	repos, err := s.Store.ListRepositories(ctx)
	if err != nil {
		return nil, err
	}
	res := &BackfillResult{}
	for _, repo := range repos {
		if repo.Organization == "" {
			res.ReposSkipped++
			continue
		}
		kind := "organization"
		if userLogins[repo.Organization] {
			kind = "user"
		}
		org, created, err := s.EnsureOrganization(ctx, repo.Organization, kind)
		if err != nil {
			return res, fmt.Errorf("ensure organization %s: %w", repo.Organization, err)
		}
		if created {
			res.OrgsCreated = append(res.OrgsCreated, org.ID)
		}
		if repo.OrganizationID == org.ID {
			res.ReposSkipped++
			continue
		}
		repo.OrganizationID = org.ID
		if err := s.Store.UpdateRepository(ctx, repo); err != nil {
			return res, fmt.Errorf("link repository %s: %w", repo.ID, err)
		}
		res.ReposLinked++
	}
	return res, nil
}

// ListOrganizations returns all organization rows.
func (s *Service) ListOrganizations(ctx context.Context) ([]*store.Organization, error) {
	return s.Store.ListOrganizations(ctx)
}

// RegisterPerson upserts an identity. orgLogins are resolved (and created
// with default kind if needed) to organization IDs for affiliation.
func (s *Service) RegisterPerson(ctx context.Context, id, githubLogin, displayName string, emails, orgLogins []string) (*store.Person, error) {
	if id == "" {
		id = fmt.Sprintf("person:%s", githubLogin)
	}
	if githubLogin == "" {
		return nil, fmt.Errorf("github login is required")
	}
	var orgIDs []string
	for _, login := range orgLogins {
		org, _, err := s.EnsureOrganization(ctx, login, "")
		if err != nil {
			return nil, err
		}
		orgIDs = append(orgIDs, org.ID)
	}
	now := time.Now().UTC()
	if existing, err := s.Store.GetPerson(ctx, id); err == nil {
		existing.GitHubLogin = githubLogin
		if displayName != "" {
			existing.DisplayName = displayName
		}
		if len(emails) > 0 {
			existing.EmailIdentities = emails
		}
		if len(orgIDs) > 0 {
			existing.OrgIDs = orgIDs
		}
		existing.UpdatedAt = now
		if err := s.Store.UpdatePerson(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	p := &store.Person{
		ID:              id,
		GitHubLogin:     githubLogin,
		DisplayName:     displayName,
		EmailIdentities: emails,
		OrgIDs:          orgIDs,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.Store.CreatePerson(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ListPeople returns all identities.
func (s *Service) ListPeople(ctx context.Context) ([]*store.Person, error) {
	return s.Store.ListPeople(ctx)
}
