// Project and spec types

export type SpecStatus = 'not_started' | 'draft' | 'evaluated' | 'approved'

export type EvalDecision = 'pass' | 'conditional' | 'fail'

// Score labels for 1-5 integer scores
export const SCORE_LABELS = ['', 'Unacceptable', 'Major Revisions', 'Acceptable', 'Good', 'Excellent']
export const SCORE_COLORS = ['', 'red', 'orange', 'yellow', 'green', 'blue']

export interface Finding {
  category: string
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  message: string
  // V2 fields
  code?: string     // Reason code (e.g., "AMBIGUOUS_REQUIREMENT")
  location?: string // Reference to where issue was found
}

// V2: Per-dimension score
export interface DimensionScore {
  id: string
  name: string
  score: number  // 1-5 integer score
  severity: 'none' | 'minor' | 'major' | 'critical'
  confidence: number  // 0.0-1.0
  reasonCodes: string[]
  findings: Finding[]
}

export interface EvalResult {
  // V1 fields (backwards compatibility)
  score: number
  decision: EvalDecision
  findings: Finding[]

  // V2 fields
  schemaVersion?: string      // "v2" for new format
  scoreV2?: number            // 1-5 integer score
  pass?: boolean              // Explicit pass/fail gate
  confidence?: number         // 0.0-1.0
  dimensions?: DimensionScore[]  // Per-dimension scores
  blocking?: string[]         // Blocking reason codes
}

// Helper to get score label from 1-5 score
export function getScoreLabel(score: number): string {
  if (score >= 1 && score <= 5) {
    return SCORE_LABELS[score]
  }
  return 'Unknown'
}

// Helper to check if evaluation needs human review (low confidence)
export function needsHumanReview(result: EvalResult, threshold = 0.7): boolean {
  if (result.confidence && result.confidence < threshold) {
    return true
  }
  if (result.dimensions) {
    return result.dimensions.some(d => d.confidence < threshold)
  }
  return false
}

export interface Spec {
  type: string
  name: string
  path: string
  status: SpecStatus
  evalResult?: EvalResult
  content?: string
}

export interface Profile {
  name: string
  description: string
  workflow: string[]
  framework?: string // e.g., "AWS", "Big Tech", "Startup"
  specType?: string  // e.g., "Product", "Feature", "Platform"
  type?: string      // "requirements" or "implementation"
}

export interface Project {
  name: string
  path: string
  profile: Profile                       // kept for backwards compat
  requirementsMethodology?: string       // e.g., "aws-working-backwards/product"
  implementationMethodology?: string     // e.g., "aidlc", "speckit", "none"
  specs: Spec[]
  gitRemote?: string                     // e.g., "https://github.com/org/repo"
}

// Implementation methodology options
export type ImplementationMethodology = 'none' | 'aidlc' | 'speckit'

// Methodology summary for listing
export interface ImplementationMethodologySummary {
  id: string
  name: string
  description: string
}

// Project methodology configuration
export interface ProjectMethodologyConfig {
  requirementsMethodology: string
  implementationMethodology: string
}

// Organization types

export interface Organization {
  id: string
  name: string
  description?: string
  v2momPath?: string
  fiscalYearStart?: string
  createdAt: string
  updatedAt: string
}

export type V2MOMScope = 'organization' | 'project'

export interface V2MOMValue {
  id: string
  name: string
  description?: string
  priority?: number
}

export interface V2MOMMethod {
  id: string
  name: string
  description?: string
  status?: string
  owner?: string
  dueDate?: string
  priority?: number
}

export interface V2MOMItem {
  id: string
  name: string
  description?: string
  severity?: string
  mitigation?: string
}

export interface V2MOMMeasure {
  id: string
  name: string
  description?: string
  target?: string
  current?: string
  unit?: string
  progress?: number
}

export interface OrganizationV2MOM {
  id: string
  name: string
  fiscalYear?: string
  owner?: string
  scope: V2MOMScope
  parentId?: string
  path: string
  vision: string
  values: V2MOMValue[]
  methods: V2MOMMethod[]
  obstacles: V2MOMItem[]
  measures: V2MOMMeasure[]
  lastUpdated: string
}

export interface V2MOMMethodAlignment {
  projectV2momId: string
  projectMethod: string
  orgV2momId: string
  orgMethod: string
  alignmentScore: number
  notes?: string
}

export interface V2MOMSummary {
  id: string
  name: string
  parentId?: string
  path: string
}

export interface OrganizationCascade {
  organization?: Organization
  orgV2moms: OrganizationV2MOM[]
  projectV2moms: Record<string, V2MOMSummary[]>
  alignments: V2MOMMethodAlignment[]
  alignmentScores: Record<string, number>
}

// Lint types

export interface LintFinding {
  path: string
  rule: string
  message: string
  severity: 'error' | 'warning' | 'info'
}

export interface LintResult {
  project: string
  findings: LintFinding[]
  errors: number
  warnings: number
  passed: boolean
}

// UI State types

export type ViewMode = 'source' | 'rendered'

export interface EditorState {
  activeSpec: Spec | null
  viewMode: ViewMode
  isDirty: boolean
}

export interface AppState {
  projects: Project[]
  activeProject: Project | null
  activeView: 'workflow' | 'spec'
  editor: EditorState
}
