import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import type { CapabilitySummary, CapabilityStack } from '../../services/api'

interface CapabilityStackViewProps {
  projectName: string
}

/**
 * CapabilityStackView displays capability stacks for a project.
 * Shows layers, categories, and individual capabilities with prism integration.
 */
export function CapabilityStackView({ projectName }: CapabilityStackViewProps) {
  const [capabilities, setCapabilities] = useState<CapabilitySummary[]>([])
  const [selectedCapability, setSelectedCapability] = useState<CapabilityStack | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadCapabilities()
  }, [projectName])

  async function loadCapabilities() {
    setIsLoading(true)
    setError(null)
    try {
      const caps = await api.listCapabilities(projectName)
      setCapabilities(caps)
      if (caps.length > 0 && !selectedId) {
        handleCapabilitySelect(caps[0].id)
      } else if (caps.length === 0) {
        setIsLoading(false)
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load capabilities'
      setError(message)
      setIsLoading(false)
    }
  }

  async function handleCapabilitySelect(id: string) {
    setSelectedId(id)
    setIsLoading(true)
    try {
      const cap = await api.getCapability(projectName, id)
      setSelectedCapability(cap)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load capability'
      setError(message)
    } finally {
      setIsLoading(false)
    }
  }

  // Loading state
  if (isLoading && !selectedCapability) {
    return (
      <div className="flex items-center justify-center h-full bg-va-bg">
        <div className="text-center">
          <div className="animate-spin w-8 h-8 border-2 border-va-accent border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-va-text-muted">Loading capabilities...</p>
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
            Make sure the project has capability data configured.
          </p>
          <button
            onClick={loadCapabilities}
            className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  // No capabilities available
  if (capabilities.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full bg-va-bg">
        <div className="text-center max-w-md">
          <div className="text-5xl mb-4">🧱</div>
          <h2 className="text-xl font-semibold text-va-text mb-2">No Capability Stacks</h2>
          <p className="text-va-text-muted text-sm mb-4">
            This project doesn't have any capability stacks configured yet.
            Add capability definitions to visualize your technology layers.
          </p>
          <div className="text-xs text-va-text-muted">
            <code>capability/*.json</code>
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
          <h2 className="text-lg font-semibold text-va-text">Capability Stack</h2>

          {/* Capability selector */}
          {capabilities.length > 1 && (
            <select
              value={selectedId || ''}
              onChange={(e) => handleCapabilitySelect(e.target.value)}
              className="bg-va-panel border border-va-border rounded px-2 py-1 text-sm text-va-text"
            >
              {capabilities.map((cap) => (
                <option key={cap.id} value={cap.id}>
                  {cap.name}
                  {cap.domain && ` (${cap.domain})`}
                </option>
              ))}
            </select>
          )}
        </div>

        <div className="flex items-center gap-2">
          {isLoading && (
            <div className="animate-spin w-4 h-4 border-2 border-va-accent border-t-transparent rounded-full" />
          )}
          <button
            onClick={loadCapabilities}
            disabled={isLoading}
            className="p-1.5 rounded hover:bg-va-panel text-va-text-muted hover:text-va-text disabled:opacity-50 transition-colors"
            title="Refresh"
          >
            <RefreshIcon className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Capability content */}
      <div className="flex-1 overflow-auto p-4">
        {selectedCapability ? (
          <CapabilityContent capability={selectedCapability} />
        ) : (
          <div className="flex items-center justify-center h-full text-va-text-muted">
            Select a capability to view
          </div>
        )}
      </div>
    </div>
  )
}

interface CapabilityContentProps {
  capability: CapabilityStack
}

