# Maturity Model Integration

This directory contains the architecture decision records and design documentation for integrating the PRISM Maturity Model into VisionStudio.

## Documents

| Document | Description |
|----------|-------------|
| [ADR-001-embed-html-dashboard.md](./ADR-001-embed-html-dashboard.md) | Decision to embed prism-maturity HTML dashboards |
| [DESIGN.md](./DESIGN.md) | Detailed technical design and implementation |
| [PLAN.md](./PLAN.md) | Implementation checklist and task breakdown |

## Overview

### Current State

VisionStudio displays:
- Spec workflows (MRD → PRD → TRD)
- V2MOM cascade visualization
- Evaluation findings

### Target State

Add maturity model visualization that shows:
- Capability maturity levels (M1-M5)
- Goal progress tracking
- Phase/initiative roadmaps
- Maturity aggregation across capability stack

### Approach

**Phase 1 (Current):** Embed prism-maturity HTML dashboards via iframe
**Phase 2 (Future):** Native React components with tighter integration

## Related Repositories

| Repo | Role |
|------|------|
| [prism-maturity](https://github.com/grokify/prism-maturity) | Maturity model engine and HTML dashboard |
| [prism-capability](https://github.com/grokify/prism-capability) | Capability stack definitions |
| [visionapp](https://github.com/ProductBuildersHQ/visionapp) | VisionStudio desktop application |

## Quick Links

- [prism-maturity dashboard/html.go](https://github.com/grokify/prism-maturity/blob/main/dashboard/html.go)
- [VisionStudio V2MOM components](../../../desktop/renderer/src/components/v2mom/)
