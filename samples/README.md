# VisionApp Samples

Reference implementations for structuring V2MOM, capability stack, maturity model, and roadmap artifacts.

## Directory Structure

```
samples/
├── grafana/          # Full reference implementation (real project)
│   ├── v2mom/        # Hierarchical V2MOMs (company + teams)
│   ├── capability/   # Capability stack with domains
│   ├── maturity/     # Multi-domain maturity model
│   └── roadmap/      # Product roadmap items
│
└── simple/           # Simplified onboarding example
    ├── v2mom/        # Single V2MOM
    ├── capability/   # Single capability stack
    ├── maturity/     # 3-domain maturity model
    └── roadmap/      # Simple roadmap
```

## Grafana Reference

The [Grafana reference](./grafana/README.md) demonstrates:

- **Hierarchical V2MOMs** - Top-level + team-specific V2MOMs with `parentId` linking
- **Multi-domain capability stack** - 5 domains (Core Platform, Security, Quality, API/DX, Operations)
- **Multi-domain maturity model** - Each domain has M1-M5 levels with criteria
- **Integrated roadmap** - Items reference capabilities and V2MOM goals

This is what a "real" enterprise project structure looks like.

## Simple Sample

The [Simple sample](./simple/README.md) provides:

- **Single V2MOM** - No hierarchy, clear values/methods/measures
- **Single capability stack** - 8 capabilities across 3 layers
- **3-domain maturity model** - Frontend, Backend, Operations
- **4 SLIs** - Render time, API latency, test coverage, deploy frequency
- **Quick onboarding** - Easy to understand and copy

## Hierarchy Pattern

```
                    Company V2MOM
                         │
         ┌───────────────┼───────────────┐
         │               │               │
    Team A V2MOM    Team B V2MOM    Team C V2MOM
         │               │               │
         └───────────────┼───────────────┘
                         │
              Capability Stack (shared)
                         │
              Maturity Model (shared)
                         │
                     Roadmap
```

## Schemas

All files follow prism-* schemas:

| Artifact | Schema Location |
|----------|-----------------|
| V2MOM | `prism-roadmap/schema/v2mom.schema.json` |
| Capability | `prism-capability/schema/capability-stack.schema.json` |
| Maturity Model | `prism-maturity/schema/prism-maturity-model.schema.json` |
| Maturity State | `prism-maturity/schema/prism-maturity-state.schema.json` |
| Roadmap | `prism-roadmap/schema/roadmap-item.schema.json` |

## Using in VisionStudio

1. Copy sample directory to your project
2. Rename to match your project
3. Update `project.json` with your metadata
4. Customize V2MOMs, capabilities, and maturity model
5. Open in VisionStudio to visualize

## Adding to prism-maturity/prism-capability

The hierarchy support for multi-project/multi-team structures may be added to:

- **prism-maturity** - Support for project manifest referencing multiple maturity models
- **prism-capability** - Support for capability stack hierarchy/composition

See the Grafana `project.json` for the proposed structure.
