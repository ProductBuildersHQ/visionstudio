# ROADMAP: Phase-Level Work Locking for Multi-Agent Coordination

**Initiative:** `INIT-VISIONSTUDIO-012`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Proposed
**Date:** 2026-10-02

RMI IDs use the repo slug `VISIONSTUDIO`. Phase status is derived from
member RMI statuses and is never set directly.

## Phase 1 — Schema & Store

**Theme:** `PhaseAssignment` exists as a queryable entity, mirroring `Assignment`

- [ ] `RMI-VISIONSTUDIO-559` Ent schema: `PhaseAssignment` entity (worker, status, lease_expires_at, workspace, handoff, timestamps) edged to `Phase`; codegen
- [ ] `RMI-VISIONSTUDIO-560` `store.PhaseAssignment` struct and `PhaseAssignmentStore` interface (Create/Get/GetActive/ListActive/Update), mirroring `AssignmentStore`
  - Depends on: `RMI-VISIONSTUDIO-559`
- [ ] `RMI-VISIONSTUDIO-561` Ent-backed implementation of `PhaseAssignmentStore`
  - Depends on: `RMI-VISIONSTUDIO-560`
- [ ] `RMI-VISIONSTUDIO-562` Round-trip store tests for `PhaseAssignment`, mirroring `assignment_test.go`
  - Depends on: `RMI-VISIONSTUDIO-561`

## Phase 2 — Service Enforcement

**Theme:** A reservation actually blocks other workers from the phase's RMIs

- [ ] `RMI-VISIONSTUDIO-563` `pkg/service`: `ReservePhase`/`ReleasePhaseReservation`/`RenewPhaseLease`, reusing `pkg/assignment`'s pure lease functions against `store.PhaseAssignment`
  - Depends on: `RMI-VISIONSTUDIO-561`
- [ ] `RMI-VISIONSTUDIO-564` `WorkReadyFilters.Worker` field; `WorkReady` skips RMIs whose phase has an active reservation held by a different worker
  - Depends on: `RMI-VISIONSTUDIO-563`
- [ ] `RMI-VISIONSTUDIO-565` `ClaimRMI` rejects claiming an RMI whose phase is reserved by a different worker (`pcerr.StateConflict`)
  - Depends on: `RMI-VISIONSTUDIO-563`
- [ ] `RMI-VISIONSTUDIO-566` Unit tests: reservation conflict, expiry, cross-worker rejection, same-worker pass-through for both `WorkReady` and `ClaimRMI`
  - Depends on: `RMI-VISIONSTUDIO-564`, `RMI-VISIONSTUDIO-565`

## Phase 3 — CLI & Docs

**Theme:** Exposed, documented, and proven against a real multi-agent initiative

- [ ] `RMI-VISIONSTUDIO-567` CLI: `work reserve-phase`/`release-phase`/`renew-phase`; `claim-phase` gains `--reserve` flag
  - Depends on: `RMI-VISIONSTUDIO-563`
- [ ] `RMI-VISIONSTUDIO-568` CLI: `work status` lists active phase reservations alongside RMI assignments
  - Depends on: `RMI-VISIONSTUDIO-563`
- [ ] `RMI-VISIONSTUDIO-569` Docs: CLI/dashboard guide update; dogfood by reserving `INIT-STANDARDSENGINE-001/phase-1` from one session and confirming a second session's `work ready`/`work claim` excludes it
  - Depends on: `RMI-VISIONSTUDIO-567`, `RMI-VISIONSTUDIO-568`, `RMI-VISIONSTUDIO-566`
