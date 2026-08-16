# ROADMAP: Portfolio WIP Management — Status Sweep and Status View

**Initiative:** `INIT-VISIONSTUDIO-011`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Planned
**Date:** 2026-08-16

Source: direct follow-on from manually closing INIT-VISIONSTUDIO-005/009/010
and INIT-AGENTPROTOCOLS-001 this session (see PRD for the two gaps that
prompted it).

## Phase 1 — Initiative Sweep CLI

**Theme:** Find initiatives whose RMI completion has outrun their recorded status, verified against real per-repo git state

- [ ] `RMI-VISIONSTUDIO-537` `visionstudio initiative sweep [--format json]` — list non-terminal initiatives (proposed/planned/executing) with ≥1 RMI and all RMIs completed; for each, resolve every distinct repository referenced by its RMIs and report git state (clean/dirty, ahead/behind cached remote-tracking ref, not found locally, not registered) using the same best-effort git-shell-out posture as `registry doctor`. Report-only — never transitions or records anything itself.

## Phase 2 — Dashboard Status View

**Theme:** See the whole initiative portfolio grouped by lifecycle stage, not just by Program

- [ ] `RMI-VISIONSTUDIO-538` Dashboard "By Status" view: initiatives grouped into pipeline-ordered columns (proposed/planned/executing/delivery_complete/releasing/released/closed/cancelled), reusing the existing `/api/execution` response (no new endpoint); a Sidebar entry point alongside the existing All Initiatives / By Program navigation; respects `visibleInitiatives` hidden-initiative filtering

## Phase 3 — Docs

**Theme:** New surfaces are documented where the rest of the CLI/dashboard already is

- [ ] `RMI-VISIONSTUDIO-539` Document `initiative sweep` in the dashboard guide's RMI CLI section and the new Status view in the dashboard tour; note both in this session's CLAUDE.md if a durable convention emerges (e.g. the "committed + pushed closes an initiative" pattern codified in the sweep report's language)
