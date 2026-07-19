import type { DimensionScore, Finding } from '../../types'
import { getScoreLabel, SCORE_COLORS } from '../../types'
import { useState } from 'react'

interface DimensionScoreCardProps {
  dimension: DimensionScore
}

export function DimensionScoreCard({ dimension }: DimensionScoreCardProps) {
  const [expanded, setExpanded] = useState(false)

  const scoreColor = SCORE_COLORS[dimension.score] || 'gray'
  const scoreLabel = getScoreLabel(dimension.score)

  // Severity determines the card border styling
  const severityStyles: Record<string, { border: string; bg: string }> = {
    none: { border: 'border-va-border', bg: 'bg-va-panel' },
    minor: { border: 'border-yellow-500/30', bg: 'bg-yellow-500/5' },
    major: { border: 'border-orange-500/30', bg: 'bg-orange-500/5' },
    critical: { border: 'border-red-500/30', bg: 'bg-red-500/5' },
  }
  const style = severityStyles[dimension.severity] || severityStyles.none

  // Score color mapping
  const scoreColors: Record<string, string> = {
    red: 'text-red-500 bg-red-500/10',
    orange: 'text-orange-500 bg-orange-500/10',
    yellow: 'text-yellow-500 bg-yellow-500/10',
    green: 'text-green-500 bg-green-500/10',
    blue: 'text-blue-500 bg-blue-500/10',
  }
  const scoreStyle = scoreColors[scoreColor] || 'text-va-text bg-va-panel'

  const hasFindings = dimension.findings && dimension.findings.length > 0
  const lowConfidence = dimension.confidence < 0.7

  return (
    <div className={`rounded-lg border ${style.border} ${style.bg} overflow-hidden`}>
      {/* Header - clickable to expand */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-3 py-2 flex items-center justify-between hover:bg-va-panel/50 transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-va-text">{dimension.name}</span>
          {lowConfidence && (
            <span className="text-[10px] px-1.5 py-0.5 bg-va-warning/20 text-va-warning rounded">
              Low Conf.
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className={`text-xs font-semibold px-2 py-0.5 rounded ${scoreStyle}`}>
            {dimension.score}/5
          </span>
          <span className="text-[10px] text-va-text-muted">{scoreLabel}</span>
          {hasFindings && (
            <span className="text-[10px] text-va-text-muted">
              {expanded ? '▼' : '▶'} {dimension.findings.length}
            </span>
          )}
        </div>
      </button>

      {/* Reason codes */}
      {dimension.reasonCodes && dimension.reasonCodes.length > 0 && (
        <div className="px-3 pb-2 flex flex-wrap gap-1">
          {dimension.reasonCodes.map((code, idx) => (
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
      <div className="px-3 pb-2">
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-va-text-muted">Conf:</span>
          <div className="flex-1 h-1 bg-va-border rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full ${lowConfidence ? 'bg-va-warning' : 'bg-va-success'}`}
              style={{ width: `${Math.round(dimension.confidence * 100)}%` }}
            />
          </div>
          <span className="text-[10px] text-va-text-muted">
            {Math.round(dimension.confidence * 100)}%
          </span>
        </div>
      </div>

      {/* Expanded findings */}
      {expanded && hasFindings && (
        <div className="border-t border-va-border bg-va-bg px-3 py-2 space-y-2">
          {dimension.findings.map((finding, idx) => (
            <DimensionFinding key={idx} finding={finding} />
          ))}
        </div>
      )}
    </div>
  )
}

function DimensionFinding({ finding }: { finding: Finding }) {
  const severityStyles: Record<string, { badge: string; border: string }> = {
    critical: { badge: 'bg-red-500 text-white', border: 'border-l-red-500' },
    high: { badge: 'bg-orange-500 text-white', border: 'border-l-orange-500' },
    medium: { badge: 'bg-yellow-500 text-black', border: 'border-l-yellow-500' },
    low: { badge: 'bg-blue-500 text-white', border: 'border-l-blue-500' },
    info: { badge: 'bg-gray-500 text-white', border: 'border-l-gray-500' },
  }
  const style = severityStyles[finding.severity] || severityStyles.info

  return (
    <div className={`border-l-2 ${style.border} pl-2`}>
      <div className="flex items-center gap-1.5 mb-0.5">
        <span className={`text-[9px] font-bold px-1 py-0.5 rounded ${style.badge}`}>
          {finding.severity.toUpperCase()}
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
      <p className="text-[11px] text-va-text leading-snug">{finding.message}</p>
    </div>
  )
}
