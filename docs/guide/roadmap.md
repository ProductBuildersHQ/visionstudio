# Roadmap

The Roadmap view provides timeline-based planning for initiatives and milestones.

## Overview

VisionStudio roadmaps help you:

- Plan initiatives across quarters
- Track milestone completion
- Link initiatives to V2MOM methods
- Visualize dependencies between initiatives

## Roadmap Structure

```
Roadmap
├── Q1 2024
│   ├── Initiative: API v2 Launch
│   │   ├── Milestone: Design Complete
│   │   ├── Milestone: Beta Release
│   │   └── Milestone: GA Release
│   └── Initiative: Security Hardening
└── Q2 2024
    └── Initiative: Platform Scaling
```

## Using the Roadmap

### Accessing Roadmap View

1. Click **Roadmap** in the sidebar
2. View initiatives on the timeline
3. Filter by quarter, team, or status

### Creating Initiatives

1. Click **Add Initiative**
2. Fill in initiative details:
   - Name and description
   - Start and end dates
   - Owner/team
   - Priority
3. Add milestones within the initiative

### Initiative Details

| Field | Description |
|-------|-------------|
| Name | Initiative name |
| Description | What the initiative delivers |
| Start Date | When work begins |
| End Date | Target completion |
| Status | planned, in_progress, completed, blocked |
| Priority | high, medium, low |
| Owner | Responsible team or person |

### Milestones

Each initiative contains milestones:

| Field | Description |
|-------|-------------|
| Name | Milestone name |
| Target Date | When milestone should complete |
| Status | pending, completed |
| Deliverables | What's delivered at this milestone |

## Roadmap Files

Roadmaps are stored as JSON:

```
project/
└── roadmap/
    └── project.roadmap.json
```

## Example Roadmap

```json
{
  "metadata": {
    "name": "Platform Roadmap 2024",
    "version": "1.0"
  },
  "initiatives": [
    {
      "id": "api-v2",
      "name": "API v2 Launch",
      "description": "Complete redesign of public API",
      "startDate": "2024-01-15",
      "endDate": "2024-03-31",
      "status": "in_progress",
      "priority": "high",
      "owner": "API Team",
      "milestones": [
        {
          "id": "design",
          "name": "Design Complete",
          "targetDate": "2024-02-01",
          "status": "completed"
        },
        {
          "id": "beta",
          "name": "Beta Release",
          "targetDate": "2024-03-01",
          "status": "pending"
        },
        {
          "id": "ga",
          "name": "GA Release",
          "targetDate": "2024-03-31",
          "status": "pending"
        }
      ],
      "v2momMethodId": "m1"
    }
  ]
}
```

## V2MOM Integration

Link initiatives to V2MOM methods:

1. Open an initiative
2. Click **Link to V2MOM Method**
3. Select the parent method
4. Initiative progress contributes to method completion

## Timeline View

The roadmap displays as a Gantt-style timeline:

- Horizontal bars show initiative duration
- Milestone diamonds mark key dates
- Color coding indicates status
- Dependencies shown as connecting lines

## Related

- [V2MOM Cascade](v2mom.md)
- [Capabilities](capabilities.md)
