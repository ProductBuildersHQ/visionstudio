import { useState, useEffect, useCallback } from 'react'
import { extensionRegistry } from '../../extensions/registry'
import { marketplaceClient } from '../../extensions/marketplace'
import { LoadingState, ErrorState } from '../toolkit'
import type { RegisteredExtension, MarketplaceEntry } from '../../types/extension'

interface ExtensionManagerViewProps {
  onViewSelect?: (extensionId: string, viewId: string) => void
}

type Tab = 'installed' | 'marketplace'

export function ExtensionManagerView({ onViewSelect }: ExtensionManagerViewProps) {
  const [activeTab, setActiveTab] = useState<Tab>('installed')
  const [marketplaceEntries, setMarketplaceEntries] = useState<MarketplaceEntry[]>([])
  const [isLoadingMarketplace, setIsLoadingMarketplace] = useState(false)
  const [marketplaceError, setMarketplaceError] = useState<string | null>(null)
  const [refreshKey, setRefreshKey] = useState(0)

  const installed = extensionRegistry.listInstalled()
  const allRegistered = extensionRegistry.listAll()

  const refresh = useCallback(() => setRefreshKey(k => k + 1), [])

  async function loadMarketplace() {
    setIsLoadingMarketplace(true)
    setMarketplaceError(null)
    try {
      const entries = await marketplaceClient.listExtensions()
      setMarketplaceEntries(entries)
    } catch (err) {
      setMarketplaceError(err instanceof Error ? err.message : 'Failed to load marketplace')
    } finally {
      setIsLoadingMarketplace(false)
    }
  }

  useEffect(() => {
    if (activeTab === 'marketplace') {
      loadMarketplace()
    }
  }, [activeTab])

  const handleInstall = useCallback((extensionId: string) => {
    extensionRegistry.install(extensionId)
    refresh()
  }, [refresh])

  const handleUninstall = useCallback(async (extensionId: string) => {
    await extensionRegistry.uninstall(extensionId)
    refresh()
  }, [refresh])

  // Force re-render when refreshKey changes
  void refreshKey

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-3xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-xl font-semibold text-va-text">Extensions</h1>
          <div className="flex gap-1 bg-va-panel rounded-lg p-0.5 border border-va-border">
            <button
              onClick={() => setActiveTab('installed')}
              className={`px-3 py-1.5 text-xs rounded-md transition-colors ${
                activeTab === 'installed'
                  ? 'bg-va-accent text-white'
                  : 'text-va-text-muted hover:text-va-text'
              }`}
            >
              Installed ({installed.length})
            </button>
            <button
              onClick={() => setActiveTab('marketplace')}
              className={`px-3 py-1.5 text-xs rounded-md transition-colors ${
                activeTab === 'marketplace'
                  ? 'bg-va-accent text-white'
                  : 'text-va-text-muted hover:text-va-text'
              }`}
            >
              Marketplace
            </button>
          </div>
        </div>

        {activeTab === 'installed' ? (
          <InstalledList
            extensions={installed}
            onViewSelect={onViewSelect}
            onUninstall={handleUninstall}
          />
        ) : (
          <MarketplaceList
            entries={marketplaceEntries}
            registered={allRegistered}
            isLoading={isLoadingMarketplace}
            error={marketplaceError}
            onRetry={loadMarketplace}
            onInstall={handleInstall}
            onUninstall={handleUninstall}
          />
        )}
      </div>
    </div>
  )
}

function InstalledList({
  extensions,
  onViewSelect,
  onUninstall,
}: {
  extensions: RegisteredExtension[]
  onViewSelect?: (extensionId: string, viewId: string) => void
  onUninstall: (extensionId: string) => Promise<void>
}) {
  const builtIn = extensions.filter(e => e.isBuiltIn)
  const marketplace = extensions.filter(e => !e.isBuiltIn)

  return (
    <div className="space-y-6">
      {builtIn.length > 0 && (
        <div>
          <h2 className="text-xs font-semibold text-va-text-muted uppercase tracking-wider mb-3">
            Built-in
          </h2>
          <div className="space-y-2">
            {builtIn.map(ext => (
              <ExtensionCard
                key={ext.manifest.id}
                extension={ext}
                onViewSelect={onViewSelect}
              />
            ))}
          </div>
        </div>
      )}
      {marketplace.length > 0 && (
        <div>
          <h2 className="text-xs font-semibold text-va-text-muted uppercase tracking-wider mb-3">
            Marketplace
          </h2>
          <div className="space-y-2">
            {marketplace.map(ext => (
              <ExtensionCard
                key={ext.manifest.id}
                extension={ext}
                onViewSelect={onViewSelect}
                onUninstall={onUninstall}
              />
            ))}
          </div>
        </div>
      )}
      {extensions.length === 0 && (
        <div className="text-center py-12 text-va-text-muted text-sm">
          No extensions installed. Browse the marketplace to find extensions.
        </div>
      )}
    </div>
  )
}

