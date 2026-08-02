interface EmptyStateProps {
  icon?: string
  title: string
  description?: string
  hint?: string
}

export function EmptyState({ icon, title, description, hint }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center h-full">
      <div className="text-center max-w-md">
        {icon && <div className="text-5xl mb-4">{icon}</div>}
        <h2 className="text-xl font-semibold text-gray-100 mb-2">{title}</h2>
        {description && <p className="text-gray-400 text-sm mb-4">{description}</p>}
        {hint && (
          <div className="text-xs text-gray-500">
            <code className="bg-gray-800 px-2 py-1 rounded">{hint}</code>
          </div>
        )}
      </div>
    </div>
  )
}
