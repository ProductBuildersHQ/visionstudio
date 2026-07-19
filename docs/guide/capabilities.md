# Capability Stack

The Capability Stack provides a visual way to organize and track your product or platform capabilities across domains.

## Overview

Capabilities represent what your system can do, organized into logical domains. VisionStudio helps you:

- Define capability domains (API, Platform, Security, etc.)
- Track capability maturity
- Link capabilities to roadmap initiatives
- Map capabilities to maturity assessments

## Capability Structure

```
Capability Stack
├── Domain: Core Platform
│   ├── Authentication
│   ├── Authorization
│   └── User Management
├── Domain: API
│   ├── REST API
│   ├── GraphQL
│   └── Webhooks
└── Domain: Operations
    ├── Monitoring
    ├── Logging
    └── Alerting
```

## Using Capability Stack

### Accessing Capability View

1. Click **Capabilities** in the sidebar
2. View the capability stack organized by domain
3. Expand domains to see individual capabilities

### Creating Capabilities

1. Click **Add Domain** to create a new domain
2. Within a domain, click **Add Capability**
3. Fill in capability details:
   - Name and description
   - Current maturity level
   - Owner/team
   - Related roadmap items

### Capability Details

Each capability includes:

| Field | Description |
|-------|-------------|
| Name | Capability name |
| Description | What the capability provides |
| Domain | Logical grouping |
| Maturity | Current maturity level (1-5) |
| Target | Target maturity level |
| Owner | Responsible team |

## Capability Files

Capabilities are stored as JSON:

```
project/
└── capability/
    ├── platform.capability.json
    └── domains/
        ├── api.capability.json
        ├── core-platform.capability.json
        └── operations.capability.json
```

## Example Capability

```json
{
  "metadata": {
    "name": "api-capabilities",
    "title": "API Capabilities",
    "domain": "api"
  },
  "capabilities": [
    {
      "id": "rest-api",
      "name": "REST API",
      "description": "RESTful API for external integrations",
      "maturity": {
        "current": 3,
        "target": 4
      },
      "owner": "API Team"
    },
    {
      "id": "graphql",
      "name": "GraphQL API",
      "description": "GraphQL endpoint for flexible queries",
      "maturity": {
        "current": 2,
        "target": 4
      },
      "owner": "API Team"
    }
  ]
}
```

## Integration with Maturity Model

Capabilities can be linked to maturity model dimensions:

1. Open a capability
2. Click **Link to Maturity Model**
3. Select the relevant dimension
4. Capability maturity contributes to dimension score

## Related

- [Maturity Model](maturity.md)
- [Roadmap](roadmap.md)
