interface SummaryCardProps {
  label: string
  value: string | number
  color?: string
}

export function SummaryCard({ label, value, color = 'text-va-text' }: SummaryCardProps) {
  return (
    <div className="bg-va-panel rounded-lg p-4 border border-va-border">
      <div className="text-xs text-va-text-muted uppercase tracking-wide mb-1">{label}</div>
      <div className={`text-2xl font-bold ${color}`}>{value}</div>
    </div>
  )
}
