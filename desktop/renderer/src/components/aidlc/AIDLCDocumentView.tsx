import { useState, useEffect, useCallback } from 'react'
import { api, AIDLCDocument, AIDLCQualityScore, AIDLCIssue } from '../../services/api'

interface AIDLCDocumentViewProps {
  projectName: string
  docType: string
  onBack?: () => void
}

function QualityBadge({ score }: { score: AIDLCQualityScore }) {
  const colors = {
    EXCELLENT: 'bg-green-900/50 text-green-400 border-green-500',
    GOOD: 'bg-blue-900/50 text-blue-400 border-blue-500',
    NEEDS_IMPROVEMENT: 'bg-yellow-900/50 text-yellow-400 border-yellow-500',
    POOR: 'bg-red-900/50 text-red-400 border-red-500',
  }

  return (
    <div
      className={`px-3 py-1 rounded-lg border ${colors[score.rating] || colors.POOR}`}
    >
      <span className="font-medium">{score.rating}</span>
      <span className="ml-2 opacity-75">{(score.score * 100).toFixed(0)}%</span>
    </div>
  )
}

function IssueCard({ issue }: { issue: AIDLCIssue }) {
  const severityColors = {
    critical: 'bg-red-900/30 border-red-500 text-red-400',
    high: 'bg-orange-900/30 border-orange-500 text-orange-400',
    medium: 'bg-yellow-900/30 border-yellow-500 text-yellow-400',
    low: 'bg-blue-900/30 border-blue-500 text-blue-400',
    info: 'bg-gray-800 border-gray-600 text-gray-400',
  }

  const colors =
    severityColors[issue.severity as keyof typeof severityColors] || severityColors.info

  return (
    <div className={`p-3 rounded-lg border ${colors}`}>
      <div className="flex items-center gap-2 mb-1">
        <span className="text-xs uppercase font-medium">{issue.severity}</span>
        {issue.category && (
          <span className="text-xs px-2 py-0.5 bg-va-bg rounded">{issue.category}</span>
        )}
        {issue.code && (
          <span className="text-xs font-mono text-va-text-muted">{issue.code}</span>
        )}
      </div>
      <p className="text-sm">{issue.message}</p>
      {issue.location && (
        <p className="text-xs text-va-text-muted mt-1">Location: {issue.location}</p>
      )}
      {issue.suggestion && (
        <p className="text-xs text-va-accent mt-1">Suggestion: {issue.suggestion}</p>
      )}
    </div>
  )
}

function DimensionBar({ name, score }: { name: string; score: number }) {
  const percent = score * 100
  const color =
    percent >= 80
      ? 'bg-green-500'
      : percent >= 60
        ? 'bg-blue-500'
        : percent >= 40
          ? 'bg-yellow-500'
          : 'bg-red-500'

  return (
    <div className="flex items-center gap-2">
      <span className="w-32 text-sm text-va-text-muted truncate" title={name}>
        {name}
      </span>
      <div className="flex-1 h-2 bg-va-border rounded-full overflow-hidden">
        <div className={`h-full ${color} transition-all`} style={{ width: `${percent}%` }} />
      </div>
      <span className="w-12 text-xs text-va-text-muted text-right">
        {percent.toFixed(0)}%
      </span>
    </div>
  )
}

export function AIDLCDocumentView({ projectName, docType, onBack }: AIDLCDocumentViewProps) {
  const [document, setDocument] = useState<AIDLCDocument | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const loadDocument = useCallback(async () => {
    if (!projectName || !docType) return

    setIsLoading(true)
    setError(null)

    try {
      const doc = await api.getAIDLCDocument(projectName, docType)
      setDocument(doc)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load document')
    } finally {
      setIsLoading(false)
    }
  }, [projectName, docType])

  useEffect(() => {
    loadDocument()
  }, [loadDocument])

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
          onClick={loadDocument}
          className="mt-2 px-3 py-1 bg-red-900/50 hover:bg-red-900/70 rounded text-sm"
        >
          Retry
        </button>
      </div>
    )
  }

  if (!document) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-va-text-muted">
        <p>Document not found</p>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-va-border">
        <div className="flex items-center gap-4">
          {onBack && (
            <button
              onClick={onBack}
              className="px-2 py-1 hover:bg-va-panel rounded text-va-text-muted"
            >
              &larr; Back
            </button>
          )}
          <div>
            <h2 className="text-lg font-semibold text-va-text">{document.title}</h2>
            <div className="flex items-center gap-2 text-sm text-va-text-muted">
              <span className="capitalize">{document.phase}</span>
              <span>|</span>
              <span className="capitalize">{document.status.replace('_', ' ')}</span>
              {document.path && (
                <>
                  <span>|</span>
                  <span className="font-mono text-xs">{document.path}</span>
                </>
              )}
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {document.score && <QualityBadge score={document.score} />}
          <button
            onClick={loadDocument}
            className="px-3 py-1 bg-va-panel hover:bg-va-border rounded text-sm"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-4">
        {document.score && (
          <div className="mb-6">
            <h3 className="text-sm font-semibold text-va-text mb-3">Evaluation Results</h3>

            {/* Dimensions */}
            {document.score.dimensions && Object.keys(document.score.dimensions).length > 0 && (
              <div className="mb-4 p-4 bg-va-panel rounded-lg space-y-2">
                {Object.entries(document.score.dimensions).map(([name, score]) => (
                  <DimensionBar key={name} name={name} score={score} />
                ))}
              </div>
            )}

            {/* Issues */}
            {document.score.issues && document.score.issues.length > 0 && (
              <div className="space-y-2">
                <h4 className="text-xs font-semibold text-va-text-muted uppercase">
                  Issues ({document.score.issues.length})
                </h4>
                {document.score.issues.map((issue, idx) => (
                  <IssueCard key={idx} issue={issue} />
                ))}
              </div>
            )}
          </div>
        )}

        {/* Document Content */}
        {document.content ? (
          <div className="bg-va-panel rounded-lg p-4">
            <h3 className="text-sm font-semibold text-va-text mb-3">Document Content</h3>
            <pre className="text-sm text-va-text-muted whitespace-pre-wrap font-mono">
              {document.content}
            </pre>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center h-32 text-va-text-muted bg-va-panel rounded-lg">
            <p>No content available</p>
            <p className="text-sm mt-1">Document may not exist yet</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default AIDLCDocumentView
