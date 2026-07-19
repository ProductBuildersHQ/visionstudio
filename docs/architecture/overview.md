# Architecture Overview

VisionStudio uses a desktop architecture with Electron frontend and Go backend.

## Components

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
│  │                                                          ││
│  │  Services:                                               ││
│  │  • API client (all backend communication)                ││
│  │                                                          ││
│  └──────────────────────┬──────────────────────────────────┘│
└─────────────────────────┼───────────────────────────────────┘
                          │ HTTP REST
┌─────────────────────────▼───────────────────────────────────┐
│                      Go Daemon                               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Handlers:                                               ││
│  │  • main.go - Core routes (projects, specs, maturity)     ││
│  │  • aidlc.go - AIDLC workflow                             ││
│  │  • v2mom.go - V2MOM cascade                              ││
│  │  • capability.go - Capability stack                      ││
│  │  • roadmap.go - Roadmap management                       ││
│  │  • organization.go - Organization/teams                  ││
│  │  • methodologies.go - Methodology selection              ││
│  │  • samples.go - Sample projects                          ││
│  └──────────────────────┬──────────────────────────────────┘│
│  ┌──────────────────────▼──────────────────────────────────┐│
│  │  Integrations:                                           ││
│  │  • VisionSpec v0.13.0 (profiles, AIDLC, evaluation)      ││
│  │  • structured-evaluation (LLM-as-Judge)                  ││
│  │  • Filesystem (JSON/Markdown storage)                    ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Data Storage

VisionStudio stores data as files:

```
~/.visionspec/
├── config.json           # Global configuration
├── organization.json     # Organization and team settings
└── projects.json         # Tracked project list

project-directory/
├── .visionspec/          # Project-specific config
│   ├── project.json      # Project metadata
│   └── maturity/         # Maturity models
├── specs/                # Specification documents
│   ├── mrd.md
│   ├── prd.md
│   └── ...
├── aidlc-docs/           # AIDLC deliverables (if AIDLC selected)
│   ├── inception/
│   ├── construction/
│   └── operations/
├── v2mom/                # V2MOM documents
├── capability/           # Capability definitions
├── roadmap/              # Roadmap data
└── maturity/             # Maturity assessments
```

## Design Decisions

### Why Electron + Go Daemon?

- **Electron**: Mature, battle-tested, consistent rendering across platforms
- **Go Daemon**: Reuse VisionSpec library, efficient file operations
- **HTTP API**: Enables future web app with same backend

### Why Not Wails?

- Wails v3 is still alpha
- Electron ecosystem is more mature for polished UIs
- HTTP API enables web reuse

### Dual Methodology Architecture

Separating requirements and implementation methodologies:

- **Requirements Methodology**: Defines spec workflow (what to build)
- **Implementation Methodology**: Defines development lifecycle (how to build)

This allows mixing approaches (e.g., AWS Working Backwards + AIDLC).

### File-Based Storage

- Projects are portable (just directories)
- Git-friendly (all text-based)
- No database required
- Works offline

## Data Flow

### Spec Workflow

1. User selects project in sidebar
2. UI loads project details via API
3. User selects spec in workflow diagram
4. API returns spec content (Markdown)
5. User edits in editor
6. Save triggers API PUT
7. Daemon writes to filesystem

### AIDLC Workflow

1. User navigates to AIDLC Workflow view
2. API returns phase/deliverable status
3. User selects deliverable to generate
4. API calls VisionSpec for LLM generation
5. Generated content returned and saved
6. Evaluation run on content
7. Results displayed in UI

### V2MOM Cascade

1. User navigates to V2MOM Cascade
2. API returns organization → team → project hierarchy
3. User edits V2MOM at any level
4. Changes saved via API
5. Cascade relationships maintained

## Component Communication

```
┌──────────┐     HTTP      ┌──────────┐     File I/O    ┌──────────┐
│  React   │──────────────▶│    Go    │────────────────▶│   File   │
│  UI      │◀──────────────│  Daemon  │◀────────────────│  System  │
└──────────┘               └──────────┘                 └──────────┘
                                │
                                │ Import
                                ▼
                          ┌──────────┐
                          │VisionSpec│
                          │ Library  │
                          └──────────┘
```

## Security

- Daemon only binds to localhost (127.0.0.1)
- No authentication required (local desktop app)
- Path traversal protection on all file operations
- Input validation on all API endpoints
