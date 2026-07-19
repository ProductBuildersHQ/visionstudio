# V2MOM Cascade

V2MOM (Vision, Values, Methods, Obstacles, Measures) is a strategic planning framework that cascades from organization level down to individual projects.

## Overview

VisionStudio supports hierarchical V2MOMs:

```
Organization V2MOM
    └── Team V2MOMs
            └── Project V2MOMs
```

Each level inherits context from its parent while defining its own specific goals.

## V2MOM Components

### Vision

A clear statement of what you want to achieve. Should be inspiring and specific.

### Values

The principles and beliefs that guide how you work. Prioritized list that helps with decision-making.

### Methods

The actions and initiatives you'll take to achieve the vision. Each method should be actionable and measurable.

### Obstacles

Challenges and risks that could prevent success. Identifying obstacles early enables proactive mitigation.

### Measures

How you'll know if you've succeeded. Specific metrics with targets and current values.

## Using V2MOM Cascade

### Accessing V2MOM Views

1. Click **V2MOM Cascade** in the sidebar
2. View the hierarchical V2MOM structure
3. Expand nodes to see child V2MOMs

### Creating Organization V2MOM

1. Go to **Organization Settings**
2. Click **Create V2MOM**
3. Fill in Vision, Values, Methods, Obstacles, Measures
4. Save the V2MOM

### Creating Team V2MOMs

1. Navigate to the organization V2MOM
2. Click **Add Team V2MOM**
3. Select the team
4. Create V2MOM that aligns with organization V2MOM

### Creating Project V2MOMs

1. Open a project
2. Click **V2MOM** in the sidebar
3. Create project-specific V2MOM
4. Link methods to parent V2MOM methods

## Cascade Alignment

VisionStudio helps ensure alignment across the cascade:

- **Parent Context** - View parent V2MOM when editing child
- **Method Linking** - Connect child methods to parent methods
- **Rollup Views** - See how project work contributes to team/org goals

## V2MOM Files

V2MOMs are stored as JSON files:

```
project/
└── v2mom/
    ├── project.v2mom.json
    └── teams/
        ├── team-a.v2mom.json
        └── team-b.v2mom.json
```

## Example V2MOM

```json
{
  "metadata": {
    "id": "platform-2024",
    "name": "Platform Team 2024",
    "parentId": "org-2024"
  },
  "vision": "Build the most developer-friendly platform in the industry",
  "values": [
    {"id": "v1", "name": "Developer Experience", "priority": 1},
    {"id": "v2", "name": "Reliability", "priority": 2},
    {"id": "v3", "name": "Performance", "priority": 3}
  ],
  "methods": [
    {
      "id": "m1",
      "name": "Launch self-service portal",
      "description": "Enable developers to provision resources without tickets",
      "status": "in_progress",
      "owner": "Platform Team"
    }
  ],
  "obstacles": [
    {"id": "o1", "name": "Legacy system dependencies", "severity": "high"}
  ],
  "measures": [
    {
      "id": "m1",
      "name": "Developer satisfaction",
      "target": "4.5/5",
      "current": "3.8/5"
    }
  ]
}
```

## Related

- [Organization Settings](organization.md)
- [Roadmap](roadmap.md)
