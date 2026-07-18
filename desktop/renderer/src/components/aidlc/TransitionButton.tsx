import { useState } from 'react'
import { api, AIDLCPhase, AIDLCPhaseTransitionResult } from '../../services/api'

interface TransitionButtonProps {
  projectName: string
  currentPhase: AIDLCPhase
  targetPhase: AIDLCPhase
  canAdvance: boolean
  onTransitionComplete?: (result: AIDLCPhaseTransitionResult) => void
}

const phaseLabels: Record<AIDLCPhase, string> = {
  inception: 'Inception',
  construction: 'Construction',
  operations: 'Operations',
}

const phaseOrder: AIDLCPhase[] = ['inception', 'construction', 'operations']

export function TransitionButton({
  projectName,
  currentPhase,
  targetPhase,
  canAdvance,
  onTransitionComplete,
}: TransitionButtonProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [result, setResult] = useState<AIDLCPhaseTransitionResult | null>(null)
  const [notes, setNotes] = useState('')

  const currentIndex = phaseOrder.indexOf(currentPhase)
  const targetIndex = phaseOrder.indexOf(targetPhase)
  const isForward = targetIndex > currentIndex

  const handleTransition = async () => {
    setIsLoading(true)
    setResult(null)

    try {
      const transitionResult = await api.transitionAIDLCPhase(projectName, targetPhase, {
        notes: notes || undefined,
      })
      setResult(transitionResult)
      if (transitionResult.success) {
        onTransitionComplete?.(transitionResult)
        setShowConfirm(false)
      }
    } catch (err) {
      setResult({
        success: false,
        from_phase: currentPhase,
        to_phase: targetPhase,
        blocking_issues: [err instanceof Error ? err.message : 'Transition failed'],
      })
    } finally {
      setIsLoading(false)
    }
  }

  if (!canAdvance && isForward) {
    return (
      <button
        disabled
        className="px-4 py-2 bg-gray-700 text-gray-400 rounded-lg cursor-not-allowed flex items-center gap-2"
      >
        <span>Complete phase requirements first</span>
      </button>
    )
  }

  if (showConfirm) {
    return (
      <div className="p-4 bg-va-panel rounded-lg border border-va-border space-y-4">
        <h3 className="font-semibold text-va-text">
          Transition to {phaseLabels[targetPhase]}?
        </h3>

        <p className="text-sm text-va-text-muted">
          You are about to transition from <strong>{phaseLabels[currentPhase]}</strong> to{' '}
          <strong>{phaseLabels[targetPhase]}</strong>.
        </p>

        <div>
          <label className="block text-sm text-va-text-muted mb-1">Notes (optional)</label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Reason for transition..."
            className="w-full px-3 py-2 bg-va-bg border border-va-border rounded text-sm"
            rows={2}
          />
        </div>

        {result && !result.success && (
          <div className="p-3 bg-red-900/20 border border-red-500 rounded">
            <p className="text-red-400 font-semibold mb-1">Transition Blocked</p>
            {result.blocking_docs && result.blocking_docs.length > 0 && (
              <div className="text-sm text-red-400/80">
                <p className="mb-1">Incomplete documents:</p>
                <ul className="list-disc list-inside">
                  {result.blocking_docs.map((doc) => (
                    <li key={doc}>{doc.replace(/_/g, ' ')}</li>
                  ))}
                </ul>
              </div>
            )}
            {result.blocking_issues && result.blocking_issues.length > 0 && (
              <div className="text-sm text-red-400/80 mt-2">
                <ul className="list-disc list-inside">
                  {result.blocking_issues.map((issue, idx) => (
                    <li key={idx}>{issue}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        )}

        <div className="flex gap-2">
          <button
            onClick={() => {
              setShowConfirm(false)
              setResult(null)
              setNotes('')
            }}
            className="px-4 py-2 bg-va-border hover:bg-va-border/80 rounded text-sm"
          >
            Cancel
          </button>
          <button
            onClick={handleTransition}
            disabled={isLoading}
            className={`px-4 py-2 rounded text-sm flex items-center gap-2 ${
              isForward
                ? 'bg-green-600 hover:bg-green-700 text-white'
                : 'bg-yellow-600 hover:bg-yellow-700 text-white'
            }`}
          >
            {isLoading && (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
            )}
            {isLoading ? 'Transitioning...' : 'Confirm Transition'}
          </button>
        </div>
      </div>
    )
  }

  return (
    <button
      onClick={() => setShowConfirm(true)}
      className={`px-4 py-2 rounded-lg flex items-center gap-2 ${
        isForward
          ? 'bg-green-600 hover:bg-green-700 text-white'
          : 'bg-yellow-600 hover:bg-yellow-700 text-white'
      }`}
    >
      <span>
        {isForward ? '→' : '←'} {phaseLabels[targetPhase]}
      </span>
    </button>
  )
}

export default TransitionButton
