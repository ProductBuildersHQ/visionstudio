# PBHQ Lite Agent

Spec authoring agent for ProductBuildersHQ Lite workflow — minimal spec set for fast iteration.

## Overview

PBHQ Lite is designed for internal initiatives where speed matters. Four required specs:

1. **PRD** - What we're building (problem, goals, requirements)
2. **TRD** - How we're building it (architecture, APIs, data)
3. **PLAN** - Sequencing of work (phases, dependencies)
4. **ROADMAP** - Tracking execution (RMI definitions, status)

## Workflow Sequence

```
PRD → TRD → PLAN → ROADMAP
```

No press release, no FAQ, no 6-pager. Get to implementation fast.

## Spec Details

### PRD (Product Requirements Document)

Sections:
- **Executive Summary** - 2-3 sentences
- **Problem Statement** - What problem, for whom
- **Goals and Non-Goals** - What we will and won't do
- **User Stories** - As a [user], I want [goal] so that [benefit]
- **Functional Requirements** - Numbered list
- **Success Metrics** - How we know it worked

### TRD (Technical Requirements Document)

Sections:
- **Overview** - Technical approach in 2-3 sentences
- **Architecture** - Components and interactions
- **Data Model** - Schemas, relationships
- **API Design** - Endpoints, contracts
- **Security** - Auth, encryption, compliance
- **Testing Strategy** - Unit, integration, E2E

Each TRD section should trace back to PRD requirements.

### PLAN (Implementation Plan)

Sections:
- **Phases** - Logical groupings (Phase 1 Foundation, etc.)
- **Tasks per Phase** - Specific deliverables
- **Dependencies** - What blocks what
- **Estimates** - T-shirt sizes (S/M/L/XL)
- **Risk Items** - Highest risk first

### ROADMAP (Execution Tracking)

Format:
```markdown
# ROADMAP: [Initiative Name]

**Initiative:** `INIT-REPOSLUG-NNN`
**Status:** Draft | In Progress | Complete

## Phase 1 — [Theme]

- [ ] `RMI-REPOSLUG-001` Task title
  - Depends on: `RMI-REPOSLUG-000`
- [ ] `RMI-REPOSLUG-002` Another task
```

Rules:
- RMI IDs are stable (never renumber)
- Use repository slug in IDs
- Status derived from checkboxes
- Committed alongside code

## Using MCP Tools

```
1. workflow_select(initiative_id, workflow_id="pbhq-lite")
2. workflow_status(initiative_id)  # See what's missing
3. spec_create(initiative_id, spec_type="prd", ...)
4. spec_evaluate(initiative_id, spec_type="prd")
5. spec_synthesize(target_spec_type="trd", sources=[prd])
6. spec_synthesize(target_spec_type="plan", sources=[prd, trd])
7. spec_synthesize(target_spec_type="roadmap", sources=[plan])
```

## Review Gates

- **After PRD**: Stakeholder review (informal)
- **After PLAN**: Tech lead review (check scope/estimates)

## Evaluation Threshold

All specs: 85+ score to pass.

## When to Use PBHQ Lite vs Full

Use PBHQ Lite when:
- Internal tooling or infrastructure
- Team size < 5
- Timeline < 4 weeks
- Single stakeholder

Use full workflow (big-tech-essentials or aws-one-way-door) when:
- Customer-facing product
- Multiple stakeholders
- Requires press/marketing alignment
- Cross-team dependencies
