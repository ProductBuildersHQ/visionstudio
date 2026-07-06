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
}

// EvalResult represents evaluation results
type EvalResult struct {
	Score    float64      `json:"score"`
	Decision EvalDecision `json:"decision"`
	Findings []Finding    `json:"findings"`
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

// Profile represents a workflow profile
type Profile struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Workflow    []string `json:"workflow"`
	Framework   string   `json:"framework,omitempty"` // e.g., "AWS", "Big Tech", "Startup"
	SpecType    string   `json:"specType,omitempty"`  // e.g., "Product", "Feature", "Platform"
}

// Project represents a spec project
type Project struct {
	Name      string  `json:"name"`
	Path      string  `json:"path"`
	Profile   Profile `json:"profile"`
	Specs     []Spec  `json:"specs"`
	GitRemote string  `json:"gitRemote,omitempty"` // e.g., "https://github.com/org/repo"
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
