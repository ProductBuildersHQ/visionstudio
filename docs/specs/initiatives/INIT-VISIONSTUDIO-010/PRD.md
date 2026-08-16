# PRD: Registry & RMI Lifecycle CLI

**Initiative:** `INIT-VISIONSTUDIO-010`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Planned
**Date:** 2026-08-15

> **Spec scope note:** this initiative is intentionally captured as PRD +
> ROADMAP only — no TRD/PLAN. It's reversible internal tooling (a two-way
> door), the enhancement request below is a complete requirements input, and
> the implementation detail lives in the roadmap RMIs. Add TRD/PLAN later only
> if a specific item turns out to need design debate.

## Source

Enhancement request authored by the `spec-mcp-server` session (2026-08-15/16)
while reconciling `visionstudio` registry and RMI records after the
`specification-workflow-spec` → `visionspec` repo merge. Every gap was
reproduced against the installed binary, not inferred from source. The same
class of blocker was hit earlier during the aws-door rename
(`RMI-VISIONSTUDIO-516`), which is what makes this recurring rather than a
one-off.

## Problem

The `visionstudio` registry can *create* repo records (`registry add`) and
read them, but a record cannot be **fixed, retired, or health-checked** once
the underlying repo is renamed, moved, or merged away. The sws→visionspec
merge left a permanent dangling record whose registered path no longer exists
on disk, with no CLI path to repair it and no proactive way to have noticed.
Reassigning the affected RMIs required a hand-rolled 25× shell loop. Separately,
query commands filter on the full repo ID while displaying the short name —
a silent wrong-answer trap — and none of the `list`/`get` commands emit
machine-readable output, forcing fragile `awk`-parsing of fixed-width tables.

## Goals

Map to the seven enhancement items, grouped into three coherent areas:

| Area | Enhancement items | Delivered as |
|------|-------------------|--------------|
| Registry entry lifecycle & health | §1 (update/archive/remove), §2 (doctor), §7 (add conflict detection) | `registry update`/`archive`/`remove`/`doctor`; add-time warnings; `superseded_by` field |
| Bulk RMI reassignment | §5 | `rmi bulk-update`; folded into `registry archive --reassign-rmis` |
| Query correctness & output | §3 (silent `--repo` trap), §4 (`--format json`), §6 (`initiative list` filters) | short-name resolution + did-you-mean hint; JSON output; initiative filters |

## Requirements

| # | Requirement | Item |
|---|-------------|------|
| FR-1 | A registered repo entry can be repointed (`--path/--org/--branch/--name`), archived (status + superseded-by, record preserved), or hard-removed (`--force`) | §1 |
| FR-2 | `registry doctor` flags entries whose path is missing, isn't a git working tree, or whose git remote no longer matches the registered ID | §2 |
| FR-3 | `registry add` warns (not blocks) on same-path / same-remote duplicates and on a missing-or-non-git `--path` | §7 |
| FR-4 | A repo's RMIs can be reassigned in one auditable operation, ideally as a side of `registry archive --superseded-by --reassign-rmis` | §5 |
| FR-5 | `--repo` filters resolve the short name shown in output, and a zero-exact-match `--repo` value emits a did-you-mean hint rather than a silent empty result | §3 (correctness) |
| FR-6 | `rmi list`, `rmi get`, `initiative list`, `registry list` support `--format json` | §4 |
| FR-7 | `initiative list` supports `--repo`, `--status`, `--program` filters | §6 |

## Non-Goals

- No change to the registry *data model* beyond the additive `superseded_by`
  field (the `status` field for active/archived already exists).
- No web/API surface — this is CLI-only lifecycle tooling.
- No automatic repo discovery/registration changes (`registry scan` unchanged).

## Success Metrics

- The dangling `specification-workflow-spec` record (and any future one) is
  fixable and self-surfacing: `registry doctor` reports it; `registry archive
  --superseded-by visionspec --reassign-rmis` retires it and migrates its RMIs
  in one command.
- No scripted reconciliation needs to parse fixed-width tables: every relevant
  `list`/`get` has `--format json`.
- Filtering by the repo short name shown on screen never silently returns zero
  rows.

## Design Fit (why this is low-risk)

- **Registry mutations** follow the established `setProgramHidden`/
  `setInitiativeHidden` load → no-op-if-same → set → `UpdatedAt` → persist
  pattern (documented in `CLAUDE.md`).
- **JSON output** extends the existing `--format` pattern already in
  `context`/`report`/`validate`/`roadmap`.
- **`archive`** is a `status` change (that column already exists) plus an
  additive `superseded_by` field — a safe, backwards-compatible migration.
