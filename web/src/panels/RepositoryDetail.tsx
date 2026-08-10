import { useMemo } from 'react'
import type { ExecutionResponse, APIRepository, APIRMI, APIInitiative } from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { PieChart } from '../components/charts'
import { visibleInitiatives } from '../lib/visibility'

interface RepositoryDetailProps {
  repository: APIRepository
  execution: ExecutionResponse
  onBack: () => void
  onInitiativeClick: (id: string) => void
}

export function RepositoryDetail({
  repository,
  execution,
  onBack,
  onInitiativeClick,
}: RepositoryDetailProps) {
  const rmis = useMemo(
    () => execution.rmis.filter((r) => r.repositoryId === repository.id),
    [execution.rmis, repository.id]
  )

  const initiatives = useMemo(() => {
    const initIds = new Set(rmis.map((r) => r.initiativeId).filter(Boolean))
    const linked = execution.initiatives.filter((i) => initIds.has(i.id))
    return visibleInitiatives(linked, execution.programs)
  }, [rmis, execution.initiatives, execution.programs])

  const statusDist = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rmis) {
      const s = r.status.toLowerCase()
      counts[s] = (counts[s] ?? 0) + 1
    }
    return Object.entries(counts).map(([name, value]) => ({ name, value }))
  }, [rmis])

  const completedCount = rmis.filter((r) =>
    ['completed', 'released', 'done'].includes(r.status.toLowerCase())
  ).length
  const progress = rmis.length > 0 ? completedCount / rmis.length : 0

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={onBack}
          className="text-gray-400 hover:text-white transition-colors"
        >
          &larr;
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-semibold">{repository.repositoryName}</h1>
            <StatusBadge status={repository.status} />
          </div>
          <p className="text-gray-400 text-sm mt-1">{repository.organization}</p>
        </div>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard label="RMIs" value={rmis.length.toString()} />
        <StatCard label="Initiatives" value={initiatives.length.toString()} />
        <StatCard label="Progress" value={`${Math.round(progress * 100)}%`} />
        <StatCard label="Completed" value={completedCount.toString()} />
      </div>

      {/* Repository info */}
      <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
        <h2 className="text-lg font-medium mb-3">Repository Info</h2>
        <div className="grid grid-cols-2 gap-4 text-sm">
          {repository.goModule && (
            <div>
              <span className="text-gray-400">Go Module:</span>
              <span className="ml-2 text-white font-mono text-xs">{repository.goModule}</span>
            </div>
          )}
          {repository.defaultBranch && (
            <div>
              <span className="text-gray-400">Branch:</span>
              <span className="ml-2 text-white">{repository.defaultBranch}</span>
            </div>
          )}
          {repository.domain && (
            <div>
              <span className="text-gray-400">Domain:</span>
              <span className="ml-2 text-white">{repository.domain}</span>
            </div>
          )}
          {repository.localPath && (
            <div className="col-span-2">
              <span className="text-gray-400">Path:</span>
              <span className="ml-2 text-white font-mono text-xs">{repository.localPath}</span>
            </div>
          )}
        </div>
      </div>

      {/* Two-column layout */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left: Initiatives */}
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <h2 className="text-lg font-medium mb-3">Initiatives ({initiatives.length})</h2>
          {initiatives.length === 0 ? (
            <p className="text-gray-400 text-sm">No initiatives linked</p>
          ) : (
            <div className="space-y-2">
              {initiatives.map((init) => {
                const initRmis = rmis.filter((r) => r.initiativeId === init.id)
                const completedRmis = initRmis.filter((r) =>
                  ['completed', 'released', 'done'].includes(r.status.toLowerCase())
                ).length
                const progress = initRmis.length > 0 ? completedRmis / initRmis.length : 0
                return (
                  <InitiativeRow
                    key={init.id}
                    initiative={init}
                    rmiCount={initRmis.length}
                    completedCount={completedRmis}
                    progress={progress}
                    onClick={() => onInitiativeClick(init.id)}
                  />
                )
              })}
            </div>
          )}
        </div>

        {/* Right: Status distribution */}
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <h2 className="text-lg font-medium mb-3">RMI Status</h2>
          {statusDist.length === 0 ? (
            <p className="text-gray-400 text-sm">No RMIs</p>
          ) : (
            <PieChart data={statusDist} size={140} />
          )}
        </div>
      </div>

      {/* RMIs table */}
      <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
        <h2 className="text-lg font-medium mb-3">Roadmap Items ({rmis.length})</h2>
        {rmis.length === 0 ? (
          <p className="text-gray-400 text-sm">No RMIs in this repository</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-400 border-b border-gray-700">
                  <th className="pb-2 pr-4">ID</th>
                  <th className="pb-2 pr-4">Title</th>
                  <th className="pb-2 pr-4">Initiative</th>
                  <th className="pb-2 pr-4">Status</th>
                  <th className="pb-2 pr-4">Priority</th>
                </tr>
              </thead>
              <tbody>
                {rmis.map((rmi) => (
                  <RMIRow
                    key={rmi.id}
                    rmi={rmi}
                    initiatives={initiatives}
                    onInitiativeClick={onInitiativeClick}
                  />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
      <div className="text-gray-400 text-sm">{label}</div>
      <div className="text-2xl font-semibold mt-1">{value}</div>
    </div>
  )
}

function InitiativeRow({
  initiative,
  rmiCount,
  completedCount,
  progress,
  onClick,
}: {
  initiative: APIInitiative
  rmiCount: number
  completedCount: number
  progress: number
  onClick: () => void
}) {
  const pct = Math.round(progress * 100)
  const barColor = pct >= 100 ? 'bg-green-500' : pct >= 50 ? 'bg-blue-500' : 'bg-yellow-500'

  return (
    <button
      onClick={onClick}
      className="w-full flex items-center gap-3 p-2 rounded hover:bg-gray-700 transition-colors text-left"
    >
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs text-gray-500">{initiative.id}</span>
          <StatusBadge status={initiative.status} size="sm" />
        </div>
        <p className="text-sm text-gray-300 truncate">{initiative.title}</p>
      </div>
      <div className="flex items-center gap-2 min-w-[120px]">
        <div className="w-16 bg-gray-700 rounded-full h-2">
          <div
            className={`${barColor} h-2 rounded-full transition-all`}
            style={{ width: `${pct}%` }}
          />
        </div>
        <span className="text-xs text-gray-400 w-12 text-right">{pct}%</span>
      </div>
      <span className="text-gray-500 text-xs whitespace-nowrap">{completedCount}/{rmiCount}</span>
    </button>
  )
}

function RMIRow({
  rmi,
  initiatives,
  onInitiativeClick,
}: {
  rmi: APIRMI
  initiatives: APIInitiative[]
  onInitiativeClick: (id: string) => void
}) {
  const initiative = initiatives.find((i) => i.id === rmi.initiativeId)

  return (
    <tr className="border-b border-gray-700/50 hover:bg-gray-700/30">
      <td className="py-2 pr-4 font-mono text-xs text-gray-400">{rmi.id}</td>
      <td className="py-2 pr-4 text-gray-300 max-w-xs truncate">{rmi.title}</td>
      <td className="py-2 pr-4">
        {initiative ? (
          <button
            onClick={() => onInitiativeClick(initiative.id)}
            className="text-blue-400 hover:text-blue-300 text-xs font-mono"
          >
            {initiative.id}
          </button>
        ) : (
          <span className="text-gray-500 text-xs">-</span>
        )}
      </td>
      <td className="py-2 pr-4">
        <StatusBadge status={rmi.status} size="sm" />
      </td>
      <td className="py-2 pr-4 text-gray-400">{rmi.priority || '-'}</td>
    </tr>
  )
}
