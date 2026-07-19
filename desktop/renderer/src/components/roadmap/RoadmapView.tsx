import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import type { Roadmap, RoadmapItem } from '../../services/api'

interface RoadmapViewProps {
  projectName: string
}

/**
 * RoadmapView displays the project roadmap with items organized by quarter.
 * Shows RICE scoring, status, and capability/goal references.
 */
export function RoadmapView({ projectName }: RoadmapViewProps) {
  const [roadmap, setRoadmap] = useState<Roadmap | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'board' | 'list'>('board')

  useEffect(() => {
    loadRoadmap()
  }, [projectName])

  async function loadRoadmap() {
    setIsLoading(true)
    setError(null)
    try {
      const data = await api.getRoadmap(projectName)
      setRoadmap(data)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load roadmap'
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
          <p className="text-va-text-muted">Loading roadmap...</p>
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
            Make sure the project has roadmap data configured.
          </p>
          <button
            onClick={loadRoadmap}
            className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  // No roadmap available
  if (!roadmap || !roadmap.items || roadmap.items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-va-bg">
        <div className="text-center max-w-md">
          <div className="text-5xl mb-4">🗺️</div>
          <h2 className="text-xl font-semibold text-va-text mb-2">No Roadmap</h2>
          <p className="text-va-text-muted text-sm mb-4">
            This project doesn't have a roadmap configured yet.
            Add roadmap items to track planned features and initiatives.
          </p>
          <div className="text-xs text-va-text-muted">
            <code>roadmap/roadmap.json</code>
          </div>
        </div>
      </div>
    )
  }

  // Group items by quarter
  const itemsByQuarter = groupByQuarter(roadmap.items)
  const quarters = Object.keys(itemsByQuarter).sort()

  return (
    <div className="h-full flex flex-col bg-va-bg">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-va-border bg-va-sidebar">
        <div className="flex items-center gap-4">
          <h2 className="text-lg font-semibold text-va-text">Roadmap</h2>
          <span className="text-sm text-va-text-muted">
            {roadmap.items.length} items
          </span>
        </div>

        <div className="flex items-center gap-2">
          {/* View toggle */}
          <div className="flex bg-va-panel rounded overflow-hidden">
            <button
              onClick={() => setViewMode('board')}
              className={`px-3 py-1 text-sm ${
                viewMode === 'board'
                  ? 'bg-va-accent text-white'
                  : 'text-va-text-muted hover:text-va-text'
              }`}
            >
              Board
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`px-3 py-1 text-sm ${
                viewMode === 'list'
                  ? 'bg-va-accent text-white'
                  : 'text-va-text-muted hover:text-va-text'
              }`}
            >
              List
            </button>
          </div>

          <button
            onClick={loadRoadmap}
            disabled={isLoading}
            className="p-1.5 rounded hover:bg-va-panel text-va-text-muted hover:text-va-text disabled:opacity-50 transition-colors"
            title="Refresh"
          >
            <RefreshIcon className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-4">
        {viewMode === 'board' ? (
          <BoardView quarters={quarters} itemsByQuarter={itemsByQuarter} />
        ) : (
          <ListView items={roadmap.items} />
        )}
      </div>
    </div>
  )
}

function groupByQuarter(items: RoadmapItem[]): Record<string, RoadmapItem[]> {
  const grouped: Record<string, RoadmapItem[]> = {}
  for (const item of items) {
    const quarter = item.quarter || 'Unscheduled'
    if (!grouped[quarter]) {
      grouped[quarter] = []
    }
    grouped[quarter].push(item)
  }
  return grouped
}

interface BoardViewProps {
  quarters: string[]
  itemsByQuarter: Record<string, RoadmapItem[]>
}

