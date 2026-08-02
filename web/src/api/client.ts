import type {
  ExecutionResponse,
  SpendResponse,
  MaturityResponse,
  SpecsResponse,
} from './types'

const BASE_URL = '/api'

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`)
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export async function getExecution(): Promise<ExecutionResponse> {
  return fetchJSON<ExecutionResponse>('/execution')
}

export async function getSpend(): Promise<SpendResponse> {
  return fetchJSON<SpendResponse>('/spend')
}

export async function getMaturity(): Promise<MaturityResponse> {
  return fetchJSON<MaturityResponse>('/maturity')
}

export async function getSpecs(): Promise<SpecsResponse> {
  return fetchJSON<SpecsResponse>('/specs')
}
