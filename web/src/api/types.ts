// API types for VisionStudio frontend
//
// Type pipeline: Go structs → JSON Schema → Zod → TypeScript
// See: pkg/apitypes/types.go (source of truth)
//
// To regenerate: go generate ./pkg/apitypes && cd web && npm run generate:types

// Re-export types from compat (normalized with default arrays)
export type {
  JudgeResult,
  SpecWorkflow,
  SpecWorkflowPhase,
  SpecFileRole,
  SpecsResponse,
  ExecutionResponse,
  APIInitiative,
  APIPhase,
  APIRMI,
  APIRepository,
  APIRMIDependency,
  APIInitiativeDependency,
  APIProgram,
  SpecFile,
  SpecFilesResponse,
} from './compat'

// Re-export helper
export { getScore } from './compat'

// Re-export generated schemas for runtime validation
export {
  JudgeResultSchema,
  SpecWorkflowSchema,
  SpecsResponseSchema,
  ExecutionResponseSchema,
  APIInitiativeSchema,
  APIPhaseSchema,
  APIRMISchema,
  APIRepositorySchema,
  APIRMIDependencySchema,
  APIProgramSchema,
  SpecFileSchema,
  SpecFilesResponseSchema,
} from './types.gen'

// Additional types not yet in Go apitypes

export interface APITokens {
  inputTokens: number
  outputTokens: number
  cacheReadTokens: number
  cacheCreationTokens: number
  totalTokens: number
  costUsd: number
}

export interface APIStatusCount {
  status: string
  count: number
}

export interface APITimeBucket {
  period: string
  start: string
  end: string
  totals: APITokens
  byModel?: Record<string, APITokens>
}

export interface SpendResponse {
  total?: APITokens
  byModel?: Record<string, APITokens>
  byInitiative?: Record<string, APITokens>
  byPhase?: Record<string, APITokens>
  byRmi?: Record<string, APITokens>
  byWeek?: APITimeBucket[]
  byMonth?: APITimeBucket[]
  hasData: boolean
  dataNote?: string
}

export interface CapabilityModel {
  id: string
  name: string
  description?: string
  max_level: number
  dimensions?: Dimension[]
}

export interface Dimension {
  key: string
  name: string
  sources?: string[]
  levels?: Level[]
}

export interface Level {
  level: number
  name: string
  description?: string
}

export interface MaturityAssessment {
  id: string
  initiative_id: string
  model_id: string
  assessed_at: string
  scores?: DimensionScore[]
}

export interface DimensionScore {
  dimension_key: string
  level: number
  notes?: string
}

export interface MaturityResponse {
  models: CapabilityModel[]
  assessments: MaturityAssessment[]
}

// SCALE types
export interface ScaleResponse {
  framework?: ScaleFramework
  assessment?: ScaleAssessmentInfo
  rollup?: ScaleRollup
  hasData: boolean
  dataNote?: string
}

export interface ScaleFramework {
  id: string
  name: string
  description?: string
  domains: ScaleDomain[]
}

export interface ScaleDomain {
  id: string
  name: string
  description?: string
  status?: string
  capabilities: ScaleCapability[]
}

export interface ScaleCapability {
  id: string
  name: string
  description?: string
  metrics: ScaleMetric[]
}

export interface ScaleMetric {
  id: string
  name: string
  description?: string
  aspect: string
  consumptionKind?: string
  unit?: string
  direction?: string
  targetValue?: number
  targetBy?: string
  owner?: string
  value?: number
  numerator?: number
  denominator?: number
  attainment?: number
  note?: string
}

export interface ScaleAssessmentInfo {
  period: string
  asOf?: string
  observations: number
  narratives?: ScaleNarrative[]
}

export interface ScaleNarrative {
  id: string
  kind: string
  title?: string
  body: string
}

export interface ScaleRollup {
  aspects: ScaleAspectScore[]
}

export interface ScaleAspectScore {
  aspect: string
  letter: string
  displayName: string
  score: number
  eligible: number
  observed: number
}

// Leverage Graph types
export interface LeverageGraph {
  generatedAt: string
  ecosystem: string
  scope: string
  modules: LeverageModule[]
  edges: LeverageEdge[]
  summary: LeverageSummary
}

export interface LeverageModule {
  id: string
  name: string
  path: string
  kind: 'internal' | 'external' | 'stdlib'
  org?: string
  version?: string
  stats: ModuleStats
  tags?: string[]
}

export interface ModuleStats {
  directDependents: number
  transitiveDependents?: number
  directDependencies: number
  depth?: number
  leverageScore?: number
}

export interface LeverageEdge {
  from: string
  to: string
  kind: 'direct' | 'indirect'
  version?: string
}

export interface LeverageSummary {
  totalModules: number
  internalModules: number
  externalModules: number
  totalEdges: number
  internalEdges: number
  internalRatio: number
  topLeveraged?: ModuleRank[]
  topConsumers?: ModuleRank[]
  orphans?: string[]
  clusters?: ModuleCluster[]
}

export interface ModuleRank {
  moduleId: string
  dependents: number
  direct?: number
  indirect?: number
  score?: number
}

export interface ModuleCluster {
  id: string
  name?: string
  modules: string[]
  reason?: string
}
