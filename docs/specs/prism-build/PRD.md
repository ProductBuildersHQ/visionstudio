# PRISM Control — Product Requirements Document (PRD)

**Status:** Draft
**Date:** 2026-07-23
**Initiative:** `INIT-PRISMCONTROL-001`
**Repository:** `github.com/ProductBuildersHQ/prism-build`

## 1. Problem

We manage an ecosystem of hundreds of git repositories. Initiatives routinely span 4–5 repositories, and we run 4–5 initiatives concurrently — meaning ~25 repositories are under active, coordinated change at any time, worked on by ~5 concurrent Claude Code sessions plus human developers.

Today, coordination state lives in per-repo Markdown files (`README.md`, `docs/specs/ROADMAP.md` with phase checklists). This breaks down at scale:

- There is no stable, machine-readable location for cross-repo initiative state; every agent session must rediscover file paths and conventions per repo.
- There is no safe mechanism for concurrent sessions to claim work without colliding or duplicating effort.
- There is no attribution linking commits, PRs, and releases back to the initiative that motivated them, so initiative progress and cost cannot be measured.
- Phase checklists conflate grouping with completion tracking, allowing items to silently disappear when a phase is marked complete.

## 2. Product Definition

PRISM Control is a **Product Delivery Control Plane**: a headless system that holds the canonical, queryable representation of cross-repository initiatives, their per-repository Roadmap Items (RMIs), work assignments, and delivery evidence — and continuously reconciles planned state against what actually shipped.

It is **not** an execution engine, an issue tracker, a metrics computation engine, or a UI. Execution systems (Claude Code sessions, and later Beads, GitHub Issues, Jira) do the work; measurement systems (omnidevx, devfolio) compute metrics; VisionStudio renders the human view.

## 3. Users and Use Cases

### Claude Code session (primary user)

- Create an initiative and decompose it into per-repo RMIs (planning).
- Query "what work is ready, unblocked, and unclaimed?" — optionally filtered by repository.
- Claim a task with a time-boxed lease; get back the commit trailer to use.
- Update task status with evidence (commit SHAs, PR URLs, test results) and compact handoff notes.
- Release or renew leases; abandoned leases expire and work returns to the pool.

### Developer (single-user initially)

- Review initiative status and phase progress via CLI or VisionStudio.
- Generate an end-to-end initiative report: duration, repos covered, commit counts and conventional-commit distribution, releases shipped.
- Validate ecosystem consistency (dangling references, stale leases, unattributed commits).

### VisionStudio (rendering consumer)

- Read planning/coordination state directly from the Dolt SQL server (read-only) to render initiative lists, progress, and roadmap views.
- Render productivity metrics only via devfolio-generated, disclosure-scoped outputs (existing pattern preserved).

### CI / automation

- Validate commit trailers against known RMIs.
- Ingest git history and `CHANGELOG.json` into delivery evidence.
- Export JSONL snapshots for backup and offline agent context.

## 4. Primary Scenario

1. A developer (or a Claude session on their behalf) creates `INIT-AVATAR-001` spanning 5 repositories, decomposed into RMIs grouped into themed phases.
2. Five Claude Code sessions start concurrently. Each calls `work_ready`, claims an RMI task with a 4-hour lease, and receives the `Refs:` trailer for its commits.
3. Sessions execute in product repos, committing with the trailer, then update task status with evidence and handoff notes via MCP or `prismctl`.
4. A session dies mid-task; its lease expires and the task returns to the ready pool with the handoff record intact. Another session resumes from the handoff.
5. On completion, ingest attributes all commits to RMIs; `prismctl report initiative INIT-AVATAR-001` produces the end-to-end report; VisionStudio renders progress throughout.

## 5. Goals

- **G1** — Single well-known coordination point replacing per-repo README/ROADMAP coordination.
- **G2** — Safe concurrent work distribution for ≥5 simultaneous agent sessions (lease-based claims, atomic via SQL transactions).
- **G3** — Durable initiative→RMI→commit attribution via git trailers, ingested once into the database.
- **G4** — End-to-end initiative measurement: duration (lifecycle timestamps), repos and commits covered, conventional-commit distribution, releases shipped.
- **G5** — Library-first architecture: all logic in reusable Go packages (SDK), with thin CLI (Cobra), MCP server (official Go SDK), and future adapters on top.
- **G6** — Queryable by downstream consumers: VisionStudio (SQL), devfolio/omnidevx (joins on assignments and evidence), future multi-tenant web UI.

## 6. Non-Goals (v1)

