# Org, Person, and Visibility Modeling in the Registry — Roadmap

**Initiative:** `INIT-VISIONSTUDIO-007`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-<REPOSLUG>-<NNN>`. Phase status is derived from member RMIs — a phase is complete only when all its required RMIs are complete.

This initiative uses the RMI-VISIONSTUDIO-4xx block (3xx belongs to INIT-VISIONSTUDIO-006).

## Phase 1 — Orgs, People, and Visibility

**Theme:** Organizations and people as first-class Dolt entities, GitHub-derived repo visibility, two-filter export safety rail, per-org and practitioner slices

- [ ] `RMI-VISIONSTUDIO-401` Organization and Person entities with backfill
  - Organization (login, kind organization|user, website, release_page_url) + Person (github_login, email_identities, org affiliations with role); backfill three orgs from registry strings + grokify Person row as owner of all three; Repository gains organization edge, legacy string kept; prism-core Person alignment settled first
- [ ] `RMI-VISIONSTUDIO-402` Repository visibility ingest
  - visibility public|private|unknown (default unknown, never treated as public); ingested via gh repo view or releaselog fetch layer, refreshed on ingest runs; UI warning for unknown
- [ ] `RMI-VISIONSTUDIO-403` Two-filter export safety rail
  - External projections require repo visibility public AND initiative flag public; private/unknown repos never export regardless of initiative flags; must land before or with RMI-VISIONSTUDIO-309/310
- [ ] `RMI-VISIONSTUDIO-404` Per-org, private-focus, and practitioner slices
  - Per-org rollups (releases/initiatives/RMIs/spend via org edge, replaces string matching); private-repo focus list with active initiatives; Person-to-orgs-to-repos resolution consumed by ACTS portfolio-period lens
