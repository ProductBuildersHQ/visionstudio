# TRD — INIT-VISIONSTUDIO-007: Org, Person, and Visibility Modeling in the Registry

**Initiative:** INIT-VISIONSTUDIO-007
**Status:** proposed

## T1 — Organization entity

```text
Organization
├── id          ("github.com/plexusone")
├── login       ("plexusone")
├── kind        ("organization" | "user")   // grokify is a user account
├── display_name, website, release_page_url (optional)
├── created_at / updated_at
└── edges: repositories (O2M), members (M2M → Person, through affiliation)
```

- `Repository.organization` (string) stays for back-compat; a new
  `organization` edge becomes the queryable relation. Backfill creates the
  three org rows from existing strings; new `registry add` resolves or
  creates the org row.

## T2 — Person entity

```text
Person
├── id            ("person:grokify")
├── github_login  ("grokify")
├── display_name, email_identities []string   // commit-author matching
├── created_at / updated_at
└── edges: organizations (M2M with role: owner|member|contributor)
```

- Align field naming with `prism-core`'s Person primitive where it exists
  (import or mirror — decide at implementation; do not fork semantics).
- First data: grokify/John as `owner` of grokify, plexusone,
  ProductBuildersHQ. Email identities cover commit-author variants so
  DevFolio/ACTS can attribute commits to the person.
- External issue reporters are *not* modeled as Person rows; internal/
  external derivation (ACTS defect tiers) needs only "is this login an
  org member," answerable from the affiliation edge.

## T3 — Repository visibility

- `Repository.visibility` (`public` | `private` | `unknown` default).
- **Ingested, not hand-maintained:** from GitHub via `gh repo view --json
  visibility` or the releaselog fetch layer during ingest runs; refreshed
  whenever release ingest touches the repo. `unknown` renders as a warning
  in the UI, never as public.

## T4 — Safety rails

External projections (VS-006 roadmap export, anything feeding public
pages) apply **two independent filters**:

1. Repo-level: `visibility == "public"` — `unknown` and `private` never
   export.
2. Initiative-level: `visibility == "public"` flag (VS-006 T6).

A public-flagged initiative whose repos are all private exports nothing by
default; naming a private repo publicly must be a deliberate override, not
a side effect.

## T5 — Slices and consumers

- **Per-org rollups:** releases, initiatives, RMIs, spend by Organization
  edge (replaces string matching); enables per-org release pages
  (PBHQSITE stamp-out) to enumerate coverage from the DB.
- **Private-repo focus list:** registered private repos with their active
  initiatives — the "some private ones we're looking to focus on" view.
- **ACTS practitioner lens:** "all repos across the person's orgs"
  resolved via Person → Organizations → Repositories; the
  practitioner-period rollup (ACTS TRD T3) consumes this instead of
  assuming "everything local."

## Risks

1. **Identity sprawl** — keep Person minimal (one row now); resist
   modeling every contributor until a consumer needs it.
2. **Visibility drift** — a repo flipped private/public on GitHub between
   ingests; mitigated by refresh-on-ingest and `unknown`-is-not-public.
3. **prism-core alignment** — if Person diverges from prism-core's
   primitive, ACTS/PRISM interop pays later; settle import-vs-mirror
   before coding.
