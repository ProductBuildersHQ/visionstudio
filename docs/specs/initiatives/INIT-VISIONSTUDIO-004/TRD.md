# TRD: Dual-Loop Lifecycle with Waterline Ratification

**Initiative:** INIT-VISIONSTUDIO-004
**Status:** Draft
**Author:** John Wang
**Date:** 2026-08-07

## 1. Overview

This TRD specifies the technical implementation of the dual-loop lifecycle
(PRD: INIT-VISIONSTUDIO-004). It is grounded in the current codebase:

- `pkg/initiative/initiative.go` — the existing validated forward-only status
  state machine (`proposed → planned → executing → delivery_complete →
  releasing → released → closed | cancelled`) with `Transition()` /
  `ValidTransition()` and lifecycle timestamp stamping. **This is the template
  the ratification gate extends — not a new mechanism.**
- `ent/schema/initiative.go`, `ent/schema/phase.go`, `ent/schema/roadmapitem.go`.
- `pkg/service/initiative.go` (`CreatePhase`, `CreateInitiative`, `CreateRMI`).
- `pkg/mcpserver/server.go` — MCP tool registration.
- `pkg/ingest/eval.go` — the disk-ingest precedent for a mechanical snapshot.
- `web/src/panels/InitiativeDetail.tsx` — `PhaseCard` / `ProgressBar`, the UI
  anchor. `web/src/panels/MaturityPanel.tsx` — data-viz precedent.
- `cmd/daemon/aidlc.go` — an existing, independent phase-gate engine
  (`CanTransitionTo`) that this initiative **unifies**, not duplicates.

## 2. Loop Catalog Integration

### 2.1 Source of truth

Stations, gates, and seams are **not defined in VisionStudio**. They already
exist upstream in `specification-workflow-spec`:

- `pkg/loop/loop.go` — types `System > Loop > Station` with `Station.Actor`
  (`human`|`ai`|`human+ai`), `Station.Gate bool`, `Station.GateAuthority`
  (`human`|`policy`), `AutonomyNote`, and `Seam`.
- `pkg/loops/default/pbhq-two-loop.yaml` — the enumeration:
  - Product Loop: `sense`, `hypothesize`, `define`, `approve`◆, `measure`,
    `validate-grow`.
  - Builder Loop: `accept`◆, `plan`, `approve`◆, `build`, `verify`◆, `ship`◆.
  - Seams: `product.approve → builder.accept` ("Product Baseline"),
    `builder.ship → product.measure` ("Telemetry"),
    `builder.plan → product.define` ("Spec ambiguity escalation").

### 2.2 New loader package: `pkg/looprepo`

- Bump `go.mod` dependency on `specification-workflow-spec` to a version that
  exports `pkg/loop` (currently only `pkg/template`, `pkg/workflow`,
  `pkg/workflows` are imported; `go.mod` pins `v0.2.0`).
- Add `pkg/looprepo` mirroring the existing `pkg/specworkflow` pattern: load the
  default two-loop system, expose an **ordered station list** with a stable
  fully-qualified id `{loop}.{station}` (e.g. `product.approve`,
  `builder.accept`) and an integer `ordinal` for waterline comparison.
- The canonical ordinal ordering is the Product Loop spine followed by the
  Builder Loop spine, joined at the Product Baseline seam:
  `sense(0) hypothesize(1) define(2) product.approve(3) accept(4) plan(5)
  builder.approve(6) build(7) verify(8) ship(9) measure(10) validate-grow(11)`.
  (`measure`/`validate-grow` are post-ship Product-Loop stations.)

## 3. Data Model Changes

### 3.1 `Phase` — station typing (additive, optional)

`Phase` today has only `id`, `sequence_number`, `title`, `theme` and means
"execution roadmap phase." Add **optional** fields so execution phases and loop
stations coexist without a breaking migration:

```go
field.String("loop_id").MaxLen(32).Optional(),        // "product" | "builder"
field.String("station_id").MaxLen(64).Optional(),     // "product.approve", ...
field.Int("station_ordinal").Optional().Nillable(),   // cached from looprepo
```

A `Phase` with `station_id` set is a **loop station**; one without remains a
plain execution phase. No per-phase status field is added — station status is
**derived** (§3.3).

### 3.2 `Initiative` — the waterline high-water mark

Add three fields (mirroring the existing Optional/Nillable timestamp pattern):

```go
field.String("ratified_station").MaxLen(64).Optional(),          // "builder.approve"
field.Time("ratified_at").Optional().Nillable(),
field.String("ratified_by").MaxLen(128).Optional(),             // human identity
```

**One field (`ratified_station`) encodes the entire waterline.** There is no
per-station approval row.

### 3.3 Derived station status

Station status is computed, never stored:

```
external   — station has no Phase row (done off-system, e.g. free web tier)
authored   — Phase row exists, station_ordinal > ratified ordinal
ratified   — Phase row exists, station_ordinal <= ratified ordinal
pending     — station is post-build and not yet reached
```

Add `pkg/loopstate` with a pure function
`Derive(stations []Station, ratifiedStation string) map[string]Status` — no DB,
fully unit-testable.

### 3.4 Product Baseline snapshot

