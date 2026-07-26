import type { ReactNode } from 'react'

interface ViewHeaderProps {
  title: string
  subtitle?: string
  isLoading?: boolean
  actions?: ReactNode
}

export function ViewHeader({ title, subtitle, isLoading, actions }: ViewHeaderProps) {
  return (
    <div className="flex items-center justify-between px-4 py-2 border-b border-va-border bg-va-sidebar">
      <div className="flex items-center gap-4">
        <h2 className="text-lg font-semibold text-va-text">{title}</h2>
        {subtitle && (
          <span className="text-sm text-va-text-muted">{subtitle}</span>
        )}
      </div>

      <div className="flex items-center gap-2">
        {isLoading && (
          <div className="animate-spin w-4 h-4 border-2 border-va-accent border-t-transparent rounded-full" />
        )}
        {actions}
      </div>
    </div>
  )
}

interface RefreshButtonProps {
  onClick: () => void
  disabled?: boolean
}

export function RefreshButton({ onClick, disabled }: RefreshButtonProps) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className="p-1.5 rounded hover:bg-va-panel text-va-text-muted hover:text-va-text disabled:opacity-50 transition-colors"
      title="Refresh"
    >
      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
        />
      </svg>
    </button>
  )
}
