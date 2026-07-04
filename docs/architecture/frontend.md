# Frontend Architecture

The frontend is a React/TypeScript application running in Electron.

## Directory Structure

```
desktop/
├── main/                    # Electron main process
│   └── index.ts
├── renderer/                # React application
│   └── src/
│       ├── App.tsx          # Root component
│       ├── main.tsx         # Entry point
│       ├── index.css        # Tailwind styles
│       ├── components/
│       │   ├── layout/      # AppLayout, Sidebar
│       │   ├── project/     # WorkflowDiagram
│       │   ├── editor/      # SpecEditor
│       │   └── llm/         # LLMPanel
│       ├── services/
│       │   └── api.ts       # HTTP client
│       └── types/
│           └── index.ts     # TypeScript types
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## Components

### Layout

- **AppLayout** - 3-panel layout (sidebar, LLM, main)
- **Sidebar** - Project tree, profile selector, spec list

### Project

- **WorkflowDiagram** - Visual workflow with status indicators

### Editor

- **SpecEditor** - Markdown editor with source/rendered toggle

### LLM

- **LLMPanel** - Chat interface for LLM assistance

## State Management

State is managed in `App.tsx` using React hooks:

- `projects` - List of projects
- `activeProject` - Currently selected project
- `activeSpec` - Currently editing spec
- `specContent` - Editor content
- `isDirty` - Unsaved changes flag

## API Service

The `api.ts` service provides typed HTTP client:

```typescript
import { api } from './services/api'

const projects = await api.listProjects()
const spec = await api.getSpec('project', 'mrd')
await api.saveSpec('project', 'mrd', content)
```

## Styling

- **Tailwind CSS v4** - Utility-first CSS
- **Custom theme** - VisionStudio color palette (`va-*` colors)
