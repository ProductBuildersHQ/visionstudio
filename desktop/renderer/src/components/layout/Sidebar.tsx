import { useState } from 'react'
import type { Project, Spec } from '../../types'
import type { Rubric } from '@plexusone/structured-evaluation'
import { extensionRegistry } from '../../extensions/registry'
import { ProjectInfoModal } from './ProjectInfoModal'

interface SidebarProps {
  projects: Project[]
  activeProject: Project | null
  onProjectSelect: (project: Project) => void
  onSpecSelect: (spec: Spec) => void
  onViewSelect: (extensionId: string, viewId: string) => void
  onMethodologyClick?: () => void
  onExtensionsClick?: () => void
  activeView?: { extensionId: string; viewId: string }
  activeSpec: Spec | null
  onAddProjectClick: () => void
  onRemoveProject: (projectName: string) => void
  isConnected?: boolean
}

function getImplMethodologyLabel(methodology?: string): string {
  switch (methodology) {
    case 'aidlc':
      return 'AIDLC'
    case 'speckit':
      return 'SpecKit'
    case 'none':
    default:
      return 'None'
  }
}

function getReqMethodologyLabel(methodology?: string, profile?: string): string {
  const name = methodology || profile || 'startup'
  if (name.startsWith('aws-working-backwards')) {
    const parts = name.split('/')
    return parts.length > 1 ? `AWS WB ${parts[1].charAt(0).toUpperCase() + parts[1].slice(1)}` : 'AWS WB'
  }
  if (name.includes('-')) {
    return name.split('-').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ')
  }
  return name.charAt(0).toUpperCase() + name.slice(1)
}

function StatusIndicator({ spec }: { spec: Spec }) {
  const getStatusColor = () => {
    if (!spec.evalResult) {
      return spec.status === 'not_started' ? 'bg-va-text-muted' : 'bg-va-border'
    }
    if (spec.evalResult.overallDecision === 'pass') return 'bg-va-success'
    if (spec.evalResult.overallDecision === 'conditional') return 'bg-va-warning'
    return 'bg-va-error'
  }

  const getStatusIcon = () => {
    if (spec.status === 'not_started') return '○'
    if (!spec.evalResult) return '◐'
    if (spec.evalResult.overallDecision === 'pass') return '✓'
    if (spec.evalResult.overallDecision === 'conditional') return '⚠'
    return '✗'
  }

  return (
    <span className={`inline-flex items-center justify-center w-5 h-5 rounded-full text-xs ${getStatusColor()}`}>
      {getStatusIcon()}
    </span>
  )
}

function ScoreBadge({ evalResult }: { evalResult?: Rubric }) {
  if (!evalResult || evalResult.intScore === undefined) return null

  const getDecisionColor = () => {
    if (evalResult.overallDecision === 'pass') return 'text-va-success'
    if (evalResult.overallDecision === 'conditional') return 'text-va-warning'
    return 'text-va-error'
  }

  return (
    <span className={`text-xs font-mono ${getDecisionColor()}`}>
      {evalResult.intScore}/5
    </span>
  )
}

