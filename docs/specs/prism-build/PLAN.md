# PRISM Control — Implementation Plan (PLAN)

**Status:** Draft
**Date:** 2026-07-23
**Initiative:** `INIT-PRISMCONTROL-001`

Execution is organized by phase (see `ROADMAP.md` for the RMI-level breakdown). Each phase has a theme, exit criteria, and builds on the previous. We review and execute by phase, not by individual RMI.

## Phase 1 — Foundation & Storage

**Theme:** Prove the stack, model the domain.

- Ent + Dolt compatibility spike: Ent/Atlas migration vs Dolt DDL, CRUD, `CALL DOLT_COMMIT` via raw SQL, `dolt_history_*` queries. Decide migration mode (Atlas vs hand-managed `schema.sql`) before the schema hardens.
- Repository scaffold: module, root SDK facade skeleton, CI (build/test/lint), golangci-lint config.
- Core domain types with `json` tags; JSON Schemas generated via `invopop/jsonschema`, validated with `schemago lint`, embedded via `//go:embed`.
- Dolt schema + Ent storage layer for `initiatives`, `repositories`, `phases`, `roadmap_items`, `rmi_dependencies`, `assignments`, `delivery_evidence`.
- Store interface + in-memory fake; unit-of-work wrapper (Ent tx + Dolt commit).

**Exit criteria:** spike decision recorded in TRD; `go test ./...` green with domain + service tests running against the fake store; small Dolt integration suite passes locally.

## Phase 2 — Coordination Core (CLI)

**Theme:** Plan, claim, and update work through `prismctl`.

- `prismctl db init|serve`; `registry add|list`.
- `initiative create|list|get|status|transition` (lifecycle timestamps stamped on transition); `phase add|list` with derived status.
- `rmi create|update|get|list` with dependency edges and acceptance criteria.
- `work ready` (ready ∧ unblocked ∧ unclaimed, `--repo`/`--initiative` filters).
- `claim` / `renew` / `release` with lease enforcement in the service layer; `claim` prints the `Refs:` trailer.
- `update` with status, evidence refs, handoff JSON.
- `export` JSONL snapshots + manifest into `exports/`.

**Exit criteria:** two terminal sessions can concurrently claim different RMIs of one initiative with no collisions; a simulated dead session's lease expires and the RMI returns to `work ready` output.

## Phase 3 — Agent Interface

**Theme:** Make Claude Code sessions first-class users.

- Session protocol written into this repo's `CLAUDE.md` (plan → claim → trailer → update → handoff); one-line pointer added to product repos' `CLAUDE.md`.
- `prismctl mcp` stdio server via the official Go MCP SDK: `initiative_list/get/create`, `rmi_create`, `work_ready`, `task_claim/release`, `task_update`, `report_initiative`. Same service methods as the CLI.
- `.mcp.json` template for product repos.

**Exit criteria:** a real Claude Code session plans a toy initiative, claims a task via MCP, commits with the trailer, and updates status — end to end without manual SQL or file editing.

## Phase 4 — Attribution & Reporting

**Theme:** Close the loop from plan to evidence to report.

- Trailer + conventional-commit parser (`pkg/evidence`), attribution precedence (trailer → PR ref → branch name).
- `ingest git` with per-repo high-water marks; `ingest changelog` for `CHANGELOG.json`.
- `report initiative` (JSON + Markdown): duration, phases, repos, commit distribution by type/scope/repo, releases, unattributed residual.
- `validate`: trailer↔RMI consistency, dangling/cyclic dependencies, expired leases, status coherence.

**Exit criteria:** report for a seeded initiative matches hand-computed expectations; ingest is idempotent (re-runs add nothing).

## Phase 5 — Pilot & Dogfood

**Theme:** Run real initiatives; fix what the pilot breaks.

- Pilot: model `INIT-OMNIDEVX-EFFECTIVENESS-001` (omnidevx, omnidevx-core, omni-github, omni-openai, devfolio) and drive it with concurrent Claude Code sessions.
- Dogfood: migrate this repo's `docs/specs/ROADMAP.md` into the running system as `INIT-PRISMCONTROL-001` (RMI IDs already match, so this is a data import).
- Protocol/tooling refinements from pilot friction; measure the success metrics from the PRD (claim collisions, unattributed-commit rate).

**Exit criteria:** PRD §8 success metrics met on the pilot; this project tracks itself.

## Phase 6 — Ecosystem Integrations (post-v1, planned)

**Theme:** Feed the consumers.

- VisionStudio: read-only SQL datasource documentation + any needed views for initiative browsing.
- omnidevx: per-initiative token/effort join via `assignments.worker` session IDs.
- devfolio: initiative dimension on velocity reports; structured-changelog RMI-ref schema extension upstream.
- Release sets + dependency-ordered release planning (tables from the ideation design).
- `ROADMAP.json` / `ROADMAP.md` projections into product repos.

