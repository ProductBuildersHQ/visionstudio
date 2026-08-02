import { useState } from 'react'
import type { ExecutionResponse } from '../api/types'
import type { NavTarget } from '../App'

interface SidebarProps {
  collapsed: boolean
  onToggleCollapse: () => void
  execution: ExecutionResponse
  navTarget: NavTarget
  onNavigate: (target: NavTarget) => void
  apiStatus: 'loading' | 'connected' | 'error'
}

export function Sidebar({
  collapsed,
  onToggleCollapse,
  execution,
  navTarget,
  onNavigate,
  apiStatus,
}: SidebarProps) {
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['initiatives']))
  const [expandedPrograms, setExpandedPrograms] = useState<Set<string>>(new Set())

  const toggleSection = (section: string) => {
    const next = new Set(expandedSections)
    if (next.has(section)) {
      next.delete(section)
    } else {
      next.add(section)
    }
    setExpandedSections(next)
  }

  const toggleProgram = (programId: string) => {
    const next = new Set(expandedPrograms)
    if (next.has(programId)) {
      next.delete(programId)
    } else {
      next.add(programId)
    }
    setExpandedPrograms(next)
  }

  const visiblePrograms = execution.programs.filter((p) => !p.hidden)
  const standaloneInitiatives = execution.initiatives.filter((i) => !i.programId)

  const isActive = (target: NavTarget): boolean => {
    if (target.section !== navTarget.section) return false
    if (target.section === 'maturity' || target.section === 'spend') return true
    if (target.section === 'initiatives' && navTarget.section === 'initiatives') {
      if (target.view === 'all' && navTarget.view === 'all') return true
      if (target.view === 'program' && navTarget.view === 'program') {
        return target.programId === navTarget.programId
      }
      if (target.view === 'standalone' && navTarget.view === 'standalone') return true
      if (target.view === 'initiative' && navTarget.view === 'initiative') {
        return target.initiativeId === navTarget.initiativeId
      }
    }
    return false
  }

  return (
    <aside
      className={`${
        collapsed ? 'w-12' : 'w-64'
      } bg-gray-800 border-r border-gray-700 flex flex-col transition-all duration-200`}
    >
      {/* Header */}
      <div className="p-3 border-b border-gray-700 flex items-center justify-between">
        {!collapsed && (
          <div className="flex items-center gap-2">
            <span className="font-semibold text-sm">VisionStudio</span>
            <span
              className={`h-2 w-2 rounded-full ${
                apiStatus === 'connected'
                  ? 'bg-green-500'
                  : apiStatus === 'error'
                  ? 'bg-red-500'
                  : 'bg-yellow-500'
              }`}
              title={apiStatus}
            />
          </div>
        )}
        <button
          onClick={onToggleCollapse}
          className="p-1 hover:bg-gray-700 rounded text-gray-400 hover:text-gray-200"
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? '»' : '«'}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-2">
        {collapsed ? (
          <CollapsedNav navTarget={navTarget} onNavigate={onNavigate} />
        ) : (
          <>
            {/* Initiatives Section */}
            <NavSection
              label="Initiatives"
              icon="📋"
              expanded={expandedSections.has('initiatives')}
              onToggle={() => toggleSection('initiatives')}
              active={navTarget.section === 'initiatives'}
              onClick={() => onNavigate({ section: 'initiatives', view: 'all' })}
            >
              {/* Programs */}
              {visiblePrograms.map((program) => {
                const programInits = execution.initiatives.filter((i) => i.programId === program.id)
                return (
                  <NavGroup
                    key={program.id}
                    label={program.name}
                    count={programInits.length}
                    expanded={expandedPrograms.has(program.id)}
                    onToggle={() => toggleProgram(program.id)}
                    active={
                      navTarget.section === 'initiatives' &&
                      navTarget.view === 'program' &&
                      navTarget.programId === program.id
                    }
                    onClick={() => onNavigate({ section: 'initiatives', view: 'program', programId: program.id })}
                  >
                    {programInits.map((init) => (
                      <NavItem
                        key={init.id}
                        label={init.id}
                        sublabel={init.title}
                        progress={init.progress}
                        active={isActive({ section: 'initiatives', view: 'initiative', initiativeId: init.id })}
                        onClick={() =>
                          onNavigate({ section: 'initiatives', view: 'initiative', initiativeId: init.id })
                        }
                      />
                    ))}
                  </NavGroup>
                )
              })}

              {/* Standalone */}
              {standaloneInitiatives.length > 0 && (
                <NavGroup
                  label="Standalone"
                  count={standaloneInitiatives.length}
                  expanded={expandedPrograms.has('_standalone')}
                  onToggle={() => toggleProgram('_standalone')}
                  active={navTarget.section === 'initiatives' && navTarget.view === 'standalone'}
                  onClick={() => onNavigate({ section: 'initiatives', view: 'standalone' })}
                >
                  {standaloneInitiatives.map((init) => (
                    <NavItem
                      key={init.id}
                      label={init.id}
                      sublabel={init.title}
                      progress={init.progress}
                      active={isActive({ section: 'initiatives', view: 'initiative', initiativeId: init.id })}
                      onClick={() =>
                        onNavigate({ section: 'initiatives', view: 'initiative', initiativeId: init.id })
                      }
                    />
                  ))}
                </NavGroup>
              )}
            </NavSection>

            {/* Maturity Section */}
            <NavSection
              label="Maturity"
              icon="📈"
              expanded={false}
              active={navTarget.section === 'maturity'}
              onClick={() => onNavigate({ section: 'maturity' })}
            />

            {/* Spend Section */}
            <NavSection
              label="Spend"
              icon="💰"
              expanded={false}
              active={navTarget.section === 'spend'}
              onClick={() => onNavigate({ section: 'spend' })}
            />
          </>
        )}
      </nav>
    </aside>
  )
}

