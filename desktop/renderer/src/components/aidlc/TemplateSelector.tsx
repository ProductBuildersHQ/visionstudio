import { useState, useEffect, useCallback } from 'react'
import { api, AIDLCPhase, AIDLCTemplate, AIDLCDocument } from '../../services/api'

interface TemplateSelectorProps {
  projectName: string
  isOpen: boolean
  onClose: () => void
  onDocumentCreated?: (document: AIDLCDocument) => void
  filterPhase?: AIDLCPhase
}

const phaseLabels: Record<AIDLCPhase, string> = {
  inception: 'Inception',
  construction: 'Construction',
  operations: 'Operations',
}

const phaseColors: Record<AIDLCPhase, string> = {
  inception: 'bg-blue-900/50 text-blue-400 border-blue-500',
  construction: 'bg-yellow-900/50 text-yellow-400 border-yellow-500',
  operations: 'bg-green-900/50 text-green-400 border-green-500',
}

export function TemplateSelector({
  projectName,
  isOpen,
  onClose,
  onDocumentCreated,
  filterPhase,
}: TemplateSelectorProps) {
  const [templates, setTemplates] = useState<AIDLCTemplate[]>([])
  const [selectedTemplate, setSelectedTemplate] = useState<AIDLCTemplate | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isCreating, setIsCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [formData, setFormData] = useState({
    title: '',
    author: '',
    description: '',
  })

  const loadTemplates = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const allTemplates = await api.listAIDLCTemplates()
      const filtered = filterPhase
        ? allTemplates.filter((t) => t.phase === filterPhase)
        : allTemplates
      setTemplates(filtered)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load templates')
    } finally {
      setIsLoading(false)
    }
  }, [filterPhase])

  useEffect(() => {
    if (isOpen) {
      loadTemplates()
    }
  }, [isOpen, loadTemplates])

  const handleCreate = async () => {
    if (!selectedTemplate) return

    setIsCreating(true)
    setError(null)

    try {
      const document = await api.createAIDLCDocument(projectName, selectedTemplate.doc_type, {
        title: formData.title || selectedTemplate.name,
        author: formData.author || undefined,
        description: formData.description || undefined,
      })
      onDocumentCreated?.(document)
      handleClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create document')
    } finally {
      setIsCreating(false)
    }
  }

  const handleClose = () => {
    setSelectedTemplate(null)
    setFormData({ title: '', author: '', description: '' })
    setError(null)
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-va-bg border border-va-border rounded-lg w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-va-border">
          <h2 className="text-lg font-semibold text-va-text">
            {selectedTemplate ? 'Create Document' : 'Select Template'}
          </h2>
          <button
            onClick={handleClose}
            className="text-va-text-muted hover:text-va-text text-xl"
          >
            &times;
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-4">
          {isLoading ? (
            <div className="flex items-center justify-center h-32">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-va-accent"></div>
            </div>
          ) : error && !selectedTemplate ? (
            <div className="p-4 bg-red-900/20 border border-red-500 rounded-lg">
              <p className="text-red-400">{error}</p>
              <button
                onClick={loadTemplates}
                className="mt-2 px-3 py-1 bg-red-900/50 hover:bg-red-900/70 rounded text-sm"
              >
                Retry
              </button>
            </div>
          ) : selectedTemplate ? (
            /* Template detail/create form */
            <div className="space-y-4">
              <button
                onClick={() => setSelectedTemplate(null)}
                className="text-va-text-muted hover:text-va-text text-sm"
              >
                &larr; Back to templates
              </button>

              <div className="p-4 bg-va-panel rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <span
                    className={`px-2 py-0.5 text-xs rounded border ${phaseColors[selectedTemplate.phase]}`}
                  >
                    {phaseLabels[selectedTemplate.phase]}
                  </span>
                </div>
                <h3 className="text-lg font-semibold mb-1">{selectedTemplate.name}</h3>
                <p className="text-sm text-va-text-muted mb-4">{selectedTemplate.description}</p>

                {selectedTemplate.sections.length > 0 && (
                  <div className="mb-4">
                    <h4 className="text-sm font-semibold text-va-text mb-2">Sections</h4>
                    <div className="space-y-1">
                      {selectedTemplate.sections.map((section) => (
                        <div
                          key={section.id}
                          className="flex items-center gap-2 text-sm text-va-text-muted"
                        >
                          <span>{section.required ? '•' : '○'}</span>
                          <span>{section.title}</span>
                          {section.required && (
                            <span className="text-xs text-va-accent">(required)</span>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* Form */}
              <div className="space-y-3">
                <div>
                  <label className="block text-sm text-va-text-muted mb-1">
                    Title (optional)
                  </label>
                  <input
                    type="text"
                    value={formData.title}
                    onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                    placeholder={selectedTemplate.name}
                    className="w-full px-3 py-2 bg-va-bg border border-va-border rounded text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm text-va-text-muted mb-1">
                    Author (optional)
                  </label>
                  <input
                    type="text"
                    value={formData.author}
                    onChange={(e) => setFormData({ ...formData, author: e.target.value })}
                    placeholder="Your name"
                    className="w-full px-3 py-2 bg-va-bg border border-va-border rounded text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm text-va-text-muted mb-1">
                    Description (optional)
                  </label>
                  <textarea
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder="Brief description..."
                    className="w-full px-3 py-2 bg-va-bg border border-va-border rounded text-sm"
                    rows={2}
                  />
                </div>
              </div>

              {error && (
                <div className="p-3 bg-red-900/20 border border-red-500 rounded">
                  <p className="text-red-400 text-sm">{error}</p>
                </div>
              )}
            </div>
          ) : (
            /* Template list */
            <div className="space-y-4">
              {(['inception', 'construction', 'operations'] as AIDLCPhase[])
                .filter((phase) => !filterPhase || phase === filterPhase)
                .map((phase) => {
                  const phaseTemplates = templates.filter((t) => t.phase === phase)
                  if (phaseTemplates.length === 0) return null

                  return (
                    <div key={phase}>
                      <h3 className="text-sm font-semibold text-va-text-muted mb-2 capitalize">
                        {phase} Phase
                      </h3>
                      <div className="grid gap-2">
                        {phaseTemplates.map((template) => (
                          <button
                            key={template.doc_type}
                            onClick={() => setSelectedTemplate(template)}
                            className="text-left p-3 bg-va-panel hover:bg-va-border rounded-lg transition"
                          >
                            <div className="flex items-center gap-2 mb-1">
                              <span className="font-medium">{template.name}</span>
                              <span
                                className={`px-2 py-0.5 text-xs rounded border ${phaseColors[template.phase]}`}
                              >
                                {phaseLabels[template.phase]}
                              </span>
                            </div>
                            <p className="text-sm text-va-text-muted line-clamp-2">
                              {template.description}
                            </p>
                          </button>
                        ))}
                      </div>
                    </div>
                  )
                })}
            </div>
          )}
        </div>

        {/* Footer */}
        {selectedTemplate && (
          <div className="flex justify-end gap-2 px-4 py-3 border-t border-va-border">
            <button
              onClick={handleClose}
              className="px-4 py-2 bg-va-border hover:bg-va-border/80 rounded text-sm"
            >
              Cancel
            </button>
            <button
              onClick={handleCreate}
              disabled={isCreating}
              className="px-4 py-2 bg-va-accent hover:bg-va-accent/80 text-white rounded text-sm flex items-center gap-2"
            >
              {isCreating && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              )}
              {isCreating ? 'Creating...' : 'Create Document'}
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

export default TemplateSelector
