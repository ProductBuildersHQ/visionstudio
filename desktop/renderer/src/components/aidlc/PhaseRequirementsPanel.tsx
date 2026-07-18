import { useState, useEffect, useCallback } from 'react'
import { api, AIDLCPhase, AIDLCPhaseRequirements } from '../../services/api'

interface PhaseRequirementsPanelProps {
  projectName: string
  onDocumentClick?: (docType: string) => void
}

function PhaseCard({
  requirements,
  isCurrent,
  onDocumentClick,
}: {
  requirements: AIDLCPhaseRequirements
  isCurrent: boolean
  onDocumentClick?: (docType: string) => void
}) {
  const phaseColors = {
    inception: 'border-blue-500 bg-blue-900/20',
    construction: 'border-yellow-500 bg-yellow-900/20',
    operations: 'border-green-500 bg-green-900/20',
  }

  const phaseIcons = {
    inception: '📋',
    construction: '🔨',
    operations: '🚀',
  }

  const progressColor =
    requirements.progress_percent >= 100
      ? 'bg-green-500'
      : requirements.progress_percent >= 50
        ? 'bg-blue-500'
        : 'bg-gray-500'

  return (
    <div
      className={`p-4 rounded-lg border ${
        phaseColors[requirements.phase] || 'border-gray-500 bg-gray-900/20'
      } ${isCurrent ? 'ring-2 ring-va-accent' : ''}`}
    >
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-xl">{phaseIcons[requirements.phase] || '📄'}</span>
          <h3 className="text-lg font-semibold capitalize">{requirements.phase}</h3>
          {isCurrent && (
            <span className="px-2 py-0.5 text-xs bg-va-accent/20 text-va-accent rounded">
              Current
            </span>
          )}
        </div>
        <span className="text-sm text-va-text-muted">
          {requirements.completed_docs.length}/{requirements.required_docs.length}
        </span>
      </div>

      {/* Progress bar */}
      <div className="h-2 bg-va-border rounded-full overflow-hidden mb-3">
        <div
          className={`h-full ${progressColor} transition-all`}
          style={{ width: `${requirements.progress_percent}%` }}
        />
      </div>

      {/* Documents */}
      <div className="space-y-2">
        {requirements.required_docs.map((docType) => {
          const isCompleted = requirements.completed_docs.includes(docType)
          const isPending = requirements.pending_docs.includes(docType)

          return (
            <button
              key={docType}
              onClick={() => onDocumentClick?.(docType)}
              className={`w-full flex items-center justify-between px-3 py-2 rounded text-sm ${
                isCompleted
                  ? 'bg-green-900/30 text-green-400'
                  : isPending
                    ? 'bg-gray-800 text-va-text-muted hover:bg-gray-700'
                    : 'bg-blue-900/30 text-blue-400'
              }`}
            >
              <span className="flex items-center gap-2">
                {isCompleted ? '✓' : isPending ? '○' : '◉'}
                <span className="capitalize">{docType.replace(/_/g, ' ')}</span>
              </span>
              <span className="text-xs opacity-75">
                {isCompleted ? 'Complete' : isPending ? 'Pending' : 'In Progress'}
              </span>
            </button>
          )
        })}
      </div>

      {/* Advance status */}
      <div className="mt-3 pt-3 border-t border-va-border">
        {requirements.can_advance ? (
          <div className="flex items-center gap-2 text-green-400 text-sm">
            <span>✓</span>
            <span>Ready to advance</span>
          </div>
        ) : (
          <div className="flex items-center gap-2 text-va-text-muted text-sm">
            <span>○</span>
            <span>
              Complete {requirements.pending_docs.length} more document
              {requirements.pending_docs.length !== 1 ? 's' : ''} to advance
            </span>
          </div>
        )}
      </div>
    </div>
  )
}

export function PhaseRequirementsPanel({
  projectName,
  onDocumentClick,
}: PhaseRequirementsPanelProps) {
  const [currentPhase, setCurrentPhase] = useState<AIDLCPhase>('inception')
  const [requirements, setRequirements] = useState<AIDLCPhaseRequirements[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const loadRequirements = useCallback(async () => {
    if (!projectName) return

    setIsLoading(true)
    setError(null)

    try {
      const data = await api.getAIDLCPhaseRequirements(projectName)
      setCurrentPhase(data.currentPhase)
      setRequirements(data.requirements)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load requirements')
    } finally {
      setIsLoading(false)
    }
  }, [projectName])

  useEffect(() => {
    loadRequirements()
  }, [loadRequirements])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-va-accent"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 bg-red-900/20 border border-red-500 rounded-lg">
        <p className="text-red-400">{error}</p>
        <button
          onClick={loadRequirements}
          className="mt-2 px-3 py-1 bg-red-900/50 hover:bg-red-900/70 rounded text-sm"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-va-text">Phase Requirements</h2>
        <button
          onClick={loadRequirements}
          className="px-3 py-1 bg-va-panel hover:bg-va-border rounded text-sm"
        >
          Refresh
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {requirements.map((req) => (
          <PhaseCard
            key={req.phase}
            requirements={req}
            isCurrent={req.phase === currentPhase}
            onDocumentClick={onDocumentClick}
          />
        ))}
      </div>
    </div>
  )
}

export default PhaseRequirementsPanel
