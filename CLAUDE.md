# CLAUDE.md — VisionStudio

Project-specific guidelines for Claude Code in VisionStudio.

## Project Overview

VisionStudio is an LLM-powered specification authoring and evaluation tool. It provides a dashboard for managing initiatives, roadmap items, and spec quality via LLM-as-a-Judge evaluations. It is the top of the ProductBuildersHQ "spec stack" (`visionstudio → visionspec → specification-workflow-spec`).

Deeper architecture docs live in `docs/architecture/` (`overview.md`, `daemon.md`, `frontend.md`, `types.md`, `ecosystem.md`). Read those for the full picture; this file is for durable conventions and gotchas, not a structural map.

## Architecture & Entry Points

**Two Go entry points** (`cmd/`):

| Binary | Purpose |
|--------|---------|
| `cmd/visionstudio` | Current cobra CLI + unified daemon (dashboard, db, ingest, spec/initiative/roadmap/maturity commands, MCP). This is the primary one. |
| `cmd/daemon` | Legacy standalone REST server (aidlc/v2mom/capability/etc. handlers). Being superseded by `visionstudio dashboard`. |

**Two frontends** — a migration is in progress:

| Dir | What it is |
|-----|-----------|
| `web/` (`visionstudio-web`) | Current React + Vite SPA served by the Go daemon (`visionstudio dashboard --unified`). Prefer this for new UI work. |
| `desktop/` (`visionstudio`) | Electron app (own `main/` + `renderer/`). Matches the README architecture diagram; older. |

**`go.work` multi-repo workspace** — building requires these sibling repos checked out alongside visionstudio (they are `use`d locally, not via published modules):

- `../../grokify/prism-maturity`, `../../grokify/prism-roadmap`
- `../../plexusone/devfolio`
- `../prism-build`

Do not add `replace` directives to `go.mod` for these — the workspace handles them, and `replace` directives must not be pushed (see Pre-Push Checklist in the global CLAUDE.md).

## Type Pipeline (Go → TypeScript)

VisionStudio uses a **Go-first type pipeline** to keep backend and frontend types synchronized:

```
Go structs (pkg/apitypes/types.go)
    ↓ go generate ./pkg/apitypes
JSON Schema (pkg/apitypes/schema/*.schema.json)
    ↓ npm run generate:types (in web/)
Zod schemas (web/src/api/schemas.gen.ts)
    ↓ z.infer<>
TypeScript types (web/src/api/types.gen.ts)
```

### Rules

1. **Go types are the source of truth** — never hand-edit `*.gen.ts` files
2. **Use camelCase JSON tags** — matches JavaScript conventions (`json:"initiativeId"`, not `json:"initiative_id"`)
3. **Use `rubric.Rubric` directly** — JudgeResult.Report uses `rubric.Rubric` from `structured-evaluation` without conversion
4. **Regenerate after changes**:
   ```bash
   go generate ./pkg/apitypes
   cd web && npm run generate:types
   ```

### Compat Layer

The compat layer (`web/src/api/compat.ts`) normalizes optional fields to required fields with defaults. It re-exports the generated rubric types directly:

```typescript
// Re-exported from types.gen.ts (generated from rubric.Rubric)
export type { Rubric, CategoryResult, Finding, Decision, NextSteps }

// JudgeResult uses Rubric directly
interface JudgeResult {
  id: string
  report?: Rubric  // rubric.Rubric from structured-evaluation
}
```

## Database

VisionStudio uses Dolt (MySQL-compatible) via Ent ORM for persistent storage. The store layer (`pkg/store`) handles all database operations.

### Migrations

```bash
# Initialize or migrate database
go run ./cmd/visionstudio db init --migrate

# Generate Ent schema after changes
go generate ./ent
```

### Store vs API Types

| Layer | Package | JSON Style | Purpose |
|-------|---------|------------|---------|
| Store | `pkg/store` | snake_case | Database/internal |
| API | `pkg/apitypes` | camelCase | HTTP responses |

Convert between them in API handlers (see `storeJudgeResultToAPI` in `cmd/visionstudio/api.go`).

