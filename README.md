# VisionStudio

LLM-powered desktop application for specification authoring and evaluation.

## Overview

VisionStudio provides an integrated workspace for creating, evaluating, and iterating on product specifications using the VisionSpec methodology. It combines structured spec workflows with AI-assisted writing and LLM-as-a-Judge evaluation.

## Features

- **Project Management** - Create and manage multiple spec projects
- **Profile-Driven Workflows** - Select from profiles (aws-product, big-tech-product, etc.)
- **Visual Workflow Diagram** - See spec sequence and status at a glance
- **Markdown Editor** - Source and rendered view toggle
- **LLM-as-a-Judge Evaluation** - Evaluate specs against profile rubrics
- **LLM Writing Assistant** - Context-aware chat for spec assistance

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Electron Desktop App                      │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              React/TypeScript Frontend                   ││
│  │  • Sidebar (projects, specs)                            ││
│  │  • Workflow diagram                                      ││
│  │  • Markdown editor                                       ││
│  │  • LLM chat panel                                        ││
│  └──────────────────────┬──────────────────────────────────┘│
└─────────────────────────┼───────────────────────────────────┘
                          │ HTTP/WebSocket
┌─────────────────────────▼───────────────────────────────────┐
│                      Go Daemon                               │
│  • REST API for projects/specs                              │
│  • VisionSpec integration                                    │
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