- Beads integration (deferred; RMI model must not depend on it).
- Web UI (VisionStudio is the UI; a hosted multi-tenant web app is a future consumer, enabled by keeping `organization` scoping in the schema).
- Release orchestration/execution (release *plan* data lands in a later phase; executing releases stays in existing skills/workflows).
- Metrics computation (omnidevx and devfolio own developer-experience metrics, collection, and pricing; PRISM Control contributes the initiative dimension and stable IDs, and computes only initiative-scoped token *attribution* — Phase 8, FR13).
- `ROADMAP.json`/`ROADMAP.md` projections pushed into product repos (later phase).
- Alignment with existing `prism-*` spec repositories (incorporated later; they are not yet tied to the development workflow).
- Multi-user auth/tenancy (schema-ready via `organization` columns, not implemented).

## 7. Requirements

### Functional

- **FR1** — CRUD for initiatives with lifecycle status and explicit transition timestamps.
- **FR2** — CRUD for RMIs with stable IDs, phase membership, dependencies (including cross-repo edges), and acceptance criteria.
- **FR3** — Phases as themed groupings of RMIs; phase status is always derived from member RMI statuses, never set directly.
- **FR4** — Repository registry (catalog of participating repos with domain, status, dependencies).
- **FR5** — Ready-work query: ready ∧ unblocked ∧ unclaimed, filterable by repo/initiative.
- **FR6** — Lease-based assignments keyed by session identifier (compatible with omnidevx Claude Code collector session IDs); expiry returns work to the pool.
- **FR7** — Delivery evidence: commits (via `Refs:` trailer ingest with per-repo high-water marks), PRs, releases, changelog entries.
- **FR8** — Initiative report (JSON + Markdown): duration, repos, RMI completion, commit distribution by type/scope/repo, releases, unattributed-commit residual.
- **FR9** — JSONL export snapshots committed to git (backup, offline agent context, portability).
- **FR10** — Interfaces: Go SDK (root package facade), `prismctl` CLI, `prismctl mcp` stdio MCP server — all over one shared service layer.
- **FR11** — Validation: trailer↔RMI consistency, dangling dependencies, expired leases, phase/RMI status coherence.
- **FR12** *(Phase 7)* — Deterministic context assembly: a context package for any phase or RMI, built from the execution graph, derived repository set, spec-file references, and prerequisite phase handoffs — reproducible (byte-identical at fixed revisions) so agent sessions start from authoritative state instead of conversation history.
- **FR13** *(Phase 8)* — Token attribution reporting: token spend joined to assignments/RMIs/initiatives (session + time window, workspace fallback, explicit unattributed residual) with initiative and quarterly report modes — initiatives → RMIs → tokens by category and cost by model. PRISM's figures are a consistent subset of devfolio's overall spend (same omnidevx events, same pricing — never a local pricing table); period reports state coverage of overall spend explicitly (TRD §16).

### Non-Functional

- **NFR1** — Dolt is the canonical store; every logical write is a SQL transaction followed by a Dolt commit (auditable history).
- **NFR2** — Unit tests preferred over integration tests; domain logic must be testable without a running Dolt (store interface + fake).
- **NFR3** — Single binary distribution (`prismctl`).
- **NFR4** — Reports and exports must not require Dolt-proprietary features at read time (portable to plain MySQL for a future hosted UI).

## 8. Success Metrics

- The omnidevx effectiveness initiative (pilot) is planned, executed, and reported end-to-end in PRISM Control.
- ≥5 concurrent sessions coordinate without claim collisions or lost work.
- <5% of commits in participating repos during an initiative window are unattributed.
- `prismctl report initiative` answers duration, repos, commit distribution, and releases with no manual data assembly.
- This repository's own roadmap (`INIT-PRISMCONTROL-001`) is migrated from `docs/specs/ROADMAP.md` into the running system (dogfood).

## 9. Future Directions

- Multi-tenant hosted web UI (read replica, auth, org scoping).
- Beads and other execution providers as adapters under RMIs.
- Release sets and dependency-ordered release orchestration.
- Deterministic `ROADMAP.json`/`ROADMAP.md` projections into product repos.
- devfolio initiative-dimension reports; omnidevx per-initiative token/cost joins.
- Alignment with `prism-roadmap` specification IRs.
- Provider cache adapters (Claude cache breakpoints, CacheLane-style lanes) over context packages — evaluate only after Phase 7 deterministic assembly ships.
- Session-scope recommendations (continue/compact/fork) derived from ingested commit evidence overlap.
