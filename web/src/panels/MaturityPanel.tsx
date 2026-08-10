import { useState, useEffect, useMemo } from 'react'
import { getMaturity, getScale, getLeverage, getExecution } from '../api/client'
import type { MaturityResponse, MaturityAssessment, CapabilityModel, ScaleResponse, ScaleMetric, LeverageGraph, ExecutionResponse } from '../api/types'
import { RadarChart, type RadarAxis, type RadarDataset } from '../components/charts'
import { LoadingState, ErrorState, EmptyState } from '../components'
import { hiddenInitiativeIds } from '../lib/visibility'

type ViewMode = 'scale' | 'leverage' | 'models'

export function MaturityPanel() {
  const [viewMode, setViewMode] = useState<ViewMode>('scale')
  const [maturityData, setMaturityData] = useState<MaturityResponse | null>(null)
  const [scaleData, setScaleData] = useState<ScaleResponse | null>(null)
  const [leverageData, setLeverageData] = useState<LeverageGraph | null>(null)
  // Fetched alongside the maturity data above so the Capability Models
  // initiative filter can exclude hidden initiatives; kept optional
  // (unlike maturity/scale) so a slow or failed execution fetch doesn't
  // block this page the way it isn't blocked by leverage data either.
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedModel, setSelectedModel] = useState<string | null>(null)

  const reload = () => {
    setError(null)
    Promise.all([getMaturity(), getScale(), getLeverage()])
      .then(([m, s, l]) => {
        setMaturityData(m)
        setScaleData(s)
        setLeverageData(l)
        if (m.models.length > 0 && !selectedModel) {
          setSelectedModel(m.models[0].id)
        }
      })
      .catch((err: Error) => setError(err.message))
    getExecution()
      .then(setExecution)
      .catch(() => {
        // Non-fatal: the initiative filter just won't exclude hidden
        // initiatives if this fails.
      })
  }

  useEffect(() => {
    reload()
  }, [])

  if (error) {
    return <ErrorState message={error} onRetry={reload} />
  }

  if (!maturityData || !scaleData) {
    return <LoadingState message="Loading maturity data..." />
  }

  return (
    <div className="space-y-6">
      {/* View Mode Selector */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Maturity</h2>
        <div className="flex gap-2">
          <button
            onClick={() => setViewMode('scale')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              viewMode === 'scale'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            SCALE Platform
          </button>
          <button
            onClick={() => setViewMode('leverage')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              viewMode === 'leverage'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Leverage Graph
          </button>
          <button
            onClick={() => setViewMode('models')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              viewMode === 'models'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Capability Models
          </button>
        </div>
      </div>

      {viewMode === 'scale' && <ScaleView data={scaleData} />}

      {viewMode === 'leverage' && leverageData && <LeverageView data={leverageData} />}

      {viewMode === 'models' && (
        <ModelsView
          data={maturityData}
          execution={execution}
          selectedModel={selectedModel}
          onSelectModel={setSelectedModel}
        />
      )}
    </div>
  )
}

function ScaleView({ data }: { data: ScaleResponse }) {
  const [expandedCap, setExpandedCap] = useState<string | null>(null)

  if (!data.hasData || !data.framework) {
    return (
      <EmptyState
        title="SCALE data unavailable"
        description={data.dataNote ?? 'Platform adoption metrics not available.'}
        hint="Check SCALE catalog configuration"
      />
    )
  }

  const domain = data.framework.domains[0]
  if (!domain) {
    return <EmptyState title="No domains" description="No SCALE domains found" />
  }

  return (
    <div className="space-y-6">
      {/* SCALE Aspect Tiles */}
      {data.rollup && (
        <div className="grid grid-cols-5 gap-3">
          {data.rollup.aspects.map((aspect) => (
            <AspectTile key={aspect.aspect} aspect={aspect} />
          ))}
        </div>
      )}

      {/* Assessment Info */}
      {data.assessment && (
        <div className="bg-gray-800 rounded-lg p-4 flex items-center justify-between">
          <div>
            <span className="text-sm text-gray-400">Period: </span>
            <span className="font-medium">{data.assessment.period}</span>
          </div>
          <div>
            <span className="text-sm text-gray-400">As of: </span>
            <span className="font-medium">{data.assessment.asOf}</span>
          </div>
          <div>
            <span className="text-sm text-gray-400">Observations: </span>
            <span className="font-medium">{data.assessment.observations}</span>
          </div>
        </div>
      )}

      {/* Domain Header */}
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="text-xl font-semibold">{domain.name}</h3>
        {domain.description && (
          <p className="text-gray-400 mt-1 text-sm">{domain.description}</p>
        )}
      </div>

      {/* Capabilities */}
      <div className="space-y-3">
        {domain.capabilities.map((cap) => (
          <CapabilityCard
            key={cap.id}
            capability={cap}
            expanded={expandedCap === cap.id}
            onToggle={() => setExpandedCap(expandedCap === cap.id ? null : cap.id)}
          />
        ))}
      </div>
    </div>
  )
}

function AspectTile({ aspect }: { aspect: { aspect: string; letter: string; displayName: string; score: number; eligible: number; observed: number } }) {
  const getColor = (score: number) => {
    if (score >= 80) return 'bg-green-600'
    if (score >= 60) return 'bg-blue-600'
    if (score >= 40) return 'bg-yellow-600'
    if (score > 0) return 'bg-orange-600'
    return 'bg-gray-600'
  }

  return (
    <div className={`${getColor(aspect.score)} rounded-lg p-4 text-center`}>
      <div className="text-3xl font-bold">{aspect.letter}</div>
      <div className="text-sm font-medium mt-1">{aspect.displayName}</div>
      <div className="text-2xl font-semibold mt-2">{aspect.score.toFixed(0)}%</div>
      <div className="text-xs opacity-75 mt-1">
        {aspect.observed}/{aspect.eligible} metrics
      </div>
    </div>
  )
}

function CapabilityCard({
  capability,
  expanded,
  onToggle,
}: {
  capability: { id: string; name: string; description?: string; metrics: ScaleMetric[] }
  expanded: boolean
  onToggle: () => void
}) {
  const metricsByAspect = useMemo(() => {
    const groups: Record<string, ScaleMetric[]> = {}
    for (const m of capability.metrics) {
      if (!groups[m.aspect]) groups[m.aspect] = []
      groups[m.aspect].push(m)
    }
    return groups
  }, [capability.metrics])

  const observedCount = capability.metrics.filter((m) => m.value !== undefined).length
  const avgAttainment = useMemo(() => {
    const withAttainment = capability.metrics.filter((m) => m.attainment !== undefined)
    if (withAttainment.length === 0) return null
    return withAttainment.reduce((sum, m) => sum + (m.attainment ?? 0), 0) / withAttainment.length
  }, [capability.metrics])

  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden">
      <button
        onClick={onToggle}
        className="w-full p-4 flex items-center justify-between text-left hover:bg-gray-750"
      >
        <div>
          <h4 className="font-medium">{capability.name}</h4>
          {capability.description && (
            <p className="text-sm text-gray-400 mt-1">{capability.description}</p>
          )}
        </div>
        <div className="flex items-center gap-4">
          <div className="text-right">
            <div className="text-sm text-gray-400">{observedCount}/{capability.metrics.length} observed</div>
            {avgAttainment !== null && (
              <div className="text-lg font-semibold">{(avgAttainment * 100).toFixed(0)}%</div>
            )}
          </div>
          <span className="text-gray-500">{expanded ? '▼' : '▶'}</span>
        </div>
      </button>

      {expanded && (
        <div className="border-t border-gray-700 p-4">
          {Object.entries(metricsByAspect).map(([aspect, metrics]) => (
            <div key={aspect} className="mb-4 last:mb-0">
              <div className="text-xs font-medium text-gray-500 uppercase mb-2">
                {aspectDisplayName(aspect)}
              </div>
              <div className="space-y-2">
                {metrics.map((m) => (
                  <MetricRow key={m.id} metric={m} />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function MetricRow({ metric }: { metric: ScaleMetric }) {
  const hasValue = metric.value !== undefined

  return (
    <div className="flex items-center justify-between py-2 px-3 bg-gray-900 rounded">
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium truncate">{metric.name}</div>
        {metric.note && (
          <div className="text-xs text-gray-500 truncate">{metric.note}</div>
        )}
      </div>
      <div className="flex items-center gap-4 ml-4">
        {hasValue && (
          <div className="text-right">
            <div className="font-mono text-sm">
              {metric.numerator !== undefined && metric.denominator !== undefined
                ? `${metric.numerator}/${metric.denominator}`
                : formatValue(metric.value!, metric.unit)}
            </div>
            {metric.attainment !== undefined && (
              <AttainmentBar attainment={metric.attainment} />
            )}
          </div>
        )}
        {!hasValue && (
          <span className="text-xs text-gray-600">No data</span>
        )}
        {metric.targetValue !== undefined && (
          <div className="text-xs text-gray-500">
            Target: {formatValue(metric.targetValue, metric.unit)}
          </div>
        )}
      </div>
    </div>
  )
}

function AttainmentBar({ attainment }: { attainment: number }) {
  const pct = Math.min(attainment * 100, 100)
  const color = pct >= 80 ? 'bg-green-500' : pct >= 50 ? 'bg-yellow-500' : 'bg-red-500'

  return (
    <div className="w-16 h-1.5 bg-gray-700 rounded-full overflow-hidden">
      <div className={`h-full ${color}`} style={{ width: `${pct}%` }} />
    </div>
  )
}

function formatValue(value: number, unit?: string): string {
  if (unit === 'percent') return `${value.toFixed(1)}%`
  if (unit === 'count') return value.toFixed(0)
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toFixed(1)
}

function aspectDisplayName(aspect: string): string {
  const names: Record<string, string> = {
    standards: 'Standards',
    consumption: 'Consumption',
    automation: 'Automation',
    leverage: 'Leverage',
    effectiveness: 'Effectiveness',
  }
  return names[aspect] ?? aspect
}

function ModelsView({
  data,
  execution,
  selectedModel,
  onSelectModel,
}: {
  data: MaturityResponse
  execution: ExecutionResponse | null
  selectedModel: string | null
  onSelectModel: (id: string) => void
}) {
  if (data.models.length === 0) {
    return (
      <EmptyState
        title="No maturity models defined"
        description="Create a capability model to track maturity assessments."
        hint="prismctl maturity model create"
      />
    )
  }

  const model = data.models.find((m) => m.id === selectedModel)
  const modelAssessments = (data.assessments ?? []).filter((a) => a.model_id === selectedModel)

  return (
    <div className="space-y-6">
      {/* Model Selector */}
      <div className="flex gap-2 flex-wrap">
        {data.models.map((m) => (
          <button
            key={m.id}
            onClick={() => onSelectModel(m.id)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              selectedModel === m.id
                ? 'bg-blue-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            {m.name}
          </button>
        ))}
      </div>

      {model && <ModelView model={model} assessments={modelAssessments} execution={execution} />}
    </div>
  )
}

function ModelView({
  model,
  assessments,
  execution,
}: {
  model: CapabilityModel
  assessments: MaturityAssessment[]
  execution: ExecutionResponse | null
}) {
  const [selectedInitiatives, setSelectedInitiatives] = useState<Set<string>>(new Set())

  // Assessments for a hidden initiative are excluded from every view below
  // (chips, radar, dimension averages, and the table), not just the filter
  // chips -- otherwise a hidden initiative's data would still leak into
  // aggregates and the table even though it can't be selected.
  const visibleAssessments = useMemo(() => {
    if (!execution) return assessments
    const hidden = hiddenInitiativeIds(execution.initiatives, execution.programs)
    return assessments.filter((a) => !hidden.has(a.initiative_id))
  }, [assessments, execution])

  const initiativeIds = useMemo(
    () => [...new Set(visibleAssessments.map((a) => a.initiative_id))],
    [visibleAssessments]
  )

  useEffect(() => {
    if (initiativeIds.length > 0 && selectedInitiatives.size === 0) {
      setSelectedInitiatives(new Set(initiativeIds.slice(0, 3)))
    }
  }, [initiativeIds])

  const radarAxes = useMemo((): RadarAxis[] => {
    if (!model.dimensions || model.dimensions.length < 3) return []
    return model.dimensions.map((dim) => ({
      key: dim.key,
      label: dim.name,
      max: model.max_level,
    }))
  }, [model])

  const radarDatasets = useMemo((): RadarDataset[] => {
    return visibleAssessments
      .filter((a) => selectedInitiatives.has(a.initiative_id))
      .map((a) => ({
        name: a.initiative_id,
        values: Object.fromEntries(
          (a.scores ?? []).map((s) => [s.dimension_key, s.level])
        ),
      }))
  }, [visibleAssessments, selectedInitiatives])

  const avgByDimension = useMemo(() => {
    const sums: Record<string, { total: number; count: number }> = {}
    for (const a of visibleAssessments) {
      for (const s of a.scores ?? []) {
        if (!sums[s.dimension_key]) sums[s.dimension_key] = { total: 0, count: 0 }
        sums[s.dimension_key].total += s.level
        sums[s.dimension_key].count++
      }
    }
    return Object.fromEntries(
      Object.entries(sums).map(([k, v]) => [k, v.count > 0 ? v.total / v.count : 0])
    )
  }, [visibleAssessments])

  const toggleInitiative = (id: string) => {
    setSelectedInitiatives((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  return (
    <div className="space-y-6">
      {/* Model Header */}
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="text-xl font-semibold">{model.name}</h3>
        {model.description && (
          <p className="text-gray-400 mt-1">{model.description}</p>
        )}
        <div className="flex items-center gap-4 mt-3 text-sm text-gray-500">
          <span>{model.dimensions?.length ?? 0} dimensions</span>
          <span>Max level: {model.max_level}</span>
          <span>{visibleAssessments.length} assessments</span>
        </div>
      </div>

      {/* Initiative Filter */}
      {initiativeIds.length > 1 && (
        <div className="flex flex-wrap gap-2">
          {initiativeIds.map((id) => (
            <button
              key={id}
              onClick={() => toggleInitiative(id)}
              className={`px-3 py-1.5 rounded text-xs font-mono transition-colors ${
                selectedInitiatives.has(id)
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {id}
            </button>
          ))}
        </div>
      )}

      {/* Radar Chart */}
      {radarAxes.length >= 3 && radarDatasets.length > 0 && (
        <div className="bg-gray-800 rounded-lg p-6">
          <h4 className="text-sm font-medium text-gray-400 mb-4">Maturity Radar</h4>
          <div className="flex justify-center">
            <RadarChart axes={radarAxes} datasets={radarDatasets} />
          </div>
        </div>
      )}

      {/* Dimensions Grid */}
      {model.dimensions && model.dimensions.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-sm font-medium text-gray-400">Dimensions</h4>
          <div className="grid grid-cols-2 gap-4">
            {model.dimensions.map((dim) => (
              <DimensionCard
                key={dim.key}
                dimension={dim}
                maxLevel={model.max_level}
                avgLevel={avgByDimension[dim.key] ?? 0}
                assessments={visibleAssessments.filter((a) =>
                  selectedInitiatives.has(a.initiative_id)
                )}
              />
            ))}
          </div>
        </div>
      )}

      {/* Assessments Table */}
      {visibleAssessments.length > 0 && (
        <div className="space-y-4">
          <h4 className="text-sm font-medium text-gray-400">Assessments</h4>
          <div className="bg-gray-800 rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-900">
                <tr>
                  <th className="text-left p-3 font-medium">Initiative</th>
                  <th className="text-left p-3 font-medium">Assessed</th>
                  <th className="text-left p-3 font-medium">Dimensions</th>
                  <th className="text-right p-3 font-medium">Avg Level</th>
                </tr>
              </thead>
              <tbody>
                {visibleAssessments
                  .sort((a, b) => new Date(b.assessed_at).getTime() - new Date(a.assessed_at).getTime())
                  .map((a) => {
                    const avg = a.scores?.length
                      ? a.scores.reduce((sum, s) => sum + s.level, 0) / a.scores.length
                      : 0
                    return (
                      <tr key={a.id} className="border-t border-gray-700">
                        <td className="p-3 font-mono text-xs">{a.initiative_id}</td>
                        <td className="p-3 text-gray-400">
                          {new Date(a.assessed_at).toLocaleDateString()}
                        </td>
                        <td className="p-3">{a.scores?.length ?? 0} scored</td>
                        <td className="p-3 text-right">
                          <LevelBadge level={avg} max={model.max_level} />
                        </td>
                      </tr>
                    )
                  })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {visibleAssessments.length === 0 && (
        <EmptyState
          title="No assessments"
          description={`No initiatives have been assessed against ${model.name} yet.`}
          hint="prismctl maturity assess"
        />
      )}
    </div>
  )
}

function DimensionCard({
  dimension,
  maxLevel,
  avgLevel,
  assessments,
}: {
  dimension: { key: string; name: string; sources?: string[]; levels?: { level: number; name: string; description?: string }[] }
  maxLevel: number
  avgLevel: number
  assessments: MaturityAssessment[]
}) {
  const latestScore = assessments
    .flatMap((a) => a.scores ?? [])
    .find((s) => s.dimension_key === dimension.key)

  const levelDef = dimension.levels?.find((l) => l.level === latestScore?.level)

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-2">
        <h5 className="font-medium">{dimension.name}</h5>
        {avgLevel > 0 && (
          <span className="text-xs text-gray-500">Avg: {avgLevel.toFixed(1)}</span>
        )}
      </div>
      {dimension.sources && dimension.sources.length > 0 && (
        <div className="text-xs text-gray-500 mb-2">
          Sources: {dimension.sources.join(', ')}
        </div>
      )}
      <div className="flex gap-1 mb-2">
        {Array.from({ length: maxLevel }, (_, i) => {
          const level = i + 1
          const isScored = latestScore && latestScore.level >= level
          return (
            <div
              key={level}
              className={`h-2 flex-1 rounded ${
                isScored ? 'bg-blue-500' : 'bg-gray-700'
              }`}
              title={dimension.levels?.find((l) => l.level === level)?.name ?? `Level ${level}`}
            />
          )
        })}
      </div>
      {latestScore && levelDef && (
        <div className="text-sm text-gray-400">
          Level {latestScore.level}: {levelDef.name}
        </div>
      )}
      {latestScore?.notes && (
        <div className="text-xs text-gray-500 mt-1 italic">{latestScore.notes}</div>
      )}
    </div>
  )
}

function LevelBadge({ level, max }: { level: number; max: number }) {
  const pct = max > 0 ? level / max : 0
  const color =
    pct >= 0.8 ? 'bg-green-500' : pct >= 0.5 ? 'bg-blue-500' : pct >= 0.3 ? 'bg-yellow-500' : 'bg-gray-600'

  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium text-white ${color}`}>
      {level.toFixed(1)}
    </span>
  )
}

function trimModulePath(path: string): string {
  return path.replace(/^github\.com\//, '')
}

function LeverageView({ data }: { data: LeverageGraph }) {
  const [filter, setFilter] = useState<'all' | 'internal' | 'orphans'>('internal')
  const [expandedOrg, setExpandedOrg] = useState<string | null>(null)

  const { summary } = data

  const internalModules = useMemo(() => {
    return data.modules.filter(m => m.kind === 'internal')
  }, [data.modules])

  const modulesByOrg = useMemo(() => {
    const orgs: Record<string, typeof internalModules> = {}
    for (const m of internalModules) {
      const org = m.org || 'unknown'
      if (!orgs[org]) orgs[org] = []
      orgs[org].push(m)
    }
    for (const org of Object.keys(orgs)) {
      orgs[org].sort((a, b) => b.stats.directDependents - a.stats.directDependents)
    }
    return orgs
  }, [internalModules])

  const orphanModules = useMemo(() => {
    return (summary.orphans ?? []).map(id => data.modules.find(m => m.id === id)).filter(Boolean)
  }, [summary.orphans, data.modules])

  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="text-sm text-gray-400">Total Modules</div>
          <div className="text-2xl font-bold">{summary.totalModules}</div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="text-sm text-gray-400">Internal</div>
          <div className="text-2xl font-bold text-blue-400">{summary.internalModules}</div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="text-sm text-gray-400">Internal Ratio</div>
          <div className="text-2xl font-bold text-green-400">{summary.internalRatio.toFixed(1)}%</div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <div className="text-sm text-gray-400">Orphans</div>
          <div className="text-2xl font-bold text-orange-400">{summary.orphans?.length ?? 0}</div>
          <div className="text-xs text-gray-500">reuse opportunities</div>
        </div>
      </div>

      {/* Top Leveraged and Top Consumers side by side */}
      <div className="grid grid-cols-2 gap-4">
        {/* Top Leveraged */}
        {summary.topLeveraged && summary.topLeveraged.length > 0 && (
          <div className="bg-gray-800 rounded-lg p-4">
            <h3 className="font-semibold mb-3">Top Leveraged <span className="text-gray-400 font-normal text-sm">(most depended upon)</span></h3>
            <div className="space-y-2">
              {summary.topLeveraged.map((m, i) => (
                <div key={m.moduleId} className="flex items-center gap-3">
                  <span className="text-gray-500 w-6 text-right">{i + 1}.</span>
                  <div className="flex-1 min-w-0">
                    <div className="font-mono text-sm truncate" title={m.moduleId}>{trimModulePath(m.moduleId)}</div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-green-400 font-bold">{m.dependents}</span>
                    <div className="w-16 bg-gray-700 rounded-full h-2">
                      <div
                        className="bg-green-500 h-2 rounded-full"
                        style={{ width: `${Math.min((m.dependents / (summary.topLeveraged![0].dependents || 1)) * 100, 100)}%` }}
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Top Consumers */}
        {summary.topConsumers && summary.topConsumers.length > 0 && (
          <div className="bg-gray-800 rounded-lg p-4">
            <h3 className="font-semibold mb-3">Top Consumers <span className="text-gray-400 font-normal text-sm">(most internal deps)</span></h3>
            <div className="space-y-2">
              {summary.topConsumers.map((m, i) => (
                <div key={m.moduleId} className="flex items-center gap-3">
                  <span className="text-gray-500 w-6 text-right">{i + 1}.</span>
                  <div className="flex-1 min-w-0">
                    <div className="font-mono text-sm truncate" title={m.moduleId}>{trimModulePath(m.moduleId)}</div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right min-w-[90px]">
                      <span className="text-blue-400 font-bold">{m.dependents}</span>
                      <span className="text-xs text-gray-500 ml-1">
                        ({m.direct ?? 0}d/{m.indirect ?? 0}i)
                      </span>
                    </div>
                    <div className="w-16 bg-gray-700 rounded-full h-2">
                      <div
                        className="bg-blue-500 h-2 rounded-full"
                        style={{ width: `${Math.min((m.dependents / (summary.topConsumers![0].dependents || 1)) * 100, 100)}%` }}
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2">
        <button
          onClick={() => setFilter('internal')}
          className={`px-3 py-1.5 rounded text-sm ${filter === 'internal' ? 'bg-blue-600' : 'bg-gray-700 hover:bg-gray-600'}`}
        >
          By Organization ({Object.keys(modulesByOrg).length})
        </button>
        <button
          onClick={() => setFilter('orphans')}
          className={`px-3 py-1.5 rounded text-sm ${filter === 'orphans' ? 'bg-orange-600' : 'bg-gray-700 hover:bg-gray-600'}`}
        >
          Orphans ({orphanModules.length})
        </button>
      </div>

      {/* Modules by Org */}
      {filter === 'internal' && (
        <div className="space-y-3">
          {Object.entries(modulesByOrg).map(([org, modules]) => (
            <div key={org} className="bg-gray-800 rounded-lg overflow-hidden">
              <button
                onClick={() => setExpandedOrg(expandedOrg === org ? null : org)}
                className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-700"
              >
                <div className="flex items-center gap-3">
                  <span className="font-semibold">{org}</span>
                  <span className="text-sm text-gray-400">{modules.length} modules</span>
                </div>
                <span className={`transition-transform ${expandedOrg === org ? 'rotate-180' : ''}`}>▼</span>
              </button>
              {expandedOrg === org && (
                <div className="border-t border-gray-700 max-h-80 overflow-y-auto">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-900 sticky top-0">
                      <tr>
                        <th className="px-4 py-2 text-left">Module</th>
                        <th className="px-4 py-2 text-right">Dependents</th>
                        <th className="px-4 py-2 text-right">Dependencies</th>
                        <th className="px-4 py-2 text-right">Leverage</th>
                      </tr>
                    </thead>
                    <tbody>
                      {modules.map(m => (
                        <tr key={m.id} className="border-t border-gray-700 hover:bg-gray-750">
                          <td className="px-4 py-2 font-mono truncate max-w-xs">{m.name}</td>
                          <td className="px-4 py-2 text-right">
                            {m.stats.directDependents > 0 ? (
                              <span className="text-green-400">{m.stats.directDependents}</span>
                            ) : (
                              <span className="text-gray-500">0</span>
                            )}
                          </td>
                          <td className="px-4 py-2 text-right text-gray-400">{m.stats.directDependencies}</td>
                          <td className="px-4 py-2 text-right">
                            {(m.stats.leverageScore ?? 0) > 1 ? (
                              <span className="text-green-400">{(m.stats.leverageScore ?? 0).toFixed(1)}x</span>
                            ) : (
                              <span className="text-gray-500">{(m.stats.leverageScore ?? 0).toFixed(1)}x</span>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Orphans */}
      {filter === 'orphans' && (
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-sm text-gray-400 mb-4">
            Internal modules with no dependents. These represent potential reuse opportunities.
          </p>
          <div className="grid grid-cols-2 gap-2 max-h-96 overflow-y-auto">
            {orphanModules.map(m => m && (
              <div key={m.id} className="px-3 py-2 bg-gray-700 rounded font-mono text-sm truncate">
                {m.path}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Meta */}
      <div className="text-xs text-gray-500 text-right">
        Generated {new Date(data.generatedAt).toLocaleString()} • {data.ecosystem} • {data.scope}
      </div>
    </div>
  )
}
