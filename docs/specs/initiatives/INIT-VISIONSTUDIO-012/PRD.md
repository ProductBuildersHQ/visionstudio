# PRD: Phase-Level Work Locking for Multi-Agent Coordination

**Initiative:** `INIT-VISIONSTUDIO-012`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Proposed
**Date:** 2026-10-02

## Source

Found while setting up multi-agent coordination on `INIT-STANDARDSENGINE-001`
(`aistandardsio/standards-engine`) across two independent Claude Code
sessions. `work claim-phase` turned out to be a one-shot batch claim of
whatever RMIs are `ready`/unblocked/unclaimed at the moment it runs — not a
standing reservation. Nothing stops a second worker from claiming an RMI the
instant a dependency clears, even if another worker is mid-way through the
same phase and would naturally pick it up next.

Every repo in this ecosystem already treats the phase, not the individual
RMI, as the practical unit of execution and review (`ROADMAP.md`'s own
convention: "review and execution happen by phase, not individual RMI").
VisionStudio's locking primitive doesn't match that reality yet.

## Problem

Program → Initiative → Phase → RMI already gives ordered, grouped epics with
a real dependency graph (`RMIDependency`) — strictly more than a flat,
unordered Jira epic. What's missing is a lock at the grain teams actually
work at: a worker can claim individual RMIs (leased, lease-enforced) but
cannot claim the phase they belong to, so there's no way to tell other
workers "I'm taking this whole phase" and have VisionStudio enforce it as
new RMIs in it unblock over time.

## Goals

- A worker can reserve a phase so other workers' `work ready`/`work claim`
  treat every RMI in it as unavailable — not just the ones currently
  claimed — until the reservation is released, completed, or its lease
  expires.
- Reservation is additive to the existing RMI-level claim/release/complete
  flow, not a replacement: the reserving worker still claims individual
  RMIs and gets individual git trailers exactly as today.
- `work status` shows active phase reservations alongside RMI assignments,
  so any session can see who holds what at a glance.
- Backward compatible: existing scripts/agents calling `work claim-phase`
  without opting in see no behavior change.

## Non-goals

- Distributed consensus or merge of genuinely concurrent edits within one
  phase — if a phase has two simultaneously-unblocked, dependency-free RMIs
  that don't share files (see `RMI-STDENGINE-001`/`-006` as a live example),
  splitting them across workers at RMI grain remains the right call; phase
  locking is for the common case of a single coherent, mostly-sequential
  body of work.
- Any change to Program/Initiative/Phase ordering or the dependency model —
  already adequate for ordered, grouped epics.
- Auto-granting RMI claims to a phase's reservation holder. The holder still
  calls `work claim`/`work claim-phase` explicitly per RMI, preserving the
  one-git-trailer-per-RMI discipline.
