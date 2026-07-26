interface EmptyStateProps {
  icon?: string
  title: string
  description?: string
  hint?: string
}

export function EmptyState({ icon, title, description, hint }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center h-full bg-va-bg">
      <div className="text-center max-w-md">
        {icon && <div className="text-5xl mb-4">{icon}</div>}
        <h2 className="text-xl font-semibold text-va-text mb-2">{title}</h2>
        {description && (
          <p className="text-va-text-muted text-sm mb-4">{description}</p>
        )}
        {hint && (
          <div className="text-xs text-va-text-muted">
            <code>{hint}</code>
          </div>
        )}
      </div>
    </div>
  )
}
