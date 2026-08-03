# PRD: Workflow-Scoped Spec MCP Server and Agent Workflows

**Initiative:** INIT-VISIONSTUDIO-003  
**Status:** Draft  
**Author:** John Wang  
**Date:** 2026-08-02

## Problem Statement

AI coding agents (Claude Code, Cursor, etc.) need structured access to specification workflows when building products. Currently:

1. **No workflow context** — Agents don't know which specs are required for a given methodology (AWS Working Backwards vs PBHQ-Lite)
2. **No synthesis guidance** — Agents can't determine that Press Release should be written before FAQ, or that TRD depends on PRD
3. **No evaluation integration** — Specs are written without LLM-as-judge quality gates
4. **Fragmented tooling** — visionstudio has MCP tools for initiatives/RMIs but not for spec authoring workflows

## Solution

Extend visionstudio's MCP server with **workflow-scoped spec operations** that let agents:

1. Select and follow a specification workflow (aws-working-backwards, pbhq-lite, pbhq-standard)
2. Create, read, and evaluate specs within that workflow's structure
3. Synthesize specs from sources following the workflow's DAG
4. Track progress through workflow phases and approval gates

## User Stories

### As an AI coding agent...

1. **I want to list available workflows** so I can present options to the user or infer from initiative type
2. **I want to select a workflow for a project** so subsequent spec operations are scoped correctly
3. **I want to see workflow status** showing which specs exist, their evaluation state, and what's next
4. **I want to create specs from workflow templates** so I start with the right structure
5. **I want to evaluate specs against workflow rubrics** so I know when quality gates pass
6. **I want to synthesize specs from sources** following the workflow's dependency rules (MRD → Press → FAQ)
7. **I want to add custom specs beyond the workflow** for project-specific documentation

### As a product manager...

1. **I want agents to follow our methodology** (AWS Working Backwards, PBHQ, etc.) consistently
2. **I want approval gates enforced** before agents proceed to dependent specs
3. **I want visibility into spec status** across the workflow

## MCP Tool Design

### Workflow Tools

| Tool | Description |
|------|-------------|
| `workflow_list` | List available workflows with their required/optional specs |
| `workflow_select` | Activate a workflow for the current project/initiative |
| `workflow_status` | Show current position: specs exist/missing, eval state, gates, next steps |

### Spec Tools (Workflow-Scoped)

| Tool | Description |
|------|-------------|
| `spec_list` | List specs defined by active workflow + status (missing/draft/evaluated/approved) |
| `spec_read` | Read spec content by type |
| `spec_create` | Create spec from workflow template |
| `spec_evaluate` | Run LLM-as-judge evaluation against workflow rubric |
| `spec_synthesize` | Generate spec from sources using workflow synthesis rules |
| `spec_add` | Add custom spec beyond workflow (extensions, project-specific) |

### Context Model

```
Project/Initiative
  └── Active Workflow (selected or inferred from initiative type)
       ├── Required specs: [mrd, prd, trd, roadmap]
       ├── Optional specs: [press, faq, uxd]
       ├── Synthesis rules: {press: [mrd], faq: [mrd, press], ...}
       └── Gates: [stakeholder_review after press, tech_lead after trd]
  └── Additional specs (beyond workflow)
       └── [custom-analysis.md, competitor-research.md]
```

### Workflow Selection Logic

1. **Explicit**: User/agent calls `workflow_select`
2. **Inferred**: From initiative type (feature → pbhq-standard, maintenance → quick-fix)
3. **None**: Freeform mode — all spec operations available without workflow structure

## Built-in Workflows

### quick-fix
- **Use case**: Small fixes, maintenance tasks
- **Required**: ROADMAP.md
- **Optional**: None

### pbhq-lite  
- **Use case**: Refactors, migrations
- **Required**: PLAN.md, ROADMAP.md
- **Optional**: PRD.md, TRD.md

### pbhq-standard
- **Use case**: Features, compliance initiatives
- **Required**: PRD.md, TRD.md, PLAN.md, ROADMAP.md
- **Optional**: UXD.md

### aws-working-backwards
- **Use case**: Major product initiatives
- **Required**: Press Release, FAQ, 6-Pager, PRD.md, TRD.md, PLAN.md, ROADMAP.md
- **Synthesis**: press ← mrd, faq ← [mrd, press], 6p ← [press, faq], prd ← [6p], trd ← [prd]
- **Gates**: Stakeholder review after 6-Pager, Tech lead review after TRD

## Integration Points

### specification-workflow-spec
- Import `pkg/profile` for workflow/profile schema types
- Import `pkg/layout` for filesystem conventions
- Import `pkg/spectype` for spec type registry

### visionspec (future)
- Template loading for spec creation
- Rubric loading for evaluation
- LLM synthesis prompts

### structured-evaluation
- LLM-as-judge evaluation execution
- Finding aggregation and verdicts

## Success Metrics

1. **Adoption**: 80% of new initiatives use workflow-scoped spec tools
2. **Quality**: Specs passing evaluation on first submission increases 40%
3. **Consistency**: Cross-project spec structure variance decreases 60%

## Dashboard Spec Viewer

The visionstudio dashboard should provide an enhanced spec viewing experience:

### Deep Links

Each spec should have a shareable URL:
- `/initiative/{INIT-ID}/spec/{SPEC-TYPE}` (e.g., `/initiative/INIT-VISIONSTUDIO-003/spec/prd`)
- URL updates as user navigates specs
- Copy link button for sharing

### Display Modes

Toggle between two viewing modes:

| Mode | Description |
|------|-------------|
| **Display** | Rendered HTML from markdown (default) |
| **Markdown** | Raw markdown source |

Display mode uses:
- markdown-it for parsing
- Prose/typography styles for headings, lists, tables, code blocks
- Syntax highlighting for code fences
- Mermaid diagram rendering

### PDF Export

Download spec as PDF:
- Uses browser print-to-PDF or html2pdf.js
- Includes spec metadata header (initiative, date, version)
- Preserves formatting from Display mode

### Shared Components

Leverage `@grokify/markdown-editor` from web-tools for:
- `parseMarkdown()` — markdown to HTML
- `MdePreview` — styled preview component
- `proseStyles` — typography CSS

## Out of Scope

- Visual workflow editor (future)
- Custom workflow definition UI (use YAML/Go for now)
- Multi-tenant workflow isolation (single-user CLI focus first)

## Dependencies

- INIT-VISIONSTUDIO-001: Unified backend (for Dolt store)
- specification-workflow-spec repo (schema types)
