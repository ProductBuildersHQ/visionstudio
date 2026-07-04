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
}

// Project represents a spec project
type Project struct {
	Name    string  `json:"name"`
	Path    string  `json:"path"`
	Profile Profile `json:"profile"`
	Specs   []Spec  `json:"specs"`
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
