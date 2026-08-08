import { useState, useEffect, useMemo } from 'react'
import { getSpend, getExecution } from '../api/client'
import type { SpendResponse, ExecutionResponse, APIRMI, APIInitiative } from '../api/types'
import { BarChart, DonutChart, StackedBarChart } from '../components/charts'
import { LoadingState, ErrorState, EmptyState } from '../components'

type TimeRange = 'week' | 'month' | 'quarter' | 'all'

export function PerformancePanel() {
  const [spend, setSpend] = useState<SpendResponse | null>(null)
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [timeRange, setTimeRange] = useState<TimeRange>('month')

  const reload = () => {
    setError(null)
    Promise.all([getSpend(), getExecution()])
      .then(([s, e]) => {
        setSpend(s)
        setExecution(e)
      })
      .catch((err: Error) => setError(err.message))
  }

  useEffect(() => {
    reload()
  }, [])

  // Filter spend data based on time range
  const filteredSpend = useMemo(() => {
    if (!spend) return null
    if (timeRange === 'all' || !spend.byWeek || spend.byWeek.length === 0) {
      return spend
    }

    const now = new Date()
    let cutoffDate: Date
    switch (timeRange) {
      case 'week':
        cutoffDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      case 'month':
        cutoffDate = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
        break
      case 'quarter':
        cutoffDate = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000)
        break
      default:
        return spend
    }

    // Filter weeks within the time range
    const filteredWeeks = spend.byWeek.filter((w) => new Date(w.start) >= cutoffDate)

    // Aggregate totals from filtered weeks
    const total = {
      inputTokens: 0,
      outputTokens: 0,
      cacheReadTokens: 0,
      cacheCreationTokens: 0,
      totalTokens: 0,
      costUsd: 0,
    }
    const byModel: Record<string, typeof total> = {}

    for (const week of filteredWeeks) {
      if (week.totals) {
        total.inputTokens += week.totals.inputTokens
        total.outputTokens += week.totals.outputTokens
        total.cacheReadTokens += week.totals.cacheReadTokens
        total.cacheCreationTokens += week.totals.cacheCreationTokens
        total.totalTokens += week.totals.totalTokens
        total.costUsd += week.totals.costUsd
      }
      if (week.byModel) {
        for (const [model, tokens] of Object.entries(week.byModel)) {
          if (!byModel[model]) {
            byModel[model] = { inputTokens: 0, outputTokens: 0, cacheReadTokens: 0, cacheCreationTokens: 0, totalTokens: 0, costUsd: 0 }
          }
          byModel[model].inputTokens += tokens.inputTokens
          byModel[model].outputTokens += tokens.outputTokens
          byModel[model].cacheReadTokens += tokens.cacheReadTokens
          byModel[model].cacheCreationTokens += tokens.cacheCreationTokens
          byModel[model].totalTokens += tokens.totalTokens
          byModel[model].costUsd += tokens.costUsd
        }
      }
    }

    return {
      ...spend,
      total,
      byModel,
      byWeek: filteredWeeks,
    }
  }, [spend, timeRange])

  if (error) {
    return <ErrorState message={error} onRetry={reload} />
  }

  if (!spend || !execution || !filteredSpend) {
    return <LoadingState message="Loading performance data..." />
  }

  return (
    <div className="space-y-8">
      {/* Time Range Selector */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Performance</h2>
        <div className="flex gap-1">
          {(['week', 'month', 'quarter', 'all'] as TimeRange[]).map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-3 py-1.5 rounded text-sm ${
                timeRange === range
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {range === 'all' ? 'All Time' : range.charAt(0).toUpperCase() + range.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {/* Cost Summary */}
      <SpendSummary spend={filteredSpend} />

      {/* Model Breakdown Charts */}
      <ModelBreakdownCharts spend={filteredSpend} />

      {/* Time Series Charts */}
      <TimeSeriesCharts spend={filteredSpend} />

      {/* Monthly History */}
      <MonthlyHistory spend={spend} />

      {/* Accomplishments - what shipped */}
      <AccomplishmentsSection
        rmis={execution.rmis ?? []}
        initiatives={execution.initiatives ?? []}
        spend={filteredSpend}
        timeRange={timeRange}
      />

      {/* Cost by Initiative */}
      <CostByInitiative spend={filteredSpend} execution={execution} />
    </div>
  )
}

function SpendSummary({ spend }: { spend: SpendResponse }) {
  if (!spend.hasData || !spend.total) {
    return (
      <EmptyState
        title="No token data"
        description={spend.dataNote ?? 'Token usage data is not available.'}
        hint="Configure omnidevx data directory"
      />
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
    <div className="grid grid-cols-5 gap-4">
      <MetricCard label="Total Tokens" value={formatTokens(spend.total.totalTokens)} />
      <MetricCard label="Input" value={formatTokens(spend.total.inputTokens)} />
      <MetricCard label="Output" value={formatTokens(spend.total.outputTokens)} />
      <MetricCard label="Cache Read" value={formatTokens(spend.total.cacheReadTokens)} />
      <MetricCard label="Cost" value={formatCost(spend.total.costUsd)} highlight />
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

function ModelBreakdownCharts({ spend }: { spend: SpendResponse }) {
  const byModel = spend.byModel

  const tokensByModel = useMemo(() => {
    if (!byModel) return []
    return Object.entries(byModel)
      .map(([name, tokens]) => ({
        name: formatModelName(name),
        value: tokens.totalTokens,
      }))
      .sort((a, b) => b.value - a.value)
  }, [byModel])

  const costByModel = useMemo(() => {
    if (!byModel) return []
    return Object.entries(byModel)
      .map(([name, tokens]) => ({
        name: formatModelName(name),
        value: tokens.costUsd,
      }))
      .sort((a, b) => b.value - a.value)
  }, [byModel])

  const tokensByCategory = useMemo(() => {
    if (!spend.total) return []
    return [
      { name: 'Input', value: spend.total.inputTokens },
      { name: 'Output', value: spend.total.outputTokens },
      { name: 'Cache Read', value: spend.total.cacheReadTokens },
      { name: 'Cache Write', value: spend.total.cacheCreationTokens },
    ].filter((d) => d.value > 0)
  }, [spend.total])

  if (!byModel || Object.keys(byModel).length === 0) {
    return null
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="text-sm font-medium text-gray-400 mb-3">Tokens by Model</h3>
        <DonutChart data={tokensByModel} height={200} />
      </div>
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="text-sm font-medium text-gray-400 mb-3">Cost by Model</h3>
        <DonutChart data={costByModel} height={200} />
      </div>
      <div className="bg-gray-800 rounded-lg p-4">
        <h3 className="text-sm font-medium text-gray-400 mb-3">Tokens by Category</h3>
        <DonutChart data={tokensByCategory} height={200} />
      </div>
    </div>
  )
}

function formatModelName(model: string): string {
  return model
    .replace('claude-', '')
    .replace('-20', ' ')
    .replace(/-/g, ' ')
}

function TimeSeriesCharts({ spend }: { spend: SpendResponse }) {
  const [view, setView] = useState<'week' | 'month'>('week')

  const buckets = view === 'week' ? spend.byWeek : spend.byMonth

  const allModels = useMemo(() => {
    if (!buckets) return []
    const models = new Set<string>()
    for (const bucket of buckets) {
      if (bucket.byModel) {
        for (const model of Object.keys(bucket.byModel)) {
          models.add(model)
        }
      }
    }
    return Array.from(models).sort()
  }, [buckets])

  const costData = useMemo(() => {
    if (!buckets) return []
    return buckets.map((bucket) => {
      const row: Record<string, string | number> = {
        period: formatPeriodLabel(bucket.period),
      }
      for (const model of allModels) {
        row[model] = bucket.byModel?.[model]?.costUsd ?? 0
      }
      return row
    })
  }, [buckets, allModels])

  const tokenData = useMemo(() => {
    if (!buckets) return []
    return buckets.map((bucket) => ({
      period: formatPeriodLabel(bucket.period),
      input: bucket.totals.inputTokens,
      output: bucket.totals.outputTokens,
      cacheRead: bucket.totals.cacheReadTokens,
      cacheWrite: bucket.totals.cacheCreationTokens,
    }))
  }, [buckets])

  if (!buckets || buckets.length === 0) {
    return null
  }

  const series = allModels.map((m) => ({
    key: m,
    name: formatModelName(m),
  }))

  const tokenSeries = [
    { key: 'input', name: 'Input' },
    { key: 'output', name: 'Output' },
    { key: 'cacheRead', name: 'Cache Read' },
    { key: 'cacheWrite', name: 'Cache Write' },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Spend Over Time</h3>
        <div className="flex gap-1">
          <button
            onClick={() => setView('week')}
            className={`px-3 py-1.5 rounded text-sm ${
              view === 'week'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Weekly
          </button>
          <button
            onClick={() => setView('month')}
            className={`px-3 py-1.5 rounded text-sm ${
              view === 'month'
                ? 'bg-blue-500 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            Monthly
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-gray-800 rounded-lg p-4">
          <h4 className="text-sm font-medium text-gray-400 mb-3">Cost by Model</h4>
          <StackedBarChart
            data={costData}
            categoryField="period"
            series={series}
            height={250}
          />
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <h4 className="text-sm font-medium text-gray-400 mb-3">Tokens by Category</h4>
          <StackedBarChart
            data={tokenData}
            categoryField="period"
            series={tokenSeries}
            height={250}
          />
        </div>
      </div>
    </div>
  )
}

function formatPeriodLabel(period: string): string {
  if (period.includes('W')) {
    const [, week] = period.split('-W')
    return `W${week}`
  }
  const [, month] = period.split('-')
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  return monthNames[parseInt(month, 10) - 1] ?? period
}

function MonthlyHistory({ spend }: { spend: SpendResponse }) {
  const [selectedMonth, setSelectedMonth] = useState<string | null>(null)

  const months = spend.byMonth ?? []
  const sortedMonths = useMemo(
    () => [...months].sort((a, b) => b.period.localeCompare(a.period)),
    [months]
  )
  const selected = selectedMonth
    ? sortedMonths.find((m) => m.period === selectedMonth)
    : sortedMonths[0]

  const modelData = useMemo(() => {
    if (!selected?.byModel) return []
    return Object.entries(selected.byModel)
      .map(([name, tokens]) => ({
        name: formatModelName(name),
        value: tokens.costUsd,
      }))
      .sort((a, b) => b.value - a.value)
  }, [selected])

  if (months.length === 0) {
    return null
  }

  const formatMonthLabel = (period: string): string => {
    const [year, month] = period.split('-')
    const monthNames = [
      'January', 'February', 'March', 'April', 'May', 'June',
      'July', 'August', 'September', 'October', 'November', 'December',
    ]
    return `${monthNames[parseInt(month, 10) - 1]} ${year}`
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
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Monthly History</h3>
        <select
          value={selected?.period ?? ''}
          onChange={(e) => setSelectedMonth(e.target.value)}
          className="bg-gray-800 text-gray-200 rounded px-3 py-1.5 text-sm border border-gray-700 focus:outline-none focus:border-blue-500"
        >
          {sortedMonths.map((m) => (
            <option key={m.period} value={m.period}>
              {formatMonthLabel(m.period)}
            </option>
          ))}
        </select>
      </div>

      {selected && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {/* Month Summary */}
          <div className="bg-gray-800 rounded-lg p-4">
            <h4 className="text-sm font-medium text-gray-400 mb-4">
              {formatMonthLabel(selected.period)} Summary
            </h4>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-sm text-gray-500">Total Cost</div>
                <div className="text-2xl font-semibold text-green-400">
                  {formatCost(selected.totals.costUsd)}
                </div>
              </div>
              <div>
                <div className="text-sm text-gray-500">Total Tokens</div>
                <div className="text-2xl font-semibold">
                  {formatTokens(selected.totals.totalTokens)}
                </div>
              </div>
              <div>
                <div className="text-sm text-gray-500">Input</div>
                <div className="text-lg">{formatTokens(selected.totals.inputTokens)}</div>
              </div>
              <div>
                <div className="text-sm text-gray-500">Output</div>
                <div className="text-lg">{formatTokens(selected.totals.outputTokens)}</div>
              </div>
              <div>
                <div className="text-sm text-gray-500">Cache Read</div>
                <div className="text-lg">{formatTokens(selected.totals.cacheReadTokens)}</div>
              </div>
              <div>
                <div className="text-sm text-gray-500">Cache Write</div>
                <div className="text-lg">{formatTokens(selected.totals.cacheCreationTokens)}</div>
              </div>
            </div>
          </div>

          {/* Model Breakdown for Month */}
          <div className="bg-gray-800 rounded-lg p-4">
            <h4 className="text-sm font-medium text-gray-400 mb-3">Cost by Model</h4>
            {modelData.length > 0 ? (
              <DonutChart data={modelData} height={200} />
            ) : (
              <p className="text-gray-500 text-sm">No model data available</p>
            )}
          </div>
        </div>
      )}

      {/* Month Comparison Table */}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-900">
            <tr>
              <th className="text-left p-3 font-medium">Month</th>
              <th className="text-right p-3 font-medium">Tokens</th>
              <th className="text-right p-3 font-medium">Cost</th>
              <th className="text-right p-3 font-medium">vs Prev</th>
            </tr>
          </thead>
          <tbody>
            {sortedMonths.map((month, idx) => {
              const prev = sortedMonths[idx + 1]
              const change = prev
                ? ((month.totals.costUsd - prev.totals.costUsd) / prev.totals.costUsd) * 100
                : null
              return (
                <tr
                  key={month.period}
                  className={`border-t border-gray-700 cursor-pointer hover:bg-gray-750 ${
                    month.period === selected?.period ? 'bg-gray-700' : ''
                  }`}
                  onClick={() => setSelectedMonth(month.period)}
                >
                  <td className="p-3">{formatMonthLabel(month.period)}</td>
                  <td className="text-right p-3 font-mono">
                    {formatTokens(month.totals.totalTokens)}
                  </td>
                  <td className="text-right p-3 font-mono text-green-400">
                    {formatCost(month.totals.costUsd)}
                  </td>
                  <td className="text-right p-3 font-mono">
                    {change !== null ? (
                      <span className={change >= 0 ? 'text-red-400' : 'text-green-400'}>
                        {change >= 0 ? '+' : ''}{change.toFixed(0)}%
                      </span>
                    ) : (
                      <span className="text-gray-600">—</span>
                    )}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function AccomplishmentsSection({
  rmis,
  initiatives,
  spend,
  timeRange,
}: {
  rmis: APIRMI[]
  initiatives: APIInitiative[]
  spend: SpendResponse
  timeRange: TimeRange
}) {
  const initMap = useMemo(
    () => new Map(initiatives.map((i) => [i.id ?? '', i.title ?? ''])),
    [initiatives]
  )

  const cutoff = useMemo(() => {
    const now = new Date()
    switch (timeRange) {
      case 'week':
        return new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
      case 'month':
        return new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
      case 'quarter':
        return new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000)
      default:
        return new Date(0)
    }
  }, [timeRange])

  const completed = useMemo(() => {
    return rmis
      .filter((r) => r.completedAt && new Date(r.completedAt) >= cutoff)
      .sort((a, b) => {
        const da = new Date(a.completedAt!).getTime()
        const db = new Date(b.completedAt!).getTime()
        return db - da
      })
  }, [rmis, cutoff])

  const byWeek = useMemo(() => {
    const weeks: Record<string, APIRMI[]> = {}
    for (const r of completed) {
      const d = new Date(r.completedAt!)
      const weekStart = getWeekStart(d)
      const key = weekStart.toISOString().slice(0, 10)
      if (!weeks[key]) weeks[key] = []
      weeks[key].push(r)
    }
    return Object.entries(weeks)
      .sort((a, b) => b[0].localeCompare(a[0]))
      .map(([week, items]) => ({ week, items }))
  }, [completed])

  const chartData = useMemo(() => {
    return byWeek
      .slice(0, 12)
      .reverse()
      .map((w) => ({
        label: formatWeekLabel(w.week),
        values: { completed: w.items.length },
      }))
  }, [byWeek])

  if (completed.length === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold mb-2">Accomplishments</h3>
        <p className="text-gray-400 text-sm">No RMIs completed in this time range.</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">
        Accomplishments <span className="text-gray-500 font-normal">({completed.length} shipped)</span>
      </h3>

      {/* Velocity Chart */}
      {chartData.length > 1 && (
        <div className="bg-gray-800 rounded-lg p-4">
          <h4 className="text-sm font-medium text-gray-400 mb-3">Weekly Velocity</h4>
          <div className="h-40">
            <BarChart
              data={chartData}
              series={['completed']}
              height={160}
            />
          </div>
        </div>
      )}

      {/* Grouped by Week */}
      <div className="space-y-4">
        {byWeek.slice(0, 8).map((group) => (
          <div key={group.week} className="bg-gray-800 rounded-lg p-4">
            <h4 className="text-sm font-medium text-gray-400 mb-3">
              Week of {formatWeekLabel(group.week)}
            </h4>
            <div className="space-y-2">
              {group.items.map((rmi) => (
                <AccomplishmentRow
                  key={rmi.id}
                  rmi={rmi}
                  initTitle={initMap.get(rmi.initiativeId ?? '')}
                  cost={rmi.id ? spend.byRmi?.[rmi.id]?.costUsd : undefined}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function AccomplishmentRow({
  rmi,
  initTitle,
  cost,
}: {
  rmi: APIRMI
  initTitle?: string
  cost?: number
}) {
  return (
    <div className="flex items-center justify-between py-2 px-3 bg-gray-900 rounded">
      <div className="flex items-center gap-3 min-w-0">
        <span className="text-lg">{typeIcon(rmi.type)}</span>
        <div className="min-w-0">
          <div className="text-sm font-medium truncate">{rmi.title}</div>
          <div className="text-xs text-gray-500 flex items-center gap-2">
            <span className="font-mono">{rmi.id}</span>
            {initTitle && <span className="truncate">• {initTitle}</span>}
          </div>
        </div>
      </div>
      <div className="flex items-center gap-3 flex-shrink-0">
        {cost !== undefined && cost > 0 && (
          <span className="text-xs text-gray-400">${cost.toFixed(2)}</span>
        )}
        <span className="text-xs text-gray-500">
          {new Date(rmi.completedAt!).toLocaleDateString()}
        </span>
      </div>
    </div>
  )
}

function CostByInitiative({
  spend,
  execution,
}: {
  spend: SpendResponse
  execution: ExecutionResponse
}) {
  if (!spend.byInitiative || Object.keys(spend.byInitiative).length === 0) {
    return null
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

  const sorted = Object.entries(spend.byInitiative).sort(
    (a, b) => b[1].totalTokens - a[1].totalTokens
  )

  return (
    <div>
      <h3 className="text-lg font-semibold mb-3">Cost by Initiative</h3>
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
            {sorted.map(([initId, tokens]) => {
              const init = (execution.initiatives ?? []).find((i) => i.id === initId)
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
  )
}

function getWeekStart(d: Date): Date {
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1)
  return new Date(d.getFullYear(), d.getMonth(), diff)
}

function formatWeekLabel(isoDate: string): string {
  const d = new Date(isoDate)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
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
