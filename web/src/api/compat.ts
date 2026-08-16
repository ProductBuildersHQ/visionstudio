// Compatibility layer for API types
// Normalizes optional fields from generated types to required fields with defaults
//
// Generated types use rubric.Rubric from structured-evaluation directly.
// This layer provides normalized interfaces with required fields for frontend use.

import type {
  JudgeResult as GenJudgeResult,
  Rubric as GenRubric,
  CategoryResult as GenCategoryResult,
  Finding as GenFinding,
  ActionItem as GenActionItem,
  NextSteps as GenNextSteps,
  Decision as GenDecision,
  SpecsResponse as GenSpecsResponse,
  ExecutionResponse as GenExecutionResponse,
  APIInitiative as GenAPIInitiative,
  APIPhase as GenAPIPhase,
  APIRMI as GenAPIRMI,
  APIRMIDependency as GenAPIRMIDependency,
  APIProgram as GenAPIProgram,
  APIRepository as GenAPIRepository,
  SpecFile as GenSpecFile,
  SpecFilesResponse as GenSpecFilesResponse,
  SpecWorkflow as GenSpecWorkflow,
  APIInitiativeDependency as GenAPIInitiativeDependency,
  APIStatusCount as GenAPIStatusCount,
} from './types.gen'

// Re-export generated rubric types for direct use
export type {
  GenRubric as Rubric,
  GenCategoryResult as CategoryResult,
  GenFinding as Finding,
  GenActionItem as ActionItem,
  GenNextSteps as NextSteps,
  GenDecision as Decision,
}

// JudgeResult with report using Rubric type from structured-evaluation
export interface JudgeResult {
  id: string
  initiativeId: string
  specPath: string
  specType?: string
  rubricId?: string
  evaluatedAt: string
  report?: GenRubric
}

export interface SpecWorkflowPhase {
  id: string
  name: string
  specs: string[]
}

export interface SpecWorkflow {
  id: string
  name: string
  description?: string
  specsRequired?: string[]
  specsOptional?: string[]
  initTypes?: string[]
  /** Spec types (uppercase, matching SpecFile.specType) in flow order */
  sequence?: string[]
  phases?: SpecWorkflowPhase[]
}

export interface SpecsResponse {
  workflows: SpecWorkflow[]
  judgeResults: JudgeResult[]
}

export interface APIInitiative {
  id: string
  title: string
  description?: string
  status: string
  type?: string
  priority?: string
  homeRepo?: string
  workflowId?: string
  programId?: string
  programName?: string
  hidden?: boolean
  progress: number
}

export interface APIPhase {
  id: string
  initiativeId: string
  sequenceNumber: number
  title: string
  theme?: string
  progress: number
}

export interface APIRMI {
  id: string
  repositoryId?: string
  initiativeId?: string
  phaseId?: string
  title: string
  description?: string
  type?: string
  status: string
  priority?: string
  sequenceNumber: number
  claimedBy?: string
  claimedAt?: string
  completedAt?: string
  tokensTotal?: number
  costUsd?: number
}

export interface APIRMIDependency {
  sourceRmiId: string
  targetRmiId: string
  relationship: string
}

export interface APIInitiativeDependency {
  sourceInitiativeId: string
  targetInitiativeId: string
  relationship: string
}

export interface APIStatusCount {
  status: string
  count: number
}

export interface APIProgram {
  id: string
  name: string
  organization?: string
  description?: string
  hidden?: boolean
  initiatives?: APIInitiative[]
}

export type SpecFileRole = 'required' | 'optional' | 'extra'

export interface SpecFile {
  initiativeId: string
  specType: string
  path: string
  content: string
  modTime?: string
  evalJson?: string
  /** Classification against the initiative's workflow; absent if unresolved */
  role?: SpecFileRole
}

export interface SpecFilesResponse {
  files: SpecFile[]
}

export interface APIRepository {
  id: string
  organization: string
  repositoryName: string
  defaultBranch?: string
  localPath?: string
  goModule?: string
  domain?: string
  status: string
}

export interface ExecutionResponse {
  programs: APIProgram[]
  initiatives: APIInitiative[]
  phases: APIPhase[]
  rmis: APIRMI[]
  repositories: APIRepository[]
  statusDistribution: APIStatusCount[]
  rmiDependencies: APIRMIDependency[]
  initiativeDependencies: APIInitiativeDependency[]
}

// Conversion functions with defaults

// Convert generated JudgeResult to normalized JudgeResult
// The report is passed through directly as it uses rubric.Rubric
export function toJudgeResult(gen: GenJudgeResult): JudgeResult {
  return {
    id: gen.id ?? '',
    initiativeId: gen.initiativeId ?? '',
    specPath: gen.specPath ?? '',
    specType: gen.specType,
    rubricId: gen.rubricId,
    evaluatedAt: gen.evaluatedAt ?? '',
    report: gen.report,
  }
}

