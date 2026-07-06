import { useState } from 'react'
import type { Spec, ViewMode, EvalResult, Finding } from '../../types'

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

// Evaluation panel component
function EvalPanel({ evalResult }: { evalResult: EvalResult }) {
  const decisionColors = {
    pass: { bg: 'bg-va-success/10', border: 'border-va-success/30', text: 'text-va-success' },
    conditional: { bg: 'bg-va-warning/10', border: 'border-va-warning/30', text: 'text-va-warning' },
    fail: { bg: 'bg-va-danger/10', border: 'border-va-danger/30', text: 'text-va-danger' },
  }
  const decision = decisionColors[evalResult.decision] || decisionColors.pass

  return (
    <div className="h-full overflow-y-auto bg-va-bg">
      {/* Header */}
      <div className={`p-4 ${decision.bg} border-b ${decision.border}`}>
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-va-text">LLM-as-a-Judge</h3>
          <span className={`text-xs font-bold ${decision.text}`}>
            {evalResult.decision.toUpperCase()}
          </span>
        </div>
        <div className="mt-2 flex items-baseline gap-2">
          <span className={`text-2xl font-bold ${decision.text}`}>
            {evalResult.score.toFixed(1)}
          </span>
          <span className="text-xs text-va-text-muted">/ 10.0</span>
        </div>
      </div>

      {/* Findings */}
      <div className="p-4">
        <h4 className="text-xs font-semibold text-va-text-muted uppercase tracking-wide mb-3">
          Findings ({evalResult.findings.length})
        </h4>
        <div className="space-y-2">
          {evalResult.findings.length === 0 ? (
            <p className="text-xs text-va-text-muted italic">No findings</p>
          ) : (
            evalResult.findings.map((finding, idx) => (
              <FindingCard key={idx} finding={finding} />
            ))
          )}
        </div>
      </div>
    </div>
  )
}

// Individual finding card with color indicators
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
      <div className="flex items-center gap-2 mb-1">
        <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${style.badge}`}>
          {finding.severity.toUpperCase()}
        </span>
        <span className="text-xs text-va-text-muted capitalize">{finding.category}</span>
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

// Rendered markdown view (basic for now, can add proper markdown renderer)
function RenderedView({ content }: { content: string }) {
  // Basic markdown rendering - TODO: use a proper markdown library
  const renderMarkdown = (md: string): string => {
    return md
      // Headers
      .replace(/^### (.*$)/gim, '<h3 class="text-lg font-semibold mt-4 mb-2">$1</h3>')
      .replace(/^## (.*$)/gim, '<h2 class="text-xl font-semibold mt-6 mb-3">$1</h2>')
      .replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold mt-8 mb-4">$1</h1>')
      // Bold
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      // Italic
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      // Code blocks
      .replace(/```([\s\S]*?)```/g, '<pre class="bg-va-panel p-3 rounded my-2 overflow-x-auto"><code>$1</code></pre>')
      // Inline code
      .replace(/`(.*?)`/g, '<code class="bg-va-panel px-1 rounded">$1</code>')
      // Lists
      .replace(/^\- (.*$)/gim, '<li class="ml-4">$1</li>')
      // Links
      .replace(/\[(.*?)\]\((.*?)\)/g, '<a href="$2" class="text-va-accent hover:underline">$1</a>')
      // Paragraphs
      .replace(/\n\n/g, '</p><p class="my-2">')
      // Line breaks
      .replace(/\n/g, '<br/>')
  }

  return (
    <div
      className="h-full overflow-y-auto p-6 prose prose-invert max-w-none"
      dangerouslySetInnerHTML={{
        __html: `<p class="my-2">${renderMarkdown(content)}</p>`,
      }}
    />
  )
}
