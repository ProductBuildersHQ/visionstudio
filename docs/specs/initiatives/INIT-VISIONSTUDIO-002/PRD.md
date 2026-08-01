# PRD — INIT-VISIONSTUDIO-002: Hosted Multi-Tenant Backend (Remote Mode)

**Initiative:** INIT-VISIONSTUDIO-002
**Status:** proposed (vision — not yet prioritized)
**Type:** feature
**Workflow:** pbhq-standard (TRD/PLAN authored when prioritized)
**Depends on:** INIT-VISIONSTUDIO-001 (Phases 1–4 complete)

## Problem

After INIT-VISIONSTUDIO-001, VisionStudio is a unified multi-domain platform —
but local-only: one user, one machine, one Dolt instance. Teams need shared
org data in a browser, and the IDE should work against that shared backend
from anywhere.

## Vision

- **Hosted daemon + Dolt** serving the same JSON API the IDE already uses
  locally; the website (`web/`) is the same React panels behind login.
- **IDE backend-picker:** local Dolt or remote URL — a base-URL + auth-token
  switch, nothing more (the HTTP-only frontend contract from
  INIT-VISIONSTUDIO-001 guarantees this).
- **Multi-tenant by org scoping** enforced in the service layer from the
  authenticated principal (org columns already exist on core entities).

## Scope (carried from INIT-VISIONSTUDIO-001 former Phase 5)

| RMI | Title |
|-----|-------|
| RMI-VISIONSTUDIO-201 | IDE backend-picker: local Dolt vs remote URL |
| RMI-VISIONSTUDIO-202 | AuthN/AuthZ for hosted daemon |
| RMI-VISIONSTUDIO-203 | Multi-tenant org scoping enforcement + isolation model decision |
| RMI-VISIONSTUDIO-204 | Hosted web shell (`web/`) — same React panels behind login |
| RMI-VISIONSTUDIO-205 | Local↔remote sync strategy (Dolt remotes/push-pull, spike) |

## Open Questions (answer before execution)

1. **Tenant isolation:** database-per-org (Dolt databases are directories —
   natural isolation, per-org branch/merge/audit) vs shared tables with org
   columns. Leaning database-per-org for Dolt's strengths.
2. **Hosted Dolt operations:** backup, scaling, upgrade story — the
   least-traveled part of the stack; needs a spike before commitment.
3. **Auth provider:** build vs integrate (GitHub OAuth is the natural fit for
   this audience).
4. **Billing/payments:** explicitly out of scope here; a later initiative.

## Success Criteria (draft)

- A user signs into the hosted website and sees only their org's programs,
  initiatives, spend, and maturity data.
- The same IDE build connects to local or hosted backend by switching a
  setting; all panels work identically.
- Hosted Dolt has a tested backup/restore procedure.