function ExtensionCard({
  extension,
  onViewSelect,
  onUninstall,
}: {
  extension: RegisteredExtension
  onViewSelect?: (extensionId: string, viewId: string) => void
  onUninstall?: (extensionId: string) => Promise<void>
}) {
  const [confirming, setConfirming] = useState(false)
  const { manifest } = extension
  const views = manifest.contributes.views ?? []

  return (
    <div className="bg-va-panel rounded-lg border border-va-border p-4">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-sm text-va-text">{manifest.name}</span>
            <span className="text-[10px] px-1.5 py-0.5 bg-va-bg border border-va-border rounded font-mono text-va-text-muted">
              v{manifest.version}
            </span>
            {extension.isActive && (
              <span className="text-[10px] px-1.5 py-0.5 bg-va-success/20 text-va-success rounded">
                Active
              </span>
            )}
          </div>
          <p className="text-xs text-va-text-muted mt-1">{manifest.description}</p>
          <div className="flex items-center gap-3 mt-2">
            <span className="text-[10px] text-va-text-muted">{manifest.publisher}</span>
            <span className="text-[10px] text-va-text-muted">{manifest.id}</span>
          </div>
        </div>
        {!extension.isBuiltIn && onUninstall && (
          <div className="shrink-0 ml-3">
            {confirming ? (
              <div className="flex gap-1">
                <button
                  onClick={() => { onUninstall(manifest.id); setConfirming(false) }}
                  className="px-2 py-1 text-[10px] bg-va-error text-white rounded hover:bg-va-error/80"
                >
                  Confirm
                </button>
                <button
                  onClick={() => setConfirming(false)}
                  className="px-2 py-1 text-[10px] bg-va-border text-va-text rounded hover:bg-va-text-muted/20"
                >
                  Cancel
                </button>
              </div>
            ) : (
              <button
                onClick={() => setConfirming(true)}
                className="px-3 py-1.5 text-xs text-va-text-muted border border-va-border rounded hover:text-va-error hover:border-va-error transition-colors"
              >
                Uninstall
              </button>
            )}
          </div>
        )}
      </div>

      {views.length > 0 && onViewSelect && (
        <div className="flex flex-wrap gap-1.5 mt-3 pt-3 border-t border-va-border">
          {views.map(view => (
            <button
              key={view.id}
              onClick={() => onViewSelect(manifest.id, view.id)}
              className="px-2 py-1 text-[10px] bg-va-bg border border-va-border rounded hover:border-va-accent hover:text-va-accent transition-colors text-va-text-muted"
            >
              {view.icon && <span className="mr-1">{view.icon}</span>}
              {view.name}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

function MarketplaceList({
  entries,
  registered,
  isLoading,
  error,
  onRetry,
  onInstall,
  onUninstall,
}: {
  entries: MarketplaceEntry[]
  registered: RegisteredExtension[]
  isLoading: boolean
  error: string | null
  onRetry: () => void
  onInstall: (extensionId: string) => void
  onUninstall: (extensionId: string) => Promise<void>
}) {
  if (isLoading) return <LoadingState message="Loading marketplace..." />
  if (error) return <ErrorState message={error} hint="Check your network connection" onRetry={onRetry} />

  const registeredMap = new Map(registered.map(r => [r.manifest.id, r]))
  const available = entries.filter(e => {
    const reg = registeredMap.get(e.id)
    return !e.builtIn && (!reg || !reg.isInstalled)
  })
  const alreadyInstalled = entries.filter(e => {
    const reg = registeredMap.get(e.id)
    return e.builtIn || (reg && reg.isInstalled)
  })

  return (
    <div className="space-y-6">
      {available.length > 0 && (
        <div>
          <h2 className="text-xs font-semibold text-va-text-muted uppercase tracking-wider mb-3">
            Available
          </h2>
          <div className="space-y-2">
            {available.map(entry => (
              <MarketplaceEntryCard
                key={entry.id}
                entry={entry}
                isInstalled={false}
                canInstall={registeredMap.has(entry.id)}
                onInstall={onInstall}
              />
            ))}
          </div>
        </div>
      )}
      {alreadyInstalled.length > 0 && (
        <div>
          <h2 className="text-xs font-semibold text-va-text-muted uppercase tracking-wider mb-3">
            Already Installed
          </h2>
          <div className="space-y-2">
            {alreadyInstalled.map(entry => {
              const reg = registeredMap.get(entry.id)
              return (
                <MarketplaceEntryCard
                  key={entry.id}
                  entry={entry}
                  isInstalled
                  isBuiltIn={reg?.isBuiltIn ?? entry.builtIn ?? false}
                  onUninstall={onUninstall}
                />
              )
            })}
          </div>
        </div>
      )}
      {entries.length === 0 && (
        <div className="text-center py-12 text-va-text-muted text-sm">
          No extensions available in the marketplace.
        </div>
      )}
    </div>
  )
}

function MarketplaceEntryCard({
  entry,
  isInstalled,
  isBuiltIn,
  canInstall,
  onInstall,
  onUninstall,
}: {
  entry: MarketplaceEntry
  isInstalled: boolean
  isBuiltIn?: boolean
  canInstall?: boolean
  onInstall?: (extensionId: string) => void
  onUninstall?: (extensionId: string) => Promise<void>
}) {
  const [confirming, setConfirming] = useState(false)

  return (
    <div className="bg-va-panel rounded-lg border border-va-border p-4">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-sm text-va-text">{entry.name}</span>
            <span className="text-[10px] px-1.5 py-0.5 bg-va-bg border border-va-border rounded font-mono text-va-text-muted">
              v{entry.version}
            </span>
            {isInstalled && (
              <span className="text-[10px] px-1.5 py-0.5 bg-va-success/20 text-va-success rounded">
                Installed
              </span>
            )}
          </div>
          <p className="text-xs text-va-text-muted mt-1">{entry.description}</p>
          <div className="flex items-center gap-3 mt-2">
            <span className="text-[10px] text-va-text-muted">{entry.publisher}</span>
            {entry.tags && entry.tags.length > 0 && (
              <div className="flex gap-1">
                {entry.tags.slice(0, 3).map(tag => (
                  <span key={tag} className="text-[10px] px-1 py-0.5 bg-va-bg rounded text-va-text-muted">
                    {tag}
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>
        <div className="shrink-0 ml-3">
          {!isInstalled && canInstall && onInstall && (
            <button
              onClick={() => onInstall(entry.id)}
              className="px-3 py-1.5 text-xs bg-va-accent text-white rounded hover:bg-va-accent/80 transition-colors"
            >
              Install
            </button>
          )}
          {!isInstalled && !canInstall && (
            <span className="px-3 py-1.5 text-xs text-va-text-muted border border-va-border rounded cursor-default" title="Extension package not available locally">
              Not Available
            </span>
          )}
          {isInstalled && !isBuiltIn && onUninstall && (
            confirming ? (
              <div className="flex gap-1">
                <button
                  onClick={() => { onUninstall(entry.id); setConfirming(false) }}
                  className="px-2 py-1 text-[10px] bg-va-error text-white rounded hover:bg-va-error/80"
                >
                  Confirm
                </button>
                <button
                  onClick={() => setConfirming(false)}
                  className="px-2 py-1 text-[10px] bg-va-border text-va-text rounded hover:bg-va-text-muted/20"
                >
                  Cancel
                </button>
              </div>
            ) : (
              <button
                onClick={() => setConfirming(true)}
                className="px-3 py-1.5 text-xs text-va-text-muted border border-va-border rounded hover:text-va-error hover:border-va-error transition-colors"
              >
                Uninstall
              </button>
            )
          )}
        </div>
      </div>
    </div>
  )
}
