// Project and spec types

export type SpecStatus = 'not_started' | 'draft' | 'evaluated' | 'approved'

export type EvalDecision = 'pass' | 'conditional' | 'fail'

export interface EvalResult {
  score: number
  decision: EvalDecision
  findings: Finding[]
}

export interface Finding {
  category: string
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  message: string
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
}

export interface Project {
  name: string
  path: string
  profile: Profile
  specs: Spec[]
  gitRemote?: string // e.g., "https://github.com/org/repo"
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
