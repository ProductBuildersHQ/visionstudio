# Maturity Model Integration: Design Document

**Version:** 1.0.0
**Date:** 2026-07-14
**Related ADR:** [ADR-001-embed-html-dashboard.md](./ADR-001-embed-html-dashboard.md)

## Overview

This document details the implementation of maturity model visualization in VisionStudio by embedding prism-maturity HTML dashboards.

## Data Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                         VisionStudio                                 │
│  ┌──────────┐    ┌─────────────────────┐    ┌──────────────────┐   │
│  │ Sidebar  │───▶│ MaturityModelView   │───▶│ iframe           │   │
│  │          │    │ (React Component)   │    │ (HTML Dashboard) │   │
│  └──────────┘    └─────────────────────┘    └──────────────────┘   │
│                           │                          ▲              │
│                           │ fetch()                  │              │
│                           ▼                          │              │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Go Daemon (port 8765)                     │   │
│  │  ┌─────────────────────────────────────────────────────┐    │   │
│  │  │ GET /api/projects/{name}/maturity/dashboard         │    │   │
│  │  │     → prism-maturity.Dashboard.ToHTML()             │────┘   │
│  │  └─────────────────────────────────────────────────────┘        │
│  │  ┌─────────────────────────────────────────────────────┐        │
│  │  │ GET /api/projects/{name}/maturity/models            │        │
│  │  │     → List available maturity models                │        │
│  │  └─────────────────────────────────────────────────────┘        │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                │ reads
                                ▼
                    ┌───────────────────────┐
                    │  .visionspec/         │
                    │  └── maturity/        │
                    │      ├── model.json   │
                    │      └── state.json   │
                    └───────────────────────┘
```

## API Specification

### GET /api/projects/{project}/maturity/dashboard

Returns a complete HTML document for the maturity dashboard.

**Parameters:**

| Name | In | Type | Description |
|------|-----|------|-------------|
| project | path | string | Project name |
| theme | query | string | `light` or `dark` (default: `dark`) |
| embed | query | boolean | If true, omit `<html>` wrapper for iframe (default: `true`) |

**Response:** `text/html`

```html
<!DOCTYPE html>
<html>
<head>
  <script src="https://cdn.jsdelivr.net/npm/echarts@5.5.0/dist/echarts.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/d3@7/dist/d3.min.js"></script>
  <style>/* dashboard styles */</style>
</head>
<body>
  <div id="dashboard"><!-- widgets --></div>
  <script>/* initialization */</script>
</body>
</html>
```

### GET /api/projects/{project}/maturity/models

Returns available maturity model definitions.

**Response:** `application/json`

```json
{
  "models": [
    {
      "id": "capability-maturity",
      "name": "Capability Maturity Model",
      "description": "5-level maturity assessment across capability stack",
      "dimensions": 12,
      "lastUpdated": "2026-07-14T10:30:00Z"
    }
  ]
}
```

### GET /api/projects/{project}/maturity/models/{modelId}

Returns a specific maturity model with current state.

**Response:** `application/json`

```json
{
  "model": {
    "id": "capability-maturity",
    "name": "Capability Maturity Model",
    "levels": [
      {"level": 1, "name": "Reactive", "color": "#ef4444"},
      {"level": 2, "name": "Basic", "color": "#f59e0b"},
      {"level": 3, "name": "Defined", "color": "#eab308"},
      {"level": 4, "name": "Managed", "color": "#22c55e"},
      {"level": 5, "name": "Optimizing", "color": "#3b82f6"}
    ],
    "dimensions": [
      {
        "id": "security",
        "name": "Security",
        "currentLevel": 3,
        "targetLevel": 4,
        "capabilities": [
          {"id": "auth", "name": "Authentication", "level": 4},
          {"id": "authz", "name": "Authorization", "level": 3}
        ]
      }
    ],
    "overallScore": 3.2
  }
}
```

## Frontend Components

### MaturityModelView.tsx

```typescript
import { useState, useEffect, useRef } from 'react'
import { api } from '../../services/api'

interface MaturityModelViewProps {
  projectName: string
}

