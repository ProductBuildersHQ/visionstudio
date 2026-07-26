import { useEffect, useState } from 'react'
import { api } from '../../services/api'
import { LoadingState, ErrorState } from '../toolkit'
import type {
  DashforgeDashboard,
  DashforgeWidget,
  DashforgeMetricConfig,
  DashforgeChartConfig,
  DashforgeTableConfig,
  DevXPeriodEntry,
} from './types'

/**
 * DevXDashboardView renders the OmniDevX dashboard-IR export produced by
 * `devfolio devx dashboard`. VisionStudio never queries the OmniDevX event
 * store or computes metrics itself — it only renders whatever dashboard
 * JSON devfolio already generated and disclosure-scoped.
 *
 * Supports both the default dashboard.json and per-period reports
 * (weekly, monthly, quarterly) with model breakdown charts.
 */
export function DevXDashboardView() {
  const [dashboard, setDashboard] = useState<DashforgeDashboard | null>(null)
  const [periods, setPeriods] = useState<DevXPeriodEntry[]>([])
  const [selectedPeriod, setSelectedPeriod] = useState<DevXPeriodEntry | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadPeriods()
  }, [])

  useEffect(() => {
    if (selectedPeriod) {
      loadPeriodDashboard(selectedPeriod)
    } else {
      loadDefaultDashboard()
    }
  }, [selectedPeriod])

  async function loadPeriods() {
    try {
      const p = await api.getDevXPeriods()
      setPeriods(p)
      // Auto-select most recent monthly if available
      const monthly = p.find((e) => e.type === 'monthly')
      if (monthly) {
        setSelectedPeriod(monthly)
      } else if (p.length > 0) {
        setSelectedPeriod(p[0])
      }
    } catch {
      // Periods endpoint may not exist; fall back to default dashboard
      loadDefaultDashboard()
    }
  }

  async function loadDefaultDashboard() {
    setIsLoading(true)
    setError(null)
    try {
      const dash = await api.getDevXDashboard()
      setDashboard(dash)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard')
    } finally {
      setIsLoading(false)
    }
  }

  async function loadPeriodDashboard(period: DevXPeriodEntry) {
    setIsLoading(true)
    setError(null)
    try {
      const dash = await api.getDevXPeriodDashboard(period.type, period.label)
      setDashboard(dash)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load period dashboard')
    } finally {
      setIsLoading(false)
    }
  }

  function handlePeriodChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const value = e.target.value
    if (value === '') {
      setSelectedPeriod(null)
    } else {
      const [type, label] = value.split(':')
      const period = periods.find((p) => p.type === type && p.label === label)
      if (period) setSelectedPeriod(period)
    }
  }

  if (isLoading) {
    return <LoadingState message="Loading DevX dashboard..." />
  }

  if (error) {
    return (
      <ErrorState
        message={error}
        hint="Generate a dashboard with: devfolio devx dashboard --person <personId> -o ~/.plexusone/omnidevx/dashboard.json"
        onRetry={() => (selectedPeriod ? loadPeriodDashboard(selectedPeriod) : loadDefaultDashboard())}
      />
    )
  }

  if (!dashboard) return null

  const dataById = new Map(dashboard.dataSources.map((ds) => [ds.id, ds.data]))
  const metricWidgets = dashboard.widgets.filter((w) => w.type === 'metric')
  const chartWidgets = dashboard.widgets.filter((w) => w.type === 'chart')
  const tableWidgets = dashboard.widgets.filter((w) => w.type === 'table')

  return (
    <div className="h-full flex flex-col bg-va-bg">
      <div className="flex items-center justify-between px-4 py-2 border-b border-va-border bg-va-sidebar">
        <h2 className="text-lg font-semibold text-va-text">{dashboard.title}</h2>
        <div className="flex items-center gap-3">
          {periods.length > 0 && (
            <select
              value={selectedPeriod ? `${selectedPeriod.type}:${selectedPeriod.label}` : ''}
              onChange={handlePeriodChange}
              className="text-sm bg-va-panel border border-va-border rounded px-2 py-1 text-va-text"
            >
              <option value="">Default</option>
              {periods.map((p) => (
                <option key={`${p.type}:${p.label}`} value={`${p.type}:${p.label}`}>
                  {p.label} ({p.type})
                </option>
              ))}
            </select>
          )}
          <button
            onClick={() => (selectedPeriod ? loadPeriodDashboard(selectedPeriod) : loadDefaultDashboard())}
            className="p-1.5 rounded hover:bg-va-panel text-va-text-muted hover:text-va-text transition-colors text-sm"
          >
            Refresh
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4 space-y-6">
        {metricWidgets.length > 0 && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            {metricWidgets.map((w) => (
              <MetricTile key={w.id} widget={w} data={resolveData(w, dataById)} />
            ))}
          </div>
        )}

        {chartWidgets.length > 0 && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {chartWidgets.map((w) => (
              <ChartWidget
                key={w.id}
                widget={w}
                data={(resolveData(w, dataById) as Record<string, unknown>[]) || []}
              />
            ))}
          </div>
        )}

        {tableWidgets.map((w) => (
          <div key={w.id} className="bg-va-panel border border-va-border rounded-lg overflow-hidden">
            <h3 className="text-sm font-semibold text-va-text px-4 py-3 border-b border-va-border">{w.title}</h3>
            <DataTable
              config={w.config as DashforgeTableConfig}
              rows={(resolveData(w, dataById) as Record<string, unknown>[]) || []}
            />
          </div>
        ))}
      </div>
    </div>
  )
}

