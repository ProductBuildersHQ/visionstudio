# Quick Start

This walks through the `visionstudio` CLI and web dashboard. If you haven't installed it yet, see [Installation](installation.md) first.

## Start VisionStudio

```bash
visionstudio app start
```

This starts the Dolt database (if it isn't already running) and opens the dashboard in your browser at `http://localhost:9400`. Press `Ctrl-C` to stop.

## Register a Repository

VisionStudio tracks initiatives against repositories it knows about. Register one with:

```bash
visionstudio registry add --org <github-org> --name <repo-name> --path /path/to/your/repo
```

See `visionstudio registry --help` for the full set of registry commands.

## Navigating the Dashboard

### All Initiatives (`/`)

The home view lists every initiative, grouped by program. Standalone initiatives (not attached to a program) have their own view at `/standalone`.

### Program view (`/program/:programId`)

Shows every initiative under a single program.

### Initiative detail (`/initiative/:initiativeId`)

Shows an initiative's phases, roadmap items (RMIs), and its spec documents. From here you can open the spec viewer for any spec type (PRD, TRD, PLAN, ROADMAP, etc.).

### Spec Viewer (`/initiative/:initiativeId/spec/:specType`)

Shows a spec's Markdown source and rendered views, with a toggle between them, along with its LLM-as-a-Judge evaluation result if one exists — per-category scores, findings, and next steps. Supports copying a link and exporting to PDF.

### Repositories (`/repositories`)

The repository catalog, with per-repository RMI counts and progress. Click through to a repository's detail view (`/repository/*`) to see the initiatives that touch it.

### Maturity (`/maturity`)

Framework-based capability maturity assessments.

### Performance (`/performance`)

Token spend and cost, broken down by model, initiative, phase, and RMI — sourced from ingested usage data (`visionstudio db ingest-tokens`).

## Working with Initiatives from the CLI

The dashboard is a view over the same data the CLI manages. Common commands:

```bash
visionstudio initiative list
visionstudio rmi list --initiative <initiative-id>
visionstudio spec judge record <initiative-id> <spec-file> --score <1-5> --rationale "<why>" --model <model-id>
visionstudio roadmap import docs/specs/initiatives/<initiative-id>/ROADMAP.md   # sync ROADMAP.md into the database
```

Run `visionstudio --help` for the full command list, or `visionstudio <command> --help` for any subcommand.

## LLM-as-a-Judge Evaluation

Specs are evaluated against workflow rubrics (`structured-evaluation/rubric.Rubric`) — a 1–5 integer score, per-category findings with severity, and an overall pass/fail decision. Evaluation results are stored as `*.eval.json` files under `docs/specs/initiatives/<id>/evaluations/` and shown in the Spec Viewer.

## Agent Access (MCP)

Agent sessions can query initiatives, RMIs, and work assignments directly via the MCP stdio server:

```bash
visionstudio mcp
```
