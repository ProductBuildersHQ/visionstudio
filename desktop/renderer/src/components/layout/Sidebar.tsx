import { useState } from 'react'
import type { Project, Spec, EvalResult } from '../../types'
import { ProjectInfoModal } from './ProjectInfoModal'

interface SidebarProps {
  projects: Project[]
  activeProject: Project | null
  onProjectSelect: (project: Project) => void
  onSpecSelect: (spec: Spec) => void
  onWorkflowClick: () => void
  onFindingsClick: () => void
  onV2MOMClick: () => void
  onMaturityModelClick: () => void
  onCapabilitiesClick: () => void
  onRoadmapClick: () => void
  onAIDLCWorkflowClick?: () => void
  onAIDLCSyncClick?: () => void
  onMethodologyClick?: () => void
  onOrganizationClick?: () => void
  activeSpec: Spec | null
  onAddProjectClick: () => void
  onRemoveProject: (projectName: string) => void
  isConnected?: boolean
}

// Helper to get a display label for implementation methodology
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

// Helper to get a short label for requirements methodology
function getReqMethodologyLabel(methodology?: string, profile?: string): string {
  const name = methodology || profile || 'startup'
  // Extract a shorter name from "aws-working-backwards/product" -> "AWS WB Product"
  if (name.startsWith('aws-working-backwards')) {
    const parts = name.split('/')
    return parts.length > 1 ? `AWS WB ${parts[1].charAt(0).toUpperCase() + parts[1].slice(1)}` : 'AWS WB'
  }
  if (name.includes('-')) {
    return name.split('-').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ')
  }
  return name.charAt(0).toUpperCase() + name.slice(1)
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

// Score badge component - supports v2 (1-5) and legacy (0-10) scores
// Color is based on decision (pass/conditional/fail) for consistency with status indicators
function ScoreBadge({ evalResult }: { evalResult?: EvalResult }) {
  if (!evalResult) return null

  // Color based on decision for consistency with dot and header
  const getDecisionColor = () => {
    if (evalResult.decision === 'pass') return 'text-va-success'
    if (evalResult.decision === 'conditional') return 'text-va-warning'
    return 'text-va-error'
  }

  // Check for v2 format
  const isV2 = evalResult.schemaVersion === 'v2' || evalResult.scoreV2 !== undefined

  if (isV2 && evalResult.scoreV2) {
    return (
      <span className={`text-xs font-mono ${getDecisionColor()}`}>
        {evalResult.scoreV2}/5
      </span>
    )
  }

  // Legacy v1 format
  return (
    <span className={`text-xs font-mono ${getDecisionColor()}`}>
      {evalResult.score.toFixed(1)}
    </span>
  )
}

export function Sidebar({
  projects,
  activeProject,
  onProjectSelect,
  onSpecSelect,
  onWorkflowClick,
  onFindingsClick,
  onV2MOMClick,
  onMaturityModelClick,
  onCapabilitiesClick,
  onRoadmapClick,
  onAIDLCWorkflowClick,
  onAIDLCSyncClick,
  onMethodologyClick,
  onOrganizationClick,
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

  return (
    <div className="p-3 pt-12">
      {/* App branding - pt-12 accommodates macOS traffic lights + breathing room */}
      <div className="mb-5">
        <div className="text-xs text-va-text-muted tracking-wide">ProductBuildersHQ</div>
        <div className="text-lg font-semibold text-va-text">VisionStudio</div>
      </div>

      {/* Organization section */}
      {onOrganizationClick && (
        <div className="mb-4">
          <div className="text-xs font-semibold text-va-text-muted uppercase tracking-wider mb-2">
            Organization
          </div>
          <button
            onClick={onOrganizationClick}
            className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
          >
            <span>Building</span>
            <span className="flex-1">Strategy & V2MOMs</span>
          </button>
        </div>
      )}

      {/* Projects header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="text-xs font-semibold text-va-text-muted uppercase tracking-wider">
            Projects
          </span>
          {/* Connection indicator - only show when disconnected */}
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
                <p className="text-va-text mb-2">Remove "{project.name}" from tracking?</p>
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

                {/* Requirements Methodology */}
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

                {/* Implementation Methodology */}
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

                {/* Divider after methodologies */}
                <div className="border-t border-va-border my-2" />

                {/* Workflow link */}
                <button
                  onClick={onWorkflowClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-accent hover:bg-va-panel transition-colors"
                >
                  <span>📊</span>
                  <span>Workflow</span>
                </button>

                {/* V2MOM Cascade link */}
                <button
                  onClick={onV2MOMClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                >
                  <span>🎯</span>
                  <span>V2MOM Cascade</span>
                </button>

                {/* Capabilities link */}
                <button
                  onClick={onCapabilitiesClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                >
                  <span>🧱</span>
                  <span>Capabilities</span>
                </button>

                {/* Roadmap link */}
                <button
                  onClick={onRoadmapClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                >
                  <span>🗺️</span>
                  <span>Roadmap</span>
                </button>

                {/* Maturity Model link */}
                <button
                  onClick={onMaturityModelClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                >
                  <span>📈</span>
                  <span>Maturity Model</span>
                </button>

                {/* All Findings link */}
                <button
                  onClick={onFindingsClick}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                >
                  <span>📋</span>
                  <span>All Findings</span>
                </button>

                {/* AIDLC Section - only show when implementation methodology is aidlc */}
                {project.implementationMethodology === 'aidlc' && (
                  <>
                    {onAIDLCWorkflowClick && (
                      <button
                        onClick={onAIDLCWorkflowClick}
                        className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                      >
                        <span>AIDLC Workflow</span>
                      </button>
                    )}

                    {onAIDLCSyncClick && (
                      <button
                        onClick={onAIDLCSyncClick}
                        className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                      >
                        <span>AIDLC Sync</span>
                      </button>
                    )}
                  </>
                )}

                {/* SpecKit Section - only show when implementation methodology is speckit */}
                {project.implementationMethodology === 'speckit' && (
                  <button
                    onClick={() => {/* TODO: SpecKit handler */}}
                    className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                  >
                    <span>SpecKit</span>
                  </button>
                )}

                {/* Project Info link */}
                <button
                  onClick={() => setShowProjectInfo(true)}
                  className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
                >
                  <span>ℹ️</span>
                  <span>Project Info</span>
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
