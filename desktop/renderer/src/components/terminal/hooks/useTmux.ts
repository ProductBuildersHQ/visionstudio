import { useState, useEffect, useCallback } from 'react'
import type { TmuxSession, PTYSession } from '../../../types/terminal'

interface UseTmuxResult {
  sessions: TmuxSession[]
  isAvailable: boolean
  isLoading: boolean
  error: string | null
  refresh: () => Promise<void>
  attach: (sessionName: string) => Promise<PTYSession>
  create: (sessionName: string, command?: string) => Promise<PTYSession>
}

export function useTmux(): UseTmuxResult {
  const [sessions, setSessions] = useState<TmuxSession[]>([])
  const [isAvailable, setIsAvailable] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const checkAvailability = useCallback(async () => {
    try {
      const available = await window.electronAPI.tmux.isAvailable()
      setIsAvailable(available)
      return available
    } catch (err) {
      setIsAvailable(false)
      return false
    }
  }, [])

  const refresh = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const available = await checkAvailability()
      if (!available) {
        setSessions([])
        return
      }

      const sessionList = await window.electronAPI.tmux.list()
      setSessions(sessionList)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to list tmux sessions')
      setSessions([])
    } finally {
      setIsLoading(false)
    }
  }, [checkAvailability])

  const attach = useCallback(async (sessionName: string): Promise<PTYSession> => {
    if (!isAvailable) {
      throw new Error('tmux is not available')
    }
    return window.electronAPI.tmux.attach(sessionName)
  }, [isAvailable])

  const create = useCallback(async (sessionName: string, command?: string): Promise<PTYSession> => {
    if (!isAvailable) {
      throw new Error('tmux is not available')
    }
    const session = await window.electronAPI.tmux.spawn(sessionName, command)
    // Refresh the session list after creating
    await refresh()
    return session
  }, [isAvailable, refresh])

  // Initial load
  useEffect(() => {
    refresh()
  }, [refresh])

  // Poll for session changes every 10 seconds
  useEffect(() => {
    if (!isAvailable) return

    const interval = setInterval(refresh, 10000)
    return () => clearInterval(interval)
  }, [isAvailable, refresh])

  return {
    sessions,
    isAvailable,
    isLoading,
    error,
    refresh,
    attach,
    create,
  }
}

// Helper to generate project-specific session names
export function getProjectSessionName(projectName: string): string {
  const sanitized = projectName
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 32)

  return `visionstudio-${sanitized}`
}
