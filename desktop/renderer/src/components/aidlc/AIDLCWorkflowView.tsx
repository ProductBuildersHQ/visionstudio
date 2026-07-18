import { useState, useEffect, useCallback } from 'react'
import { api, AIDLCWorkflow, AIDLCWorkflowNode, AIDLCPhase } from '../../services/api'

interface AIDLCWorkflowViewProps {
  projectName: string
  onDocumentSelect?: (docType: string) => void
}

const phaseColors: Record<AIDLCPhase, { bg: string; border: string; text: string }> = {
  inception: { bg: 'bg-blue-900/30', border: 'border-blue-500', text: 'text-blue-400' },
  construction: { bg: 'bg-purple-900/30', border: 'border-purple-500', text: 'text-purple-400' },
  operations: { bg: 'bg-green-900/30', border: 'border-green-500', text: 'text-green-400' },
}

const statusColors: Record<string, { bg: string; border: string }> = {
  pending: { bg: 'bg-gray-700', border: 'border-gray-500' },
  ready: { bg: 'bg-yellow-900/50', border: 'border-yellow-500' },
  in_progress: { bg: 'bg-blue-900/50', border: 'border-blue-400' },
  completed: { bg: 'bg-green-900/50', border: 'border-green-400' },
  blocked: { bg: 'bg-red-900/50', border: 'border-red-500' },
  skipped: { bg: 'bg-gray-800', border: 'border-gray-600' },
  failed: { bg: 'bg-red-900/70', border: 'border-red-400' },
}

function StatusIcon({ status }: { status: string }) {
  switch (status) {
    case 'completed':
      return <span className="text-green-400">&#10003;</span>
    case 'in_progress':
      return <span className="text-blue-400 animate-pulse">&#9679;</span>
    case 'blocked':
    case 'failed':
      return <span className="text-red-400">&#10007;</span>
    case 'ready':
      return <span className="text-yellow-400">&#9675;</span>
    default:
      return <span className="text-gray-500">&#9675;</span>
  }
}

function NodeCard({
  node,
  onClick,
}: {
  node: AIDLCWorkflowNode
  onClick?: () => void
}) {
  const colors = statusColors[node.status] || statusColors.pending

  return (
    <button
      onClick={onClick}
      className={`w-full text-left p-3 rounded-lg border-2 ${colors.bg} ${colors.border}
        hover:brightness-110 transition-all cursor-pointer`}
    >
      <div className="flex items-center justify-between mb-1">
        <span className="font-medium text-va-text">{node.name}</span>
        <StatusIcon status={node.status} />
      </div>
      <div className="text-xs text-va-text-muted">
        {node.required ? 'Required' : 'Optional'}
        {node.automated && ' | Auto'}
      </div>
      {node.score && (
        <div className="mt-2 flex items-center gap-2">
          <span
            className={`text-xs px-2 py-0.5 rounded ${
              node.score.rating === 'EXCELLENT'
                ? 'bg-green-900/50 text-green-400'
                : node.score.rating === 'GOOD'
                  ? 'bg-blue-900/50 text-blue-400'
                  : node.score.rating === 'NEEDS_IMPROVEMENT'
                    ? 'bg-yellow-900/50 text-yellow-400'
                    : 'bg-red-900/50 text-red-400'
            }`}
          >
            {node.score.rating}
          </span>
          <span className="text-xs text-va-text-muted">
            {(node.score.score * 100).toFixed(0)}%
          </span>
        </div>
      )}
    </button>
  )
}

