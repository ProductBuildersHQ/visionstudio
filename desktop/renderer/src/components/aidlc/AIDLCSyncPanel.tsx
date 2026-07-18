import { useState, useEffect, useCallback } from 'react'
import {
  api,
  AIDLCSyncDiff,
  AIDLCSyncResult,
  AIDLCSyncAction,
  AIDLCSyncConflict,
} from '../../services/api'

interface AIDLCSyncPanelProps {
  projectName: string
}

function ActionCard({ action }: { action: AIDLCSyncAction }) {
  const directionColors = {
    to_aidlc: 'border-blue-500 bg-blue-900/20',
    from_aidlc: 'border-purple-500 bg-purple-900/20',
  }
  const actionColors = {
    create: 'text-green-400',
    update: 'text-yellow-400',
    delete: 'text-red-400',
  }

  const colors =
    directionColors[action.direction as keyof typeof directionColors] ||
    'border-gray-500 bg-gray-900/20'

  return (
    <div className={`p-3 rounded-lg border ${colors}`}>
      <div className="flex items-center justify-between mb-1">
        <span className="font-medium text-va-text">{action.doc_type}</span>
        <span
          className={`text-xs uppercase ${actionColors[action.action as keyof typeof actionColors] || 'text-gray-400'}`}
        >
          {action.action}
        </span>
      </div>
      <div className="text-xs text-va-text-muted">
        <div className="flex items-center gap-1">
          <span>From:</span>
          <span className="font-mono truncate">{action.source_path}</span>
        </div>
        <div className="flex items-center gap-1">
          <span>To:</span>
          <span className="font-mono truncate">{action.dest_path}</span>
        </div>
      </div>
      {action.reason && <p className="text-xs text-va-accent mt-1">{action.reason}</p>}
    </div>
  )
}

function ConflictCard({ conflict }: { conflict: AIDLCSyncConflict }) {
  return (
    <div className="p-3 rounded-lg border border-red-500 bg-red-900/20">
      <div className="flex items-center justify-between mb-1">
        <span className="font-medium text-red-400">{conflict.doc_type}</span>
        <span className="text-xs text-red-400 uppercase">Conflict</span>
      </div>
      <div className="text-xs text-va-text-muted">
        <div>VisionSpec: {conflict.visionspec_path}</div>
        <div>AIDLC: {conflict.aidlc_path}</div>
      </div>
      <p className="text-xs text-red-300 mt-1">{conflict.reason}</p>
    </div>
  )
}

