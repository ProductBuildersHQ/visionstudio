# Architecture Overview

VisionStudio uses a desktop architecture with Electron frontend and Go backend.

## Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Electron Desktop App                      │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              React/TypeScript Frontend                   ││
│  │  • Components (Sidebar, Editor, Workflow, LLM Panel)    ││
│  │  • Services (API client)                                ││
│  │  • Types (Project, Spec, EvalResult)                    ││
│  └──────────────────────┬──────────────────────────────────┘│
└─────────────────────────┼───────────────────────────────────┘
                          │ HTTP REST
┌─────────────────────────▼───────────────────────────────────┐
│                      Go Daemon                               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  API Routes                                              ││
│  │  • GET  /api/projects                                   ││
│  │  • GET  /api/projects/{project}/specs/{type}            ││
│  │  • PUT  /api/projects/{project}/specs/{type}            ││
│  │  • POST /api/projects/{project}/specs/{type}/evaluate   ││
│  │  • POST /api/chat                                        ││
│  └──────────────────────┬──────────────────────────────────┘│
│  ┌──────────────────────▼──────────────────────────────────┐│
│  │  Integrations                                            ││
│  │  • VisionSpec (profiles, templates, rubrics)            ││
│  │  • OmniAgent (LLM providers)                            ││
│  │  • Filesystem (specs as Markdown)                       ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
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

## Data Flow

1. User interacts with React UI
2. UI calls API service (HTTP)
3. Go daemon processes request
4. Daemon interacts with VisionSpec/filesystem
5. Response returned to UI
6. UI updates state
