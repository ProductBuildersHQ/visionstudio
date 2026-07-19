# VisionStudio Implementation Plan

## Current Focus: Organization-Level V2MOMs

### Problem Statement

Currently, V2MOMs are project-scoped only. Organizations need a way to define company-wide strategic V2MOMs that cascade down to individual projects, enabling:

- Top-down strategic alignment
- Method inheritance from org to project level
- Alignment scoring between organizational and project goals
- Cross-project visibility of strategic initiatives

### Architecture

```
Organization (Workspace)
├── Organization V2MOM (Company Strategy)
│   ├── Vision: "Be the leader in..."
│   ├── Methods: ["M1: Expand market", "M2: Improve retention"]
│   └── Measures: ["Revenue $X", "NPS > Y"]
│
├── Project A
│   └── Project V2MOM
│       ├── Aligned to: Org Method M1
│       └── Methods: ["Launch in EU", "Partner program"]
│
└── Project B
    └── Project V2MOM
        ├── Aligned to: Org Method M2
        └── Methods: ["Reduce churn", "Improve onboarding"]
```

### Implementation Phases

#### Phase 1: Backend - Organization Entity & Storage

**Files to create/modify:**

1. `visionapp/pkg/config/organization.go` (new)
   - `Organization` struct with ID, name, path
   - `LoadOrganization()`, `SaveOrganization()` functions
   - Storage at `~/.visionstudio/organization.json`

2. `visionapp/pkg/api/types.go`
   - Add `Organization` type
   - Add `OrganizationV2MOM` type with scope field
   - Add request/response types for org endpoints

3. `visionapp/cmd/daemon/organization.go` (new)
   - `GET /api/organization` - Get current organization
   - `PUT /api/organization` - Update organization settings
   - `GET /api/organization/v2moms` - List org-level V2MOMs
   - `GET /api/organization/v2moms/{id}` - Get specific org V2MOM
   - `POST /api/organization/v2moms` - Create org V2MOM
   - `PUT /api/organization/v2moms/{id}` - Update org V2MOM

4. `visionapp/cmd/daemon/main.go`
   - Register new organization routes

#### Phase 2: Backend - Project Alignment

**Files to modify:**

1. `visionapp/pkg/config/projects.go`
   - Add `OrganizationID` field to `TrackedProject`

2. `visionapp/pkg/api/types.go`
   - Add `V2MOMAlignment` type for tracking method alignment
   - Add alignment score calculation types

3. `visionapp/cmd/daemon/v2mom.go`
   - Add `GET /api/projects/{project}/v2moms/{id}/alignment` endpoint
   - Implement alignment scoring logic

#### Phase 3: Frontend - Organization View

**Files to create:**

1. `components/organization/OrganizationSettings.tsx`
   - Organization name/settings editor
   - V2MOM file path configuration

2. `components/organization/OrgV2MOMView.tsx`
   - Display organization-level V2MOMs
   - Show cascade to projects

3. `components/organization/AlignmentMatrix.tsx`
   - Matrix showing org methods vs project methods
   - Alignment scores and gaps

**Files to modify:**

1. `App.tsx`
   - Add organization state
   - Add 'organization' to ActiveView type
   - Add organization view routing

2. `components/layout/Sidebar.tsx`
   - Add "Organization" section above projects
   - Show org V2MOM summary
   - Quick access to org settings

3. `services/api.ts`
   - Add organization API methods

4. `types/index.ts`
   - Add Organization types

#### Phase 4: V2MOM Cascade Visualization

**Files to create:**

1. `components/v2mom/OrgCascadeView.tsx`
   - Full cascade from org → projects
   - Interactive drill-down
   - Alignment indicators

2. `components/v2mom/AlignmentEditor.tsx`
   - UI for linking project methods to org methods
   - Drag-and-drop alignment

### Data Model

```typescript
// Organization
interface Organization {
  id: string
  name: string
  v2momPath?: string  // Path to org V2MOMs directory
  createdAt: string
  updatedAt: string
}

// V2MOM with scope
interface V2MOM {
  id: string
  name: string
  scope: 'organization' | 'project'
  parentId?: string           // For hierarchy within scope
  alignedToOrgMethodId?: string  // Project V2MOMs align to org methods
  // ... existing V2MOM fields
}

// Alignment tracking
interface V2MOMAlignment {
  projectV2momId: string
  projectMethodId: string
  orgMethodId: string
  alignmentScore: number  // 0-100
  notes?: string
}
```

### API Endpoints

```
# Organization
GET    /api/organization                    # Get org settings
PUT    /api/organization                    # Update org settings

# Organization V2MOMs
GET    /api/organization/v2moms             # List org V2MOMs
GET    /api/organization/v2moms/{id}        # Get org V2MOM
POST   /api/organization/v2moms             # Create org V2MOM
PUT    /api/organization/v2moms/{id}        # Update org V2MOM
DELETE /api/organization/v2moms/{id}        # Delete org V2MOM

# Cascade & Alignment
GET    /api/organization/cascade            # Full org → projects cascade
GET    /api/projects/{p}/v2moms/{id}/alignment  # Get alignment for project V2MOM
PUT    /api/projects/{p}/v2moms/{id}/alignment  # Update alignment
```

### UI Mockup - Sidebar

```
VisionStudio
─────────────────

ORGANIZATION
├── Strategy (Org V2MOM)    ← Click to view/edit
├── Alignment Dashboard     ← Cross-project view
└── Settings

─────────────────

PROJECTS
├── project-alpha ▼
│   ├── WHAT: AWS WB Product
│   ├── HOW: AIDLC
│   ├── Workflow
│   ├── V2MOM Cascade       ← Shows alignment to org
│   └── [specs...]
│
└── project-beta ▼
    └── ...
```

### Success Criteria

1. Users can create organization-level V2MOMs
2. Projects can be linked to organization
3. Project V2MOMs can align to org methods
4. Cascade view shows org → project hierarchy
5. Alignment scores calculated and displayed
6. Gaps in alignment are highlighted

### Dependencies

- Existing V2MOM infrastructure (complete)
- Methodology selection feature (complete)
- Project configuration system (complete)

### Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Complex hierarchy logic | Start with 2 levels (org → project), expand later |
| Data migration | Org features are additive, no migration needed |
| Performance with many projects | Lazy-load project V2MOMs in cascade view |

---

## Next: AIDLC Enhancements

(To be detailed after Organization V2MOM implementation)

## Later: SpecKit Implementation

(To be detailed after AIDLC enhancements)