function resolveData(widget: DashforgeWidget, dataById: Map<string, unknown>): unknown {
  return widget.dataSourceId ? dataById.get(widget.dataSourceId) : undefined
}

function formatValue(value: number, cfg: DashforgeMetricConfig): string {
  const decimals = cfg.formatOptions?.decimals ?? 0
  if (cfg.format === 'percent') {
    return `${value.toFixed(decimals)}${cfg.formatOptions?.suffix ?? '%'}`
  }
  if (cfg.format === 'currency') {
    return `${cfg.formatOptions?.prefix ?? '$'}${value.toLocaleString('en-US', { maximumFractionDigits: decimals })}`
  }
  return value.toLocaleString('en-US', { maximumFractionDigits: decimals })
}

function MetricTile({ widget, data }: { widget: DashforgeWidget; data: unknown }) {
  const cfg = widget.config as DashforgeMetricConfig
  const record = data as Record<string, number> | undefined
  const raw = record?.[cfg.valueField]
  const display = typeof raw === 'number' ? formatValue(raw, cfg) : '—'

  return (
    <div className="bg-va-panel border border-va-border rounded-lg p-3.5">
      <div className="text-xs text-va-text-muted mb-1.5">{widget.title}</div>
      <div className="text-2xl font-semibold text-va-text">{display}</div>
    </div>
  )
}

