import { useEffect, useState } from 'react'
import { api } from '../../services/api'
import type {
  DashforgeDashboard,
  DashforgeWidget,
  DashforgeMetricConfig,
  DashforgeChartConfig,
  DashforgeTableConfig,
} from './types'

/**
 * DevXDashboardView renders the OmniDevX dashboard-IR export produced by
 * `devfolio devx dashboard`. VisionStudio never queries the OmniDevX event
 * store or computes metrics itself — it only renders whatever dashboard
 * JSON devfolio already generated and disclosure-scoped.
 *
 * Not project-scoped: OmniDevX activity is person/machine-scoped, so this
 * view sits alongside Organization in the sidebar rather than under a
 * specific tracked project.
 */
export function DevXDashboardView() {
  const [dashboard, setDashboard] = useState<DashforgeDashboard | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    load()
  }, [])

  async function load() {
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

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full bg-va-bg">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-2 border-va-accent border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-va-text-muted">Loading DevX dashboard...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-va-bg gap-4">
        <div className="text-center max-w-md">
          <div className="text-va-error mb-4">{error}</div>
          <p className="text-va-text-muted text-sm mb-4">
            Generate a dashboard with{' '}
            <code className="text-xs bg-va-panel px-1.5 py-0.5 rounded">
              devfolio devx dashboard --person &lt;personId&gt; -o ~/.plexusone/omnidevx/dashboard.json
            </code>
          </p>
          <button
            onClick={load}
            className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
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
        <button
          onClick={load}
          className="p-1.5 rounded hover:bg-va-panel text-va-text-muted hover:text-va-text transition-colors text-sm"
        >
          Refresh
        </button>
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
              <div key={w.id} className="bg-va-panel border border-va-border rounded-lg p-4">
                <h3 className="text-sm font-semibold text-va-text mb-3">{w.title}</h3>
                <LineChart
                  config={w.config as DashforgeChartConfig}
                  rows={(resolveData(w, dataById) as Record<string, unknown>[]) || []}
                />
              </div>
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
    // Mirrors dashforge's own viewer convention: values are pre-scaled to
    // 0-100 by the exporter, so this always just appends "%".
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

/** Minimal dependency-free line chart rendering marks[]/geometry/encode — the
 * same shape dashforge's own viewer parses — directly against the resolved
 * data rows. Handles the "line" geometry only, matching what devfolio's
 * exporter currently emits; other geometries fall back to a "not rendered"
 * note rather than silently showing nothing.
 */
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
