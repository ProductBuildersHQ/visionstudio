import { useState, useEffect, useMemo } from 'react'
import { useParams, useSearchParams, useNavigate, Link } from 'react-router-dom'
import { getSpecFiles } from '../api/client'
import type { SpecFile } from '../api/types'
import { LoadingState, ErrorState, MarkdownPreview, MarkdownSource } from '../components'

type ViewMode = 'display' | 'markdown'

export function SpecViewer() {
  const { initiativeId, specType } = useParams<{ initiativeId: string; specType: string }>()
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()

  const [specFiles, setSpecFiles] = useState<SpecFile[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const modeParam = searchParams.get('mode')
  const [mode, setMode] = useState<ViewMode>(
    modeParam === 'markdown' ? 'markdown' : 'display'
  )

  useEffect(() => {
    if (!initiativeId) return
    setLoading(true)
    setError(null)
    getSpecFiles(initiativeId)
      .then((resp) => setSpecFiles(resp.files ?? []))
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false))
  }, [initiativeId])

  const selectedSpec = useMemo(() => {
    if (!specType) return specFiles[0]
    return specFiles.find((f) => f.specType.toLowerCase() === specType.toLowerCase()) ?? specFiles[0]
  }, [specFiles, specType])

  const handleModeChange = (newMode: ViewMode) => {
    setMode(newMode)
    const newParams = new URLSearchParams(searchParams)
    if (newMode === 'markdown') {
      newParams.set('mode', 'markdown')
    } else {
      newParams.delete('mode')
    }
    setSearchParams(newParams, { replace: true })
  }

  const handleSpecChange = (newSpecType: string) => {
    navigate(`/initiative/${initiativeId}/spec/${newSpecType.toLowerCase()}${mode === 'markdown' ? '?mode=markdown' : ''}`)
  }

  const copyLink = async () => {
    const url = window.location.href
    try {
      await navigator.clipboard.writeText(url)
    } catch {
      // Fallback for older browsers
      const input = document.createElement('input')
      input.value = url
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
    }
  }

  const handleDownloadPDF = async () => {
    if (!selectedSpec) return

    const previewEl = document.querySelector('.mde-prose')
    const renderedContent = previewEl?.innerHTML ?? selectedSpec.content

    const printWindow = window.open('', '_blank')
    if (printWindow) {
      printWindow.document.write(`
        <!DOCTYPE html>
        <html>
          <head>
            <title>${initiativeId}-${selectedSpec.specType}</title>
            <style>
              body { font-family: system-ui, sans-serif; margin: 40px; line-height: 1.6; color: #1a1a1a; }
              .spec-header { margin-bottom: 24px; padding-bottom: 16px; border-bottom: 2px solid #333; }
              .spec-header h1 { margin: 0 0 8px 0; font-size: 24px; }
              .spec-meta { color: #666; font-size: 12px; }
              .mde-prose h1 { font-size: 24px; border-bottom: 1px solid #ccc; padding-bottom: 8px; margin: 24px 0 16px; }
              .mde-prose h1:first-child { margin-top: 0; }
              .mde-prose h2 { font-size: 20px; border-bottom: 1px solid #eee; padding-bottom: 6px; margin-top: 24px; }
              .mde-prose h3 { font-size: 16px; margin-top: 20px; }
              .mde-prose p { margin: 12px 0; }
              .mde-prose ul, .mde-prose ol { margin: 12px 0; padding-left: 24px; }
              .mde-prose li { margin: 4px 0; }
              .mde-prose table { width: 100%; border-collapse: collapse; margin: 16px 0; }
              .mde-prose th, .mde-prose td { border: 1px solid #ddd; padding: 8px; text-align: left; }
              .mde-prose th { background: #f5f5f5; }
              .mde-prose code { background: #f5f5f5; padding: 2px 4px; border-radius: 3px; font-size: 0.9em; }
              .mde-prose pre { background: #f5f5f5; padding: 12px; border-radius: 4px; overflow-x: auto; }
              .mde-prose pre code { background: none; padding: 0; }
              .mde-prose blockquote { border-left: 4px solid #ddd; margin: 16px 0; padding-left: 16px; color: #666; }
            </style>
          </head>
          <body>
            <div class="spec-header">
              <h1>${initiativeId} - ${selectedSpec.specType}</h1>
              <div class="spec-meta">
                Generated: ${new Date().toLocaleDateString()} |
                Path: ${selectedSpec.path}
                ${selectedSpec.modTime ? ` | Modified: ${new Date(selectedSpec.modTime).toLocaleDateString()}` : ''}
              </div>
            </div>
            <div class="mde-prose">${renderedContent}</div>
            <script>window.onload = function() { window.print(); }</script>
          </body>
        </html>
      `)
      printWindow.document.close()
    }
  }

  if (error) {
    return (
      <div className="p-6">
        <BackLink initiativeId={initiativeId} />
        <ErrorState message={error} onRetry={() => window.location.reload()} />
      </div>
    )
  }

  if (loading) {
    return (
      <div className="p-6">
        <BackLink initiativeId={initiativeId} />
        <LoadingState message="Loading spec..." />
      </div>
    )
  }

  if (specFiles.length === 0) {
    return (
      <div className="p-6">
        <BackLink initiativeId={initiativeId} />
        <div className="bg-gray-800 rounded-lg p-8 text-center">
          <div className="text-gray-400 mb-2">No spec files found</div>
          <div className="text-sm text-gray-500">
            Add specs to docs/specs/initiatives/{initiativeId}/
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-4">
      <BackLink initiativeId={initiativeId} />

      {/* Header with spec tabs and controls */}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        {/* Spec Type Tabs */}
        <div className="border-b border-gray-700 px-4 flex items-center justify-between">
          <div className="flex gap-1">
            {specFiles.map((f) => (
              <button
                key={f.specType}
                onClick={() => handleSpecChange(f.specType)}
                className={`px-3 py-2 text-sm font-medium border-b-2 transition-colors ${
                  selectedSpec?.specType === f.specType
                    ? 'border-purple-500 text-purple-400'
                    : 'border-transparent text-gray-400 hover:text-gray-200'
                }`}
              >
                {f.specType}
              </button>
            ))}
          </div>

          {/* View Mode Toggle + Actions */}
          <div className="flex items-center gap-2 py-2">
            <div className="flex rounded-md overflow-hidden border border-gray-600">
              <button
                onClick={() => handleModeChange('display')}
                className={`px-3 py-1 text-xs font-medium transition-colors ${
                  mode === 'display'
                    ? 'bg-purple-600 text-white'
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                Display
              </button>
              <button
                onClick={() => handleModeChange('markdown')}
                className={`px-3 py-1 text-xs font-medium transition-colors ${
                  mode === 'markdown'
                    ? 'bg-purple-600 text-white'
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                Markdown
              </button>
            </div>

            <button
              onClick={copyLink}
              className="px-3 py-1 text-xs font-medium bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
              title="Copy link to clipboard"
            >
              Copy Link
            </button>

            <button
              onClick={handleDownloadPDF}
              className="px-3 py-1 text-xs font-medium bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
              title="Download as PDF"
            >
              Download PDF
            </button>
          </div>
        </div>

        {/* Spec Metadata */}
        {selectedSpec && (
          <div className="px-4 py-2 border-b border-gray-700 flex items-center justify-between text-xs text-gray-500">
            <span className="font-mono">{selectedSpec.path.split('/').slice(-3).join('/')}</span>
            {selectedSpec.modTime && (
              <span>Modified: {new Date(selectedSpec.modTime).toLocaleDateString()}</span>
            )}
          </div>
        )}

        {/* Content - using @grokify/markdown-editor reusable components */}
        <div className="p-4">
          {mode === 'display' ? (
            <div className="bg-gray-900 rounded-lg p-6 max-h-[70vh] overflow-y-auto">
              <MarkdownPreview content={selectedSpec?.content ?? ''} />
            </div>
          ) : (
            <div className="max-h-[70vh] overflow-y-auto">
              <MarkdownSource content={selectedSpec?.content ?? ''} />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function BackLink({ initiativeId }: { initiativeId?: string }) {
  return (
    <Link
      to={initiativeId ? `/initiative/${initiativeId}` : '/'}
      className="text-sm text-gray-400 hover:text-gray-200 flex items-center gap-1"
    >
      <span>←</span> Back to {initiativeId ?? 'Overview'}
    </Link>
  )
}
