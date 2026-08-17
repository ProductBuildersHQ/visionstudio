// AUTO-GENERATED - DO NOT EDIT
// Generated from JSON Schema via: npm run generate:types
// Source: Go structs in pkg/apitypes/types.go

import { z } from 'zod'
import {
  RubricSchema,
  JudgeMetadataSchema,
  CategoryResultSchema,
  FindingSchema,
  ActionItemSchema,
  NextStepsSchema,
  DecisionSchema,
  ReportMetadataSchema,
  PassCriteriaSchema,
  ReferenceDataSchema,
  ChecklistResultsSchema,
  JudgeResultSchema,
  SpecWorkflowSchema,
  SpecFileSchema,
  SpecFilesResponseSchema,
  SpecsResponseSchema,
  ContextSpecSchema,
  PhaseSchema,
  ProgramSchema,
  InitiativeSchema,
  RoadmapItemSchema,
  APIRMISchema,
  APIRepositorySchema,
  APIRMIDependencySchema,
  APIInitiativeDependencySchema,
  APIStatusCountSchema,
  APIReleaseSchema,
  APIPhaseSchema,
  APIInitiativeSchema,
  APIProgramSchema,
  CreateInitiativeRequestSchema,
  CreateInitiativeResponseSchema,
  WorkflowSpecDetailSchema,
  ExecutionResponseSchema,
} from './schemas.gen'

export type Rubric = z.infer<typeof RubricSchema>
export type JudgeMetadata = z.infer<typeof JudgeMetadataSchema>
export type CategoryResult = z.infer<typeof CategoryResultSchema>
export type Finding = z.infer<typeof FindingSchema>
export type ActionItem = z.infer<typeof ActionItemSchema>
export type NextSteps = z.infer<typeof NextStepsSchema>
export type Decision = z.infer<typeof DecisionSchema>
export type ReportMetadata = z.infer<typeof ReportMetadataSchema>
export type PassCriteria = z.infer<typeof PassCriteriaSchema>
export type ReferenceData = z.infer<typeof ReferenceDataSchema>
export type ChecklistResults = z.infer<typeof ChecklistResultsSchema>
export type JudgeResult = z.infer<typeof JudgeResultSchema>
export type SpecWorkflow = z.infer<typeof SpecWorkflowSchema>
export type SpecFile = z.infer<typeof SpecFileSchema>
export type SpecFilesResponse = z.infer<typeof SpecFilesResponseSchema>
export type SpecsResponse = z.infer<typeof SpecsResponseSchema>
export type ContextSpec = z.infer<typeof ContextSpecSchema>
export type Phase = z.infer<typeof PhaseSchema>
export type Program = z.infer<typeof ProgramSchema>
export type Initiative = z.infer<typeof InitiativeSchema>
export type RoadmapItem = z.infer<typeof RoadmapItemSchema>
export type APIRMI = z.infer<typeof APIRMISchema>
export type APIRepository = z.infer<typeof APIRepositorySchema>
export type APIRMIDependency = z.infer<typeof APIRMIDependencySchema>
export type APIInitiativeDependency = z.infer<typeof APIInitiativeDependencySchema>
export type APIStatusCount = z.infer<typeof APIStatusCountSchema>
export type APIRelease = z.infer<typeof APIReleaseSchema>
export type APIPhase = z.infer<typeof APIPhaseSchema>
export type APIInitiative = z.infer<typeof APIInitiativeSchema>
export type APIProgram = z.infer<typeof APIProgramSchema>
export type CreateInitiativeRequest = z.infer<typeof CreateInitiativeRequestSchema>
export type CreateInitiativeResponse = z.infer<typeof CreateInitiativeResponseSchema>
export type WorkflowSpecDetail = z.infer<typeof WorkflowSpecDetailSchema>
export type ExecutionResponse = z.infer<typeof ExecutionResponseSchema>

// Re-export schemas for runtime validation
export {
  RubricSchema,
  JudgeMetadataSchema,
  CategoryResultSchema,
  FindingSchema,
  ActionItemSchema,
  NextStepsSchema,
  DecisionSchema,
  ReportMetadataSchema,
  PassCriteriaSchema,
  ReferenceDataSchema,
  ChecklistResultsSchema,
  JudgeResultSchema,
  SpecWorkflowSchema,
  SpecFileSchema,
  SpecFilesResponseSchema,
  SpecsResponseSchema,
  ContextSpecSchema,
  PhaseSchema,
  ProgramSchema,
  InitiativeSchema,
  RoadmapItemSchema,
  APIRMISchema,
  APIRepositorySchema,
  APIRMIDependencySchema,
  APIInitiativeDependencySchema,
  APIStatusCountSchema,
  APIReleaseSchema,
  APIPhaseSchema,
  APIInitiativeSchema,
  APIProgramSchema,
  CreateInitiativeRequestSchema,
  CreateInitiativeResponseSchema,
  WorkflowSpecDetailSchema,
  ExecutionResponseSchema,
}
