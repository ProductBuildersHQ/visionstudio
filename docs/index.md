# VisionStudio

LLM-powered desktop application for specification authoring and evaluation.

## What is VisionStudio?

VisionStudio provides an integrated workspace for creating, evaluating, and iterating on product specifications using the [VisionSpec](https://github.com/ProductBuildersHQ/visionspec) methodology. It combines structured spec workflows with AI-assisted writing and LLM-as-a-Judge evaluation.

## Key Features

- **Project Management** - Create and manage multiple spec projects
- **Profile-Driven Workflows** - Select from profiles (aws-product, big-tech-product, etc.)
- **Visual Workflow Diagram** - See spec sequence and status at a glance
- **Markdown Editor** - Source and rendered view toggle
- **LLM-as-a-Judge Evaluation** - Evaluate specs against profile rubrics
- **LLM Writing Assistant** - Context-aware chat for spec assistance

## Quick Start

```bash
# Build and run
go build -o bin/daemon ./cmd/daemon/
cd desktop && npm install && npm run dev
```

See [Installation](getting-started/installation.md) for detailed setup instructions.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Electron Desktop App                     │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              React/TypeScript Frontend                │  │
│  │  • Sidebar (projects, specs)                          │  │
│  │  • Workflow diagram                                   │  │
│  │  • Markdown editor                                    │  │
│  │  • LLM chat panel                                     │  │
│  └──────────────────────┬────────────────────────────────┘  │
└─────────────────────────┼───────────────────────────────────┘
                          │ HTTP/WebSocket
┌─────────────────────────▼───────────────────────────────────┐
│                      Go Daemon                              │
│  • REST API for projects/specs                              │
│  • VisionSpec integration                                   │
│  • LLM provider abstraction (omniagent)                     │
└─────────────────────────────────────────────────────────────┘
```

## Related Projects

- [VisionSpec](https://github.com/ProductBuildersHQ/visionspec) - Spec orchestration library
- [OmniAgent](https://github.com/plexusone/omniagent) - LLM agent interface
