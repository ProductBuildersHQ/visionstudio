import { useState, useEffect, useMemo } from 'react'
import { getExecution, getSpecs } from '../api/client'
import type {
  ExecutionResponse,
  SpecsResponse,
  APIProgram,
  APIInitiative,
  APIPhase,
  APIRMI,
  APIRMIDependency,
  JudgeResult,
} from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { ProgressBar } from '../components/ProgressBar'
import { PieChart } from '../components/charts'
import { LoadingState, ErrorState, EmptyState } from '../components'

type ViewMode = 'overview' | 'detail'

export function ProgramsPanel() {
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [specs, setSpecs] = useState<SpecsResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>('overview')
  const [selectedInitiative, setSelectedInitiative] = useState<string | null>(null)

  const reload = () => {
    setError(null)
    Promise.all([getExecution(), getSpecs()])
      .then(([e, s]) => {
        setExecution(e)
        setSpecs(s)
      })
      .catch((err: Error) => setError(err.message))
  }

  useEffect(() => {
    reload()
  }, [])

  const handleInitiativeClick = (id: string) => {
    setSelectedInitiative(id)
    setViewMode('detail')
  }

  const handleBackToOverview = () => {
    setViewMode('overview')
    setSelectedInitiative(null)
  }

  if (error) {
    return <ErrorState message={error} onRetry={reload} />
  }

  if (!execution || !specs) {
    return <LoadingState message="Loading programs data..." />
  }

  if (execution.initiatives.length === 0) {
    return (
      <EmptyState
        title="No initiatives found"
        description="Create an initiative to get started."
        hint="visionstudio initiative create INIT-XXX-001"
      />
    )
  }

  if (viewMode === 'detail' && selectedInitiative) {
    const initiative = execution.initiatives.find((i) => i.id === selectedInitiative)
    if (initiative) {
      return (
        <InitiativeDetailView
          initiative={initiative}
          execution={execution}
          specs={specs}
          onBack={handleBackToOverview}
        />
      )
    }
  }

  return (
    <OverviewView
      execution={execution}
      onInitiativeClick={handleInitiativeClick}
    />
  )
}

