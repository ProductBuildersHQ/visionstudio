// Package store defines the storage interface for VisionStudio.
// The in-memory fake in this package enables unit testing of all domain
// logic without a running Dolt instance; the doltstore subpackage provides
// the Ent-backed production implementation.
package store

import (
	"context"
	"time"

	"github.com/plexusone/structured-evaluation/rubric"
)

// Store is the persistence interface shared by all service-layer operations.
// Every mutating method runs inside a UnitOfWork; reads may be called directly.
type Store interface {
	InitiativeStore
	ProgramStore
	PhaseStore
	RMIStore
	AssignmentStore
	EvidenceStore
	RepositoryStore
	OrganizationStore
	PersonStore
	ReleaseStore
	SpecWorkflowStore
	JudgeStore
	MaturityStore
	DevXStore
	PRISMRoadmapStore
	PRISMDocumentStore
	SpecDocumentStore
}

// UnitOfWork groups a SQL transaction with a subsequent Dolt commit.
// The production implementation wraps an Ent transaction and issues
// CALL DOLT_COMMIT on success; the in-memory fake is a no-op.
type UnitOfWork interface {
	// Execute runs fn inside a transaction. If fn returns nil the
	// transaction and Dolt commit are applied; otherwise both roll back.
	Execute(ctx context.Context, fn func(ctx context.Context, s Store) error) error
}

