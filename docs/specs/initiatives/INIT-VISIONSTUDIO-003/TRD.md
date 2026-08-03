# TRD: Workflow-Scoped Spec MCP Server

**Initiative:** INIT-VISIONSTUDIO-003  
**Status:** Draft  
**Author:** John Wang  
**Date:** 2026-08-02

## Overview

This document specifies the technical implementation for adding workflow-scoped specification operations to visionstudio's MCP server.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Client (Claude Code)                 │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     visionstudio MCP Server                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Initiative   │  │ Workflow     │  │ Spec         │          │
│  │ Tools        │  │ Tools        │  │ Tools        │          │
│  │ (existing)   │  │ (new)        │  │ (new)        │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            ▼                                    │
│                    ┌──────────────┐                             │
│                    │   Service    │                             │
│                    │   Layer      │                             │
│                    └──────────────┘                             │
└─────────────────────────────────────────────────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                ▼               ▼               ▼
        ┌──────────┐    ┌──────────┐    ┌──────────┐
        │ Dolt     │    │ Filesystem│   │ spec-    │
        │ Store    │    │ (specs)   │   │ workflow │
        └──────────┘    └──────────┘    │ -spec    │
                                        └──────────┘
```

## Package Structure

```
pkg/
├── mcpserver/
│   ├── server.go           # Existing - add workflow/spec tool registration
│   ├── workflow_tools.go   # NEW - workflow_list, workflow_select, workflow_status
│   └── spec_tools.go       # NEW - spec_list, spec_create, spec_evaluate, etc.
├── specworkflow/
│   ├── seed.go             # Existing - built-in workflow definitions
│   ├── resolver.go         # NEW - workflow resolution and state
│   └── executor.go         # NEW - synthesis and evaluation execution
├── service/
│   └── spec_service.go     # NEW - spec operations service layer
└── store/
    └── spec_store.go       # NEW - spec metadata persistence
