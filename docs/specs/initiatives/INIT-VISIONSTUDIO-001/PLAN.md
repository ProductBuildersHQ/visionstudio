# PLAN — INIT-VISIONSTUDIO-001: Unified Multi-Domain Backend + Frontend

## Sequencing Rationale

Three principles drive the order:

1. **The frontend talks HTTP, never imports backend Go code.** So the panels
   don't care which repo hosts the daemon. We exploit this: build the web UI
   against a small JSON API added to prism-control's existing dashboard
   server (which already loads execution, token, maturity, and judge data)
   **before** any code migration. UI value lands early, and the API contract
   is proven by real usage before the backend moves behind it.
2. **Web-first, Electron later (TRD T5).** The unified UI is a local website
   — the successor to the :9400 dashboard review habit. Browser + Vite
   hot-reload revs far faster than Electron packaging, panel code stays
   platform-agnostic, and the same SPA artifact later serves both the
   Electron shell (Phase 6) and the hosted website (INIT-VISIONSTUDIO-002).
3. **Migrate code only after its contract is fixed.** Phase 4's code move is
   mechanical because the HTTP contract (Phase 2) and IR types (Phase 1) are
   already pinned. Migration-before-UI would mean guessing at the API shape.

Phase order:

1. **Phase 1 — IR contracts.** Pure library work, no migration risk; forces
   the json-tag cleanups in source repos everything else relies on. Includes
   the `go.work` workspace so cross-repo iteration has no tag/`go get` cycle.
2. **Phase 2 — Local web foundation.** JSON API on prism-control's server
   (RMI-PRISMCONTROL-122) + unified SPA shell + shared toolkit/chart
   primitives.
3. **Phase 3 — Unified panels.** Specs, execution, maturity radar, and
   devfolio accomplishments + token costs; composed initiative view replaces
   the :9400 Go-template dashboard. Absorbs prism-control Phase 4
   (RMI-PRISMCONTROL-116..119, cancelled/superseded).
4. **Phase 4 — Backend migration.** Move Ent/store/service from prism-control
   to visionstudio; daemon re-serves the *same* JSON API, so the web app
   re-points via base URL only.
5. **Phase 5 — Multi-domain schema + ingest.** Schema growth happens in its
   permanent home.
6. **Phase 6 — Electron alignment.** The IDE becomes a thin shell over the
   proven web app; existing desktop views untouched until then.

Remote/multi-tenant hosted mode is **out of this initiative** — it moved to
INIT-VISIONSTUDIO-002 (proposed). Constraint retained here: nothing in
Phases 1–6 may preclude it (the HTTP-only frontend contract and org columns
already keep the door open).

## Cross-Repo Work Items

Phase 1 touches source repos (additive only):

| Repo | Change |
|------|--------|
| prism-control | json tags on `pkg/store` structs; `prismctl export ir` command |
| devfolio | none expected (`output/devxdashboard` already IR-first) |
| prism-maturity | verify root `prism` package types are JSON-clean |
| prism-roadmap | verify `rmi`/`roadmap`/`goals` types are JSON-clean |
| all + visionstudio | `go.work` workspace for local dev; CI pins tagged versions |

Phase 2 adds two pieces to prism-control before the move: the JSON API
(RMI-PRISMCONTROL-122), which migrates verbatim in Phase 4, and
`prismctl spec sync` (RMI-PRISMCONTROL-123), which reconciles hand-written
spec files on disk into the `Initiative.Specs` map so the specs panel and
`/api/specs` reflect reality.

Phase 4 removes code from prism-control (with deprecation release first).

## Milestones

| Milestone | Definition of Done |
|-----------|--------------------|
| M1: IR composed | `pkg/ir` builds importing all 4 modules; schema generated + linted; go.work in place |
| M2: Web shell live | Unified SPA at localhost rendering execution data from prism-control's JSON API |
| M3: Daily driver | Specs + execution + spend/accomplishments + maturity panels composed; :9400 Go dashboard retired |
| M4: Backend moved | visionstudio daemon serves the same API from Dolt; web app unchanged; prismctl file-mode still green |
| M5: Multi-domain store | `visionstudio ingest` imports devx/maturity/roadmap IR files into Dolt |
| M6: Electron aligned | IDE loads the unified web app as a thin shell |

Remote-mode milestones belong to INIT-VISIONSTUDIO-002.

## Validation Strategy

- Unit tests against `memstore` for all service logic (library-first, no Dolt
  needed in CI).
- `go build ./...` without the `dolt` tag must stay green (pure-Go builds).
- API contract: golden-file tests on JSON API responses, written in Phase 2,
  reused unchanged in Phase 4 to prove the migration preserved the contract.
- Frontend: Vitest component tests for panels; the real gate is daily use —
  M3 means the unified web UI has replaced :9400 reviews.
- Migration (Phase 4): dolt branch backup, migrate, diff row counts.

## Rollback Points

- Phases 2–3 are purely additive (new API endpoints, new web app); rollback
  is trivial; the Go-template dashboard is retired only at M3 parity.
- Phase 4 cutover is a git + dolt branch; prism-control's DB code is deleted
  only after M4 verification, in a separate commit that can be reverted.
