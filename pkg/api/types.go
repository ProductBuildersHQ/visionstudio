package api

// SpecStatus represents the status of a spec
type SpecStatus string

const (
	SpecStatusNotStarted SpecStatus = "not_started"
	SpecStatusDraft      SpecStatus = "draft"
	SpecStatusEvaluated  SpecStatus = "evaluated"
	SpecStatusApproved   SpecStatus = "approved"
)

// EvalDecision represents the evaluation decision
type EvalDecision string

const (
	EvalDecisionPass        EvalDecision = "pass"
	EvalDecisionConditional EvalDecision = "conditional"
	EvalDecisionFail        EvalDecision = "fail"
)

// Finding represents an evaluation finding
type Finding struct {
	Category string `json:"category"`
	Severity string `json:"severity"`
	Message  string `json:"message"`

	// V2 fields
	Code     string `json:"code,omitempty"`     // Reason code (e.g., "AMBIGUOUS_REQUIREMENT")
	Location string `json:"location,omitempty"` // Reference to where issue was found (e.g., "REQ-12")
}

// DimensionScore represents a single evaluation dimension (v2)
type DimensionScore struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Score       int       `json:"score"`      // 1-5 integer score
	Severity    string    `json:"severity"`   // "none", "minor", "major", "critical"
	Confidence  float64   `json:"confidence"` // 0.0-1.0
	ReasonCodes []string  `json:"reasonCodes"`
	Findings    []Finding `json:"findings"`
}

// EvalResult represents evaluation results
type EvalResult struct {
	// V1 fields (for backwards compatibility)
	Score    float64      `json:"score"`
	Decision EvalDecision `json:"decision"`
	Findings []Finding    `json:"findings"`

	// V2 fields
	SchemaVersion string           `json:"schemaVersion,omitempty"` // "v2" for new format
	ScoreV2       int              `json:"scoreV2,omitempty"`       // 1-5 integer score
	Pass          bool             `json:"pass,omitempty"`          // Explicit pass/fail gate
	Confidence    float64          `json:"confidence,omitempty"`    // 0.0-1.0
	Dimensions    []DimensionScore `json:"dimensions,omitempty"`    // Per-dimension scores
	Blocking      []string         `json:"blocking,omitempty"`      // Blocking reason codes
}

// Spec represents a specification document
type Spec struct {
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	Status     SpecStatus  `json:"status"`
	EvalResult *EvalResult `json:"evalResult,omitempty"`
	Content    string      `json:"content,omitempty"`
}

// Profile represents a workflow profile (requirements methodology)
type Profile struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Workflow    []string `json:"workflow"`
	Framework   string   `json:"framework,omitempty"` // e.g., "AWS", "Big Tech", "Startup"
	SpecType    string   `json:"specType,omitempty"`  // e.g., "Product", "Feature", "Platform"
	Type        string   `json:"type,omitempty"`      // "requirements" or "implementation"
}

// Project represents a spec project
type Project struct {
	Name                      string  `json:"name"`
	Path                      string  `json:"path"`
	Profile                   Profile `json:"profile"`                             // kept for backwards compat
	RequirementsMethodology   string  `json:"requirementsMethodology,omitempty"`   // e.g., "aws-working-backwards/product"
	ImplementationMethodology string  `json:"implementationMethodology,omitempty"` // e.g., "aidlc", "speckit", "none"
	Specs                     []Spec  `json:"specs"`
	GitRemote                 string  `json:"gitRemote,omitempty"` // e.g., "https://github.com/org/repo"
}

// API request/response types

type ListProjectsResponse struct {
	Projects []Project `json:"projects"`
}

type GetProjectResponse struct {
	Project Project `json:"project"`
}

type GetSpecResponse struct {
	Spec Spec `json:"spec"`
}

type SaveSpecRequest struct {
	Content string `json:"content"`
}

type SaveSpecResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type EvaluateSpecResponse struct {
	Result EvalResult `json:"result"`
	Error  string     `json:"error,omitempty"`
}

type ChatRequest struct {
	Message string `json:"message"`
	Context string `json:"context,omitempty"` // Current spec content for context
}

