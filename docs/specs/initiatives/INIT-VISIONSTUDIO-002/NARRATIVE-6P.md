# Six-Pager — VisionStudio Cloud: Multi-Tenant Hosted Backend

**Initiative:** INIT-VISIONSTUDIO-002
**Decision sought:** proceed with Phase 1–2 build (tenancy, sync, public
serving) under the milestone-gated investment plan below.

## 1. Context

**We did not set out to build a platform.** Each capability below was
built independently, for its own need, and proved itself standalone
before any consolidation was planned:

| Capability | Built as | Proven by |
|------------|----------|-----------|
| Capability stacks & maturity models | prism-capability, prism-maturity | Models and assessments in active use |
| Token spend & DevX reporting | omnidevx, devfolio | Changed our own behavior: model switching and /compact discipline driven by our own devx reports |
| Working Backwards specs | visionspec | This document set is authored and evaluated through it |
| Release log | releaselog (v0.1.0 shipped 2026-02-02) | plexusone.dev/releases in production and actively used since; the format itself is proven across developer-program portals with many SDKs |
| Public roadmap board | Initiative lifecycle (visionstudio) | Slack's public Trello roadmap proved the format at scale |
| SaaS app platform (identity, OAuth2, sessions, multi-tenant orgs, React shell) | systemforge (~215 commits) + systemforge-web (~34 commits) | Built to derisk/accelerate new web apps; exercised across multiple site buildouts — **production-unproven** (see gradient below) |

The catalyst was **DoltDB for multi-repo spec and progress tracking** —
it allowed these separately-proven components to consolidate into one
app. That consolidation is **Step 1: VisionStudio Local**, and it has
proven itself as a daily-driver: initiatives, phases, RMIs, specs,
maturity, and token spend in one local Dolt-backed system, wrapped in a
standardized operating model (lifecycle-transitions-as-data, per-repo
releases recorded at changelog time, escaped-defect quality signals,
ACTS efficiency measurement) whose value compounds with history.

The honest epistemic gradient, and what this initiative actually bets on:

- **Proven:** each component individually; consolidation locally (daily
  use); consolidation's UI advantage over standalone static HTML.
- **Step 2, precedent-backed:** hosting the release log and roadmap —
  both formats proven elsewhere (dev-portal release logs, Slack's public
  board), ours proven locally; only the *serving* is new.
- **Construction-proven, production-unproven:** systemforge — the app
  platform (identity, OAuth2 server, sessions, multi-tenant
  organizations, RBAC, React shell) built precisely to derisk and
  accelerate buildouts like this one. VisionStudio Cloud is its first
  production deployment: platform and app prove out together (risk
  in §5).
- **Theorized, gated:** that hosting enables the build-out to CI/CD
  monitoring and adoption/revenue metrics. This is the part the M-ladder
  exists to test cheaply.

Three pressures point at a hosted layer: (a) our public assets — the
roadmap board and release log — need to be *served*, and we want
productbuildershq.com to host them; (b) plexusone.dev already maintains a
release page from static JSON that the same service could power, making
PBHQ's cloud the backend for a second site immediately; (c) the
aggregation features (ACTS interpretation, forecasting, release
management, CI monitoring, adoption analytics) belong where history
accumulates.

Moving to the cloud is a big move: operational commitment, public
dependencies, eventually other people's data. That is why this initiative
runs Working Backwards with explicit adoption gates rather than as a
feature build.

## 2. Customer and need

**Primary (now): Customer Zero** — ourselves: plexusone and the other
grokify properties. A solo builder running many initiatives across three
GitHub orgs, needing hosted public views and cross-machine, cross-tenant
continuity without giving up the fast, private local loop. Customer Zero
is a discipline, not just a label: every feature ships to us first, and
nothing reaches friendly users that Customer Zero doesn't run daily.

**Next (gated):** solo product builders and solo founders using coding
agents. The mission is sharper than "help people build with AI": it is
**the 1-person business built with AI, where everything is owned by one
person** — product, code, releases, roadmap, economics. Initial channels
are where those people already gather: the Lovable subreddit, r/SaaS,
and the MicroConf/TinySeed communities.

