# VisionStudio IDE Terminal - Technical Requirements Document

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│  VisionStudio Desktop (Electron)                            │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ Main Process                                            ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ││
│  │  │ PTY Manager  │  │ tmux Manager │  │ IPC Handlers │  ││
│  │  │  (node-pty)  │  │              │  │              │  ││
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  ││
│  │         │                 │                 │          ││
│  │         └─────────────────┼─────────────────┘          ││
│  │                           │ contextBridge              ││
│  └───────────────────────────┼─────────────────────────────┘│
│                              │                              │
│  ┌───────────────────────────┼─────────────────────────────┐│
│  │ Renderer Process          │                             ││
│  │  ┌──────────────┐  ┌──────┴───────┐  ┌──────────────┐  ││
│  │  │ xterm.js     │  │ useTerminal  │  │ Workflow     │  ││
│  │  │ Terminal UI  │◄─┤ Hook         │  │ Overlay      │  ││
│  │  └──────────────┘  └──────────────┘  └──────┬───────┘  ││
│  │                                             │ HTTP     ││
│  └─────────────────────────────────────────────┼───────────┘│
└─────────────────────────────────────────────────┼───────────┘
                                                  │
                                                  ▼
                                         ┌──────────────┐
                                         │ Go Daemon    │
                                         │ :8765        │
                                         └──────────────┘
```

## Component Specifications

### 1. PTY Manager (`desktop/main/pty/manager.ts`)

Manages pseudo-terminal lifecycle in the main process.

```typescript
interface PTYSession {
  id: string;
  pid: number;
  shell: string;
  cwd: string;
  cols: number;
  rows: number;
  createdAt: Date;
}

class PTYManager {
  private sessions: Map<string, IPty>;

  spawn(options: SpawnOptions): Promise<PTYSession>;
  write(sessionId: string, data: string): void;
  resize(sessionId: string, cols: number, rows: number): void;
  kill(sessionId: string): Promise<void>;
  listSessions(): PTYSession[];
}
```

**Responsibilities:**

- Spawn new PTY processes with user's default shell
- Route input/output between renderer and PTY
- Handle process lifecycle (spawn, resize, kill)
- Clean up orphaned sessions on app quit

### 2. tmux Manager (`desktop/main/pty/tmux.ts`)

Discovers and manages tmux sessions.

```typescript
interface TmuxSession {
  name: string;
  windows: number;
  created: Date;
  attached: boolean;
}

class TmuxManager {
  listSessions(): Promise<TmuxSession[]>;
  attach(sessionName: string): Promise<string>; // Returns PTY session ID
  create(sessionName: string, command?: string): Promise<string>;
  detach(sessionId: string): Promise<void>;
  hasSession(sessionName: string): Promise<boolean>;
}
```

**Session Naming Convention:**

- Project sessions: `visionstudio-{project-name}`
- Generic sessions: `visionstudio-{uuid}`

### 3. IPC Protocol (`desktop/main/ipc/terminal.ts`)

Electron IPC channels for terminal communication.

| Channel | Direction | Payload | Description |
|---------|-----------|---------|-------------|
| `terminal:spawn` | Renderer → Main | `SpawnOptions` | Create new PTY session |
| `terminal:write` | Renderer → Main | `{id, data}` | Send input to PTY |
| `terminal:resize` | Renderer → Main | `{id, cols, rows}` | Resize terminal |
| `terminal:kill` | Renderer → Main | `{id}` | Kill PTY session |
| `terminal:data` | Main → Renderer | `{id, data}` | PTY output data |
| `terminal:exit` | Main → Renderer | `{id, code}` | PTY process exited |
| `tmux:list` | Renderer → Main | - | List tmux sessions |
| `tmux:attach` | Renderer → Main | `{name}` | Attach to session |
| `tmux:spawn` | Renderer → Main | `{name, command?}` | Create new session |

**Request-Reply Pattern:**

```typescript
// Renderer
const session = await window.electronAPI.terminal.spawn({
  cwd: '/path/to/project',
  shell: process.env.SHELL
});

// Main (handler)
ipcMain.handle('terminal:spawn', async (event, options) => {
  return ptyManager.spawn(options);
});
```

**Streaming Pattern:**

```typescript
// Main (sender)
pty.onData((data) => {
  mainWindow.webContents.send('terminal:data', { id, data });
});

