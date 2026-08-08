import { useState, useEffect, useMemo } from 'react'
import { getSpecs, getExecution } from '../api/client'
import type { SpecsResponse, ExecutionResponse, JudgeResult, SpecWorkflow, APIInitiative } from '../api/compat'
import { LoadingState, ErrorState, EmptyState } from '../components'

export function SpecsPanel() {
  const [specs, setSpecs] = useState<SpecsResponse | null>(null)
  const [execution, setExecution] = useState<ExecutionResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedWorkflow, setSelectedWorkflow] = useState<string | null>(null)

  const reload = () => {
    setError(null)
    Promise.all([getSpecs(), getExecution()])
      .then(([s, e]) => {
        setSpecs(s)
        setExecution(e)
        if ((s.workflows?.length ?? 0) > 0 && !selectedWorkflow && s.workflows) {
          setSelectedWorkflow(s.workflows[0]?.id ?? null)
        }
      })
      .catch((err: Error) => setError(err.message))
  }

  useEffect(() => {
    reload()
  }, [])

  if (error) {
    return <ErrorState message={error} onRetry={reload} />
  }

  if (!specs || !execution) {
    return <LoadingState message="Loading specs data..." />
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Spec Compliance & Quality</h2>
      </div>

      {/* Workflows */}
      <WorkflowSection
        workflows={specs.workflows ?? []}
        selectedWorkflow={selectedWorkflow}
        onSelect={(id) => setSelectedWorkflow(id ?? null)}
        initiatives={execution.initiatives ?? []}
        judgeResults={specs.judgeResults ?? []}
      />

      {/* Judge Results by Initiative */}
      <JudgeResultsSection
        judgeResults={specs.judgeResults ?? []}
        initiatives={execution.initiatives ?? []}
        workflow={(specs.workflows ?? []).find((w) => w.id === selectedWorkflow)}
      />
    </div>
  )
}

function WorkflowSection({
  workflows,
  selectedWorkflow,
  onSelect,
  initiatives,
  judgeResults,
}: {
  workflows: SpecWorkflow[]
  selectedWorkflow: string | null
  onSelect: (id: string) => void
  initiatives: APIInitiative[]
  judgeResults: JudgeResult[]
}) {
  if (workflows.length === 0) {
    return (
      <EmptyState
        title="No workflows defined"
        description="Create a spec workflow to track documentation requirements."
        hint="prismctl spec workflow create"
      />
    )
  }

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-medium text-gray-400">Spec Workflows</h3>
      <div className="grid grid-cols-2 gap-4">
        {workflows.map((wf) => (
          <WorkflowCard
            key={wf.id}
            workflow={wf}
            selected={selectedWorkflow === wf.id}
            onClick={() => onSelect(wf.id ?? '')}
            initiatives={initiatives}
            judgeResults={judgeResults}
          />
        ))}
      </div>
    </div>
  )
}

function WorkflowCard({
  workflow,
  selected,
  onClick,
  initiatives,
  judgeResults,
}: {
  workflow: SpecWorkflow
  selected: boolean
  onClick: () => void
  initiatives: APIInitiative[]
  judgeResults: JudgeResult[]
}) {
  const requiredSpecs = workflow.specsRequired ?? []
  const optionalSpecs = workflow.specsOptional ?? []

  const compliance = useMemo(() => {
    if (requiredSpecs.length === 0 || initiatives.length === 0) {
      return { compliant: 0, partial: 0, missing: 0, total: initiatives.length }
    }

    let compliant = 0
    let partial = 0
    let missing = 0

    for (const init of initiatives) {
      const initResults = judgeResults.filter((r) => r.initiativeId === init.id)
      const specsPresent = new Set(initResults.map((r) => specType(r.specPath ?? '')))
      const requiredPresent = requiredSpecs.filter((s) => specsPresent.has(s)).length

      if (requiredPresent === requiredSpecs.length) {
        compliant++
      } else if (requiredPresent > 0) {
        partial++
      } else {
        missing++
      }
    }

    return { compliant, partial, missing, total: initiatives.length }
  }, [requiredSpecs, initiatives, judgeResults])

  const complianceRate = compliance.total > 0 ? compliance.compliant / compliance.total : 0

  return (
    <button
      onClick={onClick}
      className={`w-full text-left bg-gray-800 rounded-lg p-4 border transition-colors ${
        selected ? 'border-blue-500' : 'border-transparent hover:border-gray-700'
      }`}
    >
      <div className="flex items-center justify-between mb-2">
        <h4 className="font-medium">{workflow.name}</h4>
        <span className={`text-sm font-medium ${
          complianceRate >= 0.8 ? 'text-green-400' :
          complianceRate >= 0.5 ? 'text-yellow-400' : 'text-red-400'
        }`}>
          {Math.round(complianceRate * 100)}%
        </span>
      </div>
      {workflow.description && (
        <p className="text-sm text-gray-400 mb-3">{workflow.description}</p>
      )}
      <div className="space-y-2 text-sm">
        {requiredSpecs.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-red-400 text-xs">Required</span>
            <span className="text-gray-300">{requiredSpecs.join(', ')}</span>
          </div>
        )}
        {optionalSpecs.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-gray-500 text-xs">Optional</span>
            <span className="text-gray-400">{optionalSpecs.join(', ')}</span>
          </div>
        )}
      </div>
      <div className="mt-3 flex gap-2 text-xs">
        <span className="text-green-400">{compliance.compliant} compliant</span>
        {compliance.partial > 0 && (
          <span className="text-yellow-400">{compliance.partial} partial</span>
        )}
        {compliance.missing > 0 && (
          <span className="text-red-400">{compliance.missing} missing</span>
        )}
      </div>
    </button>
  )
}

