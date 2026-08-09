# ROADMAP: Dual-Loop Lifecycle with Waterline Ratification

**Initiative:** `INIT-VISIONSTUDIO-004`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Draft
**Date:** 2026-08-07

## Phase 1 — Loop Model Foundation

**Theme:** Consume the upstream loop catalog and add the schema to represent stations + the waterline

- [ ] `RMI-VISIONSTUDIO-206` Bump specification-workflow-spec and add `pkg/looprepo` loader (stations, gates, seams, canonical ordinals)
- [ ] `RMI-VISIONSTUDIO-207` Add optional station typing fields to `Phase` Ent schema (`loop_id`, `station_id`, `station_ordinal`)
  - Depends on: `RMI-VISIONSTUDIO-206`
- [ ] `RMI-VISIONSTUDIO-208` Add waterline high-water-mark fields to `Initiative` schema (`ratified_station`, `ratified_at`, `ratified_by`)
- [ ] `RMI-VISIONSTUDIO-209` Add `pkg/loopstate.Derive` — pure derived station-status function (external/authored/ratified/pending)
  - Depends on: `RMI-VISIONSTUDIO-206`, `RMI-VISIONSTUDIO-208`
- [ ] `RMI-VISIONSTUDIO-210` Product Baseline snapshot on ratify (reuse `pkg/ingest` disk-copy pattern; versioned + immutable)
  - Depends on: `RMI-VISIONSTUDIO-208`

## Phase 2 — Batch Write Path & Gate

**Theme:** Author everything in one call; ratify with one gesture; gate Build

- [ ] `RMI-VISIONSTUDIO-211` Add `phase_create` MCP tool (missing today; supports station typing)
  - Depends on: `RMI-VISIONSTUDIO-207`
- [ ] `RMI-VISIONSTUDIO-212` Add `initiative_author` batch MCP tool (initiative + stations + RMIs + ideation ref + define specs, transactional, `authored`)
  - Depends on: `RMI-VISIONSTUDIO-211`
- [ ] `RMI-VISIONSTUDIO-213` Extend `pkg/initiative` state machine with `authored → planned` `Ratify` transition + artifact invariant
  - Depends on: `RMI-VISIONSTUDIO-209`, `RMI-VISIONSTUDIO-210`
- [ ] `RMI-VISIONSTUDIO-214` Add `initiative_ratify` write action (waterline set, snapshot on crossing `product.approve`, Build-unlock at `builder.approve`)
  - Depends on: `RMI-VISIONSTUDIO-213`
- [ ] `RMI-VISIONSTUDIO-215` Unify daemon AIDLC phase-gate with dual-loop (`aidlc.go` → Builder-Loop stations via shared derivation)
  - Depends on: `RMI-VISIONSTUDIO-209`

## Phase 3 — Waterline UI

**Theme:** The screenshot-worthy dual-cycle view

- [ ] `RMI-VISIONSTUDIO-216` `DualLoopView` component — Product + Builder tracks joined at the Product Baseline seam
  - Depends on: `RMI-VISIONSTUDIO-214`
- [ ] `RMI-VISIONSTUDIO-217` Waterline ratify interaction (click-through-station, fill below the line, Build lights up)
  - Depends on: `RMI-VISIONSTUDIO-216`
- [ ] `RMI-VISIONSTUDIO-218` Route `/initiative/:id/loop` + link from `InitiativeDetail` Phases section
  - Depends on: `RMI-VISIONSTUDIO-216`
- [ ] `RMI-VISIONSTUDIO-219` Wire ratify action to backend; optimistic state + artifact-invariant error surfacing
  - Depends on: `RMI-VISIONSTUDIO-217`
- [ ] `RMI-VISIONSTUDIO-220` Checkbox escape-hatch for non-contiguous approval (advanced mode)
  - Depends on: `RMI-VISIONSTUDIO-217`

## Phase 4 — Metrics & Documentation

**Theme:** Autonomy metrics, drift detection, and the write-up

- [ ] `RMI-VISIONSTUDIO-221` Loop-duration + time-in-Define-vs-Build metrics on Maturity API/panel
  - Depends on: `RMI-VISIONSTUDIO-214`
- [ ] `RMI-VISIONSTUDIO-222` Spec-drift detector (latest baseline snapshot vs. current define specs)
  - Depends on: `RMI-VISIONSTUDIO-210`
- [ ] `RMI-VISIONSTUDIO-223` Read-only `GET /api/initiative/:id/loop` endpoint
  - Depends on: `RMI-VISIONSTUDIO-209`
- [ ] `RMI-VISIONSTUDIO-224` Docs + screenshots of the dual-loop waterline view
  - Depends on: `RMI-VISIONSTUDIO-218`

## Notes

- RMI IDs continue from the existing visionstudio sequence (last observed:
  `RMI-VISIONSTUDIO-205`); this initiative starts at `206`.
- Phase status is derived from member RMI statuses, never set directly.
- **Phase 1 is the MVP substrate** — schema + derivation + snapshot make the
  model real even before the UI exists.
- **Phase 2 is the functional MVP** — an agent can batch-author and a human can
  ratify via MCP/CLI, gating Build, with no UI yet.
- **Phase 3 is the payoff** — the waterline view that makes the dual cycle
  legible and screenshot-worthy.
- Commits carry the `Refs: RMI-VISIONSTUDIO-NNN` trailer per org convention.
