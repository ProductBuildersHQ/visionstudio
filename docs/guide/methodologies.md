# Methodology Selection

VisionStudio supports dual methodology selection, separating **what to build** from **how to build**.

## Overview

When creating a project, you select two methodologies:

1. **Requirements Methodology** - Defines the spec workflow for capturing requirements
2. **Implementation Methodology** - Defines the development lifecycle approach

## Requirements Methodologies

Available requirements methodologies define which specs to create and in what order.

| Methodology | Description | Specs |
|-------------|-------------|-------|
| AWS Working Backwards (Product) | Amazon-style product development | Press Release, FAQ, MRD, PRD, UXD, TRD, IRD |
| AWS Working Backwards (Feature) | Feature-level working backwards | Press Release, FAQ, PRD, TRD |
| Big Tech Product | Combined best practices | MRD, PRD, UXD, TRD |
| Enterprise | Comprehensive post-PMF | Full spec suite |
| Lean Startup | MVP-focused | Problem, Solution, MVP |
| Jobs To Be Done | Outcome-focused | JTBD, Solution |
| Shape Up | Basecamp-style | Pitch, Bet |
| Design Thinking | Human-centered | Empathy, Define, Ideate, Prototype |

### Changing Requirements Methodology

1. Click the requirements methodology label in sidebar
2. Select new methodology from dropdown
3. Confirm change (existing specs preserved)
4. New workflow becomes active

## Implementation Methodologies

Available implementation methodologies define how development proceeds.

| Methodology | Description | Features |
|-------------|-------------|----------|
| AIDLC | AWS AI-Driven Development Lifecycle | 3 phases, 12 deliverables, LLM generation |
| SpecKit | GitHub SpecKit | Spec-driven development |
| None | No implementation methodology | Specs only |

### Changing Implementation Methodology

1. Click the implementation methodology label in sidebar
2. Select new methodology from dropdown
3. Confirm change
4. New views appear in sidebar

## Sidebar Menu Updates

The sidebar dynamically updates based on selected methodology:

**When AIDLC is selected:**

```
Project Name
├── Requirements: AWS Working Backwards
├── Implementation: AIDLC
├── ─────────────────
├── AIDLC Workflow
├── AIDLC Sync
├── ─────────────────
├── V2MOM Cascade
├── Capabilities
├── Roadmap
├── Maturity Model
├── ─────────────────
├── MRD ◐
├── PRD ○
└── ...
```

**When None is selected:**

```
Project Name
├── Requirements: AWS Working Backwards
├── Implementation: None
├── ─────────────────
├── V2MOM Cascade
├── Capabilities
├── Roadmap
├── Maturity Model
├── ─────────────────
├── MRD ◐
├── PRD ○
└── ...
```

## Project Configuration

Methodology selection is stored in project configuration:

```json
{
  "name": "my-project",
  "requirementsMethodology": "aws-working-backwards/product",
  "implementationMethodology": "aidlc"
}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/methodologies/requirements` | GET | List available requirements methodologies |
| `/api/methodologies/implementation` | GET | List available implementation methodologies |
| `/api/projects/{project}/methodology` | GET | Get project methodology |
| `/api/projects/{project}/methodology` | PUT | Update project methodology |

## Related

- [Projects](projects.md)
- [AIDLC Workflow](aidlc.md)
- [Workflows](workflows.md)
