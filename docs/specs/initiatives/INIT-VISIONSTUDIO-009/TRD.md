# TRD: Authentic AWS Working Backwards + Web Creation Surface

**Initiative:** `INIT-VISIONSTUDIO-009`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Delivered (retroactive record)
**Date:** 2026-08-14

> Retroactive record — see PRD. Key technical decisions as built.

## TD-1: Empty-sources synthesis override

The Press Release is human-authored, but `enterprise` (the parent profile)
defines `press ← mrd`, and the per-key extends-merge would silently inherit
it. An explicit `press: {sources: []}` override in both AWS profiles makes
"human-authored" a positive assertion the conformance tests can pin —
omission is not removal under `maps.Copy` merge semantics.

## TD-2: Door names, scale decoupled

Amazon classifies decisions by reversibility, not product/feature. Renaming
required touching every reference across three repos; the existing
`retiredRemap` machinery (built in 005 for `pbhq-standard`) absorbed the
migration. Lesson encoded in a regression test: bulk renames can clobber the
remap's own keys (`"aws-product": "aws-one-way-door"` became self-referential
under a careless sed) — the keys must stay the OLD ids.

## TD-3: LP grounding at three layers

Principles live in `methodology` metadata (all 16, canonical reference URL),
as template guidance comments (visible when viewing templates — the guidance
IS the content), and as rubric categories with weights summing to 1.0
(structured-evaluation `RubricSet`, embedded via `//go:embed`, parsed at
init). The door profiles carry door-*differentiated* `decision_reversibility`
criteria — the only place the two rubric sets intentionally diverge.

## TD-4: First mutation endpoint

`POST /api/initiatives` uses `apitypes` types directly (no dual runtime
struct — that gotcha applies only to the legacy `/api/execution` family).
Validation: ID allowlist regex, catalog-loader workflow check, 409 via
GetInitiative precheck. `svc.CreateInitiative` keeps both workflow records
(edge + selection row) in step, same as the CLI path.

## TD-5: Template/rubric serving

`GET /api/workflows/{id}/specs/{type}`: template as raw markdown (comments
visible by design), rubric as a JSON string (`rubricJson`, following the
`evalJson` precedent) with a loose typed view-model in the modal — avoids
generating TS types for the full RubricSet shape.

## Testing

Upstream: D2-conformance + gates + `TestLeadershipPrincipleGrounding` (16
principles, reference URL, LP categories per door). VisionStudio: handler
tests for create (5 cases) and spec-detail (3 cases), remap regression test,
full suite 23/23 + lint + tsc at each step.
