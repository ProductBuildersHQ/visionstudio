# TRD — INIT-VISIONSTUDIO-002: VisionStudio Cloud — Multi-Tenant Hosted Backend

**Initiative:** INIT-VISIONSTUDIO-002
**Status:** planned

## T1 — Tenancy model

- **Tenant entity:** `tenant_id`, display name, slug (URL-facing), kind
  (`company` | `personal` | `open-source`), created_at.
- **Membership:** Person↔Tenant M2M with role (owner | member). Person is
  the INIT-VISIONSTUDIO-007 entity — coordinate; a person belongs to
  multiple tenants (e.g., company + personal/OSS).
- **Team scope = persona uniformity.** A small AI-native startup team
  (~3 founders who all build — the GitHub-early-founders shape) is just a
  tenant with several members; deliberately supported with no additional
  machinery. Enterprise needs — persona differentiation (PM vs. Eng),
  granular permissions, org controls, SSO — are explicitly out of scope
  for this initiative; owner|member is the entire role model.
- **Tenant ↔ Organization:** a tenant *contains* one or more GitHub orgs
  (PBHQ tenant ↔ ProductBuildersHQ org; plexusone tenant ↔ plexusone org;
  personal tenant ↔ grokify user-org). Org rows (VS-007) gain a tenant
  edge cloud-side.
- **Isolation: database-per-tenant** (settling the original open
  question). Dolt databases are directories — natural isolation, per-tenant
  branch/merge/history/audit, and tenant deletion = database removal.
  Shared-table + tenant-column rejected.

## T1a — Platform substrate: systemforge

The cloud app is built on `grokify/systemforge` (Go) and
`grokify/systemforge-web` (React) — the ecosystem's SaaS platform module,
built to derisk and accelerate web-app buildouts. What it supplies vs.
what stays custom:

