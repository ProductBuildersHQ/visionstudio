# VisionStudio

LLM-powered platform for roadmap execution, specification authoring, and evaluation.

## What is VisionStudio?

VisionStudio tracks cross-repository initiatives, roadmap items, and specification quality in one place. It combines a Go CLI and web dashboard (`visionstudio`) backed by Dolt (a MySQL-compatible, Git-like versioned database) with LLM-as-a-Judge evaluation of specs against workflow rubrics.

## Key Features

### Roadmap Execution

- **Programs → Initiatives → Phases → RMIs** - Hierarchical tracking of cross-repository roadmap items, with dependencies and progress rollups
- **Hiding** - Programs and initiatives can each be hidden from the dashboard, with hiding a program cascading to its initiatives
- **Repositories** - Repository catalog with per-repo RMI counts and progress
- **Performance** - Token spend and cost tracking by model, initiative, phase, and RMI
- **Maturity Assessments** - Framework-based capability maturity scoring, plus SCALE platform adoption and code-leverage/reuse graphs
- **MCP Server** - Stdio server exposing initiatives/RMIs/work assignments to agent sessions

### Specification Authoring & Evaluation

- **Spec Viewer** - Source and rendered Markdown views per initiative, with PDF export
- **LLM-as-a-Judge Evaluation** - Evaluate specs against workflow rubrics, with per-category findings and next steps
- **Workflow Sync** - Sync `ROADMAP.md` files and spec workflows with the database

### CLI

- **One-command startup** - `visionstudio app start` brings up the database and web UI together
- **Database lifecycle** - `visionstudio db {init,start,stop,restart,status,commit}` for standalone database management
- **Actionable diagnostics** - connection failures explain how to fix them instead of raw driver errors
- **Serve from anywhere** - the web UI is embedded in the binary, no separate frontend build needed to use it

## Quick Start

```bash
git clone https://github.com/ProductBuildersHQ/visionstudio.git
cd visionstudio
go build -o bin/visionstudio ./cmd/visionstudio

# Bring up the database and web UI together, opens your browser
./bin/visionstudio app start
```

See [Installation](getting-started/installation.md) for the full walkthrough.

## Architecture

```
┌───────────────────────────────────────────────────────────────┐
│                    visionstudio binary                        │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │      web/ — React + Vite SPA, embedded via go:embed     │  │
│  │  • Programs → Initiatives → Phases → RMIs               │  │
│  │  • Repositories, Performance (token spend)               │  │
│  │  • Spec viewer + LLM-as-a-Judge evaluations              │  │
│  └──────────────────────┬──────────────────────────────────┘  │
│                          │ HTTP (same port)                   │
│  ┌──────────────────────▼──────────────────────────────────┐  │
│  │      cmd/visionstudio — cobra CLI + JSON API             │  │
│  │  • app/ui/db lifecycle commands                          │  │
│  │  • initiative/rmi/phase/spec/roadmap/maturity commands   │  │
│  │  • MCP stdio server for agent sessions                   │  │
│  └──────────────────────┬──────────────────────────────────┘  │
└─────────────────────────┼─────────────────────────────────────┘
                          │ pkg/store (Ent)
┌─────────────────────────▼─────────────────────────────────────┐
│                 Dolt (MySQL-compatible, Git-like)              │
└───────────────────────────────────────────────────────────────┘
```

See [Architecture Overview](architecture/overview.md) for the full picture, including the type pipeline and the legacy daemon/Electron architecture it's replacing.

!!! note "Legacy: Electron desktop app"
    An earlier surface — the `cmd/daemon` REST server plus an Electron desktop frontend — covers per-project spec authoring directly on the filesystem (no database): AIDLC workflow, V2MOM cascades, capability stacks, and organization/team settings. It's superseded by the `visionstudio` CLI and web dashboard above and isn't under active feature development. See [Go Daemon](architecture/daemon.md), [Frontend](architecture/frontend.md), and the [legacy user guide](guide/projects.md).

## Documentation

- [Installation](getting-started/installation.md) - System requirements and setup
- [Quick Start](getting-started/quickstart.md) - Get started in minutes
- [Dashboard Guide](dashboard/overview.md) - Complete feature documentation for the web dashboard
- [Architecture](architecture/overview.md) - Technical deep dive
- [Releases](releases/unreleased.md) - Release notes and changelog

## Related Projects

- [VisionSpec](https://github.com/ProductBuildersHQ/visionspec) - Spec orchestration library
- [AIDLC Framework](https://github.com/ProductBuildersHQ/productbuildershq-frameworks) - AI-Driven Development Lifecycle
