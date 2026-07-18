import { AIDLCQualityScore, AIDLCIssue } from '../../services/api'

interface EvaluationResultsPanelProps {
  score: AIDLCQualityScore
  compact?: boolean
}

function QualityBadge({ score, size = 'normal' }: { score: AIDLCQualityScore; size?: 'normal' | 'large' }) {
  const colors = {
    EXCELLENT: 'bg-green-900/50 text-green-400 border-green-500',
    GOOD: 'bg-blue-900/50 text-blue-400 border-blue-500',
    NEEDS_IMPROVEMENT: 'bg-yellow-900/50 text-yellow-400 border-yellow-500',
    POOR: 'bg-red-900/50 text-red-400 border-red-500',
  }

  const sizeClasses = size === 'large' ? 'px-4 py-2 text-lg' : 'px-3 py-1'

  return (
    <div className={`rounded-lg border ${colors[score.rating] || colors.POOR} ${sizeClasses}`}>
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
      <span className="w-12 text-xs text-va-text-muted text-right">{percent.toFixed(0)}%</span>
    </div>
  )
}

function IssueSummary({ issues }: { issues: AIDLCIssue[] }) {
  const counts = {
    critical: issues.filter((i) => i.severity === 'critical').length,
    high: issues.filter((i) => i.severity === 'high').length,
    medium: issues.filter((i) => i.severity === 'medium').length,
    low: issues.filter((i) => i.severity === 'low').length,
  }

  return (
    <div className="flex items-center gap-3 text-sm">
      {counts.critical > 0 && (
        <span className="text-red-400">{counts.critical} critical</span>
      )}
      {counts.high > 0 && (
        <span className="text-orange-400">{counts.high} high</span>
      )}
      {counts.medium > 0 && (
        <span className="text-yellow-400">{counts.medium} medium</span>
      )}
      {counts.low > 0 && (
        <span className="text-blue-400">{counts.low} low</span>
      )}
    </div>
  )
}

export function EvaluationResultsPanel({ score, compact = false }: EvaluationResultsPanelProps) {
  const hasDimensions = score.dimensions && Object.keys(score.dimensions).length > 0
  const hasIssues = score.issues && score.issues.length > 0

  if (compact) {
    return (
      <div className="flex items-center justify-between p-3 bg-va-panel rounded-lg">
        <QualityBadge score={score} />
        {hasIssues && <IssueSummary issues={score.issues!} />}
        {score.evaluated_at && (
          <span className="text-xs text-va-text-muted">
            {new Date(score.evaluated_at).toLocaleDateString()}
          </span>
        )}
      </div>
    )
  }

  return (
    <div className="p-4 bg-va-panel rounded-lg space-y-4">
      {/* Header with badge */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-va-text">Evaluation Results</h3>
        <QualityBadge score={score} size="large" />
      </div>

      {/* Dimensions */}
      {hasDimensions && (
        <div className="space-y-2">
          <h4 className="text-xs font-semibold text-va-text-muted uppercase">Dimensions</h4>
          {Object.entries(score.dimensions!).map(([name, value]) => (
            <DimensionBar key={name} name={name} score={value} />
          ))}
        </div>
      )}

      {/* Issues */}
      {hasIssues && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h4 className="text-xs font-semibold text-va-text-muted uppercase">
              Issues ({score.issues!.length})
            </h4>
            <IssueSummary issues={score.issues!} />
          </div>
          <div className="space-y-2 max-h-64 overflow-auto">
            {score.issues!.map((issue, idx) => (
              <IssueCard key={idx} issue={issue} />
            ))}
          </div>
        </div>
      )}

      {/* Timestamp */}
      {score.evaluated_at && (
        <div className="pt-2 border-t border-va-border text-xs text-va-text-muted">
          Evaluated: {new Date(score.evaluated_at).toLocaleString()}
        </div>
      )}
    </div>
  )
}

// Export individual components for reuse
export { QualityBadge, IssueCard, DimensionBar, IssueSummary }

export default EvaluationResultsPanel
