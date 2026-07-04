import { ipcMain, IpcMainInvokeEvent } from 'electron'
import { PTYManager, SpawnOptions } from '../pty/manager.js'
import { TmuxManager } from '../pty/tmux.js'

export function registerTerminalIPC(ptyManager: PTYManager, tmuxManager: TmuxManager): void {
  // Terminal handlers
  ipcMain.handle('terminal:spawn', async (_event: IpcMainInvokeEvent, options: SpawnOptions) => {
    return ptyManager.spawn(options)
  })

  ipcMain.on('terminal:write', (_event, { id, data }: { id: string; data: string }) => {
    ptyManager.write(id, data)
  })

  ipcMain.on('terminal:resize', (_event, { id, cols, rows }: { id: string; cols: number; rows: number }) => {
    ptyManager.resize(id, cols, rows)
  })

  ipcMain.handle('terminal:kill', async (_event: IpcMainInvokeEvent, { id }: { id: string }) => {
    await ptyManager.kill(id)
  })

  // tmux handlers
  ipcMain.handle('tmux:list', async () => {
    return tmuxManager.listSessions()
  })

  ipcMain.handle('tmux:attach', async (_event: IpcMainInvokeEvent, { name }: { name: string }) => {
    return tmuxManager.attach(name, ptyManager)
  })

  ipcMain.handle('tmux:spawn', async (_event: IpcMainInvokeEvent, { name, command }: { name: string; command?: string }) => {
    return tmuxManager.spawn(name, ptyManager, command)
  })

  ipcMain.handle('tmux:isAvailable', async () => {
    return tmuxManager.isAvailable()
  })
}
