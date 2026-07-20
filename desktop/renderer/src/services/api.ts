import type { Project, Spec, EvalResult, Profile, LintResult, ImplementationMethodologySummary, ProjectMethodologyConfig, Organization, OrganizationV2MOM, OrganizationCascade } from '../types'
import type { V2MOM, V2MOMCascade, V2MOMAlignment } from '../components/v2mom/types'
import type { MaturityModel, MaturityModelSummary } from '../components/maturity-model/types'
import type { DashforgeDashboard } from '../components/devx/types'

// Sample types
export interface SampleSummary {
  id: string
  name: string
  description: string
  complexity: 'simple' | 'enterprise'
  path: string
  fileCounts: Record<string, number>
}

export interface SampleDetail extends SampleSummary {
  projectJson?: Record<string, unknown>
  readme?: string
}

// Capability types
export interface CapabilitySummary {
  id: string
  name: string
  domain?: string
  path: string
}

export interface CapabilityStack {
  metadata: Record<string, unknown>
  layers: Record<string, unknown>[]
  categories: Record<string, unknown>[]
  capabilities: Record<string, unknown>[]
  prismIntegration?: Record<string, unknown>
}

// Roadmap types
export interface RoadmapItem {
  id: string
  title: string
  description: string
  status: string
  priority: string
  quarter: string
  effort: string
  rice?: Record<string, unknown>
  capability_refs?: string[]
  goal_refs?: string[]
}

export interface Roadmap {
  metadata: Record<string, unknown>
  items: RoadmapItem[]
}

// AIDLC types
export type AIDLCPhase = 'inception' | 'construction' | 'operations'
export type AIDLCDocStatus = 'pending' | 'draft' | 'in_progress' | 'completed' | 'blocked'
export type AIDLCQualityRating = 'EXCELLENT' | 'GOOD' | 'NEEDS_IMPROVEMENT' | 'POOR'

export interface AIDLCIssue {
  severity: string
  category: string
  code?: string
  message: string
  location?: string
  suggestion?: string
}

export interface AIDLCQualityScore {
  rating: AIDLCQualityRating
  score: number
  issues?: AIDLCIssue[]
  dimensions?: Record<string, number>
  evaluated_at?: string
}

export interface AIDLCDocument {
  type: string
  phase: AIDLCPhase
  path: string
  title: string
  status: AIDLCDocStatus
  content?: string
  score?: AIDLCQualityScore
  updated_at: string
}

export interface AIDLCState {
  current_phase: AIDLCPhase
  current_document?: string
  completed_docs: string[]
  pending_docs: string[]
  in_progress_docs?: string[]
  document_scores?: Record<string, AIDLCQualityScore>
  phase_progress?: Record<AIDLCPhase, number>
  overall_progress: number
  last_updated: string
}

export interface AIDLCWorkflowNode {
  id: string
  doc_type: string
  phase: AIDLCPhase
  name: string
  description?: string
  status: string
  score?: AIDLCQualityScore
  depends_on?: string[]
  blocks?: string[]
  required: boolean
  automated?: boolean
}

export interface AIDLCWorkflowPhase {
  id: string
  name: string
  description?: string
  order: number
  node_ids: string[]
}

export interface AIDLCWorkflowEdge {
  from: string
  to: string
  type: string
  label?: string
}

export interface AIDLCWorkflow {
  name: string
  description?: string
  phases: AIDLCWorkflowPhase[]
  nodes: Record<string, AIDLCWorkflowNode>
  edges: AIDLCWorkflowEdge[]
  progress: WorkflowProgress
}

export interface AIDLCSyncAction {
  direction: string
  doc_type: string
  source_path: string
  dest_path: string
  action: string
  reason?: string
}

export interface AIDLCSyncConflict {
  doc_type: string
  visionspec_path: string
  aidlc_path: string
  visionspec_mod_time: string
  aidlc_mod_time: string
  reason: string
}

