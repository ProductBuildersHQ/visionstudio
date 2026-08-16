import { useEffect, useMemo, useState } from 'react'
import { getWorkflowSpecDetail, type WorkflowSpecDetail } from '../api/client'

type DetailTab = 'template' | 'rubric'

/** Loose view of a structured-evaluation RubricSet — only what we render. */
interface RubricView {
  id?: string
  name?: string
  version?: string
  description?: string
  evaluationType?: string
  categories?: {
    id?: string
    name?: string
    description?: string
    weight?: number
    required?: boolean
    scale?: {
      type?: string
      options?: { value?: string; label?: string; criteria?: string[] }[]
    }
  }[]
}

export function SpecTypeDetailModal({
  workflowId,
  specType,
  initialTab,
  onClose,
}: {
  workflowId: string
  specType: string
  initialTab?: DetailTab
  onClose: () => void
}) {
  const [tab, setTab] = useState<DetailTab>(initialTab ?? 'template')
  const [detail, setDetail] = useState<WorkflowSpecDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(null)
    getWorkflowSpecDetail(workflowId, specType)
      .then(setDetail)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false))
  }, [workflowId, specType])

  const rubric = useMemo<RubricView | null>(() => {
    if (!detail?.rubricJson) return null
    try {
      return JSON.parse(detail.rubricJson) as RubricView
    } catch {
      return null
    }
  }, [detail?.rubricJson])

  // Auto-fall to the tab that has content when the preferred one is empty.
  useEffect(() => {
    if (!detail) return
    if (tab === 'template' && !detail.template && rubric) setTab('rubric')
    if (tab === 'rubric' && !rubric && detail.template) setTab('template')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detail, rubric])

  return (
    <div
      className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-gray-800 rounded-lg w-full max-w-3xl max-h-[90vh] flex flex-col shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-4 border-b border-gray-700 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h2 className="font-medium">{specType}</h2>
            <span className="text-xs px-2 py-0.5 bg-purple-500/10 border border-purple-500/30 text-purple-300 rounded">
              {workflowId}
            </span>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-200 text-lg leading-none"
            aria-label="Close"
          >
            ×
          </button>
        </div>

        <div className="border-b border-gray-700 px-4 flex gap-1">
          <ModalTab active={tab === 'template'} disabled={!detail?.template} onClick={() => setTab('template')}>
            Template
          </ModalTab>
          <ModalTab active={tab === 'rubric'} disabled={!rubric} onClick={() => setTab('rubric')}>
            Judge Rubric
          </ModalTab>
        </div>

        <div className="p-4 overflow-y-auto">
          {loading && <div className="text-gray-400 text-center py-8">Loading…</div>}
          {error && <div className="text-red-400 text-center py-8">{error}</div>}
          {!loading && !error && detail && (
            tab === 'template' ? (
              detail.template ? (
                <pre className="text-xs text-gray-300 bg-gray-900 rounded-lg p-4 whitespace-pre-wrap font-mono leading-relaxed">
                  {detail.template}
                </pre>
              ) : (
                <EmptyNote what="template" workflowId={workflowId} specType={specType} />
              )
            ) : rubric ? (
              <RubricDetail rubric={rubric} />
            ) : (
              <EmptyNote what="rubric" workflowId={workflowId} specType={specType} />
            )
          )}
        </div>
      </div>
    </div>
  )
}

function ModalTab({
  active,
  disabled,
  onClick,
  children,
}: {
  active: boolean
  disabled?: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`px-3 py-2 text-sm font-medium border-b-2 transition-colors ${
        active
          ? 'border-purple-500 text-purple-400'
          : disabled
            ? 'border-transparent text-gray-600 cursor-not-allowed'
            : 'border-transparent text-gray-400 hover:text-gray-200'
      }`}
    >
      {children}
    </button>
  )
}

function EmptyNote({ what, workflowId, specType }: { what: string; workflowId: string; specType: string }) {
  return (
    <div className="text-center py-8">
      <div className="text-gray-400 mb-1">
        No {what} defined for {specType} in {workflowId}
      </div>
      <div className="text-sm text-gray-500">
        Defined in the specification-workflow-spec catalog
      </div>
    </div>
  )
}

function RubricDetail({ rubric }: { rubric: RubricView }) {
  const [expanded, setExpanded] = useState<string | null>(null)

  return (
    <div className="space-y-4">
      <div>
        <div className="font-medium">{rubric.name ?? rubric.id}</div>
        {rubric.description && (
          <p className="text-sm text-gray-400 mt-1">{rubric.description}</p>
        )}
      </div>

      <div className="divide-y divide-gray-700 border border-gray-700 rounded-lg overflow-hidden">
        {(rubric.categories ?? []).map((c) => {
          const key = c.id ?? c.name ?? ''
          const isOpen = expanded === key
          return (
            <div key={key}>
              <button
                onClick={() => setExpanded(isOpen ? null : key)}
                className="w-full flex items-center justify-between p-3 hover:bg-gray-750 transition-colors text-left"
              >
                <div className="flex items-center gap-2 min-w-0">
                  <span className="text-gray-500 shrink-0">{isOpen ? '▼' : '▶'}</span>
                  <span className="text-sm font-medium truncate">{c.name ?? c.id}</span>
                  {c.required && (
                    <span className="text-[10px] px-1 py-0.5 bg-purple-500/10 border border-purple-500/30 text-purple-300 rounded shrink-0">
                      required
                    </span>
                  )}
                </div>
                {typeof c.weight === 'number' && (
                  <span className="text-xs text-gray-400 shrink-0 ml-2">
                    {Math.round(c.weight * 100)}%
                  </span>
                )}
              </button>
              {isOpen && (
                <div className="px-3 pb-3 space-y-2">
                  {c.description && (
                    <p className="text-xs text-gray-400">{c.description}</p>
                  )}
                  {(c.scale?.options ?? []).map((o) => (
                    <div key={o.value ?? o.label} className="bg-gray-900 rounded p-2">
                      <span
                        className={`text-xs font-medium ${
                          o.value === 'pass'
                            ? 'text-green-400'
                            : o.value === 'partial'
                              ? 'text-yellow-400'
                              : 'text-red-400'
                        }`}
                      >
                        {o.label ?? o.value}
                      </span>
                      <ul className="mt-1 space-y-0.5">
                        {(o.criteria ?? []).map((crit, i) => (
                          <li key={i} className="text-xs text-gray-300">
                            {crit}
                          </li>
                        ))}
                      </ul>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
