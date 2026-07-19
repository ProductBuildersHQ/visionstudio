# VisionStudio Samples Integration - Product Requirements Document

## Problem Statement

VisionStudio users need a way to:

1. **Learn the PRISM ecosystem** - Understand V2MOMs, capability stacks, and maturity models through working examples
2. **Quickly bootstrap projects** - Initialize new projects from reference samples instead of starting from scratch
3. **Visualize hierarchical V2MOMs** - View company and team-level V2MOMs in a cascade view
4. **Track capability maturity** - See capability stacks linked to maturity models with progress tracking
5. **Manage product roadmaps** - View roadmap items linked to capabilities and strategic goals

Currently, the Grafana and Simple samples exist but VisionStudio cannot load or display them.

## Target Users

- **Product Managers**: Define V2MOMs, prioritize roadmap items, track maturity progress
- **Engineering Leads**: Understand capability landscape, identify gaps, plan improvements
- **Executives**: View high-level maturity dashboards, track strategic goal progress
- **New Users**: Learn PRISM concepts through interactive samples

## User Stories

### Sample Discovery & Loading

- **US-1**: As a new user, I want to see available sample projects so that I can learn from working examples
- **US-2**: As a user, I want to initialize a project from a sample so that I don't start from scratch
- **US-3**: As a user, I want sample metadata (description, complexity) so that I can choose the right one

### V2MOM Integration

- **US-4**: As a PM, I want to view hierarchical V2MOMs so that I can see how team goals align with company vision
- **US-5**: As a PM, I want to see V2MOM cascade (company → teams) so that I understand goal relationships
- **US-6**: As a PM, I want V2MOM methods linked to roadmap items so that I can track execution

### Capability Stack Integration

- **US-7**: As an engineering lead, I want to view capability stacks so that I understand our product landscape
- **US-8**: As a user, I want capabilities linked to maturity levels so that I see improvement opportunities
- **US-9**: As a user, I want to see capability dependencies so that I understand architectural relationships

### Roadmap Integration

- **US-10**: As a PM, I want to view roadmap items so that I can track planned work
- **US-11**: As a PM, I want roadmap items linked to capabilities so that I understand technical impact
- **US-12**: As a PM, I want RICE scores displayed so that I can prioritize effectively

## Acceptance Criteria

### Sample Discovery

- [ ] `/api/samples` endpoint lists available samples with metadata
- [ ] Sample metadata includes: name, description, complexity level, file counts
- [ ] UI shows sample cards in project creation dialog
- [ ] "Initialize from Sample" copies sample files to new project

### V2MOM Display

- [ ] V2MOM cascade view shows hierarchy (top-level + teams)
- [ ] Each V2MOM displays: Vision, Values, Methods, Obstacles, Measures
- [ ] Team V2MOMs link to parent via `parentId`
- [ ] Click on team expands/collapses details

### Capability Stack Display

- [ ] Capability stack visualization shows layers and capabilities
- [ ] Each capability shows: name, status, owner, dependencies
- [ ] `prismRef` links capabilities to maturity domains
- [ ] Click on capability shows detail panel with maturity criteria

### Roadmap Display

- [ ] Roadmap view shows items grouped by quarter
- [ ] Each item displays: title, status, priority, RICE score
- [ ] Items link to capabilities and V2MOM goals
- [ ] Filtering by status, priority, quarter

## Sample Reference Implementations

### Simple Sample

| Component | Files | Purpose |
|-----------|-------|---------|
| V2MOM | 1 | Single V2MOM for onboarding |
| Capabilities | 1 | 8 capabilities, 3 layers |
| Maturity | 2 | 3 domains, 4 SLIs |
| Roadmap | 1 | 4 items |

### Grafana Reference

| Component | Files | Purpose |
|-----------|-------|---------|
| V2MOM | 6 | 1 top-level + 5 team V2MOMs |
| Capabilities | 6 | 1 top-level + 5 domain stacks |
| Maturity | 12 | 1 top-level + 5 domain models + states |
| Roadmap | 1 | 9 items |

## Out of Scope (V1)

- Editing V2MOMs in UI (read-only first)
- Creating new samples from existing projects
- Sample versioning and updates
- Cloud sync of samples

## Success Metrics

| Metric | Target |
|--------|--------|
| Sample adoption rate | 50% of new projects |
| V2MOM view usage | 30% of sessions |
| Capability view usage | 25% of sessions |
| Time to first value | -50% reduction |

## Dependencies

- prism-maturity schemas
- prism-capability schemas
- prism-roadmap schemas
- Grafana and Simple sample files (complete)
