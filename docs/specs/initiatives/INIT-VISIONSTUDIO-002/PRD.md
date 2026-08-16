# PRD — INIT-VISIONSTUDIO-002: VisionStudio Cloud — Multi-Tenant Hosted Backend

**Initiative:** INIT-VISIONSTUDIO-002
**Status:** planned (re-prioritized 2026-08: build now, dogfood-first, no billing)
**Type:** feature
**Workflow:** pbhq-lite

## Problem

VisionStudio is an effective local, single-developer system (one user, one
machine, one Dolt instance). Three forces now pull parts of it into the
cloud:

1. **Hosting our own public assets.** The roadmap board and release log
   need to be *served*, not just statically exported — and we want
   productbuildershq.com to host them. plexusone.dev's existing release
   log (and a future roadmap) could be powered by the same service,
   making PBHQ's cloud the backend for a second site on day one.
2. **Cross-machine, cross-tenant identity.** A single developer belongs to
   multiple contexts — a company, a personal/open-source identity. The
   local system should sync each project to the right tenant.
3. **Aggregation-value features** — ACTS interpretation, forecasting,
   release management, CI/CD monitoring, eventually adoption (Stripe/MAU)
   — get better with history and correlation, and belong in the cloud.

## Vision

**VisionStudio Cloud**: a multi-tenant hosted backend that local
single-developer VisionStudio instances sync to.

- **Tenancy:** every cloud tenant has a `tenant_id`. A person can be a
  member of multiple tenants (company tenant, personal/OSS tenant). One
  local install syncs different projects to different tenants.
- **Local stays fast and private:** the solo developer's loop remains
  local — develop fast, watch token spend. Raw session logs and source
  never sync. The cloud receives the sync contract's entities only.
- **Cloud owns aggregation:** hosted roadmap boards, release log,
  release management, ACTS *interpretation* (judging, coaching,
  longitudinal reports — consuming synced summaries and locally-redacted
  excerpts, never raw transcripts), forecasting, CI/CD monitoring, and
  later adoption analytics.
- **Dogfood first, billing never (yet):** tenants #1–#3 are our own orgs
  (ProductBuildersHQ, plexusone, grokify). No auth needed for the first
  deliverable (public read-only serving); no billing in this initiative at
  all. Freemium tiers (free: 2–5 deletable projects, monthly ACTS report,
  release log, token-vs-conventional-commit report, GitHub CI monitoring;
  paid: cadence/depth/retention, forecasting, adoption connectors, seats)
  are *designed for* here, implemented in a later initiative.

## First workloads (the reason to build now)

1. **Hosted roadmap board** — productbuildershq.com/roadmap served from
   the cloud (VS-006's board data as a public, CDN-cacheable endpoint).
2. **Hosted release log** — serving the releaselog-widget JSON for
   productbuildershq.com/releases *and* plexusone.dev/releases: the
   plexusone site re-points its existing widget from a committed static
   file to a cloud URL. Two sites, one service, zero new UI.

## Goals

1. **G1 — Tenant model.** Tenant entity + Person↔Tenant membership
   (multi-tenant users); database-per-tenant isolation on Dolt.
2. **G2 — Sync contract and sync-up.** Explicit list of entities that
   sync (initiatives, phases, RMIs, releases, orgs, visibility…) and that
   never sync (local paths, raw logs, raw ACTS evidence); local pushes to
   tenant-scoped Dolt remotes.
3. **G3 — Public serving MVP.** Read-only, tenant/org-scoped JSON
   endpoints for roadmap board and release log, enforcing the two-filter
   visibility rail at serve time; both sites re-pointed.
4. **G4 — Authenticated remote mode** (carried from original scope):
   auth, org scoping, hosted web shell, IDE backend-picker.
5. **G5 — Cloud ACTS seam.** Free monthly ACTS report generated
   cloud-side from synced evidence; metering groundwork for tiers.

## Non-Goals

- **Billing/payments** — later initiative; this one meters but never
  charges.
- **Third-party tenants** — design for them, onboard only our own orgs.
- **Moving the local loop to the cloud** — local VisionStudio remains the
  primary developer experience; cloud is sync target and serving layer.
- **ACTS capture in the cloud** — capture and redaction stay local
  (INIT-ACTS-001 privacy model); only interpretation moves.

## Success Criteria

- plexusone.dev/releases renders from a VisionStudio Cloud URL with no
  visible change to visitors.
- productbuildershq.com serves /roadmap and /releases from the cloud.
- A local instance syncs selected projects to two different tenants from
  one machine.
- Private repos/initiatives provably never appear in any served payload
  (two-filter rail at serve time).
- Hosted Dolt has a tested backup/restore procedure.
