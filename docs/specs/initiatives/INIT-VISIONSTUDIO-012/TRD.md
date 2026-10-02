# TRD: Phase-Level Work Locking for Multi-Agent Coordination

**Initiative:** `INIT-VISIONSTUDIO-012`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Proposed
**Date:** 2026-10-02

## TD-1: New `PhaseAssignment` entity, not an optional edge on `Assignment`

`ent/schema/assignment.go`'s `roadmap_item` edge is `.Unique().Required()`,
and every existing call site (`isClaimed`, `ResolveAssignment`,
`CompleteWorkByRef`, …) assumes an `Assignment` always resolves to exactly
one RMI. Making that edge optional and adding a parallel `phase` edge would
touch all of them for a marginal schema savings. A separate `PhaseAssignment`
entity — same field shape as `Assignment` (`worker`, `status`,
`lease_expires_at`, `workspace`, `handoff`, timestamps), edged to `Phase`
instead — keeps `Assignment`'s existing invariants untouched and is a purely
additive migration (new table only).

## TD-2: Reuse `pkg/assignment`'s pure lease functions

`Claim`/`Renew`/`Release`/`Complete`/`ExpireStale` in `pkg/assignment/assignment.go`
already operate on field values (`Worker`, `Status`, `LeaseExpiresAt`, …), not
ent types — they're reusable for `store.PhaseAssignment` as-is, no fork.
Only the id-prefix scheme differs (`assign-<rmi-id>-<ts>` →
`phaseassign-<phase-id>-<ts>`).

## TD-3: Enforcement lives in `WorkReady` and `ClaimRMI`, not a new gate layer

Two call sites change in `pkg/service`:

- `WorkReadyFilters` gains `Worker string`. `WorkReady`'s loop adds a check
  after `isBlockedByDeps`/`isClaimed`: look up the candidate RMI's phase via
  `GetActivePhaseAssignment(ctx, rmi.PhaseID)`; if active and
  `Worker != filters.Worker` (or `filters.Worker == ""`), skip it.
- `ClaimRMI` gets the same check before creating the RMI-level `Assignment`,
  returning `pcerr.StateConflict` ("phase INIT-X-001/phase-1 is reserved by
  worker Y until <time> — ask them to release it, or wait for lease expiry")
  so a worker can't bypass a phase lock by claiming RMIs directly.

Reservation never auto-creates RMI assignments — the holder still calls
`work claim` per RMI, so the git-trailer-per-RMI discipline (`Refs:
RMI-X-NNN` on the commit that actually does the work) is unaffected.

## TD-4: CLI — opt-in flag on the existing command, plus standalone verbs

`work claim-phase` gets a `--reserve` flag that also calls `ReservePhase` in
the same invocation (the common "I'm taking this whole phase" case, one
command). Standalone `work reserve-phase <phase-id>` /
`work release-phase <phase-id>` / `work renew-phase <phase-id>` cover
reserving before any RMI in the phase is `ready` yet (e.g. everything still
`proposed`) and releasing/renewing independent of individual RMI assignments.
No flag on `claim-phase` means unchanged behavior — existing scripts/agents
are unaffected.

## TD-5: Lease duration stays explicit, not re-defaulted

`DefaultLease` (4h) in `pkg/assignment` stays as-is for RMI claims. Phase
reservations reuse the same `--lease-hours` flag and the same
`assignment.DefaultLease` fallback rather than hardcoding a longer phase
default — a phase can be a 20-minute chore or a multi-day body of work, and
baking in an assumption (e.g. 24h) would be wrong for the former and too
short for the latter. CLI `--help` text documents that phase leases
typically want a longer `--lease-hours` than the RMI default, but doesn't
enforce it.

## Testing

- `pkg/store`: round-trip tests for `PhaseAssignmentStore` mirroring the
  existing `AssignmentStore` test suite (`assignment_test.go`).
- `pkg/service`: conflict (reserve while active reservation exists),
  expiry (reserve succeeds once prior lease lapses), cross-worker `ClaimRMI`
  rejection, and same-worker `ClaimRMI`/`WorkReady` pass-through for a
  reserved phase.
- `cmd/visionstudio`: CLI smoke tests for `reserve-phase`/`release-phase`/
  `renew-phase` and `claim-phase --reserve`, matching the existing
  `work_test.go` style.