export function Sidebar({
  projects,
  activeProject,
  onProjectSelect,
  onSpecSelect,
  onViewSelect,
  onMethodologyClick,
  onExtensionsClick,
  activeView,
  activeSpec,
  onAddProjectClick,
  onRemoveProject,
  isConnected,
}: SidebarProps) {
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(
    new Set(activeProject ? [activeProject.name] : [])
  )
  const [showProjectInfo, setShowProjectInfo] = useState(false)
  const [confirmingRemove, setConfirmingRemove] = useState<string | null>(null)

  const handleRemoveClick = (e: React.MouseEvent, projectName: string) => {
    e.stopPropagation()
    setConfirmingRemove(projectName)
  }

  const handleConfirmRemove = (projectName: string) => {
    onRemoveProject(projectName)
    setConfirmingRemove(null)
  }

  const toggleProject = (projectName: string) => {
    const newExpanded = new Set(expandedProjects)
    if (newExpanded.has(projectName)) {
      newExpanded.delete(projectName)
    } else {
      newExpanded.add(projectName)
    }
    setExpandedProjects(newExpanded)
  }

  const globalSections = extensionRegistry.getGlobalSidebarSections()
  const projectSections = extensionRegistry.getProjectSidebarSections()

  return (
    <div className="p-3 pt-12">
      {/* App branding */}
      <div className="mb-5">
        <div className="text-xs text-va-text-muted tracking-wide">ProductBuildersHQ</div>
        <div className="text-lg font-semibold text-va-text">VisionStudio</div>
      </div>

      {/* Global sections from extensions */}
      {globalSections.map(({ extensionId, section }) => (
        <div key={section.id} className="mb-4">
          <div className="text-xs font-semibold text-va-text-muted uppercase tracking-wider mb-2">
            {section.label}
          </div>
          {section.items.map(item => (
            <button
              key={item.viewId}
              onClick={() => onViewSelect(extensionId, item.viewId)}
              className={`w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left transition-colors ${
                activeView?.extensionId === extensionId && activeView?.viewId === item.viewId
                  ? 'bg-va-panel text-va-text'
                  : 'text-va-text-muted hover:text-va-text hover:bg-va-panel'
              }`}
            >
              {item.icon && <span>{item.icon}</span>}
              <span className="flex-1">{item.label}</span>
            </button>
          ))}
        </div>
      ))}

      {/* Extensions link */}
      {onExtensionsClick && (
        <div className="mb-4">
          <button
            onClick={onExtensionsClick}
            className={`w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left transition-colors ${
              activeView?.extensionId === '_system' && activeView?.viewId === 'extensions'
                ? 'bg-va-panel text-va-text'
                : 'text-va-text-muted hover:text-va-text hover:bg-va-panel'
            }`}
          >
            <span>Extensions</span>
          </button>
        </div>
      )}

      {/* Projects header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="text-xs font-semibold text-va-text-muted uppercase tracking-wider">
            Projects
          </span>
          {isConnected === false && (
            <span
              className="text-va-warning"
              title="Disconnected from daemon - reconnecting..."
            >
              <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
            </span>
          )}
        </div>
        <button
          onClick={onAddProjectClick}
          className="w-5 h-5 flex items-center justify-center text-va-text-muted hover:text-va-text hover:bg-va-panel rounded transition-colors"
          title="Add Project"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
        </button>
      </div>

      {/* Project list */}
      {projects.map((project) => {
        const isExpanded = expandedProjects.has(project.name)
        const isActive = activeProject?.name === project.name

        return (
          <div key={project.name} className="mb-1">
            {/* Project header */}
            <div className="flex items-center group">
              <button
                onClick={() => {
                  toggleProject(project.name)
                  onProjectSelect(project)
                }}
                className={`flex-1 flex items-center gap-2 px-2 py-1.5 rounded-l text-sm text-left hover:bg-va-panel transition-colors ${
                  isActive ? 'bg-va-panel text-va-text' : 'text-va-text-muted'
                }`}
              >
                <span className="text-xs">{isExpanded ? '▼' : '►'}</span>
                <span className="truncate">{project.name}</span>
              </button>
              <button
                onClick={(e) => handleRemoveClick(e, project.name)}
                className="w-6 h-6 flex items-center justify-center text-va-text-muted hover:text-va-error opacity-0 group-hover:opacity-100 transition-all"
                title="Remove project"
              >
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* Remove confirmation */}
            {confirmingRemove === project.name && (
              <div className="ml-4 mt-1 p-2 bg-va-panel border border-va-border rounded text-sm">
                <p className="text-va-text mb-2">Remove &quot;{project.name}&quot; from tracking?</p>
                <p className="text-xs text-va-text-muted mb-2">Files will not be deleted.</p>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleConfirmRemove(project.name)}
                    className="px-2 py-1 bg-va-error hover:bg-va-error/80 text-white rounded text-xs"
                  >
                    Remove
                  </button>
                  <button
                    onClick={() => setConfirmingRemove(null)}
                    className="px-2 py-1 bg-va-border hover:bg-va-text-muted/20 text-va-text rounded text-xs"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}

            {/* Expanded project content */}
            {isExpanded && isActive && (
              <div className="ml-4 mt-1 space-y-1">
                {/* Methodology Section */}
                <div className="px-2 py-1 text-xs text-va-text-muted uppercase tracking-wider">
                  Methodologies
                </div>

                <button
                  onClick={onMethodologyClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors group"
                  title="Click to change requirements methodology"
                >
                  <span>WHAT</span>
                  <span className="flex-1 truncate text-va-text">
                    {getReqMethodologyLabel(project.requirementsMethodology, project.profile.name)}
                  </span>
                  <span className="text-xs text-va-text-muted opacity-0 group-hover:opacity-100">Edit</span>
                </button>

                <button
                  onClick={onMethodologyClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors group"
                  title="Click to change implementation methodology"
                >
                  <span>HOW</span>
                  <span className="flex-1 truncate text-va-text">
                    {getImplMethodologyLabel(project.implementationMethodology)}
                  </span>
                  <span className="text-xs text-va-text-muted opacity-0 group-hover:opacity-100">Edit</span>
                </button>

                <div className="border-t border-va-border my-2" />

                {/* Dynamic extension sections */}
                {projectSections.map(({ extensionId, section }) => (
                  <div key={section.id}>
                    {section.items.map(item => (
                      <button
                        key={item.viewId}
                        onClick={() => onViewSelect(extensionId, item.viewId)}
                        className={`w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left transition-colors ${
                          activeView?.extensionId === extensionId && activeView?.viewId === item.viewId
                            ? 'text-va-accent bg-va-panel'
                            : 'text-va-text-muted hover:text-va-text hover:bg-va-panel'
                        }`}
                      >
                        {item.icon && <span>{item.icon}</span>}
                        <span>{item.label}</span>
                      </button>
                    ))}
                  </div>
                ))}

                {/* Project Info link */}
                <button
                  onClick={() => setShowProjectInfo(true)}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                >
                  <span>ℹ️</span>
                  <span>Project Info</span>
                </button>

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
                    <ScoreBadge evalResult={spec.evalResult} />
                  </button>
                ))}
              </div>
            )}
          </div>
        )
      })}

      {/* Project Info Modal */}
      {showProjectInfo && activeProject && (
        <ProjectInfoModal
          project={activeProject}
          onClose={() => setShowProjectInfo(false)}
        />
      )}
    </div>
  )
}
