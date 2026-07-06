import { useEffect, useRef, useCallback } from 'react'

// Event types matching the Go backend
export type FileEventType =
  | 'connected'
  | 'file_created'
  | 'file_modified'
  | 'file_deleted'
  | 'file_renamed'
  | 'eval_started'
  | 'eval_complete'
  | 'lint_complete'
  | 'spec_updated'
  | 'workflow_changed'

export interface FileEvent {
  type: FileEventType
  project: string
  path?: string
  specType?: string
  timestamp: string
  data?: Record<string, unknown>
}

export interface UseProjectEventsOptions {
  onEvent?: (event: FileEvent) => void
  onFileChanged?: (event: FileEvent) => void
  onEvalComplete?: (event: FileEvent) => void
  onWorkflowChanged?: (event: FileEvent) => void
  onConnected?: () => void
  onDisconnected?: () => void
  enabled?: boolean
}

const API_BASE = 'http://127.0.0.1:8765/api'

/**
 * Hook to subscribe to real-time project events via SSE
 */
export function useProjectEvents(
  projectName: string | null | undefined,
  options: UseProjectEventsOptions = {}
) {
  const {
    onEvent,
    onFileChanged,
    onEvalComplete,
    onWorkflowChanged,
    onConnected,
    onDisconnected,
    enabled = true,
  } = options

  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  const connect = useCallback(() => {
    if (!projectName || !enabled) return

    // Clean up existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    const url = `${API_BASE}/projects/${encodeURIComponent(projectName)}/events`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    // Handle all event types
    const eventTypes: FileEventType[] = [
      'connected',
      'file_created',
      'file_modified',
      'file_deleted',
      'file_renamed',
      'eval_started',
      'eval_complete',
      'lint_complete',
      'spec_updated',
      'workflow_changed',
    ]

    eventTypes.forEach((eventType) => {
      eventSource.addEventListener(eventType, (e: MessageEvent) => {
        try {
          const event: FileEvent = JSON.parse(e.data)

          // Call generic handler
          onEvent?.(event)

          // Call specific handlers
          switch (event.type) {
            case 'connected':
              onConnected?.()
              break
            case 'file_created':
            case 'file_modified':
            case 'file_deleted':
            case 'file_renamed':
              onFileChanged?.(event)
              break
            case 'eval_complete':
              onEvalComplete?.(event)
              break
            case 'workflow_changed':
              onWorkflowChanged?.(event)
              break
          }
        } catch (err) {
          console.error('Failed to parse SSE event:', err)
        }
      })
    })

    eventSource.onerror = () => {
      console.warn('SSE connection error, will reconnect...')
      onDisconnected?.()
      eventSource.close()

      // Reconnect after delay
      reconnectTimeoutRef.current = setTimeout(() => {
        connect()
      }, 3000)
    }
  }, [projectName, enabled, onEvent, onFileChanged, onEvalComplete, onWorkflowChanged, onConnected, onDisconnected])

  useEffect(() => {
    connect()

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
    }
  }, [connect])

  // Return a function to manually reconnect
  return { reconnect: connect }
}
