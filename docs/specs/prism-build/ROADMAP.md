# PRISM Control — Roadmap

**Initiative:** `INIT-PRISMCONTROL-001`
**Repository:** `github.com/ProductBuildersHQ/prism-build`
**Status:** All 6 phases completed (delivery_complete)

Follow-on work is tracked as separate initiatives under `PROG-PRISM-CONTROL`:

- `INIT-PRISMCONTROL-002` — Deterministic Context Assembly ([ROADMAP-CONTEXT-ASSEMBLY.md](ROADMAP-CONTEXT-ASSEMBLY.md), TRD §15)
- `INIT-TOKENATTRIB-001` — Token Attribution & Portfolio Reporting ([ROADMAP-TOKEN-ATTRIBUTION.md](ROADMAP-TOKEN-ATTRIBUTION.md), TRD §16)

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-PRISMCONTROL-NNN`. Phase status is derived from member RMIs — a phase is complete only when all its required RMIs are complete. This file migrates into the running PRISM Control system in Phase 5 (dogfood).

## Phase 1 — Foundation & Storage

**Theme:** Prove the stack, model the domain.
**Status:** Completed — 5 of 5 items completed

- [x] `RMI-PRISMCONTROL-001` Ent + Dolt compatibility spike
  - Acceptance: Ent migration (or documented `schema.sql` fallback), CRUD, `CALL DOLT_COMMIT`, and `dolt_history_*` query all demonstrated; decision recorded in TRD §4
- [x] `RMI-PRISMCONTROL-002` Repository scaffold: module, SDK facade skeleton, CI, lint
- [x] `RMI-PRISMCONTROL-003` Core domain types + generated JSON Schemas (invopop + schemago, `//go:embed`)
  - Depends on: `RMI-PRISMCONTROL-002`
- [x] `RMI-PRISMCONTROL-004` Dolt schema + Ent storage layer for all v1 tables
  - Depends on: `RMI-PRISMCONTROL-001`, `RMI-PRISMCONTROL-003`
- [x] `RMI-PRISMCONTROL-005` Store interface, in-memory fake, unit-of-work (tx + Dolt commit)
  - Depends on: `RMI-PRISMCONTROL-004`

## Phase 2 — Coordination Core (CLI)

**Theme:** Plan, claim, and update work through `prismctl`.
**Status:** Completed — 6 of 6 items completed

- [x] `RMI-PRISMCONTROL-006` `prismctl db init|serve` + `registry add|list|scan|deps|unpushed` (gitscan integration via `gogit/scanner`)
  - Depends on: `RMI-PRISMCONTROL-005`
  - Acceptance: `registry scan` imports repos from org directories; `registry deps` shows topological order; `registry unpushed` flags dirty repos
- [x] `RMI-PRISMCONTROL-007` Initiative commands with lifecycle transitions + timestamps; phase commands with derived status
  - Depends on: `RMI-PRISMCONTROL-006`
- [x] `RMI-PRISMCONTROL-008` RMI commands: create/update/get/list, dependency edges, acceptance criteria
  - Depends on: `RMI-PRISMCONTROL-007`
- [x] `RMI-PRISMCONTROL-009` `work ready` query (ready ∧ unblocked ∧ unclaimed; repo/initiative filters)
  - Depends on: `RMI-PRISMCONTROL-008`
- [x] `RMI-PRISMCONTROL-010` Lease-based `claim|renew|release` + `update` with evidence and handoff; claim prints trailer
  - Depends on: `RMI-PRISMCONTROL-009`
- [x] `RMI-PRISMCONTROL-011` `export` JSONL snapshots + manifest into `exports/`
  - Depends on: `RMI-PRISMCONTROL-007`

## Phase 3 — Agent Interface

**Theme:** Make Claude Code sessions first-class users.
**Status:** Completed — 3 of 3 items completed

- [x] `RMI-PRISMCONTROL-012` Session protocol in `CLAUDE.md` (plan → claim → trailer → update → handoff) + product-repo pointer line
  - Depends on: `RMI-PRISMCONTROL-010`
- [x] `RMI-PRISMCONTROL-013` `prismctl mcp` stdio server with 9 core tools (official Go MCP SDK v1.6.1)
  - Depends on: `RMI-PRISMCONTROL-010`
