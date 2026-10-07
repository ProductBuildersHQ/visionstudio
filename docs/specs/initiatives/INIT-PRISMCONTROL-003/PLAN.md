# Unified Specs + Execution + Spend Platform — Implementation Plan (PLAN)

**Status:** Executing
**Date:** 2026-08-01
**Initiative:** `INIT-PRISMCONTROL-003`
**Program:** `PROG-PRISM-CONTROL`

## Context

Three systems currently cover adjacent concerns:

- **visionspec** — spec workflow definitions (AWS Working Backwards, Big Tech product/feature) and LLM-as-a-Judge rubrics, as files and skills.
- **visionstudio** — Electron IDE for authoring specs and viewing judge results.
- **prism-control** — execution tracking (programs → initiatives → phases → RMIs) with Dolt storage, attribution, and reporting.

This initiative merges the specs layer into prism-control's database so spec building and spec execution become one system, standardizes where specs live in repos, and positions prism-control as the backend for a refactored visionstudio (Electron shell + extractable web UI).

**Explicitly out of scope (already delivered):**

- Token spend ingestion, attribution, and reporting — delivered by `INIT-TOKENATTRIB-001` (`pkg/tokens` TokenSource over omnidevx events and `devx` Dolt, session→assignment join, `prismctl report tokens`, dashboard cost panels). Phase 4 reuses this data; nothing re-implements it.

## Design Decisions

### D1 — Features map to initiatives, not a new entity

No `Feature` entity. A feature is an initiative with `type=feature`. Long-lived capability areas that need more work later spawn new initiatives. `Initiative.type` enum: `feature | maintenance | migration | compliance | refactor`, default `feature`.

### D2 — Canonical spec location

```
{repo}/docs/specs/
├── PRD.md                              # repo-level (optional)
├── TRD.md                              # repo-level (optional)
├── ROADMAP.md                          # master roadmap (prismctl roadmap sync)
└── initiatives/
    └── {INIT-ID}/
        ├── PRD.md | TRD.md | PLAN.md   # per workflow requirement
        ├── ROADMAP.md                  # prismctl syncs phases/RMIs
        └── .judge/                     # judge results (git-tracked JSON)
```

Legacy layouts (`features/{name}/`, `parity-YYYY-MM-DD/`, `v{X.Y.Z}/`) stay in place; `prismctl validate` flags initiatives whose `specs` map points outside the convention. New initiatives scaffold into the canonical path.

### D3 — Workflow gradient

| Workflow | Specs required | Default for type |
|----------|---------------|------------------|
| `quick-fix` | ROADMAP | maintenance |
| `pbhq-lite` | PLAN, ROADMAP | refactor, migration |
| `pbhq-standard` | PRD, TRD, PLAN, ROADMAP | feature, compliance |
| `aws-working-backwards` | PRFAQ, PRD, TRD, PLAN, ROADMAP | (opt-in for big bets) |

Workflows are rows in the database (SpecWorkflow), not hardcoded, so visionspec's larger enterprise flows can be registered later. An initiative can override its default via `workflow_id`.

### D4 — Judge results live in both places

`prismctl spec judge` writes a `JudgeResult` row (queryable, trend-able) and mirrors a JSON snapshot into `docs/specs/initiatives/{INIT-ID}/.judge/` (reviewable in PRs, survives DB resets).

### D5 — visionstudio refactor direction

Extract shared React components (`packages/ui/`) usable from both the Electron shell and a future multi-tenant Next.js host. prism-control is the single backend: execution data, spec workflow status, judge scores, token reports, maturity assessments — all served from the same Dolt database.

## Phase 1 — Schema Foundation

**Theme:** Model workflows, rubrics, and results; type initiatives.

- `RMI-101` `Initiative.type` enum field (feature/maintenance/migration/compliance/refactor, default feature) + `workflow_id` reference
- `RMI-102` `SpecWorkflow` entity: id, name, description, specs_required, specs_optional, init_types
- `RMI-103` `JudgeRubric` entity: workflow FK, spec_type, criteria JSON, prompt_template
- `RMI-104` `JudgeResult` entity: initiative FK, spec_path, rubric FK, score, rationale, model, evaluated_at
- `RMI-120` Seed built-in workflows (quick-fix, pbhq-lite, pbhq-standard, aws-working-backwards) + type→workflow defaults
- ~~`RMI-105`/`RMI-106`~~ cancelled — token entities duplicate `INIT-TOKENATTRIB-001`

**Exit criteria:** migrations run against Dolt; `prismctl initiative get` shows type and effective workflow; seeded workflows queryable; `go test ./...` green.

## Phase 2 — CLI Commands

**Theme:** Scaffold, validate, and judge specs from the CLI.

- `RMI-107` `prismctl spec init INIT-XXX` — scaffold `docs/specs/initiatives/{INIT-ID}/` from the initiative's workflow templates
- `RMI-108` `prismctl spec validate INIT-XXX` — required specs exist and are non-empty; exit code + JSON output
- `RMI-109` `prismctl spec judge INIT-XXX [--rubric --model]` — run LLM evaluation, store JudgeResult, mirror to `.judge/`
- `RMI-121` Spec path standardization — convention in docs, `validate` rule for non-conforming spec paths
- ~~`RMI-110`/`RMI-111`~~ cancelled — spend import/report duplicate `INIT-TOKENATTRIB-001`

**Exit criteria:** a new initiative can go create → spec init → author → spec validate → spec judge end-to-end; validate catches a deliberately missing required spec.

## Phase 3 — Maturity Model

**Theme:** Capability models and point-in-time assessments.

- `RMI-112` `CapabilityModel` entity (domains + levels JSON; sample DORA model loads)
- `RMI-113` `MaturityAssessment` entity (model FK, initiative/repository scope, per-domain scores with evidence)
- `RMI-114` `prismctl maturity assess` — interactive domain-by-domain assessment
- `RMI-115` `prismctl maturity report` — current levels, delta vs previous, gap to target

**Exit criteria:** assess and report round-trip on a sample model; gaps can seed new initiatives (documented workflow).

## Phase 4 — visionstudio Integration

**Theme:** One UI over the unified backend.

- `RMI-116` Extract shared UI components into `packages/ui/` (Electron + web hosts)
- `RMI-117` Execution dashboard panel (initiatives/phases/RMIs from prism-control API)
- `RMI-118` Spend visualization panel — renders existing `report tokens` data; no new attribution logic
- `RMI-119` Maturity radar chart panel (current vs target vs previous)

**Exit criteria:** visionstudio desktop shows specs, execution status, spend, and maturity for a real initiative from the shared Dolt database.

## Dependencies & Risks

- **Depends on (delivered):** `INIT-TOKENATTRIB-001` for all spend data.
- **Coordination:** visionspec framework content (rubrics, templates) needs conversion into SpecWorkflow/JudgeRubric rows — done incrementally, starting with pbhq workflows; AWS WB imported when first needed.
- **Risk:** visionstudio refactor (RMI-116) is the largest single item; if extraction stalls, phases 1–3 still deliver standalone CLI value.
- **Risk:** judge quality depends on rubric prompts; start with one rubric per spec type and iterate using stored JudgeResults as the eval set.
