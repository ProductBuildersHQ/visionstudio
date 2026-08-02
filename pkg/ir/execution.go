// Package ir defines the intermediate representation types for VisionStudio.
// Types are imported from their source modules to ensure a single source of
// truth and compile-time drift detection. See TRD T1.
package ir

import (
	pbstore "github.com/ProductBuildersHQ/prism-build/pkg/store"
	vsstore "github.com/ProductBuildersHQ/visionstudio/pkg/store"
	prismmaturity "github.com/grokify/prism-maturity"
	"github.com/grokify/prism-roadmap/goals"
	"github.com/grokify/prism-roadmap/goals/okr"
	"github.com/grokify/prism-roadmap/roadmap"
	"github.com/plexusone/devfolio/contributor"
	"github.com/plexusone/devfolio/output/devxdashboard"
)

// Execution domain types — aliases to prism-build/pkg/store.
// These represent the core execution tracking entities.

type (
	Program              = pbstore.Program
	Initiative           = pbstore.Initiative
	Phase                = pbstore.Phase
	RoadmapItem          = pbstore.RoadmapItem
	Assignment           = pbstore.Assignment
	DeliveryEvidence     = pbstore.DeliveryEvidence
	Repository           = pbstore.Repository
	ContextSpec          = pbstore.ContextSpec
	Handoff              = pbstore.Handoff
	RMIDependency        = pbstore.RMIDependency
	InitiativeDependency = pbstore.InitiativeDependency
	RepositoryDependency = pbstore.RepositoryDependency
)

// Spec workflow and judging types.

type (
	SpecWorkflow = pbstore.SpecWorkflow
	JudgeRubric  = pbstore.JudgeRubric
	JudgeResult  = pbstore.JudgeResult
)

// Maturity model types from prism-build (Dolt-backed).

type (
	CapabilityModel    = pbstore.CapabilityModel
	MaturityAssessment = pbstore.MaturityAssessment
	Dimension          = pbstore.Dimension
	Level              = pbstore.Level
	DimensionScore     = pbstore.DimensionScore
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
