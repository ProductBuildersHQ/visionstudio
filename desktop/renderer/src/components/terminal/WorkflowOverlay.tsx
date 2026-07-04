import { useState, useEffect, useCallback } from 'react'
import { api, WorkflowStatus } from '../../services/api'

interface WorkflowOverlayProps {
  projectName: string | null
  onClose?: () => void
}

const POLL_INTERVAL = 5000 // 5 seconds

export function WorkflowOverlay({ projectName, onClose }: WorkflowOverlayProps) {
  const [status, setStatus] = useState<WorkflowStatus | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isVisible, setIsVisible] = useState(true)

  const fetchStatus = useCallback(async () => {
    if (!projectName) {
      setStatus(null)
      setIsLoading(false)
      return
    }

    try {
      const data = await api.getWorkflowStatus(projectName)
      setStatus(data)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch status')
    } finally {
      setIsLoading(false)
    }
  }, [projectName])

  // Initial fetch and polling
  useEffect(() => {
    fetchStatus()

    const interval = setInterval(fetchStatus, POLL_INTERVAL)
    return () => clearInterval(interval)
  }, [fetchStatus])

  const handleClose = useCallback(() => {
    setIsVisible(false)
    onClose?.()
  }, [onClose])

  if (!isVisible || !projectName) {
    return null
  }

  return (
    <div className="absolute top-2 right-2 z-10">
      <div className="bg-va-panel border border-va-border rounded-lg shadow-lg p-3 min-w-64">
        {/* Header */}
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm font-medium text-va-text">Workflow Status</span>
          <button
            onClick={handleClose}
            className="text-va-text-muted hover:text-va-text"
            title="Close"
          >
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M1 1l12 12M13 1L1 13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          </button>
        </div>

        {isLoading && !status ? (
          <div className="text-sm text-va-text-muted">Loading...</div>
        ) : error ? (
          <div className="text-sm text-va-error">{error}</div>
        ) : status ? (
          <div className="space-y-3">
            {/* Current Phase */}
            <div className="flex items-center space-x-2">
              <span className="text-xs text-va-text-muted uppercase">Phase:</span>
              <span className="text-sm text-va-text capitalize">{status.currentPhase}</span>
            </div>

            {/* Progress Bar */}
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs text-va-text-muted">Progress</span>
                <span className="text-xs text-va-text">{Math.round(status.progress * 100)}%</span>
              </div>
              <div className="h-1.5 bg-va-sidebar rounded-full overflow-hidden">
                <div
                  className="h-full bg-va-accent transition-all duration-300"
                  style={{ width: `${status.progress * 100}%` }}
                />
              </div>
            </div>

            {/* Spec Statuses */}
            <div>
              <div className="text-xs text-va-text-muted uppercase mb-1">Specs</div>
              <div className="flex flex-wrap gap-1">
                {Object.entries(status.specStatuses).map(([specType, specStatus]) => (
                  <SpecChip key={specType} type={specType} status={specStatus} />
                ))}
              </div>
            </div>

            {/* Blocked By */}
            {status.blockedBy && status.blockedBy.length > 0 && (
              <div>
                <div className="text-xs text-va-warning uppercase mb-1">Blocked by</div>
                <div className="text-sm text-va-text-muted">
                  {status.blockedBy.join(', ')}
                </div>
              </div>
            )}

            {/* Last Updated */}
            <div className="text-xs text-va-text-muted">
              Updated: {new Date(status.lastUpdated).toLocaleTimeString()}
            </div>
          </div>
        ) : (
          <div className="text-sm text-va-text-muted">No status available</div>
        )}
      </div>
    </div>
  )
}

interface SpecChipProps {
  type: string
  status: string
}

function SpecChip({ type, status }: SpecChipProps) {
  const statusColors: Record<string, string> = {
    not_started: 'bg-va-text-muted/20 text-va-text-muted',
    draft: 'bg-va-warning/20 text-va-warning',
    evaluated: 'bg-va-accent/20 text-va-accent',
    approved: 'bg-va-success/20 text-va-success',
  }

  const statusIcons: Record<string, string> = {
    not_started: '',
    draft: '',
    evaluated: '',
    approved: '',
  }

  const colorClass = statusColors[status] || statusColors.not_started
  const icon = statusIcons[status] || ''

  return (
    <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-xs ${colorClass}`}>
      {icon && <span className="mr-0.5">{icon}</span>}
      {type}
    </span>
  )
}
