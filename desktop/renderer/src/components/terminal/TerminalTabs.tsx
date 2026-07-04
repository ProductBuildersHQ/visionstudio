import { useState, useRef, useEffect } from 'react'
import type { TerminalTab, TmuxSession } from '../../types/terminal'

interface TerminalTabsProps {
  tabs: TerminalTab[]
  activeTabId: string | null
  onTabSelect: (tabId: string) => void
  onTabClose: (tabId: string) => void
  onNewTab: () => void
  onToggleExpand: () => void
  isExpanded: boolean
  tmuxSessions?: TmuxSession[]
  tmuxAvailable?: boolean
  onTmuxAttach?: (sessionName: string) => void
  onTmuxCreate?: (sessionName: string) => void
  showWorkflowOverlay?: boolean
  onToggleWorkflowOverlay?: () => void
}

export function TerminalTabs({
  tabs,
  activeTabId,
  onTabSelect,
  onTabClose,
  onNewTab,
  onToggleExpand,
  isExpanded,
  tmuxSessions = [],
  tmuxAvailable = false,
  onTmuxAttach,
  onTmuxCreate,
  showWorkflowOverlay = false,
  onToggleWorkflowOverlay,
}: TerminalTabsProps) {
  const [showTmuxMenu, setShowTmuxMenu] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setShowTmuxMenu(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <div className="flex items-center justify-between px-2 py-1 bg-va-sidebar border-b border-va-border">
      <div className="flex items-center space-x-1 overflow-x-auto">
        {tabs.map((tab) => (
          <Tab
            key={tab.id}
            tab={tab}
            isActive={tab.id === activeTabId}
            onSelect={() => onTabSelect(tab.id)}
            onClose={() => onTabClose(tab.id)}
          />
        ))}

        {/* New Tab Button */}
        <button
          onClick={onNewTab}
          className="flex items-center justify-center w-6 h-6 text-va-text-muted hover:text-va-text hover:bg-va-panel rounded"
          title="New Terminal (Cmd+T)"
        >
          <PlusIcon />
        </button>
      </div>

      <div className="flex items-center space-x-1">
        {/* Workflow Overlay Toggle */}
        {onToggleWorkflowOverlay && (
          <button
            onClick={onToggleWorkflowOverlay}
            className={`flex items-center space-x-1 px-2 py-1 text-sm rounded ${
              showWorkflowOverlay
                ? 'text-va-accent bg-va-accent/10'
                : 'text-va-text-muted hover:text-va-text hover:bg-va-panel'
            }`}
            title="Toggle Workflow Status"
          >
            <WorkflowIcon />
            <span className="text-xs">Status</span>
          </button>
        )}

        {/* tmux Sessions Dropdown */}
        {tmuxAvailable && (
          <div className="relative" ref={menuRef}>
            <button
              onClick={() => setShowTmuxMenu(!showTmuxMenu)}
              className="flex items-center space-x-1 px-2 py-1 text-sm text-va-text-muted hover:text-va-text hover:bg-va-panel rounded"
              title="tmux Sessions"
            >
              <TmuxIcon />
              <span className="text-xs">tmux</span>
              <ChevronDownIcon />
            </button>

            {showTmuxMenu && (
              <div className="absolute right-0 top-full mt-1 w-56 bg-va-panel border border-va-border rounded shadow-lg z-50">
                <div className="py-1">
                  {tmuxSessions.length > 0 ? (
                    <>
                      <div className="px-3 py-1 text-xs text-va-text-muted uppercase">Sessions</div>
                      {tmuxSessions.map((session) => (
                        <button
                          key={session.name}
                          onClick={() => {
                            onTmuxAttach?.(session.name)
                            setShowTmuxMenu(false)
                          }}
                          className="w-full text-left px-3 py-2 text-sm text-va-text hover:bg-va-sidebar flex items-center justify-between"
                        >
                          <span className="truncate">{session.name}</span>
                          <span className="text-xs text-va-text-muted">
                            {session.windows} win{session.windows !== 1 ? 's' : ''}
                            {session.attached && ' (attached)'}
                          </span>
                        </button>
                      ))}
                      <div className="border-t border-va-border my-1" />
                    </>
                  ) : (
                    <div className="px-3 py-2 text-sm text-va-text-muted">No sessions</div>
                  )}
                  <button
                    onClick={() => {
                      const name = `visionstudio-${Date.now()}`
                      onTmuxCreate?.(name)
                      setShowTmuxMenu(false)
                    }}
                    className="w-full text-left px-3 py-2 text-sm text-va-accent hover:bg-va-sidebar flex items-center space-x-2"
                  >
                    <PlusIcon />
                    <span>New tmux Session</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Expand/Collapse Button */}
        <button
          onClick={onToggleExpand}
          className="flex items-center justify-center w-6 h-6 text-va-text-muted hover:text-va-text hover:bg-va-panel rounded"
          title={isExpanded ? 'Collapse (Cmd+`)' : 'Expand (Cmd+`)'}
        >
          {isExpanded ? <ChevronDownIcon /> : <ChevronUpIcon />}
        </button>
      </div>
    </div>
  )
}

interface TabProps {
  tab: TerminalTab
  isActive: boolean
  onSelect: () => void
  onClose: () => void
}

function Tab({ tab, isActive, onSelect, onClose }: TabProps) {
  return (
    <div
      className={`
        flex items-center space-x-2 px-3 py-1 rounded cursor-pointer
        ${isActive ? 'bg-va-panel text-va-text' : 'text-va-text-muted hover:text-va-text hover:bg-va-panel/50'}
      `}
      onClick={onSelect}
    >
      <TerminalIcon />
      <span className="text-sm truncate max-w-32">{tab.title}</span>
      {tab.isTmux && (
        <span className="text-xs px-1 bg-va-accent/20 text-va-accent rounded">tmux</span>
      )}
      <button
        onClick={(e) => {
          e.stopPropagation()
          onClose()
        }}
        className="w-4 h-4 flex items-center justify-center text-va-text-muted hover:text-va-text hover:bg-va-border/50 rounded"
        title="Close"
      >
        <CloseIcon />
      </button>
    </div>
  )
}

// Icons
function PlusIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M7 1v12M1 7h12" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  )
}

function CloseIcon() {
  return (
    <svg width="10" height="10" viewBox="0 0 10 10" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M1 1l8 8M9 1L1 9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  )
}

function TerminalIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="1" y="2" width="12" height="10" rx="1" stroke="currentColor" strokeWidth="1.2" />
      <path d="M3 7l2 2-2 2" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M7 10h3" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
    </svg>
  )
}

function ChevronDownIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M3 5l4 4 4-4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function ChevronUpIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path d="M3 9l4-4 4 4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function TmuxIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="1" y="2" width="12" height="10" rx="1" stroke="currentColor" strokeWidth="1.2" />
      <path d="M1 5h12" stroke="currentColor" strokeWidth="1.2" />
      <path d="M7 5v7" stroke="currentColor" strokeWidth="1.2" />
    </svg>
  )
}

function WorkflowIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="3" cy="7" r="2" stroke="currentColor" strokeWidth="1.2" />
      <circle cx="11" cy="3" r="2" stroke="currentColor" strokeWidth="1.2" />
      <circle cx="11" cy="11" r="2" stroke="currentColor" strokeWidth="1.2" />
      <path d="M5 6.5l4-2.5M5 7.5l4 2.5" stroke="currentColor" strokeWidth="1.2" />
    </svg>
  )
}