- [x] `RMI-PRISMCONTROL-014` End-to-end agent walkthrough: session plans, claims via MCP, commits with trailer, updates status
  - Depends on: `RMI-PRISMCONTROL-012`, `RMI-PRISMCONTROL-013`

## Phase 4 — Attribution & Reporting

**Theme:** Close the loop from plan to evidence to report.
**Status:** Completed — 5 of 5 items completed

- [x] `RMI-PRISMCONTROL-015` Trailer + conventional-commit parser with attribution precedence (trailer → PR ref → branch)
  - Depends on: `RMI-PRISMCONTROL-005`
- [x] `RMI-PRISMCONTROL-016` `ingest git` with per-repo high-water marks (idempotent)
  - Depends on: `RMI-PRISMCONTROL-015`
- [x] `RMI-PRISMCONTROL-017` `ingest changelog` for `CHANGELOG.json` (structured-changelog)
  - Depends on: `RMI-PRISMCONTROL-015`
- [x] `RMI-PRISMCONTROL-018` `report initiative` (JSON + Markdown): duration, phases, repos, commit distribution, releases, unattributed residual
  - Depends on: `RMI-PRISMCONTROL-016`
- [x] `RMI-PRISMCONTROL-019` `validate`: trailer↔RMI consistency, dependency cycles, expired leases, status coherence
  - Depends on: `RMI-PRISMCONTROL-016`

## Phase 5 — Pilot & Dogfood

**Theme:** Run real initiatives; fix what the pilot breaks.
**Status:** Completed — 5 of 5 items completed

- [x] `RMI-PRISMCONTROL-020` Pilot: model and drive `INIT-UIFORGE-001` across 3 repos with full session protocol
  - Depends on: `RMI-PRISMCONTROL-014`, `RMI-PRISMCONTROL-018`
  - Delivered: piloted with INIT-UIFORGE-001 (uiforge, agentos, agentos-web) — 6 phases, 37 RMIs, plan→claim→execute→complete cycle
- [x] `RMI-PRISMCONTROL-021` Dogfood: import this ROADMAP.md into the running system as `INIT-PRISMCONTROL-001`
  - Depends on: `RMI-PRISMCONTROL-020`
- [x] `RMI-PRISMCONTROL-022` Protocol and tooling refinements from pilot friction; verify PRD §8 success metrics
  - Depends on: `RMI-PRISMCONTROL-020`
  - Delivered: `pkg/pcerr` structured error codes, `rmi update-phase` bulk transitions, `--ready` flag on `claim-phase`, CLAUDE.md session protocol updated with new syntax
- [x] `RMI-PRISMCONTROL-028` RMI ID resolution for work commands (accept RMI IDs, resolve to active assignment)
  - `work complete/release/renew/update` accept RMI IDs (resolve to active assignment internally); `work complete` accepts variadic RMI IDs
- [x] `RMI-PRISMCONTROL-029` Phase-level batch commands: `claim-phase` and `complete-phase`
  - `work claim-phase <phase-id>` claims all ready RMIs; `work complete-phase <phase-id>` completes all in-progress RMIs
  - Depends on: `RMI-PRISMCONTROL-028`
- [ ] `RMI-PRISMCONTROL-042` Replanning tooling: `rmi move` (re-parent to another initiative/phase) and `rmi update --required`
  - Also `phase remove` for empty phases; friction found while splitting follow-on work into `INIT-PRISMCONTROL-002` and `INIT-TOKENATTRIB-001`

## Phase 6 — Ecosystem Integrations (post-v1)

**Theme:** Feed the consumers.
**Status:** Completed — 5 of 5 items completed

- [x] `RMI-PRISMCONTROL-023` VisionStudio read-only SQL datasource: docs + initiative-browsing views
- [x] `RMI-PRISMCONTROL-024` omnidevx join: per-initiative token/effort via `assignments.worker` session IDs
- [x] `RMI-PRISMCONTROL-025` devfolio initiative dimension; structured-changelog RMI-ref schema extension
- [x] `RMI-PRISMCONTROL-026` Release sets + dependency-ordered release planning (topological stages)
- [x] `RMI-PRISMCONTROL-027` `ROADMAP.json`/`ROADMAP.md` projections into product repos

