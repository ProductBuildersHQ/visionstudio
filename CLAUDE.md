# CLAUDE.md — VisionStudio

Project-specific guidelines for Claude Code in VisionStudio.

## Project Overview

VisionStudio is an LLM-powered specification authoring and evaluation tool. It provides a dashboard for managing initiatives, roadmap items, and spec quality via LLM-as-a-Judge evaluations. It is the top of the ProductBuildersHQ "spec stack" (`visionstudio → visionspec`). specification-workflow-spec, the former standalone home for the workflow-definition contract, was merged into visionspec in v0.16.0 and archived.

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

**Gotcha — a brand-new top-level `apitypes` struct needs two registrations, not one.** `pkg/apitypes/gen/main.go` lists which Go types get a `.schema.json` file (`go generate ./pkg/apitypes` silently emits nothing for a struct missing from that list). Separately, `web/scripts/generate-types.mjs`'s `schemaNames` array decides which of those `.schema.json` files actually get converted to Zod/TS — it is **not** auto-discovered from the `schema/` directory, so a schema file can exist and still produce no `types.gen.ts`/`schemas.gen.ts` export, failing as `Module has no exported member 'X'` in `compat.ts` with no other error. Add the new type name to both lists.

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

**Gotcha — `/api/execution`'s types are NOT `pkg/apitypes`.** `cmd/visionstudio/api.go` defines its *own* local `APIProgram`/`APIInitiative`/`APIPhase`/`APIRMI`/etc. structs — those are what actually get marshaled for `/api/execution`. `pkg/apitypes` has parallel structs of the same names, but they only exist as the JSON-schema source feeding the `web/` TypeScript pipeline (`pkg/apitypes/gen/main.go` lists exactly which types it schema-generates). Adding a field to one struct without the other means the frontend's generated type either won't have it, or will claim a field the runtime JSON never sends. When adding a field to any `/api/execution`-reachable type, update **both** `pkg/apitypes/types.go` and the matching local struct + its literal construction in `cmd/visionstudio/api.go`, then regenerate (`go generate ./pkg/apitypes && cd web && npm run generate:types`). `Program.Hidden`/`Initiative.Hidden` are the reference example of doing this correctly in both places.

## Dashboard

The unified dashboard serves both the React frontend and JSON API. Two equivalent ways to run it:

```bash
# One-command: starts the database (if needed) and serves UI+API together
visionstudio app start

# Or, for local dev against an already-running database
go run ./cmd/visionstudio dashboard --port 9401 --unified

# Frontend dev (hot reload) — pair with `dashboard --port 9401` (no --unified) for the API
cd web && npm run dev
```

`app start` is the primary entry point end users reach for; `dashboard --unified` is the equivalent for local development from source. See `visionstudio app --help` / `visionstudio ui --help` for the standalone lifecycle commands (`db start/stop`, `ui --address`, etc.).

### API Endpoints

| Endpoint | Returns |
|----------|---------|
| `/api/execution` | Programs, initiatives, phases, RMIs |
| `/api/spend` | Token spend/cost data (Performance panel) |
| `/api/maturity` | Capability models + assessments |
| `/api/scale` | SCALE platform-adoption metrics |
| `/api/scale/report` | SCALE data as a chart-ready report IR |
| `/api/leverage` | Code-leverage/reuse dependency graph |
| `/api/specs` | Spec workflows, judge results |
| `/api/spec-files/{id}` | Spec file contents for an initiative (role-labeled, workflow-ordered) |
| `POST /api/initiatives` | Create an initiative (the API's first and only mutation endpoint; backs the dashboard's New Initiative form) |

All handlers live in `cmd/visionstudio/api.go`; check there directly rather than trusting this table to stay exhaustive as routes are added.

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

## Conventions

### Using the CLI to track work (agents operating in any repo)

The CLI's `--help` is the authoritative, self-contained manual for *using* visionstudio — an agent should learn the entity model, ID formats, prerequisites, and workflows from it rather than from this file or the source. `visionstudio --help` carries the entity model (Program → Initiative → Phase → RMI), ID conventions, and the canonical create sequence; the commands with load-bearing contracts (`registry add`, `initiative create`, `spec init`, `roadmap import`, `phase add`, `rmi create`, `work claim`) each document theirs in their own `--help`, including the exact ROADMAP.md format `roadmap import` parses. **When CLI behavior changes, update the corresponding `Long`/`Example` text in the same commit** — the help being complete enough to use the tool cold is a maintained property, not an accident (added after an agent had to excavate all of this from source).

### Spec workflows: visionspec is the single source of truth

All default (non-user-custom) spec workflow definitions live in `github.com/ProductBuildersHQ/visionspec`'s embedded catalog (25 profiles: `pbhq-lite`, `quick-fix`, `aws-one-way-door`, `aws-two-way-door`, …) — `pkg/workflows`, merged in from the former standalone specification-workflow-spec repo in v0.16.0. VisionStudio only consumes them via `pkg/specworkflow`'s `Loader`; **never** define or fork a workflow locally — the old hardcoded `BuiltInWorkflows()` catalog was removed precisely because it silently diverged from upstream (same IDs, different required-doc sets). Key pieces:

