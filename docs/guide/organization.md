# Organization Settings

Organization settings configure the top-level context for V2MOM cascades and team management.

## Overview

The organization represents your company or department and provides:

- Organization-level V2MOM (top of cascade)
- Team definitions
- Shared configuration

## Accessing Organization Settings

1. Click **Organization** in the sidebar
2. View and edit organization details
3. Manage teams

## Organization Configuration

### Basic Settings

| Field | Description |
|-------|-------------|
| Name | Organization name |
| Description | Brief description |
| Fiscal Year Start | When fiscal year begins (for roadmap planning) |

### Organization V2MOM

The organization V2MOM is the top of the cascade:

1. Go to Organization Settings
2. Click **V2MOM** tab
3. Create or edit organization V2MOM
4. Team V2MOMs inherit context from here

## Team Management

### Viewing Teams

1. Go to Organization Settings
2. Click **Teams** tab
3. View list of teams

### Creating Teams

1. Click **Add Team**
2. Fill in team details:
   - Name
   - Description
   - Parent team (for hierarchy)
3. Save team

### Team Configuration

Each team has:

| Field | Description |
|-------|-------------|
| ID | Unique identifier |
| Name | Team name |
| Description | Team description |
| Parent | Parent team (optional) |
| Members | Team members (optional) |

## Configuration File

Organization settings are stored in:

```
~/.visionspec/organization.json
```

### Example Configuration

```json
{
  "organization": {
    "id": "acme-corp",
    "name": "Acme Corporation",
    "description": "Building the future",
    "fiscalYearStart": "2024-01-01"
  },
  "teams": [
    {
      "id": "platform",
      "name": "Platform Team",
      "description": "Core platform engineering"
    },
    {
      "id": "api",
      "name": "API Team",
      "description": "Public API development",
      "parentId": "platform"
    }
  ],
  "v2mom": {
    "vision": "Be the leading platform provider",
    "values": [
      {"id": "v1", "name": "Customer First", "priority": 1}
    ],
    "methods": [],
    "obstacles": [],
    "measures": []
  }
}
```

## V2MOM Hierarchy

Organization settings establish the V2MOM hierarchy:

```
Organization V2MOM
├── Platform Team V2MOM
│   ├── Core Platform V2MOM
│   └── Infrastructure V2MOM
├── API Team V2MOM
└── Quality Team V2MOM
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/organization` | GET | Get organization details |
| `/api/organization` | PUT | Update organization |
| `/api/organization/v2mom` | GET | Get organization V2MOM |
| `/api/organization/v2mom` | PUT | Update organization V2MOM |
| `/api/organization/teams` | GET | List teams |
| `/api/organization/teams` | POST | Create team |

## Related

- [V2MOM Cascade](v2mom.md)
- [Projects](projects.md)