export function MaturityModelView({ projectName }: MaturityModelViewProps) {
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dashboardHtml, setDashboardHtml] = useState<string>('')
  const iframeRef = useRef<HTMLIFrameElement>(null)

  useEffect(() => {
    loadDashboard()
  }, [projectName])

  async function loadDashboard() {
    setIsLoading(true)
    setError(null)
    try {
      const html = await api.getMaturityDashboard(projectName, { theme: 'dark' })
      setDashboardHtml(html)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard')
    } finally {
      setIsLoading(false)
    }
  }

  function handleRefresh() {
    loadDashboard()
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-va-text-muted">Loading maturity dashboard...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-4">
        <div className="text-va-error">{error}</div>
        <button
          onClick={handleRefresh}
          className="px-4 py-2 bg-va-accent text-white rounded hover:bg-va-accent/80"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-va-border">
        <h2 className="text-lg font-semibold text-va-text">Maturity Model</h2>
        <button
          onClick={handleRefresh}
          className="p-1.5 rounded hover:bg-va-panel text-va-text-muted hover:text-va-text"
          title="Refresh"
        >
          <RefreshIcon className="w-4 h-4" />
        </button>
      </div>

      {/* Dashboard iframe */}
      <div className="flex-1 overflow-hidden">
        <iframe
          ref={iframeRef}
          srcDoc={dashboardHtml}
          className="w-full h-full border-0"
          sandbox="allow-scripts allow-same-origin"
          title="Maturity Model Dashboard"
        />
      </div>
    </div>
  )
}

function RefreshIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
    </svg>
  )
}
```

### API Service Extension

Add to `services/api.ts`:

```typescript
export const api = {
  // ... existing methods

  async getMaturityDashboard(
    project: string,
    options?: { theme?: 'light' | 'dark' }
  ): Promise<string> {
    const params = new URLSearchParams()
    if (options?.theme) params.set('theme', options.theme)
    params.set('embed', 'true')

    const response = await fetch(
      `${API_BASE}/projects/${project}/maturity/dashboard?${params}`
    )
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }
    return response.text()
  },

  async listMaturityModels(project: string): Promise<MaturityModelSummary[]> {
    const data = await fetchJSON<{ models: MaturityModelSummary[] }>(
      `${API_BASE}/projects/${project}/maturity/models`
    )
    return data.models || []
  },

  async getMaturityModel(project: string, modelId: string): Promise<MaturityModel> {
    const data = await fetchJSON<{ model: MaturityModel }>(
      `${API_BASE}/projects/${project}/maturity/models/${modelId}`
    )
    return data.model
  },
}
```

### Type Definitions

Add to `types/maturity.ts`:

```typescript
export interface MaturityLevel {
  level: number
  name: string
  description?: string
  color: string
}

export interface MaturityCapability {
  id: string
  name: string
  level: number
  targetLevel?: number
}

export interface MaturityDimension {
  id: string
  name: string
  description?: string
  currentLevel: number
  targetLevel: number
  capabilities: MaturityCapability[]
}

export interface MaturityModel {
  id: string
  name: string
  description?: string
  levels: MaturityLevel[]
  dimensions: MaturityDimension[]
  overallScore: number
  lastUpdated: string
}

export interface MaturityModelSummary {
  id: string
  name: string
  description?: string
  dimensions: number
  lastUpdated: string
}
```

## Backend Implementation

### Go Daemon Handler

Add to `cmd/daemon/main.go`:

```go
import (
    "github.com/grokify/prism-maturity/dashboard"
)

