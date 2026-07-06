import * as pty from 'node-pty'
import { BrowserWindow } from 'electron'
import { randomUUID } from 'crypto'

export interface SpawnOptions {
  cwd?: string
  shell?: string
  cols?: number
  rows?: number
  env?: Record<string, string>
}

export interface PTYSession {
  id: string
  pid: number
  shell: string
  cwd: string
  cols: number
  rows: number
  createdAt: string
}

interface ActiveSession {
  pty: pty.IPty
  info: PTYSession
}

const MAX_SESSIONS = 8

export class PTYManager {
  private sessions: Map<string, ActiveSession> = new Map()
  private mainWindow: BrowserWindow | null = null

  setMainWindow(window: BrowserWindow | null): void {
    this.mainWindow = window
  }

  async spawn(options: SpawnOptions): Promise<PTYSession> {
    if (this.sessions.size >= MAX_SESSIONS) {
      throw new Error(`Maximum sessions (${MAX_SESSIONS}) reached`)
    }

    const shell = options.shell || process.env.SHELL || '/bin/zsh'
    let cwd = options.cwd || process.env.HOME || '/'
    const cols = options.cols || 80
    const rows = options.rows || 24

    // Convert relative paths to absolute (relative to HOME)
    if (!cwd.startsWith('/')) {
      cwd = process.env.HOME || '/'
    }

    const id = randomUUID()

    const ptyProcess = pty.spawn(shell, [], {
      name: 'xterm-256color',
      cols,
      rows,
      cwd,
      env: {
        ...process.env,
        ...options.env,
        TERM: 'xterm-256color',
        COLORTERM: 'truecolor',
      },
    })

    const sessionInfo: PTYSession = {
      id,
      pid: ptyProcess.pid,
      shell,
      cwd,
      cols,
      rows,
      createdAt: new Date().toISOString(),
    }

    this.sessions.set(id, {
      pty: ptyProcess,
      info: sessionInfo,
    })

    // Handle PTY output
    ptyProcess.onData((data: string) => {
      console.log('[PTYManager] Data from PTY:', id, 'length:', data.length, 'mainWindow:', !!this.mainWindow)
      if (this.mainWindow && !this.mainWindow.isDestroyed()) {
        console.log('[PTYManager] Sending to renderer via IPC, channel: terminal:data')
        try {
          this.mainWindow.webContents.send('terminal:data', { id, data })
          console.log('[PTYManager] IPC send completed')
        } catch (err) {
          console.error('[PTYManager] IPC send error:', err)
        }
      } else {
        console.log('[PTYManager] Cannot send - mainWindow is null or destroyed')
      }
    })

    // Handle PTY exit
    ptyProcess.onExit(({ exitCode }) => {
      if (this.mainWindow && !this.mainWindow.isDestroyed()) {
        this.mainWindow.webContents.send('terminal:exit', { id, code: exitCode })
      }
      this.sessions.delete(id)
    })

    return sessionInfo
  }

  write(sessionId: string, data: string): void {
    const session = this.sessions.get(sessionId)
    if (!session) {
      console.warn(`PTYManager: Session ${sessionId} not found`)
      return
    }
    session.pty.write(data)
  }

  resize(sessionId: string, cols: number, rows: number): void {
    const session = this.sessions.get(sessionId)
    if (!session) {
      console.warn(`PTYManager: Session ${sessionId} not found`)
      return
    }
    session.pty.resize(cols, rows)
    session.info.cols = cols
    session.info.rows = rows
  }

  async kill(sessionId: string): Promise<void> {
    const session = this.sessions.get(sessionId)
    if (!session) {
      console.warn(`PTYManager: Session ${sessionId} not found`)
      return
    }
    session.pty.kill()
    this.sessions.delete(sessionId)
  }

  listSessions(): PTYSession[] {
    return Array.from(this.sessions.values()).map((s) => s.info)
  }

  killAll(): void {
    for (const [id, session] of this.sessions) {
      try {
        session.pty.kill()
      } catch (err) {
        console.error(`Failed to kill session ${id}:`, err)
      }
    }
    this.sessions.clear()
  }

  getSession(sessionId: string): PTYSession | undefined {
    return this.sessions.get(sessionId)?.info
  }
}
