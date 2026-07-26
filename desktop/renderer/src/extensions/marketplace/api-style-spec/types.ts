export interface EvaluationReport {
  metadata?: EvaluationMetadata
  reviewType?: string
  rubricId?: string
  rubricVersion?: string
  categories: CategoryResult[]
  findings: EvaluationFinding[]
  passCriteria?: PassCriteria
  decision?: EvaluationDecision
  overallDecision: string
  nextSteps?: NextSteps
  summary?: string
}

export interface EvaluationMetadata {
  document?: string
  documentTitle?: string
  generatedAt?: string
  generatedBy?: string
  toolVersion?: string
}

export interface CategoryResult {
  category: string
  score: string
  numericScore: number
  weight?: number
  required?: boolean
  reasoning: string
  findings?: EvaluationFinding[]
}

export interface EvaluationFinding {
  severity: string
  category: string
  finding: string
  recommendation?: string
  location?: string
  ruleId?: string
}

export interface PassCriteria {
  minCategoriesPassing?: string
  maxFindings?: FindingLimits
}

export interface FindingLimits {
  critical?: number
  high?: number
  medium?: number
  low?: number
}

export interface EvaluationDecision {
  status: string
  reasoning: string
  categoryCounts?: CategoryCounts
  findingCounts?: FindingCounts
}

export interface CategoryCounts {
  pass: number
  partial: number
  fail: number
  total: number
}

export interface FindingCounts {
  critical: number
  high: number
  medium: number
  low: number
  total: number
}

export interface NextSteps {
  immediate?: ActionItem[]
  recommended?: ActionItem[]
}

export interface ActionItem {
  action: string
  category?: string
  effort?: string
  priority?: number
}

export interface LintReport {
  violations: Violation[]
  summary: LintSummary
  metadata: LintMetadata
}

export interface Violation {
  ruleId: string
  severity: string
  message: string
  path?: string
  line?: number
  column?: number
  suggestion?: string
  exampleFix?: string
  ruleURL?: string
  confidence?: number
  fixPriority?: string
  relatedRules?: string[]
}

export interface LintSummary {
  errors: number
  warnings: number
  infos: number
  hints: number
  total: number
}

export interface LintMetadata {
  specFile?: string
  profile?: string
  duration?: string
  rulesEvaluated?: number
}

export interface StyleRule {
  id: string
  title: string
  category: string
  severity: string
  rationale?: string
  examples?: {
    good?: string[]
    bad?: string[]
  }
}

export interface StyleProfile {
  name: string
  description?: string
  version?: string
  rules: StyleRule[]
  categories?: string[]
}

export interface FixSuggestion {
  ruleId: string
  path: string
  currentValue?: string
  suggestedValue: string
  diff?: string
  confidence: number
  reasoning?: string
  breaking?: boolean
  breakingReason?: string
}

export interface FixReport {
  suggestions: FixSuggestion[]
  fixedCount: number
  unfixedCount: number
  unfixedRules?: string[]
}

export interface LintResponse {
  status: string
  violations: Violation[]
  summary: LintSummary
  profile: string
}

export interface ProfileInfo {
  name: string
  description: string
}
