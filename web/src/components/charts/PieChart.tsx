import { useState } from 'react'
import { getColor } from './colors'

export interface PieSlice {
  name: string
  value: number
}

interface PieChartProps {
  data: PieSlice[]
  size?: number
  innerRadius?: number
  showLegend?: boolean
}

export function PieChart({
  data,
  size = 180,
  innerRadius = 40,
  showLegend = true,
}: PieChartProps) {
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)

  if (data.length === 0) {
    return (
      <div className="text-sm text-gray-400 h-40 flex items-center justify-center">
        No data
      </div>
    )
  }

  const total = data.reduce((sum, d) => sum + d.value, 0)
  if (total === 0) {
    return (
      <div className="text-sm text-gray-400 h-40 flex items-center justify-center">
        No data
      </div>
    )
  }

  const cx = size / 2
  const cy = size / 2
  const outerR = size / 2 - 10
  const innerR = innerRadius

  let cumAngle = -Math.PI / 2

  const slices = data.map((d, i) => {
    const pct = d.value / total
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

    return { ...d, pct, path, color: getColor(i) }
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
          <text
            x={cx}
            y={cy}
            textAnchor="middle"
            dominantBaseline="middle"
            className="fill-white text-sm font-semibold"
          >
            {(slices[hoverIdx].pct * 100).toFixed(1)}%
          </text>
        )}
      </svg>
      {showLegend && (
        <div className="flex-1 text-xs space-y-1">
          {slices.map((s, i) => (
            <div
              key={i}
              className={`flex items-center gap-2 ${hoverIdx === i ? 'font-semibold' : ''}`}
              onMouseEnter={() => setHoverIdx(i)}
              onMouseLeave={() => setHoverIdx(null)}
            >
              <span
                className="w-2.5 h-2.5 rounded-sm inline-block flex-shrink-0"
                style={{ background: s.color }}
              />
              <span className="text-gray-300 truncate">{s.name}</span>
              <span className="text-gray-500 ml-auto tabular-nums">
                {s.value.toLocaleString()}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
