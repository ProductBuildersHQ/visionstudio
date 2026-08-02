import { useState, useEffect } from 'react'
import { getSpecs, getExecution } from '../api/client'
import type { SpecsResponse, ExecutionResponse, JudgeResult } from '../api/types'

export function SpecsPanel() {
  const [specs, setSpecs] = useState<SpecsResponse | null>(null)
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([getSpecs(), getExecution()])
      .then(([s, e]) => {
        setSpecs(s)
        setExecution(e)
      })
      .catch((err: Error) => setError(err.message))
  }, [])

  if (error) {
    return <div className="text-red-400">Error: {error}</div>
  }

  if (!specs || !execution) {
    return <div className="text-gray-400">Loading...</div>
  }

  // Group judge results by initiative
  const resultsByInit = specs.judgeResults.reduce((acc, r) => {
    if (!acc[r.initiative_id]) {
      acc[r.initiative_id] = []
    }
    acc[r.initiative_id].push(r)
    return acc
  }, {} as Record<string, JudgeResult[]>)

  return (
    <div className="space-y-6">
      {/* Workflows */}
      <div>
        <h3 className="text-lg font-semibold mb-3">Spec Workflows</h3>
        <div className="grid grid-cols-2 gap-4">
          {specs.workflows.map((wf) => (
            <div key={wf.id} className="bg-gray-800 rounded-lg p-4">
              <h4 className="font-medium">{wf.name}</h4>
              {wf.description && (
                <p className="text-sm text-gray-400 mt-1">{wf.description}</p>
              )}
              <div className="mt-3 space-y-1">
                {wf.specs_required && wf.specs_required.length > 0 && (
                  <div className="text-sm">
                    <span className="text-gray-500">Required:</span>{' '}
                    <span className="text-gray-300">{wf.specs_required.join(', ')}</span>
                  </div>
                )}
                {wf.specs_optional && wf.specs_optional.length > 0 && (
                  <div className="text-sm">
                    <span className="text-gray-500">Optional:</span>{' '}
                    <span className="text-gray-400">{wf.specs_optional.join(', ')}</span>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Judge Results by Initiative */}
      {Object.keys(resultsByInit).length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-3">Judge Results</h3>
          <div className="space-y-4">
            {Object.entries(resultsByInit).map(([initId, results]) => {
              const init = execution.initiatives.find((i) => i.id === initId)
              return (
                <div key={initId} className="bg-gray-800 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <div>
                      <span className="font-mono text-sm text-gray-400">{initId}</span>
                      {init && (
                        <span className="text-gray-300 ml-2">{init.title}</span>
                      )}
                    </div>
                    <AverageScore results={results} />
                  </div>
                  <div className="space-y-2">
                    {results
                      .sort((a, b) => new Date(b.evaluated_at).getTime() - new Date(a.evaluated_at).getTime())
                      .map((r) => (
                        <JudgeResultRow key={r.id} result={r} />
                      ))}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {specs.judgeResults.length === 0 && (
        <div className="text-gray-400">
          No judge results yet. Run <code className="bg-gray-800 px-1 rounded">prismctl spec judge</code> to evaluate specs.
        </div>
      )}
    </div>
  )
}

function AverageScore({ results }: { results: JudgeResult[] }) {
  const avg = results.reduce((sum, r) => sum + r.score, 0) / results.length
  const color = avg >= 7 ? 'text-green-400' : avg >= 4 ? 'text-yellow-400' : 'text-red-400'
  return (
    <span className={`font-semibold ${color}`}>
      Avg: {avg.toFixed(1)}
    </span>
  )
}

function JudgeResultRow({ result }: { result: JudgeResult }) {
  const scoreColor =
    result.score >= 7 ? 'bg-green-500' : result.score >= 4 ? 'bg-yellow-500' : 'bg-red-500'
  const specName = result.spec_path.split('/').pop() ?? result.spec_path

  return (
    <div className="flex items-center justify-between py-2 px-3 bg-gray-900 rounded">
      <div className="flex items-center gap-3">
        <span className="text-sm font-medium">{specName}</span>
        {result.model && (
          <span className="text-xs text-gray-500">{result.model}</span>
        )}
      </div>
      <div className="flex items-center gap-3">
        <span className="text-xs text-gray-400">
          {new Date(result.evaluated_at).toLocaleDateString()}
        </span>
        <span className={`px-2 py-0.5 rounded text-xs font-medium text-white ${scoreColor}`}>
          {result.score.toFixed(1)}
        </span>
      </div>
    </div>
  )
}
