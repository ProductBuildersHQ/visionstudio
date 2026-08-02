import { useState } from 'react'
import { getColor, formatLargeNumber, niceCeil } from './colors'

export interface BarDataPoint {
  label: string
  values: Record<string, number>
}

interface BarChartProps {
  data: BarDataPoint[]
  series: string[]
  width?: number
  height?: number
  stacked?: boolean
}

export function BarChart({
  data,
  series,
  width = 480,
  height = 200,
  stacked = true,
}: BarChartProps) {
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)

  if (data.length === 0 || series.length === 0) {
    return (
      <div className="text-sm text-gray-400 h-40 flex items-center justify-center">
        No data
      </div>
    )
  }

  const padL = 50
  const padR = 12
  const padT = 20
  const padB = 30
  const plotW = width - padL - padR
  const plotH = height - padT - padB
  const n = data.length
  const barW = Math.min(40, plotW / n - 4)

  const maxV = stacked
    ? Math.max(1, ...data.map((d) => series.reduce((sum, s) => sum + (d.values[s] ?? 0), 0)))
    : Math.max(1, ...data.flatMap((d) => series.map((s) => d.values[s] ?? 0)))
  const niceMax = niceCeil(maxV)

  const x = (i: number) => padL + (plotW / n) * (i + 0.5)
  const y = (v: number) => padT + plotH - (plotH * v) / (niceMax || 1)

  return (
    <div className="relative">
      <div className="flex gap-4 mb-2 text-xs text-gray-400 flex-wrap">
        {series.map((s, i) => (
          <span key={s} className="flex items-center gap-1.5">
            <span
              className="w-2.5 h-2.5 rounded-sm inline-block"
              style={{ background: getColor(i) }}
            />
            {s}
          </span>
        ))}
      </div>
      <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-auto overflow-visible">
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
              {formatLargeNumber(v)}
            </text>
          </g>
        ))}
        {data.map((d, i) => {
          let cumY = 0
          return (
            <g
              key={i}
              onMouseEnter={() => setHoverIdx(i)}
              onMouseLeave={() => setHoverIdx(null)}
            >
              {series.map((s, si) => {
                const v = d.values[s] ?? 0
                const barH = (plotH * v) / (niceMax || 1)
                const yPos = y(cumY + v)
                if (stacked) cumY += v
                return (
                  <rect
                    key={s}
                    x={x(i) - barW / 2}
                    y={yPos}
                    width={barW}
                    height={barH}
                    fill={getColor(si)}
                    opacity={hoverIdx === null || hoverIdx === i ? 1 : 0.5}
                    className="transition-opacity"
                  />
                )
              })}
            </g>
          )
        })}
        {data.map((d, i) => (
          <text
            key={i}
            x={x(i)}
            y={height - 8}
            textAnchor="middle"
            className="fill-gray-500"
            fontSize={9}
          >
            {d.label}
          </text>
        ))}
      </svg>
      {hoverIdx !== null && (
        <div className="absolute top-0 right-0 bg-gray-900 border border-gray-700 rounded px-2 py-1.5 text-xs pointer-events-none">
          <div className="text-gray-400 mb-0.5">{data[hoverIdx].label}</div>
          {series.map((s, si) => (
            <div key={s} className="flex items-center gap-2 text-gray-300">
              <span
                className="w-2 h-2 rounded-sm"
                style={{ background: getColor(si) }}
              />
              {s}: <strong>{(data[hoverIdx].values[s] ?? 0).toLocaleString()}</strong>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
