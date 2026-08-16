# PRD: Portfolio WIP Management — Status Sweep and Status View

**Initiative:** `INIT-VISIONSTUDIO-011`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Planned
**Date:** 2026-08-16

> **Spec scope note:** this initiative is intentionally captured as PRD +
> ROADMAP only — no TRD/PLAN. It's reversible internal tooling (a two-way
> door): a CLI report command and a client-side dashboard grouping, both
> additive and low-risk. Add TRD/PLAN later only if a specific item turns out
> to need design debate.

## Source

Direct follow-on from closing four initiatives by hand this session
(`INIT-VISIONSTUDIO-005/009/010`, `INIT-AGENTPROTOCOLS-001`). Two problems
showed up doing that work manually:

1. `INIT-AGENTPROTOCOLS-001` sat at `executing` with all 5 RMIs already
   `completed` — and its work actually spanned three repositories
   (`agent-protocols`, `mcp-google`, `omniskill`), not just its recorded home
   repo. Finding this required manually cross-referencing RMI status against
   real git state in every referenced repo. Nothing surfaces this
   automatically today.
2. With 37 non-closed initiatives in the portfolio (as of this initiative's
   creation), the dashboard has no way to see the pipeline as a whole — only
   grouped by Program, or one initiative at a time. Managing WIP (how many
   things are `executing` vs. sitting `delivery_complete` unreleased vs.
   still `proposed`) requires a manual CLI query every time.

## Problem

There is no low-effort way to (a) find initiatives whose RMI completion has
outrun their recorded lifecycle status, verified against real per-repo git
state rather than trusted RMI status alone, or (b) see the whole initiative
portfolio grouped by where it sits in the lifecycle pipeline.

## Goals

| Area | Delivered as |
|------|--------------|
| Find status-behind-reality initiatives, cross-repo verified | `visionstudio initiative sweep` |
| See portfolio WIP at a glance | Dashboard "By Status" view, alongside the existing By Program view |

## Requirements

| # | Requirement |
|---|-------------|
| FR-1 | `initiative sweep` lists non-terminal initiatives (`proposed`/`planned`/`executing`) that have at least one RMI and every RMI `completed` |
| FR-2 | For each candidate, resolve every distinct `repository_id` referenced by its RMIs (not just the initiative's home repo) |
| FR-3 | For each distinct repo with a registered local path, report git state: clean/dirty working tree, ahead/behind the cached remote-tracking ref, or not found on disk. Repos with no registered local path are reported as unverifiable, not silently skipped |
| FR-4 | `sweep` is report-only — it never calls `initiative transition` or `release record` itself. A human (or agent) reviews the report and acts |
| FR-5 | `--format json` on `sweep`, matching the existing list/get pattern |
| FR-6 | Dashboard gains a status-grouped view: initiatives bucketed into columns by lifecycle status (`proposed` → `cancelled`, pipeline order), reusing the existing `execution` API response — no new endpoint |
| FR-7 | The Status view respects the existing hidden-initiative convention (`visibleInitiatives`) |

## Non-Goals

- No automatic status transitions or release recording — closing an
  initiative always stays a deliberate, reviewed action (per this session's
  established convention: committed + pushed is enough to close, but a human
  decides when).
- No RMI-content-vs-spec verification (i.e., re-checking whether a
  `completed` RMI's shipped code actually matches its written description).
  That kind of review — which caught 4 of 12 RMIs with real gaps in this
  session's manual audit — stays a human/agent judgment call, not something
  `sweep` can mechanically verify.
- No new API endpoint for the Status view; it reuses `/api/execution`.
- No cross-repo `git fetch` inside `sweep` (network calls in a routine
  report command are a footgun); git-state checks read cached refs only,
  same posture as `registry doctor`'s best-effort local checks.

## Success Metrics

- Running `visionstudio initiative sweep` after this session's manual
  INIT-AGENTPROTOCOLS-001 audit reproduces that finding automatically (all
  RMIs completed, 3 repos, all clean) without hand-written scripting.
- The dashboard's Status view answers "what's actually in flight right now"
  in one screen, without switching between per-program views or shelling
  into the CLI.

## Design Fit (why this is low-risk)

- `sweep` follows `registry doctor`'s established pattern: report-only,
  best-effort git shell-outs, non-fatal on lookup failure.
- The Status view is a second grouping function alongside
  `InitiativesOverview`'s existing `GroupedInitiatives` (by Program) —
  additive, no change to the API surface or data model.
