import { execSync, exec } from 'child_process'
import { PTYManager, PTYSession } from './manager.js'

export interface TmuxSession {
  name: string
  windows: number
  created: string
  attached: boolean
}

// Validate session names to prevent injection
const SESSION_NAME_PATTERN = /^[a-zA-Z0-9_-]+$/

export class TmuxManager {
  private tmuxPath: string | null = null
  private available: boolean | null = null

  constructor() {
    this.detectTmux()
  }

  private detectTmux(): void {
    try {
      this.tmuxPath = execSync('which tmux', { encoding: 'utf-8' }).trim()
      this.available = true
    } catch {
      this.tmuxPath = null
      this.available = false
    }
  }

  isAvailable(): boolean {
    if (this.available === null) {
      this.detectTmux()
    }
    return this.available === true
  }

  async listSessions(): Promise<TmuxSession[]> {
    if (!this.isAvailable()) {
      return []
    }

    return new Promise((resolve) => {
      exec('tmux list-sessions -F "#{session_name}|#{session_windows}|#{session_created}|#{session_attached}"', (error, stdout) => {
        if (error) {
          // No sessions or tmux not running
          resolve([])
          return
        }

        const sessions = stdout
          .trim()
          .split('\n')
          .filter((line) => line.length > 0)
          .map((line) => {
            const [name, windows, created, attached] = line.split('|')
            return {
              name,
              windows: parseInt(windows, 10) || 1,
              created: new Date(parseInt(created, 10) * 1000).toISOString(),
              attached: attached === '1',
            }
          })

        resolve(sessions)
      })
    })
  }

  async hasSession(name: string): Promise<boolean> {
    const sessions = await this.listSessions()
    return sessions.some((s) => s.name === name)
  }

  private validateSessionName(name: string): void {
    if (!SESSION_NAME_PATTERN.test(name)) {
      throw new Error(`Invalid session name: ${name}. Only alphanumeric, underscore, and hyphen allowed.`)
    }
    if (name.length > 64) {
      throw new Error('Session name too long (max 64 characters)')
    }
  }

  async attach(sessionName: string, ptyManager: PTYManager): Promise<PTYSession> {
    this.validateSessionName(sessionName)

    if (!this.isAvailable()) {
      throw new Error('tmux is not available')
    }

    const exists = await this.hasSession(sessionName)
    if (!exists) {
      throw new Error(`tmux session '${sessionName}' does not exist`)
    }

    // Spawn a PTY that attaches to the tmux session
    return ptyManager.spawn({
      shell: this.tmuxPath!,
      cwd: process.env.HOME,
    }).then((session) => {
      // Send attach command to the PTY
      ptyManager.write(session.id, `attach-session -t ${sessionName}\n`)
      return session
    })
  }

  async spawn(sessionName: string, ptyManager: PTYManager, command?: string): Promise<PTYSession> {
    this.validateSessionName(sessionName)

    if (!this.isAvailable()) {
      throw new Error('tmux is not available')
    }

    const exists = await this.hasSession(sessionName)

    // Spawn a PTY running tmux
    const session = await ptyManager.spawn({
      shell: this.tmuxPath!,
      cwd: process.env.HOME,
    })

    if (exists) {
      // Attach to existing session
      ptyManager.write(session.id, `attach-session -t ${sessionName}\n`)
    } else {
      // Create new session
      const cmdPart = command ? ` "${command.replace(/"/g, '\\"')}"` : ''
      ptyManager.write(session.id, `new-session -s ${sessionName}${cmdPart}\n`)
    }

    return session
  }

  getProjectSessionName(projectName: string): string {
    // Sanitize project name for tmux session
    const sanitized = projectName
      .toLowerCase()
      .replace(/[^a-z0-9-]/g, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '')
      .slice(0, 32)

    return `visionstudio-${sanitized}`
  }
}
