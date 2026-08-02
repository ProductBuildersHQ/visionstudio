interface ProgressBarProps {
  progress: number
  className?: string
}

export function ProgressBar({ progress, className = '' }: ProgressBarProps) {
  const pct = Math.round(progress * 100)
  const bgColor =
    pct >= 100 ? 'bg-green-500' : pct >= 50 ? 'bg-blue-500' : 'bg-yellow-500'

  return (
    <div className={`h-1.5 bg-gray-700 rounded-full overflow-hidden ${className}`}>
      <div
        className={`h-full ${bgColor} transition-all duration-300`}
        style={{ width: `${pct}%` }}
      />
    </div>
  )
}
