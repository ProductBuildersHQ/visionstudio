package service

import (
	"context"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// CreateDevXPeriodReport persists a developer experience period report.
func (s *Service) CreateDevXPeriodReport(ctx context.Context, report *store.DevXPeriodReport) error {
	return s.Store.CreateDevXPeriodReport(ctx, report)
}

// GetDevXPeriodReport retrieves a DevX period report by ID.
func (s *Service) GetDevXPeriodReport(ctx context.Context, id string) (*store.DevXPeriodReport, error) {
	return s.Store.GetDevXPeriodReport(ctx, id)
}

// ListDevXPeriodReports returns reports for a person.
func (s *Service) ListDevXPeriodReports(ctx context.Context, personID string) ([]*store.DevXPeriodReport, error) {
	return s.Store.ListDevXPeriodReports(ctx, personID)
}

// ListDevXPeriodReportsByRepo returns reports for a repository.
func (s *Service) ListDevXPeriodReportsByRepo(ctx context.Context, repoID string) ([]*store.DevXPeriodReport, error) {
	return s.Store.ListDevXPeriodReportsByRepo(ctx, repoID)
}

// ListDevXPeriodReportsByOrg returns reports for an organization.
func (s *Service) ListDevXPeriodReportsByOrg(ctx context.Context, org string) ([]*store.DevXPeriodReport, error) {
	return s.Store.ListDevXPeriodReportsByOrg(ctx, org)
}

// CreatePRISMRoadmap persists a prism-roadmap artifact.
func (s *Service) CreatePRISMRoadmap(ctx context.Context, roadmap *store.PRISMRoadmap) error {
	return s.Store.CreatePRISMRoadmap(ctx, roadmap)
}

// GetPRISMRoadmap retrieves a PRISM roadmap by ID.
func (s *Service) GetPRISMRoadmap(ctx context.Context, id string) (*store.PRISMRoadmap, error) {
	return s.Store.GetPRISMRoadmap(ctx, id)
}

// ListPRISMRoadmaps returns all PRISM roadmaps.
func (s *Service) ListPRISMRoadmaps(ctx context.Context) ([]*store.PRISMRoadmap, error) {
	return s.Store.ListPRISMRoadmaps(ctx)
}

// ListPRISMRoadmapsByRepo returns PRISM roadmaps for a repository.
func (s *Service) ListPRISMRoadmapsByRepo(ctx context.Context, repoID string) ([]*store.PRISMRoadmap, error) {
	return s.Store.ListPRISMRoadmapsByRepo(ctx, repoID)
}

// UpdatePRISMRoadmap updates a PRISM roadmap.
func (s *Service) UpdatePRISMRoadmap(ctx context.Context, roadmap *store.PRISMRoadmap) error {
	return s.Store.UpdatePRISMRoadmap(ctx, roadmap)
}

// CreatePRISMGoal persists a prism-roadmap goals artifact.
func (s *Service) CreatePRISMGoal(ctx context.Context, goal *store.PRISMGoal) error {
	return s.Store.CreatePRISMGoal(ctx, goal)
}

// GetPRISMGoal retrieves a PRISM goal by ID.
func (s *Service) GetPRISMGoal(ctx context.Context, id string) (*store.PRISMGoal, error) {
	return s.Store.GetPRISMGoal(ctx, id)
}

// ListPRISMGoals returns PRISM goals for a repository.
func (s *Service) ListPRISMGoals(ctx context.Context, repoID string) ([]*store.PRISMGoal, error) {
	return s.Store.ListPRISMGoals(ctx, repoID)
}

// UpdatePRISMGoal updates a PRISM goal.
func (s *Service) UpdatePRISMGoal(ctx context.Context, goal *store.PRISMGoal) error {
	return s.Store.UpdatePRISMGoal(ctx, goal)
}

// CreatePRISMDocument persists a prism-maturity document.
func (s *Service) CreatePRISMDocument(ctx context.Context, doc *store.PRISMDocument) error {
	return s.Store.CreatePRISMDocument(ctx, doc)
}

// GetPRISMDocument retrieves a PRISM document by ID.
func (s *Service) GetPRISMDocument(ctx context.Context, id string) (*store.PRISMDocument, error) {
	return s.Store.GetPRISMDocument(ctx, id)
}

// ListPRISMDocuments returns all PRISM documents.
func (s *Service) ListPRISMDocuments(ctx context.Context) ([]*store.PRISMDocument, error) {
	return s.Store.ListPRISMDocuments(ctx)
}

// ListPRISMDocumentsByOrg returns PRISM documents for an organization.
func (s *Service) ListPRISMDocumentsByOrg(ctx context.Context, org string) ([]*store.PRISMDocument, error) {
	return s.Store.ListPRISMDocumentsByOrg(ctx, org)
}

// ListPRISMDocumentsByRepo returns PRISM documents for a repository.
func (s *Service) ListPRISMDocumentsByRepo(ctx context.Context, repoID string) ([]*store.PRISMDocument, error) {
	return s.Store.ListPRISMDocumentsByRepo(ctx, repoID)
}

// UpdatePRISMDocument updates a PRISM document.
func (s *Service) UpdatePRISMDocument(ctx context.Context, doc *store.PRISMDocument) error {
	return s.Store.UpdatePRISMDocument(ctx, doc)
}

// CreateSpecDocument persists a spec document registry entry.
func (s *Service) CreateSpecDocument(ctx context.Context, doc *store.SpecDocument) error {
	return s.Store.CreateSpecDocument(ctx, doc)
}

// GetSpecDocument retrieves a spec document by ID.
func (s *Service) GetSpecDocument(ctx context.Context, id string) (*store.SpecDocument, error) {
	return s.Store.GetSpecDocument(ctx, id)
}

// ListSpecDocuments returns all spec documents.
func (s *Service) ListSpecDocuments(ctx context.Context) ([]*store.SpecDocument, error) {
	return s.Store.ListSpecDocuments(ctx)
}

// ListSpecDocumentsByRepo returns spec documents for a repository.
func (s *Service) ListSpecDocumentsByRepo(ctx context.Context, repoID string) ([]*store.SpecDocument, error) {
	return s.Store.ListSpecDocumentsByRepo(ctx, repoID)
}

// ListSpecDocumentsByInitiative returns spec documents for an initiative.
func (s *Service) ListSpecDocumentsByInitiative(ctx context.Context, initiativeID string) ([]*store.SpecDocument, error) {
	return s.Store.ListSpecDocumentsByInitiative(ctx, initiativeID)
}

// UpdateSpecDocument updates a spec document.
func (s *Service) UpdateSpecDocument(ctx context.Context, doc *store.SpecDocument) error {
	return s.Store.UpdateSpecDocument(ctx, doc)
}

// DeleteSpecDocument removes a spec document.
func (s *Service) DeleteSpecDocument(ctx context.Context, id string) error {
	return s.Store.DeleteSpecDocument(ctx, id)
}
