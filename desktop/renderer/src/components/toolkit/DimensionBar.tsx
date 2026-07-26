interface DimensionBarProps {
  label: string
  value: number
  maxValue?: number
  color?: string
}

export function DimensionBar({ label, value, maxValue = 5, color }: DimensionBarProps) {
  const percent = Math.round((value / maxValue) * 100)

  const barColor = color ?? getBarColor(value, maxValue)

  return (
    <div className="flex items-center gap-3">
      <span className="text-xs text-va-text-muted w-32 truncate">{label}</span>
      <div className="flex-1 h-2 bg-va-bg rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${barColor}`}
          style={{ width: `${percent}%` }}
        />
      </div>
      <span className="text-xs text-va-text tabular-nums w-8 text-right">
        {value}/{maxValue}
      </span>
    </div>
  )
}

function getBarColor(value: number, maxValue: number): string {
  const ratio = value / maxValue
  if (ratio >= 0.8) return 'bg-va-success'
  if (ratio >= 0.6) return 'bg-blue-500'
  if (ratio >= 0.4) return 'bg-va-warning'
  return 'bg-va-error'
}
