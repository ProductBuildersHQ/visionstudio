import { useState, useEffect } from 'react'
import type { Profile } from '../../types'
import type { SampleSummary } from '../../services/api'
import { api } from '../../services/api'
import { SamplePicker } from '../samples'

interface AddProjectModalProps {
  onClose: () => void
  onAdd: (name: string, path: string, profile: string, initialize: boolean) => Promise<void>
}

export function AddProjectModal({ onClose, onAdd }: AddProjectModalProps) {
  const [name, setName] = useState('')
  const [path, setPath] = useState('')
  const [profile, setProfile] = useState('startup')
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [initialize, setInitialize] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [selectedSample, setSelectedSample] = useState<SampleSummary | null>(null)
  const [showSamples, setShowSamples] = useState(false)

  // Load available profiles on mount
  useEffect(() => {
    console.log('[AddProjectModal] Loading profiles...')
    api.listProfiles()
      .then(data => {
        console.log('[AddProjectModal] Profiles loaded:', data)
        setProfiles(data)
      })
      .catch(err => console.error('[AddProjectModal] Failed to load profiles:', err))
  }, [])

  const handleBrowse = async () => {
    console.log('[AddProjectModal] Browse clicked')
    console.log('[AddProjectModal] electronAPI:', window.electronAPI)
    console.log('[AddProjectModal] electronAPI.dialog:', window.electronAPI?.dialog)
    try {
      // Use Electron dialog via preload
      const selectedPath = await window.electronAPI?.dialog?.selectDirectory()
      console.log('[AddProjectModal] Selected path:', selectedPath)
      if (selectedPath) {
        setPath(selectedPath)
        // Auto-fill name from directory name if empty
        if (!name) {
          const dirName = selectedPath.split('/').pop() || ''
          setName(dirName)
        }
      }
    } catch (err) {
      console.error('[AddProjectModal] Failed to open directory picker:', err)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (!name.trim()) {
      setError('Project name is required')
      return
    }
    if (!path.trim()) {
      setError('Directory path is required')
      return
    }

    setIsSubmitting(true)
    try {
      await onAdd(name.trim(), path.trim(), profile, initialize)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add project')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div
        className="bg-va-sidebar border border-va-border rounded-lg shadow-xl w-full max-w-md mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-va-border">
          <h2 className="text-lg font-semibold text-va-text">Add Project</h2>
          <button
            onClick={onClose}
            className="text-va-text-muted hover:text-va-text transition-colors"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit}>
          <div className="p-4 space-y-4">
            {/* Error message */}
            {error && (
              <div className="p-2 bg-va-error/20 border border-va-error/50 rounded text-sm text-va-error">
                {error}
              </div>
            )}

            {/* Project Name */}
            <div>
              <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider block mb-1">
                Project Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="my-project"
                className="w-full bg-va-panel border border-va-border rounded px-3 py-2 text-va-text placeholder:text-va-text-muted focus:outline-none focus:border-va-accent"
                autoFocus
              />
            </div>

            {/* Directory Path */}
            <div>
              <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider block mb-1">
                Specs Directory
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={path}
                  onChange={(e) => setPath(e.target.value)}
                  placeholder="/path/to/project/docs/specs"
                  className="flex-1 bg-va-panel border border-va-border rounded px-3 py-2 text-va-text font-mono text-sm placeholder:text-va-text-muted focus:outline-none focus:border-va-accent"
                />
                <button
                  type="button"
                  onClick={handleBrowse}
                  className="px-3 py-2 bg-va-panel hover:bg-va-border text-va-text rounded border border-va-border transition-colors text-sm"
                >
                  Browse
                </button>
              </div>
              <p className="mt-1 text-xs text-va-text-muted">
                Select the directory containing your spec files (e.g., project/docs/specs)
              </p>
            </div>

            {/* Profile */}
            <div>
              <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider block mb-1">
                Workflow Profile
              </label>
              {profiles.length === 0 ? (
                <div className="w-full bg-va-panel border border-va-border rounded px-3 py-2 text-va-text-muted text-sm">
                  Loading profiles...
                </div>
              ) : (
                <select
                  value={profile}
                  onChange={(e) => setProfile(e.target.value)}
                  className="w-full bg-va-panel border border-va-border rounded px-3 py-2 text-va-text focus:outline-none focus:border-va-accent appearance-none cursor-pointer"
                  style={{ backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%236b7280'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E")`, backgroundRepeat: 'no-repeat', backgroundPosition: 'right 0.5rem center', backgroundSize: '1.5em 1.5em', paddingRight: '2.5rem' }}
                >
                  {profiles.map((p) => (
                    <option key={p.name} value={p.name} className="bg-va-panel text-va-text">
                      {p.name} - {p.description}
                    </option>
                  ))}
                </select>
              )}
              {profiles.length > 0 && (
                <p className="mt-1 text-xs text-va-text-muted">
                  {profiles.find(p => p.name === profile)?.description}
                </p>
              )}
            </div>

            {/* Initialize checkbox */}
            <div className="flex items-start gap-3">
              <input
                type="checkbox"
                id="initialize"
                checked={initialize}
                onChange={(e) => {
                  setInitialize(e.target.checked)
                  if (!e.target.checked) {
                    setSelectedSample(null)
                    setShowSamples(false)
                  }
                }}
                className="mt-1 w-4 h-4 rounded border-va-border bg-va-panel text-va-accent focus:ring-va-accent"
              />
              <div>
                <label htmlFor="initialize" className="text-sm text-va-text cursor-pointer">
                  Initialize project structure
                </label>
                <p className="text-xs text-va-text-muted">
                  Creates directories (source/, gtm/, technical/, eval/) and scaffolds spec files from profile templates
                </p>
              </div>
            </div>

            {/* Sample selection */}
            {initialize && (
              <div>
                <button
                  type="button"
                  onClick={() => setShowSamples(!showSamples)}
                  className="flex items-center gap-2 text-sm text-va-accent hover:text-va-accent/80 transition-colors"
                >
                  <span>{showSamples ? '▼' : '►'}</span>
                  <span>
                    {selectedSample
                      ? `Using sample: ${selectedSample.name}`
                      : 'Initialize from sample (optional)'}
                  </span>
                </button>

                {showSamples && (
                  <div className="mt-3">
                    <SamplePicker
                      selectedSampleId={selectedSample?.id || null}
                      onSelectSample={(sample) => {
                        setSelectedSample(sample)
                        if (sample) {
                          // Auto-fill name from sample if empty
                          if (!name) {
                            setName(sample.name.toLowerCase().replace(/\s+/g, '-'))
                          }
                        }
                      }}
                    />
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="px-4 py-3 border-t border-va-border flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-1.5 bg-va-panel hover:bg-va-border text-va-text rounded transition-colors text-sm"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-4 py-1.5 bg-va-accent hover:bg-va-accent/80 text-white rounded transition-colors text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSubmitting ? 'Adding...' : 'Add Project'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
