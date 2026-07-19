import { useState, useMemo } from 'react'
import { marked } from 'marked'
import type { Spec, ViewMode, EvalResult, Finding } from '../../types'
import { getScoreLabel, needsHumanReview } from '../../types'
import { DimensionScoreCard } from '../eval/DimensionScoreCard'

// Configure marked for GitHub-flavored markdown with tables
marked.setOptions({
  gfm: true,
  breaks: true,
})

interface SpecEditorProps {
  spec: Spec
  onContentChange: (content: string) => void
  onSave: () => void
  isDirty: boolean
}

export function SpecEditor({ spec, onContentChange, onSave, isDirty }: SpecEditorProps) {
  // Default to rendered view for reading
  const [viewMode, setViewMode] = useState<ViewMode>('rendered')
  const [showEvalPanel, setShowEvalPanel] = useState(!!spec.evalResult)

  const hasEval = !!spec.evalResult

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 bg-va-sidebar border-b border-va-border">
        <div className="flex items-center gap-3">
          <h2 className="text-sm font-semibold text-va-text">
            {spec.name}
            {isDirty && <span className="text-va-warning ml-1">•</span>}
          </h2>
          <span className="text-xs text-va-text-muted">{spec.path}</span>
        </div>

        <div className="flex items-center gap-3">
          {/* View mode toggle */}
          <div className="flex bg-va-panel rounded overflow-hidden border border-va-border">
            <button
              onClick={() => setViewMode('source')}
              className={`px-3 py-1 text-xs font-medium transition-colors ${
                viewMode === 'source'
                  ? 'bg-va-accent text-white'
                  : 'text-va-text-muted hover:text-va-text'
              }`}
            >
              Source
            </button>
            <button
              onClick={() => setViewMode('rendered')}
              className={`px-3 py-1 text-xs font-medium transition-colors ${
                viewMode === 'rendered'
                  ? 'bg-va-accent text-white'
                  : 'text-va-text-muted hover:text-va-text'
              }`}
            >
              Rendered
            </button>
          </div>

          {/* Eval panel toggle */}
          {hasEval && (
            <button
              onClick={() => setShowEvalPanel(!showEvalPanel)}
              className={`px-3 py-1 text-xs font-medium rounded border transition-colors flex items-center gap-1.5 ${
                showEvalPanel
                  ? 'bg-va-accent/20 border-va-accent text-va-accent'
                  : 'border-va-border text-va-text-muted hover:text-va-text hover:border-va-text-muted'
              }`}
            >
              <EvalBadge decision={spec.evalResult!.decision} />
              Eval
            </button>
          )}

          {/* Save button */}
          <button
            onClick={onSave}
            disabled={!isDirty}
            className="px-3 py-1 bg-va-success text-white rounded text-xs font-medium disabled:opacity-50 disabled:cursor-not-allowed hover:bg-va-success/80 transition-colors"
          >
            Save
          </button>
        </div>
      </div>

      {/* Content area */}
      <div className="flex-1 overflow-hidden flex">
        {/* Main editor/view */}
        <div className={`${showEvalPanel && hasEval ? 'w-2/3' : 'w-full'} h-full overflow-hidden transition-all`}>
          {viewMode === 'source' ? (
            <SourceEditor
              content={spec.content || ''}
              onChange={onContentChange}
            />
          ) : (
            <RenderedView content={spec.content || ''} />
          )}
        </div>

        {/* Eval panel */}
        {showEvalPanel && hasEval && (
          <div className="w-1/3 border-l border-va-border overflow-hidden">
            <EvalPanel evalResult={spec.evalResult!} />
          </div>
        )}
      </div>
    </div>
  )
}

// Small badge to show eval decision status
function EvalBadge({ decision }: { decision: string }) {
  const colors = {
    pass: 'bg-va-success',
    conditional: 'bg-va-warning',
    fail: 'bg-va-danger',
  }
  return (
    <span className={`w-2 h-2 rounded-full ${colors[decision as keyof typeof colors] || 'bg-va-text-muted'}`} />
  )
}