// Initiative represents a cross-repository initiative.
type Initiative struct {
	ID           string `json:"id"`
	Organization string `json:"organization"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status"`
	InitType     string `json:"init_type,omitempty"`
	WorkflowID   string `json:"workflow_id,omitempty"`
	Priority     string `json:"priority,omitempty"`
	HomeRepo     string `json:"home_repo,omitempty"`
	Workspace    string `json:"workspace,omitempty"`
	ProgramID    string `json:"program_id,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`
	// Visibility is internal (default) | public. Only public initiatives
	// may appear in external projections, and only via the publicrail
	// two-filter predicate (repo visibility is the second filter).
	Visibility         string            `json:"visibility,omitempty"`
	Specs              map[string]string `json:"specs,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	PlannedAt          *time.Time        `json:"planned_at,omitempty"`
	ExecutingAt        *time.Time        `json:"executing_at,omitempty"`
	DeliveryCompleteAt *time.Time        `json:"delivery_complete_at,omitempty"`
	ReleasedAt         *time.Time        `json:"released_at,omitempty"`
	ClosedAt           *time.Time        `json:"closed_at,omitempty"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// Phase is a themed grouping of RMIs within an initiative.
// Phase status is always derived from member RMIs — never stored.
type Phase struct {
	ID             string `json:"id"`
	InitiativeID   string `json:"initiative_id"`
	SequenceNumber int    `json:"sequence_number"`
	Title          string `json:"title"`
	Theme          string `json:"theme,omitempty"`
}

// RoadmapItem (RMI) is a deliverable within a single repository.
type RoadmapItem struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repository_id"`
	InitiativeID string `json:"initiative_id,omitempty"`
	PhaseID      string `json:"phase_id,omitempty"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	ItemType     string `json:"item_type"`
	Status       string `json:"status"`
	Priority     string `json:"priority,omitempty"`
	Required     bool   `json:"required"`
	// Origin records how the RMI's scope was identified -- see pkg/rmi's
	// Origin constants (spec, implementation, acceptance_testing,
	// discussion). Defaults to "spec".
	Origin             string       `json:"origin,omitempty"`
	SequenceNumber     int          `json:"sequence_number"`
	AcceptanceCriteria []string     `json:"acceptance_criteria,omitempty"`
	ContextSpec        *ContextSpec `json:"context_spec,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	CompletedAt        *time.Time   `json:"completed_at,omitempty"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

// ContextSpec contains explicit overrides for context assembly.
type ContextSpec struct {
	ExtraRepos   []string `json:"extra_repos,omitempty"`
	IncludeSpecs []string `json:"include_specs,omitempty"`
	ExcludeSpecs []string `json:"exclude_specs,omitempty"`
}

// RMIDependency is a directed edge between two RMIs.
type RMIDependency struct {
	SourceRMIID  string `json:"source_rmi_id"`
	TargetRMIID  string `json:"target_rmi_id"`
	Relationship string `json:"relationship"`
}

// InitiativeDependency is a directed edge between two initiatives.
type InitiativeDependency struct {
	SourceInitiativeID string `json:"source_initiative_id"`
	TargetInitiativeID string `json:"target_initiative_id"`
	Relationship       string `json:"relationship"`
}

// Assignment is a lease-based work claim by an agent session.
type Assignment struct {
	ID             string     `json:"id"`
	RMIID          string     `json:"rmi_id"`
	Worker         string     `json:"worker,omitempty"`
	Status         string     `json:"status"`
	LeaseExpiresAt time.Time  `json:"lease_expires_at"`
	Workspace      string     `json:"workspace,omitempty"`
	Handoff        *Handoff   `json:"handoff,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Handoff carries compact state for session continuity.
type Handoff struct {
	Completed  []string `json:"completed"`
	Remaining  []string `json:"remaining"`
	Decisions  []string `json:"decisions"`
	NextAction string   `json:"next_action"`
}

// DeliveryEvidence links a commit, PR, release, or changelog entry to an RMI.
type DeliveryEvidence struct {
	ID           string     `json:"id"`
	RMIID        string     `json:"rmi_id"`
	EvidenceType string     `json:"evidence_type"`
	Reference    string     `json:"reference"`
	CommitType   string     `json:"commit_type,omitempty"`
	CommitScope  string     `json:"commit_scope,omitempty"`
	OccurredAt   *time.Time `json:"occurred_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Repository is a catalog entry for a participating repository.
type Repository struct {
	ID              string `json:"id"`
	Organization    string `json:"organization"`
	RepositoryName  string `json:"repository_name"`
	DefaultBranch   string `json:"default_branch,omitempty"`
	LocalPath       string `json:"local_path,omitempty"`
	GoModule        string `json:"go_module,omitempty"`
	Domain          string `json:"domain,omitempty"`
	Status          string `json:"status"`
	IngestHighWater string `json:"ingest_high_water,omitempty"`
	OrganizationID  string `json:"organization_id,omitempty"`
	// Visibility is public|private|unknown, ingested from GitHub — never
	// hand-maintained. "unknown" must NEVER be treated as public.
	Visibility string `json:"visibility,omitempty"`
	// SupersededBy points at the repository ID that replaced this one
	// (set by `registry archive --superseded-by`). Empty unless archived
	// as part of a merge/rename.
	SupersededBy string `json:"superseded_by,omitempty"`
}

// Organization is a first-class GitHub organization or a user account
// acting as one (e.g. grokify). The Repository.Organization string remains
// for back-compat; OrganizationID carries the queryable edge.
type Organization struct {
	ID             string    `json:"id"` // e.g. "github.com/plexusone"
	Login          string    `json:"login"`
	Kind           string    `json:"kind"` // organization | user
	DisplayName    string    `json:"display_name,omitempty"`
	Website        string    `json:"website,omitempty"`
	ReleasePageURL string    `json:"release_page_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Person is an identity: a human with a GitHub login and the commit-author
// email variants used to attribute work. Affiliation roles are deferred to
// the systemforge membership substrate; OrgIDs records plain affiliation.
type Person struct {
	ID              string    `json:"id"` // e.g. "person:grokify"
	GitHubLogin     string    `json:"github_login"`
	DisplayName     string    `json:"display_name,omitempty"`
	EmailIdentities []string  `json:"email_identities,omitempty"`
	OrgIDs          []string  `json:"org_ids,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Program is an organizational grouping of related initiatives.
type Program struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Organization string    `json:"organization"`
	Description  string    `json:"description,omitempty"`
	Hidden       bool      `json:"hidden,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RepositoryDependency is a directed edge between two repositories
// derived from go.mod dependency analysis.
type RepositoryDependency struct {
	SourceRepositoryID string `json:"source_repository_id"`
	TargetRepositoryID string `json:"target_repository_id"`
	DependencyType     string `json:"dependency_type"`
}

// ProgramStore defines persistence for programs.
type ProgramStore interface {
	CreateProgram(ctx context.Context, prog *Program) error
	GetProgram(ctx context.Context, id string) (*Program, error)
	ListPrograms(ctx context.Context) ([]*Program, error)
	UpdateProgram(ctx context.Context, prog *Program) error
}

// InitiativeStore defines persistence for initiatives.
type InitiativeStore interface {
	CreateInitiative(ctx context.Context, init *Initiative) error
	GetInitiative(ctx context.Context, id string) (*Initiative, error)
	ListInitiatives(ctx context.Context) ([]*Initiative, error)
	UpdateInitiative(ctx context.Context, init *Initiative) error
	CreateInitiativeDependency(ctx context.Context, dep *InitiativeDependency) error
	ListInitiativeDependencies(ctx context.Context, initiativeID string) ([]*InitiativeDependency, error)
	ListAllInitiativeDependencies(ctx context.Context) ([]*InitiativeDependency, error)
}

// PhaseStore defines persistence for phases.
type PhaseStore interface {
	CreatePhase(ctx context.Context, phase *Phase) error
	ListPhases(ctx context.Context, initiativeID string) ([]*Phase, error)
	DeletePhase(ctx context.Context, id string) error
}

// RMIStore defines persistence for roadmap items and dependencies.
type RMIStore interface {
	CreateRMI(ctx context.Context, rmi *RoadmapItem) error
	GetRMI(ctx context.Context, id string) (*RoadmapItem, error)
	ListRMIs(ctx context.Context, initiativeID string) ([]*RoadmapItem, error)
	ListAllRMIs(ctx context.Context) ([]*RoadmapItem, error)
	ListRMIsByRepo(ctx context.Context, repoID string) ([]*RoadmapItem, error)
	ListRMIsByStatus(ctx context.Context, status string) ([]*RoadmapItem, error)
	UpdateRMI(ctx context.Context, rmi *RoadmapItem) error
	CreateDependency(ctx context.Context, dep *RMIDependency) error
	ListDependencies(ctx context.Context, rmiID string) ([]*RMIDependency, error)
	ListAllDependencies(ctx context.Context) ([]*RMIDependency, error)
}

// AssignmentStore defines persistence for lease-based work claims.
type AssignmentStore interface {
	CreateAssignment(ctx context.Context, a *Assignment) error
	GetAssignment(ctx context.Context, id string) (*Assignment, error)
	GetActiveAssignment(ctx context.Context, rmiID string) (*Assignment, error)
	ListActiveAssignments(ctx context.Context) ([]*Assignment, error)
	ListAllAssignments(ctx context.Context) ([]*Assignment, error)
	UpdateAssignment(ctx context.Context, a *Assignment) error
}

// EvidenceStore defines persistence for delivery evidence.
type EvidenceStore interface {
	CreateEvidence(ctx context.Context, ev *DeliveryEvidence) error
	ListEvidenceByRMI(ctx context.Context, rmiID string) ([]*DeliveryEvidence, error)
	ListEvidenceByInitiative(ctx context.Context, initiativeID string) ([]*DeliveryEvidence, error)
	ListAllEvidence(ctx context.Context) ([]*DeliveryEvidence, error)
}

// RepositoryStore defines persistence for the repository catalog.
type RepositoryStore interface {
	CreateRepository(ctx context.Context, repo *Repository) error
	GetRepository(ctx context.Context, id string) (*Repository, error)
	ListRepositories(ctx context.Context) ([]*Repository, error)
	ListRepositoriesByOrg(ctx context.Context, org string) ([]*Repository, error)
	UpdateRepository(ctx context.Context, repo *Repository) error
	DeleteRepository(ctx context.Context, id string) error
	CreateRepoDependency(ctx context.Context, dep *RepositoryDependency) error
	ListRepoDependencies(ctx context.Context, repoID string) ([]*RepositoryDependency, error)
	ListAllRepoDependencies(ctx context.Context) ([]*RepositoryDependency, error)
}

// Release is a per-repository release (a shipped git tag). ID is
// "<repository-id>@<tag>". InitiativeIDs and RMIIDs carry the
// associations: which initiatives and roadmap items this release shipped.
type Release struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Tag          string    `json:"tag"`
	ReleasedAt   time.Time `json:"released_at"`
	URL          string    `json:"url,omitempty"`
	NotesRef     string    `json:"notes_ref,omitempty"`
	// Body is release-notes text captured as match evidence for
	// AI-assisted historical backfill (RMI-VISIONSTUDIO-315) — never
	// treated as fact, never auto-interpreted.
	Body          string    `json:"body,omitempty"`
	InitiativeIDs []string  `json:"initiative_ids,omitempty"`
	RMIIDs        []string  `json:"rmi_ids,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ReleaseStore defines persistence for releases and their associations.
type ReleaseStore interface {
	CreateRelease(ctx context.Context, rel *Release) error
	GetRelease(ctx context.Context, id string) (*Release, error)
	ListReleases(ctx context.Context) ([]*Release, error)
	ListReleasesByRepo(ctx context.Context, repoID string) ([]*Release, error)
	ListReleasesByInitiative(ctx context.Context, initiativeID string) ([]*Release, error)
	UpdateRelease(ctx context.Context, rel *Release) error
	DeleteRelease(ctx context.Context, id string) error
}

// OrganizationStore defines persistence for first-class organizations.
type OrganizationStore interface {
	CreateOrganization(ctx context.Context, org *Organization) error
	GetOrganization(ctx context.Context, id string) (*Organization, error)
	GetOrganizationByLogin(ctx context.Context, login string) (*Organization, error)
	ListOrganizations(ctx context.Context) ([]*Organization, error)
	UpdateOrganization(ctx context.Context, org *Organization) error
}

// PersonStore defines persistence for identities.
type PersonStore interface {
	CreatePerson(ctx context.Context, p *Person) error
	GetPerson(ctx context.Context, id string) (*Person, error)
	ListPeople(ctx context.Context) ([]*Person, error)
	UpdatePerson(ctx context.Context, p *Person) error
}

// SpecWorkflow defines a specification workflow template.
type SpecWorkflow struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	SpecsRequired []string `json:"specs_required,omitempty"`
	SpecsOptional []string `json:"specs_optional,omitempty"`
	InitTypes     []string `json:"init_types,omitempty"`
}

// JudgeResult stores an LLM-as-a-Judge evaluation result.
// It wraps structured-evaluation's rubric.Rubric with additional metadata
// for persistence and querying.
type JudgeResult struct {
	ID           string    `json:"id"`
	InitiativeID string    `json:"initiative_id"`
	SpecPath     string    `json:"spec_path"`
	SpecType     string    `json:"spec_type,omitempty"`
	RubricID     string    `json:"rubric_id"`
	EvaluatedAt  time.Time `json:"evaluated_at"`

	// Report is the full structured-evaluation rubric report.
	// Contains categories, findings, scores, confidence, and decision.
	Report *rubric.Rubric `json:"report"`
}

// Score returns the overall integer score (1-5) from the report.
func (j *JudgeResult) Score() int {
	if j.Report == nil {
		return 0
	}
	return int(j.Report.IntScore)
}

// Pass returns whether the evaluation passed.
func (j *JudgeResult) Pass() bool {
	if j.Report == nil {
		return false
	}
	return j.Report.Pass
}

// Model returns the judge model from the report metadata.
func (j *JudgeResult) Model() string {
	if j.Report == nil || j.Report.Judge == nil {
		return ""
	}
	return j.Report.Judge.Model
}

// Rationale returns the summary from the report.
func (j *JudgeResult) Rationale() string {
	if j.Report == nil {
		return ""
	}
	return j.Report.Summary
}

// SpecWorkflowStore defines persistence for spec workflows.
type SpecWorkflowStore interface {
	CreateSpecWorkflow(ctx context.Context, wf *SpecWorkflow) error
	GetSpecWorkflow(ctx context.Context, id string) (*SpecWorkflow, error)
	ListSpecWorkflows(ctx context.Context) ([]*SpecWorkflow, error)
	UpdateSpecWorkflow(ctx context.Context, wf *SpecWorkflow) error
	DeleteSpecWorkflow(ctx context.Context, id string) error

	// Initiative workflow selection
	SelectWorkflowForInitiative(ctx context.Context, initiativeID, workflowID string) error
	GetWorkflowForInitiative(ctx context.Context, initiativeID string) (*InitiativeWorkflow, error)
}

// JudgeStore defines persistence for judge results. Rubrics are resolved live
// from the visionspec catalog (see cmd/visionstudio spec judge), not persisted
// here.
type JudgeStore interface {
	CreateJudgeResult(ctx context.Context, result *JudgeResult) error
	ListJudgeResults(ctx context.Context, initiativeID string) ([]*JudgeResult, error)
}

// Dimension is a capability area within a maturity model.
type Dimension struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Levels      []Level  `json:"levels,omitempty"`
	Sources     []string `json:"sources,omitempty"`
}

// Level describes a maturity level within a dimension.
type Level struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CapabilityModel defines a capability maturity framework.
type CapabilityModel struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Dimensions  []Dimension `json:"dimensions,omitempty"`
	MaxLevel    int         `json:"max_level"`
}

