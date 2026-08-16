# FAQ — VisionStudio Cloud

## Customer FAQ

**Q: Does my code or my coding-agent transcripts go to the cloud?**
No. The sync contract is explicit: source code, raw session logs, raw
ACTS evidence, and machine-local paths never sync. The cloud receives the
project record — initiatives, roadmap items, releases, repo metadata,
visibility flags — plus ACTS *summaries* that pass a local redaction gate.

**Q: What's free?**
2–5 projects (deletable), the hosted release log and roadmap board, a
monthly ACTS report, the token-vs-conventional-commit report, and GitHub
CI monitoring. Paid tiers add project count, weekly/on-demand ACTS
analysis, forecasting, longer retention, adoption analytics (Stripe/MAU),
and team seats.

**Q: Who is this for?**
The 1-person business built with AI — solo product builders and solo
founders who own everything: product, code, releases, roadmap, and
economics. If you're building alone with coding agents and want your
work tracked, measured, and selectively published without giving up
local-first privacy, you're the customer.

**Q: What about teams?**
Small AI-native startup teams are supported — think GitHub with its
three early founders. If everyone on the team builds (same persona as
the solo builder), you're a tenant with a few members; the multi-tenant,
multi-user model handles it with no special machinery.

**Q: And enterprises?**
Carefully, and later — but not locked out. Enterprises have different
roles, personas, and requirements — PM vs. Eng structures, org controls,
integrations — that the AI-native builder model doesn't have. We won't
*pursue* enterprise or build persona infrastructure until the
solo/small-team segment is a profitable revenue line; but the fence is
persona focus, not customer size — if a customer grows and still wants
us, we grow with them. And per the 37signals/Basecamp precedent, we're
comfortable being outgrown: customers may graduate past us while the
product stays focused and profitable. The line we hold: teams where
everyone is a Product Builder are in; chasing organizations that need
persona differentiation waits until the core business earns it.

**Q: Can I belong to more than one workspace?**
Yes — that's the tenant model. One local install can sync different
projects to different tenants: your company, your personal/open-source
identity. Each tenant is an isolated database.

**Q: What happens to my public pages if the cloud is down?**
The static-export fallback serves the same JSON shapes; sites can pin to
their last static snapshot. Public pages are CDN-cached in front of the
service anyway.

**Q: Can I leave?**
Yes. Your tenant is a Dolt database — pull it, and you have the full
versioned history locally. Deleting a tenant deletes its database.

**Q: Why is a private repo missing from my public roadmap?**
By design. Public payloads require repo visibility *and* initiative
visibility to be public. Naming a private repo publicly is a deliberate
override, never a default.

**Q: Is my data used for research?**
Aggregated, anonymized statistics only — no per-user rows, no session
content, with a k-anonymity floor — feeding the ProductBuildersHQ Annual
Research Report on Agentic Coding. Free-tier ACTS analysis includes this
aggregate participation (it's part of what funds the free analysis);
disabling ACTS LLM analysis is the full opt-out.

## Internal FAQ

**Q: Aren't we building too much ourselves?**
We already built it — separately, for separate needs, before any platform
intent existed: maturity models (prism-capability/prism-maturity), token
spend and DevX reports (omnidevx/devfolio), Working Backwards specs
(visionspec), the release log (releaselog, v0.1.0 2026-02-02, serving
plexusone.dev in production since). DoltDB multi-repo tracking was the
catalyst that let proven components consolidate into one app. This
initiative is not greenfield platform-building; it is hosting a
consolidation that already works. The only genuinely new bet is that
hosting enables CI/CD and adoption/revenue build-out — and that bet is
gated, not assumed.

**Q: Why build the cloud now if local works so well?**
Because we need it ourselves: productbuildershq.com should host the
roadmap and release log, and plexusone.dev's release page can be powered
by the same service. Two real consumers on day one, zero third-party
risk.

**Q: What does this cost us to run?**
Infrastructure is $0 incremental: an already-paid ~$600/month server sits
idle (its former sites are non-operational) and comfortably hosts Dolt
plus the serving service — the cloud *recoups* an existing expense. A
~$10/month Lightsail instance is the isolation alternative. The domain is
already owned. The only tenant-scaling cost is LLM tokens for cloud ACTS
reports, metered per tenant from day one.

**Q: How do we get customer evidence before ProductHunt?**
Build in public from existing channels — X (x.com/Grokify), Medium,
Reddit, Dev.to — with the hosted roadmap/release pages as living demos
and cleansed reports as content. M3 friendly users are recruited from
respondents. If nobody bites, that's cheap, early gate data.

**Q: Why no billing in this initiative?**
Milestones gate investment (see NARRATIVE-6P). Billing is a separate
initiative that only opens after friendly free users are active (M3) —
by then metering has produced real per-tenant cost data to price against.

**Q: Isn't free ACTS analysis expensive?**
It has real LLM cost per user, which is why the free tier is *monthly*
cadence on an efficient configuration, and why per-tenant metering ships
with the first cloud report. Cadence and depth are the paid axis, not the
existence of analysis.

**Q: What do we build the cloud app on?**
systemforge + systemforge-web — our own SaaS platform module (identity
with multi-tenant organizations and plan tiers, OAuth2 server with GitHub
login, sessions, React shell/tenant-context packages), built precisely to
derisk and accelerate buildouts like this. Honest status:
construction-proven across multiple site buildouts, production-unproven —
VisionStudio Cloud is its first production deployment, which is both the
acceleration and a named risk.

**Q: What's the riskiest assumption?**
Hosted Dolt operations — the least-traveled part of the stack. That's why
backup/restore is a Phase 1 RMI with a tested procedure, and why tenants
are only our own orgs until M2 is passed.

**Q: One-way or two-way doors?**
One-way: committing public URLs that other sites depend on; holding
tenant data at all (retention/deletion obligations). Two-way: JSON
shapes (versioned), tier boundaries (not in code yet), auth provider
(GitHub OAuth first, swappable behind Person mapping).

**Q: Why paid users before ProductHunt?**
Publicity converts best against a proven willingness to pay. >$100 MRR is
a small bar, but it converts the launch from "look at this" to "people
pay for this."
