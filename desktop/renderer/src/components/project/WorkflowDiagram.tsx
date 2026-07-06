import type { Project, Spec } from '../../types'

interface WorkflowDiagramProps {
  project: Project
  onSpecClick: (spec: Spec) => void
}

export function WorkflowDiagram({ project, onSpecClick }: WorkflowDiagramProps) {
  // Group specs into rows for visualization
  const sourceSpecs = project.specs.filter((s) =>
    ['mrd', 'opportunity-spec'].includes(s.type)
  )
  const gtmSpecs = project.specs.filter((s) =>
    ['press', 'faq', 'narrative-1p', 'narrative-6p'].includes(s.type)
  )
  const productSpecs = project.specs.filter((s) =>
    ['prd', 'uxd'].includes(s.type)
  )
  const techSpecs = project.specs.filter((s) =>
    ['trd', 'tpd', 'ird'].includes(s.type)
  )

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-sm font-semibold text-va-text mb-1">
            Workflow: {project.profile.name}
          </h1>
          <p className="text-base text-va-text-muted">
            Click on any spec to open it in the editor
          </p>
        </div>

        {/* Legend */}
        <div className="flex items-center gap-6 mb-6 text-base">
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-full bg-va-success" />
            <span className="text-va-text-muted">Pass</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-full bg-va-warning" />
            <span className="text-va-text-muted">Needs Work</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-full bg-va-error" />
            <span className="text-va-text-muted">Fail</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-full bg-va-text-muted" />
            <span className="text-va-text-muted">Not Started</span>
          </div>
        </div>

        {/* Workflow diagram */}
        <div className="space-y-2">
          {/* Source row */}
          {sourceSpecs.length > 0 && (
            <WorkflowRow
              title="Source"
              description="Human-authored discovery documents"
              specs={sourceSpecs}
              onSpecClick={onSpecClick}
            />
          )}

          {/* Arrow */}
          {sourceSpecs.length > 0 && gtmSpecs.length > 0 && <FlowArrow />}

          {/* GTM row */}
          {gtmSpecs.length > 0 && (
            <WorkflowRow
              title="Go-to-Market"
              description="LLM-generated vision documents"
              specs={gtmSpecs}
              onSpecClick={onSpecClick}
            />
          )}

          {/* Arrow */}
          {gtmSpecs.length > 0 && productSpecs.length > 0 && <FlowArrow />}

          {/* Product row */}
          {productSpecs.length > 0 && (
            <WorkflowRow
              title="Product"
              description="Requirements and experience design"
              specs={productSpecs}
              onSpecClick={onSpecClick}
            />
          )}

          {/* Arrow */}
          {productSpecs.length > 0 && techSpecs.length > 0 && <FlowArrow />}

          {/* Technical row */}
          {techSpecs.length > 0 && (
            <WorkflowRow
              title="Technical"
              description="LLM-generated implementation specs"
              specs={techSpecs}
              onSpecClick={onSpecClick}
            />
          )}
        </div>
      </div>
    </div>
  )
}

function WorkflowRow({
  title,
  description,
  specs,
  onSpecClick,
}: {
  title: string
  description: string
  specs: Spec[]
  onSpecClick: (spec: Spec) => void
}) {
  return (
    <div className="bg-va-panel rounded-lg p-4 border border-va-border">
      <div className="mb-3">
        <h3 className="text-xs font-medium text-va-text uppercase tracking-wide">{title}</h3>
        <p className="text-base text-va-text-muted">{description}</p>
      </div>
      <div className="flex flex-wrap gap-3">
        {specs.map((spec, idx) => (
          <div key={spec.type} className="flex items-center gap-2">
            <SpecNode spec={spec} onClick={() => onSpecClick(spec)} />
            {idx < specs.length - 1 && (
              <span className="text-va-text-muted">→</span>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

function SpecNode({ spec, onClick }: { spec: Spec; onClick: () => void }) {
  const getStatusColor = () => {
    if (!spec.evalResult) {
      return spec.status === 'not_started'
        ? 'border-va-text-muted'
        : 'border-va-border'
    }
    if (spec.evalResult.decision === 'pass') return 'border-va-success'
    if (spec.evalResult.decision === 'conditional') return 'border-va-warning'
    return 'border-va-error'
  }

  const getStatusBg = () => {
    if (!spec.evalResult) return 'bg-va-sidebar'
    if (spec.evalResult.decision === 'pass') return 'bg-va-success/10'
    if (spec.evalResult.decision === 'conditional') return 'bg-va-warning/10'
    return 'bg-va-error/10'
  }

  return (
    <button
      onClick={onClick}
      className={`px-4 py-3 rounded-lg border-2 ${getStatusColor()} ${getStatusBg()} hover:bg-va-bg transition-colors text-left min-w-[120px]`}
    >
      <div className="text-sm font-medium text-va-text">{spec.name}</div>
      {spec.evalResult && (
        <div className="text-xs text-va-text-muted mt-1">
          Score: {spec.evalResult.score.toFixed(1)}
        </div>
      )}
      {spec.status === 'not_started' && (
        <div className="text-xs text-va-text-muted mt-1">Not started</div>
      )}
    </button>
  )
}

function FlowArrow() {
  return (
    <div className="flex justify-center py-0">
      <div className="text-va-text-muted text-sm">↓</div>
    </div>
  )
}