// DimensionScore captures the assessment for a single dimension.
type DimensionScore struct {
	Level     int    `json:"level"`
	Rationale string `json:"rationale,omitempty"`
	Evidence  string `json:"evidence,omitempty"`
}

// MaturityAssessment captures a point-in-time capability assessment.
type MaturityAssessment struct {
	ID           string                    `json:"id"`
	ModelID      string                    `json:"model_id"`
	InitiativeID string                    `json:"initiative_id,omitempty"`
	Organization string                    `json:"organization,omitempty"`
	Scores       map[string]DimensionScore `json:"scores,omitempty"`
	OverallScore *float64                  `json:"overall_score,omitempty"`
	Summary      string                    `json:"summary,omitempty"`
	AssessedBy   string                    `json:"assessed_by,omitempty"`
	Model        string                    `json:"model,omitempty"`
	AssessedAt   time.Time                 `json:"assessed_at"`
}

// MaturityStore defines persistence for capability models and assessments.
type MaturityStore interface {
	CreateCapabilityModel(ctx context.Context, model *CapabilityModel) error
	GetCapabilityModel(ctx context.Context, id string) (*CapabilityModel, error)
	ListCapabilityModels(ctx context.Context) ([]*CapabilityModel, error)
	UpdateCapabilityModel(ctx context.Context, model *CapabilityModel) error
	CreateMaturityAssessment(ctx context.Context, assessment *MaturityAssessment) error
	GetMaturityAssessment(ctx context.Context, id string) (*MaturityAssessment, error)
	ListMaturityAssessments(ctx context.Context, initiativeID string) ([]*MaturityAssessment, error)
	ListMaturityAssessmentsByOrg(ctx context.Context, org string) ([]*MaturityAssessment, error)
}

