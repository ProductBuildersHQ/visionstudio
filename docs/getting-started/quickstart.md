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

The sidebar has four sections: **Initiatives** (programs, standalone initiatives, and their RMIs), **Repositories**, **Maturity**, and **Performance**. See the [Dashboard Guide](../dashboard/overview.md) for a full tour of each — what every panel shows, every clickable element, and what the status colors and icons mean.

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

Specs are evaluated against workflow rubrics (`structured-evaluation/rubric.Rubric`) — a 1–5 integer score, per-category findings with severity, and an overall pass/fail decision. See [Specs & Evaluation](../dashboard/specs-and-evaluation.md) for how results are recorded and displayed.

## Agent Access (MCP)

Agent sessions can query initiatives, RMIs, and work assignments directly via the MCP stdio server:

```bash
visionstudio mcp
```
