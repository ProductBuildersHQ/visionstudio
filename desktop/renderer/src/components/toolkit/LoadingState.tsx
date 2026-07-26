interface LoadingStateProps {
  message?: string
}

export function LoadingState({ message = 'Loading...' }: LoadingStateProps) {
  return (
    <div className="flex items-center justify-center h-full bg-va-bg">
      <div className="text-center">
        <div className="animate-spin w-8 h-8 border-2 border-va-accent border-t-transparent rounded-full mx-auto mb-4" />
        <p className="text-va-text-muted">{message}</p>
      </div>
    </div>
  )
}
