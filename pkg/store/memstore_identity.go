package store

import (
	"context"
	"fmt"
	"sort"
)

// --- OrganizationStore ---

func (m *MemStore) CreateOrganization(_ context.Context, org *Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.organizations[org.ID]; exists {
		return fmt.Errorf("organization %s already exists", org.ID)
	}
	m.organizations[org.ID] = org
	return nil
}

func (m *MemStore) GetOrganization(_ context.Context, id string) (*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	org, ok := m.organizations[id]
	if !ok {
		return nil, fmt.Errorf("organization %s not found", id)
	}
	return org, nil
}

func (m *MemStore) GetOrganizationByLogin(_ context.Context, login string) (*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, org := range m.organizations {
		if org.Login == login {
			return org, nil
		}
	}
	return nil, fmt.Errorf("organization with login %s not found", login)
}

func (m *MemStore) ListOrganizations(_ context.Context) ([]*Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Organization, 0, len(m.organizations))
	for _, org := range m.organizations {
		result = append(result, org)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (m *MemStore) UpdateOrganization(_ context.Context, org *Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.organizations[org.ID]; !ok {
		return fmt.Errorf("organization %s not found", org.ID)
	}
	m.organizations[org.ID] = org
	return nil
}

// --- PersonStore ---

func (m *MemStore) CreatePerson(_ context.Context, p *Person) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.people[p.ID]; exists {
		return fmt.Errorf("person %s already exists", p.ID)
	}
	m.people[p.ID] = p
	return nil
}

func (m *MemStore) GetPerson(_ context.Context, id string) (*Person, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.people[id]
	if !ok {
		return nil, fmt.Errorf("person %s not found", id)
	}
	return p, nil
}

func (m *MemStore) ListPeople(_ context.Context) ([]*Person, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Person, 0, len(m.people))
	for _, p := range m.people {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (m *MemStore) UpdatePerson(_ context.Context, p *Person) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.people[p.ID]; !ok {
		return fmt.Errorf("person %s not found", p.ID)
	}
	m.people[p.ID] = p
	return nil
}
