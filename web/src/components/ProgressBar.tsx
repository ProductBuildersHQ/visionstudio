interface ProgressBarProps {
  progress: number
  /** Fraction (0-1) of the total that's cancelled, not just incomplete —
   * rendered as a red segment so it doesn't read the same as untouched work. */
  cancelledProgress?: number
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

export function ProgressBar({
  progress,
  cancelledProgress = 0,
  className = '',
  size = 'md',
}: ProgressBarProps) {
  const pct = Math.round(progress * 100)
  const cancelledPct = Math.min(Math.round(cancelledProgress * 100), 100 - pct)
  // Nothing left pending once completed + cancelled account for the whole
  // bar -- there's no more in-flight work, so the completed slice is done,
  // not merely "in progress" (blue). A still-open remainder keeps the usual
  // in-progress coloring.
  const resolved = pct + cancelledPct >= 100
  const bgColor = resolved ? 'bg-green-500' : pct >= 50 ? 'bg-blue-500' : 'bg-yellow-500'

  const heights = {
    sm: 'h-1',
    md: 'h-1.5',
    lg: 'h-2',
  }

  return (
    <div className={`${heights[size]} bg-gray-700 rounded-full overflow-hidden flex ${className}`}>
      <div className={`h-full ${bgColor} transition-all duration-300`} style={{ width: `${pct}%` }} />
      {cancelledPct > 0 && (
        <div
          className="h-full bg-red-500 transition-all duration-300"
          style={{ width: `${cancelledPct}%` }}
          title={`${cancelledPct}% cancelled`}
        />
      )}
    </div>
  )
}
