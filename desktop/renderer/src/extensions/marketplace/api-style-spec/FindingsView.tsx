import { useState, useEffect } from 'react'
import { IssueCard, SeverityDot, LoadingState, ErrorState, EmptyState } from '../../../components/toolkit'
import type { ExtensionViewProps } from '../../../types/extension'
import type { EvaluationReport, EvaluationFinding } from './types'

type GroupBy = 'category' | 'severity'

export function FindingsExplorer({ context }: ExtensionViewProps) {
  const [report, setReport] = useState<EvaluationReport | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [groupBy, setGroupBy] = useState<GroupBy>('category')
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const data = await context.api.getProjectData<EvaluationReport>('evaluation')
      setReport(data)
      if (data) {
        const groups = groupFindings(data.findings, 'category')
        setExpandedGroups(new Set(Object.keys(groups)))
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load findings')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [context.projectName])

  useEffect(() => {
    if (report) {
      const groups = groupFindings(report.findings, groupBy)
      setExpandedGroups(new Set(Object.keys(groups)))
    }
  }, [groupBy])

  if (isLoading) return <LoadingState message="Loading findings..." />
  if (error) return <ErrorState message={error} onRetry={load} />
  if (!report || report.findings.length === 0) {
    return (
      <EmptyState
        icon="📋"
        title="No Findings"
        description="No evaluation findings found. Run an evaluation to see results."
      />
    )
  }

  const groups = groupFindings(report.findings, groupBy)

  const toggleGroup = (key: string) => {
    const next = new Set(expandedGroups)
    if (next.has(key)) {
      next.delete(key)
    } else {
      next.add(key)
    }
    setExpandedGroups(next)
  }

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold text-va-text">Evaluation Findings</h1>
            <p className="text-sm text-va-text-muted">{report.findings.length} findings across {report.categories.length} categories</p>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-va-text-muted">Group by:</span>
            <button
              onClick={() => setGroupBy('category')}
              className={`px-2 py-1 text-xs rounded ${groupBy === 'category' ? 'bg-va-accent text-white' : 'bg-va-panel text-va-text-muted hover:text-va-text'}`}
            >
              Category
            </button>
            <button
              onClick={() => setGroupBy('severity')}
              className={`px-2 py-1 text-xs rounded ${groupBy === 'severity' ? 'bg-va-accent text-white' : 'bg-va-panel text-va-text-muted hover:text-va-text'}`}
            >
              Severity
            </button>
          </div>
        </div>

        {/* Severity summary */}
        <div className="bg-va-panel rounded-lg p-4 border border-va-border mb-6">
          <div className="flex gap-6">
            {(['critical', 'high', 'medium', 'low'] as const).map((sev) => (
              <SeverityDot
                key={sev}
                severity={sev}
                label={sev.charAt(0).toUpperCase() + sev.slice(1)}
                count={report.findings.filter(f => f.severity === sev).length}
              />
            ))}
          </div>
        </div>

        {/* Grouped findings */}
        <div className="space-y-3">
          {Object.entries(groups).map(([key, findings]) => (
            <div key={key} className="bg-va-panel rounded-lg border border-va-border overflow-hidden">
              <button
                onClick={() => toggleGroup(key)}
                className="w-full px-4 py-3 flex items-center justify-between hover:bg-va-border/30 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <span className="text-xs text-va-text-muted">{expandedGroups.has(key) ? '▼' : '►'}</span>
                  <span className="font-medium text-sm text-va-text capitalize">{key}</span>
                </div>
                <span className="text-xs text-va-text-muted">{findings.length} finding{findings.length !== 1 ? 's' : ''}</span>
              </button>
              {expandedGroups.has(key) && (
                <div className="border-t border-va-border divide-y divide-va-border">
                  {findings.map((finding, idx) => (
                    <IssueCard
                      key={idx}
                      issue={{
                        severity: finding.severity,
                        category: groupBy === 'severity' ? finding.category : undefined,
                        message: finding.finding,
                        recommendation: finding.recommendation,
                        location: finding.location,
                        code: finding.ruleId,
                      }}
                    />
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function groupFindings(findings: EvaluationFinding[], by: GroupBy): Record<string, EvaluationFinding[]> {
  const groups: Record<string, EvaluationFinding[]> = {}
  const severityOrder = ['critical', 'high', 'medium', 'low']

  for (const finding of findings) {
    const key = by === 'category' ? finding.category : finding.severity
    if (!groups[key]) groups[key] = []
    groups[key].push(finding)
  }

  if (by === 'severity') {
    const ordered: Record<string, EvaluationFinding[]> = {}
    for (const sev of severityOrder) {
      if (groups[sev]) ordered[sev] = groups[sev]
    }
    return ordered
  }

  return groups
}
