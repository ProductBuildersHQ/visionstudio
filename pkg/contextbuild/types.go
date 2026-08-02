// Package contextbuild assembles deterministic context packages for agent sessions.
// A context package contains all information needed to work on a phase or RMI,
// ordered stable→volatile with provenance revisions. Go structs are the source
// of truth; JSON Schema is generated via invopop/jsonschema.
//
// Design: TRD §15; origin: IDEATION_CHAT_CACHE-OPTIMIZATION.md.
package contextbuild

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// StabilityClass indicates how frequently a section's content changes.
// Sections are ordered from most stable to most volatile.
type StabilityClass string

const (
	// Stable content rarely changes (program/initiative definitions, decisions).
	Stable StabilityClass = "stable"

	// PhaseStable content changes at phase boundaries (phase objectives, member RMIs).
	PhaseStable StabilityClass = "phase_stable"

	// RMIStable content changes per RMI (acceptance criteria, dependencies, repo).
	RMIStable StabilityClass = "rmi_stable"

	// Volatile content changes frequently (assignment state, evidence).
	Volatile StabilityClass = "volatile"
)

// Provenance tracks the source revision for a section.
type Provenance struct {
	// Source is the data source type: "dolt" for graph data, "git" for spec files.
	Source string `json:"source"`

	// Revision is the Dolt commit hash or git SHA.
	Revision string `json:"revision"`

	// Timestamp is when the revision was created.
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// ContextPackage is the complete assembled context for a phase or RMI.
// All sections are ordered stable→volatile for optimal cache utilization.
type ContextPackage struct {
	// Version is the schema version for forwards compatibility.
	Version string `json:"version"`

	// TargetType is "phase" or "rmi".
	TargetType string `json:"target_type"`

	// TargetID is the phase ID or RMI ID this package is built for.
	TargetID string `json:"target_id"`

	// BuildTimestamp is when this package was assembled (operational metadata, last).
	BuildTimestamp time.Time `json:"build_timestamp"`

	// Sections contains the context data ordered by stability class.
	Sections Sections `json:"sections"`

	// DerivedRepos is the computed set of repositories relevant to this context.
	DerivedRepos []DerivedRepo `json:"derived_repos"`
}

// Sections groups context data by stability class, ordered stable→volatile.
type Sections struct {
	// Program contains program-level context (stability: stable).
	Program *ProgramSection `json:"program,omitempty"`

	// Initiative contains initiative-level context (stability: stable).
	Initiative InitiativeSection `json:"initiative"`

	// Phase contains phase-level context (stability: phase_stable).
	Phase PhaseSection `json:"phase"`

	// PrerequisiteHandoffs contains handoffs from prior phases (stability: phase_stable).
	PrerequisiteHandoffs []PhaseHandoff `json:"prerequisite_handoffs,omitempty"`

	// CurrentRMI contains the target RMI context (stability: rmi_stable).
	// Only present when TargetType is "rmi".
	CurrentRMI *RMISection `json:"current_rmi,omitempty"`

	// SpecReferences contains paths to spec files with revisions (stability: rmi_stable).
	SpecReferences []SpecReference `json:"spec_references,omitempty"`

	// Assignment contains current assignment state (stability: volatile).
	Assignment *AssignmentSection `json:"assignment,omitempty"`
}

// ProgramSection contains program-level context.
type ProgramSection struct {
	Stability  StabilityClass `json:"stability"`
	Provenance Provenance     `json:"provenance"`

	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// InitiativeSection contains initiative-level context.
type InitiativeSection struct {
	Stability  StabilityClass `json:"stability"`
	Provenance Provenance     `json:"provenance"`

	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority,omitempty"`
	HomeRepo    string            `json:"home_repo,omitempty"`
	Specs       map[string]string `json:"specs,omitempty"`
	Decisions   []string          `json:"decisions,omitempty"`
}

// PhaseSection contains phase-level context.
type PhaseSection struct {
	Stability  StabilityClass `json:"stability"`
	Provenance Provenance     `json:"provenance"`

	ID             string `json:"id"`
	Title          string `json:"title"`
	Theme          string `json:"theme,omitempty"`
	SequenceNumber int    `json:"sequence_number"`
	DerivedStatus  string `json:"derived_status"`

	// MemberRMIs lists all RMIs in this phase with summary info.
	MemberRMIs []RMISummary `json:"member_rmis"`

	// DependencyEdges lists inter-RMI dependencies within this phase.
	DependencyEdges []DependencyEdge `json:"dependency_edges,omitempty"`
}

// RMISummary is a compact representation of an RMI for phase context.
type RMISummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Required bool   `json:"required"`
	Sequence int    `json:"sequence"`
	RepoID   string `json:"repo_id"`
}

// DependencyEdge represents a directed dependency between two RMIs.
type DependencyEdge struct {
	SourceID     string `json:"source_id"`
	TargetID     string `json:"target_id"`
	Relationship string `json:"relationship"`
}

// PhaseHandoff contains handoff data from a completed prior phase.
type PhaseHandoff struct {
	Stability  StabilityClass `json:"stability"`
	Provenance Provenance     `json:"provenance"`

	PhaseID       string   `json:"phase_id"`
	PhaseTitle    string   `json:"phase_title"`
	DerivedStatus string   `json:"derived_status"`
	Completed     []string `json:"completed,omitempty"`
	Decisions     []string `json:"decisions,omitempty"`
}

// RMISection contains detailed RMI context for the target RMI.
type RMISection struct {
	Stability  StabilityClass `json:"stability"`
	Provenance Provenance     `json:"provenance"`

	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description,omitempty"`
	ItemType           string   `json:"item_type"`
	Status             string   `json:"status"`
	Priority           string   `json:"priority,omitempty"`
	Required           bool     `json:"required"`
	SequenceNumber     int      `json:"sequence_number"`
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
	RepositoryID       string   `json:"repository_id"`

	// Dependencies lists RMIs this one depends on.
	Dependencies []DependencyEdge `json:"dependencies,omitempty"`

	// Dependents lists RMIs that depend on this one.
	Dependents []DependencyEdge `json:"dependents,omitempty"`
}

// SpecReference points to a spec file with its git revision.
type SpecReference struct {
	Stability  StabilityClass `json:"stability"`
	Provenance Provenance     `json:"provenance"`

	// Path is the file path relative to the repository root.
	Path string `json:"path"`

	// RepoID is the repository containing this spec.
	RepoID string `json:"repo_id"`

	// LocalPath is the absolute path on disk (for agent access).
	LocalPath string `json:"local_path,omitempty"`
}

// AssignmentSection contains current assignment state.
type AssignmentSection struct {
	Stability  StabilityClass `json:"stability"`
	Provenance Provenance     `json:"provenance"`

	ID             string     `json:"id"`
	RMIID          string     `json:"rmi_id"`
	Worker         string     `json:"worker"`
	Status         string     `json:"status"`
	LeaseExpiresAt time.Time  `json:"lease_expires_at"`
	Workspace      string     `json:"workspace,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`

	// Handoff contains structured handoff state from prior sessions.
	Handoff *HandoffData `json:"handoff,omitempty"`

	// Evidence lists delivery evidence collected so far.
	Evidence []EvidenceRef `json:"evidence,omitempty"`
}

// HandoffData carries compact state for session continuity.
type HandoffData struct {
	Completed  []string `json:"completed,omitempty"`
	Remaining  []string `json:"remaining,omitempty"`
	Decisions  []string `json:"decisions,omitempty"`
	NextAction string   `json:"next_action,omitempty"`
}

// EvidenceRef is a compact reference to delivery evidence.
type EvidenceRef struct {
	Type       string     `json:"type"`
	Reference  string     `json:"reference"`
	OccurredAt *time.Time `json:"occurred_at,omitempty"`
}

// DerivedRepo is a repository in the computed context set.
type DerivedRepo struct {
	// ID is the repository ID (e.g., github.com/org/repo).
	ID string `json:"id"`

	// Role indicates why this repo is included:
	// "primary", "dependency_rmi", "repo_dependency", "explicit".
	Role string `json:"role"`

	// LocalPath is the absolute path on disk.
	LocalPath string `json:"local_path,omitempty"`

	// DefaultBranch is the repository's default branch.
	DefaultBranch string `json:"default_branch"`

	// SourceRMI is set when Role is "dependency_rmi" — the RMI that brought this repo in.
	SourceRMI string `json:"source_rmi,omitempty"`

	// SourceRepo is set when Role is "repo_dependency" — the repo that depends on this one.
	SourceRepo string `json:"source_repo,omitempty"`
}

// SchemaVersion is the current version of the context package schema.
const SchemaVersion = "1.0.0"

// RenderJSON renders the context package as indented JSON.
func (p *ContextPackage) RenderJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// RenderMarkdown renders the context package as Markdown.
func (p *ContextPackage) RenderMarkdown() string {
	var buf bytes.Buffer

	if p.TargetType == "phase" {
		fmt.Fprintf(&buf, "# Context Package: Phase %s\n\n", p.Sections.Phase.Title)
	} else {
		fmt.Fprintf(&buf, "# Context Package: RMI %s\n\n", p.TargetID)
	}

	fmt.Fprintf(&buf, "**Version:** %s\n", p.Version)
	fmt.Fprintf(&buf, "**Target:** %s (`%s`)\n", p.TargetType, p.TargetID)
	fmt.Fprintf(&buf, "**Generated:** %s\n\n", p.BuildTimestamp.Format(time.RFC3339))

	if p.Sections.Program != nil {
		fmt.Fprintf(&buf, "## Program\n\n")
		fmt.Fprintf(&buf, "- **ID:** `%s`\n", p.Sections.Program.ID)
		fmt.Fprintf(&buf, "- **Name:** %s\n", p.Sections.Program.Name)
		if p.Sections.Program.Description != "" {
			fmt.Fprintf(&buf, "- **Description:** %s\n", p.Sections.Program.Description)
		}
		fmt.Fprintf(&buf, "\n")
	}

	fmt.Fprintf(&buf, "## Initiative\n\n")
	fmt.Fprintf(&buf, "- **ID:** `%s`\n", p.Sections.Initiative.ID)
	fmt.Fprintf(&buf, "- **Title:** %s\n", p.Sections.Initiative.Title)
	fmt.Fprintf(&buf, "- **Status:** %s\n", p.Sections.Initiative.Status)
	if p.Sections.Initiative.Priority != "" {
		fmt.Fprintf(&buf, "- **Priority:** %s\n", p.Sections.Initiative.Priority)
	}
	fmt.Fprintf(&buf, "\n")

	fmt.Fprintf(&buf, "## Phase\n\n")
	fmt.Fprintf(&buf, "- **ID:** `%s`\n", p.Sections.Phase.ID)
	fmt.Fprintf(&buf, "- **Title:** %s\n", p.Sections.Phase.Title)
	if p.Sections.Phase.Theme != "" {
		fmt.Fprintf(&buf, "- **Theme:** %s\n", p.Sections.Phase.Theme)
	}
	fmt.Fprintf(&buf, "- **Status:** %s\n", p.Sections.Phase.DerivedStatus)
	fmt.Fprintf(&buf, "\n")

	if len(p.Sections.Phase.MemberRMIs) > 0 {
		fmt.Fprintf(&buf, "### Member RMIs\n\n")
		fmt.Fprintf(&buf, "| # | ID | Title | Status | Required |\n")
		fmt.Fprintf(&buf, "|---|---|-------|--------|----------|\n")
		for _, rmi := range p.Sections.Phase.MemberRMIs {
			req := "no"
			if rmi.Required {
				req = "yes"
			}
			fmt.Fprintf(&buf, "| %d | `%s` | %s | %s | %s |\n",
				rmi.Sequence, rmi.ID, rmi.Title, rmi.Status, req)
		}
		fmt.Fprintf(&buf, "\n")
	}

	if p.Sections.CurrentRMI != nil {
		fmt.Fprintf(&buf, "## Current RMI\n\n")
		fmt.Fprintf(&buf, "- **ID:** `%s`\n", p.Sections.CurrentRMI.ID)
		fmt.Fprintf(&buf, "- **Title:** %s\n", p.Sections.CurrentRMI.Title)
		fmt.Fprintf(&buf, "- **Type:** %s\n", p.Sections.CurrentRMI.ItemType)
		fmt.Fprintf(&buf, "- **Status:** %s\n", p.Sections.CurrentRMI.Status)
		fmt.Fprintf(&buf, "- **Repository:** `%s`\n", p.Sections.CurrentRMI.RepositoryID)
		fmt.Fprintf(&buf, "\n")

		if len(p.Sections.CurrentRMI.AcceptanceCriteria) > 0 {
			fmt.Fprintf(&buf, "### Acceptance Criteria\n\n")
			for _, c := range p.Sections.CurrentRMI.AcceptanceCriteria {
				fmt.Fprintf(&buf, "- %s\n", c)
			}
			fmt.Fprintf(&buf, "\n")
		}
	}

	if len(p.Sections.SpecReferences) > 0 {
		fmt.Fprintf(&buf, "## Spec References\n\n")
		fmt.Fprintf(&buf, "| Path | Repository | Revision |\n")
		fmt.Fprintf(&buf, "|------|------------|----------|\n")
		for _, ref := range p.Sections.SpecReferences {
			rev := ref.Provenance.Revision
			if len(rev) > 8 {
				rev = rev[:8]
			}
			fmt.Fprintf(&buf, "| `%s` | `%s` | `%s` |\n", ref.Path, ref.RepoID, rev)
		}
		fmt.Fprintf(&buf, "\n")
	}

	if len(p.DerivedRepos) > 0 {
		fmt.Fprintf(&buf, "## Derived Repositories\n\n")
		fmt.Fprintf(&buf, "| Repository | Role | Branch |\n")
		fmt.Fprintf(&buf, "|------------|------|--------|\n")
		for _, repo := range p.DerivedRepos {
			fmt.Fprintf(&buf, "| `%s` | %s | %s |\n", repo.ID, repo.Role, repo.DefaultBranch)
		}
		fmt.Fprintf(&buf, "\n")
	}

	return buf.String()
}
