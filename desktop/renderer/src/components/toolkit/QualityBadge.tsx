type QualityRating = 'EXCELLENT' | 'GOOD' | 'NEEDS_IMPROVEMENT' | 'POOR'

const QUALITY_STYLES: Record<QualityRating, { bg: string; text: string; label: string }> = {
  EXCELLENT: { bg: 'bg-va-success/20', text: 'text-va-success', label: 'Excellent' },
  GOOD: { bg: 'bg-blue-500/20', text: 'text-blue-400', label: 'Good' },
  NEEDS_IMPROVEMENT: { bg: 'bg-va-warning/20', text: 'text-va-warning', label: 'Needs Improvement' },
  POOR: { bg: 'bg-va-error/20', text: 'text-va-error', label: 'Poor' },
}

interface QualityBadgeProps {
  rating: string
  score?: number
  size?: 'sm' | 'md'
}

export function QualityBadge({ rating, score, size = 'sm' }: QualityBadgeProps) {
  const style = QUALITY_STYLES[rating as QualityRating] ?? QUALITY_STYLES.POOR

  const sizeClasses = size === 'sm'
    ? 'text-xs px-2 py-0.5'
    : 'text-sm px-3 py-1'

  return (
    <span className={`${sizeClasses} rounded-full font-medium ${style.bg} ${style.text}`}>
      {style.label}
      {score !== undefined && <span className="ml-1 opacity-75">({score})</span>}
    </span>
  )
}
