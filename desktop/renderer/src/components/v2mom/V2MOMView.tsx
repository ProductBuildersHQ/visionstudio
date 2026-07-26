import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import { V2MOMCascadeView } from './V2MOMCascadeView'
import { LoadingState, ErrorState, EmptyState } from '../toolkit'
import { useApp } from '../../contexts/AppContext'
import type { V2MOM, V2MOMCascade } from './types'

interface V2MOMViewProps {
  projectName?: string
}

// Internal type used by V2MOMCascadeView
interface V2MOMLevel {
  id: string
  name: string
  level: 'company' | 'department' | 'team' | 'individual'
  owner: string
  fiscalPeriod: string
  vision: string
  visionScore?: number
  sections: {
    id: string
    type: 'vision' | 'values' | 'methods' | 'obstacles' | 'measures'
    name: string
    status: 'draft' | 'complete' | 'approved'
    score?: number
  }[]
  methods: {
    id: string
    name: string
    owner: string
    status: 'not_started' | 'in_progress' | 'at_risk' | 'completed'
    capabilities?: string[]
    projects?: string[]
    parentMethod?: string
    childMethods?: string[]
  }[]
  parentId?: string
  childIds?: string[]
  alignmentScore?: number
}

/**
 * V2MOMView wraps V2MOMCascadeView with API data fetching.
 * Fetches V2MOM cascade data and transforms it for the visualization.
 */
export function V2MOMView(props: V2MOMViewProps) {
  const app = useApp()
  const projectName = props.projectName ?? app.activeProject?.name
  const [v2moms, setV2moms] = useState<V2MOMLevel[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (projectName) loadV2MOMs()
  }, [projectName])

  async function loadV2MOMs() {
    if (!projectName) return
    setIsLoading(true)
    setError(null)
    try {
      const cascade = await api.getV2MOMCascade(projectName)
      const transformed = transformCascade(cascade)
      setV2moms(transformed)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load V2MOMs'
      setError(message)
    } finally {
      setIsLoading(false)
    }
  }

  if (!projectName) return null

  if (isLoading) {
    return <LoadingState message="Loading V2MOM cascade..." />
  }

  if (error) {
    return (
      <ErrorState
        message={error}
        hint="Make sure the project has V2MOM data configured."
        onRetry={loadV2MOMs}
      />
    )
  }

  if (v2moms.length === 0) {
    return (
      <EmptyState
        icon="🎯"
        title="No V2MOMs"
        description="This project doesn't have any V2MOMs configured yet. Add V2MOM documents to visualize strategic alignment."
        hint="v2mom/*.json"
      />
    )
  }

  return (
    <V2MOMCascadeView
      v2moms={v2moms}
      onSectionClick={(v2momId, section) => {
        console.log('Section clicked:', v2momId, section)
      }}
      onMethodClick={(v2momId, method) => {
        console.log('Method clicked:', v2momId, method)
      }}
    />
  )
}

/**
 * Transform API V2MOMCascade to V2MOMLevel[] for the visualization component.
 */
function transformCascade(cascade: V2MOMCascade): V2MOMLevel[] {
  // Handle both old format (root + children arrays) and new format (v2moms array)
  const v2momList: V2MOM[] = []

  // Check for the cascade.v2moms array (from types.ts)
  if ('v2moms' in cascade && Array.isArray(cascade.v2moms)) {
    v2momList.push(...cascade.v2moms)
  }

  // Also check for root/children structure (from API response)
  const cascadeAny = cascade as unknown as Record<string, unknown>
  if (cascadeAny.root) {
    v2momList.push(cascadeAny.root as V2MOM)
  }
  if (cascadeAny.children && Array.isArray(cascadeAny.children)) {
    v2momList.push(...(cascadeAny.children as V2MOM[]))
  }

  return v2momList.map(transformV2MOM)
}

/**
 * Transform a single V2MOM to V2MOMLevel format.
 */
function transformV2MOM(v2mom: V2MOM): V2MOMLevel {
  // Extract metadata if present
  const metadata = (v2mom as unknown as Record<string, unknown>).metadata as Record<string, unknown> | undefined

  return {
    id: v2mom.id || (metadata?.id as string) || 'unknown',
    name: v2mom.name || (metadata?.name as string) || 'Unnamed V2MOM',
    level: v2mom.level || 'team',
    owner: v2mom.owner || (metadata?.owner as string) || 'Unknown',
    fiscalPeriod: v2mom.fiscalPeriod || (metadata?.fiscalPeriod as string) || '',
    vision: v2mom.vision || '',
    visionScore: undefined,
    sections: buildSections(v2mom),
    methods: (v2mom.methods || []).map((m) => ({
      id: m.id || '',
      name: m.name || '',
      owner: m.owner || '',
      status: m.status || 'not_started',
      capabilities: m.capabilities,
      projects: m.projects,
      parentMethod: m.parentMethod,
      childMethods: m.childMethods,
    })),
    parentId: v2mom.parentId || (metadata?.parentId as string),
    childIds: v2mom.childIds,
    alignmentScore: v2mom.alignmentScore,
  }
}

/**
 * Build section badges from V2MOM data.
 */
function buildSections(v2mom: V2MOM): V2MOMLevel['sections'] {
  const sections: V2MOMLevel['sections'] = []

  // Add vision section
  if (v2mom.vision) {
    sections.push({
      id: 'vision',
      type: 'vision',
      name: 'Vision',
      status: 'complete',
    })
  }

  // Add values section
  if (v2mom.values && v2mom.values.length > 0) {
    sections.push({
      id: 'values',
      type: 'values',
      name: 'Values',
      status: 'complete',
    })
  }

  // Add methods section
  if (v2mom.methods && v2mom.methods.length > 0) {
    sections.push({
      id: 'methods',
      type: 'methods',
      name: 'Methods',
      status: 'complete',
    })
  }

  // Add obstacles section
  if (v2mom.obstacles && v2mom.obstacles.length > 0) {
    sections.push({
      id: 'obstacles',
      type: 'obstacles',
      name: 'Obstacles',
      status: 'complete',
    })
  }

  // Add measures section
  if (v2mom.measures && v2mom.measures.length > 0) {
    sections.push({
      id: 'measures',
      type: 'measures',
      name: 'Measures',
      status: 'complete',
    })
  }

  return sections
}

export default V2MOMView
