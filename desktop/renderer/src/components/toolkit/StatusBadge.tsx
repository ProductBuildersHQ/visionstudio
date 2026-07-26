const STATUS_STYLES: Record<string, string> = {
  completed: 'bg-va-success/20 text-va-success',
  done: 'bg-va-success/20 text-va-success',
  pass: 'bg-va-success/20 text-va-success',
  active: 'bg-va-warning/20 text-va-warning',
  in_progress: 'bg-va-warning/20 text-va-warning',
  'in-progress': 'bg-va-warning/20 text-va-warning',
  conditional: 'bg-va-warning/20 text-va-warning',
  partial: 'bg-va-warning/20 text-va-warning',
  blocked: 'bg-va-error/20 text-va-error',
  fail: 'bg-va-error/20 text-va-error',
  pending: 'bg-va-text-muted/20 text-va-text-muted',
  not_started: 'bg-va-text-muted/20 text-va-text-muted',
  draft: 'bg-va-text-muted/20 text-va-text-muted',
}

interface StatusBadgeProps {
  status: string
  className?: string
}

export function StatusBadge({ status, className = '' }: StatusBadgeProps) {
  const style = STATUS_STYLES[status.toLowerCase()] ?? STATUS_STYLES.pending

  return (
    <span className={`text-xs px-1.5 py-0.5 rounded ${style} ${className}`}>
      {status}
    </span>
  )
}
