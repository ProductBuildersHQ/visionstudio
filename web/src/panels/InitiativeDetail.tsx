import { useState, useMemo } from 'react'
import type {
  ExecutionResponse,
  SpecsResponse,
  APIInitiative,
  APIPhase,
  APIRMI,
  APIRMIDependency,
  JudgeResult,
  SpecWorkflow,
} from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { ProgressBar } from '../components/ProgressBar'
import { PieChart } from '../components/charts'

interface InitiativeDetailProps {
  initiative: APIInitiative
  execution: ExecutionResponse
  specs: SpecsResponse
  onBack: () => void
}

export function InitiativeDetail({
  initiative,
  execution,
  specs,
  onBack,
}: InitiativeDetailProps) {
  const phases = execution.phases.filter((p) => p.initiativeId === initiative.id)
  const rmis = execution.rmis.filter((r) => r.initiativeId === initiative.id)
  const rmiDeps = (execution.rmiDependencies ?? []).filter((d) =>
    rmis.some((r) => r.id === d.sourceRmiId || r.id === d.targetRmiId)
  )
  const initDeps = (execution.initiativeDependencies ?? []).filter(
    (d) => d.sourceInitiativeId === initiative.id || d.targetInitiativeId === initiative.id
  )
  const judgeResults = (specs.judgeResults ?? []).filter((r) => r.initiative_id === initiative.id)

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
              <h1 className="text-xl font-semibold">{initiative.id}</h1>
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

      {/* Two-Column Layout: Definition + Execution Stats */}
      <div className="grid grid-cols-2 gap-6">
        {/* Definition Section */}
        <div className="space-y-4">
          <h2 className="text-lg font-medium text-purple-400">Definition</h2>
          <SpecsSection judgeResults={judgeResults} workflows={specs.workflows ?? []} />
        </div>

        {/* Execution Stats */}
        <div className="space-y-4">
          <h2 className="text-lg font-medium text-blue-400">Execution</h2>
          <div className="grid grid-cols-2 gap-4">
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
              <div className="h-12 mt-1 flex items-center justify-center">
                {statusDist.length > 0 && (
                  <PieChart data={statusDist} size={48} showLegend={false} />
                )}
              </div>
            </div>
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

      {/* Phases */}
      <div className="space-y-4">
        <h2 className="text-lg font-medium">Phases</h2>
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

const PBHQ_LITE_SPECS = ['PRD', 'TRD', 'PLAN', 'ROADMAP'] as const

function SpecsSection({
  judgeResults,
  workflows,
}: {
  judgeResults: JudgeResult[]
  workflows: SpecWorkflow[]
}) {
  const avgScore =
    judgeResults.length > 0
      ? judgeResults.reduce((sum, r) => sum + r.score, 0) / judgeResults.length
      : 0

  const resultsByType = useMemo(() => {
    const map: Record<string, JudgeResult> = {}
    for (const r of judgeResults) {
      const type = specType(r.spec_path)
      const existing = map[type]
      if (!existing || new Date(r.evaluated_at) > new Date(existing.evaluated_at)) {
        map[type] = r
      }
    }
    return map
  }, [judgeResults])

  const pbhqWorkflow = workflows.find((w) => w.name === 'pbhq-lite')

  return (
    <div className="bg-gray-800 rounded-lg p-4 space-y-4">
      {/* Workflow Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-purple-400">PBHQ Lite</span>
          {pbhqWorkflow && (
            <span className="text-xs text-gray-500">PRD → TRD → PLAN → ROADMAP</span>
          )}
        </div>
        {judgeResults.length > 0 && (
          <span
            className={`text-lg font-semibold ${
              avgScore >= 7 ? 'text-green-400' : avgScore >= 4 ? 'text-yellow-400' : 'text-red-400'
            }`}
          >
            {avgScore.toFixed(1)} avg
          </span>
        )}
      </div>

      {/* Workflow Diagram */}
      <div className="flex items-center gap-1">
        {PBHQ_LITE_SPECS.map((spec, i) => {
          const result = resultsByType[spec]
          const hasSpec = !!result
          const score = result?.score ?? 0

          return (
            <div key={spec} className="flex items-center">
              <div
                className={`px-3 py-2 rounded text-xs font-medium border ${
                  hasSpec
                    ? score >= 7
                      ? 'bg-green-500/20 border-green-500/50 text-green-300'
                      : score >= 4
                      ? 'bg-yellow-500/20 border-yellow-500/50 text-yellow-300'
                      : 'bg-red-500/20 border-red-500/50 text-red-300'
                    : 'bg-gray-700 border-gray-600 text-gray-400'
                }`}
                title={hasSpec ? `Score: ${score.toFixed(1)}` : 'Not evaluated'}
              >
                {spec}
                {hasSpec && <span className="ml-1 opacity-70">{score.toFixed(1)}</span>}
              </div>
              {i < PBHQ_LITE_SPECS.length - 1 && (
                <span className="text-gray-600 px-1">→</span>
              )}
            </div>
          )
        })}
      </div>

      {/* Detailed Results */}
      {judgeResults.length > 0 && (
        <div className="space-y-2 max-h-32 overflow-y-auto border-t border-gray-700 pt-3">
          {judgeResults
            .sort((a, b) => new Date(b.evaluated_at).getTime() - new Date(a.evaluated_at).getTime())
            .map((r) => (
              <div key={r.id} className="flex justify-between text-sm">
                <span className="text-gray-300 truncate" title={r.spec_path}>
                  {r.spec_path.split('/').pop()}
                </span>
                <div className="flex items-center gap-2 flex-shrink-0">
                  <span className="text-xs text-gray-500">
                    {new Date(r.evaluated_at).toLocaleDateString()}
                  </span>
                  <span
                    className={`px-2 py-0.5 rounded ${
                      r.score >= 7
                        ? 'bg-green-500/30 text-green-300'
                        : r.score >= 4
                        ? 'bg-yellow-500/30 text-yellow-300'
                        : 'bg-red-500/30 text-red-300'
                    }`}
                  >
                    {r.score.toFixed(1)}
                  </span>
                </div>
              </div>
            ))}
        </div>
      )}
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
          <span
            className="text-xs text-gray-500 flex-shrink-0"
            title={`Requires: ${depTitles.join(', ')}`}
          >
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

function specType(path: string): string {
  const name = path.split('/').pop()?.toLowerCase() ?? ''
  if (name.includes('prd')) return 'PRD'
  if (name.includes('trd')) return 'TRD'
  if (name.includes('plan')) return 'PLAN'
  if (name.includes('roadmap')) return 'ROADMAP'
  return name.replace('.md', '').toUpperCase()
}
