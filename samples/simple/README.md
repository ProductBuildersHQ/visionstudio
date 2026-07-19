# Simple Sample

A simplified reference implementation for learning the PRISM framework.

## Overview

This sample demonstrates the core concepts of:

- **V2MOM**: Vision, Values, Methods, Obstacles, and Measures
- **Capability Stack**: Layered view of product capabilities
- **Maturity Model**: M1-M5 maturity levels with SLI-backed criteria
- **Roadmap**: Prioritized items linked to capabilities and goals

## Structure

```
simple/
├── project.json                    # Project manifest
├── v2mom/
│   └── simple.v2mom.json          # Single V2MOM
├── capability/
│   └── simple.capability.json     # Single capability stack
├── maturity/
│   ├── simple.model.json          # Maturity model (3 domains)
│   └── simple.state.json          # Current maturity state
├── roadmap/
│   └── simple.roadmap.json        # Product roadmap
└── README.md                       # This file
```

## Domains

| Domain | Current Level | Target | Description |
|--------|---------------|--------|-------------|
| Frontend | M3 | M4 | User interface performance |
| Backend | M3 | M4 | API quality and latency |
| Operations | M3 | M4 | CI/CD and deployment |

## SLIs

| SLI | Current | Target | Unit |
|-----|---------|--------|------|
| Page Render Time | 850ms | 500ms | ms |
| API Latency P99 | 320ms | 200ms | ms |
| Test Coverage | 65% | 80% | % |
| Deploy Frequency | 3/week | 5/week | count |

## Getting Started

1. Review `project.json` to understand the file structure
2. Read `v2mom/simple.v2mom.json` for strategic goals
3. Explore `capability/simple.capability.json` for capabilities
4. Study `maturity/simple.model.json` for maturity criteria
5. Check `maturity/simple.state.json` for current state

## Comparison with Grafana Sample

| Aspect | Simple | Grafana |
|--------|--------|---------|
| V2MOMs | 1 | 6 (1 top-level + 5 teams) |
| Capability Stacks | 1 | 6 (1 top-level + 5 domains) |
| Domains | 3 | 5 top-level, 15+ sub-domains |
| SLIs | 4 | 40+ |
| Hierarchy | Flat | Multi-level |

This sample is intentionally minimal to help you understand the concepts before exploring the more comprehensive Grafana reference implementation.

## See Also

- [Grafana Reference](../grafana/README.md) - Full implementation example