- `specworkflow.Resolve(loader, init)` — the only way to answer "which workflow applies": explicit `Initiative.WorkflowID`, else `DefaultWorkflowForType` (`maintenance`/`refactor`/`migration`→`quick-fix`, else `pbhq-lite`). Never resolve from the DB.
- The `spec_workflows` table is an **index/cache** of the catalog, not a definition source; `visionstudio workflow sync` refreshes it (idempotent, also remaps initiatives off retired IDs — see `retiredRemap` in `pkg/specworkflow/seed.go`).
- Spec-type IDs (`prd`, `opportunity-spec`) ↔ filenames (`PRD.md`, `OPPORTUNITY-SPEC.md`) convert via `specworkflow.SpecFileName` and `deriveSpecType` (api.go) — keep them inverse of each other when adding doc types.
- Two per-initiative workflow records exist and must be kept in step when switching: `Initiative.WorkflowID` (the edge) and the `InitiativeWorkflow` selection row (read by synthesis/eval via `GetWorkflowForInitiative`) — `initiative update --workflow` updates both.
- Profile YAML gotcha (upstream): the execution block's key is `execution:`, not `workflow:` — yaml.v3 silently drops unknown keys, so a wrong key parses as "no execution ordering" with no error (pbhq-lite shipped that way for a while). Several other profiles still carry unparsed keys (`rubric_extensions`, `cadence`, `cycles`, …).
- The authoritative flow definitions for AWS Working Backwards are visionspec's D2 diagrams (`visionspec/docs/diagrams/aws-{one-way-door,two-way-door}-flow.d2`); the profiles are kept conformant via visionspec's own `pkg/workflows/execution_test.go`.

### Hiding an entity from the dashboard

`Program.Hidden` and `Initiative.Hidden` are the established pattern for "excluded from listings/navigation, still reachable by direct URL/ID" — reach for this shape rather than inventing a new one if another entity (e.g. Repository) needs the same capability:

- `ent/schema/<entity>.go`: `field.Bool("hidden").Default(false)` (additive migration, safe on existing rows)
- `pkg/store`: `<Entity>.Hidden bool`, wired through the doltstore Create/Update/entity→store conversion (unconditional `SetHidden(...)`, mirroring `Program`'s doltstore methods)
- Both API type layers (see the `/api/execution` gotcha above)
- `cmd/visionstudio/<entity>.go`: `<entity> hide/show <id>` cobra subcommands (see `setProgramHidden`/`setInitiativeHidden` for the exact pattern — load, no-op if already in the target state, else set + `UpdatedAt = time.Now()` + persist), plus a `HIDDEN` column on the `list` command
- Frontend: filter with `web/src/lib/visibility.ts`'s `isInitiativeVisible`/`visibleInitiatives`/`hiddenInitiativeIds` (or the equivalent for a new entity) at *every* place that entity is listed — grep for existing consumers rather than assuming there's one obvious list view; `MaturityPanel`'s Capability Models tab was fixed after initially missing 3 of 4 places it read initiative data (chips, radar/dimension aggregates, and the assessments table were all separate leaks off the same raw prop)

### Path safety for caller-derived path components

Use `github.com/grokify/mogo/os/osutil`'s `ValidatePathComponent` (allowlist check on a single ID) and `JoinSecure` (containment check via `filepath.Rel`/`filepath.IsAbs`) at request boundaries and immediately before any filesystem call built from user/agent-supplied input. `FindFirstExistingSecure` wraps the common "try N filename patterns for an ID" lookup. These were added upstream to mogo (not a local package) specifically because CodeQL's `go/path-injection` query recognizes the `filepath.Rel`+`filepath.IsAbs` idiom as a sanitizer but does *not* recognize a `strings.HasPrefix`-based containment check wrapped in a helper function — verified empirically, so don't "simplify" back to a manual `strings.Contains(path, "..")` check or a local reimplementation.

### Closing an initiative doesn't require a git tag

An initiative can be transitioned all the way to `closed` once its work is **committed and pushed**, with release notes/CHANGELOG.json prepared for the version it'll eventually ship under — it does not need to wait for that version to actually be tagged. Tagging is a separate, later, batched action once a release owner decides enough initiatives have accumulated to cut. This exists because initiatives close on their own schedule but releases are batched across several of them (see `docs/releases/unreleased.md`'s pattern of holding several initiatives' worth of changes before a version is cut and tagged). Before closing:

- Verify **every distinct repository referenced by the initiative's RMIs**, not just its `Home repo` field — an initiative's RMIs can span multiple repos (e.g. `INIT-AGENTPROTOCOLS-001` touched `agent-protocols`, `mcp-google`, and `omniskill`). `visionstudio initiative sweep` automates finding candidates (all RMIs `completed`, status not yet advanced) and reports per-repo git state, but never transitions or records anything itself — it's a starting point for review, not an auto-closer.
- Two complementary directions, pick based on what you're starting from: `initiative sweep` is initiative-first ("is this initiative done, across every repo it touches?"); `visionstudio release candidates --repo <repo-id>` is repo-first ("I'm about to release this repo — which touching initiatives are `ready` to close, `partial` [this repo's done, others aren't], or already `already_attached`?"). Use `candidates` when releasing a specific repo, since a repo's RMIs are usually a subset of several different initiatives' full scope.
- A `completed` RMI status is not proof the shipped code matches what the RMI describes — spot-check gaps between spec and implementation before trusting the status field (a 12-RMI audit this session found 4 with real gaps despite `completed` status and a matching commit).
- `visionstudio release record --repo <id> --tag <version>` works before the tag exists — it's a DB upsert with no git validation, keyed off a `CHANGELOG.json` entry (add `--strict` to require one). Record it, then `initiative transition <id> releasing`, `released`, `closed` in sequence.
- A "resolved" RMI is `completed` **or** `cancelled` — cancelled work leaves nothing pending, the same as done. Every tool that judges whether work is finished treats them equivalently: `pkg/initiative.DerivePhaseStatus`, `initiative sweep`'s and `release candidates`' `ready`/`partial` verdicts, and `ProgressBar`'s green-vs-blue coloring in the dashboard. `INIT-PRISMCONTROL-003` (13 completed + 8 cancelled RMIs, 0 pending) is the reference example that caught the original gap — `DerivePhaseStatus` only checked for literal `completed`, so two fully-resolved phases showed `in_progress`/`planned` and neither `sweep` nor `candidates` ever flagged the initiative. If you add a new "is this done" check anywhere, treat cancelled as resolved from the start rather than rediscovering this the same way.

### RMI provenance: scope discovered after the initial spec is normal, and gets tracked, not hidden

Real work never matches its spec exactly — an initiative's original PRD/ROADMAP is a starting point, not a ceiling. When new scope turns out to be needed partway through (or after) delivering an initiative, add it as a new RMI (new phase if it doesn't fit an existing one) rather than silently folding it into an existing RMI's implementation or leaving it untracked. Tag it via `RoadmapItem.Origin` (`--origin` on `rmi create`/`rmi update`, or `origin` on the `rmi_create` MCP tool) so the historical record — and eventually telemetry on how complete specs actually are at project start — stays honest:

