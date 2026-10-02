# PLAN: Phase-Level Work Locking for Multi-Agent Coordination

**Initiative:** `INIT-VISIONSTUDIO-012`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Proposed
**Date:** 2026-10-02

## Sequencing

**Phase 1 (Schema & Store)** has to land first and alone — everything else
depends on `PhaseAssignment` existing as a queryable entity.

**Phase 2 (Service Enforcement)** is the actual feature: schema with no
enforcement is inert. `ReservePhase`/`ReleasePhaseReservation` land first
(the write path), then the two read-path guards (`WorkReady`, `ClaimRMI`)
that make a reservation actually mean something.

**Phase 3 (CLI & Docs)** exposes it and proves it works: the CLI verbs,
`work status` visibility, and a real dogfood run against
`INIT-STANDARDSENGINE-001`'s Phase 1 — the exact scenario that surfaced the
gap — before calling this done.

## Milestones

| Milestone | Definition of done |
|-----------|--------------------|
| M1: Entity | `PhaseAssignment` ent schema + codegen; store struct and interface defined |
| M2: Store impl | Ent-backed `PhaseAssignmentStore` implementation; round-trip tests pass |
| M3: Reservation lifecycle | `ReservePhase`/`ReleasePhaseReservation`/`RenewPhaseLease` in `pkg/service`, reusing `pkg/assignment`'s pure lease functions |
| M4: Enforcement | `WorkReady` and `ClaimRMI` both respect an active phase reservation held by a different worker; same-worker pass-through verified |
| M5: CLI | `reserve-phase`/`release-phase`/`renew-phase` + `claim-phase --reserve`; `work status` lists phase reservations |
| M6: Dogfood | Reserve `INIT-STANDARDSENGINE-001/phase-1` from one session, confirm a second session's `work ready`/`work claim` correctly excludes it, then release |

## Risks

- **Two independent locks can drift** — a worker reserves a phase but never
  claims any RMI in it, or releases RMIs individually without releasing the
  phase reservation. Mitigated by `work status` surfacing both and lease
  expiry as the backstop (reservations expire like RMI assignments do).
- **Additive migration, low schema risk** — new table only, no changes to
  `Assignment`/`RoadmapItem`/`Phase`'s existing fields or edges.
- **Opt-in by default** — `--reserve` on `claim-phase` is off unless passed,
  so existing automation sees no behavior change until it's updated to use
  the new flag deliberately.
