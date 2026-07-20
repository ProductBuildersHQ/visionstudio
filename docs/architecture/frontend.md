# Frontend Architecture

The frontend is a React/TypeScript application running in Electron.

## Directory Structure

```
desktop/
├── main/                    # Electron main process
│   └── index.ts
├── renderer/                # React application
│   └── src/
│       ├── App.tsx          # Root component, view routing
│       ├── main.tsx         # Entry point
│       ├── index.css        # Tailwind styles
│       ├── components/
│       │   ├── layout/      # Layout components
│       │   │   ├── Sidebar.tsx
│       │   │   ├── AddProjectModal.tsx
│       │   │   └── MethodologySelector.tsx
│       │   ├── project/     # Core project views
│       │   │   ├── WorkflowDiagram.tsx
│       │   │   └── FindingsView.tsx
│       │   ├── editor/      # Spec editing
│       │   │   └── SpecEditor.tsx
│       │   ├── aidlc/       # AIDLC workflow
│       │   │   ├── AIDLCWorkflowView.tsx
│       │   │   ├── AIDLCDocumentView.tsx
│       │   │   ├── AIDLCSyncPanel.tsx
│       │   │   └── ...
│       │   ├── v2mom/       # V2MOM cascade
│       │   │   ├── V2MOMCascadeView.tsx
│       │   │   └── V2MOMView.tsx
│       │   ├── capability-stack/
│       │   │   └── CapabilityStackView.tsx
│       │   ├── roadmap/
│       │   │   └── RoadmapView.tsx
│       │   ├── maturity-model/
│       │   │   └── MaturityModelView.tsx
│       │   ├── organization/
│       │   │   └── OrganizationView.tsx
│       │   ├── samples/
│       │   │   └── SamplePicker.tsx
│       │   ├── devx/        # DevX usage dashboard (not project-scoped)
│       │   │   └── DevXDashboardView.tsx
│       │   ├── eval/        # Evaluation display
│       │   │   └── DimensionScoreCard.tsx
│       │   └── llm/         # LLM assistance
│       │       └── LLMPanel.tsx
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

- **Sidebar** - Project tree, methodology display, navigation menu
- **AddProjectModal** - Create new project with methodology selection
- **MethodologySelector** - Change requirements/implementation methodology

### Project Views

- **WorkflowDiagram** - Visual spec workflow with status indicators
- **FindingsView** - Evaluation findings list

### Editor

- **SpecEditor** - Markdown editor with source/rendered toggle, evaluation display

### AIDLC

- **AIDLCWorkflowView** - Phase overview and deliverable status
- **AIDLCDocumentView** - Individual document editing/generation
- **AIDLCSyncPanel** - Sync status with VisionSpec
- **TransitionButton** - Phase transition control
- **PhaseRequirementsPanel** - Phase completion checklist
- **EvaluationResultsPanel** - Document evaluation display
- **TemplateSelector** - Template selection for generation

### V2MOM

- **V2MOMCascadeView** - Hierarchical V2MOM display
- **V2MOMView** - Individual V2MOM editor

### Strategic Planning

- **CapabilityStackView** - Visual capability management
- **RoadmapView** - Timeline-based initiative planning
- **MaturityModelView** - Maturity assessment dashboard

### Organization

- **OrganizationView** - Organization settings and team management

### Samples

- **SamplePicker** - Browse and import sample projects

### Evaluation

- **DimensionScoreCard** - Display evaluation dimension scores

### LLM

- **LLMPanel** - Chat interface for LLM assistance

## View Routing

Views are managed in `App.tsx` with an `ActiveView` type:

```typescript
type ActiveView =
  | 'workflow'
  | 'spec'
  | 'findings'
  | 'aidlc-workflow'
  | 'aidlc-sync'
  | 'v2mom-cascade'
  | 'capabilities'
  | 'roadmap'
  | 'maturity-model'
  | 'organization'
  | 'methodology-settings'
  | 'samples'
  | 'devx-dashboard'
```

## State Management

State is managed in `App.tsx` using React hooks:

```typescript
// Project state
const [projects, setProjects] = useState<Project[]>([])
const [activeProject, setActiveProject] = useState<Project | null>(null)
const [activeView, setActiveView] = useState<ActiveView>('workflow')

// Spec state
const [activeSpec, setActiveSpec] = useState<Spec | null>(null)
const [specContent, setSpecContent] = useState<string>('')
const [isDirty, setIsDirty] = useState(false)

// Modal state
const [showMethodologySelector, setShowMethodologySelector] = useState(false)
```

## API Service

The `api.ts` service provides typed HTTP client for all endpoints:

```typescript
import { api } from './services/api'

// Projects
const projects = await api.listProjects()
const project = await api.getProject('my-project')

// Specs
const spec = await api.getSpec('project', 'mrd')
await api.saveSpec('project', 'mrd', content)
const evalResult = await api.evaluateSpec('project', 'mrd')

// AIDLC
const workflow = await api.getAIDLCWorkflow('project')
const phases = await api.getAIDLCPhases('project')
await api.transitionAIDLCPhase('project', 'construction')

// V2MOM
const cascade = await api.getV2MOMCascade()
const v2mom = await api.getV2MOM('project', 'team-v2mom')

// Capabilities
const caps = await api.listCapabilities('project')

// Roadmap
const roadmap = await api.getRoadmap('project')

// Maturity
const models = await api.listMaturityModels('project')

// Methodology
await api.updateMethodology('project', {
  requirementsMethodology: 'aws-working-backwards/product',
  implementationMethodology: 'aidlc'
})

// Samples
const samples = await api.listSamples()
await api.importSample('grafana', '/path/to/destination')
```

## Types

Key TypeScript types in `types/index.ts`:

```typescript
interface Project {
  name: string
  path: string
  profile: Profile
  requirementsMethodology?: string
  implementationMethodology?: string
  specs: Spec[]
}

interface AIDLCWorkflow {
  currentPhase: string
  phases: AIDLCPhase[]
  completionPercentage: number
}

interface V2MOM {
  vision: string
  values: V2MOMValue[]
  methods: V2MOMMethod[]
  obstacles: V2MOMItem[]
  measures: V2MOMMeasure[]
}

interface CapabilityStack {
  metadata: Metadata
  capabilities: Capability[]
}

interface Roadmap {
  initiatives: Initiative[]
}

interface MaturityModel {
  dimensions: MaturityDimension[]
  overallScore: number
}
```

## Styling

- **Tailwind CSS v4** - Utility-first CSS
- **Custom theme** - VisionStudio color palette (`va-*` colors)
- **Dark mode** - Full dark mode support via Tailwind

## Conditional Rendering

Sidebar menu items are conditionally rendered based on methodology:

```tsx
{project.implementationMethodology === 'aidlc' && (
  <>
    <MenuItem icon="🔄" label="AIDLC Workflow" onClick={...} />
    <MenuItem icon="🔁" label="AIDLC Sync" onClick={...} />
  </>
)}
```