**Core theory (and its risk):** our development model is proven for
*ourselves* — plexusone.dev is a working solo-product-builder operation.
The theory is that this model generalizes to other solo builders, and we
de-risk it by staying as close as possible to the model we know is
proven: target people whose situation matches ours before anyone
else's. The explicit risk is that it generalizes less than we think
(see §5).

**Also in scope (multi-tenant, multi-user build-out):** small AI-native
startup teams — think GitHub itself with its three early founders. A
few founders who *all build* are the same persona as the solo builder,
just sharing a tenant: the Person↔Tenant membership model supports them
with zero new machinery. The scope line is **persona uniformity**: teams
where everyone is a product builder are in; organizations that need
persona differentiation are not.

**Later (approached carefully, not locked out):** enterprises. Different
roles, personas, and requirements — PM vs. Eng structure, org controls,
integration surface — that the AI-native builder model doesn't need.
The pattern we expect to matter there is the "1-person product builder
inside a large enterprise" (the kind we study in our Block/Jack Dorsey
analysis on the ProductBuildersHQ site), but serving it means platform
infrastructure we won't build until the solo/small-team segment is a
**profitable revenue line** (gate M6).

Two postures govern any growth beyond the core segment:

- **Grow with our customers.** The fence is *persona focus* (Product
  Builders), not customer size. If a 3-founder tenant grows into a
  50-person company and still wants us, we follow that demand — organic,
  customer-led expansion is welcome; *outbound* enterprise pursuit and
  persona dilution are what's fenced.
- **Willing to be outgrown.** Per the 37signals/Basecamp precedent: it is
  acceptable for customers to outgrow the product while it remains
  profitable. We optimize for a durable, profitable core serving Product
  Builders — not for retaining every customer at any complexity cost.

The underserved need: every existing tool tracks either project management
(Linear, GitHub Projects) or spend (usage dashboards) or releases
(changelogs). None connects *delivered releases ↔ roadmap ↔ token economics
↔ working practice* — and none is local-first with a cloud that receives
only the project record.

## 3. Approach

- **Tenancy:** every tenant has a `tenant_id`; a person belongs to
  multiple tenants (company, personal/OSS). Database-per-tenant on Dolt —
  natural isolation, versioned history, deletion = database removal.
- **Sync:** Dolt-native push/pull to tenant remotes under an explicit
  contract — never source, raw logs, local paths, or unredacted ACTS
  evidence. Local stays the fast private loop; the cloud is sync target
  and serving layer.
- **First workloads:** public read-only JSON — `/t/<tenant>/roadmap.json`
  and `/t/<tenant>/releases.json` (releaselog IR shape, so plexusone.dev
  re-points its existing widget by URL). No auth needed; not blocked on
  the unfinished local web foundation (INIT-VISIONSTUDIO-001).
- **Then:** GitHub OAuth + hosted shell (blocked on VS-001 Phases 2–3),
  free monthly cloud ACTS report, per-tenant metering. Billing is a
  future initiative, gated below.

## 4. Goals — milestone-gated investment

Each milestone is a gate: the next stage's investment does not start
until the gate passes. Regressions close the gate again.

**M1 — Launched, basic functions working.**
Hosted roadmap + release log live for the PBHQ tenant;
plexusone.dev/releases served from the cloud with no visitor-visible
change; tested backup/restore. *Gate: both sites serving from cloud for
2+ weeks without manual intervention.* (≈ end of Phase 2.)

**M2 — Part of our daily routine.**
The cloud views are used the way VisionStudio Local is used today —
daily: unshipped queue worked in cloud, releases recorded at changelog
time land visibly, weekly review happens against hosted views. *Gate:
4+ consecutive weeks of daily-routine use; we'd notice within a day if it
broke.* (During/after Phase 3.)

