# VisionStudio IDE Terminal - Implementation Plan

## Phase Overview

| Phase | Focus | Deliverables |
|-------|-------|--------------|
| 1 | Core Infrastructure | PTY manager, IPC bridge, basic terminal |
| 2 | Tab Management | Multi-tab UI, state management |
| 3 | tmux Integration | Session discovery, attach/create |
| 4 | Workflow Integration | Status overlay, daemon endpoint |
| 5 | Polish | Security audit, preferences, shortcuts |

---

## Phase 1: Core Terminal Infrastructure

**Goal**: Establish the foundational terminal architecture.

### New Files

| File | Purpose |
|------|---------|
| `desktop/main/preload.ts` | Context bridge for IPC |
| `desktop/main/pty/manager.ts` | PTY lifecycle management |
| `desktop/main/ipc/terminal.ts` | IPC handlers for terminal |
| `desktop/renderer/src/components/terminal/TerminalInstance.tsx` | xterm.js wrapper |
| `desktop/renderer/src/types/terminal.ts` | TypeScript interfaces |

### Modified Files

| File | Changes |
|------|---------|
| `desktop/main/index.ts` | Add preload script, initialize PTY manager |
| `desktop/package.json` | Add node-pty, xterm.js dependencies |

### Dependencies

```json
{
  "dependencies": {
    "xterm": "^5.3.0",
    "xterm-addon-fit": "^0.8.0",
    "xterm-addon-web-links": "^0.9.0"
  },
  "devDependencies": {
    "node-pty": "^1.0.0"
  }
}
```

### Verification

- [ ] `npm run dev` starts without errors
- [ ] Terminal component renders in isolation
- [ ] PTY spawns shell process
- [ ] Input/output flows through IPC

---

## Phase 2: Tab Management and Layout

**Goal**: Enable multi-terminal workflow with tab UI.

### New Files

| File | Purpose |
|------|---------|
| `desktop/renderer/src/components/terminal/TerminalPanel.tsx` | Container with tabs and terminals |
| `desktop/renderer/src/components/terminal/TerminalTabs.tsx` | Tab bar component |
| `desktop/renderer/src/components/terminal/hooks/useTerminal.ts` | Terminal state management |

### Modified Files

| File | Changes |
|------|---------|
| `desktop/renderer/src/components/layout/AppLayout.tsx` | Add resizable bottom panel |
| `desktop/renderer/src/App.tsx` | Add terminal state and context |
| `desktop/renderer/src/index.css` | Terminal panel styles |

### UI Specification

```
┌─────────────────────────────────────────────────┐
│ [Tab 1] [Tab 2] [+]              [▼ Sessions]  │
├─────────────────────────────────────────────────┤
│ $ _                                             │
│                                                 │
│                                                 │
└─────────────────────────────────────────────────┘
```

### Verification

- [ ] New tab creates new terminal session
- [ ] Tab switch changes active terminal
- [ ] Close tab kills PTY process
- [ ] Panel resizes with drag handle

---

## Phase 3: tmux Integration

**Goal**: Persist terminal sessions across app restarts.

### New Files

| File | Purpose |
|------|---------|
| `desktop/main/pty/tmux.ts` | tmux session management |
| `desktop/renderer/src/components/terminal/hooks/useTmux.ts` | tmux React hook |

### IPC Channels

| Channel | Handler |
|---------|---------|
| `tmux:list` | List available sessions |
| `tmux:attach` | Attach to existing session |
| `tmux:spawn` | Create new session |
| `tmux:detach` | Detach from session |

### tmux Session Naming

```
visionstudio-{project-slug}    # Project-specific sessions
visionstudio-default           # Fallback session
```

### Verification

- [ ] Sessions dropdown lists tmux sessions
- [ ] Attach reconnects to existing session
- [ ] New session creates with naming convention
- [ ] Session persists after app quit/restart

---

## Phase 4: Workflow Integration

**Goal**: Connect terminal to VisionSpec workflow status.

### New Files

| File | Purpose |
|------|---------|
| `desktop/renderer/src/components/terminal/WorkflowOverlay.tsx` | Status overlay UI |

### Modified Files

| File | Changes |
|------|---------|
| `cmd/daemon/main.go` | Add `/api/workflow/status` endpoint |
| `pkg/api/types.go` | Add `WorkflowStatus` type |
| `desktop/renderer/src/services/api.ts` | Add workflow status method |

### API Endpoint

```
GET /api/projects/{project}/workflow/status

Response:
{
  "status": {
    "currentPhase": "product",
    "completedPhases": ["source", "gtm"],
    "progress": 0.6,
    "specStatuses": {
      "mrd": "approved",
      "prd": "evaluated",
      "trd": "draft"
    },
    "lastUpdated": "2024-01-15T10:30:00Z"
  }
}
```

### Overlay Behavior

- Appears on hover over workflow indicator
- Updates via polling (5 second interval)
- Color-coded spec chips
- Dismissible with click

### Verification

- [ ] Workflow endpoint returns valid status
- [ ] Overlay displays current phase
- [ ] Status updates on spec file change
- [ ] Colors match sidebar spec status

---

## Phase 5: Polish

**Goal**: Production-ready quality and user experience.

### Security Audit

| Area | Check |
|------|-------|
| IPC surface | Validate all inputs in main process |
| Path traversal | Sanitize cwd before PTY spawn |
| Shell injection | Use array args, not shell string |
| Resource limits | Cap PTY sessions at 8 |

### Error Handling

- PTY spawn failure → Show error in terminal UI
- tmux not installed → Graceful fallback to raw PTY
- IPC timeout → Retry with exponential backoff
- Process crash → Auto-reconnect with notification

### User Preferences

| Setting | Default | Storage |
|---------|---------|---------|
| Font family | SF Mono | localStorage |
| Font size | 13 | localStorage |
| Theme | Dark | localStorage |
| Default shell | $SHELL | System |
| Scrollback lines | 5000 | localStorage |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd+`` | Toggle terminal panel |
| `Cmd+T` | New terminal tab |
| `Cmd+W` | Close active tab |
| `Cmd+Shift+]` | Next tab |
| `Cmd+Shift+[` | Previous tab |
| `Cmd+K` | Clear terminal |

### Verification

- [ ] All shortcuts work as documented
- [ ] Preferences persist across sessions
- [ ] Error states display helpful messages
- [ ] No console errors in production build

---

## File Change Summary

### New Files (13)

```
desktop/main/
├── preload.ts
├── pty/
│   ├── manager.ts
│   └── tmux.ts
└── ipc/
    └── terminal.ts

desktop/renderer/src/
├── components/terminal/
│   ├── TerminalInstance.tsx
│   ├── TerminalPanel.tsx
│   ├── TerminalTabs.tsx
│   ├── WorkflowOverlay.tsx
│   └── hooks/
│       ├── useTerminal.ts
│       └── useTmux.ts
└── types/
    └── terminal.ts
```

### Modified Files (6)

```
desktop/
├── main/index.ts
├── package.json
└── renderer/src/
    ├── App.tsx
    ├── components/layout/AppLayout.tsx
    ├── services/api.ts
    └── index.css

cmd/daemon/main.go
pkg/api/types.go
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| node-pty native build fails | Pin to stable version, document build deps |
| tmux version incompatibility | Detect version, use common subset of features |
| Performance with many terminals | Lazy render inactive terminals |
| IPC message overflow | Implement backpressure |