export interface AIDLCSyncDiff {
  visionspec_dir: string
  aidlc_docs_dir: string
  actions: AIDLCSyncAction[]
  conflicts?: AIDLCSyncConflict[]
  computed_at: string
}

export interface AIDLCSyncResult {
  direction: string
  created?: string[]
  updated?: string[]
  skipped?: string[]
  errors?: string[]
  completed_at: string
}

// Phase requirements and transitions
export interface AIDLCPhaseRequirements {
  phase: AIDLCPhase
  required_docs: string[]
  completed_docs: string[]
  pending_docs: string[]
  progress_percent: number
  can_advance: boolean
}

export interface AIDLCPhaseTransitionResult {
  success: boolean
  from_phase: AIDLCPhase
  to_phase: AIDLCPhase
  blocking_docs?: string[]
  blocking_issues?: string[]
}

// Templates
export interface AIDLCTemplateSection {
  id: string
  title: string
  required: boolean
  description?: string
}

export interface AIDLCTemplate {
  doc_type: string
  name: string
  description: string
  phase: AIDLCPhase
  sections: AIDLCTemplateSection[]
  content?: string
}

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

  // Methodologies
  async listRequirementsMethodologies(): Promise<Profile[]> {
    const data = await fetchJSON<{ profiles: Profile[] }>(`${API_BASE}/methodologies/requirements`)
    return data.profiles
  },

  async listImplementationMethodologies(): Promise<ImplementationMethodologySummary[]> {
    const data = await fetchJSON<{ methodologies: ImplementationMethodologySummary[] }>(
      `${API_BASE}/methodologies/implementation`
    )
    return data.methodologies
  },

  async getProjectMethodology(project: string): Promise<ProjectMethodologyConfig> {
    const data = await fetchJSON<{ config: ProjectMethodologyConfig; error?: string }>(
      `${API_BASE}/projects/${project}/methodology`
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.config
  },

  async updateProjectMethodology(
    project: string,
    config: Partial<ProjectMethodologyConfig>
  ): Promise<ProjectMethodologyConfig> {
    const data = await fetchJSON<{ config: ProjectMethodologyConfig; error?: string }>(
      `${API_BASE}/projects/${project}/methodology`,
      {
        method: 'PUT',
        body: JSON.stringify(config),
      }
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.config
  },

  // Organization
  async getOrganization(): Promise<Organization | null> {
    const data = await fetchJSON<{ organization: Organization | null; error?: string }>(
      `${API_BASE}/organization`
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.organization
  },

  async createOrganization(name: string, description?: string): Promise<Organization> {
    const data = await fetchJSON<{ organization: Organization; error?: string }>(
      `${API_BASE}/organization`,
      {
        method: 'POST',
        body: JSON.stringify({ name, description }),
      }
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.organization
  },

  async updateOrganization(updates: Partial<Organization>): Promise<Organization> {
    const data = await fetchJSON<{ organization: Organization; error?: string }>(
      `${API_BASE}/organization`,
      {
        method: 'PUT',
        body: JSON.stringify(updates),
      }
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.organization
  },

  async deleteOrganization(): Promise<void> {
    const data = await fetchJSON<{ success: boolean; error?: string }>(
      `${API_BASE}/organization`,
      { method: 'DELETE' }
    )
    if (data.error) {
      throw new Error(data.error)
    }
  },

  async listOrganizationV2MOMs(): Promise<OrganizationV2MOM[]> {
    const data = await fetchJSON<{ v2moms: OrganizationV2MOM[]; error?: string }>(
      `${API_BASE}/organization/v2moms`
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.v2moms || []
  },

  async getOrganizationV2MOM(v2momId: string): Promise<OrganizationV2MOM> {
    const data = await fetchJSON<{ v2mom: OrganizationV2MOM; error?: string }>(
      `${API_BASE}/organization/v2moms/${v2momId}`
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.v2mom
  },

  async getOrganizationCascade(): Promise<OrganizationCascade> {
    const data = await fetchJSON<{ cascade: OrganizationCascade; error?: string }>(
      `${API_BASE}/organization/cascade`
    )
    if (data.error) {
      throw new Error(data.error)
    }
    return data.cascade
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

  // V2MOM
  async listV2MOMs(project: string): Promise<V2MOM[]> {
    const data = await fetchJSON<{ v2moms: V2MOM[]; error?: string }>(
      `${API_BASE}/projects/${project}/v2moms`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.v2moms || []
  },

  async getV2MOM(project: string, v2momId: string): Promise<V2MOM> {
    const data = await fetchJSON<{ v2mom: V2MOM; error?: string }>(
      `${API_BASE}/projects/${project}/v2moms/${v2momId}`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.v2mom
  },

  async getV2MOMCascade(project: string): Promise<V2MOMCascade> {
    const data = await fetchJSON<{ cascade: V2MOMCascade; error?: string }>(
      `${API_BASE}/projects/${project}/v2moms/cascade`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.cascade
  },

  async getV2MOMAlignment(
    project: string,
    childV2MOMId: string,
    parentV2MOMId: string
  ): Promise<V2MOMAlignment> {
    const data = await fetchJSON<{ alignment: V2MOMAlignment; error?: string }>(
      `${API_BASE}/projects/${project}/v2moms/${childV2MOMId}/alignment/${parentV2MOMId}`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.alignment
  },

  // Maturity Model
  async listMaturityModels(project: string): Promise<MaturityModelSummary[]> {
    const data = await fetchJSON<{ models: MaturityModelSummary[]; error?: string }>(
      `${API_BASE}/projects/${project}/maturity/models`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.models || []
  },

  async getMaturityModel(project: string, modelId: string): Promise<MaturityModel> {
    const data = await fetchJSON<{ model: MaturityModel; error?: string }>(
      `${API_BASE}/projects/${project}/maturity/models/${modelId}`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.model
  },

  async getMaturityDashboard(
    project: string,
    options?: { theme?: 'light' | 'dark'; modelId?: string }
  ): Promise<string> {
    const params = new URLSearchParams()
    if (options?.theme) params.set('theme', options.theme)
    if (options?.modelId) params.set('model', options.modelId)
    params.set('embed', 'true')

    const response = await fetch(
      `${API_BASE}/projects/${project}/maturity/dashboard?${params}`
    )

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    return response.text()
  },

  // Samples
  async listSamples(): Promise<SampleSummary[]> {
    const data = await fetchJSON<{ samples: SampleSummary[] }>(`${API_BASE}/samples`)
    return data.samples || []
  },

  async getSample(sampleId: string): Promise<SampleDetail> {
    const data = await fetchJSON<{ sample: SampleDetail; error?: string }>(
      `${API_BASE}/samples/${sampleId}`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.sample
  },

  // Capabilities
  async listCapabilities(project: string): Promise<CapabilitySummary[]> {
    const data = await fetchJSON<{ capabilities: CapabilitySummary[]; error?: string }>(
      `${API_BASE}/projects/${project}/capabilities`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.capabilities || []
  },

  async getCapability(project: string, capabilityId: string): Promise<CapabilityStack> {
    const data = await fetchJSON<{ capability: CapabilityStack; error?: string }>(
      `${API_BASE}/projects/${project}/capabilities/${capabilityId}`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.capability
  },

  // Roadmap
  async getRoadmap(project: string): Promise<Roadmap> {
    const data = await fetchJSON<{ roadmap: Roadmap; error?: string }>(
      `${API_BASE}/projects/${project}/roadmap`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.roadmap
  },

  // AIDLC (AWS AI DLC Workflow)
  async getAIDLCState(project: string): Promise<AIDLCState> {
    const data = await fetchJSON<{ state: AIDLCState; error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/state`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.state
  },

  async getAIDLCWorkflow(project: string): Promise<{ workflow: AIDLCWorkflow; mermaid: string }> {
    const data = await fetchJSON<{ workflow: AIDLCWorkflow; mermaid: string; error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/workflow`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return { workflow: data.workflow, mermaid: data.mermaid }
  },

  async listAIDLCDocuments(project: string): Promise<AIDLCDocument[]> {
    const data = await fetchJSON<{ documents: AIDLCDocument[]; error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/documents`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.documents || []
  },

  async getAIDLCDocument(project: string, docId: string): Promise<AIDLCDocument> {
    const data = await fetchJSON<{ document: AIDLCDocument; error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/documents/${docId}`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.document
  },

  async getAIDLCSyncDiff(project: string): Promise<AIDLCSyncDiff> {
    const data = await fetchJSON<{ diff: AIDLCSyncDiff; error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/sync/diff`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.diff
  },

  async syncAIDLC(project: string, options?: { direction?: string; dryRun?: boolean }): Promise<AIDLCSyncResult> {
    const data = await fetchJSON<{ result: AIDLCSyncResult; error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/sync`,
      {
        method: 'POST',
        body: JSON.stringify({
          direction: options?.direction || 'bidirectional',
          dry_run: options?.dryRun || false,
        }),
      }
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.result
  },

  async getAIDLCPhaseRequirements(project: string): Promise<{ currentPhase: AIDLCPhase; requirements: AIDLCPhaseRequirements[] }> {
    const data = await fetchJSON<{ current_phase: AIDLCPhase; requirements: AIDLCPhaseRequirements[]; error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/phase/requirements`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return { currentPhase: data.current_phase, requirements: data.requirements }
  },

  async transitionAIDLCPhase(
    project: string,
    targetPhase: AIDLCPhase,
    options?: { force?: boolean; approvedBy?: string; notes?: string }
  ): Promise<AIDLCPhaseTransitionResult> {
    const data = await fetchJSON<AIDLCPhaseTransitionResult & { error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/phase/transition`,
      {
        method: 'POST',
        body: JSON.stringify({
          target_phase: targetPhase,
          force: options?.force || false,
          approved_by: options?.approvedBy,
          notes: options?.notes,
        }),
      }
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data
  },

  async listAIDLCTemplates(): Promise<AIDLCTemplate[]> {
    const data = await fetchJSON<{ templates: AIDLCTemplate[]; error?: string }>(
      `${API_BASE}/projects/_/aidlc/templates`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.templates || []
  },

  async getAIDLCTemplate(docType: string): Promise<AIDLCTemplate> {
    const data = await fetchJSON<{ template: AIDLCTemplate; error?: string }>(
      `${API_BASE}/projects/_/aidlc/templates/${docType}`
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.template
  },

  async createAIDLCDocument(
    project: string,
    docType: string,
    options?: { title?: string; author?: string; description?: string; overwrite?: boolean }
  ): Promise<AIDLCDocument> {
    const data = await fetchJSON<{ document: AIDLCDocument; error?: string }>(
      `${API_BASE}/projects/${project}/aidlc/documents/create`,
      {
        method: 'POST',
        body: JSON.stringify({
          doc_type: docType,
          title: options?.title,
          author: options?.author,
          description: options?.description,
          overwrite: options?.overwrite || false,
        }),
      }
    )

    if (data.error) {
      throw new Error(data.error)
    }

    return data.document
  },

  // DevX (OmniDevX dashboard passthrough — not project-scoped)
  async getDevXDashboard(): Promise<DashforgeDashboard> {
    const response = await fetch(`${API_BASE}/devx/dashboard`)
    if (!response.ok) {
      const body = await response.json().catch(() => null) as { error?: string } | null
      throw new Error(body?.error || `HTTP ${response.status}: ${response.statusText}`)
    }
    return response.json()
  },
}
