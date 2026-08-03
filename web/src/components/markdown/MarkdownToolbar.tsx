interface ToolbarAction {
  id: string
  label: string
  icon?: string
  onClick: () => void
  disabled?: boolean
}

interface MarkdownToolbarProps {
  actions: ToolbarAction[]
  className?: string
}

export function MarkdownToolbar({ actions, className = '' }: MarkdownToolbarProps) {
  return (
    <div className={`mde-toolbar ${className}`}>
      {actions.map((action) => (
        <button
          key={action.id}
          onClick={action.onClick}
          disabled={action.disabled}
          className="mde-toolbar-btn"
          title={action.label}
        >
          {action.icon ?? action.label}
        </button>
      ))}
    </div>
  )
}
