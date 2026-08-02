import { useState, useEffect, useMemo } from 'react'
import { getMaturity } from '../api/client'
import type { MaturityResponse, MaturityAssessment, CapabilityModel } from '../api/types'
import { RadarChart, type RadarAxis, type RadarDataset } from '../components/charts'
import { LoadingState, ErrorState, EmptyState } from '../components'

export function MaturityPanel() {
  const [data, setData] = useState<MaturityResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedModel, setSelectedModel] = useState<string | null>(null)

  const reload = () => {
    setError(null)
    getMaturity()
      .then((d) => {
        setData(d)
        if (d.models.length > 0 && !selectedModel) {
          setSelectedModel(d.models[0].id)
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
    return <LoadingState message="Loading maturity data..." />
  }

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
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Maturity Assessments</h2>
        <div className="flex gap-2">
          {data.models.map((m) => (
            <button
              key={m.id}
              onClick={() => setSelectedModel(m.id)}
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
      </div>

      {model && (
        <ModelView model={model} assessments={modelAssessments} />
      )}
    </div>
  )
}

function ModelView({
  model,
  assessments,
}: {
  model: CapabilityModel
  assessments: MaturityAssessment[]
}) {
  const [selectedInitiatives, setSelectedInitiatives] = useState<Set<string>>(new Set())

  const initiativeIds = useMemo(
    () => [...new Set(assessments.map((a) => a.initiative_id))],
    [assessments]
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
    return assessments
      .filter((a) => selectedInitiatives.has(a.initiative_id))
      .map((a) => ({
        name: a.initiative_id,
        values: Object.fromEntries(
          (a.scores ?? []).map((s) => [s.dimension_key, s.level])
        ),
      }))
  }, [assessments, selectedInitiatives])

  const avgByDimension = useMemo(() => {
    const sums: Record<string, { total: number; count: number }> = {}
    for (const a of assessments) {
      for (const s of a.scores ?? []) {
        if (!sums[s.dimension_key]) sums[s.dimension_key] = { total: 0, count: 0 }
        sums[s.dimension_key].total += s.level
        sums[s.dimension_key].count++
      }
    }
    return Object.fromEntries(
      Object.entries(sums).map(([k, v]) => [k, v.count > 0 ? v.total / v.count : 0])
    )
  }, [assessments])

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
          <span>{assessments.length} assessments</span>
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
                assessments={assessments.filter((a) =>
                  selectedInitiatives.has(a.initiative_id)
                )}
              />
            ))}
          </div>
        </div>
      )}

      {/* Assessments Table */}
      {assessments.length > 0 && (
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
                {assessments
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

      {assessments.length === 0 && (
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
