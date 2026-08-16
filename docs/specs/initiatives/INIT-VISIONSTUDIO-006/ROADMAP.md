# Roadmap Board — Releases, Shipped Marks, and Public Roadmap View — Roadmap

**Initiative:** `INIT-VISIONSTUDIO-006`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-<REPOSLUG>-<NNN>`. Phase status is derived from member RMIs — a phase is complete only when all its required RMIs are complete.

This initiative uses the RMI-VISIONSTUDIO-3xx block.

## Phase 1 — Release Model and Ingest

**Theme:** Release entity with per-repo tags and initiative associations; trailer-chain auto-association proven on real git history via backfill

- [ ] `RMI-VISIONSTUDIO-301` Release Ent entity and edges
  - Release (repo@tag ID, tag, released_at, url, notes_ref) with repository edge + M2M initiatives + M2M roadmap_items; migration; follows //go:build dolt pattern
- [ ] `RMI-VISIONSTUDIO-302` Release CLI verbs
  - release create/list/show/attach/detach; attach accepts initiative or RMI IDs; release list --repo and --initiative filters
- [ ] `RMI-VISIONSTUDIO-303` CHANGELOG.json and tag ingest with trailer-chain association
  - PRIMARY source CHANGELOG.json (structured-changelog: releases[].version/date/highlights -> tag/released_at/notes_ref); git tags as fallback; git log prevTag..tag Refs trailers -> RMIs -> initiatives; semver-like filter with override; gh enrichment; extends existing ingest family; release-record verb or release-skill ingest step makes the changelog update the natural recording moment
- [ ] `RMI-VISIONSTUDIO-304` Historical backfill with coverage report
  - Walk all historical tag ranges; report trailer-coverage fraction per repo; associations confidence-labeled; no time-proximity guessing
- [ ] `RMI-VISIONSTUDIO-313` Changelog release gate — manufacture the habit
  - CHANGELOG.json (structured-changelog) is OUR convention, not a market standard; instead of presuming it, visionstudio release record / release transition GATES on it and OFFERS to scaffold: generate CHANGELOG.json entries from conventional commits via schangelog (with AI classification fallback for plain commits); habit becomes tooling output, not entry fee — key mitigation for the market-discipline generalization risk (VS-002 6-pager §5)
- [ ] `RMI-VISIONSTUDIO-314` GitHub Releases API as a release ingest source
  - Closes a verified completeness gap: tag-based ingest undercounts against real GitHub Release history (found live — omniskill: 3 releases from tags vs 11 GitHub-verified releases; 8 missing). Add GitHub Releases API (gh api or the releaselog fetch layer/its JSON export) as a source alongside CHANGELOG.json and tags; release body/notes text captured for RMI-315 to use as match evidence; precedence and conflict handling with existing sources documented
- [ ] `RMI-VISIONSTUDIO-315` AI-assisted historical backfill matching
  - For releases with no trailer-derived initiative/RMI match (expected and normal for the ~10 years of pre-adoption history across PBHQ/plexusone/grokify — see CLAUDE.md "AI-assisted historical backfill matching"): an interactive review tool (`visionstudio release backfill-match`, spec-judge show/record pattern — Mode A, subscription-session, zero marginal API cost per ACTS TRD T11) surfaces candidate initiative matches per unmatched release using changelog/commit/release-note text as evidence, each suggestion labeled Analyst inference with a confidence score and the evidence snippet it was drawn from; NO auto-attach — every match requires explicit human confirmation (`release attach`), mirroring the no-time-proximity-guessing rule live ingest already enforces

## Phase 2 — Board and Forcing Function

**Theme:** Lifecycle-derived board columns in the web UI; unshipped queue, validate rule, and dashboard badge make mark-it-released the path of least resistance

- [ ] `RMI-VISIONSTUDIO-305` Roadmap board panel
  - Columns Under Consideration/Planned/In Progress/Delivered/Shipped derived from lifecycle status; release chips (repo@tag); filters program/repo/period; cancelled hidden by default
- [ ] `RMI-VISIONSTUDIO-306` Unshipped queue and validate rule
  - delivery_complete/releasing with zero releases, sorted by staleness; visionstudio validate warns past threshold (default 14d); dashboard badge count
- [ ] `RMI-VISIONSTUDIO-307` Initiative visibility flag
  - visibility field internal (default) / public on Initiative; CLI + UI toggle; flag-flip is the publication approval
- [ ] `RMI-VISIONSTUDIO-308` Portfolio cleanup pass
  - Run the unshipped queue to zero once: attach releases + transition to released, or consciously park/cancel; activates the ACTS quality signal across the back catalog
- [ ] `RMI-VISIONSTUDIO-312` Internal releases panel
  - Releases view in the web UI over the Release entity (TRD T4a): filter by repo/period/initiative, drill to shipped initiatives+RMIs and changelog entry; covers private repos the external /releases/ page never shows; releaselog fetch layer with repo-scoped token as private-repo enrichment source (token+output internal-only, feeds Dolt)

## Phase 3 — Public Export and ACTS Handshake

**Theme:** Slack-style public roadmap page generated statically; per-repo released_at documented as the refined ACTS acceptance mark

- [ ] `RMI-VISIONSTUDIO-309` Static roadmap export
  - JSON + dependency-free HTML board of public-flagged initiatives with release chips; generate-don't-host
- [ ] `RMI-VISIONSTUDIO-310` Site publish integration
  - Export wired into the productbuildershq.com publish flow; flagship initiatives flagged public as the first content
- [ ] `RMI-VISIONSTUDIO-311` ACTS per-repo acceptance marks
  - Document + expose Release.released_at as the per-repo acceptance mark for ACTS escaped-defect tiering (INIT-ACTS-001 TRD T10); store read path agreed with ACTS Phase 3
