import type { SpendResponse, MaturityResponse, ScaleResponse, LeverageGraph } from './types'
import type {
  ExecutionResponse as GenExecutionResponse,
  SpecsResponse as GenSpecsResponse,
  SpecFilesResponse as GenSpecFilesResponse,
  CreateInitiativeRequest,
  CreateInitiativeResponse,
  WorkflowSpecDetail,
} from './types.gen'
import {
  toExecutionResponse,
  toSpecsResponse,
  toSpecFilesResponse,
  type ExecutionResponse,
  type SpecsResponse,
  type SpecFilesResponse,
} from './compat'

const BASE_URL = '/api'

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`)
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export async function getExecution(): Promise<ExecutionResponse> {
  const raw = await fetchJSON<GenExecutionResponse>('/execution')
  return toExecutionResponse(raw)
}

export async function getSpend(): Promise<SpendResponse> {
  return fetchJSON<SpendResponse>('/spend')
}

export async function getMaturity(): Promise<MaturityResponse> {
  return fetchJSON<MaturityResponse>('/maturity')
}

export async function getSpecs(): Promise<SpecsResponse> {
  const raw = await fetchJSON<GenSpecsResponse>('/specs')
  return toSpecsResponse(raw)
}

export async function getSpecFiles(initiativeId: string): Promise<SpecFilesResponse> {
  const raw = await fetchJSON<GenSpecFilesResponse>(`/spec-files/${encodeURIComponent(initiativeId)}`)
  return toSpecFilesResponse(raw)
}

export async function createInitiative(req: CreateInitiativeRequest): Promise<CreateInitiativeResponse> {
  const res = await fetch(`${BASE_URL}/initiatives`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  if (!res.ok) {
    const detail = (await res.text()).trim()
    throw new Error(detail || `API error: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<CreateInitiativeResponse>
}

export type { CreateInitiativeRequest, CreateInitiativeResponse, WorkflowSpecDetail }

export async function getWorkflowSpecDetail(workflowId: string, specType: string): Promise<WorkflowSpecDetail> {
  return fetchJSON<WorkflowSpecDetail>(
    `/workflows/${encodeURIComponent(workflowId)}/specs/${encodeURIComponent(specType)}`
  )
}

export async function getScale(): Promise<ScaleResponse> {
  return fetchJSON<ScaleResponse>('/scale')
}

export async function getLeverage(): Promise<LeverageGraph> {
  return fetchJSON<LeverageGraph>('/leverage')
}