Reuse the `pkg/ingest/eval.go` disk-ingest precedent (which already writes
`SpecDocument`/`JudgeResult` off disk). On ratify crossing `product.approve`,
snapshot the current define specs (PRD/TRD/etc.) into a **versioned, immutable**
`SpecDocument` set tagged `baseline=true` + a monotonically increasing
`baseline_version`. The snapshot is a **byte-copy off disk performed
server-side** — no agent involvement, zero coding-agent tokens.

## 4. State Machine Extension

Extend `pkg/initiative`:

- Add status `StatusAuthored = "authored"` between `proposed` and `planned`.
- Add a guarded transition `authored → planned` **named `Ratify`** that:
  1. requires `ratified_station != ""` and the artifact invariant for that
     station to hold (e.g. `builder.approve` requires a Product Baseline);
  2. stamps `ratified_at` / `ratified_by`;
  3. if the waterline crossed `product.approve`, triggers the §3.4 snapshot.
- **Build unlock:** the existing `planned → executing` transition additionally
  requires `ratifiedOrdinal >= ordinal("builder.approve")`. Below that, Build is
  refused with a typed error naming the missing stations.

Waterline moves are **idempotent and monotonic-forward by default**; moving the
line *backward* (kickback) is allowed and re-locks Build.

## 5. Write Path (MCP)

`pkg/mcpserver/server.go` gains three tools (registration alongside
`initiativeCreateTool`, `rmiCreateTool`):

| Tool | Behavior |
|------|----------|
| `phase_create` | Missing today. Wraps `Service.CreatePhase`; accepts optional `station_id`/`loop_id`; resolves `station_ordinal` via `looprepo`. |
| `initiative_author` | **Batch.** One call writes `Initiative` (status `authored`) + its station `Phase`s + `RMI`s + optional `ideation_source` ref + define `SpecDocument`s, in a transaction. This is the agent's one-pass authoring made atomic. |
| `initiative_ratify` | Sets `ratified_station` (+ `ratified_by`), runs the §4 `Ratify` transition, performs the §3.4 snapshot when crossing `product.approve`. Enforces the artifact invariant; returns the derived station-status map. |

The read-only HTTP API (`cmd/vistudio/api.go`) gains `GET
/api/initiative/{id}/loop` returning stations + derived status + waterline for
the SPA. **Writes stay on the MCP/CLI surface** per existing architecture.

## 6. AIDLC Unification

`cmd/daemon/aidlc.go` already implements a single-loop phase gate
(`workflow.CanTransitionTo(targetPhase, rules)`), independent of
`pkg/initiative`. Per the framework, **AI-DLC *is* a Builder-Loop instance**.
To avoid a third gate engine:

- Map the daemon's AIDLC phases onto Builder-Loop stations via `looprepo`.
- Route AIDLC transition checks through the shared `pkg/loopstate` derivation so
  there is one gate concept. (Full daemon refactor may be staged; the ADR
  records the decision and the interim boundary.)

## 7. Frontend

New component `web/src/panels/DualLoopView.tsx`:

- Renders the Product Loop and Builder Loop as two horizontal tracks joined at
  the Product Baseline seam (echarts or SVG; MaturityPanel is the viz precedent).
- Draws the **waterline** across both tracks: stations at/below `ratified_station`
  filled (ratified), above it outlined (authored), post-build dimmed (pending),
  off-system marked (external). `builder.build` lights up only when the line
  reaches `builder.approve`.
- Interaction: click a station → "ratify through here" → `POST` to the ratify
  action → optimistic fill; artifact-invariant failures surface inline.
- Advanced toggle exposes per-station **checkboxes** for the non-contiguous
  escape hatch.
- Route `/initiative/:initiativeId/loop`, linked from the `Phases` section of
  `InitiativeDetail.tsx` (`PhaseCard`).

This view is the intended **screenshot surface** representing the dual cycle.

## 8. Metrics

Extend `MaturityPanel` / maturity API:

- **Product-Loop duration** = `ratified_at − ideation_source.date`.
- **Time-in-Define vs Build** = `executing_at − ratified_at` vs.
  `ratified_at − created_at`.
- **Spec drift** = diff(current define specs, latest `baseline` snapshot).

## 9. Testing

- `pkg/loopstate` — table-driven unit tests for `Derive` across every waterline
  position (pure, no DB).
- `pkg/initiative` — extend the existing transition tests: `authored → planned`
  invariant, Build-unlock threshold, kickback re-lock.
- `pkg/looprepo` — golden test that the loaded station ordinals match §2.2.
- Snapshot: baseline immutability + versioning on repeated ratify.
- MCP: `initiative_author` transactional rollback on partial failure.

## 10. Migration & Compatibility

- All new Ent fields are Optional/Nillable → **no breaking migration**; existing
  initiatives read as `external` for every station and behave exactly as today.
- Existing execution `Phase`s (no `station_id`) are untouched.
- The `authored` status is additive; initiatives created by the old
  `initiative_create` path continue to start at `proposed`.

## 11. Risks

| Risk | Mitigation |
|------|-----------|
| Two gate engines diverge further | §6 unification; ADR records the single-gate decision |
| Baseline snapshot bloats the store | Immutable + versioned; store diffs, not full copies, past v1 |
| Waterline ordinal disagreements across loops | Single canonical ordering in `looprepo` (§2.2), golden-tested |
| Upstream `pkg/loop` API churn | Pin the bumped version; wrap behind `pkg/looprepo` |
