import { contextBridge, ipcRenderer } from 'electron'

// Debug: Listen to ALL IPC events
ipcRenderer.on('terminal:data', (event, payload) => {
  console.log('[Preload DEBUG] Raw terminal:data event received:', payload)
})

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
      console.log('[Preload] onData registering listener')
      const channel = 'terminal:data'
      const handler = (_event: Electron.IpcRendererEvent, payload: { id: string; data: string }) => {
        console.log('[Preload] Handler called! payload:', JSON.stringify(payload).slice(0, 100))
        try {
          callback(payload.id, payload.data)
        } catch (err) {
          console.error('[Preload] Callback error:', err)
        }
      }
      ipcRenderer.on(channel, handler)
      console.log('[Preload] onData listener registered for channel:', channel)
      // Return cleanup function
      return () => {
        console.log('[Preload] Removing listener for channel:', channel)
        ipcRenderer.removeListener(channel, handler)
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