function PhaseColumn({
  phase,
  nodes,
  onNodeClick,
}: {
  phase: { id: string; name: string; node_ids: string[] }
  nodes: Record<string, AIDLCWorkflowNode>
  onNodeClick?: (docType: string) => void
}) {
  const colors = phaseColors[phase.id as AIDLCPhase] || phaseColors.inception
  const phaseNodes = phase.node_ids.map((id) => nodes[id]).filter(Boolean)
  const completed = phaseNodes.filter((n) => n.status === 'completed').length
  const total = phaseNodes.length
  const progress = total > 0 ? (completed / total) * 100 : 0

  return (
    <div className={`flex-1 p-4 rounded-lg ${colors.bg} border ${colors.border}`}>
      <div className="flex items-center justify-between mb-3">
        <h3 className={`font-semibold ${colors.text}`}>{phase.name}</h3>
        <span className="text-xs text-va-text-muted">
          {completed}/{total}
        </span>
      </div>

      {/* Progress bar */}
      <div className="h-1 bg-va-border rounded-full mb-4 overflow-hidden">
        <div
          className={`h-full ${colors.text.replace('text-', 'bg-')} transition-all`}
          style={{ width: `${progress}%` }}
        />
      </div>

      {/* Nodes */}
      <div className="space-y-2">
        {phaseNodes.map((node) => (
          <NodeCard
            key={node.id}
            node={node}
            onClick={() => onNodeClick?.(node.doc_type)}
          />
        ))}
      </div>
    </div>
  )
}

export function AIDLCWorkflowView({ projectName, onDocumentSelect }: AIDLCWorkflowViewProps) {
  const [workflow, setWorkflow] = useState<AIDLCWorkflow | null>(null)
  const [mermaid, setMermaid] = useState<string>('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'columns' | 'diagram'>('columns')

  const loadWorkflow = useCallback(async () => {
    if (!projectName) return

    setIsLoading(true)
    setError(null)

    try {
      const data = await api.getAIDLCWorkflow(projectName)
      setWorkflow(data.workflow)
      setMermaid(data.mermaid)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load workflow')
    } finally {
      setIsLoading(false)
    }
  }, [projectName])

  useEffect(() => {
    loadWorkflow()
  }, [loadWorkflow])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-va-accent"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 bg-red-900/20 border border-red-500 rounded-lg">
        <p className="text-red-400">{error}</p>
        <button
          onClick={loadWorkflow}
          className="mt-2 px-3 py-1 bg-red-900/50 hover:bg-red-900/70 rounded text-sm"
        >
          Retry
        </button>
      </div>
    )
  }

  if (!workflow) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-va-text-muted">
        <p>No AIDLC workflow found</p>
        <p className="text-sm mt-2">Create aidlc-docs/ directory to get started</p>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-4 px-4 pt-4">
        <div>
          <h2 className="text-lg font-semibold text-va-text">{workflow.name}</h2>
          <p className="text-sm text-va-text-muted">{workflow.description}</p>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-sm text-va-text-muted">
            {workflow.progress.completed}/{workflow.progress.total} completed
          </span>
          <div className="w-24 h-2 bg-va-border rounded-full overflow-hidden">
            <div
              className="h-full bg-va-accent transition-all"
              style={{ width: `${workflow.progress.percent}%` }}
            />
          </div>
          <button
            onClick={() => setViewMode(viewMode === 'columns' ? 'diagram' : 'columns')}
            className="px-3 py-1 bg-va-panel hover:bg-va-border rounded text-sm"
          >
            {viewMode === 'columns' ? 'Diagram' : 'Columns'}
          </button>
          <button
            onClick={loadWorkflow}
            className="px-3 py-1 bg-va-panel hover:bg-va-border rounded text-sm"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto px-4 pb-4">
        {viewMode === 'columns' ? (
          <div className="flex gap-4 min-h-full">
            {workflow.phases.map((phase) => (
              <PhaseColumn
                key={phase.id}
                phase={phase}
                nodes={workflow.nodes}
                onNodeClick={onDocumentSelect}
              />
            ))}
          </div>
        ) : (
          <div className="bg-va-panel rounded-lg p-4 overflow-auto">
            <pre className="text-xs text-va-text-muted whitespace-pre-wrap">{mermaid}</pre>
          </div>
        )}
      </div>
    </div>
  )
}

export default AIDLCWorkflowView
