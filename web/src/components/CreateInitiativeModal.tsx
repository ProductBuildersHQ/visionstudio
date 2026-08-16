import { useMemo, useState } from 'react'
import type { APIProgram, SpecWorkflow } from '../api/types'
import { createInitiative } from '../api/client'
import { defaultWorkflowForType, requiredSpecTypes } from '../lib/workflow'

const INIT_TYPES = ['feature', 'maintenance', 'refactor', 'migration', 'compliance'] as const
const PRIORITIES = ['', 'critical', 'high', 'medium', 'low'] as const

const ID_PATTERN = /^[A-Za-z0-9_-]+$/

export function CreateInitiativeModal({
  workflows,
  programs,
  defaultProgramId,
  onClose,
  onCreated,
}: {
  workflows: SpecWorkflow[]
  programs: APIProgram[]
  defaultProgramId?: string
  onClose: () => void
  onCreated: (id: string) => void
}) {
  const [id, setId] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [initType, setInitType] = useState<string>('feature')
  const [priority, setPriority] = useState<string>('')
  const [programId, setProgramId] = useState<string>(defaultProgramId ?? '')
  const [workflowId, setWorkflowId] = useState<string>(defaultWorkflowForType('feature'))
  const [workflowTouched, setWorkflowTouched] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const sortedWorkflows = useMemo(
    () => [...workflows].sort((a, b) => a.id.localeCompare(b.id)),
    [workflows]
  )
  const selectedWorkflow = workflows.find((w) => w.id === workflowId)

  const handleTypeChange = (t: string) => {
    setInitType(t)
    // Follow the type's default workflow until the user picks one explicitly.
    if (!workflowTouched) {
      setWorkflowId(defaultWorkflowForType(t))
    }
  }

  const idValid = id === '' || ID_PATTERN.test(id)
  const canSubmit = id !== '' && ID_PATTERN.test(id) && title !== '' && workflowId !== '' && !submitting

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!canSubmit) return
    setSubmitting(true)
    setError(null)
    try {
      const resp = await createInitiative({
        id,
        title,
        description: description || undefined,
        initType,
        priority: priority || undefined,
        workflowId,
        programId: programId || undefined,
      })
      onCreated(resp.id ?? id)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setSubmitting(false)
    }
  }

  return (
    <div
      className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-gray-800 rounded-lg w-full max-w-lg max-h-[90vh] overflow-y-auto shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <form onSubmit={handleSubmit}>
          <div className="p-4 border-b border-gray-700 flex items-center justify-between">
            <h2 className="font-medium">New Initiative</h2>
            <button
              type="button"
              onClick={onClose}
              className="text-gray-400 hover:text-gray-200 text-lg leading-none"
              aria-label="Close"
            >
              ×
            </button>
          </div>

          <div className="p-4 space-y-4">
            <Field label="ID" required>
              <input
                type="text"
                value={id}
                onChange={(e) => setId(e.target.value.trim())}
                placeholder="INIT-MYPROJECT-001"
                className={`w-full bg-gray-900 border rounded px-3 py-2 text-sm font-mono ${
                  idValid ? 'border-gray-600' : 'border-red-500'
                }`}
              />
              {!idValid && (
                <p className="text-xs text-red-400 mt-1">
                  Letters, digits, '-' and '_' only
                </p>
              )}
            </Field>

            <Field label="Title" required>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm"
              />
            </Field>

            <Field label="Description">
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm"
              />
            </Field>

            <div className="grid grid-cols-2 gap-4">
              <Field label="Type">
                <select
                  value={initType}
                  onChange={(e) => handleTypeChange(e.target.value)}
                  className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm"
                >
                  {INIT_TYPES.map((t) => (
                    <option key={t} value={t}>{t}</option>
                  ))}
                </select>
              </Field>
              <Field label="Priority">
                <select
                  value={priority}
                  onChange={(e) => setPriority(e.target.value)}
                  className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm"
                >
                  {PRIORITIES.map((p) => (
                    <option key={p} value={p}>{p === '' ? '(none)' : p}</option>
                  ))}
                </select>
              </Field>
            </div>

            <Field label="Program">
              <select
                value={programId}
                onChange={(e) => setProgramId(e.target.value)}
                className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm"
              >
                <option value="">(standalone)</option>
                {programs.map((p) => (
                  <option key={p.id} value={p.id}>{p.name || p.id}</option>
                ))}
              </select>
            </Field>

            <Field label="Spec workflow" required>
              <select
                value={workflowId}
                onChange={(e) => {
                  setWorkflowId(e.target.value)
                  setWorkflowTouched(true)
                }}
                className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm"
              >
                {sortedWorkflows.map((w) => (
                  <option key={w.id} value={w.id}>{w.id}</option>
                ))}
              </select>
              {selectedWorkflow && (
                <div className="mt-2 bg-gray-900 border border-gray-700 rounded p-3 space-y-2">
                  {selectedWorkflow.description && (
                    <p className="text-xs text-gray-400">{selectedWorkflow.description}</p>
                  )}
                  <div className="flex items-center gap-1 flex-wrap">
                    {requiredSpecTypes(selectedWorkflow).map((spec, i, arr) => (
                      <span key={spec} className="flex items-center gap-1">
                        <span className="text-xs px-1.5 py-0.5 bg-purple-500/10 border border-purple-500/30 text-purple-300 rounded">
                          {spec}
                        </span>
                        {i < arr.length - 1 && <span className="text-gray-600 text-xs">→</span>}
                      </span>
                    ))}
                  </div>
                  {(selectedWorkflow.specsOptional?.length ?? 0) > 0 && (
                    <p className="text-[11px] text-gray-500">
                      Optional: {(selectedWorkflow.specsOptional ?? []).map((f) => f.replace(/\.md$/i, '')).join(', ')}
                    </p>
                  )}
                </div>
              )}
            </Field>

            {error && (
              <div className="text-sm text-red-400 bg-red-500/10 border border-red-500/30 rounded p-3">
                {error}
              </div>
            )}
          </div>

          <div className="p-4 border-t border-gray-700 flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm text-gray-300 bg-gray-700 rounded hover:bg-gray-600 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!canSubmit}
              className="px-4 py-2 text-sm text-white bg-purple-600 rounded hover:bg-purple-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? 'Creating…' : 'Create Initiative'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function Field({
  label,
  required,
  children,
}: {
  label: string
  required?: boolean
  children: React.ReactNode
}) {
  return (
    <label className="block">
      <span className="text-sm text-gray-400">
        {label}
        {required && <span className="text-purple-400 ml-0.5">*</span>}
      </span>
      <div className="mt-1">{children}</div>
    </label>
  )
}
