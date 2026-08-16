# PRD: Authentic AWS Working Backwards + Web Creation Surface

**Initiative:** `INIT-VISIONSTUDIO-009`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Delivered (retroactive record)
**Date:** 2026-08-14

> Retroactive record: this initiative logs work delivered 2026-08-10 → 2026-08-13
> immediately after INIT-VISIONSTUDIO-005, so that landing commits can carry
> `Refs:` trailers. Written after the fact; kept concise.

## Problem

INIT-VISIONSTUDIO-005 made the workflow catalog trustworthy, but the AWS
Working Backwards profiles it exposed were *adaptations*, not the method:
they started from an MRD (working *forwards* to a press release), were named
by a product/feature taxonomy Amazon doesn't use, had no iteration or review
gates, and their templates/rubrics referenced the Leadership Principles only
loosely. Separately, the dashboard could display initiatives but not create
them — workflow selection existed only in the CLI.

## Goals

- The AWS profiles are as **authentic** to Amazon practice as possible while
  being **effective in an AI-native environment** (human writes the vision,
  agents build the evidence, LLM-as-a-Judge scores the drafts, humans keep
  the gates).
- Initiatives can be created **from the dashboard**, selecting any catalog
  workflow with an informative preview.
- Anyone can see, from the Definition tab, **what a document should contain
  and how it will be judged** before writing it.

## Functional Requirements

| # | Requirement |
|---|-------------|
| FR-1 | Both AWS profiles start with the human-authored Press Release; the FAQ synthesizes from it; MRD/OpportunitySpec are optional post-FAQ deepening |
| FR-2 | Profiles are named by decision reversibility (`aws-one-way-door`, `aws-two-way-door`); ceremony differs by door (6-pager + decision meeting required vs optional); both deepening docs available in both |
| FR-3 | PR/FAQ iteration is first-class (iteration trigger + prfaq_review gates); existing initiatives on old IDs migrate automatically |
| FR-4 | All 16 Leadership Principles carried in profile metadata with source reference; templates and rubrics encode the transcripts' judgeable teachings |
| FR-5 | `POST /api/initiatives` creates an initiative with catalog-validated workflow; the dashboard exposes it via a New Initiative form with per-workflow preview |
| FR-6 | Clicking a workflow-diagram document (or spec-viewer buttons) shows its authoring template and judge rubric from the embedded catalog |

## Success Metrics

- Zero references to the retired profile IDs outside historical release notes; live DB fully remapped by `workflow sync`.
- Conformance suites pin the flows (D2 alignment), the gates, and the LP grounding — profile drift fails CI.
- An initiative can go from "doesn't exist" to "created on aws-one-way-door with its flow rendered" entirely in the dashboard.
