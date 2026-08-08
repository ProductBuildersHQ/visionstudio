import { useMemo } from 'react'
import type { ExecutionResponse, APIRepository } from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { PieChart } from '../components/charts'

interface RepositoriesPanelProps {
  execution: ExecutionResponse
  onRepositoryClick: (id: string) => void
}

export function RepositoriesPanel({ execution, onRepositoryClick }: RepositoriesPanelProps) {
  const repositories = execution.repositories
  const rmis = execution.rmis

  const orgGroups = useMemo(() => {
    const groups: Record<string, APIRepository[]> = {}
    for (const repo of repositories) {
      const org = repo.organization || 'Unassigned'
      if (!groups[org]) groups[org] = []
      groups[org].push(repo)
    }
    return Object.entries(groups).sort((a, b) => a[0].localeCompare(b[0]))
  }, [repositories])

  const totalRmis = rmis.length

  const statusDist = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rmis) {
      const s = r.status.toLowerCase()
      counts[s] = (counts[s] ?? 0) + 1
    }
    return Object.entries(counts).map(([name, value]) => ({ name, value }))
  }, [rmis])

  if (repositories.length === 0) {
    return (
      <div className="text-center text-gray-400 py-12">
        <p>No repositories found</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Repositories</h1>
          <p className="text-gray-400 text-sm mt-1">
            {repositories.length} repositor{repositories.length !== 1 ? 'ies' : 'y'}, {totalRmis} RMIs
          </p>
        </div>
        {statusDist.length > 0 && (
          <PieChart data={statusDist} size={120} />
        )}
      </div>

      {/* Repository Grid by Organization */}
      {orgGroups.map(([org, repos]) => (
        <div key={org} className="space-y-3">
          <h2 className="text-lg font-medium text-gray-300">{org}</h2>
          <div className="grid gap-3 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
            {repos.map((repo) => {
              const repoRMIs = rmis.filter((r) => r.repositoryId === repo.id)
              const completedCount = repoRMIs.filter((r) =>
                ['completed', 'released', 'done'].includes(r.status.toLowerCase())
              ).length
              const progress = repoRMIs.length > 0 ? completedCount / repoRMIs.length : 0

              return (
                <RepositoryCard
                  key={repo.id}
                  repository={repo}
                  rmiCount={repoRMIs.length}
                  progress={progress}
                  onClick={() => onRepositoryClick(repo.id)}
                />
              )
            })}
          </div>
        </div>
      ))}
    </div>
  )
}

function RepositoryCard({
  repository,
  rmiCount,
  progress,
  onClick,
}: {
  repository: APIRepository
  rmiCount: number
  progress: number
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className="bg-gray-800 rounded-lg p-4 hover:bg-gray-750 transition-colors text-left w-full border border-gray-700 hover:border-gray-600"
    >
      <div className="flex items-start justify-between mb-2">
        <div className="flex-1 min-w-0">
          <h3 className="font-medium text-white truncate">{repository.repositoryName}</h3>
          <p className="text-xs text-gray-500">{repository.organization}</p>
        </div>
        <StatusBadge status={repository.status} size="sm" />
      </div>

      {repository.domain && (
        <p className="text-xs text-gray-400 mb-2">{repository.domain}</p>
      )}

      <div className="flex items-center justify-between text-sm">
        <span className="text-gray-400">{rmiCount} RMIs</span>
        <span className="text-gray-400">{Math.round(progress * 100)}%</span>
      </div>

      {/* Progress bar */}
      <div className="mt-2 h-1 bg-gray-700 rounded-full overflow-hidden">
        <div
          className="h-full bg-blue-500 rounded-full transition-all duration-300"
          style={{ width: `${progress * 100}%` }}
        />
      </div>
    </button>
  )
}