function JudgeResultsSection({
  judgeResults,
  initiatives,
  workflow,
}: {
  judgeResults: JudgeResult[]
  initiatives: APIInitiative[]
  workflow?: SpecWorkflow
}) {
  const resultsByInit = useMemo(() => {
    const map: Record<string, JudgeResult[]> = {}
    for (const r of judgeResults) {
      const initId = r.initiativeId ?? ''
      if (!map[initId]) map[initId] = []
      map[initId].push(r)
    }
    return map
  }, [judgeResults])

  const getScore = (r: JudgeResult): number => r.report?.intScore ?? 0

  const sorted = useMemo(() => {
    return Object.entries(resultsByInit).sort((a, b) => {
      const avgA = a[1].reduce((sum, r) => sum + getScore(r), 0) / a[1].length
      const avgB = b[1].reduce((sum, r) => sum + getScore(r), 0) / b[1].length
      return avgB - avgA
    })
  }, [resultsByInit])

  if (judgeResults.length === 0) {
    return (
      <EmptyState
        title="No judge results"
        description="Run the spec judge to evaluate documentation quality."
        hint="prismctl spec judge"
      />
    )
  }

  const requiredSpecs = new Set(workflow?.specsRequired ?? [])

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-medium text-gray-400">
        Judge Results by Initiative
      </h3>
      <div className="space-y-4">
        {sorted.map(([initId, results]) => {
          const init = initiatives.find((i) => i.id === initId)
          const avgScore = results.reduce((sum, r) => sum + getScore(r), 0) / results.length
          const specsPresent = new Set(results.map((r) => specType(r.specPath ?? '')))
          const missingRequired = [...requiredSpecs].filter((s) => !specsPresent.has(s))

          return (
            <div key={initId} className="bg-gray-800 rounded-lg p-4">
              <div className="flex items-center justify-between mb-3">
                <div>
                  <span className="font-mono text-sm text-gray-400">{initId}</span>
                  {init && (
                    <span className="text-gray-300 ml-2">{init.title}</span>
                  )}
                </div>
                <ScoreBadge score={avgScore} />
              </div>

              {missingRequired.length > 0 && (
                <div className="mb-3 text-xs text-red-400 flex items-center gap-2">
                  <span>Missing required:</span>
                  {missingRequired.map((s) => (
                    <span key={s} className="bg-red-500/20 px-1.5 py-0.5 rounded">{s}</span>
                  ))}
                </div>
              )}

              <div className="space-y-2">
                {results
                  .sort((a, b) => new Date(b.evaluatedAt ?? '').getTime() - new Date(a.evaluatedAt ?? '').getTime())
                  .map((r) => (
                    <JudgeResultRow key={r.id} result={r} isRequired={requiredSpecs.has(specType(r.specPath ?? ''))} />
                  ))}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function JudgeResultRow({ result, isRequired }: { result: JudgeResult; isRequired: boolean }) {
  const [expanded, setExpanded] = useState(false)
  const specPath = result.specPath ?? ''
  const specName = specPath.split('/').pop() ?? specPath
  const specTypeName = result.specType ?? specType(specPath)
  const score = result.report?.intScore ?? 0
  const model = result.report?.judge?.model
  const rationale = result.report?.summary
  const categories = result.report?.categories ?? []
  const findings = result.report?.findings ?? []
  const decision = result.report?.decision
  const nextSteps = result.report?.nextSteps
  const confidence = result.report?.confidence

  return (
    <div className="bg-gray-900 rounded">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between py-2 px-3 hover:bg-gray-850 transition-colors"
      >
        <div className="flex items-center gap-3">
          <span className="text-gray-500">{expanded ? '▼' : '▶'}</span>
          <span className={`text-xs px-1.5 py-0.5 rounded ${
            isRequired ? 'bg-red-500/20 text-red-400' : 'bg-gray-700 text-gray-400'
          }`}>
            {specTypeName}
          </span>
          <span className="text-sm font-medium">{specName}</span>
          {model && (
            <span className="text-xs text-gray-500">{model}</span>
          )}
        </div>
        <div className="flex items-center gap-3">
          {findings.length > 0 && (
            <span className="text-xs text-yellow-400">{findings.length} findings</span>
          )}
          <span className="text-xs text-gray-500">
            {result.evaluatedAt ? new Date(result.evaluatedAt).toLocaleDateString() : ''}
          </span>
          <DecisionBadge status={decision?.status} />
          <ScoreBadge score={score} />
        </div>
      </button>
      {expanded && (
        <div className="px-3 pb-3 pt-1 border-t border-gray-800 space-y-3">
          {/* Summary & Confidence */}
          <div className="flex items-start justify-between">
            {rationale && (
              <p className="text-sm text-gray-400 flex-1">{rationale}</p>
            )}
            {confidence !== undefined && confidence > 0 && (
              <span className="text-xs text-gray-500 ml-2">
                {Math.round(confidence * 100)}% confidence
              </span>
            )}
          </div>

          {/* Categories */}
          {categories.length > 0 && (
            <div className="space-y-2">
              <h5 className="text-xs font-medium text-gray-500 uppercase">Categories</h5>
              <div className="grid gap-2">
                {categories.map((cat, i) => (
                  <div key={i} className="bg-gray-800 rounded p-2">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium">{cat.category}</span>
                      <div className="flex items-center gap-2">
                        {cat.confidence !== undefined && cat.confidence > 0 && (
                          <span className="text-xs text-gray-500">{Math.round(cat.confidence * 100)}%</span>
                        )}
                        <ScoreBadge score={cat.intScore ?? 0} />
                      </div>
                    </div>
                    {cat.reasoning && (
                      <p className="text-xs text-gray-400">{cat.reasoning}</p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Findings */}
          {findings.length > 0 && (
            <div className="space-y-2">
              <h5 className="text-xs font-medium text-gray-500 uppercase">Findings</h5>
              <div className="space-y-2">
                {findings.map((f, i) => (
                  <div key={i} className="bg-gray-800 rounded p-2">
                    <div className="flex items-center gap-2 mb-1">
                      <SeverityBadge severity={f.severity ?? 'medium'} />
                      <span className="text-xs text-gray-500">{f.category}</span>
                    </div>
                    <p className="text-sm">{f.title}</p>
                    {f.description && f.description !== f.title && (
                      <p className="text-xs text-gray-400 mt-1">{f.description}</p>
                    )}
                    {f.recommendation && (
                      <p className="text-xs text-blue-400 mt-1">{f.recommendation}</p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Next Steps */}
          {nextSteps && ((nextSteps.immediate?.length ?? 0) > 0 || (nextSteps.recommended?.length ?? 0) > 0) && (
            <div className="space-y-2">
              <h5 className="text-xs font-medium text-gray-500 uppercase">Next Steps</h5>
              {(nextSteps.immediate?.length ?? 0) > 0 && (
                <div className="space-y-1">
                  <span className="text-xs text-red-400">Immediate:</span>
                  {nextSteps.immediate?.map((a, i) => (
                    <div key={i} className="text-sm text-gray-300 pl-2 border-l-2 border-red-500">
                      {a.action}
                    </div>
                  ))}
                </div>
              )}
              {(nextSteps.recommended?.length ?? 0) > 0 && (
                <div className="space-y-1">
                  <span className="text-xs text-blue-400">Recommended:</span>
                  {nextSteps.recommended?.map((a, i) => (
                    <div key={i} className="text-sm text-gray-300 pl-2 border-l-2 border-blue-500">
                      {a.action}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function DecisionBadge({ status }: { status?: string }) {
  if (!status) return null
  const colors: Record<string, string> = {
    pass: 'bg-green-500/20 text-green-400',
    conditional: 'bg-yellow-500/20 text-yellow-400',
    fail: 'bg-red-500/20 text-red-400',
    human_review: 'bg-purple-500/20 text-purple-400',
  }
  return (
    <span className={`px-1.5 py-0.5 rounded text-xs ${colors[status] ?? colors.conditional}`}>
      {status}
    </span>
  )
}

function SeverityBadge({ severity }: { severity: string }) {
  const colors: Record<string, string> = {
    critical: 'bg-red-600 text-white',
    high: 'bg-red-500 text-white',
    medium: 'bg-yellow-500 text-black',
    low: 'bg-blue-500 text-white',
    info: 'bg-gray-500 text-white',
  }
  return (
    <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${colors[severity] ?? colors.medium}`}>
      {severity}
    </span>
  )
}

function ScoreBadge({ score }: { score: number }) {
  const color =
    score >= 4 ? 'bg-green-500' : score >= 3 ? 'bg-yellow-500' : 'bg-red-500'
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium text-white ${color}`}>
      {score}/5
    </span>
  )
}

function specType(path: string): string {
  const name = path.split('/').pop()?.toLowerCase() ?? ''
  if (name.includes('prd')) return 'PRD'
  if (name.includes('trd')) return 'TRD'
  if (name.includes('plan')) return 'PLAN'
  if (name.includes('roadmap')) return 'ROADMAP'
  return name.replace('.md', '').toUpperCase()
}
