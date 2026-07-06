import type { Project } from '../../types'

interface ProjectInfoModalProps {
  project: Project
  onClose: () => void
}

export function ProjectInfoModal({ project, onClose }: ProjectInfoModalProps) {
  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div
        className="bg-va-sidebar border border-va-border rounded-lg shadow-xl w-full max-w-md mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-va-border">
          <h2 className="text-lg font-semibold text-va-text">Project Info</h2>
          <button
            onClick={onClose}
            className="text-va-text-muted hover:text-va-text transition-colors"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          {/* Project Name */}
          <div>
            <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider">
              Project Name
            </label>
            <p className="mt-1 text-va-text font-medium">{project.name}</p>
          </div>

          {/* Working Directory */}
          <div>
            <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider">
              Working Directory
            </label>
            <p className="mt-1 text-va-text font-mono text-sm break-all bg-va-panel px-2 py-1.5 rounded">
              {project.path}
            </p>
          </div>

          {/* Git Remote */}
          <div>
            <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider">
              Git Remote
            </label>
            {project.gitRemote ? (
              <a
                href={project.gitRemote.replace(/\.git$/, '')}
                target="_blank"
                rel="noopener noreferrer"
                className="mt-1 block text-va-accent hover:underline font-mono text-sm break-all"
              >
                {project.gitRemote}
              </a>
            ) : (
              <p className="mt-1 text-va-text-muted italic text-sm">Not configured</p>
            )}
          </div>

          {/* Divider */}
          <div className="border-t border-va-border" />

          {/* Profile Section */}
          <div>
            <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider">
              Profile
            </label>
            <p className="mt-1 text-va-text font-medium">{project.profile.name}</p>
            <p className="text-sm text-va-text-muted">{project.profile.description}</p>
          </div>

          {/* Framework & Spec Type */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider">
                Framework
              </label>
              <p className="mt-1 text-va-text">
                {project.profile.framework || <span className="text-va-text-muted italic">Not set</span>}
              </p>
            </div>
            <div>
              <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider">
                Spec Type
              </label>
              <p className="mt-1 text-va-text">
                {project.profile.specType || <span className="text-va-text-muted italic">Not set</span>}
              </p>
            </div>
          </div>

          {/* Workflow */}
          <div>
            <label className="text-xs font-medium text-va-text-muted uppercase tracking-wider">
              Workflow Steps
            </label>
            <div className="mt-1 flex flex-wrap gap-1.5">
              {project.profile.workflow.map((step, index) => (
                <span
                  key={step}
                  className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-va-panel text-va-text"
                >
                  <span className="text-va-text-muted mr-1">{index + 1}.</span>
                  {step}
                </span>
              ))}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="px-4 py-3 border-t border-va-border flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-1.5 bg-va-panel hover:bg-va-border text-va-text rounded transition-colors text-sm"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  )
}
