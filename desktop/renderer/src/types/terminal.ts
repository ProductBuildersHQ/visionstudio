// Terminal types for IPC communication

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

export interface TmuxSession {
  name: string
  windows: number
  created: string
  attached: boolean
}

export interface TerminalTab {
  id: string
  sessionId: string
  title: string
  isTmux: boolean
  tmuxSession?: string
}

// IPC message types
export interface TerminalDataMessage {
  id: string
  data: string
}

export interface TerminalExitMessage {
  id: string
  code: number
}

export interface TerminalResizeMessage {
  id: string
  cols: number
  rows: number
}

// Electron API interface exposed via preload
export interface ElectronTerminalAPI {
  spawn: (options: SpawnOptions) => Promise<PTYSession>
  write: (id: string, data: string) => void
  resize: (id: string, cols: number, rows: number) => void
  kill: (id: string) => Promise<void>
  onData: (callback: (id: string, data: string) => void) => () => void
  onExit: (callback: (id: string, code: number) => void) => () => void
}

export interface ElectronTmuxAPI {
  list: () => Promise<TmuxSession[]>
  attach: (name: string) => Promise<PTYSession>
  spawn: (name: string, command?: string) => Promise<PTYSession>
  isAvailable: () => Promise<boolean>
}

export interface ElectronAPI {
  terminal: ElectronTerminalAPI
  tmux: ElectronTmuxAPI
}

// Extend Window interface
declare global {
  interface Window {
    electronAPI: ElectronAPI
  }
}
