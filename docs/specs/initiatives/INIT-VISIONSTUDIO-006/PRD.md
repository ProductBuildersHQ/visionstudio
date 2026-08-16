# PRD — INIT-VISIONSTUDIO-006: Roadmap Board — Releases, Shipped Marks, and Public Roadmap View

**Initiative:** INIT-VISIONSTUDIO-006
**Status:** proposed
**Type:** feature
**Workflow:** pbhq-lite

## Problem

The initiative portfolio has grown past 25 entries — some fully executed,
some partial, some untouched — and three gaps have emerged:

1. **No delivered-vs-proposed view.** The dashboard shows execution state,
   but nothing answers "what shipped recently, what's in flight, what's
   proposed" at a glance — the view Slack famously maintained as a public
   Trello roadmap board.
2. **"Shipped in what release?" is unanswerable.** There is no Release
   entity. `release plan` orders dependencies, but nothing records that an
   initiative's work actually landed in `repoX v0.14.0`. Multi-repo
   initiatives need multiple release associations — one per repo.
3. **Lifecycle transitions lack a forcing function.** Several initiatives
   sit at `executing` or `delivery_complete` indefinitely. That's not just
   untidy: the ACTS quality signal (escaped defects) anchors on
   `released_at`, so unmarked initiatives contribute no quality data.

## Vision

A **roadmap board** over the existing initiative lifecycle, backed by
per-repo releases:

- Columns derive from lifecycle status: Under Consideration (`proposed`) →
  Planned (`planned`) → In Progress (`executing`) → Delivered
  (`delivery_complete`/`releasing`) → **Shipped** (`released`/`closed`,
  showing release chips like `visionstudio v0.3.0`, `acts v0.1.0`).
- **Releases are per repo** (a git tag + timestamp + changelog link);
  an initiative associates with every release that carried its work, and a
  release lists every initiative it shipped.
- **Association is automatic**: commits between a tag and its predecessor
  carry `Refs: RMI-<SLUG>-<NNN>` trailers → RMIs → initiatives. The trailer
  convention we already follow makes release↔initiative mapping computable
  from git history, including retroactively.
- The board is the **forcing function**: initiatives stuck in Delivered
  with no release attached are visibly nagging; `validate` flags them.
- A **public export** (static page for productbuildershq.com) shows the
  Slack-style outward view, restricted to initiatives explicitly marked
  public.

## Goals

1. **G1 — Release entity.** Per-repo Release (repo, tag, `released_at`,
   changelog/URL) with many-to-many initiative↔release association;
   optional RMI-level association for precision.
2. **G2 — Ingest from git.** Scan tags (and GitHub releases) across
   registered repos; auto-associate via `Refs:` trailers in tag ranges;
   backfill history for existing repos.
3. **G3 — Board view.** Lifecycle-derived columns in the web UI with
   release chips, filters (program, repo, period), and drill-down to the
   initiative detail.
4. **G4 — Forcing function.** An "unshipped" queue (Delivered with no
   release), a `validate` rule, and a dashboard badge — mark-it-released
   becomes the path of least resistance.
5. **G5 — Public export.** Static roadmap page generated from the board,
   showing only initiatives flagged public; suitable for
   productbuildershq.com.
6. **G6 — ACTS handshake.** Per-repo `released_at` refines the ACTS
   acceptance mark: escaped-defect tiering can anchor per repo-release
   rather than one timestamp per initiative.
7. **G7 — Internal/external duality.** Dolt is the internal correlation
   store holding *all* releases (public and private repos); the
   visionstudio UI renders the internal releases panel and full board.
   External artifacts — the releaselog page (INIT-PBHQSITE-001) and the
   roadmap export — are projections of public subsets, never the source
   of truth.

## Non-Goals

- Voting, comments, or subscriptions on the public board (Trello-style
  interactivity) — static export only.
- Hosting infrastructure beyond static export to the existing site.
- Perfect backfill — historical associations are best-effort from trailer
  coverage, labeled with confidence like all evidence.
- Release *automation* (tagging, changelog generation) — schangelog and the
  release skill already own that; this initiative only records outcomes.

## Success Criteria

- Every registered repo's historical tags ingested; ≥1 existing initiative
  auto-associated to its releases purely from trailers.
- All current `delivery_complete` initiatives either marked `released` with
  release associations or consciously parked — the unshipped queue reaches
  zero once.
- The board renders the full portfolio with accurate columns and release
  chips.
- A public static export exists with at least the flagship initiatives
  visible.
- ACTS (INIT-ACTS-001) can read per-repo `released_at` for its quality
  signal without schema changes on its side.
