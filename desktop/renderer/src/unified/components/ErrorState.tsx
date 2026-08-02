interface ErrorStateProps {
  message: string
  hint?: string
  onRetry?: () => void
}

export function ErrorState({ message, hint, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center h-full gap-4">
      <div className="text-center max-w-md">
        <div className="text-red-400 mb-4">{message}</div>
        {hint && <p className="text-gray-400 text-sm mb-4">{hint}</p>}
        {onRetry && (
          <button
            onClick={onRetry}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          >
            Retry
          </button>
        )}
      </div>
    </div>
  )
}
