import { useState, useEffect } from 'react'
import { getSpend, getExecution } from '../api/client'
import type { SpendResponse, ExecutionResponse } from '../api/types'

export function SpendPanel() {
  const [spend, setSpend] = useState<SpendResponse | null>(null)
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([getSpend(), getExecution()])
      .then(([s, e]) => {
        setSpend(s)
        setExecution(e)
      })
      .catch((err: Error) => setError(err.message))
  }, [])

  if (error) {
    return <div className="text-red-400">Error: {error}</div>
  }

  if (!spend || !execution) {
    return <div className="text-gray-400">Loading...</div>
  }

  if (!spend.hasData) {
    return (
      <div className="text-gray-400">
        <p>No token data available.</p>
        {spend.dataNote && <p className="text-sm mt-2">{spend.dataNote}</p>}
      </div>
    )
  }

  const formatTokens = (n: number): string => {
    if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
    if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
    return n.toString()
  }

  const formatCost = (usd: number): string => {
    if (usd >= 1000) return `$${(usd / 1000).toFixed(1)}K`
    if (usd >= 1) return `$${usd.toFixed(2)}`
    return `$${usd.toFixed(4)}`
  }

  return (
    <div className="space-y-6">
      {/* Total Summary */}
      {spend.total && (
        <div className="grid grid-cols-4 gap-4">
          <MetricCard label="Total Tokens" value={formatTokens(spend.total.totalTokens)} />
          <MetricCard label="Input" value={formatTokens(spend.total.inputTokens)} />
          <MetricCard label="Output" value={formatTokens(spend.total.outputTokens)} />
          <MetricCard label="Cost" value={formatCost(spend.total.costUsd)} highlight />
        </div>
      )}

      {/* By Initiative */}
      {spend.byInitiative && Object.keys(spend.byInitiative).length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-3">By Initiative</h3>
          <div className="bg-gray-800 rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-900">
                <tr>
                  <th className="text-left p-3 font-medium">Initiative</th>
                  <th className="text-right p-3 font-medium">Tokens</th>
                  <th className="text-right p-3 font-medium">Cost</th>
                </tr>
              </thead>
              <tbody>
                {Object.entries(spend.byInitiative)
                  .sort((a, b) => b[1].totalTokens - a[1].totalTokens)
                  .map(([initId, tokens]) => {
                    const init = execution.initiatives.find((i) => i.id === initId)
                    return (
                      <tr key={initId} className="border-t border-gray-700">
                        <td className="p-3">
                          <div className="font-mono text-xs text-gray-400">{initId}</div>
                          {init && <div className="text-gray-300">{init.title}</div>}
                        </td>
                        <td className="text-right p-3 font-mono">
                          {formatTokens(tokens.totalTokens)}
                        </td>
                        <td className="text-right p-3 font-mono text-green-400">
                          {formatCost(tokens.costUsd)}
                        </td>
                      </tr>
                    )
                  })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

function MetricCard({
  label,
  value,
  highlight = false,
}: {
  label: string
  value: string
  highlight?: boolean
}) {
  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="text-sm text-gray-400">{label}</div>
      <div className={`text-2xl font-semibold mt-1 ${highlight ? 'text-green-400' : ''}`}>
        {value}
      </div>
    </div>
  )
}