- `spec` (default) — in the initiative's original PRD/ROADMAP before implementation began
- `implementation` — the agent found it necessary or clearly beneficial while implementing another RMI in the same initiative; not from testing a finished feature, just from being in the code (e.g. `StatusBadge` needing colors for four more statuses, discovered while building the workflow-status board)
- `acceptance_testing` — a human found the gap using the already-shipped result, often after the initiative first closed (reopen it — see above — rather than tracking the fix outside any initiative)
- `discussion` — scope proposed directly by a human in conversation; not derived from testing a shipped artifact or mid-build discovery, but also not part of the original spec

The discipline that keeps this from being scope creep: each addition traces to a **concrete finding** the work surfaced ("I noticed X is broken/missing while doing Y"), never speculation ("this would also be nice" goes back to the human as a new proposal). Small inline fixes that are directly load-bearing for the RMI in progress stay folded into it; only things substantial enough to be their own reviewable unit get a dedicated RMI.

### Rebuilding the embedded web UI

`web/dist/.gitkeep` is the only tracked file in `web/dist/` — the real build output is gitignored. Running `npm run build` (or `go install ./cmd/visionstudio` after it) deletes `.gitkeep` as part of clearing the output dir; `git checkout -- web/dist/.gitkeep` afterward, or a `git status` diff will show a phantom deletion.

## Troubleshooting

See `docs/getting-started/troubleshooting.md` for the full write-up (kept current there; this is a pointer, not the source). Quick index:

- **`Error 1105 (HY000): table "X" does not have column "Y"`** — the local Dolt database is behind the Ent schema in code (a schema-changing commit landed via `git pull`/another session, and nobody ran a migration against *this* database afterward). Fix: `visionstudio db init --migrate`. It's additive-only (adds missing columns/tables, never drops), so it's always safe to run, including as a first troubleshooting step when anything API-related throws a raw SQL error.
- **Dashboard shows old behavior after a fix, or serves stale data** — the running UI+API process predates your rebuild. `visionstudio app restart` replaces a tracked process in place (see the "closing an initiative" convention above for the `ui.pid` mechanism); if it can't find one to stop, the process predates that tracking and needs a manual kill once.
- **`cannot reach the VisionStudio database at 127.0.0.1:13306`** — Dolt itself isn't running, most often because an earlier `app start` (which auto-stops the database it started, on exit) exited. `visionstudio db start`, or `app start`/`app restart` to bring both up together.

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
  apitypes/            #   API types (camelCase JSON) — schema source for the TS pipeline; NOT the runtime source for /api/execution (see Store vs API Types gotcha)
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