function DataTable({ config, rows }: { config: DashforgeTableConfig; rows: Record<string, unknown>[] }) {
  if (rows.length === 0) {
    return <div className="p-4 text-sm text-va-text-muted">No data</div>
  }
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-va-border">
            {config.columns.map((col) => (
              <th
                key={col.field}
                className={`text-xs uppercase tracking-wide text-va-text-muted font-semibold px-4 py-2 ${
                  col.align === 'right' ? 'text-right' : 'text-left'
                }`}
              >
                {col.header ?? col.field}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-b border-va-border last:border-0">
              {config.columns.map((col) => (
                <td
                  key={col.field}
                  className={`px-4 py-2 text-va-text-muted ${col.align === 'right' ? 'text-right tabular-nums' : ''}`}
                >
                  {String(row[col.field] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

/** ChartWidget dispatches to the appropriate renderer based on mark geometries */
function ChartWidget({ widget, data }: { widget: DashforgeWidget; data: Record<string, unknown>[] }) {
  const config = widget.config as DashforgeChartConfig
  const geometries = new Set(config.marks.map((m) => m.geometry))

  return (
    <div className="bg-va-panel border border-va-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-va-text mb-3">{widget.title}</h3>
      {geometries.has('pie') ? (
        <PieChart config={config} rows={data} />
      ) : geometries.has('bar') ? (
        <BarChart config={config} rows={data} />
      ) : (
        <LineChart config={config} rows={data} />
      )}
    </div>
  )
}

const CHART_COLORS = [
  '#2a78d6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6',
  '#06b6d4', '#ec4899', '#84cc16', '#f97316', '#6366f1',
]

/** Pie/Donut chart for model breakdowns */
function PieChart({ config, rows }: { config: DashforgeChartConfig; rows: Record<string, unknown>[] }) {
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)

  if (rows.length === 0) {
    return <div className="text-sm text-va-text-muted h-40 flex items-center justify-center">No data</div>
  }

  const mark = config.marks.find((m) => m.geometry === 'pie')
  if (!mark) return null

  const valueField = mark.encode.value ?? 'value'
  const nameField = mark.encode.name ?? 'model'

  const total = rows.reduce((sum, r) => sum + Number(r[valueField] ?? 0), 0)
  if (total === 0) {
    return <div className="text-sm text-va-text-muted h-40 flex items-center justify-center">No data</div>
  }

  const size = 180
  const cx = size / 2
  const cy = size / 2
  const outerR = 70
  const innerR = 40 // donut hole

  let cumAngle = -Math.PI / 2

  const slices = rows.map((r, i) => {
    const value = Number(r[valueField] ?? 0)
    const pct = value / total
    const startAngle = cumAngle
    const endAngle = cumAngle + pct * 2 * Math.PI
    cumAngle = endAngle

    const largeArc = pct > 0.5 ? 1 : 0
    const x1 = cx + outerR * Math.cos(startAngle)
    const y1 = cy + outerR * Math.sin(startAngle)
    const x2 = cx + outerR * Math.cos(endAngle)
    const y2 = cy + outerR * Math.sin(endAngle)
    const ix1 = cx + innerR * Math.cos(startAngle)
    const iy1 = cy + innerR * Math.sin(startAngle)
    const ix2 = cx + innerR * Math.cos(endAngle)
    const iy2 = cy + innerR * Math.sin(endAngle)

    const path = [
      `M ${x1} ${y1}`,
      `A ${outerR} ${outerR} 0 ${largeArc} 1 ${x2} ${y2}`,
      `L ${ix2} ${iy2}`,
      `A ${innerR} ${innerR} 0 ${largeArc} 0 ${ix1} ${iy1}`,
      'Z',
    ].join(' ')

    return {
      name: String(r[nameField] ?? ''),
      value,
      pct,
      path,
      color: CHART_COLORS[i % CHART_COLORS.length],
    }
  })

  return (
    <div className="flex items-start gap-4">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        {slices.map((s, i) => (
          <path
            key={i}
            d={s.path}
            fill={s.color}
            opacity={hoverIdx === null || hoverIdx === i ? 1 : 0.4}
            onMouseEnter={() => setHoverIdx(i)}
            onMouseLeave={() => setHoverIdx(null)}
            className="transition-opacity cursor-pointer"
          />
        ))}
        {hoverIdx !== null && (
          <text x={cx} y={cy} textAnchor="middle" dominantBaseline="middle" className="fill-va-text text-sm font-semibold">
            {(slices[hoverIdx].pct * 100).toFixed(1)}%
          </text>
        )}
      </svg>
      <div className="flex-1 text-xs space-y-1">
        {slices.map((s, i) => (
          <div
            key={i}
            className={`flex items-center gap-2 ${hoverIdx === i ? 'font-semibold' : ''}`}
            onMouseEnter={() => setHoverIdx(i)}
            onMouseLeave={() => setHoverIdx(null)}
          >
            <span className="w-2.5 h-2.5 rounded-sm inline-block flex-shrink-0" style={{ background: s.color }} />
            <span className="text-va-text truncate">{s.name}</span>
            <span className="text-va-text-muted ml-auto tabular-nums">{s.value.toLocaleString()}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

/** Stacked bar chart for period breakdowns by model */
function BarChart({ rows }: { config: DashforgeChartConfig; rows: Record<string, unknown>[] }) {
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)

  if (rows.length === 0) {
    return <div className="text-sm text-va-text-muted h-40 flex items-center justify-center">No data</div>
  }

  // Get all model keys from data (excluding 'period' field)
  const modelKeys = new Set<string>()
  rows.forEach((r) => {
    Object.keys(r).forEach((k) => {
      if (k !== 'period' && typeof r[k] === 'number') {
        modelKeys.add(k)
      }
    })
  })
  const models = Array.from(modelKeys)

  if (models.length === 0) {
    return <div className="text-sm text-va-text-muted h-40 flex items-center justify-center">No data</div>
  }

  const width = 480
  const height = 200
  const padL = 50
  const padR = 12
  const padT = 20
  const padB = 30
  const plotW = width - padL - padR
  const plotH = height - padT - padB
  const n = rows.length
  const barW = Math.min(40, plotW / n - 4)

  // Calculate max stacked value
  const maxV = Math.max(
    1,
    ...rows.map((r) => models.reduce((sum, m) => sum + Number(r[m] ?? 0), 0))
  )
  const niceMax = niceCeil(maxV)

  const x = (i: number) => padL + (plotW / n) * (i + 0.5)
  const y = (v: number) => padT + plotH - (plotH * v) / (niceMax || 1)

  return (
    <div className="relative">
      <div className="flex gap-4 mb-2 text-xs text-va-text-muted flex-wrap">
        {models.map((m, i) => (
          <span key={m} className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-sm inline-block" style={{ background: CHART_COLORS[i % CHART_COLORS.length] }} />
            {m}
          </span>
        ))}
      </div>
      <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-auto overflow-visible">
        {[0, niceMax / 2, niceMax].map((v) => (
          <g key={v}>
            <line x1={padL} y1={y(v)} x2={width - padR} y2={y(v)} className="stroke-va-border" strokeWidth={1} />
            <text x={padL - 6} y={y(v) + 3} textAnchor="end" className="fill-va-text-muted" fontSize={9}>
              {formatLargeNumber(v)}
            </text>
          </g>
        ))}
        {rows.map((r, i) => {
          let cumY = 0
          return (
            <g key={i} onMouseEnter={() => setHoverIdx(i)} onMouseLeave={() => setHoverIdx(null)}>
              {models.map((m, mi) => {
                const v = Number(r[m] ?? 0)
                const barH = (plotH * v) / (niceMax || 1)
                const yPos = y(cumY + v)
                cumY += v
                return (
                  <rect
                    key={m}
                    x={x(i) - barW / 2}
                    y={yPos}
                    width={barW}
                    height={barH}
                    fill={CHART_COLORS[mi % CHART_COLORS.length]}
                    opacity={hoverIdx === null || hoverIdx === i ? 1 : 0.5}
                    className="transition-opacity"
                  />
                )
              })}
            </g>
          )
        })}
        {rows.map((r, i) => (
          <text key={i} x={x(i)} y={height - 8} textAnchor="middle" className="fill-va-text-muted" fontSize={9}>
            {String(r['period'] ?? '')}
          </text>
        ))}
      </svg>
      {hoverIdx !== null && (
        <div className="absolute top-0 right-0 bg-va-bg border border-va-border rounded px-2 py-1.5 text-xs pointer-events-none">
          <div className="text-va-text-muted mb-0.5">{String(rows[hoverIdx]['period'] ?? '')}</div>
          {models.map((m, mi) => (
            <div key={m} className="flex items-center gap-2 text-va-text">
              <span className="w-2 h-2 rounded-sm" style={{ background: CHART_COLORS[mi % CHART_COLORS.length] }} />
              {m}: <strong>{Number(rows[hoverIdx][m] ?? 0).toLocaleString()}</strong>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

/** Line chart for time series */
function LineChart({ config, rows }: { config: DashforgeChartConfig; rows: Record<string, unknown>[] }) {
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)

  const unsupported = config.marks.filter((m) => m.geometry !== 'line')
  const marks = config.marks.filter((m) => m.geometry === 'line')
  if (marks.length === 0 || rows.length === 0) {
    return <div className="text-sm text-va-text-muted h-40 flex items-center justify-center">No data</div>
  }

  const width = 480
  const height = 200
  const padL = 40
  const padR = 12
  const padT = 12
  const padB = 24
  const plotW = width - padL - padR
  const plotH = height - padT - padB
  const n = rows.length

  const seriesValues = marks.map((m) => rows.map((r) => Number(r[m.encode.y ?? ''] ?? 0)))
  const maxV = Math.max(1, ...seriesValues.flat())
  const niceMax = niceCeil(maxV)

  const x = (i: number) => padL + (n === 1 ? plotW / 2 : (plotW * i) / (n - 1))
  const y = (v: number) => padT + plotH - (plotH * v) / (niceMax || 1)

  const xField = marks[0].encode.x ?? ''

  return (
    <div className="relative">
      {marks.length > 1 && (
        <div className="flex gap-4 mb-2 text-xs text-va-text-muted">
          {marks.map((m) => (
            <span key={m.id} className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-sm inline-block" style={{ background: m.style?.color }} />
              {m.name ?? m.id}
            </span>
          ))}
        </div>
      )}
      <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-auto overflow-visible">
        {[0, niceMax / 2, niceMax].map((v) => (
          <g key={v}>
            <line x1={padL} y1={y(v)} x2={width - padR} y2={y(v)} className="stroke-va-border" strokeWidth={1} />
            <text x={padL - 6} y={y(v) + 3} textAnchor="end" className="fill-va-text-muted" fontSize={9}>
              {Math.round(v).toLocaleString()}
            </text>
          </g>
        ))}
        {[0, Math.floor((n - 1) / 2), n - 1].map((i) => (
          <text key={i} x={x(i)} y={height - 4} textAnchor="middle" className="fill-va-text-muted" fontSize={9}>
            {String(rows[i][xField] ?? '')}
          </text>
        ))}
        {marks.map((m, mi) => (
          <polyline
            key={m.id}
            fill="none"
            stroke={m.style?.color ?? '#2a78d6'}
            strokeWidth={2}
            strokeLinejoin="round"
            strokeLinecap="round"
            points={seriesValues[mi].map((v, i) => `${x(i)},${y(v)}`).join(' ')}
          />
        ))}
        {hoverIdx !== null && (
          <line x1={x(hoverIdx)} x2={x(hoverIdx)} y1={padT} y2={padT + plotH} className="stroke-va-border" strokeDasharray="2 2" />
        )}
        {marks.map((m, mi) =>
          hoverIdx !== null ? (
            <circle
              key={m.id}
              cx={x(hoverIdx)}
              cy={y(seriesValues[mi][hoverIdx])}
              r={4}
              fill={m.style?.color ?? '#2a78d6'}
              className="stroke-va-panel"
              strokeWidth={2}
            />
          ) : null
        )}
        {rows.map((_, i) => (
          <rect
            key={i}
            x={padL + (plotW * (i - 0.5)) / Math.max(1, n - 1)}
            y={padT}
            width={plotW / Math.max(1, n - 1)}
            height={plotH}
            fill="transparent"
            onMouseEnter={() => setHoverIdx(i)}
            onMouseLeave={() => setHoverIdx(null)}
          />
        ))}
      </svg>
      {hoverIdx !== null && (
        <div className="absolute top-0 right-0 bg-va-bg border border-va-border rounded px-2 py-1.5 text-xs pointer-events-none">
          <div className="text-va-text-muted mb-0.5">{String(rows[hoverIdx][xField] ?? '')}</div>
          {marks.map((m, mi) => (
            <div key={m.id} className="text-va-text">
              {m.name ?? m.id}: <strong>{seriesValues[mi][hoverIdx].toLocaleString()}</strong>
            </div>
          ))}
        </div>
      )}
      {unsupported.length > 0 && (
        <div className="text-xs text-va-text-muted mt-2">
          {unsupported.length} series with unsupported geometry ({unsupported.map((m) => m.geometry).join(', ')}) not shown.
        </div>
      )}
    </div>
  )
}

function niceCeil(v: number): number {
  if (v <= 0) return 1
  const mag = Math.pow(10, Math.floor(Math.log10(v)))
  const norm = v / mag
  const step = norm <= 1 ? 1 : norm <= 2 ? 2 : norm <= 5 ? 5 : 10
  return step * mag
}

function formatLargeNumber(v: number): string {
  if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(1)}M`
  if (v >= 1_000) return `${(v / 1_000).toFixed(0)}K`
  return Math.round(v).toLocaleString()
}