**M3 — Friendly free users.**
Auth on; 3–10 invited builders on the free tier (2–5 projects, monthly
ACTS report). *Gate: ≥2 external tenants active in a normal week without
hand-holding; onboarding is one command; metering shows per-tenant COGS.*
**Passing M3 opens the billing initiative** — priced against real
metering data.

**M4 — Paid users.**
A small number of paying users. *Gate: >$100 MRR sustained for a month,
with churn understood.* Deliberately small: the bar is proof of
willingness to pay, not revenue.

**M5 — Public launch (ProductHunt).**
Only after M4 — publicity lands on a proven conversion story. *Gate
review before launch: support load, abuse controls, deletion/retention
obligations, free-tier COGS at 10× signups.*

**M6 — Profitable revenue line (the branch-out gate).**
The solo/small-team segment becomes a real business: *sustained revenue
exceeding its full costs — metered COGS, operations, and support time —
for a quarter.* **No proactive branch-out (outbound enterprise pursuit,
new personas, new segments) before M6.** Depth beats breadth: until the
core segment is profitable, all investment deepens it rather than
widening the market. Growing *with* an existing customer who grows is
not branch-out (see §2 postures); building new persona infrastructure
to chase customers we don't have is.

### Investment sizing

- **Infrastructure: $0 incremental.** We already pay ~$600/month for a
  server that is currently idle (it previously hosted sites that are now
  non-operational). Repurposing it converts an unrecouped expense into
  product infrastructure; it comfortably hosts the Dolt server and the
  serving service. Alternative if we prefer isolation: a ~$10/month AWS
  Lightsail instance suffices for the MVP (and matches existing Lightsail
  deployment experience — PROG-OMNIAGENT-LIGHTSAIL, omnideploy). Either
  way, hosting is not a cost gate.
