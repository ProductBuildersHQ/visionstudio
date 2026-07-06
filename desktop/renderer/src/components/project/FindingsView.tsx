import type { Project, Spec, Finding } from '../../types'

interface FindingsViewProps {
  project: Project
  onSpecClick: (spec: Spec) => void
}

interface SpecFinding extends Finding {
  specType: string
  specName: string
}

export function FindingsView({ project, onSpecClick }: FindingsViewProps) {
  // Collect all findings from all specs, preserving spec order from sidebar
  const allFindings: SpecFinding[] = []
  const specStats: { spec: Spec; findingCount: number; score: number | null }[] = []

  for (const spec of project.specs) {
    if (spec.evalResult) {
      specStats.push({
        spec,
        findingCount: spec.evalResult.findings.length,
        score: spec.evalResult.score,
      })
      for (const finding of spec.evalResult.findings) {
        allFindings.push({
          ...finding,
          specType: spec.type,
          specName: spec.name,
        })
      }
    }
  }

  // Count findings by severity
  const severityCounts = {
    critical: allFindings.filter((f) => f.severity === 'critical').length,
    high: allFindings.filter((f) => f.severity === 'high').length,
    medium: allFindings.filter((f) => f.severity === 'medium').length,
    low: allFindings.filter((f) => f.severity === 'low').length,
  }

  const totalFindings = allFindings.length
  const evaluatedSpecs = specStats.length
  const passingSpecs = specStats.filter((s) => s.spec.evalResult?.decision === 'pass').length

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
          <SummaryCard
            label="Total Findings"
            value={totalFindings}
            color="text-va-text"
          />
          <SummaryCard
            label="Specs Evaluated"
            value={`${evaluatedSpecs} / ${project.specs.length}`}
            color="text-va-text"
          />
          <SummaryCard
            label="Passing"
            value={passingSpecs}
            color="text-va-success"
          />
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
            <SeverityBadge label="Critical" count={severityCounts.critical} color="bg-red-500" />
            <SeverityBadge label="High" count={severityCounts.high} color="bg-orange-500" />
            <SeverityBadge label="Medium" count={severityCounts.medium} color="bg-yellow-500" />
            <SeverityBadge label="Low" count={severityCounts.low} color="bg-blue-500" />
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
              .filter((spec) => spec.evalResult && spec.evalResult.findings.length > 0)
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

function SummaryCard({
  label,
  value,
  color,
}: {
  label: string
  value: string | number
  color: string
}) {
  return (
    <div className="bg-va-panel rounded-lg p-4 border border-va-border">
      <div className="text-xs text-va-text-muted uppercase tracking-wide mb-1">{label}</div>
      <div className={`text-2xl font-bold ${color}`}>{value}</div>
    </div>
  )
}

function SeverityBadge({
  label,
  count,
  color,
}: {
  label: string
  count: number
  color: string
}) {
  return (
    <div className="flex items-center gap-2">
      <span className={`w-3 h-3 rounded-full ${color} ${count === 0 ? 'opacity-30' : ''}`} />
      <span className={`text-sm ${count === 0 ? 'text-va-text-muted' : 'text-va-text'}`}>
        {label}: <span className="font-semibold">{count}</span>
      </span>
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

  const { decision, score, findings } = spec.evalResult

  const decisionStyles = {
    pass: { bg: 'bg-va-success/10', border: 'border-va-success/30', badge: 'bg-va-success' },
    conditional: { bg: 'bg-va-warning/10', border: 'border-va-warning/30', badge: 'bg-va-warning' },
    fail: { bg: 'bg-va-error/10', border: 'border-va-error/30', badge: 'bg-va-error' },
  }
  const style = decisionStyles[decision] || decisionStyles.fail

  return (
    <div className={`rounded-lg border ${style.border} overflow-hidden`}>
      {/* Header */}
      <button
        onClick={onSpecClick}
        className={`w-full ${style.bg} px-4 py-3 flex items-center justify-between hover:brightness-110 transition-all`}
      >
        <div className="flex items-center gap-3">
          <span className={`w-2 h-2 rounded-full ${style.badge}`} />
          <span className="font-semibold text-va-text">{spec.name}</span>
          <span className="text-xs text-va-text-muted">({spec.type})</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-va-text-muted">
            Score: <span className="font-semibold text-va-text">{score.toFixed(1)}</span>
          </span>
          <span className="text-sm text-va-text-muted">
            {findings.length} finding{findings.length !== 1 ? 's' : ''}
          </span>
          <span className="text-xs text-va-accent">View spec →</span>
        </div>
      </button>

      {/* Findings list */}
      <div className="bg-va-bg divide-y divide-va-border">
        {findings.map((finding, idx) => (
          <FindingRow key={idx} finding={finding} />
        ))}
      </div>
    </div>
  )
}

function FindingRow({ finding }: { finding: Finding }) {
  const severityStyles: Record<string, { badge: string; border: string }> = {
    critical: { badge: 'bg-red-500 text-white', border: 'border-l-red-500' },
    high: { badge: 'bg-orange-500 text-white', border: 'border-l-orange-500' },
    medium: { badge: 'bg-yellow-500 text-black', border: 'border-l-yellow-500' },
    low: { badge: 'bg-blue-500 text-white', border: 'border-l-blue-500' },
    info: { badge: 'bg-gray-500 text-white', border: 'border-l-gray-500' },
  }
  const style = severityStyles[finding.severity] || severityStyles.low

  return (
    <div className={`px-4 py-3 border-l-4 ${style.border}`}>
      <div className="flex items-start gap-3">
        <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded shrink-0 ${style.badge}`}>
          {finding.severity.toUpperCase()}
        </span>
        <div className="flex-1 min-w-0">
          <span className="text-xs text-va-text-muted capitalize">{finding.category}</span>
          <p className="text-sm text-va-text mt-0.5">{finding.message}</p>
        </div>
      </div>
    </div>
  )
}
