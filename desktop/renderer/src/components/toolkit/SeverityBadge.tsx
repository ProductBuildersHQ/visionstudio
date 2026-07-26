const SEVERITY_STYLES: Record<string, { badge: string; border: string; dot: string }> = {
  critical: { badge: 'bg-red-500 text-white', border: 'border-l-red-500', dot: 'bg-red-500' },
  high: { badge: 'bg-orange-500 text-white', border: 'border-l-orange-500', dot: 'bg-orange-500' },
  medium: { badge: 'bg-yellow-500 text-black', border: 'border-l-yellow-500', dot: 'bg-yellow-500' },
  low: { badge: 'bg-blue-500 text-white', border: 'border-l-blue-500', dot: 'bg-blue-500' },
  info: { badge: 'bg-gray-500 text-white', border: 'border-l-gray-500', dot: 'bg-gray-500' },
}

export function getSeverityStyle(severity: string | undefined) {
  return SEVERITY_STYLES[severity ?? 'info'] ?? SEVERITY_STYLES.info
}

interface SeverityBadgeProps {
  severity: string
  size?: 'sm' | 'md'
}

export function SeverityBadge({ severity, size = 'sm' }: SeverityBadgeProps) {
  const style = getSeverityStyle(severity)
  const sizeClasses = size === 'sm'
    ? 'text-[10px] px-1.5 py-0.5'
    : 'text-xs px-2 py-1'

  return (
    <span className={`font-bold rounded shrink-0 ${sizeClasses} ${style.badge}`}>
      {severity.toUpperCase()}
    </span>
  )
}

interface SeverityDotProps {
  severity: string
  count?: number
  label?: string
}

export function SeverityDot({ severity, count, label }: SeverityDotProps) {
  const style = getSeverityStyle(severity)
  const isEmpty = count !== undefined && count === 0

  return (
    <div className="flex items-center gap-2">
      <span className={`w-3 h-3 rounded-full ${style.dot} ${isEmpty ? 'opacity-30' : ''}`} />
      {label && (
        <span className={`text-sm ${isEmpty ? 'text-va-text-muted' : 'text-va-text'}`}>
          {label}: <span className="font-semibold">{count ?? ''}</span>
        </span>
      )}
    </div>
  )
}
