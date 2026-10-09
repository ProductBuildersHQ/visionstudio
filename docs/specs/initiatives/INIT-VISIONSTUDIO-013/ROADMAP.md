# ROADMAP: Stable Work Item IDs

**Initiative:** `INIT-VISIONSTUDIO-013`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Proposed
**Date:** 2026-10-09

RMI IDs use the repo slug `VISIONSTUDIO`. Phase status is derived from
member RMI statuses and is never set directly.

## Phase 1 — Allocation

**Theme:** Numbers come from a counter that only goes up.

- [ ] `RMI-VISIONSTUDIO-572` High-water marks per slug
- [ ] `RMI-VISIONSTUDIO-573` Explicit numbers advance the mark
  - Depends on: `RMI-VISIONSTUDIO-572`
- [ ] `RMI-VISIONSTUDIO-574` ID aliases
  - Depends on: `RMI-VISIONSTUDIO-572`

## Phase 2 — Remote mode and docs

**Theme:** Cloud and local allocate the same way.

- [ ] `RMI-VISIONSTUDIO-575` Remote mode parity
  - Depends on: `RMI-VISIONSTUDIO-573`, `RMI-VISIONSTUDIO-574`, `RMI-VISIONSTUDIOCLOUD-022`
- [ ] `RMI-VISIONSTUDIO-576` Documentation and release
  - Depends on: `RMI-VISIONSTUDIO-575`
