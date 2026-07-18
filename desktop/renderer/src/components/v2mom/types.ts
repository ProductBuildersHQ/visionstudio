/**
 * V2MOM Types for VisionStudio
 *
 * Defines the data structures for V2MOM visualization,
 * including cascade relationships and ProductContext integration.
 */

// V2MOM Level in the cascade hierarchy
export type V2MOMLevelType = 'company' | 'department' | 'team' | 'individual'

// V2MOM Section types (the five elements)
export type V2MOMSectionType =
  | 'vision'
  | 'values'
  | 'methods'
  | 'obstacles'
  | 'measures'
  | 'alignment'
  | 'summary'

// Section status
export type SectionStatus = 'not_started' | 'draft' | 'complete' | 'approved'

// Method status
export type MethodStatus =
  | 'not_started'
  | 'in_progress'
  | 'at_risk'
  | 'completed'

// Obstacle severity
export type ObstacleSeverity = 'critical' | 'high' | 'medium' | 'low'

// Obstacle category
export type ObstacleCategory =
  | 'technical'
  | 'organizational'
  | 'resource'
  | 'market'
  | 'external'

// V2MOM Section document
export interface V2MOMSection {
  id: string
  type: V2MOMSectionType
  name: string
  status: SectionStatus
  path?: string // File path for editing
  score?: number // Evaluation score
  updatedAt?: string
}

// V2MOM Value (ranked)
export interface V2MOMValue {
  rank: number
  name: string
  statement: string
  tradeoffGuidance: string
}

// V2MOM Method
export interface V2MOMMethod {
  id: string
  name: string
  owner: string
  priority: 'P0' | 'P1' | 'P2'
  status: MethodStatus
  description?: string
  successCriteria?: string[]
  targetDate?: string

  // ProductContext links
  capabilities?: string[]
  projects?: string[]
  customerRequests?: string[]

  // Cascade links
  parentMethod?: string // Which parent method this supports
  childMethods?: string[] // Which child methods support this

  // Progress
  progress?: number // 0-100
}

// V2MOM Obstacle
export interface V2MOMObstacle {
  id: string
  name: string
  category: ObstacleCategory
  severity: ObstacleSeverity
  owner: string
  status: 'identified' | 'mitigating' | 'resolved' | 'accepted'
  description?: string
  affectedMethods?: string[]
  mitigationStrategies?: string[]

  // ProductContext links
  decisions?: string[]
  customerRequests?: string[]
}

// V2MOM Measure
export interface V2MOMMeasure {
  id: string
  name: string
  methodId: string // Which method this measures
  owner: string
  frequency: 'daily' | 'weekly' | 'monthly' | 'quarterly'

  // Targets
  baseline: number
  target: number
  stretch?: number
  current?: number

  // Status
  status: 'on_track' | 'at_risk' | 'off_track' | 'not_started'

  // ProductContext links
  kpis?: string[]
  outcomes?: string[]

  // Cascade
  parentMeasure?: string
  contributionWeight?: number // % contribution to parent
}

// Complete V2MOM document
export interface V2MOM {
  id: string
  name: string
  level: V2MOMLevelType
  fiscalPeriod: string
  owner: string
  ownerTeam?: string
  status: 'draft' | 'active' | 'completed' | 'archived'

  // Cascade relationships
  parentId?: string
  childIds?: string[]
  siblingIds?: string[]

  // Five elements
  vision: string
  visionRationale?: string
  values: V2MOMValue[]
  methods: V2MOMMethod[]
  obstacles: V2MOMObstacle[]
  measures: V2MOMMeasure[]

  // Section documents (for VisionSpec)
  sections: V2MOMSection[]

  // Scores
  alignmentScore?: number // How well aligned to parent (0-100)
  healthScore?: number // Overall health (0-100)

  // Timestamps
  createdAt: string
  updatedAt: string

  // ProductContext links
  productContext?: {
    strategies?: string[]
    capabilities?: string[]
    projects?: string[]
    kpis?: string[]
    decisions?: string[]
  }
}

// Alignment data between V2MOMs
export interface V2MOMAlignment {
  childV2MOMId: string
  parentV2MOMId: string

  // Vision alignment
  visionAlignmentScore: number
  supportedParentMethods: string[]

  // Method mapping
  methodMappings: {
    childMethodId: string
    parentMethodId: string
    contribution: string
    coverage: number // 0-100
  }[]

  // Measure rollup
  measureRollups: {
    childMeasureId: string
    parentMeasureId: string
    rollupType: 'sum' | 'average' | 'min' | 'milestone'
    weight: number
  }[]

  // Coverage analysis
  parentMethodCoverage: {
    methodId: string
    coverage: number
    coveredBy: string[]
    gap?: string
  }[]

  // Overall scores
  overallScore: number
  visionMethodsScore: number
  methodCoverageScore: number
  measureRollupScore: number
  dependencyScore: number
}

// V2MOM Cascade (full hierarchy)
export interface V2MOMCascade {
  rootV2MOMId: string
  v2moms: V2MOM[]
  alignments: V2MOMAlignment[]
  fiscalPeriod: string

  // Summary stats
  totalMethods: number
  completedMethods: number
  atRiskMethods: number
  averageAlignmentScore: number
}

// API response types
export interface V2MOMListResponse {
  v2moms: V2MOM[]
  total: number
}

export interface V2MOMCascadeResponse {
  cascade: V2MOMCascade
}

export interface V2MOMAlignmentResponse {
  alignment: V2MOMAlignment
}

// ProductContext entity reference (for linking)
export interface ProductContextRef {
  type:
    | 'capability'
    | 'project'
    | 'customer'
    | 'kpi'
    | 'decision'
    | 'customer_request'
    | 'strategy'
  id: string
  name?: string
}
