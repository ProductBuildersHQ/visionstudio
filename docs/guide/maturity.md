# Maturity Model

The Maturity Model view provides framework-based assessments for tracking organizational or product maturity.

## Overview

VisionStudio supports maturity models that:

- Define dimensions and goals
- Track current vs target maturity
- Generate visual dashboards
- Link to capabilities and V2MOMs

## Maturity Model Structure

```
Maturity Model
├── Dimension: Developer Experience
│   ├── Goal: Self-service provisioning
│   ├── Goal: Documentation quality
│   └── Goal: SDK availability
├── Dimension: Reliability
│   ├── Goal: SLO coverage
│   ├── Goal: Incident response
│   └── Goal: Chaos engineering
└── Dimension: Security
    ├── Goal: Authentication
    ├── Goal: Authorization
    └── Goal: Audit logging
```

## Maturity Levels

Standard 5-level maturity scale:

| Level | Name | Description |
|-------|------|-------------|
| 1 | Reactive | Ad-hoc, no defined process |
| 2 | Managed | Basic processes in place |
| 3 | Defined | Standardized and documented |
| 4 | Measured | Metrics-driven improvement |
| 5 | Optimizing | Continuous optimization |

## Using Maturity Models

### Accessing Maturity View

1. Click **Maturity Model** in the sidebar
2. View dimensions and their scores
3. Drill into dimensions to see goals

### Creating a Maturity Model

1. Click **Create Model**
2. Define dimensions relevant to your domain
3. Add goals within each dimension
4. Set target maturity levels

### Assessing Maturity

1. Open a dimension
2. For each goal, set:
   - Current level (1-5)
   - Target level (1-5)
   - Evidence/notes
3. Save assessment
4. View rollup scores

### Dashboard

The maturity dashboard shows:

- Overall maturity score
- Dimension-by-dimension breakdown
- Gap analysis (current vs target)
- Trend over time (if historical data exists)

## Maturity Files

Maturity models use two files:

```
project/
└── maturity/
    ├── project.model.json    # Model definition
    └── project.state.json    # Current assessment
```

### Model Definition

```json
{
  "metadata": {
    "id": "platform-maturity",
    "name": "Platform Maturity Model"
  },
  "dimensions": [
    {
      "id": "dx",
      "name": "Developer Experience",
      "weight": 0.3,
      "goals": [
        {
          "id": "self-service",
          "name": "Self-service provisioning",
          "description": "Developers can provision without tickets"
        }
      ]
    }
  ]
}
```

### State Assessment

```json
{
  "metadata": {
    "modelId": "platform-maturity",
    "assessmentDate": "2024-01-15"
  },
  "dimensions": [
    {
      "dimensionId": "dx",
      "goals": [
        {
          "goalId": "self-service",
          "currentLevel": 2,
          "targetLevel": 4,
          "notes": "Basic CLI available, portal in development"
        }
      ]
    }
  ],
  "overallScore": 2.3
}
```

## HTML Dashboard

VisionStudio can generate an embedded HTML dashboard:

1. Open maturity model
2. Click **View Dashboard**
3. Interactive visualization opens
4. Export as standalone HTML if needed

## Integration

### With Capabilities

Link capabilities to maturity goals:

- Capability maturity contributes to goal score
- Gaps in capabilities highlight maturity gaps

### With V2MOM

Link maturity goals to V2MOM methods:

- Maturity improvement as V2MOM method
- Track progress through V2MOM measures

## Related

- [Capabilities](capabilities.md)
- [V2MOM Cascade](v2mom.md)
