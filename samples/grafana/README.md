# Grafana Reference Implementation

This directory contains a comprehensive reference implementation showing how to structure V2MOM, capability stack, maturity model, and roadmap artifacts for a real open source project.

## Why Grafana?

Grafana was selected as the reference implementation because:

| Criteria | Grafana |
|----------|---------|
| **Backend** | Go (pure) |
| **Frontend** | React/TypeScript |
| **Bug Bounty** | Intigriti (active) |
| **Maturity** | Very high (10+ years) |
| **Architecture** | Clean, well-documented |
| **Community** | Large, active |

## Directory Structure

```
grafana/
├── project.json                    # Project manifest
│
├── v2mom/
│   ├── grafana.v2mom.json          # Top-level V2MOM
│   └── teams/                      # Team-specific V2MOMs
│       ├── core-platform.v2mom.json
│       ├── security.v2mom.json
│       ├── quality.v2mom.json      # (planned)
│       ├── api-dx.v2mom.json       # (planned)
│       └── operations.v2mom.json   # (planned)
│
├── capability/
│   ├── grafana.capability.json     # Top-level capability stack
│   └── domains/                    # Domain-specific capabilities
│       └── (planned)
│
├── maturity/
│   ├── grafana.model.json          # Multi-domain maturity model
│   ├── grafana.state.json          # Current state assessment
│   └── domains/                    # Per-domain models (optional)
│       ├── core-platform/
│       ├── security/
│       ├── quality/
│       ├── api-dx/
│       └── operations/
│
└── roadmap/
    └── grafana.roadmap.json        # Product roadmap
```

## Hierarchy Pattern

This reference demonstrates the **hierarchical** pattern where:

1. **Top-level V2MOM** sets company/product vision and methods
2. **Team V2MOMs** inherit from top-level with `parentId` reference
3. **Capability Stack** defines all capabilities across domains
4. **Maturity Model** has multiple domains, each with M1-M5 levels
5. **Roadmap Items** reference capabilities and goals

```
┌─────────────────────────────────────────────────────────────┐
│                    Grafana V2MOM (Company)                   │
│  Vision: Ubiquitous observability platform                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│ Core Platform │ │   Security    │ │  Operations   │
│   Team V2MOM  │ │  Team V2MOM   │ │  Team V2MOM   │
│               │ │               │ │               │
│ parentId:     │ │ parentId:     │ │ parentId:     │
│ grafana-fy27  │ │ grafana-fy27  │ │ grafana-fy27  │
└───────────────┘ └───────────────┘ └───────────────┘
```

## Domain Teams

| Domain | Description | Owner |
|--------|-------------|-------|
| Core Platform | Dashboard engine, visualization, plugins | Core Platform Team |
| Security | Authentication, authorization, bug bounty | Security Team |
| Quality | Testing, CI/CD, reliability | Quality Team |
| API/DX | APIs, SDKs, documentation | API/DX Team |
| Operations | Alerting, HA, provisioning | Operations Team |

## Maturity Summary

| Domain | Current Level | Target Level | Target Date |
|--------|---------------|--------------|-------------|
| Core Platform | M4 (Managed) | M5 | Q4 2027 |
| Security | M4 (Managed) | M5 | Q2 2027 |
| Quality | M4 (Managed) | M5 | Q4 2027 |
| API/DX | M4 (Managed) | M5 | Q2 2027 |
| Operations | M4 (Managed) | M5 | Q4 2027 |

## Capability Stack Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Core Platform (Layer 1)                  │
│  ┌──────────────┬──────────────┬──────────────┬───────────┐ │
│  │  Dashboard   │   Plugin     │    Query     │   Viz     │ │
│  │   Engine     │   System     │   Engine     │  Library  │ │
│  └──────────────┴──────────────┴──────────────┴───────────┘ │
├─────────────────────────────────────────────────────────────┤
│                     Security (Layer 2)                       │
│  ┌──────────────┬──────────────┬──────────────┬───────────┐ │
│  │    Auth      │  Authorization │ Security   │   Bug     │ │
│  │ (OAuth/SAML) │    (RBAC)     │ Scanning   │  Bounty   │ │
│  └──────────────┴──────────────┴──────────────┴───────────┘ │
├─────────────────────────────────────────────────────────────┤
│                     Quality (Layer 3)                        │
│  ┌──────────────┬──────────────┬──────────────┐             │
│  │    Unit      │     E2E      │    API       │             │
│  │   Testing    │   Testing    │  Testing     │             │
│  └──────────────┴──────────────┴──────────────┘             │
├─────────────────────────────────────────────────────────────┤
│                     API/DX (Layer 4)                         │
│  ┌──────────────┬──────────────┐                            │
│  │   REST API   │    Docs      │                            │
│  └──────────────┴──────────────┘                            │
├─────────────────────────────────────────────────────────────┤
│                     Operations (Layer 5)                     │
│  ┌──────────────┬──────────────┬──────────────┐             │
│  │   Alerting   │     HA       │ Provisioning │             │
│  │   Engine     │              │  (Terraform) │             │
│  └──────────────┴──────────────┴──────────────┘             │
└─────────────────────────────────────────────────────────────┘
```

## Key Metrics (SLIs)

| SLI | Current | Target | Domain |
|-----|---------|--------|--------|
| Dashboard Load P99 | 2.1s | < 1s | Core Platform |
| Plugin Count | 380 | 500 | Core Platform |
| SAST Coverage | 95% | 100% | Security |
| Test Coverage | 76% | 85% | Quality |
| API Coverage | 95% | 100% | API/DX |
| Alert Channels | 38 | 50 | Operations |

## Roadmap Highlights

| Item | Quarter | Priority | Status |
|------|---------|----------|--------|
| Dashboard Performance | Q1-Q2 | P0 | In Progress |
| Plugin SDK v2 | Q1-Q3 | P0 | In Progress |
| SOC 2 Type II | Q1-Q2 | P0 | In Progress |
| Alerting Engine v2 | Q1-Q4 | P0 | In Progress |
| GraphQL API | Q2-Q4 | P1 | Planning |
| Passkey Auth | Q2-Q3 | P1 | Planning |

## References

- [Grafana GitHub](https://github.com/grafana/grafana)
- [Grafana Documentation](https://grafana.com/docs/)
- [Grafana Bug Bounty (Intigriti)](https://www.intigriti.com/programs/grafana)
- [prism-maturity](https://github.com/grokify/prism-maturity)
- [prism-capability](https://github.com/grokify/prism-capability)
- [prism-roadmap](https://github.com/grokify/prism-roadmap)
