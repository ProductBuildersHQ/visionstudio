# PRD — INIT-VISIONSTUDIO-007: Org, Person, and Visibility Modeling in the Registry

**Initiative:** INIT-VISIONSTUDIO-007
**Status:** proposed
**Type:** feature
**Workflow:** pbhq-lite

## Problem

The registry models repos well but three things it correlates over are
implicit:

1. **Organizations are strings.** `Repository.organization` is a bare
   string, so "everything under plexusone," per-org release logs, and
   per-org efficiency slices are string-match queries with no place to
   hang org-level facts (kind, website, default visibility, release-page
   URL).
2. **People don't exist.** The user `grokify` is the primary contributor
   to all three orgs (grokify, plexusone, ProductBuildersHQ) — but no
   entity records that. ACTS's portfolio-first measurement
   (practitioner-period across *all* projects) assumes exactly this
   identity and currently has nowhere to anchor it.
3. **Visibility is unknown.** Repos are public or private on GitHub, but
   the registry doesn't record which. Most active repos are public; some
   private ones are a growing focus. External projections (releaselog
   page, roadmap export) currently rely only on initiative-level flags —
   a repo-level guard is missing.

## Goals

1. **G1 — Organization entity.** github.com orgs and user-accounts-as-orgs
   (grokify) as first-class rows: kind (`organization`|`user`), website,
   release-page URL; Repository gains an edge; existing string values
   backfilled.
2. **G2 — Person entity.** People with GitHub logins and commit-author
   identities; org affiliations (owner/member/contributor). First row:
   grokify/John as owner-and-primary-contributor of all three orgs. This
   is the anchor for the ACTS practitioner-period lens.
3. **G3 — Repository visibility.** `public`|`private` on Repository,
   ingested from GitHub (not hand-maintained), refreshed with ingest runs.
4. **G4 — Safety rails and slices.** External projections filter on repo
   visibility *in addition to* initiative flags (defense in depth: a
   public-flagged initiative in a private repo exports its existence only
   deliberately). New slices: per-org rollups, private-repo focus list.

## Non-Goals

- Multi-tenant auth/permissions — this models facts, not access control
  (that's the hosted-mode initiative's concern).
- Full contributor analytics (commit counts per person) — DevFolio/OmniDevX
  own activity measurement; this initiative only models identity.
- Tracking external contributors beyond what defect-tier derivation needs
  (issue author internal/external — ACTS TRD T10 — benefits from Person
  but doesn't require modeling every reporter).

## Success Criteria

- Three Organization rows and one Person row exist with correct edges;
  registry queries can roll up by org without string matching.
- Every registered repo has accurate GitHub-derived visibility.
- The external export path provably excludes private repos regardless of
  initiative flags.
- ACTS can resolve "the practitioner's repos across all orgs" from the DB
  (consumed by its portfolio-period rollups).
- A private-repo focus list renders in the UI.