type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// WorkflowStatus represents the current workflow state for a project
type WorkflowStatus struct {
	CurrentPhase    string            `json:"currentPhase"`
	CompletedPhases []string          `json:"completedPhases"`
	Progress        float64           `json:"progress"` // 0.0 to 1.0
	SpecStatuses    map[string]string `json:"specStatuses"`
	BlockedBy       []string          `json:"blockedBy,omitempty"`
	LastUpdated     string            `json:"lastUpdated"`
}

type GetWorkflowStatusResponse struct {
	Status WorkflowStatus `json:"status"`
	Error  string         `json:"error,omitempty"`
}

// Project management types

type AddProjectRequest struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Profile    string `json:"profile"`
	Initialize bool   `json:"initialize"` // If true, create directory structure and scaffold specs
}

type AddProjectResponse struct {
	Project Project `json:"project"`
	Error   string  `json:"error,omitempty"`
}

type RemoveProjectResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ListProfilesResponse struct {
	Profiles []Profile `json:"profiles"`
}

// ImplementationMethodologySummary provides info about an implementation methodology
type ImplementationMethodologySummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListImplementationMethodologiesResponse is the response for listing implementation methodologies
type ListImplementationMethodologiesResponse struct {
	Methodologies []ImplementationMethodologySummary `json:"methodologies"`
}

// ProjectMethodologyConfig represents a project's methodology configuration
type ProjectMethodologyConfig struct {
	RequirementsMethodology   string `json:"requirementsMethodology"`
	ImplementationMethodology string `json:"implementationMethodology"`
}

// GetProjectMethodologyResponse is the response for getting a project's methodology config
type GetProjectMethodologyResponse struct {
	Config ProjectMethodologyConfig `json:"config"`
	Error  string                   `json:"error,omitempty"`
}

// UpdateProjectMethodologyRequest is the request for updating methodology selection
type UpdateProjectMethodologyRequest struct {
	RequirementsMethodology   string `json:"requirementsMethodology,omitempty"`
	ImplementationMethodology string `json:"implementationMethodology,omitempty"`
}

// UpdateProjectMethodologyResponse is the response for updating methodology selection
type UpdateProjectMethodologyResponse struct {
	Config ProjectMethodologyConfig `json:"config"`
	Error  string                   `json:"error,omitempty"`
}

// Lint types

