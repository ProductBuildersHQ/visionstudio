# TRD — INIT-VISIONSTUDIO-001: Unified Multi-Domain Backend + Frontend

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│  Per-Repo Tools (standalone, file-based, emit JSON IR)                   │
│                                                                          │
│  plexusone/devfolio          grokify/prism-maturity                      │
│  grokify/prism-roadmap       ProductBuildersHQ/prism-control (slimmed)   │
│  ProductBuildersHQ/visionspec                                            │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │ Go module imports (types) + JSON IR files
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  ProductBuildersHQ/visionstudio                                          │
│                                                                          │
│  pkg/ir/         — composed IR: imports types from source modules        │
│  pkg/store/      — Store interfaces + memstore (moved from prism-control)│
│  pkg/store/doltstore/ — Ent-backed Dolt implementation                   │
│  ent/            — Ent schema (moved, then grown to all domains)         │
│  pkg/service/    — service layer shared by daemon, CLI, MCP              │
│  cmd/daemon/     — HTTP API for IDE + website (existing, extended)       │
│  desktop/        — Electron IDE (React + Tailwind, existing)             │
│  web/            — hosted website (same React components)                │
└──────────────────────────────────────────────────────────────────────────┘
         │                                          │
         ▼                                          ▼
  Local mode: Dolt on localhost           Hosted mode: Dolt server,
  (single user, offline-capable)          multi-tenant, authenticated
```

## Dependency Direction (critical)

All imports flow **into** visionstudio; nothing imports visionstudio:

```
visionstudio ──imports──► prism-control/pkg/store   (execution structs)
visionstudio ──imports──► devfolio/output/devxdashboard (devx dashboard IR)
visionstudio ──imports──► prism-maturity (root prism pkg: MaturityModel, cells)
visionstudio ──imports──► prism-roadmap/{rmi,roadmap,goals}
```

This forbids circular repo dependencies. Consequence: prismctl's DB-backed
commands cannot import visionstudio's store. Resolution: DB-backed multi-repo
orchestration migrates to the visionstudio CLI/daemon; prismctl retains
file-mode commands (spec init/validate, roadmap sync, IR export) that need no
database.

## Key Decisions

### T1 — IR composition by import, not duplication

`visionstudio/pkg/ir` contains **no hand-copied structs**. It aliases/embeds
types from source modules and adds only composition types (e.g., a
`RepoSnapshot` that bundles one repo's execution + maturity + roadmap + devx
IR). Source modules keep schema ownership; visionstudio revs by bumping module
versions.

```go
package ir

import (
    pcstore "github.com/ProductBuildersHQ/prism-control/pkg/store"
    "github.com/plexusone/devfolio/output/devxdashboard"
    prismmaturity "github.com/grokify/prism-maturity"
    "github.com/grokify/prism-roadmap/roadmap"
)

type Initiative = pcstore.Initiative // alias: zero drift
type RoadmapItem = pcstore.RoadmapItem

