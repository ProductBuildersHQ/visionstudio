import { useState, useEffect, useRef } from 'react'
import { api } from '../../services/api'
import type { MaturityModelSummary } from './types'

interface MaturityModelViewProps {
  projectName: string
}

/**
 * MaturityModelView displays the prism-maturity HTML dashboard
 * embedded in an iframe for the selected project.
 */
export function MaturityModelView({ projectName }: MaturityModelViewProps) {
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dashboardHtml, setDashboardHtml] = useState<string>('')
  const [models, setModels] = useState<MaturityModelSummary[]>([])
  const [selectedModelId, setSelectedModelId] = useState<string | null>(null)
  const iframeRef = useRef<HTMLIFrameElement>(null)

  // Load available models on mount
  useEffect(() => {
    loadModels()
  }, [projectName])

  // Load dashboard when model is selected
  useEffect(() => {
    if (selectedModelId) {
      loadDashboard(selectedModelId)
    }
  }, [selectedModelId, projectName])

  async function loadModels() {
    try {
      const modelList = await api.listMaturityModels(projectName)
      setModels(modelList)
      if (modelList.length > 0 && !selectedModelId) {
        setSelectedModelId(modelList[0].id)
      } else if (modelList.length === 0) {
        // No models - show placeholder
        setIsLoading(false)
        setDashboardHtml('')
      }
    } catch (err) {
      console.error('Failed to load maturity models:', err)
      // Continue anyway - we'll try to load a default dashboard
      loadDashboard('default')
    }
  }

  async function loadDashboard(modelId: string) {
    setIsLoading(true)
    setError(null)
    try {
      const html = await api.getMaturityDashboard(projectName, {
        theme: 'dark',
        modelId,
      })
      setDashboardHtml(html)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load dashboard'
      setError(message)
    } finally {
      setIsLoading(false)
    }
  }

  function handleRefresh() {
    if (selectedModelId) {
      loadDashboard(selectedModelId)
    } else {
      loadDashboard('default')
    }
  }

  function handleModelChange(modelId: string) {
    setSelectedModelId(modelId)
  }

  // Loading state
  if (isLoading && !dashboardHtml) {
    return (
      <div className="flex items-center justify-center h-full bg-va-bg">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-2 border-va-accent border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-va-text-muted">Loading maturity dashboard...</p>
        </div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-va-bg gap-4">
        <div className="text-center max-w-md">
          <div className="text-va-error mb-4">{error}</div>
          <p className="text-va-text-muted text-sm mb-4">
            Make sure the project has maturity model data configured.
          </p>
          <button
            onClick={handleRefresh}
            className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  // No models available
  if (models.length === 0 && !dashboardHtml) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-va-bg">
        <div className="text-center max-w-md">
          <div className="text-5xl mb-4">📊</div>
          <h2 className="text-xl font-semibold text-va-text mb-2">No Maturity Model</h2>
          <p className="text-va-text-muted text-sm mb-4">
            This project doesn't have a maturity model configured yet.
            Add a maturity model to track capability progression.
          </p>
          <div className="text-xs text-va-text-muted">
            <code>.visionspec/maturity/model.json</code>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col bg-va-bg">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-va-border bg-va-sidebar">
        <div className="flex items-center gap-4">
          <h2 className="text-lg font-semibold text-va-text">Maturity Model</h2>

          {/* Model selector */}
          {models.length > 1 && (
            <select
              value={selectedModelId || ''}
              onChange={(e) => handleModelChange(e.target.value)}
              className="bg-va-panel border border-va-border rounded px-2 py-1 text-sm text-va-text"
            >
              {models.map((model) => (
                <option key={model.id} value={model.id}>
                  {model.name}
                </option>
              ))}
            </select>
          )}
        </div>

        <div className="flex items-center gap-2">
          {/* Loading indicator */}
          {isLoading && (
            <div className="animate-spin w-4 h-4 border-2 border-va-accent border-t-transparent rounded-full" />
          )}

          {/* Refresh button */}
          <button
            onClick={handleRefresh}
            disabled={isLoading}
            className="p-1.5 rounded hover:bg-va-panel text-va-text-muted hover:text-va-text disabled:opacity-50 transition-colors"
            title="Refresh"
          >
            <RefreshIcon className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Dashboard iframe */}
      <div className="flex-1 overflow-hidden">
        {dashboardHtml ? (
          <iframe
            ref={iframeRef}
            srcDoc={dashboardHtml}
            className="w-full h-full border-0"
            sandbox="allow-scripts allow-same-origin"
            title="Maturity Model Dashboard"
          />
        ) : (
          <div className="flex items-center justify-center h-full text-va-text-muted">
            Select a maturity model to view
          </div>
        )}
      </div>
    </div>
  )
}

function RefreshIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
      />
    </svg>
  )
}
