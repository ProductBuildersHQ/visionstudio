import { useState } from 'react'
import { getColor } from './colors'

export interface RadarAxis {
  key: string
  label: string
  max: number
}

export interface RadarDataset {
  name: string
  values: Record<string, number>
  color?: string
}

interface RadarChartProps {
  axes: RadarAxis[]
  datasets: RadarDataset[]
  size?: number
}

export function RadarChart({ axes, datasets, size = 300 }: RadarChartProps) {
  const [hoverDataset, setHoverDataset] = useState<number | null>(null)

  if (axes.length < 3) {
    return (
      <div className="text-sm text-gray-400 h-40 flex items-center justify-center">
        Radar chart requires at least 3 axes
      </div>
    )
  }

  const cx = size / 2
  const cy = size / 2
  const radius = size / 2 - 40
  const levels = 5
  const angleStep = (2 * Math.PI) / axes.length

  const getPoint = (axisIdx: number, value: number, max: number) => {
    const angle = -Math.PI / 2 + axisIdx * angleStep
    const r = (value / max) * radius
    return {
      x: cx + r * Math.cos(angle),
      y: cy + r * Math.sin(angle),
    }
  }

  const getAxisEnd = (axisIdx: number) => {
    const angle = -Math.PI / 2 + axisIdx * angleStep
    return {
      x: cx + radius * Math.cos(angle),
      y: cy + radius * Math.sin(angle),
      labelX: cx + (radius + 20) * Math.cos(angle),
      labelY: cy + (radius + 20) * Math.sin(angle),
    }
  }

  return (
    <div className="flex items-start gap-6">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        {/* Grid levels */}
        {Array.from({ length: levels }, (_, i) => {
          const levelRadius = (radius * (i + 1)) / levels
          const points = axes
            .map((_, ai) => {
              const angle = -Math.PI / 2 + ai * angleStep
              return `${cx + levelRadius * Math.cos(angle)},${cy + levelRadius * Math.sin(angle)}`
            })
            .join(' ')
          return (
            <polygon
              key={i}
              points={points}
              fill="none"
              className="stroke-gray-700"
              strokeWidth={1}
            />
          )
        })}

        {/* Axis lines */}
        {axes.map((_, i) => {
          const end = getAxisEnd(i)
          return (
            <line
              key={i}
              x1={cx}
              y1={cy}
              x2={end.x}
              y2={end.y}
              className="stroke-gray-600"
              strokeWidth={1}
            />
          )
        })}

        {/* Axis labels */}
        {axes.map((axis, i) => {
          const end = getAxisEnd(i)
          return (
            <text
              key={axis.key}
              x={end.labelX}
              y={end.labelY}
              textAnchor="middle"
              dominantBaseline="middle"
              className="fill-gray-400 text-xs"
            >
              {axis.label}
            </text>
          )
        })}

        {/* Data polygons */}
        {datasets.map((dataset, di) => {
          const color = dataset.color ?? getColor(di)
          const points = axes
            .map((axis, ai) => {
              const value = dataset.values[axis.key] ?? 0
              const pt = getPoint(ai, value, axis.max)
              return `${pt.x},${pt.y}`
            })
            .join(' ')

          return (
            <g key={di}>
              <polygon
                points={points}
                fill={color}
                fillOpacity={hoverDataset === null || hoverDataset === di ? 0.3 : 0.1}
                stroke={color}
                strokeWidth={2}
                strokeOpacity={hoverDataset === null || hoverDataset === di ? 1 : 0.3}
                className="transition-all cursor-pointer"
                onMouseEnter={() => setHoverDataset(di)}
                onMouseLeave={() => setHoverDataset(null)}
              />
              {/* Data points */}
              {axes.map((axis, ai) => {
                const value = dataset.values[axis.key] ?? 0
                const pt = getPoint(ai, value, axis.max)
                return (
                  <circle
                    key={ai}
                    cx={pt.x}
                    cy={pt.y}
                    r={hoverDataset === di ? 5 : 3}
                    fill={color}
                    opacity={hoverDataset === null || hoverDataset === di ? 1 : 0.3}
                    className="transition-all"
                  />
                )
              })}
            </g>
          )
        })}
      </svg>

      {/* Legend */}
      {datasets.length > 0 && (
        <div className="text-sm space-y-2">
          {datasets.map((dataset, di) => {
            const color = dataset.color ?? getColor(di)
            return (
              <div
                key={di}
                className={`flex items-center gap-2 cursor-pointer ${
                  hoverDataset === di ? 'font-semibold' : ''
                }`}
                onMouseEnter={() => setHoverDataset(di)}
                onMouseLeave={() => setHoverDataset(null)}
              >
                <span
                  className="w-3 h-3 rounded-sm"
                  style={{ background: color }}
                />
                <span className="text-gray-300">{dataset.name}</span>
              </div>
            )
          })}
          {hoverDataset !== null && (
            <div className="mt-3 pt-3 border-t border-gray-700 space-y-1">
              {axes.map((axis) => (
                <div key={axis.key} className="flex justify-between text-xs">
                  <span className="text-gray-400">{axis.label}</span>
                  <span className="text-gray-300">
                    {datasets[hoverDataset].values[axis.key] ?? 0} / {axis.max}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
