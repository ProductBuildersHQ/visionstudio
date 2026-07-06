import type { Project, Spec, EvalResult, Profile, LintResult } from '../types'

const API_BASE = 'http://127.0.0.1:8765/api'

// Workflow status type
export interface WorkflowStatus {
  currentPhase: string
  completedPhases: string[]
  progress: number
  specStatuses: Record<string, string>
  blockedBy?: string[]
  lastUpdated: string
}

// Workflow node
export interface WorkflowNode {
  id: string
  name: string
  description?: string
  type: string   // "source", "gtm", "technical", "output"
  phase: string
  status: string // "pending", "ready", "in_progress", "completed", "blocked", "skipped"
  depends_on?: string[]
  automated?: boolean
  metadata?: Record<string, unknown>
}

// Workflow phase
export interface WorkflowPhase {
  id: string
  name: string
  description?: string
  order: number
  nodes: string[] // Node IDs
}

// Workflow progress
export interface WorkflowProgress {
  completed: number
  total: number
  percent: number
}

// Full workflow
export interface Workflow {
  name: string
  description?: string
  phases: WorkflowPhase[]
  nodes: Record<string, WorkflowNode>
  progress: WorkflowProgress
}

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`)
  }

  return response.json()
}

export const api = {
  // Health check
  async health(): Promise<{ status: string }> {
    return fetchJSON(`${API_BASE}/health`)
  },

  // Projects
  async listProjects(): Promise<Project[]> {
    const data = await fetchJSON<{ projects: Project[] }>(`${API_BASE}/projects`)
    return data.projects
  },

  async getProject(name: string): Promise<Project> {
    const data = await fetchJSON<{ project: Project }>(`${API_BASE}/projects/${name}`)
    return data.project
  },

  async addProject(name: string, path: string, profile: string, initialize = false): Promise<Project> {
    const data = await fetchJSON<{ project: Project; error?: string }>(
      `${API_BASE}/projects`,
      {
        method: 'POST',
        body: JSON.stringify({ name, path, profile, initialize }),
      }
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.project
  },

  async removeProject(name: string): Promise<void> {
    const data = await fetchJSON<{ success: boolean; error?: string }>(
      `${API_BASE}/projects/${name}`,
      { method: 'DELETE' }
    )
    if (data.error) {
      throw new Error(data.error)
    }
  },

  async listProfiles(): Promise<Profile[]> {
    const data = await fetchJSON<{ profiles: Profile[] }>(`${API_BASE}/profiles`)
    return data.profiles
  },

  async lintProject(project: string): Promise<LintResult> {
    const data = await fetchJSON<{ result: LintResult; error?: string }>(
      `${API_BASE}/projects/${project}/lint`
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.result
  },

  // Specs
  async getSpec(project: string, specType: string): Promise<Spec> {
    const data = await fetchJSON<{ spec: Spec }>(
      `${API_BASE}/projects/${project}/specs/${specType}`
    )
    return data.spec
  },

  async saveSpec(project: string, specType: string, content: string): Promise<void> {
    await fetchJSON(`${API_BASE}/projects/${project}/specs/${specType}`, {
      method: 'PUT',
      body: JSON.stringify({ content }),
    })
  },

  async evaluateSpec(project: string, specType: string): Promise<EvalResult> {
    const data = await fetchJSON<{ result: EvalResult }>(
      `${API_BASE}/projects/${project}/specs/${specType}/evaluate`,
      { method: 'POST' }
    )
    return data.result
  },

  // Chat
  async chat(message: string, context?: string): Promise<string> {
    const data = await fetchJSON<{ response: string; error?: string }>(
      `${API_BASE}/chat`,
      {
        method: 'POST',
        body: JSON.stringify({ message, context }),
      }
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.response
  },

  // Workflow
  async getWorkflow(project: string): Promise<{ workflow: Workflow; mermaid: string }> {
    const data = await fetchJSON<{ workflow: Workflow; mermaid: string; error?: string }>(
      `${API_BASE}/projects/${project}/workflow`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return { workflow: data.workflow, mermaid: data.mermaid }
  },

  async getWorkflowStatus(project: string): Promise<WorkflowStatus> {
    const data = await fetchJSON<{ status: WorkflowStatus; error?: string }>(
      `${API_BASE}/projects/${project}/workflow/status`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.status
  },
}
