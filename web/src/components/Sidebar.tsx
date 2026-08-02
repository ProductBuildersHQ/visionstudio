import { useState } from 'react'
import { useLocation } from 'react-router-dom'
import type { ExecutionResponse } from '../api/types'
import type { NavTarget } from '../App'

interface SidebarProps {
  collapsed: boolean
  onToggleCollapse: () => void
  execution: ExecutionResponse
  onNavigate: (target: NavTarget) => void
  apiStatus: 'loading' | 'connected' | 'error'
}

export function Sidebar({
  collapsed,
  onToggleCollapse,
  execution,
  onNavigate,
  apiStatus,
}: SidebarProps) {
  const location = useLocation()
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

  const currentSection = (): 'initiatives' | 'maturity' | 'spend' => {
    if (location.pathname === '/maturity') return 'maturity'
    if (location.pathname === '/spend') return 'spend'
    return 'initiatives'
  }

  const isActivePath = (path: string): boolean => {
    return location.pathname === path
  }

  const isInitiativeActive = (initiativeId: string): boolean => {
    return location.pathname === `/initiative/${initiativeId}`
  }

  const isProgramActive = (programId: string): boolean => {
    return location.pathname === `/program/${programId}`
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
          <button
            onClick={() => onNavigate({ section: 'initiatives', view: 'all' })}
            className="flex items-center gap-2 hover:text-white transition-colors"
          >
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
          </button>
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
          <CollapsedNav currentSection={currentSection()} onNavigate={onNavigate} />
        ) : (
          <>
            {/* Initiatives Section */}
            <NavSection
              label="Initiatives"
              icon="📋"
              expanded={expandedSections.has('initiatives')}
              onToggle={() => toggleSection('initiatives')}
              active={currentSection() === 'initiatives'}
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
                    active={isProgramActive(program.id)}
                    onClick={() => onNavigate({ section: 'initiatives', view: 'program', programId: program.id })}
                  >
                    {programInits.map((init) => (
                      <NavItem
                        key={init.id}
                        label={init.id}
                        sublabel={init.title}
                        progress={init.progress}
                        active={isInitiativeActive(init.id)}
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
                  active={isActivePath('/standalone')}
                  onClick={() => onNavigate({ section: 'initiatives', view: 'standalone' })}
                >
                  {standaloneInitiatives.map((init) => (
                    <NavItem
                      key={init.id}
                      label={init.id}
                      sublabel={init.title}
                      progress={init.progress}
                      active={isInitiativeActive(init.id)}
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
              active={isActivePath('/maturity')}
              onClick={() => onNavigate({ section: 'maturity' })}
            />

            {/* Spend Section */}
            <NavSection
              label="Spend"
              icon="💰"
              expanded={false}
              active={isActivePath('/spend')}
              onClick={() => onNavigate({ section: 'spend' })}
            />
          </>
        )}
      </nav>
    </aside>
  )
}

function CollapsedNav({
  currentSection,
  onNavigate,
}: {
  currentSection: 'initiatives' | 'maturity' | 'spend'
  onNavigate: (target: NavTarget) => void
}) {
  return (
    <div className="flex flex-col items-center gap-1 px-1">
      <button
        onClick={() => onNavigate({ section: 'initiatives', view: 'all' })}
        className={`p-2 rounded hover:bg-gray-700 ${
          currentSection === 'initiatives' ? 'bg-gray-700' : ''
        }`}
        title="Initiatives"
      >
        📋
      </button>
      <button
        onClick={() => onNavigate({ section: 'maturity' })}
        className={`p-2 rounded hover:bg-gray-700 ${
          currentSection === 'maturity' ? 'bg-gray-700' : ''
        }`}
        title="Maturity"
      >
        📈
      </button>
      <button
        onClick={() => onNavigate({ section: 'spend' })}
        className={`p-2 rounded hover:bg-gray-700 ${
          currentSection === 'spend' ? 'bg-gray-700' : ''
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
