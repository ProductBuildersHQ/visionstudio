# VisionStudio Samples Integration - Technical Requirements Document

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  VisionStudio Desktop (Electron/React)                          │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ Frontend Components                                         │ │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐│ │
│  │  │V2MOMCascade  │ │CapabilityStack│ │RoadmapView         ││ │
│  │  │View (exists) │ │View (new)     │ │(new)               ││ │
│  │  └──────────────┘ └──────────────┘ └──────────────────────┘│ │
│  │  ┌──────────────┐ ┌──────────────┐                         │ │
│  │  │SamplePicker  │ │MaturityModel │                         │ │
│  │  │(new)         │ │View (exists) │                         │ │
│  │  └──────────────┘ └──────────────┘                         │ │
│  └────────────────────────────────────────────────────────────┘ │
│                              │ HTTP                             │
└──────────────────────────────┼──────────────────────────────────┘
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│  Go Daemon (:8765)                                               │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ API Routes                                                  │  │
│  │  /api/samples              - List available samples         │  │
│  │  /api/samples/{id}         - Get sample details             │  │
│  │  /api/projects/{p}/v2moms  - List/get V2MOMs (NEW)         │  │
│  │  /api/projects/{p}/capabilities - List/get capabilities    │  │
│  │  /api/projects/{p}/roadmap - Get roadmap items (NEW)       │  │
│  │  /api/projects/{p}/maturity/* - Maturity (EXISTS)          │  │
│  └────────────────────────────────────────────────────────────┘  │
│                              │                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ File Loaders                                                │  │
│  │  v2mom/      - Load V2MOM JSON files                        │  │
│  │  capability/ - Load capability stack JSON files             │  │
│  │  maturity/   - Load maturity model/state (EXISTS)           │  │
│  │  roadmap/    - Load roadmap JSON files                      │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│  File System                                                     │
│  ├── samples/                                                    │
│  │   ├── simple/    - Onboarding sample                         │
│  │   └── grafana/   - Enterprise reference                      │
│  └── {project}/                                                  │
│      ├── v2mom/                                                  │
│      ├── capability/                                             │
│      ├── maturity/                                               │
│      └── roadmap/                                                │
└──────────────────────────────────────────────────────────────────┘
```

## API Specifications

### 1. Sample Discovery

**GET /api/samples**

```go
type SampleSummary struct {
    ID          string            `json:"id"`          // "simple", "grafana"
    Name        string            `json:"name"`        // "Simple Sample"
    Description string            `json:"description"` // "Onboarding example..."
    Complexity  string            `json:"complexity"`  // "simple", "enterprise"
    Path        string            `json:"path"`        // "/samples/simple"
    FileCounts  map[string]int    `json:"fileCounts"`  // {"v2mom": 1, "capability": 1}
}

type ListSamplesResponse struct {
    Samples []SampleSummary `json:"samples"`
}
```

**GET /api/samples/{sampleId}**

```go
type SampleDetail struct {
    SampleSummary
    ProjectJSON json.RawMessage `json:"projectJson"` // Full project.json content
    README      string          `json:"readme"`      // README.md content
}
```

### 2. V2MOM Endpoints

**GET /api/projects/{project}/v2moms**

```go
type V2MOMSummary struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    ParentID string `json:"parentId,omitempty"`
    Path     string `json:"path"`
}

type ListV2MOMsResponse struct {
    V2MOMs []V2MOMSummary `json:"v2moms"`
}
```

**GET /api/projects/{project}/v2moms/{v2momId}**

Returns full V2MOM JSON per prism-roadmap/schema/v2mom.schema.json

**GET /api/projects/{project}/v2moms/cascade**

```go
type V2MOMCascade struct {
    Root     V2MOM       `json:"root"`
    Children []V2MOM     `json:"children"`
}
```

### 3. Capability Stack Endpoints

**GET /api/projects/{project}/capabilities**

```go
type CapabilitySummary struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Domain      string `json:"domain,omitempty"`
    Path        string `json:"path"`
}

type ListCapabilitiesResponse struct {
    Capabilities []CapabilitySummary `json:"capabilities"`
}
```

**GET /api/projects/{project}/capabilities/{capabilityId}**

Returns full capability stack JSON per prism-capability/schema/capability-stack.schema.json

### 4. Roadmap Endpoint

**GET /api/projects/{project}/roadmap**

```go
type RoadmapResponse struct {
    Metadata json.RawMessage  `json:"metadata"`
    Items    []RoadmapItem    `json:"items"`
}

type RoadmapItem struct {
    ID            string   `json:"id"`
    Title         string   `json:"title"`
    Description   string   `json:"description"`
    Status        string   `json:"status"`
    Priority      string   `json:"priority"`
    Quarter       string   `json:"quarter"`
    Effort        string   `json:"effort"`
    RICE          *RICE    `json:"rice,omitempty"`
    CapabilityRefs []string `json:"capability_refs,omitempty"`
    GoalRefs      []string `json:"goal_refs,omitempty"`
}

type RICE struct {
    Reach      int     `json:"reach"`
    Impact     int     `json:"impact"`
    Confidence float64 `json:"confidence"`
    Effort     int     `json:"effort"`
    Score      float64 `json:"score"`
}
```

## Frontend Components

### 1. SamplePicker (`components/samples/SamplePicker.tsx`)

```typescript
interface SamplePickerProps {
  onSelect: (sampleId: string) => void;
  onCancel: () => void;
}

// Displays grid of sample cards
// Shows: name, description, complexity badge, file counts
// "Use This Sample" button triggers project initialization
```

### 2. CapabilityStackView (`components/capability/CapabilityStackView.tsx`)

```typescript
interface CapabilityStackViewProps {
  projectName: string;
  capabilityId?: string; // Optional: specific capability to show
}

// Visualizes capability stack with:
// - Layer sections (vertical)
// - Capability cards with status indicators
// - Dependency lines between capabilities
// - Click to expand with prismRef details
```

### 3. RoadmapView (`components/roadmap/RoadmapView.tsx`)

```typescript
interface RoadmapViewProps {
  projectName: string;
}

// Displays roadmap items:
// - Grouped by quarter (swimlanes)
// - Card with status, priority, RICE score
// - Links to capabilities and V2MOM goals
// - Filter/sort controls
```

### 4. V2MOMCascadeView Updates

Existing component needs:
- Data binding to `/api/projects/{project}/v2moms/cascade`
- Expand/collapse for child V2MOMs
- Visualization of parent-child relationships

## File Discovery Logic

```go
// project.json-aware discovery
func (s *Server) discoverProjectFiles(projectPath string) (*ProjectFiles, error) {
    // 1. Check for project.json
    projectJSONPath := filepath.Join(projectPath, "project.json")
    if exists(projectJSONPath) {
        return parseProjectJSON(projectJSONPath)
    }

    // 2. Fallback: scan standard directories
    return &ProjectFiles{
        V2MOMs:       scanDir(projectPath, "v2mom", "*.v2mom.json"),
        Capabilities: scanDir(projectPath, "capability", "*.capability.json"),
        Maturity:     scanDir(projectPath, "maturity", "*.json"),
        Roadmap:      scanDir(projectPath, "roadmap", "*.roadmap.json"),
    }, nil
}
```

## Type Definitions

### Frontend Types (`types/prism.ts`)

```typescript
// V2MOM types (from prism-roadmap)
interface V2MOM {
  metadata: V2MOMMetadata;
  vision: string;
  values: Value[];
  methods: Method[];
  obstacles: Obstacle[];
  measures: Measure[];
}

// Capability types (from prism-capability)
interface CapabilityStack {
  metadata: CapabilityMetadata;
  layers: Layer[];
  categories: Category[];
  capabilities: Capability[];
  prismIntegration?: PrismIntegration;
}

// Roadmap types (from prism-roadmap)
interface Roadmap {
  metadata: RoadmapMetadata;
  items: RoadmapItem[];
}
```

## Security Considerations

1. **Path Validation**: All file paths must be within project directory
2. **JSON Parsing**: Use safe JSON unmarshaling with size limits
3. **Sample Isolation**: Samples are read-only, never modified
4. **No Code Execution**: JSON files are data only, no eval

## Performance Requirements

| Operation | Target Latency |
|-----------|----------------|
| List samples | < 100ms |
| Load V2MOM cascade | < 200ms |
| Load capability stack | < 200ms |
| Load roadmap | < 100ms |
| File change detection | < 500ms |

## Dependencies

### Go Packages

```go
// Already used
"encoding/json"
"net/http"
"path/filepath"
"github.com/go-chi/chi/v5"

// No new dependencies required
```

### NPM Packages

```json
{
  "dependencies": {
    // Existing - no new dependencies required
    "react": "^18.2.0",
    "tailwindcss": "^4.0.0"
  }
}
```