function SyncResultCard({ result }: { result: AIDLCSyncResult }) {
  const hasErrors = result.errors && result.errors.length > 0

  return (
    <div
      className={`p-4 rounded-lg border ${hasErrors ? 'border-red-500 bg-red-900/20' : 'border-green-500 bg-green-900/20'}`}
    >
      <div className="flex items-center justify-between mb-2">
        <span className={`font-medium ${hasErrors ? 'text-red-400' : 'text-green-400'}`}>
          Sync {hasErrors ? 'Completed with Errors' : 'Successful'}
        </span>
        <span className="text-xs text-va-text-muted">
          {new Date(result.completed_at).toLocaleTimeString()}
        </span>
      </div>

      <div className="grid grid-cols-3 gap-4 text-sm">
        <div>
          <span className="text-green-400 font-medium">{result.created?.length || 0}</span>
          <span className="text-va-text-muted ml-1">Created</span>
        </div>
        <div>
          <span className="text-yellow-400 font-medium">{result.updated?.length || 0}</span>
          <span className="text-va-text-muted ml-1">Updated</span>
        </div>
        <div>
          <span className="text-gray-400 font-medium">{result.skipped?.length || 0}</span>
          <span className="text-va-text-muted ml-1">Skipped</span>
        </div>
      </div>

      {hasErrors && (
        <div className="mt-3 text-xs text-red-300">
          <div className="font-medium mb-1">Errors:</div>
          <ul className="list-disc list-inside">
            {result.errors?.map((err, idx) => (
              <li key={idx}>{err}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

export function AIDLCSyncPanel({ projectName }: AIDLCSyncPanelProps) {
  const [diff, setDiff] = useState<AIDLCSyncDiff | null>(null)
  const [result, setResult] = useState<AIDLCSyncResult | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isSyncing, setIsSyncing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [direction, setDirection] = useState<string>('bidirectional')

  const loadDiff = useCallback(async () => {
    if (!projectName) return

    setIsLoading(true)
    setError(null)

    try {
      const data = await api.getAIDLCSyncDiff(projectName)
      setDiff(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load sync diff')
    } finally {
      setIsLoading(false)
    }
  }, [projectName])

  const handleSync = async (dryRun = false) => {
    if (!projectName) return

    setIsSyncing(true)
    setError(null)
    setResult(null)

    try {
      const res = await api.syncAIDLC(projectName, { direction, dryRun })
      setResult(res)
      if (!dryRun) {
        // Reload diff after sync
        await loadDiff()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sync failed')
    } finally {
      setIsSyncing(false)
    }
  }

  useEffect(() => {
    loadDiff()
  }, [loadDiff])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-va-accent"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 bg-red-900/20 border border-red-500 rounded-lg">
        <p className="text-red-400">{error}</p>
        <button
          onClick={loadDiff}
          className="mt-2 px-3 py-1 bg-red-900/50 hover:bg-red-900/70 rounded text-sm"
        >
          Retry
        </button>
      </div>
    )
  }

  const hasChanges = diff && (diff.actions.length > 0 || (diff.conflicts?.length || 0) > 0)

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-va-border">
        <div>
          <h2 className="text-lg font-semibold text-va-text">AIDLC Sync</h2>
          <p className="text-sm text-va-text-muted">
            Synchronize documents between VisionSpec and AIDLC
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={direction}
            onChange={(e) => setDirection(e.target.value)}
            className="px-3 py-1 bg-va-panel border border-va-border rounded text-sm"
          >
            <option value="bidirectional">Bidirectional</option>
            <option value="to_aidlc">To AIDLC</option>
            <option value="from_aidlc">From AIDLC</option>
          </select>
          <button
            onClick={() => handleSync(true)}
            disabled={isSyncing}
            className="px-3 py-1 bg-va-panel hover:bg-va-border rounded text-sm disabled:opacity-50"
          >
            {isSyncing ? 'Checking...' : 'Dry Run'}
          </button>
          <button
            onClick={() => handleSync(false)}
            disabled={isSyncing || !hasChanges}
            className="px-3 py-1 bg-va-accent hover:brightness-110 rounded text-sm disabled:opacity-50"
          >
            {isSyncing ? 'Syncing...' : 'Sync Now'}
          </button>
          <button
            onClick={loadDiff}
            className="px-3 py-1 bg-va-panel hover:bg-va-border rounded text-sm"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-4 space-y-4">
        {/* Sync Result */}
        {result && <SyncResultCard result={result} />}

        {/* Directories */}
        {diff && (
          <div className="flex gap-4 text-xs text-va-text-muted">
            <div>
              <span className="font-medium">VisionSpec:</span>
              <span className="font-mono ml-1">{diff.visionspec_dir}</span>
            </div>
            <div>
              <span className="font-medium">AIDLC:</span>
              <span className="font-mono ml-1">{diff.aidlc_docs_dir}</span>
            </div>
          </div>
        )}

        {/* Status */}
        {!hasChanges ? (
          <div className="flex flex-col items-center justify-center h-32 text-va-text-muted bg-va-panel rounded-lg">
            <span className="text-green-400 text-2xl mb-2">&#10003;</span>
            <p>Directories are in sync</p>
            <p className="text-sm mt-1">No changes detected</p>
          </div>
        ) : (
          <>
            {/* Conflicts */}
            {diff?.conflicts && diff.conflicts.length > 0 && (
              <div>
                <h3 className="text-sm font-semibold text-red-400 mb-2">
                  Conflicts ({diff.conflicts.length})
                </h3>
                <div className="space-y-2">
                  {diff.conflicts.map((conflict, idx) => (
                    <ConflictCard key={idx} conflict={conflict} />
                  ))}
                </div>
              </div>
            )}

            {/* Actions */}
            {diff?.actions && diff.actions.length > 0 && (
              <div>
                <h3 className="text-sm font-semibold text-va-text mb-2">
                  Pending Actions ({diff.actions.length})
                </h3>
                <div className="space-y-2">
                  {diff.actions.map((action, idx) => (
                    <ActionCard key={idx} action={action} />
                  ))}
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

export default AIDLCSyncPanel
