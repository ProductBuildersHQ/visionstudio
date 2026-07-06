import { useState, useCallback, useRef, useEffect } from 'react'
import type { TerminalTab, PTYSession, SpawnOptions } from '../../../types/terminal'

const MAX_TABS = 8

// Check if we're running in Electron with preload
const isElectron = (): boolean => {
  return typeof window !== 'undefined' && window.electronAPI !== undefined
}

interface UseTerminalResult {
  tabs: TerminalTab[]
  activeTabId: string | null
  isExpanded: boolean
  createTab: (options?: SpawnOptions) => Promise<void>
  closeTab: (tabId: string) => void
  setActiveTab: (tabId: string) => void
  toggleExpanded: () => void
  setExpanded: (expanded: boolean) => void
  updateTabTitle: (tabId: string, title: string) => void
}

export function useTerminal(): UseTerminalResult {
  const [tabs, setTabs] = useState<TerminalTab[]>([])
  const [activeTabId, setActiveTabId] = useState<string | null>(null)
  const [isExpanded, setIsExpanded] = useState(true)
  const tabCounter = useRef(0)

  const createTab = useCallback(async (options?: SpawnOptions) => {
    console.log('[useTerminal] createTab called, isElectron:', isElectron())
    if (!isElectron()) {
      console.error('Terminal requires Electron environment with preload script')
      return
    }

    if (tabs.length >= MAX_TABS) {
      console.warn(`Maximum tabs (${MAX_TABS}) reached`)
      return
    }

    try {
      console.log('[useTerminal] Spawning terminal with options:', options)
      const session: PTYSession = await window.electronAPI.terminal.spawn({
        cwd: options?.cwd || undefined,
        shell: options?.shell,
        cols: options?.cols || 80,
        rows: options?.rows || 24,
        env: options?.env,
      })
      console.log('[useTerminal] Session created:', session.id)

      tabCounter.current += 1
      const newTab: TerminalTab = {
        id: `tab-${tabCounter.current}`,
        sessionId: session.id,
        title: `Terminal ${tabCounter.current}`,
        isTmux: false,
      }

      setTabs((prev) => [...prev, newTab])
      setActiveTabId(newTab.id)
    } catch (err) {
      console.error('Failed to create terminal tab:', err)
    }
  }, [tabs.length])

  const closeTab = useCallback((tabId: string) => {
    const tab = tabs.find((t) => t.id === tabId)
    if (!tab) return

    // Kill the PTY session
    if (isElectron()) {
      window.electronAPI.terminal.kill(tab.sessionId).catch((err) => {
        console.error('Failed to kill terminal session:', err)
      })
    }

    setTabs((prev) => {
      const newTabs = prev.filter((t) => t.id !== tabId)

      // If we closed the active tab, activate another
      if (activeTabId === tabId && newTabs.length > 0) {
        const closedIndex = prev.findIndex((t) => t.id === tabId)
        const newActiveIndex = Math.min(closedIndex, newTabs.length - 1)
        setActiveTabId(newTabs[newActiveIndex].id)
      } else if (newTabs.length === 0) {
        setActiveTabId(null)
      }

      return newTabs
    })
  }, [tabs, activeTabId])

  const setActiveTab = useCallback((tabId: string) => {
    setActiveTabId(tabId)
  }, [])

  const toggleExpanded = useCallback(() => {
    setIsExpanded((prev) => !prev)
  }, [])

  const updateTabTitle = useCallback((tabId: string, title: string) => {
    setTabs((prev) =>
      prev.map((tab) =>
        tab.id === tabId ? { ...tab, title: title || tab.title } : tab
      )
    )
  }, [])

  // Clean up all sessions on unmount
  useEffect(() => {
    return () => {
      if (isElectron()) {
        tabs.forEach((tab) => {
          window.electronAPI.terminal.kill(tab.sessionId).catch(() => {})
        })
      }
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return {
    tabs,
    activeTabId,
    isExpanded,
    createTab,
    closeTab,
    setActiveTab,
    toggleExpanded,
    setExpanded: setIsExpanded,
    updateTabTitle,
  }
}
