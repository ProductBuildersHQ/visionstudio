# ROADMAP — INIT-VISIONSTUDIO-001: Unified Multi-Domain Backend + Frontend

Repo slug: `VISIONSTUDIO` (cross-repo RMIs carry their own repo's slug).
Phase status is derived from member RMI statuses. This initiative uses the
RMI-VISIONSTUDIO-1xx block (RMI-VISIONSTUDIO-001..002 belong to
INIT-DEVXREPORTS-001); the hosted follow-on uses the 2xx block.

## Phase 1 — IR Contract Layer

Compose domain IRs in `pkg/ir` by importing source modules (TRD T1).

| RMI | Title | Notes |
|-----|-------|-------|
| RMI-VISIONSTUDIO-101 | Scaffold `pkg/ir` + go.work workspace | Aliases for Program/Initiative/Phase/RMI/Assignment/Evidence; json tags added upstream in prism-build `pkg/store`; `go.work` spanning all 5 repos for local dev, CI pins tags |
| RMI-VISIONSTUDIO-102 | Integrate devfolio devx dashboard IR | Import `output/devxdashboard`; period reports included |
| RMI-VISIONSTUDIO-103 | Integrate prism-maturity IR | Root `prism` package (MaturityModel, cells, levels); verify JSON-clean |
| RMI-VISIONSTUDIO-104 | Integrate prism-roadmap IR | `rmi`, `roadmap`, `goals` packages; verify JSON-clean |
| RMI-VISIONSTUDIO-105 | Composed `RepoSnapshot` + JSON Schema generation | invopop/jsonschema → `schemago lint` → `go:embed` |

## Phase 2 — Local Web Foundation

Web-first (TRD T5): a unified local website — the successor to the :9400
dashboard — served against a JSON API on prism-build's existing server.
No backend code moves yet.

| RMI | Title | Notes |
|-----|-------|-------|
| RMI-PRISMBUILD-101 | JSON API on prism-build dashboard server | `/api/execution`, `/api/spend`, `/api/maturity`, `/api/specs` reusing `loadDashboardData` + maturity/judge stores; golden-file contract tests; migrates verbatim in Phase 4 |
| RMI-PRISMBUILD-102 | `prismctl spec sync` — disk→DB spec reconciliation | Scan `docs/specs/initiatives/{INIT-ID}/` in each registered repo, backfill `Initiative.Specs` map, flag legacy-location spec files for migration; keeps the specs panel honest when agents write specs by hand |
| RMI-VISIONSTUDIO-106 | Unified local web app shell | Vite + React SPA at localhost, served by daemon in prod mode / vite in dev; app code platform-agnostic for later Electron reuse |
| RMI-VISIONSTUDIO-107 | Shared UI toolkit + chart primitives in web app | Port toolkit (badges/states/bars) and pie/bar/line from DevXDashboardView; add radar primitive |

## Phase 3 — Unified Panels

The daily-driver dashboard: specs, execution, maturity, and devfolio
accomplishments + token costs in one UI. Absorbs INIT-PRISMCONTROL-003
Phase 4 (RMI-PRISMCONTROL-116..119 superseded).

| RMI | Title | Notes |
|-----|-------|-------|
| RMI-VISIONSTUDIO-108 | Execution dashboard panel | Programs → initiatives → phases → RMIs; parity with `prismctl dashboard` |
| RMI-VISIONSTUDIO-109 | Spend + accomplishments panel | devfolio-style report: what shipped per period/initiative + token/cost breakdown |
| RMI-VISIONSTUDIO-110 | Maturity radar chart panel | Renders CapabilityModel assessments (in Dolt today); domain/stage models added after Phase 5 ingest |
| RMI-VISIONSTUDIO-111 | Specs panel | Workflow compliance (specs required vs present) + judge scores per initiative |
| RMI-VISIONSTUDIO-112 | Composed initiative view | Specs + execution + spend + maturity on one screen; retire Go-template :9400 dashboard after parity |

## Phase 4 — Backend Migration from prism-build

Move Dolt/Ent/service into visionstudio (TRD T2); the daemon re-serves the
Phase 2 JSON API unchanged (contract tests prove it). The full prismctl
command surface migrates into the `vistudio` binary — no slim-down, no
feature loss; prism-build stays intact with deprecation pointers.

| RMI | Title | Notes |
|-----|-------|-------|
| RMI-VISIONSTUDIO-113 | Move Ent schema + generated code | Keep `//go:build dolt` pattern |
| RMI-VISIONSTUDIO-114 | Move Store/UnitOfWork interfaces, memstore, doltstore | Full store layer copied; embedded DB name `visionstudio` |
| RMI-VISIONSTUDIO-115 | Move `pkg/service` + JSON API onto visionstudio daemon | Same contract as RMI-PRISMBUILD-101 via `pkg/webapi`; web app re-points via base URL only; daemon serves the built SPA |
| RMI-VISIONSTUDIO-116 | visionstudio CLI for DB-backed orchestration | Initiative/RMI/assignment commands; prismctl DB commands deprecated with pointers |
| RMI-VISIONSTUDIO-117 | Dolt data migration + full prismctl feature migration (vistudio CLI) | Branch backup, copy `prismcontrol` DB to `~/.productbuildershq/visionstudio`, row-count diff verified; ALL prismctl commands (mcp, spec, maturity, context, export, ingest, report, validate, roadmap, release, dashboard) ported into the `vistudio` binary; prism-build kept intact |

## Phase 5 — Multi-Domain Schema + Ingest

Grow the Dolt schema to all domains; ingest per-repo IR files.

| RMI | Title | Notes |
|-----|-------|-------|
| RMI-VISIONSTUDIO-118 | Persist devx metrics (schema + store) | From devfolio IR; joins to initiatives for spend |
| RMI-VISIONSTUDIO-119 | Persist prism-roadmap artifacts | Roadmaps/goals keyed by repo |
| RMI-VISIONSTUDIO-120 | Persist prism-maturity domain/stage models alongside CapabilityModel | TRD T3: two maturity IRs, no forced unification; `maturityconv` mapper; radar panel gains domain/stage rendering |
| RMI-VISIONSTUDIO-121 | visionspec document registry | Spec docs/rubrics discoverable per repo/initiative |
| RMI-VISIONSTUDIO-122 | `visionstudio ingest` command | Imports `*.ir.json` from a repo into Dolt, org-scoped |

## Phase 6 — Electron Alignment (after web solidifies)

The Electron IDE adopts the unified web app once it is the proven daily
driver; existing desktop views keep working untouched until then.

| RMI | Title | Notes |
|-----|-------|-------|
| RMI-VISIONSTUDIO-123 | Electron renderer loads unified web app | Thin shell over the same SPA; single component source |
| RMI-VISIONSTUDIO-124 | Desktop-specific integration parity | Project/file access, existing views (AIDLC, V2MOM, …) reachable from unified shell |

## Follow-On Initiative

Remote backend + multi-tenant hosted mode lives in **INIT-VISIONSTUDIO-002**
(proposed): RMI-VISIONSTUDIO-201..205. Nothing in Phases 1–6 may preclude
hosted mode (HTTP-only frontend contract, org columns preserved); the
unified web app is the same artifact the hosted website will serve.
