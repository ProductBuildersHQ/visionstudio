// Package apitypes defines the canonical API response types for VisionStudio.
// These types are the source of truth for JSON Schema generation.
// TypeScript types are generated from these via the schema pipeline:
//
//	Go structs → JSON Schema → Zod → TypeScript
//
// JSON tags use camelCase to match JavaScript conventions.
// Do NOT manually edit web/src/api/types.gen.ts or schemas.gen.ts.
// Instead, modify these types and run: go generate ./pkg/apitypes
package apitypes

import (
	"time"

	"github.com/plexusone/structured-evaluation/rubric"
)

//go:generate go run ./gen/main.go

// JudgeResult is an LLM-as-a-Judge evaluation.
// Report uses rubric.Rubric directly from structured-evaluation.
type JudgeResult struct {
	ID           string         `json:"id"`
	InitiativeID string         `json:"initiativeId"`
	SpecPath     string         `json:"specPath"`
	SpecType     string         `json:"specType,omitempty"`
	RubricID     string         `json:"rubricId,omitempty"`
	EvaluatedAt  time.Time      `json:"evaluatedAt"`
	Report       *rubric.Rubric `json:"report,omitempty"`
}

// SpecWorkflow defines a specification workflow template.
// Definitions are sourced from the specification-workflow-spec catalog.
type SpecWorkflow struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	SpecsRequired []string            `json:"specsRequired,omitempty"`
	SpecsOptional []string            `json:"specsOptional,omitempty"`
	InitTypes     []string            `json:"initTypes,omitempty"`
	Sequence      []string            `json:"sequence,omitempty"`
	Phases        []SpecWorkflowPhase `json:"phases,omitempty"`
}

// SpecWorkflowPhase is a named group of spec types within a workflow's flow.
type SpecWorkflowPhase struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Specs []string `json:"specs"`
}

// SpecsResponse is the response for /api/specs.
type SpecsResponse struct {
	Workflows    []SpecWorkflow `json:"workflows"`
	JudgeResults []JudgeResult  `json:"judgeResults"`
}

// WorkflowSpecDetail is the response for /api/workflows/{id}/specs/{type}:
// the authoring template and LLM-as-a-Judge rubric for one spec type of a
// workflow, served from the specification-workflow-spec catalog.
type WorkflowSpecDetail struct {
	WorkflowID string `json:"workflowId"`
	SpecType   string `json:"specType"`
	// Template is the raw markdown authoring template (placeholders and
	// guidance comments included); empty if the workflow defines none.
	Template string `json:"template,omitempty"`
	// RubricJSON is the structured-evaluation RubricSet serialized as JSON;
	// empty if the workflow defines no rubric for this spec type.
	RubricJSON string `json:"rubricJson,omitempty"`
}

// CreateInitiativeRequest is the request body for POST /api/initiatives.
type CreateInitiativeRequest struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority,omitempty"`
	InitType    string `json:"initType,omitempty"`
	WorkflowID  string `json:"workflowId"`
	ProgramID   string `json:"programId,omitempty"`
	HomeRepo    string `json:"homeRepo,omitempty"`
}

