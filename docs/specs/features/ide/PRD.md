# VisionStudio IDE Terminal - Product Requirements Document

## Problem Statement

VisionStudio users currently experience significant context switching between the spec editor and external terminal applications. When working with VisionSpec workflows, users must:

1. Switch to an external terminal to run visionspec commands
2. Manually track workflow status and spec evaluation results
3. Lose context when jumping between spec editing and command execution
4. Manually manage tmux sessions for persistent AI agent interactions

This fragmented workflow reduces productivity and increases cognitive load.

## Target Users

- **Spec Authors**: Product managers and technical writers creating MRD, PRD, UXD specs
- **Tech Leads**: Engineers synthesizing TRD/IRD from source specs
- **AI Agent Operators**: Users running Claude Code or Codex CLI within VisionStudio

## User Stories

### Terminal Access

- **US-1**: As a spec author, I want an integrated terminal so that I can run visionspec commands without leaving the editor.
- **US-2**: As a user, I want multiple terminal tabs so that I can run parallel workflows.
- **US-3**: As a user, I want terminal sessions to persist across app restarts via tmux so that I don't lose work.

### tmux Integration

- **US-4**: As a power user, I want to attach to existing tmux sessions so that I can resume AI agent conversations.
- **US-5**: As a user, I want VisionStudio to auto-create project-specific tmux sessions (`visionstudio-{project}`) so that my work is organized.
- **US-6**: As a user, I want to list and switch between tmux sessions so that I can manage multiple contexts.

### Workflow Integration

- **US-7**: As a spec author, I want to see live workflow status in the terminal so that I know which specs need work.
- **US-8**: As a user, I want spec changes to trigger workflow status updates so that I have real-time feedback.
- **US-9**: As a user, I want to run `/eval`, `/synthesize`, `/status` commands directly so that common actions are accessible.

### AI Agent Support

- **US-10**: As an AI operator, I want Claude Code to work seamlessly in the integrated terminal so that I can use AI assistance.
- **US-11**: As an AI operator, I want terminal output to be captured so that I can review AI agent actions.

## Acceptance Criteria

### Terminal Basics

- [ ] Terminal panel appears at bottom of main content area
- [ ] Terminal supports ANSI colors and escape sequences
- [ ] Terminal is resizable via drag handle
- [ ] Terminal can be collapsed/expanded with keyboard shortcut (Cmd+`)
- [ ] Shell inherits user's default shell ($SHELL)

### Tab Management

- [ ] New tab created with Cmd+T
- [ ] Tab closed with Cmd+W (with confirmation if process running)
- [ ] Tab names show process or custom label
- [ ] Maximum 8 tabs per session

### tmux Integration

- [ ] "tmux Sessions" dropdown lists available sessions
- [ ] Clicking session attaches to it in new tab
- [ ] "New Session" creates `visionstudio-{project}` session
- [ ] Session survives VisionStudio restart

### Workflow Status

- [ ] Workflow overlay shows current phase and spec statuses
- [ ] Overlay updates within 2 seconds of spec file change
- [ ] Color coding matches sidebar (green=pass, orange=conditional, red=fail)
- [ ] Overlay can be dismissed with click

### Performance

- [ ] Terminal input latency < 50ms
- [ ] Terminal handles 10,000+ lines of output without lag
- [ ] Memory usage remains stable during long sessions

## Out of Scope (V1)

- Split terminal panes (future enhancement)
- Terminal multiplexer beyond tmux
- Custom shell profile support
- Terminal recording/playback

## Success Metrics

| Metric | Target |
|--------|--------|
| Terminal adoption rate | 80% of active users |
| Context switches per session | -60% reduction |
| Workflow command execution time | < 2 seconds |
| User satisfaction (NPS) | +10 points |

## Dependencies

- node-pty for PTY management
- xterm.js for terminal rendering
- tmux installed on user's system
- VisionSpec daemon running on port 8765