type LintFinding struct {
	Path     string `json:"path"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error", "warning", "info"
}

type LintResult struct {
	Project  string        `json:"project"`
	Findings []LintFinding `json:"findings"`
	Errors   int           `json:"errors"`
	Warnings int           `json:"warnings"`
	Passed   bool          `json:"passed"`
}

type LintProjectResponse struct {
	Result LintResult `json:"result"`
	Error  string     `json:"error,omitempty"`
}

// Workflow types

// WorkflowNode represents a node in the workflow DAG
type WorkflowNode struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Type        string         `json:"type"`   // "source", "gtm", "technical", "output"
	Phase       string         `json:"phase"`  // Phase this node belongs to
	Status      string         `json:"status"` // "pending", "ready", "in_progress", "completed", "blocked", "skipped"
	DependsOn   []string       `json:"depends_on,omitempty"`
	Automated   bool           `json:"automated,omitempty"` // True if LLM-generated
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// WorkflowPhase represents a phase in the workflow
type WorkflowPhase struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Order       int      `json:"order"`
	Nodes       []string `json:"nodes"` // Node IDs in this phase
}

// Workflow represents a complete workflow graph
type Workflow struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Phases      []WorkflowPhase         `json:"phases"`
	Nodes       map[string]WorkflowNode `json:"nodes"`
	Progress    WorkflowProgress        `json:"progress"`
}

// WorkflowProgress tracks completion progress
type WorkflowProgress struct {
	Completed int     `json:"completed"`
	Total     int     `json:"total"`
	Percent   float64 `json:"percent"`
}

type GetWorkflowResponse struct {
	Workflow Workflow `json:"workflow"`
	Mermaid  string   `json:"mermaid,omitempty"` // Mermaid diagram representation
	Error    string   `json:"error,omitempty"`
}

// Real-time event types

// FileEventType represents the type of file system event
type FileEventType string

const (
	EventFileCreated     FileEventType = "file_created"
	EventFileModified    FileEventType = "file_modified"
	EventFileDeleted     FileEventType = "file_deleted"
	EventFileRenamed     FileEventType = "file_renamed"
	EventEvalStarted     FileEventType = "eval_started"
	EventEvalComplete    FileEventType = "eval_complete"
	EventLintComplete    FileEventType = "lint_complete"
	EventSpecUpdated     FileEventType = "spec_updated"
	EventWorkflowChanged FileEventType = "workflow_changed"
)

// FileEvent represents a real-time event sent via SSE
type FileEvent struct {
	Type      FileEventType  `json:"type"`
	Project   string         `json:"project"`
	Path      string         `json:"path,omitempty"`     // Relative path within project
	SpecType  string         `json:"specType,omitempty"` // If this is a spec file
	Timestamp string         `json:"timestamp"`
	Data      map[string]any `json:"data,omitempty"` // Additional event-specific data
}

// Maturity Model types

// MaturityModelSummary provides a summary of a maturity model for listing
type MaturityModelSummary struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	DimensionCount int     `json:"dimensionCount"`
	GoalCount      int     `json:"goalCount"`
	OverallScore   float64 `json:"overallScore"`
	LastUpdated    string  `json:"lastUpdated"`
}

// Sample types for reference implementations

// SampleSummary provides a summary of an available sample
type SampleSummary struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Complexity  string         `json:"complexity"` // "simple", "enterprise"
	Path        string         `json:"path"`
	FileCounts  map[string]int `json:"fileCounts"` // {"v2mom": 1, "capability": 1, etc.}
}

// SampleDetail provides full details of a sample
type SampleDetail struct {
	SampleSummary
	ProjectJSON map[string]any `json:"projectJson,omitempty"`
	README      string         `json:"readme,omitempty"`
}

// ListSamplesResponse is the response for listing available samples
type ListSamplesResponse struct {
	Samples []SampleSummary `json:"samples"`
}

// GetSampleResponse is the response for getting sample details
type GetSampleResponse struct {
	Sample SampleDetail `json:"sample"`
	Error  string       `json:"error,omitempty"`
}

// V2MOM types for strategic planning

// V2MOMSummary provides a summary of a V2MOM for listing
type V2MOMSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parentId,omitempty"`
	Path     string `json:"path"`
}

// V2MOM represents a full V2MOM document
type V2MOM struct {
	Metadata  map[string]any   `json:"metadata"`
	Vision    string           `json:"vision"`
	Values    []map[string]any `json:"values"`
	Methods   []map[string]any `json:"methods"`
	Obstacles []map[string]any `json:"obstacles"`
	Measures  []map[string]any `json:"measures"`
}

// V2MOMCascade represents a hierarchical V2MOM structure
type V2MOMCascade struct {
	Root     V2MOM   `json:"root"`
	Children []V2MOM `json:"children"`
}

// ListV2MOMsResponse is the response for listing V2MOMs
type ListV2MOMsResponse struct {
	V2MOMs []V2MOMSummary `json:"v2moms"`
}

// GetV2MOMResponse is the response for getting a V2MOM
type GetV2MOMResponse struct {
	V2MOM V2MOM  `json:"v2mom"`
	Error string `json:"error,omitempty"`
}

// GetV2MOMCascadeResponse is the response for getting V2MOM cascade
type GetV2MOMCascadeResponse struct {
	Cascade V2MOMCascade `json:"cascade"`
	Error   string       `json:"error,omitempty"`
}

// Capability types

// CapabilitySummary provides a summary of a capability stack
type CapabilitySummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain,omitempty"`
	Path   string `json:"path"`
}

// CapabilityStack represents a full capability stack document
type CapabilityStack struct {
	Metadata         map[string]any   `json:"metadata"`
	Layers           []map[string]any `json:"layers"`
	Categories       []map[string]any `json:"categories"`
	Capabilities     []map[string]any `json:"capabilities"`
	PrismIntegration map[string]any   `json:"prismIntegration,omitempty"`
}

// ListCapabilitiesResponse is the response for listing capability stacks
type ListCapabilitiesResponse struct {
	Capabilities []CapabilitySummary `json:"capabilities"`
}