function CollapsedNav({
  navTarget,
  onNavigate,
}: {
  navTarget: NavTarget
  onNavigate: (target: NavTarget) => void
}) {
  return (
    <div className="flex flex-col items-center gap-1 px-1">
      <button
        onClick={() => onNavigate({ section: 'initiatives', view: 'all' })}
        className={`p-2 rounded hover:bg-gray-700 ${
          navTarget.section === 'initiatives' ? 'bg-gray-700' : ''
        }`}
        title="Initiatives"
      >
        📋
      </button>
      <button
        onClick={() => onNavigate({ section: 'maturity' })}
        className={`p-2 rounded hover:bg-gray-700 ${
          navTarget.section === 'maturity' ? 'bg-gray-700' : ''
        }`}
        title="Maturity"
      >
        📈
      </button>
      <button
        onClick={() => onNavigate({ section: 'spend' })}
        className={`p-2 rounded hover:bg-gray-700 ${
          navTarget.section === 'spend' ? 'bg-gray-700' : ''
        }`}
        title="Spend"
      >
        💰
      </button>
    </div>
  )
}

function NavSection({
  label,
  icon,
  expanded,
  onToggle,
  active,
  onClick,
  children,
}: {
  label: string
  icon: string
  expanded?: boolean
  onToggle?: () => void
  active: boolean
  onClick: () => void
  children?: React.ReactNode
}) {
  const hasChildren = !!children

  return (
    <div className="mb-1">
      <button
        onClick={hasChildren ? onToggle : onClick}
        className={`w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-700 ${
          active ? 'bg-gray-700 text-white' : 'text-gray-300'
        }`}
      >
        <span>{icon}</span>
        <span className="flex-1 text-left font-medium">{label}</span>
        {hasChildren && <span className="text-gray-500 text-xs">{expanded ? '▼' : '▶'}</span>}
      </button>
      {hasChildren && expanded && <div className="ml-2">{children}</div>}
    </div>
  )
}

function NavGroup({
  label,
  count,
  expanded,
  onToggle,
  active,
  onClick,
  children,
}: {
  label: string
  count: number
  expanded: boolean
  onToggle: () => void
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <div>
      <div className="flex items-center">
        <button
          onClick={(e) => {
            e.stopPropagation()
            onToggle()
          }}
          className="p-1 text-gray-500 hover:text-gray-300"
        >
          {expanded ? '▼' : '▶'}
        </button>
        <button
          onClick={onClick}
          className={`flex-1 flex items-center justify-between px-2 py-1.5 text-sm rounded hover:bg-gray-700 ${
            active ? 'bg-gray-700/50 text-white' : 'text-gray-400'
          }`}
        >
          <span className="truncate">{label}</span>
          <span className="text-xs text-gray-500">{count}</span>
        </button>
      </div>
      {expanded && <div className="ml-4">{children}</div>}
    </div>
  )
}

function NavItem({
  label,
  sublabel,
  progress,
  active,
  onClick,
}: {
  label: string
  sublabel: string
  progress: number
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left px-2 py-1.5 text-xs rounded hover:bg-gray-700 ${
        active ? 'bg-blue-500/20 text-blue-300' : 'text-gray-400'
      }`}
    >
      <div className="flex items-center justify-between">
        <span className="font-mono text-gray-500">{label}</span>
        <span className="text-gray-500">{Math.round(progress * 100)}%</span>
      </div>
      <div className="truncate text-gray-300">{sublabel}</div>
    </button>
  )
}