function CapabilityContent({ capability }: CapabilityContentProps) {
  const title = (capability.metadata?.title as string) || (capability.metadata?.name as string) || 'Capability Stack'
  const description = capability.metadata?.description as string

  return (
    <div className="space-y-6">
      {/* Metadata header */}
      <div>
        <h3 className="text-xl font-semibold text-va-text">{title}</h3>
        {description && (
          <p className="text-va-text-muted mt-1">{description}</p>
        )}
      </div>

      {/* Layers */}
      {capability.layers && capability.layers.length > 0 && (
        <div>
          <h4 className="text-sm font-semibold text-va-text-muted uppercase tracking-wider mb-3">Layers</h4>
          <div className="space-y-2">
            {capability.layers.map((layer, idx) => (
              <LayerCard key={idx} layer={layer} />
            ))}
          </div>
        </div>
      )}

      {/* Categories */}
      {capability.categories && capability.categories.length > 0 && (
        <div>
          <h4 className="text-sm font-semibold text-va-text-muted uppercase tracking-wider mb-3">Categories</h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {capability.categories.map((category, idx) => (
              <CategoryCard key={idx} category={category} />
            ))}
          </div>
        </div>
      )}

      {/* Individual Capabilities */}
      {capability.capabilities && capability.capabilities.length > 0 && (
        <div>
          <h4 className="text-sm font-semibold text-va-text-muted uppercase tracking-wider mb-3">
            Capabilities ({capability.capabilities.length})
          </h4>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {capability.capabilities.map((cap, idx) => (
              <CapabilityCard key={idx} capability={cap} />
            ))}
          </div>
        </div>
      )}

      {/* Prism Integration */}
      {capability.prismIntegration && (
        <div>
          <h4 className="text-sm font-semibold text-va-text-muted uppercase tracking-wider mb-3">
            Prism Integration
          </h4>
          <div className="bg-va-panel border border-va-border rounded p-4">
            <pre className="text-xs text-va-text overflow-auto">
              {JSON.stringify(capability.prismIntegration, null, 2)}
            </pre>
          </div>
        </div>
      )}
    </div>
  )
}

function LayerCard({ layer }: { layer: Record<string, unknown> }) {
  const name = (layer.name as string) || (layer.id as string) || 'Layer'
  const description = layer.description as string
  const order = layer.order as number

  return (
    <div className="bg-va-panel border border-va-border rounded p-3 flex items-center gap-3">
      {order !== undefined && (
        <span className="w-8 h-8 flex items-center justify-center bg-va-accent/20 text-va-accent rounded-full text-sm font-medium">
          {order}
        </span>
      )}
      <div>
        <div className="font-medium text-va-text">{name}</div>
        {description && <div className="text-sm text-va-text-muted">{description}</div>}
      </div>
    </div>
  )
}

function CategoryCard({ category }: { category: Record<string, unknown> }) {
  const name = (category.name as string) || (category.id as string) || 'Category'
  const description = category.description as string
  const capCount = (category.capabilities as unknown[])?.length || 0

  return (
    <div className="bg-va-panel border border-va-border rounded p-3">
      <div className="font-medium text-va-text">{name}</div>
      {description && <div className="text-sm text-va-text-muted mt-1">{description}</div>}
      {capCount > 0 && (
        <div className="text-xs text-va-accent mt-2">{capCount} capabilities</div>
      )}
    </div>
  )
}

function CapabilityCard({ capability }: { capability: Record<string, unknown> }) {
  const name = (capability.name as string) || (capability.id as string) || 'Capability'
  const description = capability.description as string
  const status = capability.status as string
  const level = capability.level as string

  const getStatusColor = () => {
    switch (status) {
      case 'active':
      case 'implemented':
        return 'bg-va-success'
      case 'planned':
      case 'in_progress':
        return 'bg-va-warning'
      case 'deprecated':
        return 'bg-va-error'
      default:
        return 'bg-va-text-muted'
    }
  }

  return (
    <div className="bg-va-panel border border-va-border rounded p-3">
      <div className="flex items-start justify-between gap-2">
        <div className="font-medium text-va-text">{name}</div>
        {status && (
          <span className={`w-2 h-2 rounded-full ${getStatusColor()}`} title={status} />
        )}
      </div>
      {description && <div className="text-sm text-va-text-muted mt-1 line-clamp-2">{description}</div>}
      {level && (
        <div className="text-xs text-va-accent mt-2">Level: {level}</div>
      )}
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