## Phase 7 — Deterministic Context Assembly (planned)

**Theme:** Specs are the cache, not the chat.

Agent sessions should start each phase (and re-ground before each RMI) from a deterministic context package assembled out of the PRISM graph and repository specs, instead of carrying ever-growing conversation history. Design and alignment decisions in TRD §15; ideation source in `IDEATION_CHAT_CACHE-OPTIMIZATION.md`.

- `pkg/contextbuild`: canonical context package for a phase or RMI — graph data, derived repo set (RMI repo + dependency-RMI repos + registered repo dependencies), spec-file references with revisions, stability-classed sections in stable→volatile order.
- Phase handoff projection: aggregate member-RMI assignment handoffs + evidence into one structured artifact; consumed as prerequisite-phase context.
- `prismctl context build <rmi-id|phase-id>` (JSON + Markdown); one MCP `context_build` tool over the same service method.
- Session-protocol update: phase-start orientation and per-RMI re-grounding via `context build`; continue/compact/reset guidance at RMI boundaries.
- Optional explicit `context_spec` overrides on RMIs.

**Exit criteria:** a fresh Claude Code session, given only `prismctl context build` output for a phase of a real initiative, claims and completes an RMI end-to-end without the prior transcript; `context build` output is byte-identical across two runs at the same Dolt/git revisions.

## Phase 8 — Token Attribution & Portfolio Reporting (planned)

**Theme:** Map token spend to the planning graph.

Token telemetry remains owned by omnidevx (collection, pricing) and devfolio (developer reports), stored in the separate `devx` database on the shared Dolt server. PRISM's job is attribution: joining spend to assignments, RMIs, initiatives, and programs, and reporting from the planning side. Design in TRD §16.

- `pkg/tokens`: `TokenSource` interface with two implementations — omnidevx event-JSONL reader (available now) and `devx.token_events` SQL reader (once devfolio's Dolt ingest lands). Reports are source-agnostic.
- Attribution join: `token_event.session_id = assignments.worker` scoped by the assignment's time window (a session may claim several RMIs in sequence); workspace→repository fallback for unclaimed sessions; everything else is the unattributed residual — reported, mirroring unattributed commits.
- Cost from `omnidevx-core`'s embedded pricing — never a local pricing table (devfolio's duplicated table already drifted once; don't repeat that).
- `prismctl report tokens --initiative <id>` and `prismctl report tokens --quarter 2026-Q3` (or `--since/--until`): initiatives → RMIs → tokens by category, cost by model, session counts, residual.
- Dashboard: token/cost columns on the summary, per-phase/per-RMI spend on detail views; program/initiative `description` surfaced and backfilled.

**Exit criteria:** `report tokens --quarter` over a real quarter lists every PRISM-tracked initiative with per-RMI token/cost figures; the three attribution buckets (RMI-attributed, repository-level, out-of-management) sum to devfolio's overall Token Spend for the same window, and managed spend matches devfolio filtered to the same workspaces (same events, same pricing — consistency by construction, TRD §16); the report states coverage (managed ÷ overall) rather than implying it.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Ent/Atlas migrations incompatible with Dolt DDL | Blocks storage layer | Phase-1 spike is the first RMI; fallback: hand-managed `schema.sql`, Ent no-migration mode |
| Trailer discipline slips (humans, squash merges) | Attribution quality degrades | Protocol + `claim` prints trailer; PR-body/branch fallbacks; unattributed residual surfaced in every report; CI warning → error later |
| `dolt sql-server` not running when a session starts | Sessions blocked | `prismctl db serve` helper; `exports/` JSONL as read-only fallback context |
| MCP tool schemas bloat session context | Slower/costlier sessions | Keep tool surface to the 8 core tools; everything else stays CLI-only |
| Scope creep toward Beads/releases/UI | v1 never ships | Non-goals pinned in PRD §6; release tables explicitly deferred |
| Context packages go stale against specs they describe | Agents act on outdated context | Packages reference spec paths + revisions instead of copying content; agents read the files at build-time revisions |
| Context assembly coupled to one provider's cache mechanics | Spec/tooling debt as providers change | Stability classes and canonical ordering live in the package format only; source specs stay provider-neutral (TRD §15) |

## Sequencing Notes

- The pilot (Phase 5) intentionally starts only after reporting (Phase 4) so the pilot produces the first real end-to-end report.
- Commit-trailer convention starts **now** — commits to this repo reference `RMI-PRISMCONTROL-NNN` from Phase 1 onward, so the dogfood migration retroactively attributes the project's own history.
