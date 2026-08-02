// API response types matching prism-build/cmd/prismctl/api.go

export interface APIProgram {
  id: string
  name: string
  description?: string
  hidden?: boolean
}

export interface APIInitiative {
  id: string
  title: string
  description?: string
  status: string
  type?: string
  programId?: string
  programName?: string
  homeRepo?: string
  progress: number
}

export interface APIPhase {
  id: string
  initiativeId: string
  title: string
  sequenceNumber: number
  progress: number
}

export interface APIRMI {
  id: string
  initiativeId: string
  phaseId: string
  title: string
  status: string
  type?: string
  repositoryId?: string
  sequenceNumber: number
  claimedBy?: string
  claimedAt?: string
  completedAt?: string
  tokensTotal?: number
  costUsd?: number
}

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

export interface ExecutionResponse {
  programs: APIProgram[]
  initiatives: APIInitiative[]
  phases: APIPhase[]
  rmis: APIRMI[]
  statusDistribution: APIStatusCount[]
}

export interface SpendResponse {
  total?: APITokens
  byInitiative?: Record<string, APITokens>
  byPhase?: Record<string, APITokens>
  byRmi?: Record<string, APITokens>
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

export interface SpecWorkflow {
  id: string
  name: string
  description?: string
  specs_required?: string[]
  specs_optional?: string[]
}

export interface JudgeResult {
  id: string
  initiative_id: string
  spec_path: string
  rubric_id?: string
  score: number
  rationale: string
  model?: string
  evaluated_at: string
}

export interface SpecsResponse {
  workflows: SpecWorkflow[]
  judgeResults: JudgeResult[]
}
