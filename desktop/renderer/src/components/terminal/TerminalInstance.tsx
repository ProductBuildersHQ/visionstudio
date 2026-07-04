import { useEffect, useRef, useCallback } from 'react'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import 'xterm/css/xterm.css'

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
  const cleanupRef = useRef<(() => void)[]>([])

  // Handle resize
  const handleResize = useCallback(() => {
    if (fitAddonRef.current && terminalRef.current) {
      try {
        fitAddonRef.current.fit()
        const { cols, rows } = terminalRef.current
        window.electronAPI.terminal.resize(sessionId, cols, rows)
      } catch {
        // Ignore resize errors during teardown
      }
    }
  }, [sessionId])

  useEffect(() => {
    if (!containerRef.current) return

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

    terminal.open(containerRef.current)
    fitAddon.fit()

    terminalRef.current = terminal
    fitAddonRef.current = fitAddon

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

    // Handle window resize
    const resizeObserver = new ResizeObserver(handleResize)
    resizeObserver.observe(containerRef.current)

    // Initial size sync
    const { cols, rows } = terminal
    window.electronAPI.terminal.resize(sessionId, cols, rows)

    // Store cleanup functions
    cleanupRef.current = [
      () => inputDisposable.dispose(),
      () => titleDisposable.dispose(),
      dataCleanup,
      exitCleanup,
      () => resizeObserver.disconnect(),
      () => terminal.dispose(),
    ]

    return () => {
      cleanupRef.current.forEach((cleanup) => cleanup())
      cleanupRef.current = []
      terminalRef.current = null
      fitAddonRef.current = null
    }
  }, [sessionId, onTitleChange, onExit, handleResize])

  return (
    <div
      ref={containerRef}
      className={`terminal-container ${className || ''}`}
      style={{
        width: '100%',
        height: '100%',
        backgroundColor: TERMINAL_THEME.background,
      }}
    />
  )
}
