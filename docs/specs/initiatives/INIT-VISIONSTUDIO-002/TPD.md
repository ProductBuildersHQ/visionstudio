# TPD — VisionStudio Cloud

**Status: deferred by design — author before the Phase 3 gate, updated
with Phase 1 spike results.**

The TRD carries the architecture decisions made so far (tenancy model,
sync contract, serving design, ops baseline). This TPD is the detailed
technical product design that must exist before Phase 3:

- Sync protocol details from the RMI-VISIONSTUDIO-205 spike (whole-DB
  push vs projection DB, schema-version handshake, conflict handling)

## Spike findings (RMI-VISIONSTUDIO-205, 2026-08-12 — live against the running server)

1. **SQL-wire sync works.** `CALL DOLT_REMOTE('add',…)` +
   `CALL DOLT_PUSH('remote','main')` executed over the MySQL protocol
   against the live sql-server pushed the full database to a `file://`
   remote (2,040 chunks). godolt's SQL-first design is validated.
2. **Full-fidelity round-trip.** `dolt clone` of the pushed remote
   reproduced all rows (243 releases, 36 initiatives) with versioned
   history intact.
3. **Incremental push is delta-aware** — an immediate re-push reported
   "Everything up-to-date."
4. **Whole-DB push violates the sync contract by construction** —
   confirmed empirically: all 35 `local_path` values arrived in the
   clone. Decision: whole-DB push is **dogfood-only** (our own cloud, our
   own data, acceptable through M2); a **synced-entity projection
   database** (or column scrubbing on a sync branch) is REQUIRED before
   any external tenant (M3 gate condition). RMI-VISIONSTUDIO-502's CLI
   ships whole-DB push labeled as such.
5. **Client-side needs CLI-exec** for `clone`/`init` (no server on the
   receiving end at bootstrap) — matching godolt's SQL-first +
   CLI-secondary split.
6. Remote add/remove round-trips cleanly over SQL (`dolt_remotes` table
   readable for listing).

## Backup/restore findings (RMI-VISIONSTUDIO-503, 2026-08-12 — live against the dogfood database)

1. **No SQL-callable backup.** Verified no `DOLT_BACKUP` stored procedure
   exists — `dolt backup` is CLI-exec only (`godolt.BackupAdd/Sync/Restore`).
2. **Safe against a live server.** `dolt backup sync` executed while the
   dogfood `dolt sql-server` was actively serving the same directory —
   no lock conflict, zero downtime, no need to stop the server.
3. **Verified end-to-end against real data** (not a fixture): backup →
   restore into a scratch directory → row counts compared to the live
   server via independent SQL queries — 243 releases, 36 initiatives, 4
   organizations, exact match. Runbook: `visionstudio-cloud/ops/backup-restore.md`.
4. **Open:** production backup target (S3/GCS vs. file), sync cadence,
   and a recurring automated restore-drill are undecided — tracked as
   open items in the runbook, required before M3.
- Deployment topology and hosted Dolt operational runbook (from
  RMI-VISIONSTUDIO-503, promoted from procedure to design doc)
- AuthN/AuthZ design: GitHub OAuth flows, session model, Person mapping,
  tenant-scoping enforcement points
- Metering data model (RMI-VISIONSTUDIO-508)
- Capacity/cost model for the free tier at 10× M3 signups (feeds the M5
  gate review)
