import { useEffect, useCallback, useRef, useState } from 'react'
import { TerminalTabs } from './TerminalTabs'
import { TerminalInstance } from './TerminalInstance'
import { WorkflowOverlay } from './WorkflowOverlay'
import { useTerminal } from './hooks/useTerminal'
import { useTmux } from './hooks/useTmux'
import type { TerminalTab } from '../../types/terminal'

interface TerminalPanelProps {
  height: number
  onHeightChange: (height: number) => void
  projectPath?: string
  projectName?: string
}

const MIN_HEIGHT = 100
const MAX_HEIGHT = 600
const DEFAULT_HEIGHT = 250

export function TerminalPanel({ height, onHeightChange, projectPath, projectName }: TerminalPanelProps) {
  const {
    tabs,
    activeTabId,
    isExpanded,
    createTab,
    closeTab,
    setActiveTab,
    toggleExpanded,
    updateTabTitle,
  } = useTerminal()

  const {
    sessions: tmuxSessions,
    isAvailable: tmuxAvailable,
    attach: tmuxAttach,
    create: tmuxCreate,
  } = useTmux()

  const [allTabs, setAllTabs] = useState<TerminalTab[]>([])
  const [showWorkflowOverlay, setShowWorkflowOverlay] = useState(true)
  const resizeRef = useRef<HTMLDivElement>(null)
  const isDraggingRef = useRef(false)
  const tabCounter = useRef(0)

  // Sync tabs from useTerminal hook
  useEffect(() => {
    setAllTabs(tabs)
  }, [tabs])

  // Create initial tab on mount
  useEffect(() => {
    if (tabs.length === 0) {
      createTab({ cwd: projectPath })
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd+` to toggle terminal
      if (e.metaKey && e.key === '`') {
        e.preventDefault()
        toggleExpanded()
      }
      // Cmd+T to new tab (only when terminal is focused/expanded)
      if (e.metaKey && e.key === 't' && isExpanded) {
        e.preventDefault()
        createTab({ cwd: projectPath })
      }
      // Cmd+W to close tab (only when terminal is focused/expanded)
      if (e.metaKey && e.key === 'w' && isExpanded && activeTabId) {
        e.preventDefault()
        closeTab(activeTabId)
      }
      // Cmd+Shift+] to next tab
      if (e.metaKey && e.shiftKey && e.key === ']' && tabs.length > 1) {
        e.preventDefault()
        const currentIndex = tabs.findIndex((t) => t.id === activeTabId)
        const nextIndex = (currentIndex + 1) % tabs.length
        setActiveTab(tabs[nextIndex].id)
      }
      // Cmd+Shift+[ to previous tab
      if (e.metaKey && e.shiftKey && e.key === '[' && tabs.length > 1) {
        e.preventDefault()
        const currentIndex = tabs.findIndex((t) => t.id === activeTabId)
        const prevIndex = (currentIndex - 1 + tabs.length) % tabs.length
        setActiveTab(tabs[prevIndex].id)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [tabs, activeTabId, isExpanded, createTab, closeTab, setActiveTab, toggleExpanded, projectPath])

  // Handle resize drag
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    isDraggingRef.current = true

    const startY = e.clientY
    const startHeight = height

    const handleMouseMove = (e: MouseEvent) => {
      if (!isDraggingRef.current) return

      const deltaY = startY - e.clientY
      const newHeight = Math.min(MAX_HEIGHT, Math.max(MIN_HEIGHT, startHeight + deltaY))
      onHeightChange(newHeight)
    }

    const handleMouseUp = () => {
      isDraggingRef.current = false
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
  }, [height, onHeightChange])

  const handleTabExit = useCallback((tabId: string, code: number) => {
    // Optionally auto-close tab on exit
    console.log(`Tab ${tabId} exited with code ${code}`)
  }, [])

  const handleNewTab = useCallback(() => {
    createTab({ cwd: projectPath })
  }, [createTab, projectPath])

  const handleTmuxAttach = useCallback(async (sessionName: string) => {
    try {
      const session = await tmuxAttach(sessionName)
      tabCounter.current += 1
      const newTab: TerminalTab = {
        id: `tmux-tab-${tabCounter.current}`,
        sessionId: session.id,
        title: sessionName,
        isTmux: true,
        tmuxSession: sessionName,
      }
      setAllTabs((prev) => [...prev, newTab])
      setActiveTab(newTab.id)
    } catch (err) {
      console.error('Failed to attach to tmux session:', err)
    }
  }, [tmuxAttach, setActiveTab])

  const handleTmuxCreate = useCallback(async (sessionName: string) => {
    try {
      const session = await tmuxCreate(sessionName)
      tabCounter.current += 1
      const newTab: TerminalTab = {
        id: `tmux-tab-${tabCounter.current}`,
        sessionId: session.id,
        title: sessionName,
        isTmux: true,
        tmuxSession: sessionName,
      }
      setAllTabs((prev) => [...prev, newTab])
      setActiveTab(newTab.id)
    } catch (err) {
      console.error('Failed to create tmux session:', err)
    }
  }, [tmuxCreate, setActiveTab])

  const handleToggleWorkflowOverlay = useCallback(() => {
    setShowWorkflowOverlay((prev) => !prev)
  }, [])

  if (!isExpanded) {
    // Collapsed state - just show tab bar
    return (
      <div className="border-t border-va-border bg-va-sidebar">
        <TerminalTabs
          tabs={allTabs.length > 0 ? allTabs : tabs}
          activeTabId={activeTabId}
          onTabSelect={setActiveTab}
          onTabClose={closeTab}
          onNewTab={handleNewTab}
          onToggleExpand={toggleExpanded}
          isExpanded={isExpanded}
          tmuxSessions={tmuxSessions}
          tmuxAvailable={tmuxAvailable}
          onTmuxAttach={handleTmuxAttach}
          onTmuxCreate={handleTmuxCreate}
          showWorkflowOverlay={showWorkflowOverlay}
          onToggleWorkflowOverlay={projectName ? handleToggleWorkflowOverlay : undefined}
        />
      </div>
    )
  }

  return (
    <div
      className="flex flex-col border-t border-va-border bg-va-panel"
      style={{ height }}
    >
      {/* Resize handle */}
      <div
        ref={resizeRef}
        className="h-1 cursor-ns-resize bg-va-border hover:bg-va-accent transition-colors"
        onMouseDown={handleMouseDown}
      />

      {/* Tab bar */}
      <TerminalTabs
        tabs={allTabs.length > 0 ? allTabs : tabs}
        activeTabId={activeTabId}
        onTabSelect={setActiveTab}
        onTabClose={closeTab}
        onNewTab={handleNewTab}
        onToggleExpand={toggleExpanded}
        isExpanded={isExpanded}
        tmuxSessions={tmuxSessions}
        tmuxAvailable={tmuxAvailable}
        onTmuxAttach={handleTmuxAttach}
        onTmuxCreate={handleTmuxCreate}
        showWorkflowOverlay={showWorkflowOverlay}
        onToggleWorkflowOverlay={projectName ? handleToggleWorkflowOverlay : undefined}
      />

      {/* Terminal content */}
      <div className="flex-1 overflow-hidden relative">
        {/* Workflow Overlay */}
        {showWorkflowOverlay && projectName && (
          <WorkflowOverlay
            projectName={projectName}
            onClose={() => setShowWorkflowOverlay(false)}
          />
        )}

        {(allTabs.length > 0 ? allTabs : tabs).map((tab) => (
          <div
            key={tab.id}
            className={`absolute inset-0 ${tab.id === activeTabId ? 'block' : 'hidden'}`}
          >
            <TerminalInstance
              sessionId={tab.sessionId}
              onTitleChange={(title) => updateTabTitle(tab.id, title)}
              onExit={(code) => handleTabExit(tab.id, code)}
            />
          </div>
        ))}

        {(allTabs.length > 0 ? allTabs : tabs).length === 0 && (
          <div className="flex items-center justify-center h-full text-va-text-muted">
            <button
              onClick={handleNewTab}
              className="flex items-center space-x-2 px-4 py-2 bg-va-sidebar hover:bg-va-border rounded"
            >
              <span>New Terminal</span>
              <span className="text-xs opacity-60">Cmd+T</span>
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

export { DEFAULT_HEIGHT as DEFAULT_TERMINAL_HEIGHT }
