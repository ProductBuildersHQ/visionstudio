# VisionStudio

LLM-powered desktop application for specification authoring and evaluation.

## What is VisionStudio?

VisionStudio provides an integrated workspace for creating, evaluating, and iterating on product specifications using the [VisionSpec](https://github.com/ProductBuildersHQ/visionspec) methodology.

## Key Features

- **Project Management** - Create and manage multiple spec projects with different profiles
- **Profile-Driven Workflows** - Select from methodologies like AWS Working Backwards, Big Tech, Shape Up
- **Visual Workflow Diagram** - See spec sequence, dependencies, and status at a glance
- **Markdown Editor** - Toggle between source and rendered views
- **LLM-as-a-Judge Evaluation** - Evaluate specs against profile-specific rubrics
- **LLM Writing Assistant** - Context-aware chat for spec assistance

## Quick Start

```bash
# Build and run
go build -o bin/daemon ./cmd/daemon/
cd desktop && npm install && npm run dev
```

See [Installation](getting-started/installation.md) for detailed setup instructions.

## Architecture

VisionStudio uses a desktop architecture with:

- **Electron** - Cross-platform desktop shell
- **React/TypeScript** - Frontend UI
- **Go Daemon** - Backend API server integrating with VisionSpec

```
Electron App
    │
    ▼ HTTP
Go Daemon ──► VisionSpec
    │
    ▼
Local Filesystem (specs as Markdown)
```
