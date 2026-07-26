import { useState, useEffect } from 'react'
import { SummaryCard, IssueCard, LoadingState, ErrorState, EmptyState } from '../../../components/toolkit'
import type { ExtensionViewProps } from '../../../types/extension'
import type { LintReport } from './types'

export function LintResultsView({ context }: ExtensionViewProps) {
  const [report, setReport] = useState<LintReport | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const data = await context.api.getProjectData<LintReport>('lint')
      setReport(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load lint results')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [context.projectName])

  if (isLoading) return <LoadingState message="Loading lint results..." />
  if (error) return <ErrorState message={error} onRetry={load} />
  if (!report || report.violations.length === 0) {
    return (
      <EmptyState
        icon="🔍"
        title="No Lint Violations"
        description="No violations found, or linting has not been run yet."
        hint="api-style lint --input spec.yaml --profile default"
      />
    )
  }

  const { summary } = report

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold text-va-text">Lint Results</h1>
            {report.metadata.specFile && (
              <p className="text-sm text-va-text-muted mt-0.5 font-mono">{report.metadata.specFile}</p>
            )}
          </div>
          <div className="flex items-center gap-3">
            {report.metadata.profile && (
              <span className="text-xs px-2 py-1 bg-va-panel border border-va-border rounded text-va-text-muted">
                Profile: {report.metadata.profile}
              </span>
            )}
            {report.metadata.duration && (
              <span className="text-xs text-va-text-muted">{report.metadata.duration}</span>
            )}
            <button
              onClick={load}
              className="px-3 py-1.5 text-xs bg-va-panel border border-va-border rounded hover:bg-va-border transition-colors text-va-text"
            >
              Re-lint
            </button>
          </div>
        </div>

        {/* Summary cards */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          <SummaryCard
            label="Errors"
            value={summary.errors}
            color={summary.errors > 0 ? 'text-va-error' : 'text-va-success'}
          />
          <SummaryCard
            label="Warnings"
            value={summary.warnings}
            color={summary.warnings > 0 ? 'text-va-warning' : 'text-va-text-muted'}
          />
          <SummaryCard label="Info" value={summary.infos} />
          <SummaryCard label="Hints" value={summary.hints} />
        </div>

        {/* Violations */}
        <div className="space-y-1">
          {report.violations.map((v, idx) => (
            <IssueCard
              key={idx}
              issue={{
                severity: v.severity === 'error' ? 'critical'
                  : v.severity === 'warn' ? 'high'
                  : v.severity === 'info' ? 'medium'
                  : 'low',
                message: v.message,
                code: v.ruleId,
                location: v.path
                  ? `${v.path}${v.line ? `:${v.line}` : ''}${v.column ? `:${v.column}` : ''}`
                  : undefined,
                suggestion: v.suggestion,
              }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
