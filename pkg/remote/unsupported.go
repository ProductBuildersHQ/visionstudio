package remote

import (
	"context"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// The methods in this file have no VisionStudio Cloud tenant API endpoint
// yet (the API's first slice covers initiatives and RMIs). Each returns a
// *NotSupportedError wrapping ErrNotSupported, so callers fail clearly
// instead of silently reading empty data or dropping writes.

// --- ProgramStore ---

func (s *Store) CreateProgram(_ context.Context, _ *store.Program) error {
	return notSupported("CreateProgram")
}

func (s *Store) UpdateProgram(_ context.Context, _ *store.Program) error {
	return notSupported("UpdateProgram")
}

// --- InitiativeStore ---

func (s *Store) UpdateInitiative(_ context.Context, _ *store.Initiative) error {
	return notSupported("UpdateInitiative")
}

func (s *Store) CreateInitiativeDependency(_ context.Context, _ *store.InitiativeDependency) error {
	return notSupported("CreateInitiativeDependency")
}

func (s *Store) ListInitiativeDependencies(_ context.Context, _ string) ([]*store.InitiativeDependency, error) {
	return nil, notSupported("ListInitiativeDependencies")
}

func (s *Store) ListAllInitiativeDependencies(_ context.Context) ([]*store.InitiativeDependency, error) {
	return nil, notSupported("ListAllInitiativeDependencies")
}

// --- PhaseStore ---

func (s *Store) CreatePhase(_ context.Context, _ *store.Phase) error {
	return notSupported("CreatePhase")
}

func (s *Store) DeletePhase(_ context.Context, _ string) error {
	return notSupported("DeletePhase")
}

// --- RMIStore ---

func (s *Store) UpdateRMI(_ context.Context, _ *store.RoadmapItem) error {
	return notSupported("UpdateRMI")
}

func (s *Store) CreateDependency(_ context.Context, _ *store.RMIDependency) error {
	return notSupported("CreateDependency")
}

func (s *Store) ListDependencies(_ context.Context, _ string) ([]*store.RMIDependency, error) {
	return nil, notSupported("ListDependencies")
}

func (s *Store) ListAllDependencies(_ context.Context) ([]*store.RMIDependency, error) {
	return nil, notSupported("ListAllDependencies")
}

// --- AssignmentStore ---

func (s *Store) CreateAssignment(_ context.Context, _ *store.Assignment) error {
	return notSupported("CreateAssignment")
}

func (s *Store) GetAssignment(_ context.Context, _ string) (*store.Assignment, error) {
	return nil, notSupported("GetAssignment")
}

func (s *Store) GetActiveAssignment(_ context.Context, _ string) (*store.Assignment, error) {
	return nil, notSupported("GetActiveAssignment")
}

func (s *Store) ListActiveAssignments(_ context.Context) ([]*store.Assignment, error) {
	return nil, notSupported("ListActiveAssignments")
}

func (s *Store) ListAllAssignments(_ context.Context) ([]*store.Assignment, error) {
	return nil, notSupported("ListAllAssignments")
}

func (s *Store) UpdateAssignment(_ context.Context, _ *store.Assignment) error {
	return notSupported("UpdateAssignment")
}

// --- EvidenceStore ---

func (s *Store) CreateEvidence(_ context.Context, _ *store.DeliveryEvidence) error {
	return notSupported("CreateEvidence")
}

func (s *Store) ListEvidenceByRMI(_ context.Context, _ string) ([]*store.DeliveryEvidence, error) {
	return nil, notSupported("ListEvidenceByRMI")
}

func (s *Store) ListEvidenceByInitiative(_ context.Context, _ string) ([]*store.DeliveryEvidence, error) {
	return nil, notSupported("ListEvidenceByInitiative")
}

func (s *Store) ListAllEvidence(_ context.Context) ([]*store.DeliveryEvidence, error) {
	return nil, notSupported("ListAllEvidence")
}

// --- RepositoryStore ---

func (s *Store) CreateRepository(_ context.Context, _ *store.Repository) error {
	return notSupported("CreateRepository")
}

func (s *Store) GetRepository(_ context.Context, _ string) (*store.Repository, error) {
	return nil, notSupported("GetRepository")
}

func (s *Store) ListRepositories(_ context.Context) ([]*store.Repository, error) {
	return nil, notSupported("ListRepositories")
}

func (s *Store) ListRepositoriesByOrg(_ context.Context, _ string) ([]*store.Repository, error) {
	return nil, notSupported("ListRepositoriesByOrg")
}

func (s *Store) UpdateRepository(_ context.Context, _ *store.Repository) error {
	return notSupported("UpdateRepository")
}

func (s *Store) DeleteRepository(_ context.Context, _ string) error {
	return notSupported("DeleteRepository")
}

func (s *Store) CreateRepoDependency(_ context.Context, _ *store.RepositoryDependency) error {
	return notSupported("CreateRepoDependency")
}

func (s *Store) ListRepoDependencies(_ context.Context, _ string) ([]*store.RepositoryDependency, error) {
	return nil, notSupported("ListRepoDependencies")
}

func (s *Store) ListAllRepoDependencies(_ context.Context) ([]*store.RepositoryDependency, error) {
	return nil, notSupported("ListAllRepoDependencies")
}

// --- ReleaseStore ---

func (s *Store) CreateRelease(_ context.Context, _ *store.Release) error {
	return notSupported("CreateRelease")
}

func (s *Store) GetRelease(_ context.Context, _ string) (*store.Release, error) {
	return nil, notSupported("GetRelease")
}

func (s *Store) ListReleases(_ context.Context) ([]*store.Release, error) {
	return nil, notSupported("ListReleases")
}

func (s *Store) ListReleasesByRepo(_ context.Context, _ string) ([]*store.Release, error) {
	return nil, notSupported("ListReleasesByRepo")
}

func (s *Store) ListReleasesByInitiative(_ context.Context, _ string) ([]*store.Release, error) {
	return nil, notSupported("ListReleasesByInitiative")
}

func (s *Store) UpdateRelease(_ context.Context, _ *store.Release) error {
	return notSupported("UpdateRelease")
}

func (s *Store) DeleteRelease(_ context.Context, _ string) error {
	return notSupported("DeleteRelease")
}

// --- OrganizationStore ---

func (s *Store) CreateOrganization(_ context.Context, _ *store.Organization) error {
	return notSupported("CreateOrganization")
}

func (s *Store) GetOrganization(_ context.Context, _ string) (*store.Organization, error) {
	return nil, notSupported("GetOrganization")
}

func (s *Store) GetOrganizationByLogin(_ context.Context, _ string) (*store.Organization, error) {
	return nil, notSupported("GetOrganizationByLogin")
}

func (s *Store) ListOrganizations(_ context.Context) ([]*store.Organization, error) {
	return nil, notSupported("ListOrganizations")
}

func (s *Store) UpdateOrganization(_ context.Context, _ *store.Organization) error {
	return notSupported("UpdateOrganization")
}

// --- PersonStore ---

func (s *Store) CreatePerson(_ context.Context, _ *store.Person) error {
	return notSupported("CreatePerson")
}

func (s *Store) GetPerson(_ context.Context, _ string) (*store.Person, error) {
	return nil, notSupported("GetPerson")
}

func (s *Store) ListPeople(_ context.Context) ([]*store.Person, error) {
	return nil, notSupported("ListPeople")
}

func (s *Store) UpdatePerson(_ context.Context, _ *store.Person) error {
	return notSupported("UpdatePerson")
}

// --- SpecWorkflowStore ---

func (s *Store) CreateSpecWorkflow(_ context.Context, _ *store.SpecWorkflow) error {
	return notSupported("CreateSpecWorkflow")
}

func (s *Store) GetSpecWorkflow(_ context.Context, _ string) (*store.SpecWorkflow, error) {
	return nil, notSupported("GetSpecWorkflow")
}

func (s *Store) ListSpecWorkflows(_ context.Context) ([]*store.SpecWorkflow, error) {
	return nil, notSupported("ListSpecWorkflows")
}

func (s *Store) UpdateSpecWorkflow(_ context.Context, _ *store.SpecWorkflow) error {
	return notSupported("UpdateSpecWorkflow")
}

func (s *Store) DeleteSpecWorkflow(_ context.Context, _ string) error {
	return notSupported("DeleteSpecWorkflow")
}

func (s *Store) SelectWorkflowForInitiative(_ context.Context, _ string, _ string) error {
	return notSupported("SelectWorkflowForInitiative")
}

func (s *Store) GetWorkflowForInitiative(_ context.Context, _ string) (*store.InitiativeWorkflow, error) {
	return nil, notSupported("GetWorkflowForInitiative")
}

// --- JudgeStore ---

func (s *Store) CreateJudgeResult(_ context.Context, _ *store.JudgeResult) error {
	return notSupported("CreateJudgeResult")
}

func (s *Store) ListJudgeResults(_ context.Context, _ string) ([]*store.JudgeResult, error) {
	return nil, notSupported("ListJudgeResults")
}

// --- MaturityStore ---

func (s *Store) CreateCapabilityModel(_ context.Context, _ *store.CapabilityModel) error {
	return notSupported("CreateCapabilityModel")
}

func (s *Store) GetCapabilityModel(_ context.Context, _ string) (*store.CapabilityModel, error) {
	return nil, notSupported("GetCapabilityModel")
}

func (s *Store) ListCapabilityModels(_ context.Context) ([]*store.CapabilityModel, error) {
	return nil, notSupported("ListCapabilityModels")
}

func (s *Store) UpdateCapabilityModel(_ context.Context, _ *store.CapabilityModel) error {
	return notSupported("UpdateCapabilityModel")
}

func (s *Store) CreateMaturityAssessment(_ context.Context, _ *store.MaturityAssessment) error {
	return notSupported("CreateMaturityAssessment")
}

func (s *Store) GetMaturityAssessment(_ context.Context, _ string) (*store.MaturityAssessment, error) {
	return nil, notSupported("GetMaturityAssessment")
}

func (s *Store) ListMaturityAssessments(_ context.Context, _ string) ([]*store.MaturityAssessment, error) {
	return nil, notSupported("ListMaturityAssessments")
}

func (s *Store) ListMaturityAssessmentsByOrg(_ context.Context, _ string) ([]*store.MaturityAssessment, error) {
	return nil, notSupported("ListMaturityAssessmentsByOrg")
}

// --- DevXStore ---

func (s *Store) CreateDevXPeriodReport(_ context.Context, _ *store.DevXPeriodReport) error {
	return notSupported("CreateDevXPeriodReport")
}

func (s *Store) GetDevXPeriodReport(_ context.Context, _ string) (*store.DevXPeriodReport, error) {
	return nil, notSupported("GetDevXPeriodReport")
}

func (s *Store) ListDevXPeriodReports(_ context.Context, _ string) ([]*store.DevXPeriodReport, error) {
	return nil, notSupported("ListDevXPeriodReports")
}

func (s *Store) ListDevXPeriodReportsByRepo(_ context.Context, _ string) ([]*store.DevXPeriodReport, error) {
	return nil, notSupported("ListDevXPeriodReportsByRepo")
}

func (s *Store) ListDevXPeriodReportsByOrg(_ context.Context, _ string) ([]*store.DevXPeriodReport, error) {
	return nil, notSupported("ListDevXPeriodReportsByOrg")
}

// --- PRISMRoadmapStore ---

func (s *Store) CreatePRISMRoadmap(_ context.Context, _ *store.PRISMRoadmap) error {
	return notSupported("CreatePRISMRoadmap")
}

func (s *Store) GetPRISMRoadmap(_ context.Context, _ string) (*store.PRISMRoadmap, error) {
	return nil, notSupported("GetPRISMRoadmap")
}

func (s *Store) ListPRISMRoadmaps(_ context.Context) ([]*store.PRISMRoadmap, error) {
	return nil, notSupported("ListPRISMRoadmaps")
}

func (s *Store) ListPRISMRoadmapsByRepo(_ context.Context, _ string) ([]*store.PRISMRoadmap, error) {
	return nil, notSupported("ListPRISMRoadmapsByRepo")
}

func (s *Store) UpdatePRISMRoadmap(_ context.Context, _ *store.PRISMRoadmap) error {
	return notSupported("UpdatePRISMRoadmap")
}

func (s *Store) CreatePRISMGoal(_ context.Context, _ *store.PRISMGoal) error {
	return notSupported("CreatePRISMGoal")
}

func (s *Store) GetPRISMGoal(_ context.Context, _ string) (*store.PRISMGoal, error) {
	return nil, notSupported("GetPRISMGoal")
}

func (s *Store) ListPRISMGoals(_ context.Context, _ string) ([]*store.PRISMGoal, error) {
	return nil, notSupported("ListPRISMGoals")
}

func (s *Store) UpdatePRISMGoal(_ context.Context, _ *store.PRISMGoal) error {
	return notSupported("UpdatePRISMGoal")
}

// --- PRISMDocumentStore ---

func (s *Store) CreatePRISMDocument(_ context.Context, _ *store.PRISMDocument) error {
	return notSupported("CreatePRISMDocument")
}

func (s *Store) GetPRISMDocument(_ context.Context, _ string) (*store.PRISMDocument, error) {
	return nil, notSupported("GetPRISMDocument")
}

func (s *Store) ListPRISMDocuments(_ context.Context) ([]*store.PRISMDocument, error) {
	return nil, notSupported("ListPRISMDocuments")
}

func (s *Store) ListPRISMDocumentsByOrg(_ context.Context, _ string) ([]*store.PRISMDocument, error) {
	return nil, notSupported("ListPRISMDocumentsByOrg")
}

func (s *Store) ListPRISMDocumentsByRepo(_ context.Context, _ string) ([]*store.PRISMDocument, error) {
	return nil, notSupported("ListPRISMDocumentsByRepo")
}

func (s *Store) UpdatePRISMDocument(_ context.Context, _ *store.PRISMDocument) error {
	return notSupported("UpdatePRISMDocument")
}

// --- SpecDocumentStore ---

func (s *Store) CreateSpecDocument(_ context.Context, _ *store.SpecDocument) error {
	return notSupported("CreateSpecDocument")
}

func (s *Store) GetSpecDocument(_ context.Context, _ string) (*store.SpecDocument, error) {
	return nil, notSupported("GetSpecDocument")
}

func (s *Store) ListSpecDocuments(_ context.Context) ([]*store.SpecDocument, error) {
	return nil, notSupported("ListSpecDocuments")
}

func (s *Store) ListSpecDocumentsByRepo(_ context.Context, _ string) ([]*store.SpecDocument, error) {
	return nil, notSupported("ListSpecDocumentsByRepo")
}

func (s *Store) ListSpecDocumentsByInitiative(_ context.Context, _ string) ([]*store.SpecDocument, error) {
	return nil, notSupported("ListSpecDocumentsByInitiative")
}

func (s *Store) UpdateSpecDocument(_ context.Context, _ *store.SpecDocument) error {
	return notSupported("UpdateSpecDocument")
}

func (s *Store) DeleteSpecDocument(_ context.Context, _ string) error {
	return notSupported("DeleteSpecDocument")
}
