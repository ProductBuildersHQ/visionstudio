import { useEffect, useRef } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

interface TerminalInstanceProps {
  sessionId: string
  onTitleChange?: (title: string) => void
  onExit?: (code: number) => void
  className?: string
}

// VisionStudio theme colors matching the app
const TERMINAL_THEME = {
  background: '#1b2636',
  foreground: '#e0e6ed',
  cursor: '#3b82f6',
  cursorAccent: '#1b2636',
  selectionBackground: 'rgba(59, 130, 246, 0.3)',
  selectionForeground: '#e0e6ed',
  black: '#1b2636',
  red: '#ef4444',
  green: '#22c55e',
  yellow: '#f59e0b',
  blue: '#3b82f6',
  magenta: '#a855f7',
  cyan: '#06b6d4',
  white: '#e0e6ed',
  brightBlack: '#8899a6',
  brightRed: '#f87171',
  brightGreen: '#4ade80',
  brightYellow: '#fbbf24',
  brightBlue: '#60a5fa',
  brightMagenta: '#c084fc',
  brightCyan: '#22d3ee',
  brightWhite: '#ffffff',
}

export function TerminalInstance({ sessionId, onTitleChange, onExit, className }: TerminalInstanceProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const terminalRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const isInitializedRef = useRef(false)
  const isMountedRef = useRef(true)
  const cleanupFnsRef = useRef<(() => void)[]>([])

  useEffect(() => {
    isMountedRef.current = true

    if (!containerRef.current || isInitializedRef.current) return

    const container = containerRef.current
    let initRetryCount = 0
    const maxRetries = 20
    let initTimeout: ReturnType<typeof setTimeout>

    // Wait for container to have valid dimensions before initializing
    const tryInit = () => {
      if (isInitializedRef.current) return

      const { offsetWidth, offsetHeight } = container
      console.log('[TerminalInstance] tryInit - dimensions:', offsetWidth, 'x', offsetHeight, 'retry:', initRetryCount)
      if ((offsetWidth === 0 || offsetHeight === 0) && initRetryCount < maxRetries) {
        initRetryCount++
        initTimeout = setTimeout(tryInit, 50)
        return
      }

      isInitializedRef.current = true
      initTerminal()
    }

    const initTerminal = () => {
      console.log('[TerminalInstance] initTerminal called for session:', sessionId)
      // Initialize xterm.js
      const terminal = new Terminal({
        theme: TERMINAL_THEME,
        fontFamily: 'SF Mono, Menlo, Monaco, Consolas, monospace',
        fontSize: 13,
        lineHeight: 1.2,
        cursorBlink: true,
        cursorStyle: 'block',
        scrollback: 5000,
        allowProposedApi: true,
      })

      const fitAddon = new FitAddon()
      const webLinksAddon = new WebLinksAddon()

      terminal.loadAddon(fitAddon)
      terminal.loadAddon(webLinksAddon)

      // Open terminal
      terminal.open(container)
      console.log('[TerminalInstance] Terminal opened in container')

      terminalRef.current = terminal
      fitAddonRef.current = fitAddon

      // Debounced resize handler
      let resizeTimeout: ReturnType<typeof setTimeout> | null = null
      const handleResize = () => {
        if (resizeTimeout) clearTimeout(resizeTimeout)
        resizeTimeout = setTimeout(() => {
          if (fitAddonRef.current && terminalRef.current && container) {
            const { offsetWidth, offsetHeight } = container
            if (offsetWidth === 0 || offsetHeight === 0) return

            try {
              fitAddonRef.current.fit()
              const { cols, rows } = terminalRef.current
              window.electronAPI.terminal.resize(sessionId, cols, rows)
            } catch {
              // Ignore resize errors
            }
          }
        }, 50)
      }

      // Initial fit after a short delay
      const fitTimeout = setTimeout(() => {
        try {
          fitAddon.fit()
          const { cols, rows } = terminal
          window.electronAPI.terminal.resize(sessionId, cols, rows)
        } catch {
          // Ignore initial fit errors
        }
      }, 100)

      // Set up resize observer
      const resizeObserver = new ResizeObserver(handleResize)
      resizeObserver.observe(container)

      // Handle terminal input -> PTY
      const inputDisposable = terminal.onData((data) => {
        window.electronAPI.terminal.write(sessionId, data)
      })

      // Handle title changes
      const titleDisposable = terminal.onTitleChange((title) => {
        onTitleChange?.(title)
      })

      // Handle PTY output -> terminal
      const dataCleanup = window.electronAPI.terminal.onData((id, data) => {
        if (id === sessionId) {
          console.log('[TerminalInstance] Received data for session:', id, 'length:', data.length)
          terminal.write(data)
        }
      })

      // Handle PTY exit
      const exitCleanup = window.electronAPI.terminal.onExit((id, code) => {
        if (id === sessionId) {
          terminal.write(`\r\n\x1b[90m[Process exited with code ${code}]\x1b[0m\r\n`)
          onExit?.(code)
        }
      })

      // Store cleanup functions
      cleanupFnsRef.current = [
        () => clearTimeout(fitTimeout),
        () => resizeTimeout && clearTimeout(resizeTimeout),
        () => inputDisposable.dispose(),
        () => titleDisposable.dispose(),
        dataCleanup,
        exitCleanup,
        () => resizeObserver.disconnect(),
        () => terminal.dispose(),
      ]
    }

    // Start initialization
    tryInit()

    return () => {
      isMountedRef.current = false
      clearTimeout(initTimeout)
      // Delay cleanup to handle React Strict Mode remounting
      setTimeout(() => {
        if (!isMountedRef.current) {
          cleanupFnsRef.current.forEach((fn) => fn())
          cleanupFnsRef.current = []
          terminalRef.current = null
          fitAddonRef.current = null
          // Don't reset isInitializedRef here to prevent re-init on HMR
        }
      }, 100)
    }
  }, [sessionId, onTitleChange, onExit])

  return (
    <div
      ref={containerRef}
      className={`terminal-container ${className || ''}`}
      style={{
        width: '100%',
        height: '100%',
        minWidth: 200,
        minHeight: 100,
        backgroundColor: TERMINAL_THEME.background,
      }}
    />
  )
}
