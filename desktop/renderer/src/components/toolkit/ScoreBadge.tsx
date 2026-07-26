const SCORE_COLORS: Record<number, string> = {
  5: 'bg-blue-500',
  4: 'bg-green-500',
  3: 'bg-yellow-500',
  2: 'bg-orange-500',
  1: 'bg-red-500',
}

const SCORE_LABELS: Record<number, string> = {
  5: 'Excellent',
  4: 'Good',
  3: 'Acceptable',
  2: 'Major Revisions',
  1: 'Unacceptable',
}

interface ScoreBadgeProps {
  score: number
  maxScore?: number
  showLabel?: boolean
  size?: 'sm' | 'md' | 'lg'
}

export function ScoreBadge({ score, maxScore = 5, showLabel = false, size = 'sm' }: ScoreBadgeProps) {
  const color = SCORE_COLORS[Math.min(Math.max(score, 1), 5)] ?? 'bg-gray-500'
  const label = SCORE_LABELS[Math.min(Math.max(score, 1), 5)] ?? ''

  const sizeClasses = {
    sm: 'text-xs px-1.5 py-0.5',
    md: 'text-sm px-2 py-1',
    lg: 'text-base px-3 py-1.5',
  }

  return (
    <span className={`${sizeClasses[size]} ${color} text-white font-semibold rounded`}>
      {score}/{maxScore}
      {showLabel && label && <span className="ml-1 font-normal opacity-80">{label}</span>}
    </span>
  )
}

interface ScoreDotProps {
  score: number
  className?: string
}

export function ScoreDot({ score, className = '' }: ScoreDotProps) {
  const color = SCORE_COLORS[Math.min(Math.max(score, 1), 5)] ?? 'bg-gray-500'
  return <span className={`w-3 h-3 rounded-full inline-block ${color} ${className}`} />
}