```

## Database Schema

### spec_documents Table

```sql
CREATE TABLE spec_documents (
    id VARCHAR(255) PRIMARY KEY,           -- e.g., "INIT-FOO-001/prd"
    initiative_id VARCHAR(255) NOT NULL,
    spec_type VARCHAR(64) NOT NULL,        -- e.g., "prd", "press", "trd"
    workflow_id VARCHAR(64),               -- NULL for custom specs
    file_path VARCHAR(512) NOT NULL,       -- relative to initiative spec dir
    status VARCHAR(32) NOT NULL DEFAULT 'draft',  -- draft, evaluated, approved, rejected
    checksum VARCHAR(64),                  -- SHA256 of content
    eval_score INT,                        -- 0-100 from last evaluation
    eval_verdict VARCHAR(32),              -- pass, partial, fail
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (initiative_id) REFERENCES initiatives(id),
    FOREIGN KEY (workflow_id) REFERENCES spec_workflows(id)
);
```

### initiative_workflow Table

```sql
CREATE TABLE initiative_workflow (
    initiative_id VARCHAR(255) PRIMARY KEY,
    workflow_id VARCHAR(64) NOT NULL,
    selected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (initiative_id) REFERENCES initiatives(id),
    FOREIGN KEY (workflow_id) REFERENCES spec_workflows(id)
);
```

## MCP Tool Specifications

### workflow_list

```json
{
  "name": "workflow_list",
  "description": "List available specification workflows with their required and optional specs.",
  "inputSchema": {
    "type": "object",
    "properties": {}
  }
}
```

**Response:**
```json
{
  "workflows": [
    {
      "id": "aws-working-backwards",
      "name": "AWS Working Backwards",
      "description": "Full PR/FAQ process for major product initiatives",
      "specs_required": ["press", "faq", "narrative-6p", "prd", "trd", "plan", "roadmap"],
      "specs_optional": ["mrd", "uxd"],
      "init_types": []
    }
  ]
}
```

### workflow_select

```json
{
  "name": "workflow_select",
  "description": "Activate a workflow for an initiative. Subsequent spec operations are scoped to this workflow.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "initiative_id": {"type": "string", "description": "Initiative ID"},
      "workflow_id": {"type": "string", "description": "Workflow ID to activate"}
    },
    "required": ["initiative_id", "workflow_id"]
  }
}
```

### workflow_status

```json
{
  "name": "workflow_status",
  "description": "Show current workflow position: which specs exist, their status, gates blocking, and recommended next steps.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "initiative_id": {"type": "string", "description": "Initiative ID"}
    },
    "required": ["initiative_id"]
  }
}
```

**Response:**
```json
{
  "initiative_id": "INIT-FOO-001",
  "workflow": {
    "id": "aws-working-backwards",
    "name": "AWS Working Backwards"
  },
  "specs": [
    {"type": "mrd", "status": "approved", "eval_score": 92},
    {"type": "press", "status": "evaluated", "eval_score": 78, "eval_verdict": "partial"},
    {"type": "faq", "status": "draft"},
    {"type": "prd", "status": "missing"},
    {"type": "trd", "status": "missing"}
  ],
  "gates": [
    {"after": "narrative-6p", "action": "stakeholder_review", "status": "pending"}
  ],
  "next_steps": [
    "Improve press release (score 78 < 85 threshold)",
    "Create faq spec (can synthesize from mrd + press)"
  ]
}
```

### spec_list

```json
{
  "name": "spec_list",
  "description": "List specs for an initiative. If a workflow is active, shows workflow-defined specs plus any custom additions.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "initiative_id": {"type": "string", "description": "Initiative ID"}
    },
    "required": ["initiative_id"]
  }
}
```

### spec_create

```json
{
  "name": "spec_create",
  "description": "Create a new spec from the workflow template. Writes to the initiative's spec directory.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "initiative_id": {"type": "string"},
      "spec_type": {"type": "string", "description": "Spec type (e.g., prd, trd, press)"},
      "title": {"type": "string", "description": "Optional title override"}
    },
    "required": ["initiative_id", "spec_type"]
  }
}
```

### spec_read

```json
{
  "name": "spec_read",
  "description": "Read spec content.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "initiative_id": {"type": "string"},
      "spec_type": {"type": "string"}
    },
    "required": ["initiative_id", "spec_type"]
  }
}
```

### spec_evaluate

```json
{
  "name": "spec_evaluate",
  "description": "Evaluate a spec against the workflow's rubric using LLM-as-judge.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "initiative_id": {"type": "string"},
      "spec_type": {"type": "string"},
      "model": {"type": "string", "description": "Model for evaluation (default: claude-sonnet-4-20250514)"}
    },
    "required": ["initiative_id", "spec_type"]
  }
}
```

**Response:**
```json
{
  "spec_type": "prd",
  "score": 85,
  "verdict": "pass",
  "findings": [
    {"severity": "medium", "section": "User Stories", "message": "Missing edge case coverage for offline mode"}
  ],
  "recommendation": "Ready for technical review"
}
```

### spec_synthesize

```json
{
  "name": "spec_synthesize",
  "description": "Generate a spec from source specs following workflow synthesis rules.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "initiative_id": {"type": "string"},
      "spec_type": {"type": "string", "description": "Target spec type to synthesize"},
      "model": {"type": "string", "description": "Model for synthesis (default: claude-sonnet-4-20250514)"},
      "dry_run": {"type": "boolean", "description": "Preview without writing file"}
    },
    "required": ["initiative_id", "spec_type"]
  }
}
```

### spec_add

```json
{
  "name": "spec_add",
  "description": "Add a custom spec beyond the workflow definition.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "initiative_id": {"type": "string"},
      "name": {"type": "string", "description": "Spec name (e.g., competitor-analysis)"},
      "content": {"type": "string", "description": "Initial content"},
      "category": {"type": "string", "enum": ["source", "gtm", "technical", "output"]}
    },
    "required": ["initiative_id", "name"]
  }
}
```

## Integration with specification-workflow-spec

### Import Types

```go
import (
    "github.com/ProductBuildersHQ/specification-workflow-spec/pkg/layout"
    "github.com/ProductBuildersHQ/specification-workflow-spec/pkg/profile"
    "github.com/ProductBuildersHQ/specification-workflow-spec/pkg/spectype"
)
```

### Layout Usage

```go
func specPath(initiativeID, specType, category string) string {
    l := layout.DefaultLayout()
    initDir := specworkflow.SpecDir(initiativeID)  // docs/specs/initiatives/{INIT-ID}/
    return filepath.Join(initDir, l.SpecPath(specType, category))
}
```

### Profile Resolution

```go
func resolveProfile(workflowID string) (*profile.Profile, error) {
    // Load from embedded profiles or database
    // Support profile inheritance via Extends field
}
```

## Filesystem Layout

Per initiative:
```
docs/specs/initiatives/INIT-FOO-001/
├── source/
│   ├── mrd.md
│   └── prd.md
├── gtm/
│   ├── press.md
│   ├── faq.md
│   └── narrative-6p.md
├── technical/
│   └── trd.md
├── evals/
│   ├── mrd.eval.json
│   ├── prd.eval.json
│   └── press.eval.json
├── PLAN.md
└── ROADMAP.md
```

## Implementation Phases

### Phase 1: Core Tools
- [ ] workflow_list, workflow_select, workflow_status
- [ ] spec_list, spec_read, spec_create
- [ ] Database schema for spec_documents and initiative_workflow

### Phase 2: Evaluation
- [ ] spec_evaluate with structured-evaluation integration
- [ ] Eval result persistence
- [ ] Gate enforcement

### Phase 3: Synthesis
- [ ] spec_synthesize with LLM generation
- [ ] Source validation against workflow DAG
- [ ] Dry-run preview mode

### Phase 4: Agent Workflows
- [ ] Agent rule files (similar to aidlc-workflows)
- [ ] Claude Code workflow scripts for orchestration

## Testing Strategy

1. **Unit tests**: Service layer logic, workflow resolution
2. **Integration tests**: MCP tool handlers with mock store
3. **E2E tests**: Full workflow from select → create → evaluate → synthesize

## Security Considerations

1. **File access**: Spec operations scoped to initiative directories only
2. **LLM calls**: Rate limiting for evaluate/synthesize operations
3. **Secrets**: No spec content containing secrets in database (file path only)

## Dashboard Spec Viewer Implementation

### URL Routing

Add route for deep-linked specs:

```typescript
// App.tsx routes
<Route path="/initiative/:initId/spec/:specType" element={<SpecViewer />} />
```

### API Endpoint

Add endpoint for single spec retrieval:

```go
// GET /api/specs/{initiative_id}/{spec_type}
func (s *Server) handleGetSpec(w http.ResponseWriter, r *http.Request) {
    initID := chi.URLParam(r, "initiative_id")
    specType := chi.URLParam(r, "spec_type")
    // Return: { specType, path, content, modTime }
}
```

### SpecViewer Component

```typescript
// web/src/panels/SpecViewer.tsx
import { useState } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import MarkdownIt from 'markdown-it'

