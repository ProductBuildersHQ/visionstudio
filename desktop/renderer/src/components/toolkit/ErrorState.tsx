interface ErrorStateProps {
  message: string
  hint?: string
  onRetry?: () => void
}

export function ErrorState({ message, hint, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center h-full bg-va-bg gap-4">
      <div className="text-center max-w-md">
        <div className="text-va-error mb-4">{message}</div>
        {hint && (
          <p className="text-va-text-muted text-sm mb-4">{hint}</p>
        )}
        {onRetry && (
          <button
            onClick={onRetry}
            className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80 transition-colors"
          >
            Retry
          </button>
        )}
      </div>
    </div>
  )
}
