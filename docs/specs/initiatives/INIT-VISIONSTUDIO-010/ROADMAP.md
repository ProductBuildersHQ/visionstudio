# ROADMAP: Registry & RMI Lifecycle CLI

**Initiative:** `INIT-VISIONSTUDIO-010`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Planned
**Date:** 2026-08-15

Source: enhancement request authored by the `spec-mcp-server` session while
reconciling registry/RMI records after the `specification-workflow-spec` →
`visionspec` repo merge (see PRD for the item-by-item mapping). Every gap was
reproduced against the installed binary. Prior instance of the same blocker:
`RMI-VISIONSTUDIO-516` (the aws-door rename).

Priority within the initiative follows the request's table: the §3 correctness
bug and `registry doctor` are do-first; hard-delete and add-time warnings are
lowest.

## Phase 1 — Registry Entry Lifecycle & Health

**Theme:** Make a registered repo record fixable, retirable, and self-auditing (enhancement §1, §2, §7)

- [ ] `RMI-VISIONSTUDIO-524` Add `superseded_by` field to Repository (Ent schema + store/doltstore/memstore, additive migration; the `status` field already exists for active/archived)
- [ ] `RMI-VISIONSTUDIO-525` `registry doctor` — walk every registered `--path`, flag missing directory, non-git working tree, or `git remote origin` URL that no longer matches the registered ID (highest-leverage: auto-surfaces dangling records)
- [ ] `RMI-VISIONSTUDIO-526` `registry update <repo-id> [--path --org --branch --name]` — repoint/edit an existing entry (load → no-op-if-same → set → UpdatedAt → persist, mirroring setProgramHidden)
- [ ] `RMI-VISIONSTUDIO-527` `registry archive <repo-id> [--reason --superseded-by <new-id>]` — status change preserving the record and a superseded-by pointer (preferred over remove for merge/rename)
  - Depends on: `RMI-VISIONSTUDIO-524`
- [ ] `RMI-VISIONSTUDIO-528` `registry remove <repo-id> [--force]` — hard delete for true mistakes only (requires --force; refuses if RMIs still reference it without --force)
- [ ] `RMI-VISIONSTUDIO-529` `registry add` conflict/duplicate detection — warn (not block) on an existing entry with the same real path or same git remote URL, and on a `--path` that is missing or not a git working tree

## Phase 2 — Bulk RMI Reassignment

**Theme:** Repointing a repo's RMIs is one auditable operation, not a hand-rolled loop (enhancement §5)

- [ ] `RMI-VISIONSTUDIO-530` `rmi bulk-update --repo <old-id> --set-repo <new-id> [--initiative I]` — reassign the repository field across all matching RMIs in one call (with a dry-run preview)
- [ ] `RMI-VISIONSTUDIO-531` Fold bulk reassignment into archive: `registry archive <old-id> --superseded-by <new-id> --reassign-rmis` performs the registry state change and the RMI field migration as a single auditable operation
  - Depends on: `RMI-VISIONSTUDIO-527`, `RMI-VISIONSTUDIO-530`

## Phase 3 — Query Correctness & Machine-Readable Output

**Theme:** Filters resolve what's on screen, and every list/get is scriptable (enhancement §3, §4, §6)

- [ ] `RMI-VISIONSTUDIO-532` Fix the silent `--repo` short-name trap (correctness bug): resolve short names via the registry, and emit a did-you-mean hint on any zero-exact-match `--repo` value instead of a silent empty result
- [ ] `RMI-VISIONSTUDIO-533` `--format json` on `rmi list`, `rmi get`, `initiative list`, `registry list` (extend the existing context/report/validate/roadmap pattern)
- [ ] `RMI-VISIONSTUDIO-534` `initiative list --repo/--status/--program` filters, reusing 532's short-name repo resolution
  - Depends on: `RMI-VISIONSTUDIO-532`
