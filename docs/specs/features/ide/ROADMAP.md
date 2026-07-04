# VisionStudio IDE Terminal - Feature Roadmap

## Current Release: v0.1.0

### Phase 1: Core Infrastructure (Current)

- [x] PTY manager with node-pty
- [x] IPC bridge via contextBridge
- [x] xterm.js terminal component
- [x] Basic terminal input/output

### Phase 2: Tab Management

- [ ] Multi-tab terminal panel
- [ ] Tab create/close/switch
- [ ] Resizable panel height
- [ ] Terminal state persistence

### Phase 3: tmux Integration

- [ ] tmux session discovery
- [ ] Attach to existing sessions
- [ ] Create project-specific sessions
- [ ] Session selector dropdown

### Phase 4: Workflow Integration

- [ ] Workflow status endpoint
- [ ] Status overlay component
- [ ] Real-time status updates
- [ ] Command shortcuts (/eval, /status)

### Phase 5: Polish

- [ ] Security audit
- [ ] User preferences
- [ ] Keyboard shortcuts
- [ ] Error handling

---

## Future Releases

### v0.2.0 - Enhanced Terminal

**Split Panes**

- Horizontal and vertical splits
- Drag to resize pane boundaries
- Per-pane tmux session binding

**Search & Navigation**

- Cmd+F to search terminal buffer
- Highlight all matches
- Navigate with Enter/Shift+Enter

**Copy/Paste Improvements**

- Rectangular selection
- Auto-copy on selection
- Rich paste with escaping

### v0.3.0 - AI Agent Integration

**Claude Code Support**

- Detect Claude Code sessions
- Show agent status in tab title
- Quick action buttons (approve, reject, abort)

**Codex CLI Support**

- Codex session management
- Output parsing for spec updates
- Auto-refresh workflow on completion

**Agent History**

- Log all agent interactions
- Searchable command history
- Replay previous sessions

### v0.4.0 - Collaboration

**Session Sharing**

- Share terminal session via link
- Read-only viewer mode
- Collaborative editing (pair programming)

**Recording & Playback**

- Record terminal sessions
- Playback with timing
- Export as asciicast

### v0.5.0 - Advanced Workflows

**Custom Commands**

- User-defined command aliases
- Workflow-aware command suggestions
- Command palette (Cmd+Shift+P)

**Integrated Output**

- Parse spec evaluation output
- Inline finding annotations
- Click-to-navigate to spec line

**Notifications**

- Long-running command alerts
- Error notifications
- Spec approval reminders

---

## Version Milestones

| Version | Target | Key Feature |
|---------|--------|-------------|
| v0.1.0 | Q1 2024 | Basic terminal with tmux |
| v0.2.0 | Q2 2024 | Split panes, search |
| v0.3.0 | Q3 2024 | AI agent integration |
| v0.4.0 | Q4 2024 | Collaboration features |
| v0.5.0 | Q1 2025 | Advanced workflows |

---

## Technical Debt Backlog

| Item | Priority | Notes |
|------|----------|-------|
| xterm.js upgrade path | Medium | Track xterm v6 alpha |
| node-pty alternative | Low | Consider rust-pty if perf issues |
| IPC serialization | Medium | Consider using protobuf |
| Test coverage | High | Target 80% for PTY manager |

---

## Dependencies to Watch

| Package | Current | Watch For |
|---------|---------|-----------|
| node-pty | 1.0.0 | Windows support improvements |
| xterm | 5.3.0 | WebGL renderer stability |
| electron | 33.0.0 | Context isolation changes |
| tmux | 3.4 | New control mode features |
