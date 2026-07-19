# VisionStudio Samples Integration - Implementation Plan

## Phase Overview

| Phase | Focus | Deliverables |
|-------|-------|--------------|
| 1 | Sample Discovery | List/load samples, project initialization |
| 2 | V2MOM Integration | V2MOM cascade API and UI updates |
| 3 | Capability Stack | New capability visualization component |
| 4 | Roadmap | New roadmap view component |
| 5 | Polish | Navigation, error handling, performance |

---

## Phase 1: Sample Discovery & Loading

**Goal**: Enable users to discover and initialize projects from samples.

### Backend (Go Daemon)

| File | Action | Description |
|------|--------|-------------|
| `cmd/daemon/main.go` | Modify | Add sample routes |
| `cmd/daemon/samples.go` | Create | Sample handlers |
| `pkg/api/types.go` | Modify | Add Sample types |

### New Handlers

```go
// cmd/daemon/samples.go
func (s *Server) handleListSamples(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetSample(w http.ResponseWriter, r *http.Request)
func (s *Server) handleInitFromSample(w http.ResponseWriter, r *http.Request)
```

### Routes

```
GET  /api/samples           → handleListSamples
GET  /api/samples/{id}      → handleGetSample
POST /api/projects          → handleAddProject (update to support sample init)
```

### Frontend

| File | Action | Description |
|------|--------|-------------|
| `services/api.ts` | Modify | Add sample API methods |
| `components/samples/SamplePicker.tsx` | Create | Sample selection UI |
| `components/samples/SampleCard.tsx` | Create | Individual sample card |
| `components/samples/index.ts` | Create | Exports |

### Verification

- [ ] `GET /api/samples` returns simple and grafana samples
- [ ] `GET /api/samples/grafana` returns full detail
- [ ] SamplePicker displays samples in project creation
- [ ] New project initialized from sample has all files

---

## Phase 2: V2MOM Integration

**Goal**: Connect V2MOM cascade view to backend data.

### Backend

| File | Action | Description |
|------|--------|-------------|
| `cmd/daemon/main.go` | Modify | Add v2mom routes |
| `cmd/daemon/v2mom.go` | Create | V2MOM handlers |
| `pkg/api/types.go` | Modify | Add V2MOM types |

### New Handlers

```go
// cmd/daemon/v2mom.go
func (s *Server) handleListV2MOMs(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetV2MOM(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetV2MOMCascade(w http.ResponseWriter, r *http.Request)
```

### Routes

```
GET /api/projects/{project}/v2moms           → handleListV2MOMs
GET /api/projects/{project}/v2moms/{id}      → handleGetV2MOM
GET /api/projects/{project}/v2moms/cascade   → handleGetV2MOMCascade
```

### Frontend

| File | Action | Description |
|------|--------|-------------|
| `services/api.ts` | Modify | Add v2mom API methods |
| `components/v2mom/V2MOMCascadeView.tsx` | Modify | Connect to API |
| `types/v2mom/types.ts` | Modify | Add missing types |

### Verification

- [ ] `GET /api/projects/{project}/v2moms` returns V2MOM list
- [ ] `GET /api/projects/{project}/v2moms/cascade` returns hierarchy
- [ ] V2MOMCascadeView displays data from API
- [ ] Team V2MOMs expand/collapse correctly

---

## Phase 3: Capability Stack

**Goal**: Add capability stack visualization.

### Backend

| File | Action | Description |
|------|--------|-------------|
| `cmd/daemon/main.go` | Modify | Add capability routes |
| `cmd/daemon/capability.go` | Create | Capability handlers |
| `pkg/api/types.go` | Modify | Add Capability types |

### New Handlers

```go
// cmd/daemon/capability.go
func (s *Server) handleListCapabilities(w http.ResponseWriter, r *http.Request)
func (s *Server) handleGetCapability(w http.ResponseWriter, r *http.Request)
```

### Routes

```
GET /api/projects/{project}/capabilities           → handleListCapabilities
GET /api/projects/{project}/capabilities/{id}      → handleGetCapability
```

### Frontend

| File | Action | Description |
|------|--------|-------------|
| `services/api.ts` | Modify | Add capability API methods |
| `components/capability/CapabilityStackView.tsx` | Create | Main view |
| `components/capability/CapabilityCard.tsx` | Create | Capability card |
| `components/capability/LayerSection.tsx` | Create | Layer grouping |
| `components/capability/types.ts` | Create | Type definitions |
| `components/capability/index.ts` | Create | Exports |
| `App.tsx` | Modify | Add capability view routing |
| `layout/Sidebar.tsx` | Modify | Add navigation item |

### Verification

