# PLAN — INIT-VISIONSTUDIO-002: VisionStudio Cloud — Multi-Tenant Hosted Backend

## Sequencing Rationale

1. **Public serving before auth.** The first workloads (hosted roadmap +
   release log) are public read-only JSON — they need tenancy and sync but
   **no login, no hosted SPA**. That decouples the cloud start from
   INIT-VISIONSTUDIO-001's unfinished web foundation: T3 is a small Go
   service plus existing widgets. Authenticated remote mode (Phase 3)
   is the part that waits for VS-001 Phases 2–3.
2. **Dogfood tenants only.** Tenants #1–#3 are our own orgs; plexusone.dev
   re-pointing its release widget to a PBHQ-hosted URL is the first
   external-consumer proof with zero third-party risk.
3. **Ops before traffic.** Backup/restore and upgrade for hosted Dolt are
   proven in Phase 1, before any site depends on the service.
4. **Metering before billing.** Phase 4 records per-tenant analysis cost
   from the first cloud ACTS report; billing itself is a later initiative
   with real usage data to price against.

Phase order:

1. **Phase 1 — Tenancy Foundation and Sync-Up.** Tenant entity +
   membership, sync contract + CLI, sync-strategy spike (existing
   RMI-205), hosted Dolt ops baseline.
2. **Phase 2 — Public Hosting MVP.** Serving service (roadmap.json,
   releases.json in releaselog IR shape), PBHQ site re-point,
   plexusone.dev re-point.
3. **Phase 3 — Authenticated Remote Mode.** Existing RMIs 201–204 (auth,
   org scoping, hosted web shell, IDE backend-picker). Blocked on
   INIT-VISIONSTUDIO-001 Phases 2–3.
4. **Phase 4 — Cloud ACTS and Freemium Groundwork.** Free monthly ACTS
   report cloud-side; per-tenant metering; tier definitions on paper.

## Prerequisites (read before starting any RMI)

- Phases 1–2 have **no dependency on INIT-VISIONSTUDIO-001's web
  foundation** — do not wait for it, and do not build any SPA panels here.
- Phase 1 should land after (or coordinate with) INIT-VISIONSTUDIO-007
  RMI-401 (Person/Organization entities — membership hangs off Person)
  and RMI-402/403 (visibility + export rail, reused at serve time).
- Phase 2's releases.json needs Release rows — INIT-VISIONSTUDIO-006
  Phase 1 (entity + ingest) first; the board projection needs VS-006's
  column derivation but NOT its UI panel.
- Phase 4's cloud ACTS report consumes ACTS Phase 2–3 outputs
  (INIT-ACTS-001) and its redaction gate for anything excerpt-level.

## Working Agreements

- New RMIs use the **RMI-VISIONSTUDIO-5xx block** (2xx is fragmented
  across 002/005 and on-disk 004; 3xx = 006; 4xx = 007). Existing
  201–205 keep their IDs and move to the phases where they now belong.
- Commits carry `Refs: RMI-VISIONSTUDIO-NNN` trailers.
- Freemium tier *content* lives in the PRD as design intent; nothing in
  this initiative may hard-code tier gates beyond metering counters.

## Dependencies and Coordination

- **INIT-VISIONSTUDIO-006/007:** entity and rail reuse per Prerequisites;
  the serve-time rail is the same predicate as the export rail
  (RMI-VISIONSTUDIO-403) — one implementation, two call sites.
- **INIT-PBHQSITE-001:** ships static-first regardless; its RMI-004
  stamp-out recipe gains "or point the widget at VisionStudio Cloud" once
  Phase 2 is live. Static remains the degraded-mode fallback.
- **INIT-ACTS-001:** interpretation-in-cloud placement (its PLAN notes
  local-first capture; cloud is a delivery surface, not a capture change).
- **Billing:** future initiative; nothing here blocks on it.