// DevXPeriodReport holds developer experience metrics for a period.
type DevXPeriodReport struct {
	ID            string         `json:"id"`
	Organization  string         `json:"organization,omitempty"`
	RepositoryID  string         `json:"repository_id,omitempty"`
	PersonID      string         `json:"person_id"`
	PeriodType    string         `json:"period_type"`
	PeriodLabel   string         `json:"period_label"`
	PeriodStart   time.Time      `json:"period_start"`
	PeriodEnd     time.Time      `json:"period_end"`
	Metrics       map[string]any `json:"metrics,omitempty"`
	ByModel       map[string]any `json:"by_model,omitempty"`
	CoverageScore float64        `json:"coverage_score,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// DevXStore defines persistence for developer experience metrics.
type DevXStore interface {
	CreateDevXPeriodReport(ctx context.Context, report *DevXPeriodReport) error
	GetDevXPeriodReport(ctx context.Context, id string) (*DevXPeriodReport, error)
	ListDevXPeriodReports(ctx context.Context, personID string) ([]*DevXPeriodReport, error)
	ListDevXPeriodReportsByRepo(ctx context.Context, repoID string) ([]*DevXPeriodReport, error)
	ListDevXPeriodReportsByOrg(ctx context.Context, org string) ([]*DevXPeriodReport, error)
}

// PRISMRoadmap holds prism-roadmap artifacts keyed by repo.
type PRISMRoadmap struct {
	ID           string    `json:"id"`
	Organization string    `json:"organization,omitempty"`
	RepositoryID string    `json:"repository_id"`
	Name         string    `json:"name,omitempty"`
	Phases       []any     `json:"phases,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PRISMGoal holds prism-roadmap goals artifacts.
type PRISMGoal struct {
	ID           string         `json:"id"`
	Organization string         `json:"organization,omitempty"`
	RepositoryID string         `json:"repository_id"`
	GoalType     string         `json:"goal_type"`
	Document     map[string]any `json:"document,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// PRISMRoadmapStore defines persistence for prism-roadmap artifacts.
type PRISMRoadmapStore interface {
	CreatePRISMRoadmap(ctx context.Context, roadmap *PRISMRoadmap) error
	GetPRISMRoadmap(ctx context.Context, id string) (*PRISMRoadmap, error)
	ListPRISMRoadmaps(ctx context.Context) ([]*PRISMRoadmap, error)
	ListPRISMRoadmapsByRepo(ctx context.Context, repoID string) ([]*PRISMRoadmap, error)
	UpdatePRISMRoadmap(ctx context.Context, roadmap *PRISMRoadmap) error
	CreatePRISMGoal(ctx context.Context, goal *PRISMGoal) error
	GetPRISMGoal(ctx context.Context, id string) (*PRISMGoal, error)
	ListPRISMGoals(ctx context.Context, repoID string) ([]*PRISMGoal, error)
	UpdatePRISMGoal(ctx context.Context, goal *PRISMGoal) error
}

// PRISMDocument holds prism-maturity domain/stage models.
type PRISMDocument struct {
	ID            string         `json:"id"`
	Organization  string         `json:"organization,omitempty"`
	RepositoryID  string         `json:"repository_id,omitempty"`
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	Version       string         `json:"version,omitempty"`
	Domains       []any          `json:"domains,omitempty"`
	Layers        []any          `json:"layers,omitempty"`
	Metrics       []any          `json:"metrics,omitempty"`
	Maturity      map[string]any `json:"maturity,omitempty"`
	SLIState      map[string]any `json:"sli_state,omitempty"`
	MaturityState map[string]any `json:"maturity_state,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// PRISMDocumentStore defines persistence for prism-maturity documents.
type PRISMDocumentStore interface {
	CreatePRISMDocument(ctx context.Context, doc *PRISMDocument) error
	GetPRISMDocument(ctx context.Context, id string) (*PRISMDocument, error)
	ListPRISMDocuments(ctx context.Context) ([]*PRISMDocument, error)
	ListPRISMDocumentsByOrg(ctx context.Context, org string) ([]*PRISMDocument, error)
	ListPRISMDocumentsByRepo(ctx context.Context, repoID string) ([]*PRISMDocument, error)
	UpdatePRISMDocument(ctx context.Context, doc *PRISMDocument) error
}

// SpecDocument is a visionspec document registry entry.
type SpecDocument struct {
	ID           string    `json:"id"`
	Organization string    `json:"organization,omitempty"`
	RepositoryID string    `json:"repository_id"`
	InitiativeID string    `json:"initiative_id,omitempty"`
	WorkflowID   string    `json:"workflow_id,omitempty"`
	SpecType     string    `json:"spec_type"`
	FilePath     string    `json:"file_path"`
	Title        string    `json:"title,omitempty"`
	Status       string    `json:"status,omitempty"`
	ContentHash  string    `json:"content_hash,omitempty"`
	EvalScore    *int      `json:"eval_score,omitempty"`
	EvalVerdict  string    `json:"eval_verdict,omitempty"`
	SyncedAt     time.Time `json:"synced_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// InitiativeWorkflow records which workflow is selected for an initiative.
type InitiativeWorkflow struct {
	InitiativeID string    `json:"initiative_id"`
	WorkflowID   string    `json:"workflow_id"`
	SelectedAt   time.Time `json:"selected_at"`
}

// SpecDocumentStore defines persistence for the spec document registry.
type SpecDocumentStore interface {
	CreateSpecDocument(ctx context.Context, doc *SpecDocument) error
	GetSpecDocument(ctx context.Context, id string) (*SpecDocument, error)
	ListSpecDocuments(ctx context.Context) ([]*SpecDocument, error)
	ListSpecDocumentsByRepo(ctx context.Context, repoID string) ([]*SpecDocument, error)
	ListSpecDocumentsByInitiative(ctx context.Context, initiativeID string) ([]*SpecDocument, error)
	UpdateSpecDocument(ctx context.Context, doc *SpecDocument) error
	DeleteSpecDocument(ctx context.Context, id string) error
}