func handleMaturityDashboard(w http.ResponseWriter, r *http.Request) {
    projectName := chi.URLParam(r, "project")
    theme := r.URL.Query().Get("theme")
    if theme == "" {
        theme = "dark"
    }

    // Load project maturity data
    project, err := loadProject(projectName)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    // Build dashboard from prism-maturity
    dash, err := buildMaturityDashboard(project)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Render to HTML
    html, err := dash.ToHTML(dashboard.HTMLOptions{
        EmbedData: true,
        Theme:     theme,
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(html))
}

// Register routes
r.Get("/api/projects/{project}/maturity/dashboard", handleMaturityDashboard)
r.Get("/api/projects/{project}/maturity/models", handleListMaturityModels)
r.Get("/api/projects/{project}/maturity/models/{modelId}", handleGetMaturityModel)
```

## Sidebar Integration

Update `Sidebar.tsx`:

```typescript
interface SidebarProps {
  // ... existing props
  onMaturityModelClick: () => void
}

// In the navigation section, add:
<div className="mt-4 pt-4 border-t border-va-border">
  <div className="px-2 mb-2 text-xs font-medium text-va-text-muted uppercase tracking-wider">
    Analytics
  </div>
  <button
    onClick={onMaturityModelClick}
    className="w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm text-left text-va-text-muted hover:text-va-text hover:bg-va-panel transition-colors"
  >
    <span className="text-base">📊</span>
    <span>Maturity Model</span>
  </button>
</div>
```

## App.tsx Integration

```typescript
type ActiveView = 'workflow' | 'spec' | 'findings' | 'v2mom' | 'maturity-model'

function App() {
  const [activeView, setActiveView] = useState<ActiveView>('workflow')
  // ... existing state

  const handleMaturityModelClick = () => {
    setActiveView('maturity-model')
    setActiveSpec(null)
  }

  return (
    <AppLayout
      sidebar={
        <Sidebar
          // ... existing props
          onMaturityModelClick={handleMaturityModelClick}
        />
      }
      main={
        activeView === 'maturity-model' && activeProject ? (
          <MaturityModelView projectName={activeProject.name} />
        ) : activeView === 'v2mom' ? (
          <V2MOMCascadeView ... />
        ) : // ... other views
      }
      terminal={<TerminalPanel />}
    />
  )
}
```

## Theme Synchronization

To sync VisionStudio's dark theme with the embedded dashboard:

### Option 1: Query Parameter (Implemented)

Pass `?theme=dark` to the dashboard endpoint. The prism-maturity HTML template already supports this.

### Option 2: PostMessage API (Future)

```typescript
// In MaturityModelView.tsx
useEffect(() => {
  const iframe = iframeRef.current
  if (iframe?.contentWindow) {
    iframe.contentWindow.postMessage({ type: 'SET_THEME', theme: 'dark' }, '*')
  }
}, [dashboardHtml])

// In dashboard HTML template, add listener:
window.addEventListener('message', (event) => {
  if (event.data.type === 'SET_THEME') {
    document.documentElement.setAttribute('data-theme', event.data.theme)
  }
})
```

## File Structure

```
desktop/renderer/src/
├── components/
│   ├── maturity-model/
│   │   ├── MaturityModelView.tsx    # Main view component
│   │   ├── types.ts                 # TypeScript types
│   │   └── index.ts                 # Exports
│   └── index.ts                     # Add maturity-model exports
├── services/
│   └── api.ts                       # Add maturity API methods
└── types/
    └── maturity.ts                  # Maturity type definitions

cmd/daemon/
├── main.go                          # Add maturity routes
└── maturity.go                      # Maturity handlers (new file)
```

## Testing

### Manual Testing Checklist

- [ ] Sidebar shows "Maturity Model" button
- [ ] Clicking navigates to maturity view
- [ ] Dashboard loads and displays
- [ ] Charts render correctly (ECharts, D3)
- [ ] Dark theme matches VisionStudio
- [ ] Refresh button reloads dashboard
- [ ] Error state shows retry option
- [ ] Loading state shows spinner

### Integration Tests

```typescript
describe('MaturityModelView', () => {
  it('loads dashboard on mount', async () => {
    render(<MaturityModelView projectName="test-project" />)
    expect(screen.getByText('Loading maturity dashboard...')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByTitle('Maturity Model Dashboard')).toBeInTheDocument()
    })
  })

  it('shows error state on failure', async () => {
    server.use(
      rest.get('/api/projects/:project/maturity/dashboard', (req, res, ctx) => {
        return res(ctx.status(500))
      })
    )
    render(<MaturityModelView projectName="test-project" />)
    await waitFor(() => {
      expect(screen.getByText(/Failed to load/)).toBeInTheDocument()
    })
  })
})
```

## Migration Path to Native React

When ready to move beyond iframe embedding:

1. **Extract widget components** from prism-maturity HTML template
2. **Create React wrappers** for ECharts (`react-echarts`)
3. **Port D3 bullet chart** to React component
4. **Reuse maturity data API** (already JSON-based)
5. **Gradually replace** iframe sections with native components

This allows incremental migration while maintaining functionality.
