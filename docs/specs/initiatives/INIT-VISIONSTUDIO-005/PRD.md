# PRD: Spec Workflow Single Source of Truth + Selection & Switching

**Initiative:** `INIT-VISIONSTUDIO-005`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Draft
**Date:** 2026-08-10

## Problem

VisionStudio supports different spec workflows for initiative definition (PBHQ Lite, AWS Working Backwards Product/Feature, etc.), but the support is inconsistent and partially broken:

1. **Two disagreeing workflow catalogs.** The upstream `specification-workflow-spec` library defines ~25 rich workflow profiles (synthesis rules, rubrics, templates) and is used by spec synthesis and LLM-as-a-Judge evaluation. VisionStudio *also* hardcodes its own 4-profile catalog (`pkg/specworkflow/seed.go`) used by all CRUD/CLI/API paths. The same ID means different things in each: `pbhq-lite` requires all four docs upstream but only PLAN+ROADMAP locally. An initiative's workflow can resolve differently depending on which subsystem asks.
2. **No workflow switching after creation.** `initiative create --workflow` exists, but nothing exposes changing it later, even though the store layer already supports it.
3. **The Initiative page is hardcoded to one workflow.** The frontend assumes PBHQ Lite's four documents everywhere (progress "N of 4", tab ordering, the workflow diagram literally labeled "PBHQ Lite Workflow") regardless of the initiative's actual workflow.
4. **Extra files are invisible as a concept.** Any `.md` dropped in an initiative's spec directory silently appears in the UI, untagged and sorted last. There is no way to tell a formal workflow document from a loose context file.
5. **The AWS Working Backwards profiles don't match their reference definition.** The authoritative flows (visionspec's `aws-product-flow.d2` / `aws-feature-flow.d2`) are phased DAGs; the upstream profiles are flat (no execution phases) and have missing/incorrect synthesis edges. A silent YAML parse bug (`workflow:` vs `execution:`) drops pbhq-lite's execution ordering entirely.

## Goals

- One catalog: every default (non-user-custom) workflow is defined in `specification-workflow-spec`; VisionStudio only consumes and switches between them.
- `aws-product` and `aws-feature` faithfully encode the visionspec D2 flows (phases, sequence, synthesis sources).
- A workflow can be selected at initiative creation **and changed afterward** via the CLI.
- The Initiative page adapts to the selected workflow: expected documents, progress denominator, tab order, and the workflow diagram all derive from the workflow definition.
- Spec files not in the selected workflow are clearly labeled as extra/context.

## Non-Goals

- Web UI for changing an initiative's workflow (CLI-only for now; the page shows the workflow read-only). The API surface remains GET-only.
- A "promote extra file to workflow document" action (labeling only).
- A forward-planning Release container (releases remain emergent; see ADR execution-hierarchy-naming).
- Migrating the legacy Electron/daemon surface to workflow awareness.

## User Stories

1. As a product builder, I create an initiative with `--workflow aws-feature` and the dashboard shows the OpportunitySpec → PR/FAQ → PRD flow, not PBHQ Lite's.
2. As a product builder, I realize mid-definition that a feature-level effort is actually a new product line, run `initiative update --workflow aws-product`, and the spec expectations update everywhere.
3. As a product builder, I keep an `IDEATION.md` alongside formal specs and the UI labels it "Extra" instead of pretending it is part of the workflow.
4. As an agent session, I resolve an initiative's workflow and get the same required-document answer that synthesis and evaluation use.

## Functional Requirements

| # | Requirement |
|---|-------------|
| FR-1 | All default workflows load from `specification-workflow-spec`'s embedded catalog; the local `BuiltInWorkflows()` catalog is removed |
| FR-2 | `workflow sync` upserts the DB workflow index from the upstream catalog, remaps retired IDs (`pbhq-standard`→`pbhq-lite`), and deletes retired unreferenced rows |
| FR-3 | `initiative update --workflow <id>` switches an initiative's workflow, validated against the loader |
| FR-4 | `initiative get`/`list` and `/api/execution` expose the initiative's workflow ID |
| FR-5 | `/api/specs` returns workflow definitions (required/optional docs, extends, sequence, phases) sourced from the loader |
| FR-6 | `/api/spec-files/{id}` labels each file `required`, `optional`, or `extra` per the initiative's resolved workflow |
| FR-7 | The Initiative page renders spec expectations, progress, ordering, and the workflow diagram from the assigned workflow definition |
| FR-8 | `aws-product`/`aws-feature`/`pbhq-lite` profiles parse with non-nil execution blocks matching the visionspec D2 flows; `quick-fix` exists upstream |

## Success Metrics

- Zero divergence: `workflow list`, spec scaffolding, synthesis, and evaluation all resolve any given workflow ID to the same definition (verifiable by unit test).
- An initiative on `aws-feature` shows 6 required documents on the Initiative page (opportunity-spec, press, faq, prd, uxd, trd, tpd = 7 specs, 6 required + prd) — exactly what the profile defines, with no PBHQ-Lite leakage.
- Switching workflow via CLI is reflected on the Initiative page after refresh with no manual data repair.
- All existing initiatives resolve to a valid workflow after `workflow sync` (5 explicitly assigned, 22 by type default).
