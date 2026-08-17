# VisionStudio Cloud — Multi-Tenant Hosted Backend — Roadmap

**Initiative:** `INIT-VISIONSTUDIO-002`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-<REPOSLUG>-<NNN>`. Phase status is derived from member RMIs — a phase is complete only when all its required RMIs are complete.

New RMIs use the RMI-VISIONSTUDIO-5xx block; RMI-VISIONSTUDIO-201..205 predate this restructure and keep their IDs.

**Milestone gates (canonical in NARRATIVE-6P.md §4 — phases deliver capability, milestones gate investment):**

- **M1 Launched** — both sites serving from cloud 2+ weeks unattended (≈ end Phase 2)
- **M2 Daily routine** — 4+ weeks of daily cloud-view use (during/after Phase 3)
- **M3 Friendly free users** — ≥2 external tenants active weekly, recruited via build-in-public posts (X/Medium/Reddit/Dev.to with hosted pages as living demos); passing M3 opens the billing initiative
- **M4 Paid** — >$100 MRR sustained one month
- **M5 ProductHunt launch** — only after M4; gate review on support/abuse/COGS first
- **M6 Profitable revenue line** — solo/small-team revenue exceeds full costs (COGS+ops+support) for a quarter; the branch-out gate: no PROACTIVE widening (outbound enterprise, new personas) before M6 — growing with an existing customer who grows is not branch-out; okay to be outgrown (Basecamp posture)

## Phase 1 — Tenancy Foundation and Sync-Up

**Theme:** Tenant entity and multi-tenant membership, sync contract, Dolt-native sync, hosted Dolt operations baseline — no auth, no serving yet

- [ ] `RMI-VISIONSTUDIO-501` Local tenant-assignment config (client-side only)
  - RE-SCOPED per open-core split (TRD T5a): full Tenant entity + systemforge identity/membership is CLOUD-side work, moves to private visionstudio-cloud (created at RMI-503). Public repo owns only the CLIENT record: local config mapping a registered repo to a tenant slug (cliconfig), no dependency on the cloud existing yet. Feeds RMI-502's sync command.
- [ ] `RMI-VISIONSTUDIO-205` Local-remote sync strategy spike + godolt module API
  - Dolt push/pull to tenant-scoped remotes; conflict story for single-writer-per-project; schema-version handshake; whole-DB push vs synced-entity projection DB; spike ALSO discovers the API shape for the standalone Dolt Go module (working name godolt, gogit precedent; SERVER-MODE-FIRST per operational lesson — embedded driver ruled out at 1 stable connection vs many concurrent agent sessions; SQL-first via dolt_* stored procedures over MySQL wire, CLI-exec secondary — TRD T2); 502/503 build against the module
- [ ] `RMI-VISIONSTUDIO-502` Sync contract and sync CLI
  - Explicit synced-entity list (TRD T2: initiatives incl lifecycle+visibility, phases, RMIs, releases, orgs, persons, repo metadata WITHOUT local_path); never-sync list (local paths, raw logs, raw ACTS evidence, source); visionstudio cloud login + sync --tenant with per-project tenant assignment
- [ ] `RMI-VISIONSTUDIO-503` Hosted Dolt operations baseline
  - Deploy topology, TESTED backup/restore, upgrade procedure, per-tenant size monitoring; proven with dogfood tenants before any site depends on the service
- [ ] `RMI-VISIONSTUDIO-535` Cloud-to-local pull + fast-forward-only sync policy
  - Added by MVP re-scope 2026-08-14 (sync + multi-user is the primary launch value; multi-user breaks the single-writer assumption in RMI-205's conflict story). `visionstudio pull --tenant <slug>`: Dolt-native fetch + LOCAL merge — conflicts surface locally via Dolt conflict tables, the cloud never merges divergent user work server-side. Push becomes fast-forward-only: rejected with a "pull first" error when the tenant remote has commits the client lacks. This is the entire multi-user concurrency model (the git model). Client side of the facade protocol in `RMI-VISIONSTUDIOCLOUD-002`; dogfood M1 proof: push/pull round-trip between two of our own machines before any auth exists

## Phase 2 — Public Hosting MVP: Roadmap and Release Log

**Theme:** Read-only public JSON serving with the two-filter rail; productbuildershq.com and plexusone.dev re-point their widgets — two sites, one service

- [x] `RMI-VISIONSTUDIO-504` ~~Public serving service~~ SUPERSEDED
  - Cancelled: implementation moved to `RMI-VISIONSTUDIOCLOUD-001` in the new private visionstudio-cloud repo (created at RMI-503), per the open-core repo split (TRD T5a) decided after this RMI was written. See that repo's roadmap for the live item; original scope preserved: GET /t/<tenant>/roadmap.json + /t/<tenant>/releases.json in releaselog IR shape, two-filter rail, CDN caching.
- [ ] `RMI-VISIONSTUDIO-505` productbuildershq.com powered by cloud
  - /roadmap and /releases render from cloud endpoints; static export (VS-006 RMI-309, PBHQSITE static JSON) demoted to documented degraded-mode fallback with identical shapes
- [ ] `RMI-VISIONSTUDIO-506` plexusone.dev powered by PBHQ cloud
  - plexusone tenant synced; plexusone.dev release widget re-points from committed static JSON to cloud URL (visitors see no change); roadmap page option documented; first external-consumer proof

## Phase 3 — Authenticated Multi-Tenant Launch (MVP)

**Theme:** Login (GitHub + Google SSO), multi-user orgs with invitations, authenticated bidirectional sync through the HTTPS facade — the minimal launchable multi-tenant service. Re-scoped 2026-08-14: sync + multi-user is the primary launch value; the hosted web surface (previous Phase 3 scope) moves to Phase 5 and no longer blocks launch on INIT-VISIONSTUDIO-001. Launch loop: login → create org → invite member → push/pull shared portfolio → public roadmap/releases pages served from the cloud copy (Phase 2)

- [x] `RMI-VISIONSTUDIO-202` ~~AuthN/AuthZ for hosted daemon via systemforge~~ SUPERSEDED
  - Moved to `RMI-VISIONSTUDIOCLOUD-004` in the private visionstudio-cloud repo per the open-core split (TRD T5a re-homing, RMI-504 precedent); original scope preserved there, Google SSO added. Re-narrowed 2026-08-16: the GitHub/Google federation mechanics themselves moved again, to `RMI-SYSTEMFORGE-001` in `grokify/systemforge` (built once, centrally, in `cmd/coreauth`, reusable by every relying-party site — not just visionstudio-cloud); RMI-VISIONSTUDIOCLOUD-004's remaining scope is registering as an OAuth client of coreauth and mounting the existing `@systemforge/auth` frontend package
- [x] `RMI-VISIONSTUDIO-203` ~~Multi-tenant org scoping enforcement~~ SUPERSEDED
  - Folded into `RMI-VISIONSTUDIOCLOUD-002` (the sync facade authorizes every push/pull against tenant membership) and `RMI-VISIONSTUDIOCLOUD-003` (provisioning creates the per-tenant grants) — enforcement moved to where the service layer lives
- Cloud-side MVP RMIs live in visionstudio-cloud's `docs/specs/ROADMAP.md`: `RMI-VISIONSTUDIOCLOUD-002` (HTTPS sync facade, branch-gated apply), `RMI-VISIONSTUDIOCLOUD-003` (org/tenant provisioning + invitations), `RMI-VISIONSTUDIOCLOUD-004` (coreauth OAuth client registration + `@systemforge/auth` mount)
- `RMI-SYSTEMFORGE-001` (GitHub/Google federation handler for coreauth) lives in `grokify/systemforge`, registered in the registry 2026-08-16 as a cross-org shared dependency (used across ProductBuildersHQ, plexusone, and grokify) — tracked in this phase because visionstudio-cloud's login RMI depends on it, but it is not a member RMI of either visionstudio or visionstudio-cloud

## Phase 4 — Cloud ACTS and Freemium Groundwork

**Theme:** Free monthly ACTS report generated cloud-side; per-tenant metering as the COGS ledger; tier definitions on paper, no billing

- [ ] `RMI-VISIONSTUDIO-507` Free monthly ACTS report in the cloud
  - Generated from synced summaries + locally-redacted excerpts (ACTS capture/redaction stays local per INIT-ACTS-001 privacy model); monthly cadence free by design, weekly/on-demand reserved for future paid tiers; GATED on ACTS Mode A satisfaction (TRD T11: subscription-session analysis proven locally before any API-metered automation); aggregate-research inclusion + opt-out per ACTS TRD T12
- [ ] `RMI-VISIONSTUDIO-508` Per-tenant metering
  - LLM token cost per cloud analysis recorded per tenant (ACTS self-ROI machinery reused); usage counters for projects/reports; NO billing integration — metering only, pricing data for the future billing initiative
- [ ] `RMI-VISIONSTUDIO-509` GitHub CI monitoring for the free tier
  - Added by LP aggregate finding lp-001 (unfunded promise: press + FAQ list CI monitoring in the free tier, which must exist by M3); GitHub Actions run status per registered repo surfaced per tenant; read-only, gh/API-based; feeds the quality signal (CI failures) as a side effect

## Phase 5 — Hosted App Surface

**Theme:** Hosted web shell and IDE backend-picker — moved out of the MVP path by the 2026-08-14 re-scope; blocked on INIT-VISIONSTUDIO-001 Phases 2-3 for panel content, which no longer blocks launch

- [ ] `RMI-VISIONSTUDIO-204` Hosted web shell via systemforge-web
  - @systemforge/{shell,auth,tenant,pages} provide chrome/login/tenant-switcher; VS-001's React panels render inside — panel content still BLOCKED on INIT-VISIONSTUDIO-001 Phases 2-3, shell itself is not
- [ ] `RMI-VISIONSTUDIO-201` IDE backend-picker: local Dolt vs remote URL
  - Base-URL + auth-token switch; HTTP-only frontend contract from VS-001 guarantees panel parity
