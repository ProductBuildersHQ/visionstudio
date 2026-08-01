// Package ir defines the intermediate representation types for VisionStudio.
// Types are imported from their source modules to ensure a single source of
// truth and compile-time drift detection. See TRD T1.
package ir

import "github.com/ProductBuildersHQ/prism-build/pkg/store"

// Execution domain types — aliases to prism-build/pkg/store.
// These represent the core execution tracking entities.

type (
	Program              = store.Program
	Initiative           = store.Initiative
	Phase                = store.Phase
	RoadmapItem          = store.RoadmapItem
	Assignment           = store.Assignment
	DeliveryEvidence     = store.DeliveryEvidence
	Repository           = store.Repository
	ContextSpec          = store.ContextSpec
	Handoff              = store.Handoff
	RMIDependency        = store.RMIDependency
	InitiativeDependency = store.InitiativeDependency
	RepositoryDependency = store.RepositoryDependency
)

// Spec workflow and judging types.

type (
	SpecWorkflow = store.SpecWorkflow
	JudgeRubric  = store.JudgeRubric
	JudgeResult  = store.JudgeResult
)

// Maturity model types.

type (
	CapabilityModel    = store.CapabilityModel
	MaturityAssessment = store.MaturityAssessment
	Dimension          = store.Dimension
	Level              = store.Level
	DimensionScore     = store.DimensionScore
)
