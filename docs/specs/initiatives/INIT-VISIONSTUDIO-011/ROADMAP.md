# ROADMAP: Portfolio WIP Management — Status Sweep and Status View

**Initiative:** `INIT-VISIONSTUDIO-011`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Closed
**Date:** 2026-08-16

Source: direct follow-on from manually closing INIT-VISIONSTUDIO-005/009/010
and INIT-AGENTPROTOCOLS-001 this session (see PRD for the two gaps that
prompted it).

## Phase 1 — Initiative Sweep CLI

**Theme:** Find initiatives whose RMI completion has outrun their recorded status, verified against real per-repo git state

- [x] `RMI-VISIONSTUDIO-537` `visionstudio initiative sweep [--format json]` — list non-terminal initiatives (proposed/planned/executing) with ≥1 RMI and all RMIs completed; for each, resolve every distinct repository referenced by its RMIs and report git state (clean/dirty, ahead/behind cached remote-tracking ref, not found locally, not registered) using the same best-effort git-shell-out posture as `registry doctor`. Report-only — never transitions or records anything itself.

## Phase 2 — Dashboard Status View

**Theme:** See the whole initiative portfolio grouped by lifecycle stage, not just by Program

- [x] `RMI-VISIONSTUDIO-538` Dashboard "By Status" view: initiatives grouped into pipeline-ordered columns (proposed/planned/executing/delivery_complete/releasing/released/closed/cancelled), reusing the existing `/api/execution` response (no new endpoint); a Sidebar entry point alongside the existing All Initiatives / By Program navigation; respects `visibleInitiatives` hidden-initiative filtering

## Phase 3 — Docs

**Theme:** New surfaces are documented where the rest of the CLI/dashboard already is

- [x] `RMI-VISIONSTUDIO-539` Document `initiative sweep` in the dashboard guide's RMI CLI section and the new Status view in the dashboard tour; note both in this session's CLAUDE.md if a durable convention emerges (e.g. the "committed + pushed closes an initiative" pattern codified in the sweep report's language)

## Phase 4 — Acceptance Testing Follow-ups

**Theme:** Gaps found using the shipped By Status view and progress bars live, not anticipated by the original PRD/ROADMAP

- [x] `RMI-VISIONSTUDIO-540` Show cancelled RMIs as a distinct red segment in progress bars — `INIT-OMNIAGENT-002` showed `delivery_complete` with a partial blue bar and a long gray tail that was actually ~18% cancelled, indistinguishable from untouched work. Added `cancelledProgress` to `APIInitiative`/`APIPhase` (both dual-struct layers) and a red segment in `ProgressBar` (`ad47edd`)
- [x] `RMI-VISIONSTUDIO-541` `visionstudio app restart` to replace a stale UI+API server — verifying the fix above required manually finding and killing a stale detached UI process serving an old binary, with no CLI path to do it in one step. Added `ui.pid` tracking (mirroring Dolt's `server.pid`) and `app restart` to stop-and-replace in place (`d0e0b5d`)
- [x] `RMI-VISIONSTUDIO-542` Color a resolved progress bar green, not blue, when nothing is left pending — `INIT-OMNIAGENT-002` (82% completed + 18% cancelled = fully resolved) still showed blue, the color meant for real remaining work. `ProgressBar` now shows green whenever completed + cancelled account for the whole bar (`11ac490`)

## Phase 5 — RMI Origin Tracking

**Theme:** Structured provenance for how an RMI's scope was identified — spec vs. found during implementation vs. found during acceptance testing vs. proposed in discussion — as telemetry on how complete specs actually are at initiative start

- [x] `RMI-VISIONSTUDIO-543` `RMI.Origin` field: Ent schema (default `spec`, additive migration, backfilled on existing rows), `pkg/rmi` origin constants + `ValidOrigin`, wired through doltstore/memstore and both dual-struct API layers (regenerated schema/Zod/TS), `--origin` on `rmi create`/`rmi update` plus an ORIGIN column and filter on `rmi list`, and the `rmi_create` MCP tool so an agent can self-tag `origin=implementation`. Proposed directly by the user in conversation — generalizes the lightweight description-phrase convention used in Phase 4 into a real, queryable field

## Phase 6 — Release Candidate Discovery

**Theme:** For a repo about to be released, find every non-terminal initiative touching it and whether it's ready to ride along

- [x] `RMI-VISIONSTUDIO-544` `visionstudio release candidates --repo <repo-id>` — lists every non-terminal initiative with ≥1 RMI in the given repo and classifies it: `ready` (every RMI in every repo the initiative touches is done — full close candidate), `partial` (this repo's RMIs are done but other repos still have open work — record the release but don't close yet), `not_ready` (this repo still has open work), `already_attached` (a release of this repo already lists this initiative). Report-only, same discipline as `sweep`. Proposed directly by the user: complements `initiative sweep`'s initiative-first direction with the repo-first direction an agent actually needs when releasing a specific repo

## Phase 7 — Release Visibility on Initiatives

**Theme:** Surface which repo@tag releases an initiative shipped in, wherever you'd naturally look for it

- [x] `RMI-VISIONSTUDIO-545` Show attached releases on `initiative get` and the Initiative Detail dashboard page — the `Release` entity already recorded repo-scoped tags with an initiative M2M edge, correctly handling multi-repo initiatives (one tag per repo), but neither the CLI nor the dashboard surfaced it. `GetInitiativeDetail` gains a `Releases` field; `initiative get` prints a `Releases:` section; `ExecutionResponse` (both dual-struct API layers) gains a `releases` array; `InitiativeDetail.tsx` shows a "Released in:" chip row per `repo@tag`, linked to the release URL when recorded. Proposed directly by the user, after confirming the underlying data already handled the multi-repo case correctly (`INIT-AGENTPROTOCOLS-001`'s 3 repo tags) — the gap was visibility, not recording

## Phase 8 — Cancelled RMIs Count as Resolved

**Theme:** A cancelled RMI leaves nothing pending, the same as completed — every done-check in the codebase should agree

- [x] `RMI-VISIONSTUDIO-546` Treat cancelled RMIs as resolved in `DerivePhaseStatus`, `sweep`, and `release candidates` — found via `INIT-PRISMCONTROL-003` (13 completed + 8 cancelled RMIs, 0 pending, fully resolved) still showing two phases `in_progress` and one `planned`, with neither `sweep` nor `candidates` flagging it as a close candidate. `DerivePhaseStatus` only counted literal `completed` toward required-work-done, so a phase with a cancelled required RMI never reached the completed-vs-partial branch. Fixed by tracking required-resolved (completed OR cancelled) separately, matching `ProgressBar`'s existing green-vs-blue treatment; `sweep`/`candidates` had the identical gap in their `ready`/`partial` verdicts, fixed the same way, and both now surface the cancelled count alongside the resolved count so the distinction stays visible. Found through direct use of the tools this initiative shipped, on a real case — acceptance testing in the fullest sense
