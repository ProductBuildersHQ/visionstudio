# PLAN: Stable Work Item IDs

**Initiative:** `INIT-VISIONSTUDIO-013`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Proposed
**Date:** 2026-10-09

## Sequencing

**Phase 1 (Allocation)** lands the counter first: the `IDCounter` entity, the
`IDAllocatorStore` interface on Dolt and the in-memory store, allocation
inside the insert transaction, and the migration that seeds marks from
existing data. Until that is in, nothing else has a mark to read or raise.
Explicit-number handling and aliases both build on the counter but are
independent of each other, so they can proceed in parallel once it exists.

**Phase 2 (Remote mode and docs)** makes the cloud behave the same way and
writes the policy down. Remote parity waits for both halves of Phase 1 (the
remote store needs an allocation interface and an alias resolver to
implement) and for the cloud API's endpoints, which are tracked in the cloud
repository. Documentation and release close the initiative once local and
remote agree.

## Milestones

| Milestone | Definition of done |
|-----------|--------------------|
| M1: Counter | `IDCounter` Ent schema + codegen; `IDAllocatorStore` on `MemStore` and `doltstore`; `CreateRMI`/`CreateInitiative` allocate and insert in one transaction; parallel-creates test passes |
| M2: Seeded marks | `InitIDCounters` migration step; on a seeded database `next-id` returns the same number before and after |
| M3: Explicit numbers | `--id` and `roadmap import` raise the mark; `next-id` reads the mark; a removed highest-numbered item's number is not reissued |
| M4: Aliases | `IDAlias` schema; `reidentify` commands; `get`, dependencies and work commands resolve aliases; allocation skips aliases |
| M5: Remote parity | `pkg/remote` allocates and resolves through the cloud `/ids/*` endpoints; `remotetest` covers them; `visionstudio --remote rmi create` without `--id` works and never computes a number locally |
| M6: Docs and release | ID allocation policy in the dashboard guide and remote-mode guide; changelog; release tagged after CI |

## Risks

- **Two allocators during the transition** — if the service's old
  list-and-max path survives next to the counter, the two can disagree.
  Mitigated by removing `ListAllRMIs`-based allocation in M1 rather than
  keeping it as a fallback, and by the M2 equality check.
- **Transaction plumbing is new** — `DoltUnitOfWork.Execute` has never been
  on a real write path and currently does not pass the transactional client
  through. The fix is small but touches the store constructor; it lands in M1
  with its own test before any allocation depends on it.
- **Remote mode without endpoints** — until the cloud ships its routes, a
  remote `rmi create` without `--id` must fail with a not-supported error
  instead of silently racing on a client-side scan. The remote store returns
  that error from M5's first commit, and the remote-mode guide lists it as a
  gap until parity lands.
- **Additive schema, low migration risk** — two new tables, no change to
  existing columns or edges; `InitIDCounters` only creates or raises marks,
  so re-running it is safe.