export function toSpecWorkflow(gen: GenSpecWorkflow): SpecWorkflow {
  return {
    id: gen.id ?? '',
    name: gen.name ?? '',
    description: gen.description,
    specsRequired: gen.specsRequired,
    specsOptional: gen.specsOptional,
    initTypes: gen.initTypes,
    sequence: gen.sequence,
    phases: (gen.phases ?? []).map((p) => ({
      id: p.id ?? '',
      name: p.name ?? '',
      specs: p.specs ?? [],
    })),
  }
}

export function toSpecsResponse(gen: GenSpecsResponse): SpecsResponse {
  return {
    workflows: (gen.workflows ?? []).map(toSpecWorkflow),
    judgeResults: (gen.judgeResults ?? []).map(toJudgeResult),
  }
}

export function toAPIInitiative(gen: GenAPIInitiative): APIInitiative {
  return {
    id: gen.id ?? '',
    title: gen.title ?? '',
    description: gen.description,
    status: gen.status ?? '',
    type: gen.type,
    priority: gen.priority,
    homeRepo: gen.homeRepo,
    workflowId: gen.workflowId,
    programId: gen.programId,
    programName: gen.programName,
    hidden: gen.hidden,
    progress: gen.progress ?? 0,
  }
}

export function toAPIPhase(gen: GenAPIPhase): APIPhase {
  return {
    id: gen.id ?? '',
    initiativeId: gen.initiativeId ?? '',
    sequenceNumber: gen.sequenceNumber ?? 0,
    title: gen.title ?? '',
    theme: gen.theme,
    progress: gen.progress ?? 0,
  }
}

export function toAPIRMI(gen: GenAPIRMI): APIRMI {
  return {
    id: gen.id ?? '',
    repositoryId: gen.repositoryId,
    initiativeId: gen.initiativeId,
    phaseId: gen.phaseId,
    title: gen.title ?? '',
    description: gen.description,
    type: gen.type,
    status: gen.status ?? '',
    priority: gen.priority,
    sequenceNumber: gen.sequenceNumber ?? 0,
    claimedBy: gen.claimedBy,
    claimedAt: gen.claimedAt,
    completedAt: gen.completedAt,
    tokensTotal: gen.tokensTotal,
    costUsd: gen.costUsd,
  }
}

export function toAPIRMIDependency(gen: GenAPIRMIDependency): APIRMIDependency {
  return {
    sourceRmiId: gen.sourceRmiId ?? '',
    targetRmiId: gen.targetRmiId ?? '',
    relationship: gen.relationship ?? '',
  }
}

export function toAPIInitiativeDependency(gen: GenAPIInitiativeDependency): APIInitiativeDependency {
  return {
    sourceInitiativeId: gen.sourceInitiativeId ?? '',
    targetInitiativeId: gen.targetInitiativeId ?? '',
    relationship: gen.relationship ?? '',
  }
}

export function toAPIStatusCount(gen: GenAPIStatusCount): APIStatusCount {
  return {
    status: gen.status ?? '',
    count: gen.count ?? 0,
  }
}

export function toAPIProgram(gen: GenAPIProgram): APIProgram {
  return {
    id: gen.id ?? '',
    name: gen.name ?? '',
    organization: gen.organization,
    description: gen.description,
    hidden: gen.hidden,
    initiatives: gen.initiatives?.map(toAPIInitiative),
  }
}

export function toAPIRepository(gen: GenAPIRepository): APIRepository {
  return {
    id: gen.id ?? '',
    organization: gen.organization ?? '',
    repositoryName: gen.repositoryName ?? '',
    defaultBranch: gen.defaultBranch,
    localPath: gen.localPath,
    goModule: gen.goModule,
    domain: gen.domain,
    status: gen.status ?? '',
  }
}

export function toExecutionResponse(gen: GenExecutionResponse): ExecutionResponse {
  return {
    programs: (gen.programs ?? []).map(toAPIProgram),
    initiatives: (gen.initiatives ?? []).map(toAPIInitiative),
    phases: (gen.phases ?? []).map(toAPIPhase),
    rmis: (gen.rmis ?? []).map(toAPIRMI),
    repositories: (gen.repositories ?? []).map(toAPIRepository),
    statusDistribution: (gen.statusDistribution ?? []).map(toAPIStatusCount),
    rmiDependencies: (gen.rmiDependencies ?? []).map(toAPIRMIDependency),
    initiativeDependencies: (gen.initiativeDependencies ?? []).map(toAPIInitiativeDependency),
  }
}

export function toSpecFile(gen: GenSpecFile): SpecFile {
  const role = gen.role
  return {
    initiativeId: gen.initiativeId ?? '',
    specType: gen.specType ?? '',
    path: gen.path ?? '',
    content: gen.content ?? '',
    modTime: gen.modTime,
    evalJson: gen.evalJson,
    role: role === 'required' || role === 'optional' || role === 'extra' ? role : undefined,
  }
}

export function toSpecFilesResponse(gen: GenSpecFilesResponse): SpecFilesResponse {
  return {
    files: (gen.files ?? []).map(toSpecFile),
  }
}

// Helper to get score from JudgeResult
export function getScore(r: JudgeResult): number {
  return r.report?.intScore ?? 0
}
