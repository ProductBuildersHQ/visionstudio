# PLAN: Dual-Loop Lifecycle with Waterline Ratification

**Initiative:** INIT-VISIONSTUDIO-004
**Status:** Draft
**Author:** John Wang
**Date:** 2026-08-07

## Approach

Build inside-out: make the model real first (schema + pure derivation +
snapshot), then expose it (batch author + ratify gate), then visualize it
(waterline UI), then measure it. Each phase is independently shippable and
leaves `main` green.

The guiding principle is **reuse over reinvention**. Three existing assets carry
most of the weight:

1. The upstream two-loop catalog (`specification-workflow-spec/pkg/loop` +
   `pbhq-two-loop.yaml`) — the station/gate/seam data already exists.
2. The `pkg/initiative` forward-only state machine — the ratification gate is one
   more guarded transition on a proven pattern.
3. The `pkg/ingest` disk-copy path — the Product Baseline snapshot follows it, so
   capture costs zero coding-agent tokens.

## Sequencing

| Phase | Outcome | Gate to next |
|-------|---------|--------------|
| 1 — Foundation | Stations loadable; waterline field on Initiative; status derivable; baseline snapshot works | `pkg/loopstate` unit tests green; migration is non-breaking |
| 2 — Write path & gate | Agent batch-authors; human ratifies via MCP/CLI; Build gated | End-to-end MCP flow authors → ratifies → unlocks Build |
| 3 — Waterline UI | Dual-loop view with click-to-ratify; screenshot surface | View renders both loops + waterline for a real initiative |
| 4 — Metrics & docs | Loop-duration, drift, endpoint, screenshots | Metrics visible in Maturity panel; docs published |

## Key Decisions (see ADR-001)

- **Waterline = single high-water mark.** Store `ratified_station`; derive every
  station's status. No per-station approval rows. Consistent with "phase status
  is derived, never set directly."
- **Extend `Phase`, do not add a new entity.** Optional station-typing fields let
  execution phases and loop stations coexist without a breaking migration.
- **Unify the gate.** The daemon's AIDLC engine becomes a Builder-Loop instance
  routed through the shared derivation rather than a third independent gate.
- **Mechanical snapshot.** Baseline capture is a server-side byte-copy off disk,
  never an agent summarization pass.

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Upstream `pkg/loop` not yet exported at the pinned version | RMI-206 bumps the module and wraps it behind `pkg/looprepo`; verify the export before Phase 1 proceeds |
| AIDLC unification balloons into a daemon rewrite | Stage it: RMI-215 maps phases + routes checks; a full daemon refactor is a separate initiative if needed |
| UI scope creep on the "screenshot" view | Phase 2 delivers the functional MVP with no UI; Phase 3 is purely presentational and can ship incrementally |
| RMI-ID collision | Confirmed highest existing is `205`; this initiative starts at `206` |

## Definition of Done

- An agent can call `initiative_author` once and land a fully-authored
  dual-loop initiative.
- A human can ratify through a station in one gesture; Build unlocks only at
  `builder.approve`; a Product Baseline is snapshotted at `product.approve`.
- The dual-loop waterline view renders in the `web/` SPA and is documented with
  screenshots.
- All new packages unit-tested; migration verified non-breaking against existing
  initiatives; `main` green.

## Validation

Dogfood on this very initiative: author `INIT-VISIONSTUDIO-004`'s own stations
through the new batch tool, ratify it through the waterline, and use the
resulting view as the documentation screenshot. If the tool cannot cleanly model
the initiative that specified it, the model is wrong.
