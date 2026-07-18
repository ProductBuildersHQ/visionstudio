/**
 * V2MOM Cascade Visualization Component
 *
 * Displays the hierarchical cascade of V2MOMs:
 * Company → Department → Team → Individual
 *
 * Shows alignment between levels (parent Methods → child Vision)
 * and links to ProductContext entities.
 */

import { useState } from 'react'

// V2MOM types
interface V2MOMSection {
  id: string
  type: 'vision' | 'values' | 'methods' | 'obstacles' | 'measures'
  name: string
  status: 'draft' | 'complete' | 'approved'
  score?: number
}

interface V2MOMMethod {
  id: string
  name: string
  owner: string
  status: 'not_started' | 'in_progress' | 'at_risk' | 'completed'
  capabilities?: string[]
  projects?: string[]
  parentMethod?: string
  childMethods?: string[]
}

interface V2MOMLevel {
  id: string
  name: string
  level: 'company' | 'department' | 'team' | 'individual'
  owner: string
  fiscalPeriod: string
  vision: string
  visionScore?: number
  sections: V2MOMSection[]
  methods: V2MOMMethod[]
  parentId?: string
  childIds?: string[]
  alignmentScore?: number
}

interface V2MOMCascadeViewProps {
  v2moms: V2MOMLevel[]
  currentV2MOMID?: string
  onSectionClick: (v2momId: string, section: V2MOMSection) => void
  onMethodClick: (v2momId: string, method: V2MOMMethod) => void
}

export function V2MOMCascadeView({
  v2moms,
  currentV2MOMID,
  onSectionClick,
  onMethodClick,
}: V2MOMCascadeViewProps) {
  const [expandedLevels, setExpandedLevels] = useState<Set<string>>(
    new Set(v2moms.map((v) => v.id))
  )
  const [viewMode, setViewMode] = useState<'cascade' | 'alignment' | 'methods'>(
    'cascade'
  )

  // Build cascade hierarchy
  const rootV2MOMs = v2moms.filter((v) => !v.parentId)

  const toggleLevel = (id: string) => {
    const newExpanded = new Set(expandedLevels)
    if (newExpanded.has(id)) {
      newExpanded.delete(id)
    } else {
      newExpanded.add(id)
    }
    setExpandedLevels(newExpanded)
  }

  const getChildren = (parentId: string): V2MOMLevel[] => {
    return v2moms.filter((v) => v.parentId === parentId)
  }

  return (
    <div className="h-full overflow-auto p-6">
      <div className="max-w-5xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-sm font-semibold text-va-text mb-1">
            V2MOM Cascade
          </h1>
          <p className="text-base text-va-text-muted">
            Strategic alignment from Company → Department → Team
          </p>
        </div>

        {/* View Mode Tabs */}
        <div className="flex items-center gap-2 mb-6">
          <ViewModeButton
            active={viewMode === 'cascade'}
            onClick={() => setViewMode('cascade')}
          >
            Cascade View
          </ViewModeButton>
          <ViewModeButton
            active={viewMode === 'alignment'}
            onClick={() => setViewMode('alignment')}
          >
            Alignment View
          </ViewModeButton>
          <ViewModeButton
            active={viewMode === 'methods'}
            onClick={() => setViewMode('methods')}
          >
            Methods View
          </ViewModeButton>
        </div>

        {/* Legend */}
        <div className="flex items-center gap-6 mb-6 text-base">
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-full bg-va-success" />
            <span className="text-va-text-muted">Approved</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-full bg-va-primary" />
            <span className="text-va-text-muted">Complete</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-full bg-va-warning" />
            <span className="text-va-text-muted">Draft</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-full bg-va-text-muted" />
            <span className="text-va-text-muted">Not Started</span>
          </div>
        </div>

        {/* Main Content */}
        {viewMode === 'cascade' && (
          <CascadeTree
            roots={rootV2MOMs}
            getChildren={getChildren}
            expandedLevels={expandedLevels}
            toggleLevel={toggleLevel}
            currentV2MOMID={currentV2MOMID}
            onSectionClick={onSectionClick}
            onMethodClick={onMethodClick}
          />
        )}

        {viewMode === 'alignment' && (
          <AlignmentMatrix
            v2moms={v2moms}
            currentV2MOMID={currentV2MOMID}
          />
        )}

        {viewMode === 'methods' && (
          <MethodsFlowView
            v2moms={v2moms}
            onMethodClick={onMethodClick}
          />
        )}
      </div>
    </div>
  )
}

