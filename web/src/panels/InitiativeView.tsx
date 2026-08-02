import { useState, useEffect, useMemo } from 'react'
import { getExecution, getSpend, getMaturity, getSpecs } from '../api/client'
import type {
  ExecutionResponse,
  SpendResponse,
  MaturityResponse,
  SpecsResponse,
  APIInitiative,
  APIPhase,
  APIRMI,
  JudgeResult,
  MaturityAssessment,
} from '../api/types'
import { StatusBadge, ProgressBar, LoadingState, ErrorState, EmptyState } from '../components'
import { PieChart, RadarChart, type RadarAxis, type RadarDataset } from '../components/charts'

export function InitiativeView() {
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [spend, setSpend] = useState<SpendResponse | null>(null)
  const [maturity, setMaturity] = useState<MaturityResponse | null>(null)
  const [specs, setSpecs] = useState<SpecsResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedInitiative, setSelectedInitiative] = useState<string | null>(null)

  const reload = () => {
    setError(null)
    Promise.all([getExecution(), getSpend(), getMaturity(), getSpecs()])
      .then(([e, s, m, sp]) => {
        setExecution(e)
        setSpend(s)
        setMaturity(m)
        setSpecs(sp)
        if (e.initiatives.length > 0 && !selectedInitiative) {
          setSelectedInitiative(e.initiatives[0].id)
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

  if (!execution || !spend || !maturity || !specs) {
    return <LoadingState message="Loading initiative data..." />
  }

  if (execution.initiatives.length === 0) {
    return (
      <EmptyState
        title="No initiatives found"
        description="Create an initiative to get started."
        hint="prismctl initiative create INIT-XXX-001"
      />
    )
  }

  const selectedInit = execution.initiatives.find((i) => i.id === selectedInitiative)

  return (
    <div className="flex gap-6">
      {/* Initiative Selector Sidebar */}
      <div className="w-64 flex-shrink-0 space-y-2 max-h-[calc(100vh-150px)] overflow-y-auto">
        <h3 className="text-sm font-medium text-gray-400 mb-2">Select Initiative</h3>
        {execution.programs.map((prog) => {
          const progInits = execution.initiatives.filter((i) => i.programId === prog.id)
          if (progInits.length === 0) return null
          return (
            <div key={prog.id} className="space-y-1">
              <div className="text-xs text-gray-500 font-medium px-2">{prog.name}</div>
              {progInits.map((init) => (
                <InitiativeButton
                  key={init.id}
                  initiative={init}
                  selected={selectedInitiative === init.id}
                  onClick={() => setSelectedInitiative(init.id)}
                />
              ))}
            </div>
          )
        })}
        {execution.initiatives.filter((i) => !i.programId).length > 0 && (
          <div className="space-y-1">
            <div className="text-xs text-gray-500 font-medium px-2">Standalone</div>
            {execution.initiatives
              .filter((i) => !i.programId)
              .map((init) => (
                <InitiativeButton
                  key={init.id}
                  initiative={init}
                  selected={selectedInitiative === init.id}
                  onClick={() => setSelectedInitiative(init.id)}
                />
              ))}
          </div>
        )}
      </div>

      {/* Main Content */}
      <div className="flex-1 min-w-0">
        {selectedInit ? (
          <ComposedView
            initiative={selectedInit}
            execution={execution}
            spend={spend}
            maturity={maturity}
            specs={specs}
          />
        ) : (
          <div className="text-gray-400">Select an initiative</div>
        )}
      </div>
    </div>
  )
}

function InitiativeButton({
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
      className={`w-full text-left px-2 py-1.5 rounded text-sm transition-colors ${
        selected
          ? 'bg-blue-500/20 text-blue-400'
          : 'text-gray-300 hover:bg-gray-800'
      }`}
    >
      <div className="font-mono text-xs text-gray-500">{initiative.id}</div>
      <div className="truncate">{initiative.title}</div>
    </button>
  )
}

function ComposedView({
  initiative,
  execution,
  spend,
  maturity,
  specs,
}: {
  initiative: APIInitiative
  execution: ExecutionResponse
  spend: SpendResponse
  maturity: MaturityResponse
  specs: SpecsResponse
}) {
  const phases = execution.phases.filter((p) => p.initiativeId === initiative.id)
  const rmis = execution.rmis.filter((r) => r.initiativeId === initiative.id)
  const judgeResults = (specs.judgeResults ?? []).filter((r) => r.initiative_id === initiative.id)
  const assessments = (maturity.assessments ?? []).filter((a) => a.initiative_id === initiative.id)
  const initSpend = spend.byInitiative?.[initiative.id]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-gray-800 rounded-lg p-4">
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
            <div className="text-3xl font-bold">{Math.round(initiative.progress * 100)}%</div>
            <div className="text-sm text-gray-400">complete</div>
          </div>
        </div>
      </div>

      {/* Four-Quadrant Grid */}
      <div className="grid grid-cols-2 gap-4">
        {/* Execution Quadrant */}
        <ExecutionQuadrant phases={phases} rmis={rmis} />

        {/* Spend Quadrant */}
        <SpendQuadrant spend={initSpend} rmis={rmis} byRmi={spend.byRmi} />

        {/* Specs Quadrant */}
        <SpecsQuadrant
          judgeResults={judgeResults}
          workflows={specs.workflows}
        />

        {/* Maturity Quadrant */}
        <MaturityQuadrant
          assessments={assessments}
          models={maturity.models}
        />
      </div>
    </div>
  )
}

function ExecutionQuadrant({
  phases,
  rmis,
}: {
  phases: APIPhase[]
  rmis: APIRMI[]
}) {
  const sortedPhases = useMemo(
    () => [...phases].sort((a, b) => a.sequenceNumber - b.sequenceNumber),
    [phases]
  )

  const statusDist = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rmis) {
      const s = r.status.toLowerCase()
      counts[s] = (counts[s] ?? 0) + 1
    }
    return Object.entries(counts).map(([name, value]) => ({ name, value }))
  }, [rmis])

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-medium text-blue-400">Execution</h3>
        <span className="text-sm text-gray-400">{rmis.length} RMIs</span>
      </div>

      <div className="flex gap-4 mb-4">
        <div className="w-20 h-20">
          {statusDist.length > 0 && (
            <PieChart data={statusDist} size={80} showLegend={false} innerRadius={25} />
          )}
        </div>
        <div className="flex-1 space-y-1 text-xs">
          {statusDist.map((s) => (
            <div key={s.name} className="flex justify-between">
              <span className="text-gray-400">{s.name}</span>
              <span className="text-gray-300">{s.value}</span>
            </div>
          ))}
        </div>
      </div>

      <div className="space-y-2 max-h-48 overflow-y-auto">
        {sortedPhases.map((phase) => {
          const phaseRmis = rmis.filter((r) => r.phaseId === phase.id).length
          const completed = rmis.filter(
            (r) => r.phaseId === phase.id && r.status.toLowerCase() === 'completed'
          ).length
          return (
            <div key={phase.id} className="text-sm">
              <div className="flex justify-between mb-1">
                <span className="text-gray-300 truncate">{phase.title}</span>
                <span className="text-gray-500">{completed}/{phaseRmis}</span>
              </div>
              <ProgressBar progress={phase.progress} size="sm" />
            </div>
          )
        })}
      </div>
    </div>
  )
}

