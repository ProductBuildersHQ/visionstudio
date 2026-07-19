# ADR-001: Embed prism-maturity HTML Dashboard in VisionStudio

**Status:** Proposed
**Date:** 2026-07-14
**Deciders:** @grokify

## Context

VisionStudio currently supports V2MOM visualization but lacks integration with the PRISM Maturity Model. The `prism-maturity` repository already has a sophisticated HTML dashboard generator that produces interactive visualizations including:

- ECharts-based charts (bar, line, area, scatter)
- D3.js bullet charts for maturity levels (M1-M5)
- Grid-based responsive layouts
- Dark/light theme support
- Multiple widget types (text, metric, chart, table, bullet)

We need to bring maturity model visualization into VisionStudio to provide a unified product management experience.

## Decision

We will **embed the prism-maturity HTML dashboard** directly in VisionStudio using an iframe/WebView approach as the initial integration strategy.

### Why Embed HTML?

| Factor | Embed HTML | Native React Port |
|--------|-----------|-------------------|
| **Time to implement** | Days | Weeks |
| **Visual fidelity** | Exact match | Requires recreation |
| **Maintenance burden** | Low (reuse existing) | High (dual codebases) |
| **ECharts/D3 complexity** | Already done | Must port or wrap |
| **Iteration speed** | Fast | Slow |

### Architecture

```
VisionStudio (Electron)
├── Sidebar
│   └── "Maturity Model" button
├── Main Content Area
│   └── MaturityModelView.tsx
│       └── <iframe src={dashboardHtmlUrl} />
└── Go Daemon (port 8765)
    └── GET /api/projects/{name}/maturity/dashboard
        └── Returns HTML from prism-maturity
```

### Integration Flow

1. User clicks "Maturity Model" in sidebar
2. `MaturityModelView` component mounts
3. Component fetches dashboard HTML from daemon API
4. HTML rendered in sandboxed iframe
5. PostMessage API enables bidirectional communication

## Alternatives Considered

### Option A: Native React Components (Rejected for Phase 1)

Port all visualizations to React using:
- `recharts` or `react-echarts` for charts
- `react-d3` for bullet charts
- Custom components for grid layout

**Pros:**
- Tighter integration with VisionStudio theme
- No iframe sandbox limitations
- Better accessibility

**Cons:**
- Significant development effort
- Must maintain two visualization codebases
- Risk of divergence from prism-maturity

**Decision:** Defer to Phase 2 after validating the embedded approach.

### Option B: WebView with Local File (Considered)

Generate HTML file on disk and load via `file://` protocol.

**Pros:**
- Works offline
- No network latency

**Cons:**
- File system coordination complexity
- Security restrictions in Electron
- Harder to refresh dynamically

**Decision:** Use HTTP endpoint instead for simpler state management.

## Implementation Plan

### Phase 1: Basic Embedding (This ADR)

1. **Go Daemon API Endpoint**
   - `GET /api/projects/{name}/maturity/dashboard` → HTML string
   - `GET /api/projects/{name}/maturity/data` → JSON for future use

2. **React Component**
   - `MaturityModelView.tsx` with iframe
   - Loading state and error handling
   - Refresh capability

3. **Sidebar Integration**
   - Add "Maturity Model" navigation item
   - Wire up view switching in App.tsx

### Phase 2: Enhanced Integration (Future)

- PostMessage API for drill-down navigation
- Theme synchronization (dark/light)
- Selection events from dashboard to VisionStudio
- Native React components for key widgets

## Consequences

### Positive

- Rapid delivery of maturity visualization
- Leverages battle-tested prism-maturity code
- Single source of truth for dashboard logic
- Easy to update (change in prism-maturity propagates automatically)

### Negative

- Iframe introduces visual boundary
- Limited interaction between dashboard and VisionStudio
- Must handle iframe security sandbox
- Theme mismatch possible (will address in Phase 2)

### Neutral

- Requires daemon to have prism-maturity as dependency
- Dashboard data must be accessible to daemon

## References

- [prism-maturity dashboard/html.go](https://github.com/grokify/prism-maturity/blob/main/dashboard/html.go)
- [VisionStudio V2MOM implementation](../../../desktop/renderer/src/components/v2mom/)
- [Electron webview security](https://www.electronjs.org/docs/latest/tutorial/security)