type ViewMode = 'display' | 'markdown'

export function SpecViewer() {
  const { initId, specType } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const [mode, setMode] = useState<ViewMode>(
    (searchParams.get('mode') as ViewMode) || 'display'
  )
  
  // Fetch spec content
  const { data: spec } = useQuery(['spec', initId, specType], 
    () => getSpec(initId!, specType!)
  )
  
  const md = new MarkdownIt({ html: true, linkify: true, typographer: true })
  const renderedHTML = spec ? md.render(spec.content) : ''
  
  return (
    <div>
      {/* Mode Toggle */}
      <div className="flex gap-2 mb-4">
        <button onClick={() => setMode('display')}>Display</button>
        <button onClick={() => setMode('markdown')}>Markdown</button>
        <button onClick={handleDownloadPDF}>Download PDF</button>
        <button onClick={copyLink}>Copy Link</button>
      </div>
      
      {/* Content */}
      {mode === 'display' ? (
        <div className="prose" dangerouslySetInnerHTML={{ __html: renderedHTML }} />
      ) : (
        <pre className="font-mono">{spec?.content}</pre>
      )}
    </div>
  )
}
```

### PDF Export

```typescript
async function handleDownloadPDF() {
  const element = document.querySelector('.prose')
  if (!element) return
  
  // Option 1: Browser print
  window.print()
  
  // Option 2: html2pdf.js
  const html2pdf = await import('html2pdf.js')
  html2pdf.default()
    .set({
      margin: 10,
      filename: `${initId}-${specType}.pdf`,
      html2canvas: { scale: 2 },
      jsPDF: { format: 'letter' }
    })
    .from(element)
    .save()
}
```

### Shared Markdown Components

Import from web-tools (or inline if not published to npm):

```typescript
// Option 1: npm package
import { parseMarkdown, proseStyles } from '@grokify/markdown-editor'

// Option 2: Copy shared utilities
// Copy from ~/go/src/github.com/grokify/web-tools/packages/markdown-editor/src/utils/markdown.ts
```

### Prose Styles

Use the prose styles from markdown-editor for consistent rendering:

```css
.prose { /* from mde-preview.ts proseStyles */ }
.prose h1 { font-size: 2em; border-bottom: 1px solid var(--border); }
.prose h2 { font-size: 1.5em; }
.prose code { background: var(--bg-tertiary); padding: 0.2em 0.4em; }
.prose pre { background: var(--bg-secondary); overflow-x: auto; }
.prose table { width: 100%; border-collapse: collapse; }
/* ... */
```
