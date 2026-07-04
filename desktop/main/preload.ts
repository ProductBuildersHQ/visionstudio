import { contextBridge, ipcRenderer } from 'electron'

// Type definitions for IPC communication
interface SpawnOptions {
  cwd?: string
  shell?: string
  cols?: number
  rows?: number
  env?: Record<string, string>
}

interface PTYSession {
  id: string
  pid: number
  shell: string
  cwd: string
  cols: number
  rows: number
  createdAt: string
}

interface TmuxSession {
  name: string
  windows: number
  created: string
  attached: boolean
}

// Expose terminal API to renderer via contextBridge
contextBridge.exposeInMainWorld('electronAPI', {
  terminal: {
    spawn: (options: SpawnOptions): Promise<PTYSession> => {
      return ipcRenderer.invoke('terminal:spawn', options)
    },

    write: (id: string, data: string): void => {
      ipcRenderer.send('terminal:write', { id, data })
    },

    resize: (id: string, cols: number, rows: number): void => {
      ipcRenderer.send('terminal:resize', { id, cols, rows })
    },

    kill: (id: string): Promise<void> => {
      return ipcRenderer.invoke('terminal:kill', { id })
    },

    onData: (callback: (id: string, data: string) => void): (() => void) => {
      const handler = (_event: Electron.IpcRendererEvent, data: { id: string; data: string }) => {
        callback(data.id, data.data)
      }
      ipcRenderer.on('terminal:data', handler)
      // Return cleanup function
      return () => {
        ipcRenderer.removeListener('terminal:data', handler)
      }
    },

    onExit: (callback: (id: string, code: number) => void): (() => void) => {
      const handler = (_event: Electron.IpcRendererEvent, data: { id: string; code: number }) => {
        callback(data.id, data.code)
      }
      ipcRenderer.on('terminal:exit', handler)
      // Return cleanup function
      return () => {
        ipcRenderer.removeListener('terminal:exit', handler)
      }
    },
  },

  tmux: {
    list: (): Promise<TmuxSession[]> => {
      return ipcRenderer.invoke('tmux:list')
    },

    attach: (name: string): Promise<PTYSession> => {
      return ipcRenderer.invoke('tmux:attach', { name })
    },

    spawn: (name: string, command?: string): Promise<PTYSession> => {
      return ipcRenderer.invoke('tmux:spawn', { name, command })
    },

    isAvailable: (): Promise<boolean> => {
      return ipcRenderer.invoke('tmux:isAvailable')
    },
  },
})
