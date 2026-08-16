# PLAN: Spec Workflow Single Source of Truth + Selection & Switching

**Initiative:** `INIT-VISIONSTUDIO-005`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Draft
**Date:** 2026-08-10

## Sequencing Narrative

The order is dictated by one constraint: **nothing downstream can be trusted until the catalog is right.** Every consumer — CLI, API, frontend, synthesis, evaluation — keys off workflow definitions, so the upstream profiles must be corrected first, and the two-catalog split must be collapsed second. Only then do the surfaces that *display* workflow state (CLI output, API payloads, React panels) have a stable contract to build against.

### Phase 1 — Upstream Workflow Catalog (specification-workflow-spec)

Fix the definitions at the source. The silent `workflow:`/`execution:` parse bug is first because it invalidates any assumption that profiles round-trip; then align `aws-product`/`aws-feature` to the authoritative visionspec D2 flows and add `quick-fix`. Regression tests pin the D2 conformance so future profile edits can't silently drift. This phase has no VisionStudio dependency and no schema changes.

### Phase 2 — Catalog Unification (visionstudio core)

Wire visionstudio to the corrected upstream via a local `replace` directive, then retire catalog B: `SyncFromCatalog` replaces `SeedBuiltIn`, `Resolve` goes loader-only, the type-default mapping is updated, and the store gains `DeleteSpecWorkflow`. Depends on Phase 1 (the definitions being synced must already be correct). The `GetSynthesisGuidance` PromptContext fallback lands here because it is part of making the loader the trustworthy single interface.

### Phase 3 — Selection & Switching Surfaces (CLI + API)

With one catalog, expose it: `initiative update --workflow`, loader-backed `workflow list/get/sync`, workflow visibility in `initiative get/list`, then the three API additions (runtime `workflowId`, workflow detail payload, spec-file roles) with type regeneration. Depends on Phase 2 (validation and payloads read the loader). CLI before API only because the CLI is also the migration tool the API verification relies on.

### Phase 4 — Workflow-Aware UI + Migration & Docs

The frontend consumes the new contract: dynamic spec sets, progress, ordering, diagram, Extra badges. Then run `workflow sync` against live data, verify end-to-end with a scratch aws-feature→aws-product initiative, and update docs (CLAUDE.md, dashboard guide, quickstart, CHANGELOG). Last because it exercises everything above; a UI built against the old contract would need immediate rework.

## Dependencies Between Phases

```
Phase 1 (upstream profiles) ──► Phase 2 (unification) ──► Phase 3 (CLI+API) ──► Phase 4 (UI+migration+docs)
```

Strictly linear at phase granularity; within phases, items are parallelizable except where RMI dependencies note otherwise.

## Milestones

| Milestone | Definition of done |
|-----------|-------------------|
| M1: Profiles trustworthy | Upstream tests pass incl. D2-conformance suite; 25 workflows load with correct execution blocks |
| M2: One catalog | `BuiltInWorkflows` gone; `SyncFromCatalog` + loader-only `Resolve` unit-tested; visionstudio builds against local upstream |
| M3: Switchable | `initiative update --workflow` works; APIs expose workflowId/detail/roles; generated TS types updated |
| M4: Visible & migrated | Initiative page fully workflow-driven with Extra badges; live DB synced (remaps verified); docs updated |

## Risks

- **Semantics shift for existing initiatives:** the two former 2-doc `pbhq-lite` initiatives gain PRD/TRD as required after consolidation. Accepted and flagged; no data loss.
- **Replace directive leakage:** the `go.mod` replace must not be pushed. Mitigated by explicit end-of-work reporting and the Pre-Push Checklist.
- **Filename mapping for new doc types** (`OPPORTUNITY-SPEC.md`, `NARRATIVE-6P.md`): `deriveSpecType` substring matching must be extended carefully to avoid misclassifying (e.g. `PRD.md` vs `OPPORTUNITY-SPEC.md` both contain no overlap — verified by round-trip test).
