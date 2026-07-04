import { useState } from 'react'
import type { Project, Spec } from '../../types'

interface SidebarProps {
  projects: Project[]
  activeProject: Project | null
  onProjectSelect: (project: Project) => void
  onSpecSelect: (spec: Spec) => void
  onWorkflowClick: () => void
  activeSpec: Spec | null
}

// Status indicator component
function StatusIndicator({ spec }: { spec: Spec }) {
  const getStatusColor = () => {
    if (!spec.evalResult) {
      return spec.status === 'not_started' ? 'bg-va-text-muted' : 'bg-va-border'
    }
    if (spec.evalResult.decision === 'pass') return 'bg-va-success'
    if (spec.evalResult.decision === 'conditional') return 'bg-va-warning'
    return 'bg-va-error'
  }

  const getStatusIcon = () => {
    if (spec.status === 'not_started') return '○'
    if (!spec.evalResult) return '◐'
    if (spec.evalResult.decision === 'pass') return '✓'
    if (spec.evalResult.decision === 'conditional') return '⚠'
    return '✗'
  }

  return (
    <span className={`inline-flex items-center justify-center w-5 h-5 rounded-full text-xs ${getStatusColor()}`}>
      {getStatusIcon()}
    </span>
  )
}

// Score badge component
function ScoreBadge({ score }: { score?: number }) {
  if (score === undefined) return null

  const getScoreColor = () => {
    if (score >= 8) return 'text-va-success'
    if (score >= 6) return 'text-va-warning'
    return 'text-va-error'
  }

  return (
    <span className={`text-xs font-mono ${getScoreColor()}`}>
      {score.toFixed(1)}
    </span>
  )
}

export function Sidebar({
  projects,
  activeProject,
  onProjectSelect,
  onSpecSelect,
  onWorkflowClick,
  activeSpec,
}: SidebarProps) {
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(
    new Set(activeProject ? [activeProject.name] : [])
  )

  const toggleProject = (projectName: string) => {
    const newExpanded = new Set(expandedProjects)
    if (newExpanded.has(projectName)) {
      newExpanded.delete(projectName)
    } else {
      newExpanded.add(projectName)
    }
    setExpandedProjects(newExpanded)
  }

  return (
    <div className="p-3">
      {/* Projects header */}
      <div className="text-xs font-semibold text-va-text-muted uppercase tracking-wider mb-2">
        Projects
      </div>

      {/* Project list */}
      {projects.map((project) => {
        const isExpanded = expandedProjects.has(project.name)
        const isActive = activeProject?.name === project.name

        return (
          <div key={project.name} className="mb-1">
            {/* Project header */}
            <button
              onClick={() => {
                toggleProject(project.name)
                onProjectSelect(project)
              }}
              className={`w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left hover:bg-va-panel transition-colors ${
                isActive ? 'bg-va-panel text-va-text' : 'text-va-text-muted'
              }`}
            >
              <span className="text-xs">{isExpanded ? '▼' : '►'}</span>
              <span>{project.name}</span>
            </button>

            {/* Expanded project content */}
            {isExpanded && isActive && (
              <div className="ml-4 mt-1 space-y-1">
                {/* Profile dropdown */}
                <div className="px-2 py-1">
                  <label className="text-xs text-va-text-muted block mb-1">Profile:</label>
                  <select
                    className="w-full bg-va-panel border border-va-border rounded px-2 py-1 text-sm text-va-text"
                    value={project.profile.name}
                    onChange={() => {/* TODO: profile change */}}
                  >
                    <option value={project.profile.name}>{project.profile.name}</option>
                  </select>
                </div>

                {/* Workflow link */}
                <button
                  onClick={onWorkflowClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-accent hover:bg-va-panel transition-colors"
                >
                  <span>📊</span>
                  <span>Workflow</span>
                </button>

                {/* Divider */}
                <div className="border-t border-va-border my-2" />

                {/* Spec list */}
                {project.specs.map((spec) => (
                  <button
                    key={spec.type}
                    onClick={() => onSpecSelect(spec)}
                    className={`w-full flex items-center justify-between gap-2 px-2 py-1.5 rounded text-sm text-left hover:bg-va-panel transition-colors ${
                      activeSpec?.type === spec.type ? 'bg-va-panel' : ''
                    }`}
                  >
                    <div className="flex items-center gap-2">
                      <StatusIndicator spec={spec} />
                      <span>{spec.name}</span>
                    </div>
                    <ScoreBadge score={spec.evalResult?.score} />
                  </button>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