- **Domain:** already owned (the marketing site's domain is paid).
- **Variable COGS: subscription-first, then metered.** Our own ACTS
  analysis initially runs inside the existing Claude Code subscription
  (agent-session mode, ACTS TRD T11 — the `spec judge` pattern) at zero
  marginal API cost. API-metered cloud analysis is enabled only after the
  local workflow is proven under subscription or controlled costs — so
  Phase 4's cloud report carries no unbounded spend at launch.
  Hypothesis once metered: single-digit dollars per tenant-month at the
  free tier's monthly cadence; replaced with real numbers by per-tenant
  metering (RMI-VISIONSTUDIO-508). Public-page serving is CDN-cached
  static-shaped JSON — negligible.
- **COGS partially offsets as marketing:** aggregated, anonymized ACTS
  statistics feed the planned *ProductBuildersHQ Annual Research Report
  on Agentic Coding* (ACTS TRD T12), reclassifying part of the analysis
  spend as research/marketing. Free-tier analysis includes aggregate
  inclusion (disclosed; disabling ACTS LLM is the full opt-out).
- **Build effort (estimated, systemforge-accelerated):** the cloud is
  built on **systemforge** (Go platform: identity with multi-tenant
  Organizations/Memberships/plan field, full OAuth2 server with GitHub
  login, JWT/BFF sessions, RBAC) and **systemforge-web** (React: auth,
  tenant context, shell, pre-built pages) — built for exactly this
  purpose. Estimate: **Phases 1–2 ≈ 2–4 part-time weeks** of agentic
  development (9 RMIs: tenancy mapping onto systemforge identity, sync
  CLI, ops baseline, serving endpoints, two site re-points); **Phase 3 ≈
  a similar block**, since auth/sessions/shell are integration rather
  than construction. The build is itself ACTS evidence — token spend per
  delivered RMI is measured by the system being extended, so estimates
  are replaced by measurement from the first week.

### Customer-evidence plan (build in public)

We are the proof, published. Once M1–M2 hold, we post the working system
from existing audience channels — X (x.com/Grokify), Medium
(@grokify), Reddit (u/grokify), Dev.to — with the hosted roadmap and
release pages as living demos and cleansed ACTS/efficiency reports as
content, aimed at the communities where solo builders gather (Lovable
subreddit, r/SaaS, MicroConf/TinySeed). All of this happens **before
ProductHunt**. M3's friendly users are recruited from respondents; the
M3 gate itself is unchanged (≥2 external tenants active weekly). This
converts customer evidence from assumption to a dated pipeline:
build → use daily → publish → invite → charge → launch.

Beyond our own posts, ProductBuildersHQ's research output — case studies
and statement/article analysis of other builders (e.g., the Block/Dorsey
analysis) — positions us as a **thought leader in the product builder
movement**, not just an app vendor: actual experience and code
(PlexusOne) combined with apps (VisionStudio) and teaching
(ProductBuildersHQ). The research channel compounds the recruitment
channel: people who read the analysis are the audience for the app.

## 5. Risks and alternatives

- **Hosted Dolt operations** (highest): least-traveled stack component.
  Mitigation: ops RMI in Phase 1 with *tested* restore; dogfood-only
  tenants until M2; M-gates prevent premature exposure.
- **Platform and app prove together:** VisionStudio Cloud is
  systemforge's first production deployment — a defect in either looks
  like a defect in both. Mitigation: derisking buildouts is systemforge's
  design goal (multiple full site buildouts already exercised it);
  dogfood-only tenants until M2 mean production hardening happens on
  ourselves; the auth-heavy surface (Phase 3) lands only after M1's
  no-auth serving has proven the ops story. Its multi-property user
  model — consistent users and app health across systemforge apps, so
  properties can be acquired or divested cleanly — is a long-term asset,
  not a launch dependency.
- **Free-tier COGS:** ACTS analysis is the only tenant-scaling cost —
  infra is the repurposed server. Mitigation, layered: launch analysis
  runs in subscription agent-session mode at zero marginal API cost
  (Mode A, ACTS TRD T11); API-metered automation (Mode B) cannot enable
  before Mode A is trusted; once metered, monthly cadence is the free
  tier and cadence/depth the paid axis; part of the spend reclassifies as
  research/marketing via the annual report.
- **Customer-evidence risk:** demand beyond ourselves is unproven until
  the build-in-public posts land. Mitigation: the posting plan above uses
  existing audiences and living demo pages; if posts draw no interest by
  M3, that is itself gate data — we keep an excellent internal tool and
  stop external investment cheaply.
- **Model-generalization risk:** the core theory (§2) is that our proven
  solo model transfers to other solo builders — but plexusone.dev proves
  *our* practice, not the market's. Mitigation: minimize the transfer
  distance in two steps — lookalike solo builders first (Lovable/r/SaaS/
  MicroConf-TinySeed circles), then same-persona startup teams of ~3
  founders (a supported tenant shape, not a new model). The enterprise
  extension (persona differentiation, PM-vs-Eng structure) is
  separately-gated theory requiring platform build-out we have
  deliberately not started — the persona-uniformity line is the fence.
- **Privacy failure** (serving private data): two-filter rail enforced at
  sync *and* serve; `unknown` never public; private-repo fixture test
  gates every re-point.
- **Alternative considered — static-only forever:** cheapest, but no
  cross-site serving, no tenancy, no aggregation features; keeps two
  sites duplicating pipelines. Static remains as degraded-mode fallback,
  so the alternative is preserved rather than rejected.
- **Alternative considered — third-party backend (Supabase/Postgres):**
  loses Dolt's versioned history and the git-like sync we already run
  locally; rejected while Dolt-native sync remains viable (Phase 1 spike
  is the check).

## 6. Ask

Approve Phase 1–2 (tenancy, sync contract, ops baseline, public serving;
RMI-VISIONSTUDIO-501–506 + 205) under the gates above. No billing work,
no third-party onboarding, no ProductHunt planning until their gates
open; no API-metered analysis (Mode B) until subscription-mode analysis
(Mode A) is trusted. UXD and TPD are authored before the Phase 3 gate
(hosted shell — the first surface with real UX). After M1, this document
gains an appendix of real baselines (metering, page traffic,
post-response data) and is re-judged before M3 opens.
