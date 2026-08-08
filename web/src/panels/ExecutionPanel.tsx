import { useState, useEffect, useMemo } from 'react'
import { getExecution } from '../api/client'
import type {
  ExecutionResponse,
  APIInitiative,
  APIPhase,
  APIRMI,
  APIRMIDependency,
  APIInitiativeDependency,
} from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { ProgressBar } from '../components/ProgressBar'
import { PieChart } from '../components/charts'
import { LoadingState, ErrorState, EmptyState } from '../components'

export function ExecutionPanel() {
  const [data, setData] = useState<ExecutionResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedInitiative, setSelectedInitiative] = useState<string | null>(null)

  const reload = () => {
    setError(null)
    getExecution()
      .then((d) => {
        setData(d)
        if (!selectedInitiative && d.initiatives.length > 0) {
          setSelectedInitiative(d.initiatives[0].id)
        }
      })
      .catch((err: Error) => setError(err.message))
  }

  useEffect(() => {
    reload()
  }, [])

  if (error) {
    return <ErrorState message={error} onRetry={reload} />
  }

  if (!data) {
    return <LoadingState message="Loading execution data..." />
  }

  if (data.initiatives.length === 0) {
    return (
      <EmptyState
        title="No initiatives found"
        description="Create an initiative to get started."
        hint="prismctl initiative create INIT-XXX-001"
      />
    )
  }

  const selectedInit = data.initiatives.find((i) => i.id === selectedInitiative)
  const phases = selectedInit
    ? data.phases.filter((p) => p.initiativeId === selectedInit.id)
    : []
  const rmis = selectedInit
    ? data.rmis.filter((r) => r.initiativeId === selectedInit.id)
    : []
  const rmiDeps = selectedInit
    ? (data.rmiDependencies ?? []).filter((d) =>
        rmis.some((r) => r.id === d.sourceRmiId || r.id === d.targetRmiId)
      )
    : []
  const initDeps = selectedInit
    ? (data.initiativeDependencies ?? []).filter(
        (d) =>
          d.sourceInitiativeId === selectedInit.id ||
          d.targetInitiativeId === selectedInit.id
      )
    : []

  return (
    <div className="grid grid-cols-12 gap-6">
      {/* Programs & Initiatives List */}
      <div className="col-span-4 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Initiatives</h2>
          <StatusDistributionMini rmis={data.rmis} />
        </div>
        <div className="space-y-3 max-h-[calc(100vh-200px)] overflow-y-auto pr-2">
          {data.programs.map((prog) => {
            const progInits = data.initiatives.filter((i) => i.programId === prog.id)
            if (progInits.length === 0) return null
            return (
              <div key={prog.id} className="space-y-2">
                <h3 className="text-sm font-medium text-gray-400 flex items-center gap-2">
                  <span className="h-1 w-1 rounded-full bg-gray-500" />
                  {prog.name}
                </h3>
                {progInits.map((init) => (
                  <InitiativeCard
                    key={init.id}
                    initiative={init}
                    selected={selectedInitiative === init.id}
                    onClick={() => setSelectedInitiative(init.id)}
                  />
                ))}
              </div>
            )
          })}
          {/* Standalone initiatives */}
          {data.initiatives.filter((i) => !i.programId).length > 0 && (
            <div className="space-y-2">
              <h3 className="text-sm font-medium text-gray-400 flex items-center gap-2">
                <span className="h-1 w-1 rounded-full bg-gray-500" />
                Standalone
              </h3>
              {data.initiatives
                .filter((i) => !i.programId)
                .map((init) => (
                  <InitiativeCard
                    key={init.id}
                    initiative={init}
                    selected={selectedInitiative === init.id}
                    onClick={() => setSelectedInitiative(init.id)}
                  />
                ))}
            </div>
          )}
        </div>
      </div>

      {/* Detail Panel */}
      <div className="col-span-8">
        {selectedInit ? (
          <InitiativeDetail
            initiative={selectedInit}
            phases={phases}
            rmis={rmis}
            rmiDeps={rmiDeps}
            initDeps={initDeps}
            allInitiatives={data.initiatives}
          />
        ) : (
          <div className="text-gray-400">Select an initiative to view details</div>
        )}
      </div>
    </div>
  )
}