// GetCapabilityResponse is the response for getting a capability stack
type GetCapabilityResponse struct {
	Capability CapabilityStack `json:"capability"`
	Error      string          `json:"error,omitempty"`
}

// Roadmap types

// RoadmapItem represents a single roadmap item
type RoadmapItem struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Status         string         `json:"status"`
	Priority       string         `json:"priority"`
	Quarter        string         `json:"quarter"`
	Effort         string         `json:"effort"`
	RICE           map[string]any `json:"rice,omitempty"`
	CapabilityRefs []string       `json:"capability_refs,omitempty"`
	GoalRefs       []string       `json:"goal_refs,omitempty"`
}

// Roadmap represents a full roadmap document
type Roadmap struct {
	Metadata map[string]any `json:"metadata"`
	Items    []RoadmapItem  `json:"items"`
}

// GetRoadmapResponse is the response for getting a roadmap
type GetRoadmapResponse struct {
	Roadmap Roadmap `json:"roadmap"`
	Error   string  `json:"error,omitempty"`
}

// AIDLC types for AWS AI DLC workflow integration

// AIDLCPhase represents an AIDLC workflow phase
type AIDLCPhase string

const (
	AIDLCPhaseInception    AIDLCPhase = "inception"
	AIDLCPhaseConstruction AIDLCPhase = "construction"
	AIDLCPhaseOperations   AIDLCPhase = "operations"
)

// AIDLCDocStatus represents the status of an AIDLC document
type AIDLCDocStatus string

const (
	AIDLCDocStatusPending    AIDLCDocStatus = "pending"
	AIDLCDocStatusDraft      AIDLCDocStatus = "draft"
	AIDLCDocStatusInProgress AIDLCDocStatus = "in_progress"
	AIDLCDocStatusCompleted  AIDLCDocStatus = "completed"
	AIDLCDocStatusBlocked    AIDLCDocStatus = "blocked"
)

// AIDLCQualityRating represents quality assessment rating
type AIDLCQualityRating string

const (
	AIDLCRatingExcellent        AIDLCQualityRating = "EXCELLENT"
	AIDLCRatingGood             AIDLCQualityRating = "GOOD"
	AIDLCRatingNeedsImprovement AIDLCQualityRating = "NEEDS_IMPROVEMENT"
	AIDLCRatingPoor             AIDLCQualityRating = "POOR"
)

// AIDLCIssue represents a quality issue
type AIDLCIssue struct {
	Severity   string `json:"severity"`
	Category   string `json:"category"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message"`
	Location   string `json:"location,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// AIDLCQualityScore represents evaluation results
type AIDLCQualityScore struct {
	Rating      AIDLCQualityRating `json:"rating"`
	Score       float64            `json:"score"`
	Issues      []AIDLCIssue       `json:"issues,omitempty"`
	Dimensions  map[string]float64 `json:"dimensions,omitempty"`
	EvaluatedAt string             `json:"evaluated_at,omitempty"`
}

// AIDLCDocument represents an AIDLC document
type AIDLCDocument struct {
	Type      string             `json:"type"`
	Phase     AIDLCPhase         `json:"phase"`
	Path      string             `json:"path"`
	Title     string             `json:"title"`
	Status    AIDLCDocStatus     `json:"status"`
	Content   string             `json:"content,omitempty"`
	Score     *AIDLCQualityScore `json:"score,omitempty"`
	UpdatedAt string             `json:"updated_at"`
}

// AIDLCState represents the parsed aidlc-state.md content
type AIDLCState struct {
	CurrentPhase    AIDLCPhase                    `json:"current_phase"`
	CurrentDocument string                        `json:"current_document,omitempty"`
	CompletedDocs   []string                      `json:"completed_docs"`
	PendingDocs     []string                      `json:"pending_docs"`
	InProgressDocs  []string                      `json:"in_progress_docs,omitempty"`
	DocumentScores  map[string]*AIDLCQualityScore `json:"document_scores,omitempty"`
	PhaseProgress   map[AIDLCPhase]float64        `json:"phase_progress,omitempty"`
	OverallProgress float64                       `json:"overall_progress"`
	LastUpdated     string                        `json:"last_updated"`
}

// AIDLCWorkflowNode represents a node in the workflow DAG
type AIDLCWorkflowNode struct {
	ID          string             `json:"id"`
	DocType     string             `json:"doc_type"`
	Phase       AIDLCPhase         `json:"phase"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Status      string             `json:"status"`
	Score       *AIDLCQualityScore `json:"score,omitempty"`
	DependsOn   []string           `json:"depends_on,omitempty"`
	Blocks      []string           `json:"blocks,omitempty"`
	Required    bool               `json:"required"`
	Automated   bool               `json:"automated,omitempty"`
}

