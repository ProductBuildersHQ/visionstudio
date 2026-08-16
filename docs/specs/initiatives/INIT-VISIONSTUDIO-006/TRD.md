# TRD — INIT-VISIONSTUDIO-006: Roadmap Board — Releases, Shipped Marks, and Public Roadmap View

**Initiative:** INIT-VISIONSTUDIO-006
**Status:** proposed

## T1 — Release entity (new Ent schema)

```text
Release
├── id            (string, "repo-slug@tag", e.g. "visionstudio@v0.3.0")
├── tag           (string, the git tag)
├── released_at   (time, tag/GitHub-release timestamp)
├── url           (optional, GitHub release URL)
├── notes_ref     (optional, changelog entry reference)
├── created_at / updated_at
└── edges:
    ├── repository  (unique, required → Repository)
    ├── initiatives (M2M → Initiative)
    └── roadmap_items (M2M → RoadmapItem, optional precision layer)
```

Follows existing schema conventions (`ent/schema/*.go`, `//go:build dolt`
pattern, string IDs with StorageKey). The initiative↔release M2M is the
core relationship: multi-repo initiatives accumulate one release per repo
per ship; a release lists every initiative whose work it carried.

## T2 — Trailer-chain auto-association

For each new tag on a registered repo (local path from the registry):

1. `git log prevTag..tag --format=%H %(trailers:key=Refs,valueonly)`
2. Collect `RMI-<SLUG>-<NNN>` trailers → resolve RMIs → resolve initiatives.
3. Create Release, attach RMI and initiative edges.
4. Association confidence: trailer-derived = high; commits without trailers
   contribute nothing (no time-proximity guessing — consistent with ACTS
   evidence standards).

Backfill = the same walk over all historical tag ranges. Trailer coverage
is recent, so old releases may associate sparsely; that is acceptable and
labeled.

## T3 — Ingest sources

- **Primary: `CHANGELOG.json`** (structured-changelog format,
  schangelog-validated) in each registered repo. Verified fields:
  `repository` at the top level and `releases[]` with `version`, `date`,
  and `highlights` — the Release entity's tag, `released_at`, and
  `notes_ref` come straight from it, with human-curated notes for free.
- **Secondary:** git tags in the registry's local paths (repos without
  CHANGELOG.json, or tags not yet in it).
- **GitHub-release enrichment:** URL/notes via the `grokify/releaselog`
  fetch layer (as a Go library or its JSON IR). With the public read-only
  token this doubles as the external-page fetch (INIT-PBHQSITE-001); with
  a repo-scoped token it also covers **private repos** — that token and
  its output stay internal, feeding Dolt only, never a published page.
  `gh api` remains the fallback.

**Dolt is the internal correlation store.** All releases across all repos
— public and private — land as Release rows, joined to initiatives, RMIs,
and (via `released_at`) the ACTS quality signal. The external artifacts
(releaselog page, roadmap export) are projections of public subsets; the
internal truth lives in Dolt and renders in the visionstudio UI (T4a).

## T4a — Internal releases panel

The internal counterpart of the public /releases/ page: a Releases view in
the visionstudio web UI over the Release entity — filterable by repo,
period, and initiative; each release drills down to the initiatives and
RMIs it shipped and links its changelog entry. Covers private repos that
the external page can never show. Reuses the existing toolkit/table
primitives.
- Trailer-chain association (T2) runs identically for either source — the
  source supplies the release row; git history supplies the associations.
- **Association precedence** (once INIT-SCHANGELOG-001 ships entry-level
  `rmis` in CHANGELOG.json): 1) entry-level `rmis` — human-curated, highest
  precision; 2) trailer-chain walk — automatic fallback; 3) cross-check
  between the two, disagreements flagged as curation errors.
- Extend the existing `ingest` command family (`visionstudio ingest`) with a
  releases scanner rather than adding a parallel command tree.

**The natural recording moment:** updating `CHANGELOG.json` is already the
last step of every release ritual (schangelog parse → update → validate →
generate → tag). Recording shipped status in visionstudio rides that
moment — either the release skill runs `visionstudio ingest` for the repo
after tagging, or a `visionstudio release record <repo>` verb reads the
newest CHANGELOG.json entry directly. No new habit required; the changelog
habit *is* the shipped-status habit.

## T4 — Board view

- New web-UI panel using existing toolkit primitives; columns derived from
  lifecycle status (no new status field — the board *is* the lifecycle,
  visualized):
  `proposed → Under Consideration`, `planned → Planned`,
  `executing → In Progress`, `delivery_complete|releasing → Delivered`,
  `released|closed → Shipped` (chips: `repo@tag` per associated release).
- Filters: program, repo, hidden-flag, time window (e.g. "shipped last
  quarter").
- `cancelled` initiatives excluded by default, toggleable.

## T5 — Forcing function

- **Unshipped queue:** `delivery_complete` (or `releasing`) with zero
  associated releases, sorted by staleness (`delivery_complete_at`).
- **Validate rule:** `visionstudio validate` warns on unshipped-queue
  entries older than a threshold (default 14 days).
- **Dashboard badge:** count surfaced on the main dashboard.
- On `work complete-phase` / final-phase completion, print a reminder that
  the initiative is ready for `delivery_complete` → release association →
  `released`.

## T6 — Visibility and public export

- New Initiative field `visibility` (`internal` default | `public`). Only
  `public` initiatives appear in the export; titles/descriptions are
  reviewed at flag-flip time (the flag is the approval).
- **Second filter (INIT-VISIONSTUDIO-007):** repo-level visibility guards
  the same path — private/unknown repos never export regardless of
  initiative flags. The RMI-VISIONSTUDIO-403 rail should land before or
  with RMI-VISIONSTUDIO-309/310.
- Export command emits static JSON + HTML (board layout, no JS
  dependencies) suitable for productbuildershq.com; same generate-don't-
  host philosophy as existing exports.
- **Succession path (INIT-VISIONSTUDIO-002 Phase 2):** the cloud service
  serves the same board projection at `/t/<tenant>/roadmap.json` — keep
  the export's JSON shape identical to the served shape so the site can
  switch between static and cloud sources without page changes; static
  becomes the degraded-mode fallback.

## T7 — ACTS handshake

Per-repo `released_at` gives ACTS (INIT-ACTS-001, TRD T10) a finer
acceptance mark: escaped-defect tiering for a multi-repo initiative anchors
on the release of the *repo the fix lands in*, not one global timestamp.
ACTS reads Release rows via the same store; no ACTS-side schema change.

## Risks

1. **Trailer coverage gaps** make old releases under-associated — accepted,
   confidence-labeled, improves naturally going forward.
2. **Tag semantics vary** (pre-release tags, non-semver tags) — ingest
   filters to semver-like tags by default with an override list.
3. **Visibility mistakes** on the public export — default-internal plus
   explicit per-initiative opt-in keeps the failure mode "too private,"
   never "too public."
