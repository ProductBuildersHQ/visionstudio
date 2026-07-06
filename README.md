# VisionStudio

[![](docs/images/img_visionspec-visionstudio_hero_v4.png)](https://productbuildershq.com/visionstudio/)

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/ProductBuildersHQ/visionstudio/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/ProductBuildersHQ/visionstudio/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/ProductBuildersHQ/visionstudio/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/ProductBuildersHQ/visionstudio/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/ProductBuildersHQ/visionstudio/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/ProductBuildersHQ/visionstudio/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/ProductBuildersHQ/visionstudio
 [docs-godoc-url]: https://pkg.go.dev/github.com/ProductBuildersHQ/visionstudio
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://productbuildershq.com/visionstudio
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=ProductBuildersHQ%2Fvisionstudio
 [loc-svg]: https://tokei.rs/b1/github/ProductBuildersHQ/visionstudio
 [repo-url]: https://github.com/ProductBuildersHQ/visionstudio
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/ProductBuildersHQ/visionstudio/blob/main/LICENSE

LLM-powered desktop application for specification authoring and evaluation.

## Overview

VisionStudio provides an integrated workspace for creating, evaluating, and iterating on product specifications using the VisionSpec methodology. It combines structured spec workflows with AI-assisted writing and LLM-as-a-Judge evaluation.

## Features

- 📁 **Project Management** - Create and manage multiple spec projects
- ⚡ **Profile-Driven Workflows** - Select from profiles (aws-product, big-tech-product, etc.)
- 📊 **Visual Workflow Diagram** - See spec sequence and status at a glance
- ✏️ **Markdown Editor** - Source and rendered view toggle
- ⚖️ **LLM-as-a-Judge Evaluation** - Evaluate specs against profile rubrics
- 🤖 **LLM Writing Assistant** - Context-aware chat for spec assistance

## Screen Shots

### Workflow

[![](docs/images/ss_visionstudio_series-1_img-001_workflow.png)](https://productbuildershq.com/visionstudio/)

### PRD Spec

[![](docs/images/ss_visionstudio_series-1_img-002_prd.png)](https://productbuildershq.com/visionstudio/)

### Findings List

[![](docs/images/ss_visionstudio_series-1_img-003_findings.png)](https://productbuildershq.com/visionstudio/)

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

## Development

### Prerequisites

- Go 1.23+
- Node.js 20+
- npm

### Setup

```bash
# Clone the repository
git clone https://github.com/ProductBuildersHQ/visionstudio.git
cd visionstudio

# Build the Go daemon
go build -o bin/daemon ./cmd/daemon/

# Install frontend dependencies
cd desktop && npm install

# Start development servers
./bin/daemon &                    # Start Go daemon
cd desktop && npm run dev:renderer &  # Start Vite
cd desktop && npm run dev:main    # Start Electron
```

### Project Structure

```
visionstudio/
├── cmd/daemon/          # Go daemon entry point
├── pkg/
│   └── api/             # API types and handlers
├── desktop/
│   ├── main/            # Electron main process (TypeScript)
│   ├── renderer/        # React frontend
│   │   └── src/
│   │       ├── components/
│   │       ├── services/
│   │       └── types/
│   └── package.json
├── docs/
│   └── specs/           # VisionSpec project specs
└── go.mod
```

## Related Projects

- [VisionSpec](https://github.com/ProductBuildersHQ/visionspec) - Spec orchestration library
- [OmniAgent](https://github.com/plexusone/omniagent) - LLM agent interface

## License

MIT