// RepoSnapshot composes all domain IRs for one repository.
type RepoSnapshot struct {
    Repo      string                     `json:"repo"`
    Execution *ExecutionIR               `json:"execution,omitempty"`
    Maturity  *prismmaturity.MaturityModel `json:"maturity,omitempty"`
    Roadmap   *roadmap.Roadmap           `json:"roadmap,omitempty"`
    DevX      *devxdashboard.Dashboard   `json:"devx,omitempty"`
}
```

Precondition: source-module types must be JSON-clean (json tags, no
unexported state). Where a source type is not (e.g., prism-control structs
lack json tags today), the RMI adds tags **in the source repo**.

### T2 — prism-control split

| Stays in prism-control | Moves to visionstudio |
|---|---|
| `pkg/store` struct definitions (Initiative, RMI, …) with json tags added | `Store`/`UnitOfWork` interfaces + `memstore` |
| File-mode commands: `spec init/validate`, roadmap sync, `export` | `ent/` schema + generated code |
| `pkg/specworkflow`, `pkg/initiative` (file logic) | `pkg/store/doltstore` |
| JSON IR export commands (new) | `pkg/service`, dashboard server, MCP server |
| | DB-backed prismctl commands (become visionstudio CLI) |

Struct definitions stay in prism-control so per-repo IR export works with zero
DB dependencies, and visionstudio imports them (T1). Interfaces move because
only implementations (memstore/doltstore) and the service layer use them.

Transition: prism-control marks DB-backed commands deprecated for one release,
pointing at the visionstudio CLI equivalent.

### T3 — Maturity model reconciliation

Two maturity representations exist:

- prism-control `CapabilityModel`/`MaturityAssessment` (dimension × level,
  agent-assessed, in Dolt) — from INIT-PRISMCONTROL-003 Phase 3
- prism-maturity `MaturityModel` (domain × stage cells, KPI-oriented, JSON IR)

Resolution: keep both as distinct IRs (they answer different questions), map
both into the Dolt schema, and let the maturity radar panel render either.
A converter `pkg/ir/maturityconv` provides lossy mapping where useful. No
forced unification in this initiative.

### T4 — Daemon API is the single frontend contract

The existing `cmd/daemon` HTTP API is extended with store-backed endpoints
(`/api/execution/...`, `/api/spend/...`, `/api/maturity/...`). The Electron
renderer and the website consume the same API. The renderer never talks to
Dolt directly. Local vs remote is a base-URL + auth-token switch in the IDE.

### T5 — Frontend: React everywhere, web-first

visionstudio's React + Tailwind + Vite stack is the standard, delivered
**web-first**: the unified UI is a local website (successor to the :9400
dashboard habit) served by the daemon in prod mode and Vite in dev. Panel
code is platform-agnostic; the Electron IDE adopts the same SPA as a thin
shell only after the web app is the proven daily driver (Phase 6), and the
hosted website (INIT-VISIONSTUDIO-002) serves the same artifact behind auth.
Browser + hot-reload revs much faster than Electron packaging; desktop
concerns (IPC, file access, packaging) never touch panel code.

The Go-template dashboard in prism-control remains as-is during transition
(it still works against the old DSN) but receives no new features; it is
retired once the composed web UI reaches parity (M3).

Shared components port from `desktop/renderer/src/components/toolkit` into
the web app's shared package, and chart primitives generalize (pie/bar/line
exist in DevXDashboardView; radar is new for maturity).

### T6 — Multi-tenancy via org scoping

Hosted mode adds `organization` scoping (columns already exist on Initiative,
Repository, Program, MaturityAssessment) enforced in the service layer from
the authenticated principal — not in SQL by the frontend. Local mode runs
with a single implicit org. Tenant isolation at the Dolt level (database per
org vs shared tables) is decided in Phase 5 after load/UX validation.

## Data Flow

1. **Per-repo demo:** tool reads repo files → emits `*.ir.json` → standalone
   HTML render or visionstudio panel reads the file directly.
2. **Ingest:** `visionstudio ingest <repo>` (or daemon watch) imports IR files
   into Dolt, keyed by repo + org.
3. **Live:** daemon queries Dolt via service layer → JSON API → React panels.
4. **Hosted:** same as (3) with auth + org scoping; Dolt runs server-side.

## Build/Tooling Constraints

- Dolt driver requires CGO + ICU (`CGO_CFLAGS=-I/opt/homebrew/opt/icu4c/include`
  etc.); keep the `//go:build dolt` tag pattern so default builds stay pure Go.
- JSON Schema for composed IR generated Go-first via `invopop/jsonschema`,
  linted with `schemago lint`, embedded via `go:embed`.
- Ent for all new Dolt entities; Cobra for CLI; official Go MCP SDK.

## Risks

| Risk | Mitigation |
|------|------------|
| Source-module type churn breaks pkg/ir | Version pinning; aliases isolate call sites; CI compile check |
| prism-control structs get json-tag changes that break existing Dolt rows | Tags added additively; Ent schema owns column names independently |
| Two maturity models confuse users | T3: explicit naming (capability vs domain/stage) in UI |
| Electron + hosted web drift | Single component library; web/ is a thin shell over the same panels |
| Migration breaks existing prismcontrol Dolt data | Phase 2 includes migration script + dolt branch backup before cutover |
