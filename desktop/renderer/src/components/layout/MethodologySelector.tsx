import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import type { Profile, ImplementationMethodologySummary, ProjectMethodologyConfig } from '../../types'

interface MethodologySelectorProps {
  projectName: string
  currentConfig: ProjectMethodologyConfig
  onClose: () => void
  onSave: (config: ProjectMethodologyConfig) => void
}

export function MethodologySelector({
  projectName,
  currentConfig,
  onClose,
  onSave,
}: MethodologySelectorProps) {
  const [requirementsProfiles, setRequirementsProfiles] = useState<Profile[]>([])
  const [implMethodologies, setImplMethodologies] = useState<ImplementationMethodologySummary[]>([])
  const [selectedRequirements, setSelectedRequirements] = useState(currentConfig.requirementsMethodology)
  const [selectedImplementation, setSelectedImplementation] = useState(currentConfig.implementationMethodology)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadMethodologies()
  }, [])

  const loadMethodologies = async () => {
    try {
      setIsLoading(true)
      setError(null)

      const [profiles, impls] = await Promise.all([
        api.listRequirementsMethodologies(),
        api.listImplementationMethodologies(),
      ])

      setRequirementsProfiles(profiles)
      setImplMethodologies(impls)
    } catch (err) {
      setError(`Failed to load methodologies: ${err}`)
    } finally {
      setIsLoading(false)
    }
  }

  const handleSave = async () => {
    try {
      setIsSaving(true)
      setError(null)

      const newConfig = await api.updateProjectMethodology(projectName, {
        requirementsMethodology: selectedRequirements,
        implementationMethodology: selectedImplementation,
      })

      onSave(newConfig)
    } catch (err) {
      setError(`Failed to save: ${err}`)
    } finally {
      setIsSaving(false)
    }
  }

  const hasChanges =
    selectedRequirements !== currentConfig.requirementsMethodology ||
    selectedImplementation !== currentConfig.implementationMethodology

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-va-panel border border-va-border rounded-lg shadow-xl w-full max-w-md">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-va-border">
          <h2 className="text-lg font-semibold text-va-text">Methodology Settings</h2>
          <button
            onClick={onClose}
            className="text-va-text-muted hover:text-va-text transition-colors"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-6">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin w-6 h-6 border-2 border-va-accent border-t-transparent rounded-full" />
            </div>
          ) : (
            <>
              {/* Requirements Methodology */}
              <div>
                <label className="block text-sm font-medium text-va-text mb-2">
                  Requirements Methodology
                </label>
                <p className="text-xs text-va-text-muted mb-2">
                  Defines WHAT to build - the spec structure and workflow
                </p>
                <select
                  value={selectedRequirements}
                  onChange={(e) => setSelectedRequirements(e.target.value)}
                  className="w-full bg-va-bg border border-va-border rounded px-3 py-2 text-sm text-va-text focus:outline-none focus:border-va-accent"
                >
                  {requirementsProfiles.map((profile) => (
                    <option key={profile.name} value={profile.name}>
                      {profile.name} - {profile.description}
                    </option>
                  ))}
                </select>
              </div>

              {/* Implementation Methodology */}
              <div>
                <label className="block text-sm font-medium text-va-text mb-2">
                  Implementation Methodology
                </label>
                <p className="text-xs text-va-text-muted mb-2">
                  Defines HOW to build - the implementation workflow
                </p>
                <select
                  value={selectedImplementation}
                  onChange={(e) => setSelectedImplementation(e.target.value)}
                  className="w-full bg-va-bg border border-va-border rounded px-3 py-2 text-sm text-va-text focus:outline-none focus:border-va-accent"
                >
                  {implMethodologies.map((methodology) => (
                    <option key={methodology.id} value={methodology.id}>
                      {methodology.name} - {methodology.description}
                    </option>
                  ))}
                </select>
              </div>

              {/* Implementation Details */}
              {selectedImplementation === 'aidlc' && (
                <div className="bg-va-bg border border-va-border rounded p-3">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-sm">AIDLC Phases:</span>
                  </div>
                  <div className="flex gap-2 text-xs">
                    <span className="px-2 py-1 bg-va-accent/20 text-va-accent rounded">Inception</span>
                    <span className="text-va-text-muted">→</span>
                    <span className="px-2 py-1 bg-va-warning/20 text-va-warning rounded">Construction</span>
                    <span className="text-va-text-muted">→</span>
                    <span className="px-2 py-1 bg-va-success/20 text-va-success rounded">Operations</span>
                  </div>
                </div>
              )}

              {selectedImplementation === 'speckit' && (
                <div className="bg-va-bg border border-va-border rounded p-3">
                  <p className="text-xs text-va-text-muted">
                    GitHub SpecKit integration for structured implementation specifications.
                  </p>
                </div>
              )}
            </>
          )}

          {error && (
            <div className="bg-va-error/10 border border-va-error/30 rounded p-3 text-sm text-va-error">
              {error}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-4 py-3 border-t border-va-border">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-va-text-muted hover:text-va-text transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={!hasChanges || isSaving}
            className="px-4 py-2 text-sm bg-va-accent text-white rounded hover:bg-va-accent/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isSaving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>
    </div>
  )
}
