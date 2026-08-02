interface StatusBadgeProps {
  status: string
  size?: 'sm' | 'md'
}

const statusColors: Record<string, string> = {
  completed: 'bg-green-500/20 text-green-400 border-green-500/30',
  in_progress: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  executing: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  ready: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  planned: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  proposed: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
  cancelled: 'bg-red-500/20 text-red-400 border-red-500/30',
}

export function StatusBadge({ status, size = 'md' }: StatusBadgeProps) {
  const normalizedStatus = status.toLowerCase().replace(/\s+/g, '_')
  const colorClass = statusColors[normalizedStatus] ?? statusColors.proposed
  const sizeClass = size === 'sm' ? 'text-xs px-1.5 py-0.5' : 'text-xs px-2 py-1'

  return (
    <span className={`rounded border font-medium ${colorClass} ${sizeClass}`}>
      {status.replace(/_/g, ' ')}
    </span>
  )
}
