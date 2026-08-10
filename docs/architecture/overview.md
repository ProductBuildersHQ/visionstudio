# Architecture Overview

VisionStudio has two architectures today:

- **Primary**: the `visionstudio` CLI and web dashboard, backed by Dolt.
- **Legacy** *(being phased out)*: an Electron desktop app talking to the `cmd/daemon` REST server, storing everything on the filesystem. See [Go Daemon](daemon.md) and [Frontend](frontend.md) for its details.

## Primary Architecture: visionstudio CLI + Web Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│                    visionstudio binary                       │
│  ┌─────────────────────────────────────────────────────────┐│
│  │        web/ — React + Vite SPA (go:embed all:dist)       ││
│  │  • Programs → Initiatives → Phases → RMIs                ││
│  │  • Repositories, Performance (token spend)                ││
│  │  • Spec viewer + LLM-as-a-Judge evaluations                ││
│  │  • Maturity assessments                                    ││
│  └──────────────────────┬──────────────────────────────────┘│
│                          │ HTTP (same port, cmd/visionstudio/serve.go)
│  ┌──────────────────────▼──────────────────────────────────┐│
│  │           cmd/visionstudio — cobra CLI + JSON API         ││
│  │  Commands:                                                 ││
│  │  • app/ui/db — lifecycle (one-command startup, standalone)││
│  │  • initiative/phase/rmi/program/registry/roadmap/spec/    ││
│  │    maturity/work/release — data management                ││
│  │  • ingest/export/validate/report — evidence & consistency ││
│  │  • mcp — stdio server for agent sessions                   ││
│  │  Handlers: cmd/visionstudio/api.go (JSON API, store→API    ││
│  │  converters)                                                ││
│  └──────────────────────┬──────────────────────────────────┘│
└─────────────────────────┼───────────────────────────────────┘
                          │ pkg/store (Ent)
┌─────────────────────────▼───────────────────────────────────┐
│                 Dolt (MySQL-compatible, Git-like)             │
└─────────────────────────────────────────────────────────────┘
```

### Data Storage: Dolt (via Ent)

All structured data lives in Dolt (MySQL-compatible, Git-like versioning), accessed through the Ent ORM (`pkg/store`):

- Programs, initiatives, phases, RMIs, and RMI/initiative dependencies
- Judge results and evaluations
- Spec workflows and templates
- Repository metadata
- Maturity models and assessments
- Token spend / cost events (ingested from OmniDevX)

```bash
# Initialize/migrate database
visionstudio db init --migrate
```

Spec Markdown files themselves remain on disk (`docs/specs/initiatives/<id>/*.md`, `evaluations/*.eval.json`) for Git compatibility — the database indexes and evaluates them, it doesn't replace them.

### Type Layers

| Layer | Package | JSON Style | Purpose |
|-------|---------|------------|---------|
| Store | `pkg/store` | snake_case | Database/internal |
| API | `pkg/apitypes` | camelCase | HTTP responses |
| Frontend | `web/src/api/types.gen.ts` | camelCase | TypeScript |

Conversion happens in API handlers (`cmd/visionstudio/api.go`). Go types are the source of truth — see [Type Pipeline](types.md) for the full generation flow.

### Security

- Binds to loopback by default; `visionstudio ui --address host:port` warns if you bind a non-loopback address
- No authentication (assumes trusted local/team network use)
- Request-derived path components (initiative IDs, etc.) are validated at the API boundary and re-checked immediately before every filesystem access, via `github.com/grokify/mogo/os/osutil`'s `ValidatePathComponent`/`JoinSecure`/`FindFirstExistingSecure`

## Legacy Architecture: Electron + Go Daemon

*This section documents the system being phased out — see the note at the top of this page.*

### Components

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
│  │  • DevX usage dashboard (not project-scoped)             ││
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
│  │  • devx.go - DevX dashboard passthrough                  ││
│  └──────────────────────┬──────────────────────────────────┘│
│  ┌──────────────────────▼──────────────────────────────────┐│
│  │  Integrations:                                           ││
│  │  • VisionSpec v0.14.0 (profiles, AIDLC, evaluation)      ││
│  │  • structured-evaluation (LLM-as-Judge)                  ││
│  │  • Filesystem (JSON/Markdown storage)                    ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Data Storage: Filesystem

No database — everything is JSON/Markdown on disk:

```
~/.visionspec/
├── config.json           # Global configuration
├── organization.json     # Organization and team settings
└── projects.json         # Tracked project list

project-directory/
├── .visionspec/          # Project-specific config
│   ├── project.json      # Project metadata
│   └── maturity/         # Maturity models
├── docs/specs/           # Specification documents
│   ├── PRD.md
│   ├── TRD.md
│   └── initiatives/
│       └── INIT-*/
│           ├── PLAN.md
│           ├── ROADMAP.md
│           └── evaluations/  # Judge results (*.eval.json)
├── aidlc-docs/           # AIDLC deliverables (if AIDLC selected)
├── v2mom/                # V2MOM documents
├── capability/           # Capability definitions
└── maturity/             # Maturity assessments
```

### Design Decisions

#### Why Electron + Go Daemon?

- **Electron**: Mature, battle-tested, consistent rendering across platforms
- **Go Daemon**: Reuse VisionSpec library, efficient file operations
- **HTTP API**: Enables future web app with same backend

#### Why Not Wails?

- Wails v3 is still alpha
- Electron ecosystem is more mature for polished UIs
- HTTP API enables web reuse

#### Dual Methodology Architecture

Separating requirements and implementation methodologies:

- **Requirements Methodology**: Defines spec workflow (what to build)
- **Implementation Methodology**: Defines development lifecycle (how to build)

This allows mixing approaches (e.g., AWS Working Backwards + AIDLC).

#### File-Based Storage

- Projects are portable (just directories)
- Git-friendly (all text-based)
- No database required
- Works offline

### Data Flow

#### Spec Workflow

1. User selects project in sidebar
2. UI loads project details via API
3. User selects spec in workflow diagram
4. API returns spec content (Markdown)
5. User edits in editor
6. Save triggers API PUT
7. Daemon writes to filesystem

#### AIDLC Workflow

1. User navigates to AIDLC Workflow view
2. API returns phase/deliverable status
3. User selects deliverable to generate
4. API calls VisionSpec for LLM generation
5. Generated content returned and saved
6. Evaluation run on content
7. Results displayed in UI

#### V2MOM Cascade

1. User navigates to V2MOM Cascade
2. API returns organization → team → project hierarchy
3. User edits V2MOM at any level
4. Changes saved via API
5. Cascade relationships maintained

### Component Communication

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

### Security

- Daemon only binds to localhost (127.0.0.1)
- No authentication required (local desktop app)
- Request-derived path components are validated at the API boundary and re-checked immediately before every filesystem access, via `github.com/grokify/mogo/os/osutil`'s `ValidatePathComponent`/`JoinSecure`/`FindFirstExistingSecure`
