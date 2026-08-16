# PLAN — INIT-VISIONSTUDIO-006: Roadmap Board — Releases, Shipped Marks, and Public Roadmap View

## Sequencing Rationale

1. **Model → ingest → view → export.** The Release entity and trailer-chain
   ingest come first because they prove the core bet — that existing
   `Refs:` trailer coverage auto-associates releases to initiatives — on
   real history before any UI is built. If backfill associates well, the
   board is immediately populated; if coverage is thin, we learn that while
   the fix (keep carrying trailers) is already policy.
2. **The forcing function ships with the board, not after.** The unshipped
   queue and validate rule are what convert the board from reporting into
   behavior change (the same principle as ACTS: reports that don't change
   behavior are the failure mode). Shipping the board without the nag would
   ossify the current stuck-at-executing state.
3. **Public export is last** because it depends on the visibility flag,
   settled column semantics, and at least one pass of cleanup — the first
   public board should not show a wall of stale Delivered entries.

Phase order:

1. **Phase 1 — Release model and ingest.** Entity, edges, migration, CLI
   verbs, tag scanner with trailer-chain association, historical backfill.
2. **Phase 2 — Board and forcing function.** Web-UI board panel, unshipped
   queue, validate rule, dashboard badge, visibility flag.
3. **Phase 3 — Public export and ACTS handshake.** Static export, site
   publish, per-repo acceptance-mark documentation for ACTS.

## Prerequisites (read before starting any RMI)

- **UI-dependent RMIs are blocked on INIT-VISIONSTUDIO-001 Phases 2–3.**
  The board panel (RMI-305), dashboard badge (part of RMI-306), and
  internal releases panel (RMI-312) target the unified local web app
  (Vite + React SPA + JSON API) that INIT-VISIONSTUDIO-001 delivers in its
  Phases 2–3 — which is **not built yet** (that initiative sits at Phase 1
  in progress). Do not build these panels into the legacy Electron views
  or the :9400 Go-template dashboard. Check INIT-VISIONSTUDIO-001 status
  first; if its web foundation isn't ready, deliver Phase 1 here (entity,
  CLI, ingest, backfill) plus the CLI-facing parts of RMI-306 (unshipped
  queue in `validate`) and defer the panels.
- Schema, CLI, ingest, and backfill work (RMI-301..304) has no UI
  dependency and can start immediately against the current Ent schema.

## Working Agreements

- RMI block: `RMI-VISIONSTUDIO-301+` (1xx and 2xx blocks are taken by
  INIT-VISIONSTUDIO-001/002; 0xx by earlier work).
- Commits carry `Refs: RMI-VISIONSTUDIO-3NN` trailers — this initiative's
  own releases become the first fully-associated test case.
- Backfill quality is measured, not assumed: report the fraction of
  historical tag-range commits carrying trailers.

## Dependencies and Coordination

- **ACTS (INIT-ACTS-001):** consumes per-repo `released_at` (its TRD T10);
  coordinate the store read path in ACTS Phase 3, no blocking dependency
  either direction.
- **schangelog / release skill:** tagging and changelog generation stay
  where they are; this initiative only records outcomes. The
  CHANGELOG.json update is the natural recording moment (TRD T3): the
  release skill gains one final step — `visionstudio release record
  <repo>` (or a repo-scoped ingest) after tagging — so shipped status
  updates ride the existing changelog habit with zero new ceremony.
- **productbuildershq.com:** static export lands in the existing site's
  publish flow; no new hosting.

## First Cleanup Milestone

After Phase 2, run the unshipped queue to zero once: every current
`delivery_complete` initiative either gets its releases attached and moves
to `released`, or is consciously parked/cancelled. That single pass
activates the ACTS quality signal across the whole back catalog.
