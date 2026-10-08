package remote

import (
	"context"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// The methods in this file have no VisionStudio Cloud tenant API endpoint:
// the cloud does not serve these areas to remote clients. Each returns a
// *NotSupportedError wrapping ErrNotSupported that names the area, so
// callers fail clearly instead of silently reading empty data or dropping
// writes.

// --- initiative dependencies ---

func (s *Store) CreateInitiativeDependency(_ context.Context, _ *store.InitiativeDependency) error {
	return notServed("CreateInitiativeDependency", "initiative dependencies")
}

func (s *Store) ListInitiativeDependencies(_ context.Context, _ string) ([]*store.InitiativeDependency, error) {
	return nil, notServed("ListInitiativeDependencies", "initiative dependencies")
}

func (s *Store) ListAllInitiativeDependencies(_ context.Context) ([]*store.InitiativeDependency, error) {
	return nil, notServed("ListAllInitiativeDependencies", "initiative dependencies")
}

// --- repository writes ---

func (s *Store) CreateRepository(_ context.Context, _ *store.Repository) error {
	return notServed("CreateRepository", "repository writes")
}

func (s *Store) UpdateRepository(_ context.Context, _ *store.Repository) error {
	return notServed("UpdateRepository", "repository writes")
}

func (s *Store) DeleteRepository(_ context.Context, _ string) error {
	return notServed("DeleteRepository", "repository writes")
}

// --- repository dependencies ---

func (s *Store) CreateRepoDependency(_ context.Context, _ *store.RepositoryDependency) error {
	return notServed("CreateRepoDependency", "repository dependencies")
}

func (s *Store) ListRepoDependencies(_ context.Context, _ string) ([]*store.RepositoryDependency, error) {
	return nil, notServed("ListRepoDependencies", "repository dependencies")
}

func (s *Store) ListAllRepoDependencies(_ context.Context) ([]*store.RepositoryDependency, error) {
	return nil, notServed("ListAllRepoDependencies", "repository dependencies")
}

// --- release writes ---

func (s *Store) CreateRelease(_ context.Context, _ *store.Release) error {
	return notServed("CreateRelease", "release writes")
}

func (s *Store) UpdateRelease(_ context.Context, _ *store.Release) error {
	return notServed("UpdateRelease", "release writes")
}

func (s *Store) DeleteRelease(_ context.Context, _ string) error {
	return notServed("DeleteRelease", "release writes")
}

// --- organizations ---

func (s *Store) CreateOrganization(_ context.Context, _ *store.Organization) error {
	return notServed("CreateOrganization", "organizations")
}

func (s *Store) GetOrganization(_ context.Context, _ string) (*store.Organization, error) {
	return nil, notServed("GetOrganization", "organizations")
}

func (s *Store) GetOrganizationByLogin(_ context.Context, _ string) (*store.Organization, error) {
	return nil, notServed("GetOrganizationByLogin", "organizations")
}

func (s *Store) ListOrganizations(_ context.Context) ([]*store.Organization, error) {
	return nil, notServed("ListOrganizations", "organizations")
}

func (s *Store) UpdateOrganization(_ context.Context, _ *store.Organization) error {
	return notServed("UpdateOrganization", "organizations")
}

// --- people ---

func (s *Store) CreatePerson(_ context.Context, _ *store.Person) error {
	return notServed("CreatePerson", "people")
}

func (s *Store) GetPerson(_ context.Context, _ string) (*store.Person, error) {
	return nil, notServed("GetPerson", "people")
}

func (s *Store) ListPeople(_ context.Context) ([]*store.Person, error) {
	return nil, notServed("ListPeople", "people")
}

func (s *Store) UpdatePerson(_ context.Context, _ *store.Person) error {
	return notServed("UpdatePerson", "people")
}

// --- spec workflow writes (workflows come from the embedded catalog) ---

func (s *Store) CreateSpecWorkflow(_ context.Context, _ *store.SpecWorkflow) error {
	return notServed("CreateSpecWorkflow", "spec workflow writes")
}

func (s *Store) UpdateSpecWorkflow(_ context.Context, _ *store.SpecWorkflow) error {
	return notServed("UpdateSpecWorkflow", "spec workflow writes")
}

func (s *Store) DeleteSpecWorkflow(_ context.Context, _ string) error {
	return notServed("DeleteSpecWorkflow", "spec workflow writes")
}

// --- capability models ---

func (s *Store) CreateCapabilityModel(_ context.Context, _ *store.CapabilityModel) error {
	return notServed("CreateCapabilityModel", "capability models")
}

func (s *Store) GetCapabilityModel(_ context.Context, _ string) (*store.CapabilityModel, error) {
	return nil, notServed("GetCapabilityModel", "capability models")
}

func (s *Store) ListCapabilityModels(_ context.Context) ([]*store.CapabilityModel, error) {
	return nil, notServed("ListCapabilityModels", "capability models")
}

func (s *Store) UpdateCapabilityModel(_ context.Context, _ *store.CapabilityModel) error {
	return notServed("UpdateCapabilityModel", "capability models")
}

// --- maturity assessments ---

func (s *Store) CreateMaturityAssessment(_ context.Context, _ *store.MaturityAssessment) error {
	return notServed("CreateMaturityAssessment", "maturity assessments")
}

func (s *Store) GetMaturityAssessment(_ context.Context, _ string) (*store.MaturityAssessment, error) {
	return nil, notServed("GetMaturityAssessment", "maturity assessments")
}

func (s *Store) ListMaturityAssessments(_ context.Context, _ string) ([]*store.MaturityAssessment, error) {
	return nil, notServed("ListMaturityAssessments", "maturity assessments")
}

func (s *Store) ListMaturityAssessmentsByOrg(_ context.Context, _ string) ([]*store.MaturityAssessment, error) {
	return nil, notServed("ListMaturityAssessmentsByOrg", "maturity assessments")
}

// --- DevX period reports ---

func (s *Store) CreateDevXPeriodReport(_ context.Context, _ *store.DevXPeriodReport) error {
	return notServed("CreateDevXPeriodReport", "DevX period reports")
}

func (s *Store) GetDevXPeriodReport(_ context.Context, _ string) (*store.DevXPeriodReport, error) {
	return nil, notServed("GetDevXPeriodReport", "DevX period reports")
}

func (s *Store) ListDevXPeriodReports(_ context.Context, _ string) ([]*store.DevXPeriodReport, error) {
	return nil, notServed("ListDevXPeriodReports", "DevX period reports")
}

func (s *Store) ListDevXPeriodReportsByRepo(_ context.Context, _ string) ([]*store.DevXPeriodReport, error) {
	return nil, notServed("ListDevXPeriodReportsByRepo", "DevX period reports")
}

func (s *Store) ListDevXPeriodReportsByOrg(_ context.Context, _ string) ([]*store.DevXPeriodReport, error) {
	return nil, notServed("ListDevXPeriodReportsByOrg", "DevX period reports")
}

// --- PRISM roadmaps ---

func (s *Store) CreatePRISMRoadmap(_ context.Context, _ *store.PRISMRoadmap) error {
	return notServed("CreatePRISMRoadmap", "PRISM roadmaps")
}

func (s *Store) GetPRISMRoadmap(_ context.Context, _ string) (*store.PRISMRoadmap, error) {
	return nil, notServed("GetPRISMRoadmap", "PRISM roadmaps")
}

func (s *Store) ListPRISMRoadmaps(_ context.Context) ([]*store.PRISMRoadmap, error) {
	return nil, notServed("ListPRISMRoadmaps", "PRISM roadmaps")
}

func (s *Store) ListPRISMRoadmapsByRepo(_ context.Context, _ string) ([]*store.PRISMRoadmap, error) {
	return nil, notServed("ListPRISMRoadmapsByRepo", "PRISM roadmaps")
}

func (s *Store) UpdatePRISMRoadmap(_ context.Context, _ *store.PRISMRoadmap) error {
	return notServed("UpdatePRISMRoadmap", "PRISM roadmaps")
}

// --- PRISM goals ---

func (s *Store) CreatePRISMGoal(_ context.Context, _ *store.PRISMGoal) error {
	return notServed("CreatePRISMGoal", "PRISM goals")
}

func (s *Store) GetPRISMGoal(_ context.Context, _ string) (*store.PRISMGoal, error) {
	return nil, notServed("GetPRISMGoal", "PRISM goals")
}

func (s *Store) ListPRISMGoals(_ context.Context, _ string) ([]*store.PRISMGoal, error) {
	return nil, notServed("ListPRISMGoals", "PRISM goals")
}

func (s *Store) UpdatePRISMGoal(_ context.Context, _ *store.PRISMGoal) error {
	return notServed("UpdatePRISMGoal", "PRISM goals")
}

// --- PRISM documents ---

func (s *Store) CreatePRISMDocument(_ context.Context, _ *store.PRISMDocument) error {
	return notServed("CreatePRISMDocument", "PRISM documents")
}

func (s *Store) GetPRISMDocument(_ context.Context, _ string) (*store.PRISMDocument, error) {
	return nil, notServed("GetPRISMDocument", "PRISM documents")
}

func (s *Store) ListPRISMDocuments(_ context.Context) ([]*store.PRISMDocument, error) {
	return nil, notServed("ListPRISMDocuments", "PRISM documents")
}

func (s *Store) ListPRISMDocumentsByOrg(_ context.Context, _ string) ([]*store.PRISMDocument, error) {
	return nil, notServed("ListPRISMDocumentsByOrg", "PRISM documents")
}

func (s *Store) ListPRISMDocumentsByRepo(_ context.Context, _ string) ([]*store.PRISMDocument, error) {
	return nil, notServed("ListPRISMDocumentsByRepo", "PRISM documents")
}

func (s *Store) UpdatePRISMDocument(_ context.Context, _ *store.PRISMDocument) error {
	return notServed("UpdatePRISMDocument", "PRISM documents")
}