// AIDLCWorkflowPhase represents a phase in the workflow
type AIDLCWorkflowPhase struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Order       int      `json:"order"`
	NodeIDs     []string `json:"node_ids"`
}

// AIDLCWorkflowEdge represents an edge in the workflow DAG
type AIDLCWorkflowEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Type  string `json:"type"` // dependency, blocks, suggests
	Label string `json:"label,omitempty"`
}

// AIDLCWorkflow represents the full AIDLC workflow
type AIDLCWorkflow struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description,omitempty"`
	Phases      []AIDLCWorkflowPhase         `json:"phases"`
	Nodes       map[string]AIDLCWorkflowNode `json:"nodes"`
	Edges       []AIDLCWorkflowEdge          `json:"edges"`
	Progress    WorkflowProgress             `json:"progress"`
}

// AIDLCSyncAction represents a sync operation
type AIDLCSyncAction struct {
	Direction  string `json:"direction"` // to_aidlc, from_aidlc
	DocType    string `json:"doc_type"`
	SourcePath string `json:"source_path"`
	DestPath   string `json:"dest_path"`
	Action     string `json:"action"` // create, update, delete
	Reason     string `json:"reason,omitempty"`
}

// AIDLCSyncConflict represents a sync conflict
type AIDLCSyncConflict struct {
	DocType           string `json:"doc_type"`
	VisionSpecPath    string `json:"visionspec_path"`
	AIDLCPath         string `json:"aidlc_path"`
	VisionSpecModTime string `json:"visionspec_mod_time"`
	AIDLCModTime      string `json:"aidlc_mod_time"`
	Reason            string `json:"reason"`
}

// AIDLCSyncDiff represents differences between visionspec and aidlc directories
type AIDLCSyncDiff struct {
	VisionSpecDir string              `json:"visionspec_dir"`
	AIDLCDocsDir  string              `json:"aidlc_docs_dir"`
	Actions       []AIDLCSyncAction   `json:"actions"`
	Conflicts     []AIDLCSyncConflict `json:"conflicts,omitempty"`
	ComputedAt    string              `json:"computed_at"`
}

// AIDLCSyncResult represents the result of a sync operation
type AIDLCSyncResult struct {
	Direction   string   `json:"direction"`
	Created     []string `json:"created,omitempty"`
	Updated     []string `json:"updated,omitempty"`
	Skipped     []string `json:"skipped,omitempty"`
	Errors      []string `json:"errors,omitempty"`
	CompletedAt string   `json:"completed_at"`
}

// AIDLC API request/response types

// GetAIDLCStateResponse is the response for getting AIDLC state
type GetAIDLCStateResponse struct {
	State AIDLCState `json:"state"`
	Error string     `json:"error,omitempty"`
}

