# ADR-001: Dual-Loop Lifecycle and Waterline Ratification

**Status:** Proposed
**Date:** 2026-08-07
**Initiative:** INIT-VISIONSTUDIO-004
**Deciders:** John Wang

## Context

VisionStudio first records an initiative at the **tail** of its lifecycle. In the
real working flow, an initiative is authored fully-planned in one grounded
coding-agent session (Define → Approve → Accept → Plan), then Build is kicked
off. Two economic and structural facts shape any solution:

1. **The web→coding split is a token boundary, not a loop boundary.** The Product
   Loop runs on a free web-tier agent; the metered coding agent is reserved for
   coding. Provenance capture must therefore cost zero coding-agent tokens.
2. **Stations are traversed in order but compressed.** AI-DLC velocity means the
   pre-build stations are authored together and reviewed once. What collapses is
   the number of *human gates* (many → one), not the *phases*.

The current codebase already contains most of the substrate:

- `pkg/initiative/initiative.go` — a validated forward-only status state machine.
- `ent/schema/phase.go` — a `Phase` (execution roadmap phase; no status field).
- `specification-workflow-spec/pkg/loop` + `pbhq-two-loop.yaml` — a complete,
  **unused** dual-loop station/gate/seam catalog (the module is pinned but only
  `pkg/template`/`pkg/workflow`/`pkg/workflows` are imported).
- `cmd/daemon/aidlc.go` — a **second**, independent phase-gate engine
  (`CanTransitionTo`).
- `pkg/ingest/eval.go` — a disk-copy ingest precedent.
- `web/src/panels/InitiativeDetail.tsx` — per-phase progress UI.

## Decisions

### D1 — Model the lifecycle as a phase catalog with artifact invariants, not a Jira turnstile

Stations are traversed in a canonical order, but the tool does **not** force an
initiative to walk them one transition at a time. Entry is allowed at any
station (initiatives are typically born at `builder.approve`). Integrity is
enforced by **artifact invariants** ("you cannot ratify through `builder.approve`
unless a Product Baseline exists"), not by transition sequence.

*Rejected:* a Jira-style mandatory step-by-step workflow — it does not match a
one-session authoring flow and adds ceremony with no information.

### D2 — Encode approval as a single waterline high-water mark

Because approval in a linear loop is monotonic (approving Plan is meaningless
without Define approved), the entire approval state is captured by **one field**
on `Initiative`: `ratified_station`. Every station's status
(`external`/`authored`/`ratified`/`pending`) is **derived**, never stored
per-station. This is consistent with the existing convention that phase status is
derived from members, never set directly.

*Consequences:* one gesture ("ratify through station X") handles both approve-all
and kickback (drop the line below the bad station → Build re-locks). Per-station
**checkboxes** are retained only as an advanced escape hatch for genuinely
non-contiguous approval.

*Rejected:* a `RatificationEvent` row per station — expresses invalid
non-contiguous states and adds write volume for no gain in a linear loop.

### D3 — Extend `Phase` with optional station typing rather than add a new entity

Add optional `loop_id` / `station_id` / `station_ordinal` to `Phase`. A phase
with `station_id` is a loop station; without it, a plain execution phase. This
avoids a breaking migration and lets the two concepts coexist.

*Rejected:* a separate `LoopStation` entity — duplicates the Phase↔Initiative↔RMI
edges already in place and fragments the progress-rollup UI.

### D4 — The ratification gate extends the existing state machine

Add status `authored` and a guarded `authored → planned` transition (`Ratify`)
to `pkg/initiative`, stamping `ratified_at`/`ratified_by` exactly as existing
transitions stamp `planned_at`/`executing_at`. Build (`planned → executing`)
additionally requires the waterline to have reached `builder.approve`.

*Rejected:* a new parallel gate subsystem — see D6.

### D5 — The Product Baseline snapshot is a mechanical, server-side byte-copy

On ratify crossing `product.approve`, snapshot the current define specs into a
versioned, immutable `baseline` `SpecDocument` set, reusing the `pkg/ingest`
disk-copy path. No agent is invoked. This satisfies the zero-coding-token
constraint and makes the baseline a cached answer future sessions read instead of
re-deriving intent with premium tokens.

*Rejected:* asking the coding agent to assemble/regenerate the baseline — spends
the exact tokens the web→coding split exists to protect.

### D6 — Unify the gate; do not add a third engine

Two phase-gate mechanisms already coexist (`pkg/initiative` and the daemon's
AIDLC `CanTransitionTo`). Per the two-loop framework, **AI-DLC is a Builder-Loop
instance**. AIDLC phases are mapped onto Builder-Loop stations and its checks
routed through the shared `pkg/loopstate` derivation, so the dual-loop model
*generalizes* AIDLC rather than competing with it. A full daemon refactor may be
staged separately; this ADR fixes the direction.

## Consequences

**Positive**

- VisionStudio models *both* loops — closing the same blind spot the two-loop
  framework was written to expose (the tool currently embodies it).
- One gesture ratifies; zero coding-agent tokens for provenance; non-breaking
  migration (existing initiatives read as `external` and behave as today).
- The waterline is simultaneously the app control and the framework diagram — a
  legible, screenshot-worthy artifact.

**Negative / costs**

- Requires an upstream module bump to export `pkg/loop`.
- AIDLC unification is real work; interim, two engines coexist behind a shared
  derivation boundary until fully converged.

## References

- PRD / TRD / ROADMAP / PLAN: `docs/specs/initiatives/INIT-VISIONSTUDIO-004/`
- Upstream catalog: `specification-workflow-spec/pkg/loop`,
  `pkg/loops/default/pbhq-two-loop.yaml`
- Framework: The Product Loop & Builder Loop (ProductBuildersHQ)
