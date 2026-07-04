import { app, BrowserWindow, shell } from 'electron'
import path from 'path'
import { spawn, ChildProcess } from 'child_process'
import { fileURLToPath } from 'url'
import { PTYManager } from './pty/manager.js'
import { TmuxManager } from './pty/tmux.js'
import { registerTerminalIPC } from './ipc/terminal.js'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

let mainWindow: BrowserWindow | null = null
let daemonProcess: ChildProcess | null = null

// Terminal infrastructure
const ptyManager = new PTYManager()
const tmuxManager = new TmuxManager()

const DAEMON_PORT = 8765
const isDev = process.env.NODE_ENV !== 'production'

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 1000,
    minHeight: 600,
    title: 'VisionStudio',
    backgroundColor: '#1b2636',
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js'),
    },
    titleBarStyle: 'hiddenInset',
    trafficLightPosition: { x: 16, y: 16 },
  })

  // Load the app
  if (isDev) {
    mainWindow.loadURL('http://localhost:5173')
    mainWindow.webContents.openDevTools()
  } else {
    mainWindow.loadFile(path.join(__dirname, '../renderer/dist/index.html'))
  }

  // Open external links in browser
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url)
    return { action: 'deny' }
  })

  mainWindow.on('closed', () => {
    mainWindow = null
    ptyManager.setMainWindow(null)
  })

  // Set up terminal infrastructure
  ptyManager.setMainWindow(mainWindow)
}

function startDaemon() {
  // Path to the Go daemon binary
  const daemonPath = isDev
    ? path.join(__dirname, '../../bin/daemon')
    : path.join(process.resourcesPath!, 'bin/daemon')

  console.log(`Starting daemon at ${daemonPath}`)

  daemonProcess = spawn(daemonPath, ['--port', String(DAEMON_PORT)], {
    stdio: ['ignore', 'pipe', 'pipe'],
  })

  daemonProcess.stdout?.on('data', (data: Buffer) => {
    console.log(`[daemon] ${data}`)
  })

  daemonProcess.stderr?.on('data', (data: Buffer) => {
    console.error(`[daemon] ${data}`)
  })

  daemonProcess.on('error', (err: Error) => {
    console.error('Failed to start daemon:', err)
  })

  daemonProcess.on('close', (code: number | null) => {
    console.log(`Daemon exited with code ${code}`)
  })
}

function stopDaemon() {
  if (daemonProcess) {
    daemonProcess.kill()
    daemonProcess = null
  }
}

app.whenReady().then(() => {
  // Register IPC handlers for terminal
  registerTerminalIPC(ptyManager, tmuxManager)

  startDaemon()

  // Give daemon a moment to start
  setTimeout(createWindow, 500)

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on('window-all-closed', () => {
  stopDaemon()
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('before-quit', () => {
  ptyManager.killAll()
  stopDaemon()
})
