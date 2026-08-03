import { useState, useCallback, useRef } from 'react'
import { MarkdownPreview } from './MarkdownPreview'
import { MarkdownSource } from './MarkdownSource'

type ViewMode = 'display' | 'markdown'

interface MarkdownViewerProps {
  content: string
  title?: string
  metadata?: {
    path?: string
    modTime?: string
  }
  initialMode?: ViewMode
  onModeChange?: (mode: ViewMode) => void
  className?: string
}

export function MarkdownViewer({
  content,
  title,
  metadata,
  initialMode = 'display',
  onModeChange,
  className = '',
}: MarkdownViewerProps) {
  const [mode, setMode] = useState<ViewMode>(initialMode)
  const containerRef = useRef<HTMLDivElement>(null)

  const handleModeChange = useCallback(
    (newMode: ViewMode) => {
      setMode(newMode)
      onModeChange?.(newMode)
    },
    [onModeChange]
  )

  const handleCopyLink = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(window.location.href)
    } catch {
      const input = document.createElement('input')
      input.value = window.location.href
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
    }
  }, [])

  const handleDownloadPDF = useCallback(async () => {
    const printWindow = window.open('', '_blank')
    if (printWindow) {
      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
          <head>
            <title>${title ?? 'Document'}</title>
            <style>
              body { font-family: system-ui, sans-serif; margin: 40px; line-height: 1.6; }
              h1 { border-bottom: 2px solid #333; padding-bottom: 8px; }
              h2 { border-bottom: 1px solid #ccc; padding-bottom: 6px; margin-top: 24px; }
              code { background: #f5f5f5; padding: 2px 4px; border-radius: 3px; }
              pre { background: #f5f5f5; padding: 12px; border-radius: 4px; overflow-x: auto; }
              table { border-collapse: collapse; width: 100%; }
              th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
              blockquote { border-left: 4px solid #ddd; margin: 16px 0; padding-left: 16px; color: #666; }
            </style>
          </head>
          <body>
            ${containerRef.current?.querySelector('.mde-prose')?.innerHTML ?? content}
            <script>window.onload = function() { window.print(); }</script>
          </body>
        </html>
      `)
      printWindow.document.close()
    }
  }, [content, title])

  return (
    <div ref={containerRef} className={`mde-viewer ${className}`}>
      {/* Header */}
      <div className="mde-viewer-header">
        {title && <h1 className="mde-viewer-title">{title}</h1>}
        <div className="mde-viewer-controls">
          <div className="mde-mode-toggle">
            <button
              onClick={() => handleModeChange('display')}
              className={`mde-mode-btn ${mode === 'display' ? 'active' : ''}`}
            >
              Display
            </button>
            <button
              onClick={() => handleModeChange('markdown')}
              className={`mde-mode-btn ${mode === 'markdown' ? 'active' : ''}`}
            >
              Markdown
            </button>
          </div>
          <button onClick={handleCopyLink} className="mde-action-btn">
            Copy Link
          </button>
          <button onClick={handleDownloadPDF} className="mde-action-btn">
            Download PDF
          </button>
        </div>
      </div>

      {/* Metadata */}
      {metadata && (
        <div className="mde-viewer-meta">
          {metadata.path && <span className="mde-meta-path">{metadata.path}</span>}
          {metadata.modTime && (
            <span className="mde-meta-date">
              Modified: {new Date(metadata.modTime).toLocaleDateString()}
            </span>
          )}
        </div>
      )}

      {/* Content */}
      <div className="mde-viewer-content">
        {mode === 'display' ? (
          <MarkdownPreview content={content} />
        ) : (
          <MarkdownSource content={content} />
        )}
      </div>
    </div>
  )
}