function SpendQuadrant({
  spend,
  rmis,
  byRmi,
}: {
  spend?: { totalTokens: number; costUsd: number }
  rmis: APIRMI[]
  byRmi?: Record<string, { totalTokens: number; costUsd: number }>
}) {
  const formatTokens = (n: number): string => {
    if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
    if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
    return n.toString()
  }

  const formatCost = (usd: number): string => {
    if (usd >= 100) return `$${usd.toFixed(0)}`
    if (usd >= 1) return `$${usd.toFixed(2)}`
    return `$${usd.toFixed(4)}`
  }

  const topRmis = useMemo(() => {
    if (!byRmi) return []
    return rmis
      .filter((r) => byRmi[r.id]?.costUsd > 0)
      .sort((a, b) => (byRmi[b.id]?.costUsd ?? 0) - (byRmi[a.id]?.costUsd ?? 0))
      .slice(0, 5)
  }, [rmis, byRmi])

  if (!spend || spend.totalTokens === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="font-medium text-green-400 mb-4">Spend</h3>
        <div className="text-gray-400 text-sm">No token data available</div>
      </div>
    )
  }

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-medium text-green-400">Spend</h3>
        <span className="text-lg font-semibold text-green-400">{formatCost(spend.costUsd)}</span>
      </div>

      <div className="text-sm text-gray-400 mb-4">
        {formatTokens(spend.totalTokens)} tokens
      </div>

      {topRmis.length > 0 && (
        <div className="space-y-2 max-h-36 overflow-y-auto">
          <div className="text-xs text-gray-500 mb-1">Top cost by RMI</div>
          {topRmis.map((rmi) => (
            <div key={rmi.id} className="flex justify-between text-xs">
              <span className="text-gray-300 truncate flex-1 mr-2">{rmi.title}</span>
              <span className="text-gray-400">{formatCost(byRmi![rmi.id].costUsd)}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function SpecsQuadrant({
  judgeResults,
  workflows,
}: {
  judgeResults: JudgeResult[]
  workflows: { id: string; name: string; specs_required?: string[] }[]
}) {
  const avgScore = judgeResults.length > 0
    ? judgeResults.reduce((sum, r) => sum + r.score, 0) / judgeResults.length
    : 0

  const specTypes = [...new Set(judgeResults.map((r) => specType(r.spec_path)))]
  const workflow = workflows[0]
  const requiredSpecs = new Set(workflow?.specs_required ?? [])
  const missingRequired = [...requiredSpecs].filter((s) => !specTypes.includes(s))

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-medium text-purple-400">Specs</h3>
        {judgeResults.length > 0 && (
          <span className={`text-lg font-semibold ${
            avgScore >= 7 ? 'text-green-400' : avgScore >= 4 ? 'text-yellow-400' : 'text-red-400'
          }`}>
            {avgScore.toFixed(1)}
          </span>
        )}
      </div>

      {judgeResults.length === 0 ? (
        <div className="text-gray-400 text-sm">No specs judged yet</div>
      ) : (
        <>
          <div className="flex flex-wrap gap-1 mb-3">
            {specTypes.map((type) => (
              <span
                key={type}
                className={`text-xs px-1.5 py-0.5 rounded ${
                  requiredSpecs.has(type) ? 'bg-purple-500/30 text-purple-300' : 'bg-gray-700 text-gray-400'
                }`}
              >
                {type}
              </span>
            ))}
          </div>

          {missingRequired.length > 0 && (
            <div className="text-xs text-red-400 mb-2">
              Missing: {missingRequired.join(', ')}
            </div>
          )}

          <div className="space-y-1 max-h-32 overflow-y-auto">
            {judgeResults
              .sort((a, b) => b.score - a.score)
              .slice(0, 5)
              .map((r) => (
                <div key={r.id} className="flex justify-between text-xs">
                  <span className="text-gray-300">{r.spec_path.split('/').pop()}</span>
                  <span className={`px-1.5 py-0.5 rounded ${
                    r.score >= 7 ? 'bg-green-500/30' : r.score >= 4 ? 'bg-yellow-500/30' : 'bg-red-500/30'
                  }`}>
                    {r.score.toFixed(1)}
                  </span>
                </div>
              ))}
          </div>
        </>
      )}
    </div>
  )
}

function MaturityQuadrant({
  assessments,
  models,
}: {
  assessments: MaturityAssessment[]
  models: { id: string; name: string; max_level: number; dimensions?: { key: string; name: string }[] }[]
}) {
  if (assessments.length === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="font-medium text-orange-400 mb-4">Maturity</h3>
        <div className="text-gray-400 text-sm">No assessments yet</div>
      </div>
    )
  }

  const latest = assessments[0]
  const model = models.find((m) => m.id === latest.model_id)
  const avgLevel = latest.scores?.length
    ? latest.scores.reduce((sum, s) => sum + s.level, 0) / latest.scores.length
    : 0

  const radarAxes: RadarAxis[] = model?.dimensions?.map((d) => ({
    key: d.key,
    label: d.name,
    max: model.max_level,
  })) ?? []

  const radarDataset: RadarDataset = {
    name: latest.initiative_id,
    values: Object.fromEntries((latest.scores ?? []).map((s) => [s.dimension_key, s.level])),
  }

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-medium text-orange-400">Maturity</h3>
        <span className="text-lg font-semibold text-orange-400">
          {avgLevel.toFixed(1)}/{model?.max_level ?? 5}
        </span>
      </div>

      {radarAxes.length >= 3 && (
        <div className="flex justify-center">
          <RadarChart axes={radarAxes} datasets={[radarDataset]} size={140} />
        </div>
      )}

      {model && (
        <div className="text-xs text-gray-500 text-center mt-2">
          {model.name}
        </div>
      )}
    </div>
  )
}

function specType(path: string): string {
  const name = path.split('/').pop()?.toLowerCase() ?? ''
  if (name.includes('prd')) return 'PRD'
  if (name.includes('trd')) return 'TRD'
  if (name.includes('plan')) return 'PLAN'
  if (name.includes('roadmap')) return 'ROADMAP'
  return name.replace('.md', '').toUpperCase()
}
