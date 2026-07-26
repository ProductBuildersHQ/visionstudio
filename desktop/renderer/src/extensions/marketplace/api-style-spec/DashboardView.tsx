import { useState, useEffect } from 'react'
import { SummaryCard, SeverityDot, LoadingState, ErrorState, EmptyState } from '../../../components/toolkit'
import type { ExtensionViewProps } from '../../../types/extension'
import type { EvaluationReport } from './types'

export function DashboardView({ context }: ExtensionViewProps) {
  const [report, setReport] = useState<EvaluationReport | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const data = await context.api.getProjectData<EvaluationReport>('evaluation')
      setReport(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load evaluation report')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [context.projectName])

  if (isLoading) return <LoadingState message="Loading evaluation report..." />
  if (error) return <ErrorState message={error} onRetry={load} />
  if (!report) {
    return (
      <EmptyState
        icon="📊"
        title="No Evaluation Report"
        description="Run an evaluation against your API style spec to see results here."
        hint="api-style evaluate --input spec.yaml --profile default"
      />
    )
  }

  const decision = report.decision
  const findingCounts = decision?.findingCounts
  const categoryCounts = decision?.categoryCounts

  const overallColor = report.overallDecision === 'pass'
    ? 'text-va-success'
    : report.overallDecision === 'partial'
      ? 'text-va-warning'
      : 'text-va-error'

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-5xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold text-va-text">API Style Evaluation</h1>
            {report.metadata?.documentTitle && (
              <p className="text-sm text-va-text-muted mt-0.5">{report.metadata.documentTitle}</p>
            )}
          </div>
          <div className="flex items-center gap-3">
            {report.metadata?.generatedAt && (
              <span className="text-xs text-va-text-muted">
                {new Date(report.metadata.generatedAt).toLocaleDateString()}
              </span>
            )}
            <button
              onClick={load}
              className="px-3 py-1.5 text-xs bg-va-panel border border-va-border rounded hover:bg-va-border transition-colors text-va-text"
            >
              Refresh
            </button>
          </div>
        </div>

        {/* Summary cards */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          <SummaryCard
            label="Decision"
            value={report.overallDecision.toUpperCase()}
            color={overallColor}
          />
          <SummaryCard
            label="Categories"
            value={`${categoryCounts?.pass ?? 0}/${categoryCounts?.total ?? 0} pass`}
            color={categoryCounts?.fail ? 'text-va-warning' : 'text-va-success'}
          />
          <SummaryCard
            label="Total Findings"
            value={findingCounts?.total ?? report.findings.length}
          />
          <SummaryCard
            label="Critical/High"
            value={(findingCounts?.critical ?? 0) + (findingCounts?.high ?? 0)}
            color={(findingCounts?.critical ?? 0) + (findingCounts?.high ?? 0) > 0
              ? 'text-va-error'
              : 'text-va-success'
            }
          />
        </div>

        {/* Severity breakdown */}
        <div className="bg-va-panel rounded-lg p-4 border border-va-border mb-6">
          <h2 className="text-sm font-semibold text-va-text mb-3">Findings by Severity</h2>
          <div className="flex gap-6">
            <SeverityDot severity="critical" label="Critical" count={findingCounts?.critical ?? 0} />
            <SeverityDot severity="high" label="High" count={findingCounts?.high ?? 0} />
            <SeverityDot severity="medium" label="Medium" count={findingCounts?.medium ?? 0} />
            <SeverityDot severity="low" label="Low" count={findingCounts?.low ?? 0} />
          </div>
        </div>

        {/* Category scores table */}
        <div className="bg-va-panel rounded-lg border border-va-border mb-6 overflow-hidden">
          <div className="px-4 py-3 border-b border-va-border">
            <h2 className="text-sm font-semibold text-va-text">Category Scores</h2>
          </div>
          <div className="divide-y divide-va-border">
            {report.categories.map((cat) => (
              <div key={cat.category} className="px-4 py-3 flex items-center gap-4">
                <ScoreDot score={cat.score} />
                <div className="flex-1 min-w-0">
                  <span className="text-sm font-medium text-va-text">{cat.category}</span>
                  {cat.reasoning && (
                    <p className="text-xs text-va-text-muted mt-0.5 line-clamp-1">{cat.reasoning}</p>
                  )}
                </div>
                <span className="text-sm font-mono text-va-text-muted">{cat.numericScore}/5</span>
                <ScoreLabel score={cat.score} />
              </div>
            ))}
          </div>
        </div>

        {/* Next Steps */}
        {report.nextSteps && (
          <div className="bg-va-panel rounded-lg border border-va-border overflow-hidden">
            <div className="px-4 py-3 border-b border-va-border">
              <h2 className="text-sm font-semibold text-va-text">Next Steps</h2>
            </div>
            <div className="divide-y divide-va-border">
              {report.nextSteps.immediate?.map((item, idx) => (
                <div key={`imm-${idx}`} className="px-4 py-3 flex items-start gap-3">
                  <span className="text-xs font-bold px-1.5 py-0.5 rounded bg-red-500 text-white shrink-0">
                    IMMEDIATE
                  </span>
                  <div className="flex-1">
                    <p className="text-sm text-va-text">{item.action}</p>
                    <div className="flex gap-3 mt-1">
                      {item.category && (
                        <span className="text-xs text-va-text-muted">{item.category}</span>
                      )}
                      {item.effort && (
                        <span className="text-xs text-va-text-muted">Effort: {item.effort}</span>
                      )}
                    </div>
                  </div>
                </div>
              ))}
              {report.nextSteps.recommended?.map((item, idx) => (
                <div key={`rec-${idx}`} className="px-4 py-3 flex items-start gap-3">
                  <span className="text-xs font-bold px-1.5 py-0.5 rounded bg-blue-500 text-white shrink-0">
                    RECOMMENDED
                  </span>
                  <div className="flex-1">
                    <p className="text-sm text-va-text">{item.action}</p>
                    <div className="flex gap-3 mt-1">
                      {item.category && (
                        <span className="text-xs text-va-text-muted">{item.category}</span>
                      )}
                      {item.effort && (
                        <span className="text-xs text-va-text-muted">Effort: {item.effort}</span>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Summary text */}
        {report.summary && (
          <div className="mt-6 p-4 bg-va-panel/50 rounded-lg border border-va-border">
            <p className="text-sm text-va-text-muted">{report.summary}</p>
          </div>
        )}
      </div>
    </div>
  )
}

function ScoreDot({ score }: { score: string }) {
  const colors: Record<string, string> = {
    pass: 'bg-va-success',
    partial: 'bg-va-warning',
    fail: 'bg-va-error',
  }
  return <span className={`w-3 h-3 rounded-full shrink-0 ${colors[score] ?? 'bg-va-text-muted'}`} />
}

function ScoreLabel({ score }: { score: string }) {
  const styles: Record<string, string> = {
    pass: 'text-va-success',
    partial: 'text-va-warning',
    fail: 'text-va-error',
  }
  return (
    <span className={`text-xs font-semibold uppercase ${styles[score] ?? 'text-va-text-muted'}`}>
      {score}
    </span>
  )
}
