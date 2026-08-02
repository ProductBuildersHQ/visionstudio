interface ProgressBarProps {
  progress: number
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

export function ProgressBar({ progress, className = '', size = 'md' }: ProgressBarProps) {
  const pct = Math.round(progress * 100)
  const bgColor =
    pct >= 100 ? 'bg-green-500' : pct >= 50 ? 'bg-blue-500' : 'bg-yellow-500'

  const heights = {
    sm: 'h-1',
    md: 'h-1.5',
    lg: 'h-2',
  }

  return (
    <div className={`${heights[size]} bg-gray-700 rounded-full overflow-hidden ${className}`}>
      <div
        className={`h-full ${bgColor} transition-all duration-300`}
        style={{ width: `${pct}%` }}
      />
    </div>
  )
}
