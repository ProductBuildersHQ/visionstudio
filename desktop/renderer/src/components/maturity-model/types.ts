/**
 * Maturity Model Types
 *
 * These types represent the maturity model data structures from prism-maturity.
 */

/** Maturity level definition (M1-M5) */
export interface MaturityLevel {
  level: number
  name: string
  description?: string
  color: string
}

/** A capability within a dimension */
export interface MaturityCapability {
  id: string
  name: string
  level: number
  targetLevel?: number
}

/** A dimension/domain in the maturity model */
export interface MaturityDimension {
  id: string
  name: string
  description?: string
  currentLevel: number
  targetLevel: number
  capabilities: MaturityCapability[]
}

/** Goal with maturity progression */
export interface MaturityGoal {
  id: string
  name: string
  description?: string
  owner?: string
  priority?: number
  status?: string
  currentLevel: number
  targetLevel: number
  startDate?: string
  targetDate?: string
}

/** Phase in the maturity roadmap */
export interface MaturityPhase {
  id: string
  name: string
  quarter?: string
  year?: number
  startDate: string
  endDate: string
  status?: string
  goals: MaturityGoal[]
}

/** Full maturity model */
export interface MaturityModel {
  id: string
  name: string
  description?: string
  levels: MaturityLevel[]
  dimensions: MaturityDimension[]
  goals?: MaturityGoal[]
  phases?: MaturityPhase[]
  overallScore: number
  lastUpdated: string
}

/** Summary for listing models */
export interface MaturityModelSummary {
  id: string
  name: string
  description?: string
  dimensionCount: number
  goalCount: number
  overallScore: number
  lastUpdated: string
}

/** Default maturity levels (M1-M5) */
export const DEFAULT_MATURITY_LEVELS: MaturityLevel[] = [
  { level: 1, name: 'Reactive', color: '#ef4444' },
  { level: 2, name: 'Basic', color: '#f59e0b' },
  { level: 3, name: 'Defined', color: '#eab308' },
  { level: 4, name: 'Managed', color: '#22c55e' },
  { level: 5, name: 'Optimizing', color: '#3b82f6' },
]

/** Get color for a maturity level */
export function getMaturityColor(level: number): string {
  const maturityLevel = DEFAULT_MATURITY_LEVELS.find(l => l.level === level)
  return maturityLevel?.color ?? '#6b7280'
}

/** Get name for a maturity level */
export function getMaturityName(level: number): string {
  const maturityLevel = DEFAULT_MATURITY_LEVELS.find(l => l.level === level)
  return maturityLevel?.name ?? `M${level}`
}
