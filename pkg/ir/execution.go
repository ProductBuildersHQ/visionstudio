// Package ir defines the intermediate representation types for VisionStudio.
// Types are imported from their source modules to ensure a single source of
// truth and compile-time drift detection. See TRD T1.
package ir

import (
	vsstore "github.com/ProductBuildersHQ/visionstudio/pkg/store"
	prismmaturity "github.com/grokify/prism-maturity"
	"github.com/grokify/prism-roadmap/goals"
	"github.com/grokify/prism-roadmap/goals/okr"
	"github.com/grokify/prism-roadmap/roadmap"
	"github.com/plexusone/devfolio/contributor"
	"github.com/plexusone/devfolio/output/devxdashboard"
)

// Execution domain types — aliases to this module's pkg/store.
// These represent the core execution tracking entities.

type (
	Program              = vsstore.Program
	Initiative           = vsstore.Initiative
	Phase                = vsstore.Phase
	RoadmapItem          = vsstore.RoadmapItem
	Assignment           = vsstore.Assignment
	DeliveryEvidence     = vsstore.DeliveryEvidence
	Repository           = vsstore.Repository
	ContextSpec          = vsstore.ContextSpec
	Handoff              = vsstore.Handoff
	RMIDependency        = vsstore.RMIDependency
	InitiativeDependency = vsstore.InitiativeDependency
	RepositoryDependency = vsstore.RepositoryDependency
)

// Spec workflow and judging types.

type (
	SpecWorkflow = vsstore.SpecWorkflow
	JudgeResult  = vsstore.JudgeResult
)

// Maturity model types from this module's pkg/store (Dolt-backed).

type (
	CapabilityModel    = vsstore.CapabilityModel
	MaturityAssessment = vsstore.MaturityAssessment
	Dimension          = vsstore.Dimension
	Level              = vsstore.Level
	DimensionScore     = vsstore.DimensionScore
)

// PRISM maturity framework types from prism-maturity (JSON IR).

type (
	PRISMDocument = prismmaturity.PRISMDocument
	PRISMMetric   = prismmaturity.Metric
	SLI           = prismmaturity.SLI
	SLO           = prismmaturity.SLO
	PRISMService  = prismmaturity.Service
	PRISMTeam     = prismmaturity.Team
)

// Roadmap types from prism-roadmap (JSON IR).

type (
	Roadmap           = roadmap.Roadmap
	RoadmapPhase      = roadmap.Phase
	Deliverable       = roadmap.Deliverable
	DeliverableStatus = roadmap.DeliverableStatus
	PhaseStatus       = roadmap.PhaseStatus
	RoadmapRisk       = roadmap.Risk
)

// Goals types from prism-roadmap (JSON IR).

type (
	Goals       = goals.Goals
	GoalItem    = goals.GoalItem
	ResultItem  = goals.ResultItem
	OKRDocument = okr.OKRDocument
	Objective   = okr.Objective
	KeyResult   = okr.KeyResult
)

// Contributor/devfolio types (JSON IR).

type (
	ContributorProfile = contributor.Profile
	RepoContrib        = contributor.RepoContrib
	ContributorStats   = contributor.ContributorStats
	DailyActivity      = contributor.DailyActivity
)

// DevX dashboard period report types (JSON IR).

type (
	PeriodReport     = devxdashboard.PeriodReport
	PeriodType       = devxdashboard.PeriodType
	DailyPoint       = devxdashboard.DailyPoint
	ModelPeriodPoint = devxdashboard.ModelPeriodPoint
)

// Phase 5 store types — aliases to visionstudio/pkg/store for the new domains.

type (
	DevXPeriodReport = vsstore.DevXPeriodReport
	PRISMRoadmapDB   = vsstore.PRISMRoadmap
	PRISMGoalDB      = vsstore.PRISMGoal
	PRISMDocumentDB  = vsstore.PRISMDocument
	SpecDocumentDB   = vsstore.SpecDocument
)
