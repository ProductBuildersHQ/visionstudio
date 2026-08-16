import { useState, useMemo, useEffect } from 'react'
import { Link } from 'react-router-dom'
import MarkdownIt from 'markdown-it'
import type {
  ExecutionResponse,
  SpecsResponse,
  APIInitiative,
  APIPhase,
  APIRMI,
  APIRMIDependency,
  JudgeResult,
  SpecWorkflow,
  SpecFile,
} from '../api/types'
import { getSpecFiles } from '../api/client'
import { resolveWorkflow, requiredSpecTypes } from '../lib/workflow'
import { StatusBadge } from '../components/StatusBadge'
import { ProgressBar } from '../components/ProgressBar'
import { SpecTypeDetailModal } from '../components/SpecTypeDetailModal'

interface InitiativeDetailProps {
  initiative: APIInitiative
  execution: ExecutionResponse
  specs: SpecsResponse
  onBack: () => void
}

type DetailTab = 'definition' | 'execution'

export function InitiativeDetail({
  initiative,
  execution,
  specs,
  onBack,
}: InitiativeDetailProps) {
  const phases = execution.phases.filter((p) => p.initiativeId === initiative.id)
  const rmis = execution.rmis.filter((r) => r.initiativeId === initiative.id)
  const rmiDeps = (execution.rmiDependencies ?? []).filter((d) =>
    rmis.some((r) => r.id === d.sourceRmiId || r.id === d.targetRmiId)
  )
  const initDeps = (execution.initiativeDependencies ?? []).filter(
    (d) => d.sourceInitiativeId === initiative.id || d.targetInitiativeId === initiative.id
  )
  const judgeResults = (specs.judgeResults ?? []).filter((r) => r.initiativeId === initiative.id)

  const hasExecution = phases.length > 0 || rmis.length > 0
  const [activeTab, setActiveTab] = useState<DetailTab>(hasExecution ? 'execution' : 'definition')
  const [specFiles, setSpecFiles] = useState<SpecFile[]>([])
  const [specFilesLoading, setSpecFilesLoading] = useState(true)

  useEffect(() => {
    setSpecFilesLoading(true)
    getSpecFiles(initiative.id)
      .then((resp) => setSpecFiles(resp.files ?? []))
      .catch(() => setSpecFiles([]))
      .finally(() => setSpecFilesLoading(false))
  }, [initiative.id])

  const sortedPhases = useMemo(
    () => [...phases].sort((a, b) => a.sequenceNumber - b.sequenceNumber),
    [phases]
  )

  const workflow = useMemo(
    () => resolveWorkflow(initiative, specs.workflows ?? []),
    [initiative, specs.workflows]
  )
  const requiredTotal = workflow ? requiredSpecTypes(workflow).length : 0
  const requiredPresent = specFiles.filter((f) => f.role === 'required').length

  const repoStats = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rmis) {
      if (r.repositoryId) {
        const short = r.repositoryId.split('/').pop() ?? r.repositoryId
        counts[short] = (counts[short] ?? 0) + 1
      }
    }
    return Object.entries(counts)
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => b.count - a.count)
  }, [rmis])

  const statusDist = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rmis) {
      const s = r.status.toLowerCase()
      counts[s] = (counts[s] ?? 0) + 1
    }
    return Object.entries(counts).map(([name, value]) => ({ name, value }))
  }, [rmis])

  return (
    <div className="space-y-6">
      {/* Back Button + Header */}
      <div>
        <button
          onClick={onBack}
          className="text-sm text-gray-400 hover:text-gray-200 mb-4 flex items-center gap-1"
        >
          <span>←</span> Back to Overview
        </button>
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-semibold">{initiative.id}</h1>
              <StatusBadge status={initiative.status} />
              {workflow && (
                <span
                  className="text-xs px-2 py-0.5 bg-purple-500/10 border border-purple-500/30 text-purple-300 rounded"
                  title={`Spec workflow: ${workflow.name}${initiative.workflowId ? '' : ' (default for type)'}`}
                >
                  {workflow.id}
                </span>
              )}
            </div>
            <p className="text-gray-300 mt-1">{initiative.title}</p>
            {initiative.description && (
              <p className="text-gray-500 text-sm mt-2 max-w-2xl">{initiative.description}</p>
            )}
          </div>
          <div className="text-right">
            <div className="text-2xl font-semibold">{Math.round(initiative.progress * 100)}%</div>
            <div className="text-sm text-gray-400">complete</div>
          </div>
        </div>
      </div>

      {/* Summary Cards Row */}
      <div className="grid grid-cols-4 gap-4">
        <SummaryCard
          label="Definition"
          value={
            requiredTotal > 0
              ? `${requiredPresent} of ${requiredTotal} (${Math.round((requiredPresent / requiredTotal) * 100)}%)`
              : specFiles.length > 0
                ? `${specFiles.length} specs`
                : 'No specs'
          }
          color="purple"
        />
        <SummaryCard label="Phases" value={phases.length.toString()} color="blue" />
        <SummaryCard label="RMIs" value={rmis.length.toString()} color="blue" />
        <SummaryCard label="Repos" value={repoStats.length.toString()} color="gray" />
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-gray-700">
        <nav className="flex gap-4">
          <TabButton
            active={activeTab === 'definition'}
            onClick={() => setActiveTab('definition')}
            color="purple"
          >
            Definition Details
          </TabButton>
          <TabButton
            active={activeTab === 'execution'}
            onClick={() => setActiveTab('execution')}
            color="blue"
            badge={hasExecution ? undefined : 'empty'}
          >
            Execution Details
          </TabButton>
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'definition' ? (
        <DefinitionTab
          judgeResults={judgeResults}
          workflows={specs.workflows ?? []}
          initiative={initiative}
          execution={execution}
          initDeps={initDeps}
          specFiles={specFiles}
          specFilesLoading={specFilesLoading}
        />
      ) : (
        <ExecutionTab
          phases={sortedPhases}
          rmis={rmis}
          rmiDeps={rmiDeps}
          repoStats={repoStats}
          statusDist={statusDist}
          initDeps={initDeps}
          initiative={initiative}
          execution={execution}
        />
      )}
    </div>
  )
}