// Evaluation panel component with v2 support
function EvalPanel({ evalResult }: { evalResult: EvalResult }) {
  const [showAllFindings, setShowAllFindings] = useState(false)

  // Detect v2 format
  const isV2 = evalResult.schemaVersion === 'v2' || evalResult.scoreV2 !== undefined

  const decisionColors = {
    pass: { bg: 'bg-va-success/10', border: 'border-va-success/30', text: 'text-va-success' },
    conditional: { bg: 'bg-va-warning/10', border: 'border-va-warning/30', text: 'text-va-warning' },
    fail: { bg: 'bg-va-danger/10', border: 'border-va-danger/30', text: 'text-va-danger' },
  }
  const decision = decisionColors[evalResult.decision] || decisionColors.pass

  // V2 score display
  const displayScore = isV2 && evalResult.scoreV2
    ? `${evalResult.scoreV2}/5`
    : evalResult.score.toFixed(1)
  const scoreLabel = isV2 && evalResult.scoreV2 ? getScoreLabel(evalResult.scoreV2) : null

  // Check if needs human review
  const needsReview = needsHumanReview(evalResult)

  // Has dimensions for v2 display
  const hasDimensions = evalResult.dimensions && evalResult.dimensions.length > 0

  return (
    <div className="h-full overflow-y-auto bg-va-bg">
      {/* Header with pass/fail */}
      <div className={`p-4 ${decision.bg} border-b ${decision.border}`}>
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-va-text">LLM-as-a-Judge</h3>
          <div className="flex items-center gap-2">
            {needsReview && (
              <span className="text-[10px] px-1.5 py-0.5 bg-va-warning/20 text-va-warning rounded font-medium">
                Needs Review
              </span>
            )}
            <span className={`text-xs font-bold px-2 py-0.5 rounded ${decision.text} ${decision.bg}`}>
              {evalResult.decision.toUpperCase()}
            </span>
          </div>
        </div>
        <div className="mt-2 flex items-baseline gap-2">
          <span className={`text-2xl font-bold ${decision.text}`}>
            {displayScore}
          </span>
          {scoreLabel && (
            <span className="text-sm text-va-text-muted">({scoreLabel})</span>
          )}
          {!isV2 && <span className="text-xs text-va-text-muted">/ 10.0</span>}
        </div>

        {/* V2: Confidence indicator */}
        {isV2 && evalResult.confidence !== undefined && (
          <div className="mt-3">
            <div className="flex items-center gap-2">
              <span className="text-xs text-va-text-muted">Confidence:</span>
              <div className="flex-1 h-1.5 bg-va-border rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full ${evalResult.confidence < 0.7 ? 'bg-va-warning' : 'bg-va-success'}`}
                  style={{ width: `${Math.round(evalResult.confidence * 100)}%` }}
                />
              </div>
              <span className={`text-xs font-medium ${evalResult.confidence < 0.7 ? 'text-va-warning' : 'text-va-text'}`}>
                {Math.round(evalResult.confidence * 100)}%
              </span>
            </div>
          </div>
        )}
      </div>

      {/* V2: Blocking issues */}
      {isV2 && evalResult.blocking && evalResult.blocking.length > 0 && (
        <div className="p-3 bg-va-error/5 border-b border-va-error/20">
          <h4 className="text-xs font-semibold text-va-error uppercase tracking-wide mb-2">
            Blocking Issues
          </h4>
          <div className="flex flex-wrap gap-1">
            {evalResult.blocking.map((code, idx) => (
              <span key={idx} className="text-[10px] px-1.5 py-0.5 bg-va-error/10 text-va-error border border-va-error/30 rounded font-mono">
                {code}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* V2: Dimensions grid */}
      {hasDimensions && (
        <div className="p-4 border-b border-va-border">
          <h4 className="text-xs font-semibold text-va-text-muted uppercase tracking-wide mb-3">
            Dimensions ({evalResult.dimensions!.length})
          </h4>
          <div className="space-y-2">
            {evalResult.dimensions!.map((dim, idx) => (
              <DimensionScoreCard key={idx} dimension={dim} />
            ))}
          </div>
        </div>
      )}

      {/* Findings section */}
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <h4 className="text-xs font-semibold text-va-text-muted uppercase tracking-wide">
            {hasDimensions ? 'All Findings' : 'Findings'} ({evalResult.findings.length})
          </h4>
          {hasDimensions && evalResult.findings.length > 0 && (
            <button
              onClick={() => setShowAllFindings(!showAllFindings)}
              className="text-[10px] text-va-accent hover:underline"
            >
              {showAllFindings ? 'Hide' : 'Show All'}
            </button>
          )}
        </div>

        {/* Show findings if no dimensions, or if showAllFindings is true */}
        {(!hasDimensions || showAllFindings) && (
          <div className="space-y-2">
            {evalResult.findings.length === 0 ? (
              <p className="text-xs text-va-text-muted italic">No findings</p>
            ) : (
              evalResult.findings.map((finding, idx) => (
                <FindingCard key={idx} finding={finding} />
              ))
            )}
          </div>
        )}

        {/* Quick summary when dimensions present but findings collapsed */}
        {hasDimensions && !showAllFindings && evalResult.findings.length > 0 && (
          <div className="text-xs text-va-text-muted">
            <span className="text-red-400">{evalResult.findings.filter(f => f.severity === 'critical').length} critical</span>
            {' · '}
            <span className="text-orange-400">{evalResult.findings.filter(f => f.severity === 'high').length} high</span>
            {' · '}
            <span className="text-yellow-400">{evalResult.findings.filter(f => f.severity === 'medium').length} medium</span>
            {' · '}
            <span className="text-blue-400">{evalResult.findings.filter(f => f.severity === 'low').length} low</span>
          </div>
        )}
      </div>
    </div>
  )
}

// Individual finding card with color indicators and v2 support
function FindingCard({ finding }: { finding: Finding }) {
  const severityStyles = {
    critical: {
      bg: 'bg-red-500/10',
      border: 'border-l-red-500',
      badge: 'bg-red-500 text-white',
      text: 'text-red-400',
    },
    high: {
      bg: 'bg-orange-500/10',
      border: 'border-l-orange-500',
      badge: 'bg-orange-500 text-white',
      text: 'text-orange-400',
    },
    medium: {
      bg: 'bg-yellow-500/10',
      border: 'border-l-yellow-500',
      badge: 'bg-yellow-500 text-black',
      text: 'text-yellow-400',
    },
    low: {
      bg: 'bg-blue-500/10',
      border: 'border-l-blue-500',
      badge: 'bg-blue-500 text-white',
      text: 'text-blue-400',
    },
    info: {
      bg: 'bg-va-panel',
      border: 'border-l-va-text-muted',
      badge: 'bg-va-text-muted text-va-bg',
      text: 'text-va-text-muted',
    },
  }

  const style = severityStyles[finding.severity] || severityStyles.info

  return (
    <div className={`${style.bg} ${style.border} border-l-2 rounded-r p-3`}>
      <div className="flex items-center gap-2 mb-1 flex-wrap">
        <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${style.badge}`}>
          {finding.severity.toUpperCase()}
        </span>
        <span className="text-xs text-va-text-muted capitalize">{finding.category}</span>
        {/* V2: Reason code badge */}
        {finding.code && (
          <span className="text-[10px] px-1.5 py-0.5 bg-va-panel border border-va-border rounded font-mono">
            {finding.code}
          </span>
        )}
        {/* V2: Location reference */}
        {finding.location && (
          <span className="text-[10px] text-va-text-muted">
            @ {finding.location}
          </span>
        )}
      </div>
      <p className="text-xs text-va-text leading-relaxed">{finding.message}</p>
    </div>
  )
}

// Source editor (markdown textarea for now, can upgrade to Monaco later)
function SourceEditor({
  content,
  onChange,
}: {
  content: string
  onChange: (content: string) => void
}) {
  return (
    <textarea
      value={content}
      onChange={(e) => onChange(e.target.value)}
      className="w-full h-full p-4 bg-va-bg text-va-text font-mono text-sm resize-none focus:outline-none"
      placeholder="Start writing your spec..."
      spellCheck={false}
    />
  )
}

// Rendered markdown view using marked library
function RenderedView({ content }: { content: string }) {
  const html = useMemo(() => {
    return marked.parse(content || '') as string
  }, [content])

  return (
    <div
      className="h-full overflow-y-auto p-6 markdown-content"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}
