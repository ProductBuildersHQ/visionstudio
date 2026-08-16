import { useMemo } from 'react'
import type { APIInitiative, APIProgram } from '../api/types'
import { ProgressBar } from '../components/ProgressBar'
import { visibleInitiatives } from '../lib/visibility'

// Pipeline order, matching pkg/initiative's lifecycle statuses.
const STATUS_ORDER = [
  'proposed',
  'planned',
  'executing',
  'delivery_complete',
  'releasing',
  'released',
  'closed',
  'cancelled',
] as const

const STATUS_LABELS: Record<string, string> = {
  proposed: 'Proposed',
  planned: 'Planned',
  executing: 'Executing',
  delivery_complete: 'Delivery Complete',
  releasing: 'Releasing',
  released: 'Released',
  closed: 'Closed',
  cancelled: 'Cancelled',
}

interface StatusBoardProps {
  initiatives: APIInitiative[]
  programs: APIProgram[]
  onInitiativeClick: (id: string) => void
}

export function StatusBoard({ initiatives: allInitiatives, programs, onInitiativeClick }: StatusBoardProps) {
  const initiatives = useMemo(
    () => visibleInitiatives(allInitiatives, programs),
    [allInitiatives, programs]
  )

  const byStatus = useMemo(() => {
    const groups = new Map<string, APIInitiative[]>()
    for (const status of STATUS_ORDER) groups.set(status, [])
    for (const init of initiatives) {
      const key = groups.has(init.status) ? init.status : 'proposed'
      groups.get(key)!.push(init)
    }
    for (const items of groups.values()) {
      items.sort((a, b) => b.progress - a.progress)
    }
    return groups
  }, [initiatives])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Portfolio by Status</h1>
        <p className="text-gray-400 text-sm mt-1">
          {initiatives.length} initiative{initiatives.length !== 1 ? 's' : ''} across the pipeline
        </p>
      </div>

      <div className="flex gap-4 overflow-x-auto pb-4">
        {STATUS_ORDER.map((status) => {
          const items = byStatus.get(status) ?? []
          return (
            <div key={status} className="flex-shrink-0 w-72">
              <div className="flex items-center justify-between mb-3 px-1">
                <h2 className="text-sm font-medium text-gray-300">{STATUS_LABELS[status]}</h2>
                <span className="text-xs text-gray-500 bg-gray-800 rounded-full px-2 py-0.5">
                  {items.length}
                </span>
              </div>
              <div className="space-y-2">
                {items.map((init) => (
                  <button
                    key={init.id}
                    onClick={() => onInitiativeClick(init.id)}
                    className="w-full text-left p-3 rounded-lg border border-gray-700 bg-gray-800 hover:border-gray-600 hover:bg-gray-750 transition-colors"
                  >
                    <span className="font-mono text-xs text-gray-500">{init.id}</span>
                    <h3 className="text-sm font-medium text-gray-100 mt-1 mb-2 line-clamp-2">
                      {init.title}
                    </h3>
                    <ProgressBar
                      progress={init.progress}
                      cancelledProgress={init.cancelledProgress}
                      size="sm"
                    />
                  </button>
                ))}
                {items.length === 0 && <p className="text-xs text-gray-600 italic px-1">None</p>}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
