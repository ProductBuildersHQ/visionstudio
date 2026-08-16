# TRD: Spec Workflow Single Source of Truth + Selection & Switching

**Initiative:** `INIT-VISIONSTUDIO-005`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Draft
**Date:** 2026-08-10

## Architecture

```
specification-workflow-spec (single source of truth)
  pkg/workflows/default/*/profile.yaml     25 profiles incl. aws-product,
  (embedded FS, auto-discovered)           aws-feature, pbhq-lite, quick-fix
        │
        │  workflows.DefaultLoader() → ResolvingLoader (extends chains,
        │  template/rubric inheritance via mergeWorkflows)
        ▼
visionstudio pkg/specworkflow
  Loader (existing wrapper)                synthesis + eval (already wired)
  Resolve(init) ─────────────────────────► loader-backed; store no longer consulted
  SyncFromCatalog(store, loader)           DB rows = index/cache for API listing
  DefaultWorkflowForType                   maintenance/refactor/migration→quick-fix
        │                                  feature/compliance/default→pbhq-lite
        ▼
  CLI (workflow list/get/sync,             API (/api/execution workflowId,
       initiative update --workflow)            /api/specs workflows detail,
                                                /api/spec-files roles)
        ▼
  web/ InitiativeDetail + SpecViewer       workflow-driven rendering
```

## Key Technical Decisions

### TD-1: Upstream profiles carry the flow structure

`workflow.Workflow` already has `Execution{Sequence, Phases, ReviewGates}` — the "non-flat" structure the D2 diagrams encode. The aws profiles simply never used it, and pbhq-lite's block was silently dropped because its YAML used `workflow:` where the struct tag is `yaml:"execution"` (yaml.v3 ignores unknown keys). Fix the key, add `execution:` blocks to both aws profiles mirroring the D2 phase groupings, and correct the synthesis `sources` to match the D2 edges (notably `ird ← trd`, which was missing entirely, and `tpd ← prd,trd,uxd`).

### TD-2: Spec-type IDs vs. filenames

Upstream uses spec-type IDs (`prd`, `opportunity-spec`, `narrative-6p`); VisionStudio's store/scaffolding uses filenames (`PRD.md`). Bridge with one canonical mapping: `SpecFileName(specType) = strings.ToUpper(specType) + ".md"` (e.g. `OPPORTUNITY-SPEC.md`, `NARRATIVE-6P.md`). `deriveSpecType` (api.go) remains the inverse for discovery, extended to recognize the new types.

### TD-3: DB rows demoted to an index

`SpecWorkflow` rows stop being a competing definition. `SyncFromCatalog`:

1. Upserts one row per loader-available workflow (name, description, required/optional as filenames).
2. Remaps initiatives on retired IDs (`pbhq-standard` → `pbhq-lite`).
3. Deletes rows absent from the catalog and unreferenced by any initiative (requires a new `DeleteSpecWorkflow` store method on the `SpecWorkflowStore` interface, doltstore, memstore).
4. Is idempotent; `workflow seed` remains as a deprecated alias.

`Resolve(init)` drops its store parameter and resolves purely via the loader (explicit `WorkflowID`, else `DefaultWorkflowForType(init.InitType)`).

### TD-4: Dual API struct discipline

Per the documented gotcha, every API change lands in **both** `pkg/apitypes/types.go` (schema source → TS pipeline) and the local runtime structs in `cmd/visionstudio/api.go`:

- `APIInitiative.WorkflowID` (runtime struct is missing it today — existing type/runtime mismatch).
- Workflow payload gains `extends`, `sequence`, `phases[]{id,name,specs}`.
- Spec-file entries gain `role: "required"|"optional"|"extra"`.

Regenerate: `go generate ./pkg/apitypes && cd web && npm run generate:types`.

### TD-5: GetSynthesisGuidance PromptContext fallback

`SynthesisRule` has both `guidance` and `prompt_context`; aws-feature uses only the latter, which `Loader.GetSynthesisGuidance` currently drops. Return `Guidance`, falling back to `PromptContext` when empty.

### TD-6: Frontend derives everything from data

Both `PBHQ_LITE_SPECS` constants are deleted. `InitiativeDetail`/`SpecViewer` look up the initiative's workflow (via `workflowId` from `/api/execution` + the workflow detail from `/api/specs`) and derive: expected/required doc set, progress denominator, tab sort (workflow `sequence`, then optional, then extras), and the `WorkflowDiagram` (real workflow name, per-doc boxes with eval status). Extras render with an "Extra" badge. Fallback when unset: type-based default workflow.

## Data Model Changes

| Layer | Change |
|-------|--------|
| `pkg/store` | `SpecWorkflowStore` gains `DeleteSpecWorkflow(ctx, id)`; no Ent schema changes (additive-free) |
| `pkg/specworkflow` | `BuiltInWorkflows` removed; `SyncFromCatalog`, `SpecFileName`, loader-backed `Resolve` added |
| `pkg/apitypes` + api.go | Fields per TD-4 |

No database migration required: the `spec_workflows` table shape is unchanged; only row contents are synced/remapped.

## Live Data Migration (one-time, via `workflow sync`)

| Current | After |
|---------|-------|
| `pbhq-standard` ×2 initiatives | `pbhq-lite` (identical required set) |
| `pbhq-lite` ×2 (local 2-doc meaning) | `pbhq-lite` (upstream 4-doc meaning — PRD/TRD become required; expected, flagged) |
| `quick-fix` ×1 | `quick-fix` (now upstream-defined) |
| NULL ×22 | unchanged; resolve via `DefaultWorkflowForType` |
| `aws-working-backwards` row (0 refs) | deleted |

## Development Wiring

`specification-workflow-spec` is consumed via a **local `replace` directive** in `go.mod` during this effort (no commits/tags/releases). The replace must be removed (and the upstream released/pinned) before any push, per the Pre-Push Checklist.

## Testing Strategy

- Upstream: `pkg/workflows/execution_test.go` — parse regression (pbhq-lite execution non-nil), D2 conformance for both aws profiles (sequence, phases, synthesis sources), quick-fix requirement set, template inheritance through `extends`.
- VisionStudio: unit tests for `SyncFromCatalog` (upsert/remap/delete, idempotency, memstore), `Resolve` fallback order, `SpecFileName`/`deriveSpecType` round-trip; `go vet`, `golangci-lint`, `npx tsc --noEmit`.
- End-to-end: scratch hidden initiative switched `aws-feature` → `aws-product`; verify `/api/spec-files` roles and Initiative page rendering with a loose `NOTES.md`.
