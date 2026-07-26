import type { CategoryResult, EvalFinding } from '../../types'
import { getScoreLabel, SCORE_COLORS } from '../../types'
import { useState } from 'react'

interface DimensionScoreCardProps {
  category: CategoryResult
}

export function DimensionScoreCard({ category }: DimensionScoreCardProps) {
  const [expanded, setExpanded] = useState(false)

  const intScore = category.intScore ?? 0
  const scoreColor = SCORE_COLORS[intScore] || 'gray'
  const scoreLabel = getScoreLabel(intScore)

  // Severity comes from CategoryResult.Severity — computed upstream (by
  // structured-evaluation) as the worst severity among this category's
  // findings, using the same 5-level scale as Finding.severity. There is
  // no "none" level; an empty/no-findings category has severity undefined.
  const severityStyles: Record<string, { border: string; bg: string }> = {
    critical: { border: 'border-red-500/30', bg: 'bg-red-500/5' },
    high: { border: 'border-orange-500/30', bg: 'bg-orange-500/5' },
    medium: { border: 'border-yellow-500/30', bg: 'bg-yellow-500/5' },
    low: { border: 'border-va-border', bg: 'bg-va-panel' },
    info: { border: 'border-va-border', bg: 'bg-va-panel' },
  }
  const style = (category.severity && severityStyles[category.severity]) || {
    border: 'border-va-border',
    bg: 'bg-va-panel',
  }

  // Score color mapping
  const scoreColors: Record<string, string> = {
    red: 'text-red-500 bg-red-500/10',
    orange: 'text-orange-500 bg-orange-500/10',
    yellow: 'text-yellow-500 bg-yellow-500/10',
    green: 'text-green-500 bg-green-500/10',
    blue: 'text-blue-500 bg-blue-500/10',
  }
  const scoreStyle = scoreColors[scoreColor] || 'text-va-text bg-va-panel'

  const findings = category.findings ?? []
  const hasFindings = findings.length > 0
  const lowConfidence = category.confidence !== undefined && category.confidence < 0.7

  // There's no human-readable category name in the per-evaluation report —
  // that lives in the rubric definition (RubricSet.Categories[].Name),
  // which this component doesn't have access to. Fall back to the
  // category ID until callers thread the rubric definition through.
  const displayName = category.category ?? 'Unknown'

  return (
    <div className={`rounded-lg border ${style.border} ${style.bg} overflow-hidden`}>
      {/* Header - clickable to expand */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-3 py-2 flex items-center justify-between hover:bg-va-panel/50 transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-va-text capitalize">{displayName}</span>
          {category.score && (
            <span className="text-[10px] px-1.5 py-0.5 bg-va-panel border border-va-border rounded uppercase">
              {category.score}
            </span>
          )}
          {lowConfidence && (
            <span className="text-[10px] px-1.5 py-0.5 bg-va-warning/20 text-va-warning rounded">
              Low Conf.
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className={`text-xs font-semibold px-2 py-0.5 rounded ${scoreStyle}`}>
            {intScore}/5
          </span>
          <span className="text-[10px] text-va-text-muted">{scoreLabel}</span>
          {hasFindings && (
            <span className="text-[10px] text-va-text-muted">
              {expanded ? '▼' : '▶'} {findings.length}
            </span>
          )}
        </div>
      </button>

      {/* Reasoning (chain-of-thought) */}
      {category.reasoning && (
        <div className="px-3 pb-2">
          <p className="text-[11px] text-va-text-muted leading-snug">{category.reasoning}</p>
        </div>
      )}

      {/* Reason codes */}
      {category.reasonCodes && category.reasonCodes.length > 0 && (
        <div className="px-3 pb-2 flex flex-wrap gap-1">
          {category.reasonCodes.map((code, idx) => (
            <span
              key={idx}
              className="text-[10px] px-1.5 py-0.5 bg-va-panel border border-va-border rounded font-mono"
            >
              {code}
            </span>
          ))}
        </div>
      )}

      {/* Confidence bar */}
      {category.confidence !== undefined && (
        <div className="px-3 pb-2">
          <div className="flex items-center gap-2">
            <span className="text-[10px] text-va-text-muted">Conf:</span>
            <div className="flex-1 h-1 bg-va-border rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full ${lowConfidence ? 'bg-va-warning' : 'bg-va-success'}`}
                style={{ width: `${Math.round(category.confidence * 100)}%` }}
              />
            </div>
            <span className="text-[10px] text-va-text-muted">
              {Math.round(category.confidence * 100)}%
            </span>
          </div>
        </div>
      )}

      {/* Expanded findings */}
      {expanded && hasFindings && (
        <div className="border-t border-va-border bg-va-bg px-3 py-2 space-y-2">
          {findings.map((finding, idx) => (
            <DimensionFinding key={idx} finding={finding} />
          ))}
        </div>
      )}
    </div>
  )
}

function DimensionFinding({ finding }: { finding: EvalFinding }) {
  const severityStyles: Record<string, { badge: string; border: string }> = {
    critical: { badge: 'bg-red-500 text-white', border: 'border-l-red-500' },
    high: { badge: 'bg-orange-500 text-white', border: 'border-l-orange-500' },
    medium: { badge: 'bg-yellow-500 text-black', border: 'border-l-yellow-500' },
    low: { badge: 'bg-blue-500 text-white', border: 'border-l-blue-500' },
    info: { badge: 'bg-gray-500 text-white', border: 'border-l-gray-500' },
  }
  const style = (finding.severity && severityStyles[finding.severity]) || severityStyles.info

  return (
    <div className={`border-l-2 ${style.border} pl-2`}>
      <div className="flex items-center gap-1.5 mb-0.5">
        <span className={`text-[9px] font-bold px-1 py-0.5 rounded ${style.badge}`}>
          {finding.severity?.toUpperCase() ?? 'UNKNOWN'}
        </span>
        {finding.code && (
          <span className="text-[9px] px-1 py-0.5 bg-va-panel border border-va-border rounded font-mono">
            {finding.code}
          </span>
        )}
        {finding.location && (
          <span className="text-[9px] text-va-text-muted">@ {finding.location}</span>
        )}
      </div>
      <p className="text-[11px] text-va-text leading-snug font-medium">{finding.title}</p>
      {finding.description && (
        <p className="text-[11px] text-va-text-muted leading-snug">{finding.description}</p>
      )}
    </div>
  )
}