// CreateInitiativeResponse is the response for POST /api/initiatives.
type CreateInitiativeResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Initiative represents a cross-repository initiative.
type Initiative struct {
	ID                 string            `json:"id"`
	Organization       string            `json:"organization"`
	Title              string            `json:"title"`
	Description        string            `json:"description,omitempty"`
	Status             string            `json:"status"`
	InitType           string            `json:"initType,omitempty"`
	WorkflowID         string            `json:"workflowId,omitempty"`
	Priority           string            `json:"priority,omitempty"`
	HomeRepo           string            `json:"homeRepo,omitempty"`
	Workspace          string            `json:"workspace,omitempty"`
	ProgramID          string            `json:"programId,omitempty"`
	Specs              map[string]string `json:"specs,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	PlannedAt          *time.Time        `json:"plannedAt,omitempty"`
	ExecutingAt        *time.Time        `json:"executingAt,omitempty"`
	DeliveryCompleteAt *time.Time        `json:"deliveryCompleteAt,omitempty"`
	ReleasedAt         *time.Time        `json:"releasedAt,omitempty"`
	ClosedAt           *time.Time        `json:"closedAt,omitempty"`
	UpdatedAt          time.Time         `json:"updatedAt"`
}

// Phase is a themed grouping of RMIs within an initiative.
type Phase struct {
	ID             string `json:"id"`
	InitiativeID   string `json:"initiativeId"`
	SequenceNumber int    `json:"sequenceNumber"`
	Title          string `json:"title"`
	Theme          string `json:"theme,omitempty"`
}

// RoadmapItem (RMI) is a deliverable within a single repository.
type RoadmapItem struct {
	ID                 string       `json:"id"`
	RepositoryID       string       `json:"repositoryId"`
	InitiativeID       string       `json:"initiativeId,omitempty"`
	PhaseID            string       `json:"phaseId,omitempty"`
	Title              string       `json:"title"`
	Description        string       `json:"description,omitempty"`
	ItemType           string       `json:"itemType"`
	Status             string       `json:"status"`
	Priority           string       `json:"priority,omitempty"`
	Required           bool         `json:"required"`
	SequenceNumber     int          `json:"sequenceNumber"`
	AcceptanceCriteria []string     `json:"acceptanceCriteria,omitempty"`
	ContextSpec        *ContextSpec `json:"contextSpec,omitempty"`
	CreatedAt          time.Time    `json:"createdAt"`
	CompletedAt        *time.Time   `json:"completedAt,omitempty"`
	UpdatedAt          time.Time    `json:"updatedAt"`
}

// ContextSpec contains explicit overrides for context assembly.
type ContextSpec struct {
	ExtraRepos   []string `json:"extraRepos,omitempty"`
	IncludeSpecs []string `json:"includeSpecs,omitempty"`
	ExcludeSpecs []string `json:"excludeSpecs,omitempty"`
}

// RMIDependency is a directed edge between two RMIs.
type RMIDependency struct {
	SourceRMIID  string `json:"sourceRmiId"`
	TargetRMIID  string `json:"targetRmiId"`
	Relationship string `json:"relationship"`
}

// Program is an organizational grouping of related initiatives.
type Program struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Organization string    `json:"organization"`
	Description  string    `json:"description,omitempty"`
	Hidden       bool      `json:"hidden,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// APIProgram is the API representation of a program with nested initiatives.
type APIProgram struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Organization string          `json:"organization,omitempty"`
	Description  string          `json:"description,omitempty"`
	Hidden       bool            `json:"hidden,omitempty"`
	Initiatives  []APIInitiative `json:"initiatives,omitempty"`
}

// APIInitiative is the API representation of an initiative with progress.
type APIInitiative struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Status      string  `json:"status"`
	Type        string  `json:"type,omitempty"`
	Priority    string  `json:"priority,omitempty"`
	HomeRepo    string  `json:"homeRepo,omitempty"`
	WorkflowID  string  `json:"workflowId,omitempty"`
	ProgramID   string  `json:"programId,omitempty"`
	ProgramName string  `json:"programName,omitempty"`
	Hidden      bool    `json:"hidden,omitempty"`
	Progress    float64 `json:"progress"`
	// CancelledProgress is the fraction of RMIs that are cancelled (not
	// completed, not pending) -- rendered as a distinct segment so a
	// cancelled RMI doesn't visually read the same as one that's simply
	// not started yet.
	CancelledProgress float64 `json:"cancelledProgress,omitempty"`
}

// APIPhase is the API representation of a phase.
type APIPhase struct {
	ID             string  `json:"id"`
	InitiativeID   string  `json:"initiativeId"`
	SequenceNumber int     `json:"sequenceNumber"`
	Title          string  `json:"title"`
	Theme          string  `json:"theme,omitempty"`
	Progress       float64 `json:"progress"`
	// CancelledProgress mirrors APIInitiative.CancelledProgress at phase scope.
	CancelledProgress float64 `json:"cancelledProgress,omitempty"`
}

