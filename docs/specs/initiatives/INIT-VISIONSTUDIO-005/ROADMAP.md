# ROADMAP: Spec Workflow Single Source of Truth + Selection & Switching

**Initiative:** `INIT-VISIONSTUDIO-005`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Executing
**Date:** 2026-08-10

## Phase 1 — Upstream Workflow Catalog

**Theme:** Fix specification-workflow-spec profiles at the source: parse bug, D2 alignment, quick-fix

- [x] `RMI-VISIONSTUDIO-225` Fix silent `workflow:` → `execution:` YAML key parse bug in pbhq-lite profile
- [x] `RMI-VISIONSTUDIO-226` Align aws-product profile to visionspec aws-product-flow.d2 (execution phases/sequence, synthesis edges incl. missing `ird ← trd`, `spec` optional)
- [x] `RMI-VISIONSTUDIO-227` Align aws-feature profile to visionspec aws-feature-flow.d2 (execution phases/sequence, `narrative-6p` optional, `tpd ← prd,trd,uxd`, `spec` optional)
- [x] `RMI-VISIONSTUDIO-228` Add quick-fix profile upstream (extends pbhq-lite; ROADMAP-only required)
- [x] `RMI-VISIONSTUDIO-229` Add D2-conformance and parse-regression test suite (`pkg/workflows/execution_test.go`)
  - Depends on: `RMI-VISIONSTUDIO-225`, `RMI-VISIONSTUDIO-226`, `RMI-VISIONSTUDIO-227`, `RMI-VISIONSTUDIO-228`

## Phase 2 — Catalog Unification

**Theme:** Retire visionstudio's local workflow catalog; the upstream loader becomes the only definition source

- [x] `RMI-VISIONSTUDIO-230` Wire local specification-workflow-spec via go.mod replace directive (dev-only; remove before push)
- [x] `RMI-VISIONSTUDIO-231` Add `DeleteSpecWorkflow` to SpecWorkflowStore interface, doltstore, memstore
- [x] `RMI-VISIONSTUDIO-232` Replace `BuiltInWorkflows`/`SeedBuiltIn` with `SyncFromCatalog` (upsert from loader, remap `pbhq-standard`→`pbhq-lite`, delete retired unreferenced rows, idempotent)
  - Depends on: `RMI-VISIONSTUDIO-230`, `RMI-VISIONSTUDIO-231`
- [x] `RMI-VISIONSTUDIO-233` Make `Resolve` loader-only and update `DefaultWorkflowForType` (maintenance/refactor/migration→quick-fix, else→pbhq-lite); add `SpecFileName` spec-type↔filename bridge
  - Depends on: `RMI-VISIONSTUDIO-230`
- [x] `RMI-VISIONSTUDIO-234` Fix `GetSynthesisGuidance` to fall back to `PromptContext` (aws-feature guidance silently dropped today)

## Phase 3 — Selection & Switching Surfaces

**Theme:** Expose the unified catalog through CLI and API

- [x] `RMI-VISIONSTUDIO-235` Add `initiative update --workflow` (loader-validated) and show workflow in `initiative get`/`list`
  - Depends on: `RMI-VISIONSTUDIO-233`
- [x] `RMI-VISIONSTUDIO-236` Rework `workflow list`/`get` to read from loader; add `workflow sync` (deprecate `seed` as alias)
  - Depends on: `RMI-VISIONSTUDIO-232`
- [x] `RMI-VISIONSTUDIO-237` Add `workflowId` to runtime APIInitiative, workflow detail (extends/sequence/phases) to `/api/specs`, `role` (required/optional/extra) to spec-files response — both apitypes and runtime structs, regenerate TS
  - Depends on: `RMI-VISIONSTUDIO-233`

## Phase 4 — Workflow-Aware UI, Migration & Docs

**Theme:** The Initiative page adapts to any workflow; live data migrated; docs current

- [x] `RMI-VISIONSTUDIO-238` Remove PBHQ_LITE_SPECS hardcoding; derive spec set, progress, and tab order from assigned workflow (InitiativeDetail + SpecViewer)
  - Depends on: `RMI-VISIONSTUDIO-237`
- [x] `RMI-VISIONSTUDIO-239` Dynamic WorkflowDiagram (real workflow name/sequence) + Extra badge for non-workflow files + read-only workflow display
  - Depends on: `RMI-VISIONSTUDIO-238`
- [x] `RMI-VISIONSTUDIO-240` Run `workflow sync` migration on live DB; end-to-end verify aws-feature→aws-product switch with extra-file labeling
  - Depends on: `RMI-VISIONSTUDIO-236`, `RMI-VISIONSTUDIO-239`
- [x] `RMI-VISIONSTUDIO-241` Update docs: CLAUDE.md conventions, dashboard specs-and-evaluation guide, quickstart, CHANGELOG unreleased
  - Depends on: `RMI-VISIONSTUDIO-240`
