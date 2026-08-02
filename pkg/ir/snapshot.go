package ir

import (
	prismmaturity "github.com/grokify/prism-maturity"
	"github.com/grokify/prism-roadmap/roadmap"
	"github.com/plexusone/devfolio/contributor"
	"github.com/plexusone/devfolio/output/devxdashboard"
	"github.com/plexusone/uiforge/dashboardir"
)

// RepoSnapshot composes all domain IRs for one repository.
// This is the unit of import/export for visionstudio ingest.
type RepoSnapshot struct {
	Repo      string        `json:"repo"`
	Org       string        `json:"org,omitempty"`
	Execution *ExecutionIR  `json:"execution,omitempty"`
	Maturity  *MaturityIR   `json:"maturity,omitempty"`
	Roadmap   *RoadmapIR    `json:"roadmap,omitempty"`
	DevX      *DevXIR       `json:"devx,omitempty"`
	Contrib   *ContribIR    `json:"contrib,omitempty"`
	Timestamp string        `json:"timestamp,omitempty"`
}

// ExecutionIR holds execution-tracking data for a repository.
type ExecutionIR struct {
	Initiatives []*Initiative    `json:"initiatives,omitempty"`
	Phases      []*Phase         `json:"phases,omitempty"`
	RMIs        []*RoadmapItem   `json:"rmis,omitempty"`
	Assignments []*Assignment    `json:"assignments,omitempty"`
	Evidence    []*DeliveryEvidence `json:"evidence,omitempty"`
}

// MaturityIR holds maturity assessment data.
// Supports both capability models (Dolt-backed) and PRISM documents (JSON IR).
type MaturityIR struct {
	CapabilityModels []*CapabilityModel     `json:"capabilityModels,omitempty"`
	Assessments      []*MaturityAssessment  `json:"assessments,omitempty"`
	PRISMDocument    *prismmaturity.PRISMDocument `json:"prismDocument,omitempty"`
}

// RoadmapIR holds roadmap and goals data.
type RoadmapIR struct {
	Roadmap *roadmap.Roadmap `json:"roadmap,omitempty"`
	Goals   *Goals           `json:"goals,omitempty"`
}

// DevXIR holds developer experience metrics and dashboards.
type DevXIR struct {
	PeriodReports []*devxdashboard.PeriodReport `json:"periodReports,omitempty"`
	Dashboard     *dashboardir.Dashboard        `json:"dashboard,omitempty"`
}

// ContribIR holds contributor profile and activity data.
type ContribIR struct {
	Profiles []*contributor.Profile `json:"profiles,omitempty"`
}
