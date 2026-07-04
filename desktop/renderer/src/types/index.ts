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
}

export interface Project {
  name: string
  path: string
  profile: Profile
  specs: Spec[]
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
