# PRD — INIT-VISIONSTUDIO-001: Unified Multi-Domain Backend + Frontend

**Initiative:** INIT-VISIONSTUDIO-001
**Status:** proposed
**Type:** feature
**Workflow:** pbhq-standard

## Problem

The ProductBuildersHQ ecosystem has grown organically into two kinds of tools:

1. **JSON-IR-first, per-repo tools** — devfolio, prism-maturity, prism-roadmap,
   visionspec. Each lives and revs with a repo, reads local files, and emits a
   JSON intermediate representation (IR) that any frontend can render. They
   demonstrate standalone functionality without a database.
2. **Database-first orchestration** — prism-control. Started as a Dolt DB with
   an Ent-backed service layer. It coordinates multi-repo initiatives
   (Program/Initiative/Phase/RMI), assignments, evidence, spend, and maturity —
   but nothing renders without a running Dolt instance.

Meanwhile the UI story is fragmented: prism-control has a Go-template dashboard
on `localhost:9400`, visionstudio has a React/Electron IDE, and a hosted
multi-tenant SaaS is the eventual goal. There is no single backend the IDE and
website can share, and no single frontend stack.

## Vision

VisionStudio becomes the **multi-domain platform**:

- **Backend:** the Dolt DB and Go service layer (moved from prism-control) grow
  to persist all domains — execution, devx/spend, maturity, roadmap, specs.
  One backend serves both the local IDE and the hosted website.
- **Frontend:** one React codebase powers the Electron IDE and the website.
  The IDE connects to a **local** Dolt backend or a **remote** hosted backend;
  the website always uses the hosted backend.
- **Per-repo tools stay standalone.** devfolio, prism-maturity, prism-roadmap,
  and a slimmed prism-control keep working against local files and emitting
  JSON IR. VisionStudio *imports their IR types* — it does not redefine them.

## Goals

1. **G1 — IR contract layer.** `visionstudio/pkg/ir` composes the domain IRs by
   importing types from their source modules (devfolio, prism-maturity,
   prism-roadmap, prism-control). Source modules remain the source of truth.
2. **G2 — prism-control split.** prism-control keeps its structs/JSON IR and
   per-repo file functionality (spec workflows, roadmap sync, export).
   The Dolt schema, store implementations, and service layer move to
   visionstudio.
3. **G3 — Unified persistence.** The Dolt schema grows beyond execution to
   persist devx metrics, roadmap artifacts, maturity assessments, and spec
   registry entries, with ingest commands that import per-repo IR files.
4. **G4 — Composed frontend.** React panels (execution dashboard, spend
   visualization, maturity radar) consume the unified backend; a composed
   initiative view shows execution + spend + maturity together. This absorbs
   prism-control Phase 4 (RMI-PRISMCONTROL-116..119).
Local + remote backends (hosted multi-tenant SaaS) is the follow-on
initiative **INIT-VISIONSTUDIO-002** — this initiative's execution ends at
G4, and Phases 1–4 must not preclude hosted mode.

## Non-Goals

- Rewriting devfolio / prism-maturity / prism-roadmap. They stay standalone.
- Migrating off Dolt. Dolt remains the database for local and hosted modes.
- Billing/payments for the SaaS (future initiative).
- Deprecating per-repo JSON IR files — they remain the demo/interchange format.

## Users

| User | Need |
|------|------|
| Solo developer (today: John) | Local IDE + local Dolt, full functionality offline |
| Team on hosted SaaS | Browser access, shared org data, multi-tenant isolation |
| Repo visitor | Per-repo tools demonstrate functionality with zero setup |
| Agent sessions (Claude Code) | CLI/MCP access to the same backend the UI uses |

## Success Criteria

- `visionstudio` imports IR types from all four source modules and publishes a
  composed JSON Schema that passes `schemago lint`.
- prism-control's Ent/Dolt/service code lives in visionstudio; prismctl's
  file-mode commands still work standalone in prism-control.
- The unified local website (specs + execution + spend/accomplishments +
  maturity in one UI) has replaced the :9400 dashboard as the daily driver,
  rendering live Dolt data.
- The Electron IDE loads the same web app as a thin shell (final phase, after
  the web UI solidifies).
- The IDE backend-picker (local Dolt or remote URL) is a success criterion
  for INIT-VISIONSTUDIO-002, not this initiative.
