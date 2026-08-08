# ADR: Dual-Loop Lifecycle and Waterline Ratification

Architecture decision records for the dual-loop lifecycle feature
(`INIT-VISIONSTUDIO-004`).

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-001](ADR-001-dual-loop-ratification.md) | Dual-Loop Lifecycle and Waterline Ratification | Proposed |

## Summary

VisionStudio adds a Product Loop + Builder Loop lifecycle to initiatives,
authored by the coding agent in one batch and gated by a single human
**ratification** rendered as a **waterline**. The station catalog is consumed
from `specification-workflow-spec` (`pkg/loop`), the ratification gate extends
the existing `pkg/initiative` state machine, approval is encoded as a single
high-water mark with derived per-station status, and the Product Baseline is
snapshotted server-side off disk at zero coding-agent-token cost.

See the [initiative specs](../../initiatives/INIT-VISIONSTUDIO-004/) for PRD,
TRD, ROADMAP, and PLAN.