function StatusDistributionMini({
  rmis,
}: {
  rmis: APIRMI[]
}) {
  const statusDist = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rmis) {
      const s = r.status.toLowerCase()
      counts[s] = (counts[s] ?? 0) + 1
    }
    return Object.entries(counts).map(([status, count]) => ({ status, count }))
  }, [rmis])

  const total = statusDist.reduce((sum, s) => sum + s.count, 0)
  if (total === 0) return null

  return (
    <div className="flex items-center gap-1 text-xs">
      {statusDist.map((s) => (
        <span
          key={s.status}
          className="px-1.5 py-0.5 rounded text-gray-300"
          style={{ backgroundColor: statusColor(s.status) + '33' }}
          title={`${s.status}: ${s.count}`}
        >
          {s.count}
        </span>
      ))}
    </div>
  )
}

function InitiativeCard({
  initiative,
  selected,
  onClick,
}: {
  initiative: APIInitiative
  selected: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left p-3 rounded-lg border transition-colors ${
        selected
          ? 'border-blue-500 bg-blue-500/10'
          : 'border-gray-700 bg-gray-800 hover:border-gray-600'
      }`}
    >
      <div className="flex items-center justify-between">
        <span className="font-mono text-xs text-gray-400">{initiative.id}</span>
        <StatusBadge status={initiative.status} size="sm" />
      </div>
      <div className="text-sm font-medium mt-1 truncate">{initiative.title}</div>
      <ProgressBar progress={initiative.progress} className="mt-2" size="sm" />
    </button>
  )
}

function InitiativeDetail({
  initiative,
  phases,
  rmis,
  rmiDeps,
  initDeps,
  allInitiatives,
}: {
  initiative: APIInitiative
  phases: APIPhase[]
  rmis: APIRMI[]
  rmiDeps: APIRMIDependency[]
  initDeps: APIInitiativeDependency[]
  allInitiatives: APIInitiative[]
}) {
  const sortedPhases = useMemo(
    () => [...phases].sort((a, b) => a.sequenceNumber - b.sequenceNumber),
    [phases]
  )

  const repoStats = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rmis) {
      if (r.repositoryId) {
        const short = r.repositoryId.split('/').pop() ?? r.repositoryId
        counts[short] = (counts[short] ?? 0) + 1
      }
    }
    return Object.entries(counts)
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => b.count - a.count)
  }, [rmis])

  const statusDist = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rmis) {
      const s = r.status.toLowerCase()
      counts[s] = (counts[s] ?? 0) + 1
    }
    return Object.entries(counts).map(([status, count]) => ({
      name: status,
      value: count,
    }))
  }, [rmis])

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h2 className="text-xl font-semibold">{initiative.id}</h2>
            <StatusBadge status={initiative.status} />
          </div>
          <p className="text-gray-300 mt-1">{initiative.title}</p>
          {initiative.description && (
            <p className="text-gray-500 text-sm mt-2 max-w-2xl">{initiative.description}</p>
          )}
        </div>
        <div className="text-right">
          <div className="text-2xl font-semibold">{Math.round(initiative.progress * 100)}%</div>
          <div className="text-sm text-gray-400">complete</div>
        </div>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="text-sm text-gray-400">Phases</div>
          <div className="text-2xl font-semibold mt-1">{phases.length}</div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="text-sm text-gray-400">RMIs</div>
          <div className="text-2xl font-semibold mt-1">{rmis.length}</div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="text-sm text-gray-400">Repos</div>
          <div className="text-2xl font-semibold mt-1">{repoStats.length}</div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="text-sm text-gray-400">Status</div>
          <div className="h-16 mt-1">
            {statusDist.length > 0 && <PieChart data={statusDist} size={64} showLegend={false} />}
          </div>
        </div>
      </div>

      {/* Dependencies */}
      {initDeps.length > 0 && (
        <div className="bg-gray-800 rounded-lg p-4">
          <h3 className="font-medium mb-2">Initiative Dependencies</h3>
          <div className="flex flex-wrap gap-2">
            {initDeps.map((d, i) => {
              const isSource = d.sourceInitiativeId === initiative.id
              const otherId = isSource ? d.targetInitiativeId : d.sourceInitiativeId
              const other = allInitiatives.find((init) => init.id === otherId)
              return (
                <span
                  key={i}
                  className="text-xs px-2 py-1 bg-gray-700 rounded flex items-center gap-1"
                >
                  {isSource ? (
                    <>
                      <span className="text-gray-400">requires</span>
                      <span className="font-mono">{otherId}</span>
                      {other && <span className="text-gray-500">({other.title})</span>}
                    </>
                  ) : (
                    <>
                      <span className="font-mono">{otherId}</span>
                      <span className="text-gray-400">requires this</span>
                    </>
                  )}
                </span>
              )
            })}
          </div>
        </div>
      )}

      {/* Repos */}
      {repoStats.length > 1 && (
        <div className="flex flex-wrap gap-2">
          {repoStats.map((r) => (
            <span
              key={r.name}
              className="text-xs px-2 py-1 bg-gray-800 border border-gray-700 rounded"
            >
              {r.name} <span className="text-gray-500">({r.count})</span>
            </span>
          ))}
        </div>
      )}

      {/* Phases */}
      <div className="space-y-4">
        {sortedPhases.map((phase) => {
          const phaseRmis = rmis
            .filter((r) => r.phaseId === phase.id)
            .sort((a, b) => a.sequenceNumber - b.sequenceNumber)
          const phaseDeps = rmiDeps.filter((d) =>
            phaseRmis.some((r) => r.id === d.sourceRmiId || r.id === d.targetRmiId)
          )

          return (
            <PhaseCard
              key={phase.id}
              phase={phase}
              rmis={phaseRmis}
              deps={phaseDeps}
              allRmis={rmis}
            />
          )
        })}
      </div>
    </div>
  )
}

function PhaseCard({
  phase,
  rmis,
  deps,
  allRmis,
}: {
  phase: APIPhase
  rmis: APIRMI[]
  deps: APIRMIDependency[]
  allRmis: APIRMI[]
}) {
  const [expanded, setExpanded] = useState(true)

  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between p-4 hover:bg-gray-750 transition-colors"
      >
        <div className="flex items-center gap-3">
          <span className="text-gray-500">{expanded ? '▼' : '▶'}</span>
          <h3 className="font-medium">{phase.title}</h3>
          <span className="text-sm text-gray-500">({rmis.length} RMIs)</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-400">{Math.round(phase.progress * 100)}%</span>
          <ProgressBar progress={phase.progress} className="w-24" size="sm" />
        </div>
      </button>
      {expanded && (
        <div className="border-t border-gray-700 p-4 space-y-2">
          {rmis.map((rmi) => (
            <RMIRow key={rmi.id} rmi={rmi} deps={deps} allRmis={allRmis} />
          ))}
        </div>
      )}
    </div>
  )
}

function RMIRow({
  rmi,
  deps,
  allRmis,
}: {
  rmi: APIRMI
  deps: APIRMIDependency[]
  allRmis: APIRMI[]
}) {
  const myDeps = deps.filter((d) => d.sourceRmiId === rmi.id)
  const depTitles = myDeps
    .map((d) => allRmis.find((r) => r.id === d.targetRmiId)?.id)
    .filter(Boolean)

  return (
    <div className="flex items-center justify-between py-2 px-3 bg-gray-900 rounded hover:bg-gray-850 transition-colors">
      <div className="flex items-center gap-3 min-w-0">
        <span className="text-lg" title={rmi.type ?? 'item'}>
          {typeIcon(rmi.type)}
        </span>
        <span className="text-xs font-mono text-gray-500 flex-shrink-0">{rmi.id}</span>
        <span className="text-sm truncate">{rmi.title}</span>
        {depTitles.length > 0 && (
          <span className="text-xs text-gray-500 flex-shrink-0" title={`Requires: ${depTitles.join(', ')}`}>
            → {depTitles.length}
          </span>
        )}
      </div>
      <div className="flex items-center gap-3 flex-shrink-0">
        {rmi.claimedBy && (
          <span className="text-xs text-gray-500" title={`Claimed: ${rmi.claimedAt}`}>
            {rmi.claimedBy}
          </span>
        )}
        {rmi.completedAt && (
          <span className="text-xs text-gray-500">
            {new Date(rmi.completedAt).toLocaleDateString()}
          </span>
        )}
        <StatusBadge status={rmi.status} size="sm" />
      </div>
    </div>
  )
}

function typeIcon(itemType?: string): string {
  switch (itemType?.toLowerCase()) {
    case 'capability':
      return '★'
    case 'refactor':
      return '↺'
    case 'quality':
      return '✓'
    case 'fix':
      return '⚠'
    case 'chore':
      return '⚙'
    case 'spike':
      return '⚡'
    default:
      return '•'
  }
}

function statusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'completed':
    case 'delivery_complete':
    case 'released':
    case 'closed':
      return '#22c55e'
    case 'executing':
    case 'in_progress':
      return '#3b82f6'
    case 'ready':
      return '#f59e0b'
    case 'planned':
      return '#8b5cf6'
    case 'proposed':
      return '#6b7280'
    case 'cancelled':
      return '#ef4444'
    default:
      return '#6b7280'
  }
}