**From systemforge (integrate, don't build):**

- Identity: Users, multi-tenant Organizations (name/slug/**plan** —
  freemium-tier-ready), Memberships with roles, transaction-safe
  ownership transfer, OAuth account links, API keys.
- OAuth2 server (Fosite): authorization code + PKCE, client credentials,
  refresh rotation; GitHub/Google social login handlers.
- Sessions: JWT service, BFF pattern, middleware.
- Authorization: RBAC (SpiceDB/ReBAC available; **not used** — the
  persona-uniformity rule keeps owner|member sufficient).
- Web (systemforge-web): @systemforge/{auth,tenant,api-client,shell,
  pages,design-tokens} — the hosted shell chrome; VS-001's panels render
  inside it.

**Entity mapping (pin before RMI-501):** systemforge `Organization` ≈
VisionStudio `Tenant` (its plan field carries the tier);
systemforge `User` = the auth identity behind VS-007's `Person`
(github_login/email identities map to OAuth accounts). VisionStudio
entities reference systemforge IDs rather than duplicating identity.

**Epistemic status:** construction-proven (~215 + ~34 commits, multiple
site buildouts), production-unproven — this is systemforge's first
production deployment (risk logged in the 6-pager §5). Its
multi-property user model (consistent users/app-health across
systemforge apps; properties acquirable/divestable) is a long-term
ecosystem asset, not a launch dependency.

## T2 — Sync contract and mechanism

**Syncs (public-shape entities):** Program, Initiative (incl. lifecycle
timestamps + visibility flag), Phase, RoadmapItem, Release,
InitiativeDependency/RMIDependency, Organization, Person,
Repository (metadata + visibility — **never** `local_path`), SpecDocument
references (paths, not necessarily content), delivery-evidence metadata.

**Never syncs:** `local_path` and machine-local config, raw coding-agent
session logs, raw ACTS evidence and interviews, source code, secrets.
ACTS summaries and locally-redacted excerpts sync only under the ACTS
privacy model (redaction pipeline as sync gate).

**Mechanism:** Dolt-native push/pull to tenant-scoped remotes
(RMI-VISIONSTUDIO-205 spike validates: conflict story for
single-writer-per-project, schema-migration coordination between local and
cloud versions, partial-sync question — whole-DB push vs synced-entity
projection DB). Local CLI: `visionstudio cloud login` /
`visionstudio sync --tenant <slug>`, per-project tenant assignment stored
locally.

**Packaging: standalone Dolt Go module** (decided — the ecosystem's
standard service-integration pattern, precedent `grokify/gogit` for git).
A reusable module (working name `godolt`) owns the Dolt operational
surface: remote management, push/pull/clone, backup/restore,
multi-database (per-tenant) administration, schema-version handshake.

**Server-mode-first (operational lesson, decided):** the embedded Go
driver was tried previously and ruled out — it sustains only one stable
connection, while the real workload is many concurrent AI agent sessions
(Claude Code) hitting the database; `dolt sql-server` (MySQL protocol) is
what we run today and what godolt targets. Design consequences:

- **SQL-first API:** a connection pool against sql-server; Dolt's
  version-control verbs invoked as stored procedures
  (`CALL DOLT_PUSH/DOLT_PULL/DOLT_CLONE/DOLT_BACKUP/...`), so sync works
  over the same MySQL wire as queries.
- **CLI-exec as secondary** for the operational verbs that need it
  (server lifecycle, restore-from-cold, filesystem-level tenant admin).
- **Embedded driver: not a target.**
- Cloud topology fits naturally: one `dolt sql-server` serves multiple
  databases in a directory, so database-per-tenant runs under a single
  server process with per-database access; the local and hosted sides
  speak the same protocol.

VisionStudio consumes godolt; systemforge apps and future tools get Dolt
sync without re-learning it. The RMI-205 spike doubles as the module's
API-shape discovery; RMI-502/503 build against the module, not ad-hoc
exec calls.

## T3 — Public serving MVP (no auth required)

A small Go service (not the SPA) over the tenant databases:

- `GET /t/<tenant>/roadmap.json` — the VS-006 board projection
  (lifecycle-derived columns, release chips), public subset only.
- `GET /t/<tenant>/releases.json` — **emits the releaselog JSON IR** so
  the existing `@grokify/releaselog` widget consumes it by changing only
  its URL. plexusone.dev re-points; PBHQ site embeds the same widget.
- Two-filter rail enforced **at serve time** (repo visibility == public
  AND initiative visibility == public), independent of what synced —
  defense in depth on top of the sync gate.
- CDN-cacheable (Cache-Control + ETag), CORS open — payloads are public
  by construction.
- Static-export fallback (VS-006 RMI-309, PBHQSITE static file) remains
  the degraded mode if the service is down: same JSON shapes.

## T4 — Authenticated remote mode (original scope, unchanged in intent)

- GitHub OAuth (natural fit for the audience) — answering the original
  auth open question; sessions map to Person, memberships gate tenants.
- Hosted web shell = the same React panels as local behind login —
  **depends on INIT-VISIONSTUDIO-001 Phases 2–3** (the SPA), unlike T3.
- IDE backend-picker: base-URL + token switch (HTTP-only frontend
  contract from VS-001 guarantees this).
- Service-layer tenant scoping from the authenticated principal.

## T5 — Cloud ACTS seam

- Free monthly ACTS report generated cloud-side from synced summaries
  (deterministic metrics + redacted excerpts); on-demand/weekly cadence
  reserved for paid tiers later.
- Metering from day one: per-tenant LLM token cost of cloud analyses
  recorded (ACTS self-ROI machinery) — this is the COGS ledger the
  freemium design needs, even while everything is free.

## T5a — Repository layout (decided: open-core split)

**Cloud-only features are proprietary** — the freemium funnel depends on
the cloud version not being freely self-hostable. The line: *capture and
record locally = open; aggregate, serve, and analyze in the cloud =
private.*

- **Public `visionstudio` (open core, drives adoption):** the local app —
  CLI, daemon, local panels in `web/`; entities/store/publicrail; release
  *recording*, ingest, changelog gate; static exports as degraded-mode
  fallback (same JSON shapes as the hosted versions); and the **cloud API
  client** (`visionstudio cloud login`, Dolt push sync-up, fetch-my-
  reports). A logged-in local app integrates hosted roadmap / releaselog /
  ACTS data by *calling* cloud APIs — never by containing cloud code.
- **Private `ProductBuildersHQ/visionstudio-cloud` (drives signups):**
  the cloud product — Go backend **and** React/TS frontend in one repo
  (house-normal): multi-tenant serving (releaselog JSON, hosted roadmap
  board), systemforge auth/tenancy, metering, cloud ACTS report
  generation, tenant remote hosting/provisioning, plus deployment/ops
  (server + dolt sql-server config, per-tenant data layout, proxy/CDN,
  tested backup/restore, runbooks). Depends on the public visionstudio
  module for shared types (private→public imports are friction-free);
  the frontend composes published `@systemforge/*` packages rather than
  forking `web/`.
- **ACTS splits on the same line:** capture, redaction, and Mode-A
  analysis stay local/open (the privacy story is itself the adoption
  story); the cloud report service (RMI-507), longitudinal aggregation,
  and the research-report pipeline are proprietary.
- **RMI re-homing:** when visionstudio-cloud is created and registered,
  serving/cloud RMIs (504–508 and successors) are re-homed under its repo
  slug per the cross-repo RMI convention.

## T6 — Operations

- Hosted Dolt ops baseline before any external traffic: deploy topology,
  backup/restore (tested, not theoretical), upgrade procedure,
  per-tenant size monitoring. This was the "least-traveled part of the
  stack" open question — it gets its own RMI, early.
- **Deployment target:** the existing idle ~$600/month server (repurposed
  — hosts Dolt + serving service comfortably), or a ~$10/month Lightsail
  instance if isolation is preferred (omnideploy's Lightsail target,
  INIT-OMNIDEPLOY-001, is the deploy path in that case). Decide in
  RMI-VISIONSTUDIO-503.

## Risks

1. **Hosted Dolt operational maturity** — mitigated by the ops RMI and by
   dogfood-only tenants until proven.
2. **Sync schema drift** — local and cloud on different schema versions;
   mitigate with schema-version handshake in the sync CLI and
   migration-gated pushes.
3. **Serving private data** — the rail is enforced at sync AND serve;
   `unknown` visibility never serves. Tested with a private-repo fixture
   before either site re-points.
4. **Scope gravity toward SaaS** — billing, third-party onboarding, and
   enterprise auth are explicitly out; the PRD non-goals are the fence.