function OverviewView({
  execution,
  onInitiativeClick,
}: {
  execution: ExecutionResponse
  onInitiativeClick: (id: string) => void
}) {
  const { programs, initiatives } = execution

  const statusDist = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of execution.rmis) {
      const s = r.status.toLowerCase()
      counts[s] = (counts[s] ?? 0) + 1
    }
    return Object.entries(counts).map(([name, value]) => ({ name, value }))
  }, [execution.rmis])

  const standalone = initiatives.filter((i) => !i.programId)

  return (
    <div className="space-y-8">
      {/* Summary Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold">Programs & Initiatives</h2>
          <p className="text-gray-400 text-sm mt-1">
            {programs.length} programs, {initiatives.length} initiatives, {execution.rmis.length} RMIs
          </p>
        </div>
        <div className="flex items-center gap-4">
          <div className="w-16 h-16">
            {statusDist.length > 0 && (
              <PieChart data={statusDist} size={64} showLegend={false} innerRadius={20} />
            )}
          </div>
          <div className="text-sm space-y-1">
            {statusDist.slice(0, 4).map((s) => (
              <div key={s.name} className="flex items-center gap-2">
                <span className="text-gray-400">{s.name}:</span>
                <span className="text-gray-200">{s.value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Programs */}
      {programs.filter((p) => !p.hidden).map((program) => (
        <ProgramSection
          key={program.id}
          program={program}
          initiatives={initiatives.filter((i) => i.programId === program.id)}
          onInitiativeClick={onInitiativeClick}
        />
      ))}

      {/* Standalone Initiatives */}
      {standalone.length > 0 && (
        <div className="space-y-4">
          <h3 className="text-lg font-medium text-gray-300">Standalone Initiatives</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {standalone.map((init) => (
              <InitiativeTile
                key={init.id}
                initiative={init}
                onClick={() => onInitiativeClick(init.id)}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function ProgramSection({
  program,
  initiatives,
  onInitiativeClick,
}: {
  program: APIProgram
  initiatives: APIInitiative[]
  onInitiativeClick: (id: string) => void
}) {
  if (initiatives.length === 0) return null

  const totalProgress = initiatives.reduce((sum, i) => sum + i.progress, 0) / initiatives.length

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium text-gray-200">{program.name}</h3>
          {program.description && (
            <p className="text-sm text-gray-500">{program.description}</p>
          )}
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm text-gray-400">{initiatives.length} initiatives</span>
          <div className="w-24">
            <ProgressBar progress={totalProgress} size="sm" />
          </div>
          <span className="text-sm text-gray-300">{Math.round(totalProgress * 100)}%</span>
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {initiatives.map((init) => (
          <InitiativeTile
            key={init.id}
            initiative={init}
            onClick={() => onInitiativeClick(init.id)}
          />
        ))}
      </div>
    </div>
  )
}

function InitiativeTile({
  initiative,
  onClick,
}: {
  initiative: APIInitiative
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className="w-full text-left p-4 rounded-lg border border-gray-700 bg-gray-800 hover:border-gray-600 hover:bg-gray-750 transition-colors"
    >
      <div className="flex items-center justify-between mb-2">
        <span className="font-mono text-xs text-gray-500">{initiative.id}</span>
        <StatusBadge status={initiative.status} size="sm" />
      </div>
      <h4 className="font-medium text-gray-100 mb-2 line-clamp-2">{initiative.title}</h4>
      <div className="flex items-center justify-between">
        <ProgressBar progress={initiative.progress} className="flex-1 mr-3" size="sm" />
        <span className="text-sm text-gray-400">{Math.round(initiative.progress * 100)}%</span>
      </div>
    </button>
  )
}

function InitiativeDetailView({
  initiative,
  execution,
  specs,
  onBack,
}: {
  initiative: APIInitiative
  execution: ExecutionResponse
  specs: SpecsResponse
  onBack: () => void
}) {
  const phases = execution.phases.filter((p) => p.initiativeId === initiative.id)
  const rmis = execution.rmis.filter((r) => r.initiativeId === initiative.id)
  const rmiDeps = (execution.rmiDependencies ?? []).filter((d) =>
    rmis.some((r) => r.id === d.sourceRmiId || r.id === d.targetRmiId)
  )
  const initDeps = (execution.initiativeDependencies ?? []).filter(
    (d) => d.sourceInitiativeId === initiative.id || d.targetInitiativeId === initiative.id
  )
  const judgeResults = (specs.judgeResults ?? []).filter((r) => r.initiativeId === initiative.id)

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
    return Object.entries(counts).map(([name, value]) => ({ name, value }))
  }, [rmis])

  return (
    <div className="space-y-6">
      {/* Back Button + Header */}
      <div>
        <button
          onClick={onBack}
          className="text-sm text-gray-400 hover:text-gray-200 mb-4 flex items-center gap-1"
        >
          <span>←</span> Back to Overview
        </button>
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

      {/* Initiative Dependencies */}
      {initDeps.length > 0 && (
        <div className="bg-gray-800 rounded-lg p-4">
          <h3 className="font-medium mb-2">Initiative Dependencies</h3>
          <div className="flex flex-wrap gap-2">
            {initDeps.map((d, i) => {
              const isSource = d.sourceInitiativeId === initiative.id
              const otherId = isSource ? d.targetInitiativeId : d.sourceInitiativeId
              const other = execution.initiatives.find((init) => init.id === otherId)
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

      {/* Specs Section */}
      {judgeResults.length > 0 && (
        <SpecsSection judgeResults={judgeResults} />
      )}

      {/* Phases */}
      <div className="space-y-4">
        <h3 className="text-lg font-medium">Execution Phases</h3>
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

function SpecsSection({ judgeResults }: { judgeResults: JudgeResult[] }) {
  const getScore = (r: JudgeResult): number => r.report?.intScore ?? 0
  const avgScore = judgeResults.length > 0
    ? judgeResults.reduce((sum, r) => sum + getScore(r), 0) / judgeResults.length
    : 0

  const specTypes = [...new Set(judgeResults.map((r) => specType(r.specPath)))]

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="font-medium text-purple-400">Spec Quality</h3>
        <span className={`text-lg font-semibold ${
          avgScore >= 4 ? 'text-green-400' : avgScore >= 3 ? 'text-yellow-400' : 'text-red-400'
        }`}>
          {avgScore.toFixed(1)}/5 avg
        </span>
      </div>
      <div className="flex flex-wrap gap-2 mb-3">
        {specTypes.map((type) => (
          <span key={type} className="text-xs px-2 py-1 bg-purple-500/20 text-purple-300 rounded">
            {type}
          </span>
        ))}
      </div>
      <div className="space-y-2 max-h-48 overflow-y-auto">
        {judgeResults
          .sort((a, b) => new Date(b.evaluatedAt ?? '').getTime() - new Date(a.evaluatedAt ?? '').getTime())
          .map((r) => {
            const score = getScore(r)
            return (
              <div key={r.id} className="flex justify-between text-sm">
                <span className="text-gray-300">{(r.specPath ?? '').split('/').pop()}</span>
                <span className={`px-2 py-0.5 rounded ${
                  score >= 4 ? 'bg-green-500/30 text-green-300' :
                  score >= 3 ? 'bg-yellow-500/30 text-yellow-300' :
                  'bg-red-500/30 text-red-300'
                }`}>
                  {score}/5
                </span>
              </div>
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
          <h4 className="font-medium">{phase.title}</h4>
          <span className="text-sm text-gray-500">({rmis.length} RMIs)</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-400">{Math.round((phase.progress ?? 0) * 100)}%</span>
          <ProgressBar progress={phase.progress ?? 0} className="w-24" size="sm" />
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
        <StatusBadge status={rmi.status ?? ''} size="sm" />
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

function specType(path: string): string {
  const name = path.split('/').pop()?.toLowerCase() ?? ''
  if (name.includes('prd')) return 'PRD'
  if (name.includes('trd')) return 'TRD'
  if (name.includes('plan')) return 'PLAN'
  if (name.includes('roadmap')) return 'ROADMAP'
  return name.replace('.md', '').toUpperCase()
}