function BoardView({ quarters, itemsByQuarter }: BoardViewProps) {
  return (
    <div className="flex gap-4 overflow-x-auto pb-4" style={{ minHeight: '100%' }}>
      {quarters.map((quarter) => (
        <div key={quarter} className="flex-shrink-0 w-80">
          <div className="bg-va-sidebar border border-va-border rounded-t px-3 py-2">
            <h3 className="font-semibold text-va-text">{quarter}</h3>
            <span className="text-xs text-va-text-muted">
              {itemsByQuarter[quarter].length} items
            </span>
          </div>
          <div className="bg-va-panel/50 border border-t-0 border-va-border rounded-b p-2 space-y-2 min-h-[200px]">
            {itemsByQuarter[quarter].map((item) => (
              <RoadmapCard key={item.id} item={item} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

interface ListViewProps {
  items: RoadmapItem[]
}

function ListView({ items }: ListViewProps) {
  return (
    <div className="space-y-2">
      {items.map((item) => (
        <RoadmapRow key={item.id} item={item} />
      ))}
    </div>
  )
}

function RoadmapCard({ item }: { item: RoadmapItem }) {
  return (
    <div className="bg-va-panel border border-va-border rounded p-3">
      <div className="flex items-start justify-between gap-2 mb-2">
        <h4 className="font-medium text-va-text text-sm">{item.title}</h4>
        <StatusBadge status={item.status} />
      </div>
      {item.description && (
        <p className="text-xs text-va-text-muted line-clamp-2 mb-2">{item.description}</p>
      )}
      <div className="flex items-center gap-2 flex-wrap">
        {item.priority && <PriorityBadge priority={item.priority} />}
        {item.effort && <EffortBadge effort={item.effort} />}
        {item.rice && <RICEScore rice={item.rice} />}
      </div>
      {(item.capability_refs?.length || item.goal_refs?.length) && (
        <div className="mt-2 pt-2 border-t border-va-border flex gap-2 flex-wrap">
          {item.capability_refs?.map((ref) => (
            <span key={ref} className="text-xs bg-va-accent/20 text-va-accent px-1.5 py-0.5 rounded">
              {ref}
            </span>
          ))}
          {item.goal_refs?.map((ref) => (
            <span key={ref} className="text-xs bg-va-success/20 text-va-success px-1.5 py-0.5 rounded">
              {ref}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

function RoadmapRow({ item }: { item: RoadmapItem }) {
  return (
    <div className="bg-va-panel border border-va-border rounded p-3 flex items-center gap-4">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <h4 className="font-medium text-va-text truncate">{item.title}</h4>
          <StatusBadge status={item.status} />
        </div>
        {item.description && (
          <p className="text-sm text-va-text-muted truncate mt-1">{item.description}</p>
        )}
      </div>
      <div className="flex items-center gap-2 flex-shrink-0">
        {item.quarter && (
          <span className="text-xs bg-va-panel border border-va-border px-2 py-1 rounded">
            {item.quarter}
          </span>
        )}
        {item.priority && <PriorityBadge priority={item.priority} />}
        {item.effort && <EffortBadge effort={item.effort} />}
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const getColor = () => {
    switch (status.toLowerCase()) {
      case 'completed':
      case 'done':
        return 'bg-va-success/20 text-va-success'
      case 'in_progress':
      case 'in-progress':
      case 'active':
        return 'bg-va-warning/20 text-va-warning'
      case 'blocked':
        return 'bg-va-error/20 text-va-error'
      default:
        return 'bg-va-text-muted/20 text-va-text-muted'
    }
  }

  return (
    <span className={`text-xs px-1.5 py-0.5 rounded ${getColor()}`}>
      {status}
    </span>
  )
}

function PriorityBadge({ priority }: { priority: string }) {
  const getColor = () => {
    switch (priority.toLowerCase()) {
      case 'critical':
      case 'highest':
        return 'text-va-error'
      case 'high':
        return 'text-va-warning'
      case 'medium':
        return 'text-va-text'
      default:
        return 'text-va-text-muted'
    }
  }

  return (
    <span className={`text-xs ${getColor()}`}>
      P: {priority}
    </span>
  )
}

function EffortBadge({ effort }: { effort: string }) {
  return (
    <span className="text-xs text-va-text-muted">
      {effort}
    </span>
  )
}

function RICEScore({ rice }: { rice: Record<string, unknown> }) {
  const reach = rice.reach as number
  const impact = rice.impact as number
  const confidence = rice.confidence as number
  const effort = rice.effort as number

  if (reach === undefined || impact === undefined || confidence === undefined || effort === undefined) {
    return null
  }

  const score = Math.round((reach * impact * confidence) / effort)

  return (
    <span className="text-xs text-va-accent font-mono" title={`R:${reach} I:${impact} C:${confidence} E:${effort}`}>
      RICE: {score}
    </span>
  )
}

function RefreshIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
      />
    </svg>
  )
}