## Dashboard

The unified dashboard serves both the React frontend and JSON API:

```bash
# Run dashboard (serves frontend + API on same port)
go run ./cmd/visionstudio dashboard --port 9401 --unified

# Frontend dev (hot reload)
cd web && npm run dev
```

### API Endpoints

| Endpoint | Returns |
|----------|---------|
| `/api/execution` | Programs, initiatives, phases, RMIs |
| `/api/specs` | Spec workflows, judge results |
| `/api/spec-files/{id}` | Spec file contents for an initiative |

## LLM-as-a-Judge

Judge results use `structured-evaluation/rubric.Rubric` format directly. Eval files (`*.eval.json`) must conform to this schema.

**Key fields:**

| Field | Description |
|-------|-------------|
| `intScore` | 1-5 integer score |
| `confidence` | 0.0-1.0 confidence level |
| `pass` | Boolean pass/fail |
| `decision.status` | `pass`, `conditional`, `fail`, or `human_review` |
| `categories[]` | Per-category scores, reasoning, and confidence |
| `findings[]` | Issues with severity, recommendation, and category |
| `nextSteps.immediate` | Blocking actions that must be completed |
| `nextSteps.recommended` | Suggested improvements |

**Eval file location:** `docs/specs/initiatives/*/evaluations/*.eval.json`

**Eval file format:** Must use `rubric.Rubric` schema (schemaVersion: "v2"). See `structured-evaluation` for the full schema.

## Commands

```bash
# Build
go build ./cmd/visionstudio

# Run dashboard
go run ./cmd/visionstudio dashboard --port 9401 --unified

# Database migration
go run ./cmd/visionstudio db init --migrate

# Regenerate types
go generate ./pkg/apitypes
cd web && npm run generate:types

# Lint
golangci-lint run

# Test
go test ./...
cd web && npm test
```

## File Structure

Key locations (not exhaustive — the tree changes; `ls pkg/` for the current package list):

```
cmd/visionstudio/          # Primary CLI + daemon (cobra)
  api.go               # API handlers, store→API converters
cmd/daemon/            # Legacy REST server (being superseded)
pkg/                   # ~28 domain/service packages. Notable:
  apitypes/            #   Canonical API types (camelCase JSON) — source of truth for frontend types
    types.go           #     Structs → JSON Schema
    gen/main.go        #     JSON Schema generator
    schema/            #     Generated JSON schemas
  store/               #   Database layer over ent (snake_case JSON)
  service/             #   Business logic
  webapi/              #   Web API server/helpers
  speceval/            #   LLM-as-a-Judge spec evaluation
  synthesis/           #   LLM spec-document generation
  specworkflow/        #   Workflow definitions
  initiative/ roadmap/ rmi/ maturity/ release/ report/   # domain logic
  assignment/          #   Lease-based agent work claims
  mcpserver/           #   MCP server exposure
  ingest/ reposcan/ evidence/ contextbuild/              # repo import & context assembly
ent/                   # Generated Ent ORM code (~20 schemas; dominates Go LOC)
web/                   # Current React + Vite SPA (visionstudio-web)
  src/api/
    client.ts          #   API client with compat converters
    compat.ts          #   Gen→normalized type converters
    types.gen.ts       #   Generated TypeScript types (never hand-edit)
    schemas.gen.ts     #   Generated Zod schemas (never hand-edit)
  src/panels/          #   Top-level views
  scripts/
    generate-types.mjs #   JSON Schema → Zod/TS generator
desktop/               # Older Electron app (visionstudio) — own main/ + renderer/
docs/
  architecture/        # Read these for the deep structural picture
  specs/               # PRD/TRD/PLAN/ROADMAP + initiatives/ + *.eval.json gates
```

## Dependencies

- **invopop/jsonschema**: Go → JSON Schema generation
- **json-schema-to-zod**: JSON Schema → Zod conversion (npm)
- **structured-evaluation**: LLM-as-a-Judge rubrics
- **entgo.io/ent**: ORM for Dolt/MySQL