- [ ] `GET /api/projects/{project}/capabilities` returns list
- [ ] CapabilityStackView renders layers and capabilities
- [ ] Click on capability shows detail panel
- [ ] prismRef links to maturity domain

---

## Phase 4: Roadmap Integration

**Goal**: Add roadmap visualization.

### Backend

| File | Action | Description |
|------|--------|-------------|
| `cmd/daemon/main.go` | Modify | Add roadmap route |
| `cmd/daemon/roadmap.go` | Create | Roadmap handlers |
| `pkg/api/types.go` | Modify | Add Roadmap types |

### New Handlers

```go
// cmd/daemon/roadmap.go
func (s *Server) handleGetRoadmap(w http.ResponseWriter, r *http.Request)
```

### Routes

```
GET /api/projects/{project}/roadmap → handleGetRoadmap
```

### Frontend

| File | Action | Description |
|------|--------|-------------|
| `services/api.ts` | Modify | Add roadmap API method |
| `components/roadmap/RoadmapView.tsx` | Create | Main view |
| `components/roadmap/RoadmapItem.tsx` | Create | Item card |
| `components/roadmap/QuarterLane.tsx` | Create | Quarter grouping |
| `components/roadmap/types.ts` | Create | Type definitions |
| `components/roadmap/index.ts` | Create | Exports |
| `App.tsx` | Modify | Add roadmap view routing |
| `layout/Sidebar.tsx` | Modify | Add navigation item |

### Verification

- [ ] `GET /api/projects/{project}/roadmap` returns items
- [ ] RoadmapView renders items grouped by quarter
- [ ] RICE scores display correctly
- [ ] Links to capabilities work

---

## Phase 5: Polish

**Goal**: Production-ready integration.

### Navigation Updates

| File | Changes |
|------|---------|
| `layout/Sidebar.tsx` | Order: Workflow > V2MOM > Capabilities > Roadmap > Maturity > Findings |
| `App.tsx` | Add view type constants |
| `types/index.ts` | Add ActiveView union type |

### Error Handling

- [ ] Missing files show helpful empty state
- [ ] Invalid JSON shows parse error
- [ ] API errors show toast notification

### Performance

- [ ] Lazy load views not in viewport
- [ ] Cache API responses (5 minute TTL)
- [ ] Debounce file system events

### Documentation

- [ ] Update README with samples usage
- [ ] Add inline code comments
- [ ] Update API documentation

---

## File Change Summary

### New Files (Backend - 4)

```
cmd/daemon/
├── samples.go      # Sample discovery handlers
├── v2mom.go        # V2MOM handlers
├── capability.go   # Capability handlers
└── roadmap.go      # Roadmap handlers
```

### New Files (Frontend - 12)

```
desktop/renderer/src/components/
├── samples/
│   ├── SamplePicker.tsx
│   ├── SampleCard.tsx
│   └── index.ts
├── capability/
│   ├── CapabilityStackView.tsx
│   ├── CapabilityCard.tsx
│   ├── LayerSection.tsx
│   ├── types.ts
│   └── index.ts
└── roadmap/
    ├── RoadmapView.tsx
    ├── RoadmapItem.tsx
    ├── QuarterLane.tsx
    ├── types.ts
    └── index.ts
```

### Modified Files (6)

```
cmd/daemon/main.go           # Add routes
pkg/api/types.go             # Add types
desktop/renderer/src/
├── services/api.ts          # Add API methods
├── App.tsx                  # Add views
├── components/layout/Sidebar.tsx  # Add nav items
└── components/v2mom/V2MOMCascadeView.tsx  # Connect to API
```

---

## Implementation Order

```
Phase 1: Sample Discovery
  ├── 1.1 Backend handlers (samples.go)
  ├── 1.2 API types
  ├── 1.3 Frontend SamplePicker
  └── 1.4 Project init from sample

Phase 2: V2MOM Integration
  ├── 2.1 Backend handlers (v2mom.go)
  ├── 2.2 API types
  ├── 2.3 Frontend API methods
  └── 2.4 V2MOMCascadeView data binding

Phase 3: Capability Stack (can parallel with 2)
  ├── 3.1 Backend handlers (capability.go)
  ├── 3.2 API types
  ├── 3.3 Frontend components
  └── 3.4 Navigation integration

Phase 4: Roadmap
  ├── 4.1 Backend handlers (roadmap.go)
  ├── 4.2 API types
  ├── 4.3 Frontend components
  └── 4.4 Navigation integration

Phase 5: Polish
  ├── 5.1 Error handling
  ├── 5.2 Performance optimization
  └── 5.3 Documentation
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Sample files missing | Validate sample integrity on startup |
| JSON schema drift | Pin to specific prism-* schema versions |
| Large file performance | Implement streaming/pagination |
| Navigation complexity | Progressive disclosure, collapsed by default |