// GetAIDLCWorkflowResponse is the response for getting AIDLC workflow
type GetAIDLCWorkflowResponse struct {
	Workflow AIDLCWorkflow `json:"workflow"`
	Mermaid  string        `json:"mermaid,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// ListAIDLCDocumentsResponse is the response for listing AIDLC documents
type ListAIDLCDocumentsResponse struct {
	Documents []AIDLCDocument `json:"documents"`
	Error     string          `json:"error,omitempty"`
}

// GetAIDLCDocumentResponse is the response for getting an AIDLC document
type GetAIDLCDocumentResponse struct {
	Document AIDLCDocument `json:"document"`
	Error    string        `json:"error,omitempty"`
}

// AIDLCSyncRequest is the request for triggering a sync
type AIDLCSyncRequest struct {
	Direction string `json:"direction,omitempty"` // to_aidlc, from_aidlc, bidirectional
	DryRun    bool   `json:"dry_run,omitempty"`
}

// GetAIDLCSyncDiffResponse is the response for getting sync diff
type GetAIDLCSyncDiffResponse struct {
	Diff  AIDLCSyncDiff `json:"diff"`
	Error string        `json:"error,omitempty"`
}

// AIDLCSyncResponse is the response for a sync operation
type AIDLCSyncResponse struct {
	Result AIDLCSyncResult `json:"result"`
	Error  string          `json:"error,omitempty"`
}

// AIDLCPhaseRequirements represents requirements for a phase
type AIDLCPhaseRequirements struct {
	Phase           AIDLCPhase `json:"phase"`
	RequiredDocs    []string   `json:"required_docs"`
	CompletedDocs   []string   `json:"completed_docs"`
	PendingDocs     []string   `json:"pending_docs"`
	ProgressPercent float64    `json:"progress_percent"`
	CanAdvance      bool       `json:"can_advance"`
}

// GetAIDLCPhaseRequirementsResponse is the response for phase requirements
type GetAIDLCPhaseRequirementsResponse struct {
	CurrentPhase AIDLCPhase               `json:"current_phase"`
	Requirements []AIDLCPhaseRequirements `json:"requirements"`
	Error        string                   `json:"error,omitempty"`
}

// AIDLCPhaseTransitionRequest is the request for phase transition
type AIDLCPhaseTransitionRequest struct {
	TargetPhase AIDLCPhase `json:"target_phase"`
	Force       bool       `json:"force,omitempty"` // Force transition even if not all requirements are met
	ApprovedBy  string     `json:"approved_by,omitempty"`
	Notes       string     `json:"notes,omitempty"`
}

// AIDLCPhaseTransitionResponse is the response for phase transition
type AIDLCPhaseTransitionResponse struct {
	Success        bool       `json:"success"`
	FromPhase      AIDLCPhase `json:"from_phase"`
	ToPhase        AIDLCPhase `json:"to_phase"`
	BlockingDocs   []string   `json:"blocking_docs,omitempty"`
	BlockingIssues []string   `json:"blocking_issues,omitempty"`
	Error          string     `json:"error,omitempty"`
}

// AIDLCTemplateSection represents a section in an AIDLC template
type AIDLCTemplateSection struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

// AIDLCTemplate represents an AIDLC document template
type AIDLCTemplate struct {
	DocType     string                 `json:"doc_type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Phase       AIDLCPhase             `json:"phase"`
	Sections    []AIDLCTemplateSection `json:"sections"`
	Content     string                 `json:"content,omitempty"` // Full template content
}

// ListAIDLCTemplatesResponse is the response for listing templates
type ListAIDLCTemplatesResponse struct {
	Templates []AIDLCTemplate `json:"templates"`
	Error     string          `json:"error,omitempty"`
}

// GetAIDLCTemplateResponse is the response for getting a template
type GetAIDLCTemplateResponse struct {
	Template AIDLCTemplate `json:"template"`
	Error    string        `json:"error,omitempty"`
}

// CreateAIDLCDocumentRequest is the request for creating a document
type CreateAIDLCDocumentRequest struct {
	DocType     string `json:"doc_type"`
	Title       string `json:"title,omitempty"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description,omitempty"`
	Overwrite   bool   `json:"overwrite,omitempty"` // Overwrite if exists
}

// CreateAIDLCDocumentResponse is the response for creating a document
type CreateAIDLCDocumentResponse struct {
	Document AIDLCDocument `json:"document"`
	Error    string        `json:"error,omitempty"`
}

// Organization types for workspace-level management

