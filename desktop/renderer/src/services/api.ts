import type { Project, Spec, EvalResult } from '../types'

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
