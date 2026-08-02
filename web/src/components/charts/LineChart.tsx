import { useState } from 'react'
import { getColor, niceCeil } from './colors'

export interface LineDataPoint {
  label: string
  values: Record<string, number>
}

export interface LineSeries {
  key: string
  name: string
  color?: string
}

interface LineChartProps {
  data: LineDataPoint[]
  series: LineSeries[]
  width?: number
  height?: number
}

export function LineChart({
  data,
  series,
  width = 480,
  height = 200,
}: LineChartProps) {
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)

  if (data.length === 0 || series.length === 0) {
    return (
      <div className="text-sm text-gray-400 h-40 flex items-center justify-center">
        No data
      </div>
    )
  }

  const padL = 40
  const padR = 12
  const padT = 12
  const padB = 24
  const plotW = width - padL - padR
  const plotH = height - padT - padB
  const n = data.length

  const allValues = data.flatMap((d) => series.map((s) => d.values[s.key] ?? 0))
  const maxV = Math.max(1, ...allValues)
  const niceMax = niceCeil(maxV)

  const x = (i: number) => padL + (n === 1 ? plotW / 2 : (plotW * i) / (n - 1))
  const y = (v: number) => padT + plotH - (plotH * v) / (niceMax || 1)

  return (
    <div className="relative">
      {series.length > 1 && (
        <div className="flex gap-4 mb-2 text-xs text-gray-400">
          {series.map((s, i) => (
            <span key={s.key} className="flex items-center gap-1.5">
              <span
                className="w-2.5 h-2.5 rounded-sm inline-block"
                style={{ background: s.color ?? getColor(i) }}
              />
              {s.name}
            </span>
          ))}
        </div>
      )}
      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="w-full h-auto overflow-visible"
        onMouseLeave={() => setHoverIdx(null)}
      >
        {[0, niceMax / 2, niceMax].map((v) => (
          <g key={v}>
            <line
              x1={padL}
              y1={y(v)}
              x2={width - padR}
              y2={y(v)}
              className="stroke-gray-700"
              strokeWidth={1}
            />
            <text
              x={padL - 6}
              y={y(v) + 3}
              textAnchor="end"
              className="fill-gray-500"
              fontSize={9}
            >
              {Math.round(v).toLocaleString()}
            </text>
          </g>
        ))}
        {/* X-axis labels */}
        {[0, Math.floor((n - 1) / 2), n - 1].map((i) => (
          <text
            key={i}
            x={x(i)}
            y={height - 4}
            textAnchor="middle"
            className="fill-gray-500"
            fontSize={9}
          >
            {data[i].label}
          </text>
        ))}
        {/* Lines */}
        {series.map((s, si) => (
          <polyline
            key={s.key}
            fill="none"
            stroke={s.color ?? getColor(si)}
            strokeWidth={2}
            strokeLinejoin="round"
            strokeLinecap="round"
            points={data.map((d, i) => `${x(i)},${y(d.values[s.key] ?? 0)}`).join(' ')}
          />
        ))}
        {/* Hover line */}
        {hoverIdx !== null && (
          <line
            x1={x(hoverIdx)}
            x2={x(hoverIdx)}
            y1={padT}
            y2={padT + plotH}
            className="stroke-gray-600"
            strokeDasharray="2 2"
          />
        )}
        {/* Dots */}
        {series.map((s, si) =>
          data.map((d, i) => (
            <circle
              key={`${s.key}-${i}`}
              cx={x(i)}
              cy={y(d.values[s.key] ?? 0)}
              r={hoverIdx === i ? 5 : 3}
              fill={s.color ?? getColor(si)}
              className="transition-all"
              onMouseEnter={() => setHoverIdx(i)}
            />
          ))
        )}
      </svg>
      {hoverIdx !== null && (
        <div className="absolute top-0 right-0 bg-gray-900 border border-gray-700 rounded px-2 py-1.5 text-xs pointer-events-none">
          <div className="text-gray-400 mb-0.5">{data[hoverIdx].label}</div>
          {series.map((s, si) => (
            <div key={s.key} className="flex items-center gap-2 text-gray-300">
              <span
                className="w-2 h-2 rounded-sm"
                style={{ background: s.color ?? getColor(si) }}
              />
              {s.name}: <strong>{(data[hoverIdx].values[s.key] ?? 0).toLocaleString()}</strong>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
