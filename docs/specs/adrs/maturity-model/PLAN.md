# Maturity Model Integration: Implementation Plan

**Target:** Phase 1 - HTML Embedding
**Related Docs:**

- [ADR-001-embed-html-dashboard.md](./ADR-001-embed-html-dashboard.md)
- [DESIGN.md](./DESIGN.md)

## Implementation Checklist

### 1. Go Daemon API

- [ ] Add prism-maturity dependency to `go.mod` (deferred - using embedded HTML generator)
- [x] Create handlers in `cmd/daemon/main.go` (maturity handlers inline)
- [x] Implement `GET /api/projects/{project}/maturity/dashboard`
- [x] Implement `GET /api/projects/{project}/maturity/models`
- [x] Implement `GET /api/projects/{project}/maturity/models/{modelId}`
- [x] Register routes in `main.go`
- [x] Add `MaturityModelSummary` to `pkg/api/types.go`
- [ ] Test endpoint returns valid HTML

### 2. React Components

- [x] Create `desktop/renderer/src/components/maturity-model/` directory
- [x] Create `types.ts` with maturity type definitions
- [x] Create `MaturityModelView.tsx` with iframe
- [x] Create `index.ts` exports
- [x] Add exports to `components/index.ts`

### 3. API Service

- [x] Add `getMaturityDashboard()` to `services/api.ts`
- [x] Add `listMaturityModels()` to `services/api.ts`
- [x] Add `getMaturityModel()` to `services/api.ts`

### 4. App Integration

- [x] Add `'maturity-model'` to `ActiveView` type in `App.tsx`
- [x] Add `MaturityModelView` conditional rendering
- [x] Add `handleMaturityModelClick` handler

### 5. Sidebar Navigation

- [x] Add `onMaturityModelClick` prop to `Sidebar.tsx`
- [x] Add "Maturity Model" button (between V2MOM and Findings)
- [x] Wire up click handler

### 6. Testing

- [ ] Manual test: sidebar navigation
- [ ] Manual test: dashboard rendering
- [ ] Manual test: refresh functionality
- [ ] Manual test: error handling

## Dependencies

### prism-maturity

Required features from prism-maturity:

| Feature | Package | Function |
|---------|---------|----------|
| Dashboard generation | `dashboard` | `Dashboard.ToHTML()` |
| HTML template | `dashboard` | Embedded template |
| Maturity aggregation | `dashboard` | `MaturityAggregator` |

### VisionStudio

Existing patterns to follow:

| Pattern | Reference File |
|---------|---------------|
| View component | `v2mom/V2MOMCascadeView.tsx` |
| API service | `services/api.ts` |
| View switching | `App.tsx` |
| Sidebar nav | `layout/Sidebar.tsx` |

## File Changes Summary

| File | Action | Description |
|------|--------|-------------|
| `go.mod` | Modify | Add prism-maturity dependency |
| `cmd/daemon/main.go` | Modify | Add maturity routes |
| `cmd/daemon/maturity.go` | Create | Maturity handlers |
| `components/maturity-model/types.ts` | Create | Type definitions |
| `components/maturity-model/MaturityModelView.tsx` | Create | Main view component |
| `components/maturity-model/index.ts` | Create | Exports |
| `components/index.ts` | Modify | Add maturity exports |
| `services/api.ts` | Modify | Add maturity API methods |
| `App.tsx` | Modify | Add maturity view |
| `layout/Sidebar.tsx` | Modify | Add navigation button |

## Open Questions

1. **Data location:** Where should maturity model JSON files live?
   - Option A: `.visionspec/maturity/` (alongside other project config)
   - Option B: Separate `.prism/` directory
   - Option C: Embedded in main project config

2. **Multiple models:** Should we support multiple maturity models per project?
   - Current assumption: Yes, with model selector dropdown

3. **Edit capability:** Should VisionStudio allow editing maturity assessments?
   - Phase 1: Read-only dashboard
   - Phase 2: Edit current levels via UI

## Success Criteria

- [ ] User can navigate to Maturity Model view from sidebar
- [ ] Dashboard renders with all widgets (charts, tables, metrics)
- [ ] Dashboard matches prism-maturity standalone output
- [ ] Dark theme is applied consistently
- [ ] Refresh button reloads latest data
- [ ] Graceful error handling when no maturity data exists
