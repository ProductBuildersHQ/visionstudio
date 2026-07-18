import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import { V2MOMCascadeView } from './V2MOMCascadeView'
import type { V2MOM, V2MOMCascade } from './types'

interface V2MOMViewProps {
  projectName: string
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
export function V2MOMView({ projectName }: V2MOMViewProps) {
  const [v2moms, setV2moms] = useState<V2MOMLevel[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadV2MOMs()
  }, [projectName])

  async function loadV2MOMs() {
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

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full bg-va-bg">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-2 border-va-accent border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-va-text-muted">Loading V2MOM cascade...</p>
        </div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-va-bg gap-4">
        <div className="text-center max-w-md">
          <div className="text-va-error mb-4">{error}</div>
          <p className="text-va-text-muted text-sm mb-4">
            Make sure the project has V2MOM data configured.
          </p>
          <button
            onClick={loadV2MOMs}
            className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  // No V2MOMs available
  if (v2moms.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-va-bg">
        <div className="text-center max-w-md">
          <div className="text-5xl mb-4">🎯</div>
          <h2 className="text-xl font-semibold text-va-text mb-2">No V2MOMs</h2>
          <p className="text-va-text-muted text-sm mb-4">
            This project doesn't have any V2MOMs configured yet.
            Add V2MOM documents to visualize strategic alignment.
          </p>
          <div className="text-xs text-va-text-muted">
            <code>v2mom/*.json</code>
          </div>
        </div>
      </div>
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