function ViewModeButton({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
        active
          ? 'bg-va-primary text-white'
          : 'bg-va-sidebar text-va-text-muted hover:bg-va-panel'
      }`}
    >
      {children}
    </button>
  )
}

// Cascade Tree View - shows hierarchical V2MOMs
function CascadeTree({
  roots,
  getChildren,
  expandedLevels,
  toggleLevel,
  currentV2MOMID,
  onSectionClick,
  onMethodClick,
}: {
  roots: V2MOMLevel[]
  getChildren: (parentId: string) => V2MOMLevel[]
  expandedLevels: Set<string>
  toggleLevel: (id: string) => void
  currentV2MOMID?: string
  onSectionClick: (v2momId: string, section: V2MOMSection) => void
  onMethodClick: (v2momId: string, method: V2MOMMethod) => void
}) {
  return (
    <div className="space-y-4">
      {roots.map((v2mom) => (
        <V2MOMNode
          key={v2mom.id}
          v2mom={v2mom}
          getChildren={getChildren}
          expandedLevels={expandedLevels}
          toggleLevel={toggleLevel}
          currentV2MOMID={currentV2MOMID}
          onSectionClick={onSectionClick}
          onMethodClick={onMethodClick}
          depth={0}
        />
      ))}
    </div>
  )
}

function V2MOMNode({
  v2mom,
  getChildren,
  expandedLevels,
  toggleLevel,
  currentV2MOMID,
  onSectionClick,
  onMethodClick,
  depth,
}: {
  v2mom: V2MOMLevel
  getChildren: (parentId: string) => V2MOMLevel[]
  expandedLevels: Set<string>
  toggleLevel: (id: string) => void
  currentV2MOMID?: string
  onSectionClick: (v2momId: string, section: V2MOMSection) => void
  onMethodClick: (v2momId: string, method: V2MOMMethod) => void
  depth: number
}) {
  const isExpanded = expandedLevels.has(v2mom.id)
  const isCurrent = v2mom.id === currentV2MOMID
  const children = getChildren(v2mom.id)
  const hasChildren = children.length > 0

  const levelColors = {
    company: 'border-l-purple-500',
    department: 'border-l-blue-500',
    team: 'border-l-green-500',
    individual: 'border-l-orange-500',
  }

  return (
    <div
      className={`border-l-4 ${levelColors[v2mom.level]} ${
        isCurrent ? 'bg-va-primary/10' : ''
      }`}
      style={{ marginLeft: depth * 24 }}
    >
      {/* V2MOM Header */}
      <div
        className={`p-4 bg-va-panel rounded-r-lg border border-l-0 border-va-border ${
          isCurrent ? 'border-va-primary' : ''
        }`}
      >
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            {hasChildren && (
              <button
                onClick={() => toggleLevel(v2mom.id)}
                className="text-va-text-muted hover:text-va-text"
              >
                {isExpanded ? '▼' : '▶'}
              </button>
            )}
            <span className="text-xs font-medium text-va-text-muted uppercase">
              {v2mom.level}
            </span>
            <h3 className="text-sm font-semibold text-va-text">{v2mom.name}</h3>
            {v2mom.alignmentScore && (
              <AlignmentBadge score={v2mom.alignmentScore} />
            )}
          </div>
          <div className="text-xs text-va-text-muted">
            {v2mom.owner} • {v2mom.fiscalPeriod}
          </div>
        </div>

        {/* Vision Statement */}
        <div className="text-sm text-va-text-muted italic mb-3 pl-6">
          "{v2mom.vision}"
        </div>

        {/* V2MOM Sections */}
        <div className="flex flex-wrap gap-2 pl-6">
          {v2mom.sections.map((section) => (
            <SectionBadge
              key={section.id}
              section={section}
              onClick={() => onSectionClick(v2mom.id, section)}
            />
          ))}
        </div>

        {/* Methods Summary */}
        {v2mom.methods.length > 0 && (
          <div className="mt-3 pl-6">
            <div className="text-xs text-va-text-muted mb-2">
              Methods ({v2mom.methods.length})
            </div>
            <div className="flex flex-wrap gap-2">
              {v2mom.methods.slice(0, 5).map((method) => (
                <MethodBadge
                  key={method.id}
                  method={method}
                  onClick={() => onMethodClick(v2mom.id, method)}
                />
              ))}
              {v2mom.methods.length > 5 && (
                <span className="text-xs text-va-text-muted">
                  +{v2mom.methods.length - 5} more
                </span>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Child V2MOMs */}
      {isExpanded && hasChildren && (
        <div className="mt-2">
          {/* Cascade Arrow */}
          <div className="flex items-center gap-2 py-1 pl-8">
            <span className="text-va-text-muted">↓</span>
            <span className="text-xs text-va-text-muted">
              Methods cascade to child Vision
            </span>
          </div>
          {children.map((child) => (
            <V2MOMNode
              key={child.id}
              v2mom={child}
              getChildren={getChildren}
              expandedLevels={expandedLevels}
              toggleLevel={toggleLevel}
              currentV2MOMID={currentV2MOMID}
              onSectionClick={onSectionClick}
              onMethodClick={onMethodClick}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function SectionBadge({
  section,
  onClick,
}: {
  section: V2MOMSection
  onClick: () => void
}) {
  const statusColors = {
    draft: 'bg-va-warning/20 border-va-warning',
    complete: 'bg-va-primary/20 border-va-primary',
    approved: 'bg-va-success/20 border-va-success',
  }

  const sectionLabels = {
    vision: '🎯 Vision',
    values: '💎 Values',
    methods: '🚀 Methods',
    obstacles: '🚧 Obstacles',
    measures: '📊 Measures',
  }

  return (
    <button
      onClick={onClick}
      className={`px-2 py-1 rounded border text-xs ${statusColors[section.status]} hover:opacity-80 transition-opacity`}
    >
      {sectionLabels[section.type]}
      {section.score && (
        <span className="ml-1 text-va-text-muted">{section.score}/10</span>
      )}
    </button>
  )
}

function MethodBadge({
  method,
  onClick,
}: {
  method: V2MOMMethod
  onClick: () => void
}) {
  const statusColors = {
    not_started: 'bg-va-sidebar border-va-border',
    in_progress: 'bg-va-primary/20 border-va-primary',
    at_risk: 'bg-va-warning/20 border-va-warning',
    completed: 'bg-va-success/20 border-va-success',
  }

  return (
    <button
      onClick={onClick}
      className={`px-2 py-1 rounded border text-xs ${statusColors[method.status]} hover:opacity-80 transition-opacity`}
    >
      {method.name}
      {method.capabilities && method.capabilities.length > 0 && (
        <span className="ml-1 text-va-text-muted">
          🔗{method.capabilities.length}
        </span>
      )}
    </button>
  )
}

function AlignmentBadge({ score }: { score: number }) {
  const getColor = () => {
    if (score >= 80) return 'bg-va-success text-white'
    if (score >= 60) return 'bg-va-warning text-black'
    return 'bg-va-error text-white'
  }

  return (
    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${getColor()}`}>
      {score}% aligned
    </span>
  )
}

