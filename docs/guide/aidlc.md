# AIDLC Workflow

The AIDLC (AI-Driven Development Lifecycle) workflow provides a structured approach to AI-assisted software development with three phases: Inception, Construction, and Operations.

## Overview

When you select AIDLC as your implementation methodology, VisionStudio provides dedicated views for managing the full development lifecycle.

## Phases

### Inception

The discovery and planning phase that establishes project foundations.

**Deliverables:**

| Document | Required | Description |
|----------|----------|-------------|
| Vision Document | Yes | Strategic vision and product direction |
| Requirements Spec | Yes | Functional and non-functional requirements |
| Technical Spec | Yes | APIs, data models, system interfaces |
| Architecture Spec | No | Component design, deployment topology |

### Construction

The implementation planning phase that prepares for development.

**Deliverables:**

| Document | Required | Description |
|----------|----------|-------------|
| Implementation Plan | Yes | Milestones, tasks, resource allocation |
| Test Plan | Yes | Test strategy, cases, automation |
| Integration Plan | No | External system connections |
| Security Review | Yes | Threat modeling, access control |

### Operations

The production readiness phase for deployment and monitoring.

**Deliverables:**

| Document | Required | Description |
|----------|----------|-------------|
| Runbook | Yes | Deployment procedures, troubleshooting |
| Monitoring Plan | Yes | Metrics, alerts, dashboards |
| Disaster Recovery | No | RTO/RPO, backup, failover |
| SLO Document | Yes | SLIs, SLOs, error budgets |

## Using the AIDLC Workflow

### Accessing AIDLC Views

1. Select a project with AIDLC as implementation methodology
2. Click **AIDLC Workflow** in the sidebar
3. View phase status and deliverable completion

### Generating Documents

1. Navigate to a specific phase
2. Click on a deliverable to open it
3. Use **Generate** to create initial content
4. Review and edit the generated document
5. Run evaluation to assess quality

### Phase Transitions

1. Complete all required deliverables in current phase
2. Pass phase gate review criteria
3. Click **Transition to Next Phase**
4. The workflow advances and unlocks next phase deliverables

### Sync Status

The **AIDLC Sync** panel shows synchronization between:

- AIDLC documents and VisionSpec specs
- Phase completion status
- Dependency satisfaction

## Document Dependencies

AIDLC enforces document dependencies:

```
Vision Document
    └── Requirements Spec
            ├── Technical Spec
            │       └── Implementation Plan
            │               ├── Test Plan
            │               └── Security Review
            └── Architecture Spec
                    └── Implementation Plan
```

## Evaluation

Each document is evaluated against AIDLC criteria:

- **Completeness** - All required sections present
- **Clarity** - Unambiguous and actionable
- **Traceability** - Links to upstream documents
- **Feasibility** - Realistic and achievable

Passing threshold: 3.5/5.0 weighted score

## Cost Estimates

VisionStudio displays estimated LLM costs per document:

| Phase | Estimated Cost |
|-------|---------------|
| Inception | ~$0.51 |
| Construction | ~$0.70 |
| Operations | ~$0.60 |
| **Total** | **~$1.81** |

## Related

- [Methodology Selection](methodologies.md)
- [Evaluation](evaluation.md)
- [AIDLC API Reference](aidlc-api.md)
