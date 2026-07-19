# VisionStudio

LLM-powered desktop application for specification authoring, strategic planning, and product development lifecycle management.

## What is VisionStudio?

VisionStudio provides an integrated workspace for creating, evaluating, and iterating on product specifications using the [VisionSpec](https://github.com/ProductBuildersHQ/visionspec) methodology. It combines structured spec workflows with AI-assisted writing, LLM-as-a-Judge evaluation, and comprehensive strategic planning tools.

## Key Features

### Specification Authoring

- **Project Management** - Create and manage multiple spec projects
- **Dual Methodology Selection** - Choose requirements methodology (AWS Working Backwards, Big Tech, Lean Startup, etc.) and implementation methodology (AIDLC, SpecKit)
- **Visual Workflow Diagram** - See spec sequence and status at a glance
- **Markdown Editor** - Source and rendered view toggle with live preview
- **LLM-as-a-Judge Evaluation** - Evaluate specs against profile rubrics with dimension scores
- **LLM Writing Assistant** - Context-aware chat for spec assistance

### AIDLC Integration

- **AIDLC Workflow** - AWS AI-Driven Development Lifecycle (Inception → Construction → Operations)
- **Document Generation** - LLM-powered generation of AIDLC deliverables
- **Sync Status** - Track alignment between specs and AIDLC documents
- **Phase Transitions** - Guided progression through development phases

### Strategic Planning

- **V2MOM Cascade** - Hierarchical V2MOMs from Organization → Team → Project
- **Capability Stack** - Visual capability management with domains and levels
- **Roadmap View** - Timeline-based initiative and milestone planning
- **Maturity Model** - Framework-based maturity assessments with dashboards

### Organization Management

- **Organization Settings** - Configure company/department context
- **Team Management** - Define teams and hierarchies
- **Sample Projects** - Import example projects for learning

## Quick Start

```bash
# Build and run the daemon
go build -o bin/daemon ./cmd/daemon/
./bin/daemon

# In another terminal, run the desktop app
cd desktop && npm install && npm run dev
```

See [Installation](getting-started/installation.md) for detailed setup instructions.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Electron Desktop App                      │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              React/TypeScript Frontend                   ││
│  │                                                          ││
│  │  Layout:                                                 ││
│  │  • Sidebar (projects, methodology, navigation)          ││
│  │  • Main content area (views)                             ││
│  │                                                          ││
│  │  Views:                                                  ││
│  │  • Workflow diagram + spec editor                        ││
│  │  • AIDLC workflow + document generation                  ││
│  │  • V2MOM cascade editor                                  ││
│  │  • Capability stack view                                 ││
│  │  • Roadmap timeline                                      ││
│  │  • Maturity model dashboard                              ││
│  │  • Organization settings                                 ││
│  └──────────────────────┬──────────────────────────────────┘│
└─────────────────────────┼───────────────────────────────────┘
                          │ HTTP REST
┌─────────────────────────▼───────────────────────────────────┐
│                      Go Daemon                               │
│  • REST API (60+ endpoints)                                  │
│  • VisionSpec v0.13.0 integration                            │
│  • AIDLC workflow management                                 │
│  • V2MOM cascade handling                                    │
│  • Capability, roadmap, maturity model support               │
└─────────────────────────────────────────────────────────────┘
```

## Documentation

- [Installation](getting-started/installation.md) - System requirements and setup
- [Quick Start](getting-started/quickstart.md) - Get started in minutes
- [User Guide](guide/projects.md) - Complete feature documentation
- [Architecture](architecture/overview.md) - Technical deep dive

## Related Projects

- [VisionSpec](https://github.com/ProductBuildersHQ/visionspec) - Spec orchestration library
- [AIDLC Framework](https://github.com/ProductBuildersHQ/productbuildershq-frameworks) - AI-Driven Development Lifecycle