// Organization represents the workspace/organization configuration
type Organization struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	V2MOMPath       string `json:"v2momPath,omitempty"`
	FiscalYearStart string `json:"fiscalYearStart,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// V2MOMScope indicates whether a V2MOM is organization or project level
type V2MOMScope string

const (
	V2MOMScopeOrganization V2MOMScope = "organization"
	V2MOMScopeProject      V2MOMScope = "project"
)

// OrganizationV2MOM represents an organization-level V2MOM
type OrganizationV2MOM struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	FiscalYear  string         `json:"fiscalYear,omitempty"`
	Owner       string         `json:"owner,omitempty"`
	Scope       V2MOMScope     `json:"scope"`
	ParentID    string         `json:"parentId,omitempty"`
	Path        string         `json:"path"`
	Vision      string         `json:"vision"`
	Values      []V2MOMValue   `json:"values"`
	Methods     []V2MOMMethod  `json:"methods"`
	Obstacles   []V2MOMItem    `json:"obstacles"`
	Measures    []V2MOMMeasure `json:"measures"`
	LastUpdated string         `json:"lastUpdated"`
}

// V2MOMValue represents a value in a V2MOM
type V2MOMValue struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Priority    int    `json:"priority,omitempty"`
}

// V2MOMMethod represents a method in a V2MOM
type V2MOMMethod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"` // not_started, in_progress, at_risk, completed
	Owner       string `json:"owner,omitempty"`
	DueDate     string `json:"dueDate,omitempty"`
	Priority    int    `json:"priority,omitempty"`
}

// V2MOMItem represents a generic V2MOM item (used for obstacles)
type V2MOMItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Severity    string `json:"severity,omitempty"` // low, medium, high, critical
	Mitigation  string `json:"mitigation,omitempty"`
}

// V2MOMMeasure represents a measure in a V2MOM
type V2MOMMeasure struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Target      string  `json:"target,omitempty"`
	Current     string  `json:"current,omitempty"`
	Unit        string  `json:"unit,omitempty"`
	Progress    float64 `json:"progress,omitempty"` // 0.0 to 1.0
}

// V2MOMMethodAlignment tracks alignment between project and org methods
type V2MOMMethodAlignment struct {
	ProjectV2MOMID string  `json:"projectV2momId"`
	ProjectMethod  string  `json:"projectMethod"`
	OrgV2MOMID     string  `json:"orgV2momId"`
	OrgMethod      string  `json:"orgMethod"`
	AlignmentScore float64 `json:"alignmentScore"` // 0 to 100
	Notes          string  `json:"notes,omitempty"`
}

// OrganizationCascade represents the full org → projects cascade
type OrganizationCascade struct {
	Organization    *Organization             `json:"organization"`
	OrgV2MOMs       []OrganizationV2MOM       `json:"orgV2moms"`
	ProjectV2MOMs   map[string][]V2MOMSummary `json:"projectV2moms"` // project name → v2moms
	Alignments      []V2MOMMethodAlignment    `json:"alignments"`
	AlignmentScores map[string]float64        `json:"alignmentScores"` // project name → avg score
}

// Organization API request/response types

// GetOrganizationResponse is the response for getting organization
type GetOrganizationResponse struct {
	Organization *Organization `json:"organization"`
	Error        string        `json:"error,omitempty"`
}

// CreateOrganizationRequest is the request for creating an organization
type CreateOrganizationRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	FiscalYearStart string `json:"fiscalYearStart,omitempty"`
}

// CreateOrganizationResponse is the response for creating an organization
type CreateOrganizationResponse struct {
	Organization *Organization `json:"organization"`
	Error        string        `json:"error,omitempty"`
}

// UpdateOrganizationRequest is the request for updating an organization
type UpdateOrganizationRequest struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	V2MOMPath       string `json:"v2momPath,omitempty"`
	FiscalYearStart string `json:"fiscalYearStart,omitempty"`
}

// UpdateOrganizationResponse is the response for updating an organization
type UpdateOrganizationResponse struct {
	Organization *Organization `json:"organization"`
	Error        string        `json:"error,omitempty"`
}

// ListOrganizationV2MOMsResponse is the response for listing org V2MOMs
type ListOrganizationV2MOMsResponse struct {
	V2MOMs []OrganizationV2MOM `json:"v2moms"`
	Error  string              `json:"error,omitempty"`
}

// GetOrganizationV2MOMResponse is the response for getting an org V2MOM
type GetOrganizationV2MOMResponse struct {
	V2MOM OrganizationV2MOM `json:"v2mom"`
	Error string            `json:"error,omitempty"`
}

// GetOrganizationCascadeResponse is the response for the full cascade
type GetOrganizationCascadeResponse struct {
	Cascade OrganizationCascade `json:"cascade"`
	Error   string              `json:"error,omitempty"`
}
