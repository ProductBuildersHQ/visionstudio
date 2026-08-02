import { useState, useEffect } from 'react'
import { getMaturity } from '../api/client'
import type { MaturityResponse, MaturityAssessment } from '../api/types'
import { RadarChart, type RadarAxis, type RadarDataset } from '../components/charts'

export function MaturityPanel() {
  const [data, setData] = useState<MaturityResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedModel, setSelectedModel] = useState<string | null>(null)

  useEffect(() => {
    getMaturity()
      .then((d) => {
        setData(d)
        if (d.models.length > 0) {
          setSelectedModel(d.models[0].id)
        }
      })
      .catch((err: Error) => setError(err.message))
  }, [])

  if (error) {
    return <div className="text-red-400">Error: {error}</div>
  }

  if (!data) {
    return <div className="text-gray-400">Loading...</div>
  }

  if (data.models.length === 0) {
    return <div className="text-gray-400">No capability models defined.</div>
  }

  const model = data.models.find((m) => m.id === selectedModel)
  const modelAssessments = data.assessments.filter((a) => a.model_id === selectedModel)

  return (
    <div className="space-y-6">
      {/* Model Selector */}
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

      {model && (
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold">{model.name}</h2>
            {model.description && (
              <p className="text-gray-400 mt-1">{model.description}</p>
            )}
          </div>

          {/* Radar Chart */}
          {model.dimensions && model.dimensions.length >= 3 && modelAssessments.length > 0 && (
            <div className="bg-gray-800 rounded-lg p-4">
              <h3 className="text-lg font-semibold mb-4">Maturity Radar</h3>
              <RadarChart
                axes={model.dimensions.map((dim): RadarAxis => ({
                  key: dim.key,
                  label: dim.name,
                  max: model.max_level,
                }))}
                datasets={modelAssessments.slice(0, 3).map((a): RadarDataset => ({
                  name: a.initiative_id,
                  values: Object.fromEntries(
                    (a.scores ?? []).map((s) => [s.dimension_key, s.level])
                  ),
                }))}
              />
            </div>
          )}

          {/* Dimensions */}
          {model.dimensions && model.dimensions.length > 0 && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold">Dimensions</h3>
              <div className="grid grid-cols-2 gap-4">
                {model.dimensions.map((dim) => (
                  <DimensionCard
                    key={dim.key}
                    dimension={dim}
                    maxLevel={model.max_level}
                    assessments={modelAssessments}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Assessments */}
          {modelAssessments.length > 0 && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold">Assessments</h3>
              <div className="bg-gray-800 rounded-lg overflow-hidden">
                <table className="w-full text-sm">
                  <thead className="bg-gray-900">
                    <tr>
                      <th className="text-left p-3 font-medium">Initiative</th>
                      <th className="text-left p-3 font-medium">Assessed</th>
                      <th className="text-left p-3 font-medium">Dimensions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {modelAssessments.map((a) => (
                      <tr key={a.id} className="border-t border-gray-700">
                        <td className="p-3 font-mono text-xs">{a.initiative_id}</td>
                        <td className="p-3 text-gray-400">
                          {new Date(a.assessed_at).toLocaleDateString()}
                        </td>
                        <td className="p-3">
                          {a.scores?.length ?? 0} scored
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function DimensionCard({
  dimension,
  maxLevel,
  assessments,
}: {
  dimension: { key: string; name: string; sources?: string[]; levels?: { level: number; name: string; description?: string }[] }
  maxLevel: number
  assessments: MaturityAssessment[]
}) {
  const latestScore = assessments
    .flatMap((a) => a.scores ?? [])
    .find((s) => s.dimension_key === dimension.key)

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-2">
        <h4 className="font-medium">{dimension.name}</h4>
        {dimension.sources && dimension.sources.length > 0 && (
          <span className="text-xs text-gray-500">{dimension.sources.join(', ')}</span>
        )}
      </div>
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
            />
          )
        })}
      </div>
      {latestScore && (
        <div className="text-sm text-gray-400">
          Level {latestScore.level}: {dimension.levels?.find((l) => l.level === latestScore.level)?.name ?? ''}
        </div>
      )}
    </div>
  )
}
