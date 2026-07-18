import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import type { Organization, OrganizationV2MOM, OrganizationCascade } from '../../types'

interface OrganizationViewProps {
  onClose?: () => void
}

type TabType = 'overview' | 'v2moms' | 'cascade' | 'settings'

export function OrganizationView({ onClose }: OrganizationViewProps) {
  const [organization, setOrganization] = useState<Organization | null>(null)
  const [v2moms, setV2MOMs] = useState<OrganizationV2MOM[]>([])
  const [cascade, setCascade] = useState<OrganizationCascade | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('overview')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isCreating, setIsCreating] = useState(false)
  const [newOrgName, setNewOrgName] = useState('')

  useEffect(() => {
    loadOrganization()
  }, [])

  const loadOrganization = async () => {
    try {
      setIsLoading(true)
      setError(null)

      const org = await api.getOrganization()
      setOrganization(org)

      if (org) {
        const [v2momList, cascadeData] = await Promise.all([
          api.listOrganizationV2MOMs(),
          api.getOrganizationCascade(),
        ])
        setV2MOMs(v2momList)
        setCascade(cascadeData)
      }
    } catch (err) {
      setError(`Failed to load organization: ${err}`)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateOrganization = async () => {
    if (!newOrgName.trim()) return

    try {
      setIsCreating(true)
      const org = await api.createOrganization(newOrgName.trim())
      setOrganization(org)
      setNewOrgName('')
    } catch (err) {
      setError(`Failed to create organization: ${err}`)
    } finally {
      setIsCreating(false)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin w-8 h-8 border-2 border-va-accent border-t-transparent rounded-full" />
      </div>
    )
  }

  // No organization - show create form
  if (!organization) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="bg-va-panel border border-va-border rounded-lg p-8 max-w-md text-center">
          <div className="text-4xl mb-4">Building</div>
          <h2 className="text-xl font-semibold text-va-text mb-2">Create Your Organization</h2>
          <p className="text-va-text-muted mb-6">
            Set up your organization to manage company-wide V2MOMs that cascade to individual projects.
          </p>

          <div className="space-y-4">
            <input
              type="text"
              value={newOrgName}
              onChange={(e) => setNewOrgName(e.target.value)}
              placeholder="Organization name"
              className="w-full px-4 py-2 bg-va-bg border border-va-border rounded text-va-text focus:outline-none focus:border-va-accent"
            />
            <button
              onClick={handleCreateOrganization}
              disabled={!newOrgName.trim() || isCreating}
              className="w-full px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isCreating ? 'Creating...' : 'Create Organization'}
            </button>
          </div>

          {error && (
            <p className="mt-4 text-va-error text-sm">{error}</p>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-4 border-b border-va-border">
        <div>
          <h1 className="text-xl font-semibold text-va-text">{organization.name}</h1>
          {organization.description && (
            <p className="text-sm text-va-text-muted">{organization.description}</p>
          )}
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="text-va-text-muted hover:text-va-text"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex border-b border-va-border px-6">
        {(['overview', 'v2moms', 'cascade', 'settings'] as TabType[]).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
              activeTab === tab
                ? 'text-va-accent border-va-accent'
                : 'text-va-text-muted border-transparent hover:text-va-text'
            }`}
          >
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6">
        {error && (
          <div className="bg-va-error/10 border border-va-error/30 rounded p-4 mb-4 text-va-error">
            {error}
          </div>
        )}

        {activeTab === 'overview' && (
          <OverviewTab v2moms={v2moms} cascade={cascade} />
        )}

        {activeTab === 'v2moms' && (
          <V2MOMsTab v2moms={v2moms} />
        )}

        {activeTab === 'cascade' && (
          <CascadeTab cascade={cascade} />
        )}

        {activeTab === 'settings' && (
          <SettingsTab organization={organization} onUpdate={loadOrganization} />
        )}
      </div>
    </div>
  )
}

// Overview Tab
function OverviewTab({
  v2moms,
  cascade,
}: {
  v2moms: OrganizationV2MOM[]
  cascade: OrganizationCascade | null
}) {
  const projectCount = cascade ? Object.keys(cascade.projectV2moms).length : 0

  return (
    <div className="space-y-6">
      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="bg-va-panel border border-va-border rounded-lg p-4">
          <div className="text-3xl font-bold text-va-accent">{v2moms.length}</div>
          <div className="text-sm text-va-text-muted">Org V2MOMs</div>
        </div>
        <div className="bg-va-panel border border-va-border rounded-lg p-4">
          <div className="text-3xl font-bold text-va-success">{projectCount}</div>
          <div className="text-sm text-va-text-muted">Projects</div>
        </div>
        <div className="bg-va-panel border border-va-border rounded-lg p-4">
          <div className="text-3xl font-bold text-va-warning">--</div>
          <div className="text-sm text-va-text-muted">Avg Alignment</div>
        </div>
      </div>

      {/* Primary V2MOM */}
      {v2moms.length > 0 && (
        <div className="bg-va-panel border border-va-border rounded-lg p-6">
          <h3 className="text-lg font-semibold text-va-text mb-4">Organization Strategy</h3>
          <V2MOMCard v2mom={v2moms[0]} />
        </div>
      )}

      {/* Quick links */}
      <div className="bg-va-panel border border-va-border rounded-lg p-4">
        <h3 className="text-sm font-semibold text-va-text-muted uppercase mb-3">Quick Actions</h3>
        <div className="space-y-2">
          <button className="w-full text-left px-3 py-2 rounded hover:bg-va-bg text-va-text text-sm">
            Create New V2MOM
          </button>
          <button className="w-full text-left px-3 py-2 rounded hover:bg-va-bg text-va-text text-sm">
            View Alignment Dashboard
          </button>
        </div>
      </div>
    </div>
  )
}

// V2MOM Card component
function V2MOMCard({ v2mom }: { v2mom: OrganizationV2MOM }) {
  return (
    <div className="space-y-4">
      {/* Vision */}
      <div>
        <h4 className="text-sm font-medium text-va-text-muted mb-1">Vision</h4>
        <p className="text-va-text">{v2mom.vision || 'No vision defined'}</p>
      </div>

      {/* Methods summary */}
      <div>
        <h4 className="text-sm font-medium text-va-text-muted mb-2">
          Methods ({v2mom.methods?.length || 0})
        </h4>
        <div className="space-y-2">
          {v2mom.methods?.slice(0, 3).map((method, i) => (
            <div key={method.id || i} className="flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-va-accent/20 text-va-accent text-xs flex items-center justify-center">
                {i + 1}
              </span>
              <span className="text-va-text text-sm">{method.name}</span>
            </div>
          ))}
          {(v2mom.methods?.length || 0) > 3 && (
            <p className="text-xs text-va-text-muted">
              +{(v2mom.methods?.length || 0) - 3} more methods
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

// V2MOMs Tab
function V2MOMsTab({ v2moms }: { v2moms: OrganizationV2MOM[] }) {
  if (v2moms.length === 0) {
    return (
      <div className="text-center py-12 text-va-text-muted">
        <p>No organization V2MOMs yet.</p>
        <p className="text-sm mt-2">
          Create V2MOM files in your organization's V2MOM directory.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {v2moms.map((v2mom) => (
        <div
          key={v2mom.id}
          className="bg-va-panel border border-va-border rounded-lg p-4"
        >
          <div className="flex items-start justify-between mb-3">
            <div>
              <h3 className="text-lg font-semibold text-va-text">{v2mom.name}</h3>
              {v2mom.fiscalYear && (
                <span className="text-sm text-va-text-muted">{v2mom.fiscalYear}</span>
              )}
            </div>
            {v2mom.owner && (
              <span className="text-xs text-va-text-muted">{v2mom.owner}</span>
            )}
          </div>
          <V2MOMCard v2mom={v2mom} />
        </div>
      ))}
    </div>
  )
}

// Cascade Tab
function CascadeTab({ cascade }: { cascade: OrganizationCascade | null }) {
  if (!cascade) {
    return (
      <div className="text-center py-12 text-va-text-muted">
        Loading cascade data...
      </div>
    )
  }

  const projectNames = Object.keys(cascade.projectV2moms)

  return (
    <div className="space-y-6">
      {/* Org V2MOMs */}
      <div>
        <h3 className="text-sm font-semibold text-va-text-muted uppercase mb-3">
          Organization Level
        </h3>
        {cascade.orgV2moms.length === 0 ? (
          <p className="text-va-text-muted text-sm">No organization V2MOMs</p>
        ) : (
          <div className="bg-va-panel border border-va-border rounded-lg p-4">
            {cascade.orgV2moms.map((v2mom) => (
              <div key={v2mom.id} className="flex items-center gap-2">
                <span className="text-va-accent">Org</span>
                <span className="text-va-text">{v2mom.name}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Arrow */}
      <div className="flex justify-center">
        <svg className="w-6 h-6 text-va-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
        </svg>
      </div>

      {/* Project V2MOMs */}
      <div>
        <h3 className="text-sm font-semibold text-va-text-muted uppercase mb-3">
          Project Level ({projectNames.length} projects)
        </h3>
        {projectNames.length === 0 ? (
          <p className="text-va-text-muted text-sm">No projects with V2MOMs</p>
        ) : (
          <div className="space-y-3">
            {projectNames.map((projectName) => (
              <div
                key={projectName}
                className="bg-va-panel border border-va-border rounded-lg p-4"
              >
                <h4 className="font-medium text-va-text mb-2">{projectName}</h4>
                <div className="space-y-1">
                  {cascade.projectV2moms[projectName].map((v2mom) => (
                    <div key={v2mom.id} className="text-sm text-va-text-muted flex items-center gap-2">
                      <span>--</span>
                      <span>{v2mom.name}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// Settings Tab
function SettingsTab({
  organization,
  onUpdate,
}: {
  organization: Organization
  onUpdate: () => void
}) {
  const [name, setName] = useState(organization.name)
  const [description, setDescription] = useState(organization.description || '')
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSave = async () => {
    try {
      setIsSaving(true)
      setError(null)
      await api.updateOrganization({ name, description })
      onUpdate()
    } catch (err) {
      setError(`Failed to save: ${err}`)
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="max-w-lg space-y-6">
      <div>
        <label className="block text-sm font-medium text-va-text mb-2">
          Organization Name
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-3 py-2 bg-va-bg border border-va-border rounded text-va-text focus:outline-none focus:border-va-accent"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-va-text mb-2">
          Description
        </label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          className="w-full px-3 py-2 bg-va-bg border border-va-border rounded text-va-text focus:outline-none focus:border-va-accent"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-va-text mb-2">
          V2MOM Directory
        </label>
        <input
          type="text"
          value={organization.v2momPath || '~/.visionstudio/v2moms/'}
          disabled
          className="w-full px-3 py-2 bg-va-bg border border-va-border rounded text-va-text-muted"
        />
        <p className="text-xs text-va-text-muted mt-1">
          Place your organization V2MOM JSON files in this directory.
        </p>
      </div>

      {error && (
        <p className="text-va-error text-sm">{error}</p>
      )}

      <button
        onClick={handleSave}
        disabled={isSaving}
        className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80 disabled:opacity-50"
      >
        {isSaving ? 'Saving...' : 'Save Changes'}
      </button>
    </div>
  )
}