// Alignment Matrix View - shows method-to-method alignment
function AlignmentMatrix({
  v2moms,
  currentV2MOMID,
}: {
  v2moms: V2MOMLevel[]
  currentV2MOMID?: string
}) {
  // Group by level
  const levels = ['company', 'department', 'team', 'individual'] as const
  const byLevel = levels.reduce(
    (acc, level) => {
      acc[level] = v2moms.filter((v) => v.level === level)
      return acc
    },
    {} as Record<string, V2MOMLevel[]>
  )

  return (
    <div className="space-y-6">
      <div className="text-sm text-va-text-muted">
        Shows how Methods at each level support Methods at the level above
      </div>

      {levels.map((level, idx) => {
        const levelV2MOMs = byLevel[level]
        if (levelV2MOMs.length === 0) return null

        const parentLevel = idx > 0 ? levels[idx - 1] : null
        const parentV2MOMs = parentLevel ? byLevel[parentLevel] : []

        return (
          <div key={level}>
            <h3 className="text-xs font-semibold text-va-text uppercase mb-2">
              {level} Level
            </h3>

            {levelV2MOMs.map((v2mom) => (
              <div
                key={v2mom.id}
                className={`p-4 rounded-lg border mb-2 ${
                  v2mom.id === currentV2MOMID
                    ? 'border-va-primary bg-va-primary/10'
                    : 'border-va-border bg-va-panel'
                }`}
              >
                <div className="font-medium text-sm text-va-text mb-2">
                  {v2mom.name}
                </div>

                {/* Alignment to parent */}
                {parentV2MOMs.length > 0 && (
                  <div className="text-xs text-va-text-muted mb-2">
                    Supporting {parentLevel} methods:
                  </div>
                )}

                <div className="grid grid-cols-3 gap-2">
                  {v2mom.methods.map((method) => (
                    <div
                      key={method.id}
                      className="p-2 rounded bg-va-sidebar text-xs"
                    >
                      <div className="font-medium text-va-text">
                        {method.name}
                      </div>
                      {method.parentMethod && (
                        <div className="text-va-text-muted mt-1">
                          ↑ {method.parentMethod}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            ))}

            {idx < levels.length - 1 && byLevel[levels[idx + 1]]?.length > 0 && (
              <div className="flex justify-center py-2">
                <span className="text-va-text-muted">↓ cascades to</span>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}

// Methods Flow View - shows method execution across the cascade
function MethodsFlowView({
  v2moms,
  onMethodClick,
}: {
  v2moms: V2MOMLevel[]
  onMethodClick: (v2momId: string, method: V2MOMMethod) => void
}) {
  // Collect all methods with their V2MOM context
  const allMethods = v2moms.flatMap((v2mom) =>
    v2mom.methods.map((method) => ({
      ...method,
      v2momId: v2mom.id,
      v2momName: v2mom.name,
      level: v2mom.level,
    }))
  )

  // Group by status
  const statusGroups = {
    not_started: allMethods.filter((m) => m.status === 'not_started'),
    in_progress: allMethods.filter((m) => m.status === 'in_progress'),
    at_risk: allMethods.filter((m) => m.status === 'at_risk'),
    completed: allMethods.filter((m) => m.status === 'completed'),
  }

  return (
    <div className="grid grid-cols-4 gap-4">
      {Object.entries(statusGroups).map(([status, methods]) => (
        <div key={status}>
          <h3 className="text-xs font-semibold text-va-text uppercase mb-2 flex items-center gap-2">
            <StatusDot status={status as keyof typeof statusGroups} />
            {status.replace('_', ' ')}
            <span className="text-va-text-muted font-normal">
              ({methods.length})
            </span>
          </h3>

          <div className="space-y-2">
            {methods.map((method) => (
              <button
                key={`${method.v2momId}-${method.id}`}
                onClick={() =>
                  onMethodClick(method.v2momId, {
                    id: method.id,
                    name: method.name,
                    owner: method.owner,
                    status: method.status,
                    capabilities: method.capabilities,
                    projects: method.projects,
                    parentMethod: method.parentMethod,
                  })
                }
                className="w-full p-2 rounded-lg bg-va-panel border border-va-border hover:border-va-primary text-left transition-colors"
              >
                <div className="text-sm font-medium text-va-text">
                  {method.name}
                </div>
                <div className="text-xs text-va-text-muted mt-1">
                  {method.v2momName} • {method.owner}
                </div>
                {method.capabilities && method.capabilities.length > 0 && (
                  <div className="text-xs text-va-primary mt-1">
                    🔗 {method.capabilities.join(', ')}
                  </div>
                )}
              </button>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

function StatusDot({
  status,
}: {
  status: 'not_started' | 'in_progress' | 'at_risk' | 'completed'
}) {
  const colors = {
    not_started: 'bg-va-text-muted',
    in_progress: 'bg-va-primary',
    at_risk: 'bg-va-warning',
    completed: 'bg-va-success',
  }

  return <span className={`w-2 h-2 rounded-full ${colors[status]}`} />
}

export default V2MOMCascadeView