// Renderer (listener)
window.electronAPI.terminal.onData((id, data) => {
  terminalRef.current.write(data);
});
```

### 4. Preload Script (`desktop/main/preload.ts`)

Context bridge exposing safe IPC methods.

```typescript
import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('electronAPI', {
  terminal: {
    spawn: (options) => ipcRenderer.invoke('terminal:spawn', options),
    write: (id, data) => ipcRenderer.send('terminal:write', { id, data }),
    resize: (id, cols, rows) => ipcRenderer.send('terminal:resize', { id, cols, rows }),
    kill: (id) => ipcRenderer.invoke('terminal:kill', { id }),
    onData: (callback) => {
      ipcRenderer.on('terminal:data', (_, data) => callback(data.id, data.data));
    },
    onExit: (callback) => {
      ipcRenderer.on('terminal:exit', (_, data) => callback(data.id, data.code));
    },
  },
  tmux: {
    list: () => ipcRenderer.invoke('tmux:list'),
    attach: (name) => ipcRenderer.invoke('tmux:attach', { name }),
    spawn: (name, command) => ipcRenderer.invoke('tmux:spawn', { name, command }),
  },
});
```

### 5. Terminal Instance (`desktop/renderer/src/components/terminal/TerminalInstance.tsx`)

xterm.js wrapper component.

```typescript
interface TerminalInstanceProps {
  sessionId: string;
  onTitleChange?: (title: string) => void;
  onExit?: (code: number) => void;
}

const TerminalInstance: React.FC<TerminalInstanceProps> = ({
  sessionId,
  onTitleChange,
  onExit,
}) => {
  const terminalRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<Terminal>();
  const fitAddonRef = useRef<FitAddon>();

  useEffect(() => {
    // Initialize xterm.js with addons
    // Subscribe to IPC data events
    // Handle resize events
    // Cleanup on unmount
  }, [sessionId]);

  return <div ref={terminalRef} className="terminal-container" />;
};
```

**xterm.js Configuration:**

```typescript
const terminal = new Terminal({
  theme: {
    background: '#1b2636',
    foreground: '#e0e6ed',
    cursor: '#3b82f6',
    selection: 'rgba(59, 130, 246, 0.3)',
  },
  fontFamily: 'SF Mono, Menlo, Monaco, monospace',
  fontSize: 13,
  lineHeight: 1.2,
  cursorBlink: true,
  allowProposedApi: true,
});
```

### 6. Workflow Status Endpoint

**Go Daemon Addition (`cmd/daemon/main.go`):**

```go
// GET /api/projects/{project}/workflow/status
func (s *Server) handleGetWorkflowStatus(w http.ResponseWriter, r *http.Request) {
    projectName := chi.URLParam(r, "project")
    project, err := s.loadProject(projectName)
    if err != nil {
        s.writeJSON(w, http.StatusNotFound, api.ErrorResponse{Error: err.Error()})
        return
    }

    status := s.computeWorkflowStatus(project)
    s.writeJSON(w, http.StatusOK, api.GetWorkflowStatusResponse{Status: status})
}
```

**Type Definition (`pkg/api/types.go`):**

```go
type WorkflowStatus struct {
    CurrentPhase    string            `json:"currentPhase"`
    CompletedPhases []string          `json:"completedPhases"`
    Progress        float64           `json:"progress"` // 0.0 to 1.0
    SpecStatuses    map[string]string `json:"specStatuses"`
    BlockedBy       []string          `json:"blockedBy,omitempty"`
    LastUpdated     time.Time         `json:"lastUpdated"`
}

type GetWorkflowStatusResponse struct {
    Status WorkflowStatus `json:"status"`
}
```

## Security Considerations

### IPC Surface Hardening

1. **Context Isolation**: `contextIsolation: true` ensures renderer cannot access Node.js
2. **Preload Whitelist**: Only expose necessary APIs through contextBridge
3. **Input Validation**: Sanitize all IPC inputs in main process
4. **Path Traversal**: Validate cwd paths before spawning PTY

### Shell Execution

1. **No Shell Injection**: Use spawn with args array, not shell string
2. **Environment Inheritance**: Only pass safe environment variables
3. **Process Limits**: Enforce max concurrent PTY sessions (8)
4. **Cleanup**: Kill orphaned processes on app exit

### tmux Security

1. **Socket Access**: Use user's tmux socket only
2. **Session Names**: Validate session names against allowlist pattern
3. **Command Escaping**: Escape commands passed to tmux

## Performance Requirements

| Metric | Requirement | Implementation |
|--------|-------------|----------------|
| Input latency | < 50ms | Direct PTY write, no buffering |
| Output throughput | 10MB/s | xterm.js flow control |
| Memory per terminal | < 50MB | Scrollback limit (5000 lines) |
| Startup time | < 200ms | Lazy addon loading |

## Dependencies

### NPM Packages

| Package | Version | Purpose |
|---------|---------|---------|
| `node-pty` | ^1.0.0 | PTY process management |
| `xterm` | ^5.3.0 | Terminal emulator UI |
| `xterm-addon-fit` | ^0.8.0 | Auto-resize terminal |
| `xterm-addon-web-links` | ^0.9.0 | Clickable URLs |

### System Requirements

- macOS 10.15+ (for PTY support)
- tmux 3.0+ (optional, for session management)
- User's default shell (bash, zsh, fish)

## Testing Strategy

### Unit Tests

- PTY Manager: spawn/kill lifecycle
- tmux Manager: session discovery
- IPC handlers: message routing

### Integration Tests

- Terminal ↔ PTY communication
- tmux session attach/detach
- Workflow status polling

### E2E Tests

- Open terminal, type command, see output
- Attach to tmux session, detach
- Workflow overlay updates on spec change
