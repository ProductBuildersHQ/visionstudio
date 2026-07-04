import { useState } from 'react'
import type { Spec, ViewMode } from '../../types'

interface SpecEditorProps {
  spec: Spec
  onContentChange: (content: string) => void
  onSave: () => void
  isDirty: boolean
}

export function SpecEditor({ spec, onContentChange, onSave, isDirty }: SpecEditorProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('source')

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

      {/* Editor content */}
      <div className="flex-1 overflow-hidden">
        {viewMode === 'source' ? (
          <SourceEditor
            content={spec.content || ''}
            onChange={onContentChange}
          />
        ) : (
          <RenderedView content={spec.content || ''} />
        )}
      </div>
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