function SummaryCard({
  label,
  value,
  color,
}: {
  label: string
  value: string
  color: 'purple' | 'blue' | 'gray'
}) {
  const colorClass = {
    purple: 'text-purple-400',
    blue: 'text-blue-400',
    gray: 'text-gray-400',
  }[color]

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="text-sm text-gray-400">{label}</div>
      <div className={`text-xl font-semibold mt-1 ${colorClass}`}>{value}</div>
    </div>
  )
}

function TabButton({
  active,
  onClick,
  color,
  badge,
  children,
}: {
  active: boolean
  onClick: () => void
  color: 'purple' | 'blue'
  badge?: string
  children: React.ReactNode
}) {
  const activeColor = color === 'purple' ? 'border-purple-500 text-purple-400' : 'border-blue-500 text-blue-400'

  return (
    <button
      onClick={onClick}
      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
        active
          ? activeColor
          : 'border-transparent text-gray-400 hover:text-gray-200'
      }`}
    >
      {children}
      {badge && (
        <span className="ml-2 text-xs px-1.5 py-0.5 bg-gray-700 rounded text-gray-500">
          {badge}
        </span>
      )}
    </button>
  )
}

function DefinitionTab({
  judgeResults,
  workflows,
  initiative,
  execution,
  initDeps,
  specFiles,
  specFilesLoading,
}: {
  judgeResults: JudgeResult[]
  workflows: SpecWorkflow[]
  initiative: APIInitiative
  execution: ExecutionResponse
  initDeps: { sourceInitiativeId: string; targetInitiativeId: string; relationship: string }[]
  specFiles: SpecFile[]
  specFilesLoading: boolean
}) {
  return (
    <div className="space-y-6">
      {/* Workflow Diagram */}
      <WorkflowDiagram judgeResults={judgeResults} workflows={workflows} specFiles={specFiles} initiative={initiative} />

      {/* Spec Files Viewer */}
      <SpecFilesViewer
        specFiles={specFiles}
        loading={specFilesLoading}
        initiativeId={initiative.id}
        workflowId={resolveWorkflow(initiative, workflows)?.id}
      />

      {/* Judge Results Detail */}
      <JudgeResultsDetail judgeResults={judgeResults} />

      {/* Initiative Dependencies */}
      {initDeps.length > 0 && (
        <DependenciesSection
          initDeps={initDeps}
          initiative={initiative}
          execution={execution}
        />
      )}
    </div>
  )
}

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: true,
})

function SpecFilesViewer({
  specFiles,
  loading,
  initiativeId,
  workflowId,
}: {
  specFiles: SpecFile[]
  loading: boolean
  initiativeId: string
  workflowId?: string
}) {
  const [selectedSpec, setSelectedSpec] = useState<string | null>(null)
  const [detailTab, setDetailTab] = useState<'template' | 'rubric' | null>(null)

  // Files arrive from the API already sorted into workflow order
  // (required by sequence, then optional, then extras).
  const sortedSpecFiles = specFiles
  const selected = sortedSpecFiles.find((f) => f.specType === selectedSpec) ?? sortedSpecFiles[0]

  const renderedHTML = useMemo(() => {
    if (!selected?.content) return ''
    return md.render(selected.content)
  }, [selected?.content])

  if (loading) {
    return (
      <div className="bg-gray-800 rounded-lg p-6 text-center">
        <div className="text-gray-400">Loading spec files...</div>
      </div>
    )
  }

  if (specFiles.length === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-6 text-center">
        <div className="text-gray-400 mb-2">No spec files found on disk</div>
        <div className="text-sm text-gray-500">
          Add specs to docs/specs/initiatives/{'{INITIATIVE_ID}'}/
        </div>
      </div>
    )
  }

  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden">
      {/* Spec Type Tabs */}
      <div className="border-b border-gray-700 px-4 flex items-center justify-between">
        <div className="flex gap-1">
          {sortedSpecFiles.map((f) => (
            <button
              key={f.specType}
              onClick={() => setSelectedSpec(f.specType)}
              className={`px-3 py-2 text-sm font-medium border-b-2 transition-colors ${
                (selectedSpec ?? sortedSpecFiles[0]?.specType) === f.specType
                  ? 'border-purple-500 text-purple-400'
                  : 'border-transparent text-gray-400 hover:text-gray-200'
              }`}
            >
              {f.specType}
              {f.role === 'extra' && (
                <span className="ml-1.5 text-[10px] px-1 py-0.5 bg-gray-700 rounded text-gray-400 align-middle">
                  Extra
                </span>
              )}
            </button>
          ))}
        </div>
        <div className="flex items-center gap-2">
          {workflowId && (
            <>
              <button
                onClick={() => setDetailTab('template')}
                className="px-3 py-1 text-xs font-medium bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
              >
                Template
              </button>
              <button
                onClick={() => setDetailTab('rubric')}
                className="px-3 py-1 text-xs font-medium bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
              >
                Rubric
              </button>
            </>
          )}
          <Link
            to={`/initiative/${initiativeId}/spec/${selected.specType.toLowerCase()}`}
            className="px-3 py-1 text-xs font-medium bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
          >
            Open Full View
          </Link>
        </div>
      </div>

      {detailTab && workflowId && (
        <SpecTypeDetailModal
          workflowId={workflowId}
          specType={selected.specType}
          initialTab={detailTab}
          onClose={() => setDetailTab(null)}
        />
      )}

      {/* Spec Content */}
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <span className="text-xs text-gray-500 font-mono">{selected.path.split('/').slice(-3).join('/')}</span>
          {selected.modTime && (
            <span className="text-xs text-gray-500">
              Modified: {new Date(selected.modTime).toLocaleDateString()}
            </span>
          )}
        </div>
        <div
          className="spec-prose bg-gray-900 rounded-lg p-4 max-h-96 overflow-y-auto"
          dangerouslySetInnerHTML={{ __html: renderedHTML }}
        />
      </div>
    </div>
  )
}

function ExecutionTab({
  phases,
  rmis,
  rmiDeps,
  repoStats,
  statusDist,
  initDeps,
  initiative,
  execution,
}: {
  phases: APIPhase[]
  rmis: APIRMI[]
  rmiDeps: APIRMIDependency[]
  repoStats: { name: string; count: number }[]
  statusDist: { name: string; value: number }[]
  initDeps: { sourceInitiativeId: string; targetInitiativeId: string; relationship: string }[]
  initiative: APIInitiative
  execution: ExecutionResponse
}) {
  if (phases.length === 0 && rmis.length === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-8 text-center">
        <div className="text-gray-400 mb-2">No execution data yet</div>
        <div className="text-sm text-gray-500">
          Define phases and RMIs in ROADMAP.md to track execution progress
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Status Distribution */}
      {statusDist.length > 0 && (
        <div className="bg-gray-800 rounded-lg px-4 py-3">
          <div className="flex items-center gap-4">
            <h3 className="font-medium text-sm text-gray-300">RMI Status</h3>
            <div className="flex items-center gap-4">
              {statusDist.map((s) => (
                <div key={s.name} className="flex items-center gap-1.5 text-sm">
                  <span className="text-blue-400 font-semibold tabular-nums">{s.value}</span>
                  <span className="text-gray-500 capitalize">{s.name.replace('_', ' ')}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Repos */}
      {repoStats.length > 1 && (
        <div className="flex flex-wrap gap-2">
          {repoStats.map((r) => (
            <span
              key={r.name}
              className="text-xs px-2 py-1 bg-gray-800 border border-gray-700 rounded"
            >
              {r.name} <span className="text-gray-500">({r.count})</span>
            </span>
          ))}
        </div>
      )}

      {/* Initiative Dependencies */}
      {initDeps.length > 0 && (
        <DependenciesSection
          initDeps={initDeps}
          initiative={initiative}
          execution={execution}
        />
      )}

      {/* Phases */}
      <div className="space-y-4">
        <h2 className="text-lg font-medium">Phases</h2>
        {phases.map((phase) => {
          const phaseRmis = rmis
            .filter((r) => r.phaseId === phase.id)
            .sort((a, b) => a.sequenceNumber - b.sequenceNumber)
          const phaseDeps = rmiDeps.filter((d) =>
            phaseRmis.some((r) => r.id === d.sourceRmiId || r.id === d.targetRmiId)
          )

          return (
            <PhaseCard
              key={phase.id}
              phase={phase}
              rmis={phaseRmis}
              deps={phaseDeps}
              allRmis={rmis}
            />
          )
        })}
      </div>
    </div>
  )
}

function DependenciesSection({
  initDeps,
  initiative,
  execution,
}: {
  initDeps: { sourceInitiativeId: string; targetInitiativeId: string; relationship: string }[]
  initiative: APIInitiative
  execution: ExecutionResponse
}) {
  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <h3 className="font-medium mb-2">Initiative Dependencies</h3>
      <div className="flex flex-wrap gap-2">
        {initDeps.map((d, i) => {
          const isSource = d.sourceInitiativeId === initiative.id
          const otherId = isSource ? d.targetInitiativeId : d.sourceInitiativeId
          const other = execution.initiatives.find((init) => init.id === otherId)
          return (
            <span
              key={i}
              className="text-xs px-2 py-1 bg-gray-700 rounded flex items-center gap-1"
            >
              {isSource ? (
                <>
                  <span className="text-gray-400">requires</span>
                  <span className="font-mono">{otherId}</span>
                  {other && <span className="text-gray-500">({other.title})</span>}
                </>
              ) : (
                <>
                  <span className="font-mono">{otherId}</span>
                  <span className="text-gray-400">requires this</span>
                </>
              )}
            </span>
          )
        })}
      </div>
    </div>
  )
}

function WorkflowDiagram({
  judgeResults,
  workflows,
  specFiles,
  initiative,
}: {
  judgeResults: JudgeResult[]
  workflows: SpecWorkflow[]
  specFiles: SpecFile[]
  initiative: APIInitiative
}) {
  const resultsByType = useMemo(() => {
    const map: Record<string, JudgeResult> = {}
    for (const r of judgeResults) {
      const type = specType(r.specPath)
      const existing = map[type]
      if (!existing || new Date(r.evaluatedAt) > new Date(existing.evaluatedAt)) {
        map[type] = r
      }
    }
    return map
  }, [judgeResults])

  const specFilesByType = useMemo(() => {
    const map: Record<string, SpecFile> = {}
    for (const f of specFiles) {
      map[f.specType] = f
    }
    return map
  }, [specFiles])

  const workflow = useMemo(
    () => resolveWorkflow(initiative, workflows),
    [initiative, workflows]
  )
  const diagramSpecs = workflow ? requiredSpecTypes(workflow) : []
  const [detailSpec, setDetailSpec] = useState<string | null>(null)
  const getScore = (r: JudgeResult): number => r.report?.intScore ?? 0
  const avgScore =
    judgeResults.length > 0
      ? judgeResults.reduce((sum, r) => sum + getScore(r), 0) / judgeResults.length
      : 0

  if (!workflow || diagramSpecs.length === 0) {
    return null
  }

  return (
    <div className="bg-gray-800 rounded-lg p-4 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-purple-400">{workflow.name} Workflow</span>
          <span className="text-xs text-gray-500">{diagramSpecs.join(' → ')}</span>
          <span className="text-[10px] text-gray-600">click a document for its template &amp; rubric</span>
        </div>
        {judgeResults.length > 0 && (
          <span
            className={`text-lg font-semibold ${
              avgScore >= 4 ? 'text-green-400' : avgScore >= 3 ? 'text-yellow-400' : 'text-red-400'
            }`}
          >
            {avgScore.toFixed(1)} avg
          </span>
        )}
      </div>

      {/* Workflow Diagram */}
      <div className="flex items-center justify-center gap-2 py-4 flex-wrap">
        {diagramSpecs.map((spec, i) => {
          const result = resultsByType[spec]
          const hasJudgeResult = !!result
          const hasSpecFile = !!specFilesByType[spec]
          const score = result ? getScore(result) : 0

          // Determine state: evaluated (with score), present (file exists), or missing
          let stateClass: string
          let tooltip: string
          let bottomText: string | null = null

          if (hasJudgeResult) {
            stateClass = score >= 4
              ? 'bg-green-500/20 border-green-500 text-green-300'
              : score >= 3
              ? 'bg-yellow-500/20 border-yellow-500 text-yellow-300'
              : 'bg-red-500/20 border-red-500 text-red-300'
            tooltip = `Score: ${score}/5`
            bottomText = `${score}/5`
          } else if (hasSpecFile) {
            stateClass = 'bg-blue-500/20 border-blue-500 text-blue-300'
            tooltip = 'Spec exists (not evaluated)'
            bottomText = '✓'
          } else {
            stateClass = 'bg-gray-700 border-gray-600 text-gray-400'
            tooltip = 'Not created'
          }

          return (
            <div key={spec} className="flex items-center">
              <button
                onClick={() => setDetailSpec(spec)}
                className={`px-4 py-3 rounded-lg text-sm font-medium border-2 transition-all cursor-pointer hover:brightness-125 ${stateClass}`}
                title={`${tooltip} — click for template & rubric`}
              >
                <div className="text-center">
                  <div>{spec}</div>
                  {bottomText && <div className="text-xs opacity-70 mt-1">{bottomText}</div>}
                </div>
              </button>
              {i < diagramSpecs.length - 1 && (
                <span className="text-gray-500 px-2 text-lg">→</span>
              )}
            </div>
          )
        })}
      </div>

      {detailSpec && (
        <SpecTypeDetailModal
          workflowId={workflow.id}
          specType={detailSpec}
          onClose={() => setDetailSpec(null)}
        />
      )}
    </div>
  )
}

function JudgeResultsDetail({ judgeResults }: { judgeResults: JudgeResult[] }) {
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const getScore = (r: JudgeResult): number => r.report?.intScore ?? 0

  if (judgeResults.length === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-6 text-center">
        <div className="text-gray-400 mb-2">No spec evaluations yet</div>
        <div className="text-sm text-gray-500">
          Run LLM-as-a-Judge on your specs to see quality scores
        </div>
      </div>
    )
  }

  const sorted = [...judgeResults].sort(
    (a, b) => new Date(b.evaluatedAt).getTime() - new Date(a.evaluatedAt).getTime()
  )

  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden">
      <div className="p-4 border-b border-gray-700">
        <h3 className="font-medium">LLM-as-a-Judge Results</h3>
      </div>
      <div className="divide-y divide-gray-700">
        {sorted.map((r) => {
          const score = getScore(r)
          const model = r.report?.judge?.model
          const rationale = r.report?.summary
          const pass = r.report?.pass
          return (
            <div key={r.id}>
              <button
                onClick={() => setExpandedId(expandedId === r.id ? null : r.id)}
                className="w-full flex items-center justify-between p-4 hover:bg-gray-750 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <span className="text-gray-500">{expandedId === r.id ? '▼' : '▶'}</span>
                  <span className="text-xs px-2 py-0.5 bg-purple-500/20 text-purple-300 rounded">
                    {r.specType ?? specType(r.specPath)}
                  </span>
                  <span className="text-sm text-gray-300">{r.specPath.split('/').pop()}</span>
                  {model && <span className="text-xs text-gray-500">{model}</span>}
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-500">
                    {new Date(r.evaluatedAt).toLocaleDateString()}
                  </span>
                  <span
                    className={`px-2 py-1 rounded text-sm font-medium ${
                      pass !== undefined
                        ? pass
                          ? 'bg-green-500/30 text-green-300'
                          : 'bg-red-500/30 text-red-300'
                        : score >= 4
                        ? 'bg-green-500/30 text-green-300'
                        : score >= 3
                        ? 'bg-yellow-500/30 text-yellow-300'
                        : 'bg-red-500/30 text-red-300'
                    }`}
                  >
                    {score}/5
                  </span>
                </div>
              </button>
              {expandedId === r.id && rationale && (
                <div className="px-4 pb-4 pt-0 ml-8 text-sm text-gray-400 whitespace-pre-wrap">
                  {rationale}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

function PhaseCard({
  phase,
  rmis,
  deps,
  allRmis,
}: {
  phase: APIPhase
  rmis: APIRMI[]
  deps: APIRMIDependency[]
  allRmis: APIRMI[]
}) {
  const [expanded, setExpanded] = useState(true)

  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between p-4 hover:bg-gray-750 transition-colors"
      >
        <div className="flex items-center gap-3">
          <span className="text-gray-500">{expanded ? '▼' : '▶'}</span>
          <h4 className="font-medium">{phase.title}</h4>
          <span className="text-sm text-gray-500">({rmis.length} RMIs)</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-400">{Math.round(phase.progress * 100)}%</span>
          <ProgressBar
            progress={phase.progress}
            cancelledProgress={phase.cancelledProgress}
            className="w-24"
            size="sm"
          />
        </div>
      </button>
      {expanded && (
        <div className="border-t border-gray-700 p-4 space-y-2">
          {rmis.map((rmi) => (
            <RMIRow key={rmi.id} rmi={rmi} deps={deps} allRmis={allRmis} />
          ))}
        </div>
      )}
    </div>
  )
}

function RMIRow({
  rmi,
  deps,
  allRmis,
}: {
  rmi: APIRMI
  deps: APIRMIDependency[]
  allRmis: APIRMI[]
}) {
  const myDeps = deps.filter((d) => d.sourceRmiId === rmi.id)
  const depTitles = myDeps
    .map((d) => allRmis.find((r) => r.id === d.targetRmiId)?.id)
    .filter(Boolean)

  return (
    <div className="flex items-center justify-between py-2 px-3 bg-gray-900 rounded hover:bg-gray-850 transition-colors">
      <div className="flex items-center gap-3 min-w-0">
        <span className="text-lg" title={rmi.type ?? 'item'}>
          {typeIcon(rmi.type)}
        </span>
        <span className="text-xs font-mono text-gray-500 flex-shrink-0">{rmi.id}</span>
        <span className="text-sm truncate">{rmi.title}</span>
        {depTitles.length > 0 && (
          <span
            className="text-xs text-gray-500 flex-shrink-0"
            title={`Requires: ${depTitles.join(', ')}`}
          >
            → {depTitles.length}
          </span>
        )}
      </div>
      <div className="flex items-center gap-3 flex-shrink-0">
        {rmi.claimedBy && (
          <span className="text-xs text-gray-500" title={`Claimed: ${rmi.claimedAt}`}>
            {rmi.claimedBy}
          </span>
        )}
        {rmi.completedAt && (
          <span className="text-xs text-gray-500">
            {new Date(rmi.completedAt).toLocaleDateString()}
          </span>
        )}
        <StatusBadge status={rmi.status} size="sm" />
      </div>
    </div>
  )
}

function typeIcon(itemType?: string): string {
  switch (itemType?.toLowerCase()) {
    case 'capability':
      return '★'
    case 'refactor':
      return '↺'
    case 'quality':
      return '✓'
    case 'fix':
      return '⚠'
    case 'chore':
      return '⚙'
    case 'spike':
      return '⚡'
    default:
      return '•'
  }
}

function specType(path: string): string {
  const name = path.split('/').pop()?.toLowerCase() ?? ''
  if (name.includes('prd')) return 'PRD'
  if (name.includes('trd')) return 'TRD'
  if (name.includes('plan')) return 'PLAN'
  if (name.includes('roadmap')) return 'ROADMAP'
  return name.replace('.md', '').toUpperCase()
}
