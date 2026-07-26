import type { Project, Spec, EvalFinding } from '../../types'
import { needsHumanReview } from '../../types'
import { SummaryCard, SeverityDot, IssueCard } from '../toolkit'
import { useApp } from '../../contexts/AppContext'

interface FindingsViewProps {
  project?: Project
  onSpecClick?: (spec: Spec) => void
}

interface SpecFinding extends EvalFinding {
  specType: string
  specName: string
}

export function FindingsView(props: FindingsViewProps) {
  const app = useApp()
  const project = props.project ?? app.activeProject
  const onSpecClick = props.onSpecClick ?? app.navigateToSpec
  if (!project) return null
  const allFindings: SpecFinding[] = []
  const specStats: { spec: Spec; findingCount: number; score: number | undefined }[] = []

  for (const spec of project.specs) {
    if (spec.evalResult) {
      const findings = spec.evalResult.findings ?? []
      specStats.push({
        spec,
        findingCount: findings.length,
        score: spec.evalResult.intScore,
      })
      for (const finding of findings) {
        allFindings.push({
          ...finding,
          specType: spec.type,
          specName: spec.name,
        })
      }
    }
  }

  const severityCounts = {
    critical: allFindings.filter((f) => f.severity === 'critical').length,
    high: allFindings.filter((f) => f.severity === 'high').length,
    medium: allFindings.filter((f) => f.severity === 'medium').length,
    low: allFindings.filter((f) => f.severity === 'low').length,
  }

  const totalFindings = allFindings.length
  const evaluatedSpecs = specStats.length
  const passingSpecs = specStats.filter((s) => s.spec.evalResult?.overallDecision === 'pass').length

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-xl font-semibold text-va-text mb-1">
            LLM-as-a-Judge Findings
          </h1>
          <p className="text-sm text-va-text-muted">
            Consolidated view of all evaluation findings across specs
          </p>
        </div>

        {/* Summary cards */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          <SummaryCard label="Total Findings" value={totalFindings} />
          <SummaryCard
            label="Specs Evaluated"
            value={`${evaluatedSpecs} / ${project.specs.length}`}
          />
          <SummaryCard label="Passing" value={passingSpecs} color="text-va-success" />
          <SummaryCard
            label="Needs Work"
            value={evaluatedSpecs - passingSpecs}
            color={evaluatedSpecs - passingSpecs > 0 ? 'text-va-warning' : 'text-va-text-muted'}
          />
        </div>

        {/* Severity breakdown */}
        <div className="bg-va-panel rounded-lg p-4 border border-va-border mb-6">
          <h2 className="text-sm font-semibold text-va-text mb-3">Findings by Severity</h2>
          <div className="flex gap-6">
            <SeverityDot severity="critical" label="Critical" count={severityCounts.critical} />
            <SeverityDot severity="high" label="High" count={severityCounts.high} />
            <SeverityDot severity="medium" label="Medium" count={severityCounts.medium} />
            <SeverityDot severity="low" label="Low" count={severityCounts.low} />
          </div>
        </div>

        {/* Findings by spec */}
        {totalFindings === 0 ? (
          <div className="bg-va-panel rounded-lg p-8 border border-va-border text-center">
            <p className="text-va-text-muted">No findings yet. Evaluate specs to see results.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {project.specs
              .filter((spec) => spec.evalResult && (spec.evalResult.findings?.length ?? 0) > 0)
              .map((spec) => (
                <SpecFindingsCard
                  key={spec.type}
                  spec={spec}
                  onSpecClick={() => onSpecClick(spec)}
                />
              ))}
          </div>
        )}
      </div>
    </div>
  )
}

function SpecFindingsCard({
  spec,
  onSpecClick,
}: {
  spec: Spec
  onSpecClick: () => void
}) {
  if (!spec.evalResult) return null

  const evalResult = spec.evalResult
  const findings = evalResult.findings ?? []
  const overallDecision = evalResult.overallDecision

  const displayScore = evalResult.intScore !== undefined ? `${evalResult.intScore}/5` : '—'

  const decisionStyles: Record<string, { bg: string; border: string; badge: string }> = {
    pass: { bg: 'bg-va-success/10', border: 'border-va-success/30', badge: 'bg-va-success' },
    conditional: { bg: 'bg-va-warning/10', border: 'border-va-warning/30', badge: 'bg-va-warning' },
    fail: { bg: 'bg-va-error/10', border: 'border-va-error/30', badge: 'bg-va-error' },
    human_review: { bg: 'bg-va-warning/10', border: 'border-va-warning/30', badge: 'bg-va-warning' },
  }
  const style = (overallDecision && decisionStyles[overallDecision]) || decisionStyles.fail

  const needsReview = needsHumanReview(evalResult)

  return (
    <div className={`rounded-lg border ${style.border} overflow-hidden`}>
      <button
        onClick={onSpecClick}
        className={`w-full ${style.bg} px-4 py-3 flex items-center justify-between hover:brightness-110 transition-all`}
      >
        <div className="flex items-center gap-3">
          <span className={`w-2 h-2 rounded-full ${style.badge}`} />
          <span className="font-semibold text-va-text">{spec.name}</span>
          <span className="text-xs text-va-text-muted">({spec.type})</span>
          {needsReview && (
            <span className="text-[10px] px-1.5 py-0.5 bg-va-warning/20 text-va-warning rounded">
              Needs Review
            </span>
          )}
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-va-text-muted">
            Score: <span className="font-semibold text-va-text">{displayScore}</span>
          </span>
          {evalResult.confidence !== undefined && (
            <span className="text-sm text-va-text-muted">
              Conf: <span className="font-semibold text-va-text">{Math.round(evalResult.confidence * 100)}%</span>
            </span>
          )}
          <span className="text-sm text-va-text-muted">
            {findings.length} finding{findings.length !== 1 ? 's' : ''}
          </span>
          <span className="text-xs text-va-accent">View spec →</span>
        </div>
      </button>

      {evalResult.blocking && evalResult.blocking.length > 0 && (
        <div className="bg-va-error/5 px-4 py-2 border-b border-va-border">
          <span className="text-xs text-va-error font-semibold">Blocking: </span>
          {evalResult.blocking.map((code, idx) => (
            <span key={idx} className="text-xs text-va-error bg-va-error/10 px-1.5 py-0.5 rounded mr-1">
              {code}
            </span>
          ))}
        </div>
      )}

      <div className="bg-va-bg divide-y divide-va-border">
        {findings.map((finding, idx) => (
          <IssueCard key={idx} issue={finding} />
        ))}
      </div>
    </div>
  )
}