// APIRMI is the API representation of a roadmap item.
type APIRMI struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repositoryId,omitempty"`
	InitiativeID string `json:"initiativeId,omitempty"`
	PhaseID      string `json:"phaseId,omitempty"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Type         string `json:"type,omitempty"`
	Status       string `json:"status"`
	Priority     string `json:"priority,omitempty"`
	// Origin records how the RMI's scope was identified: spec,
	// implementation, acceptance_testing, or discussion (see pkg/rmi).
	Origin         string  `json:"origin,omitempty"`
	SequenceNumber int     `json:"sequenceNumber"`
	ClaimedBy      string  `json:"claimedBy,omitempty"`
	ClaimedAt      string  `json:"claimedAt,omitempty"`
	CompletedAt    string  `json:"completedAt,omitempty"`
	TokensTotal    int64   `json:"tokensTotal,omitempty"`
	CostUSD        float64 `json:"costUsd,omitempty"`
}

// APIRMIDependency is the API representation of an RMI dependency.
type APIRMIDependency struct {
	SourceRMIID  string `json:"sourceRmiId"`
	TargetRMIID  string `json:"targetRmiId"`
	Relationship string `json:"relationship"`
}

// APIInitiativeDependency is the API representation of an initiative dependency.
type APIInitiativeDependency struct {
	SourceInitiativeID string `json:"sourceInitiativeId"`
	TargetInitiativeID string `json:"targetInitiativeId"`
	Relationship       string `json:"relationship"`
}

// APIRepository is the API representation of a repository.
type APIRepository struct {
	ID             string `json:"id"`
	Organization   string `json:"organization"`
	RepositoryName string `json:"repositoryName"`
	DefaultBranch  string `json:"defaultBranch,omitempty"`
	LocalPath      string `json:"localPath,omitempty"`
	GoModule       string `json:"goModule,omitempty"`
	Domain         string `json:"domain,omitempty"`
	Status         string `json:"status"`
}

// ExecutionResponse is the response for /api/execution.
type ExecutionResponse struct {
	Programs               []APIProgram              `json:"programs"`
	Initiatives            []APIInitiative           `json:"initiatives"`
	Phases                 []APIPhase                `json:"phases"`
	RMIs                   []APIRMI                  `json:"rmis"`
	Repositories           []APIRepository           `json:"repositories"`
	StatusDistribution     []APIStatusCount          `json:"statusDistribution,omitempty"`
	RMIDependencies        []APIRMIDependency        `json:"rmiDependencies,omitempty"`
	InitiativeDependencies []APIInitiativeDependency `json:"initiativeDependencies,omitempty"`
	Releases               []APIRelease              `json:"releases,omitempty"`
}

// APIStatusCount is a status with its count.
type APIStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// APIRelease is a recorded repo@tag release, with the initiatives it's
// attached to -- an initiative can be attached to releases from multiple
// repos (its RMIs span more than one), so "what did this initiative ship
// in" is a list, not a single tag.
type APIRelease struct {
	ID            string   `json:"id"`
	RepositoryID  string   `json:"repositoryId"`
	Tag           string   `json:"tag"`
	ReleasedAt    string   `json:"releasedAt"`
	URL           string   `json:"url,omitempty"`
	InitiativeIDs []string `json:"initiativeIds,omitempty"`
}

// SpecFile represents a spec document read from disk.
type SpecFile struct {
	InitiativeID string `json:"initiativeId"`
	SpecType     string `json:"specType"`
	Path         string `json:"path"`
	Content      string `json:"content"`
	ModTime      string `json:"modTime,omitempty"`
	EvalJSON     string `json:"evalJson,omitempty"`
	// Role classifies the file against the initiative's workflow:
	// "required", "optional", or "extra" (present on disk but not part of
	// the selected workflow's spec set).
	Role string `json:"role,omitempty"`
}

// SpecFilesResponse is the response for /api/spec-files/{initiativeId}.
type SpecFilesResponse struct {
	Files []SpecFile `json:"files"`
}
