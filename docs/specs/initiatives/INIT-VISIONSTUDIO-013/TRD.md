# TRD: Stable Work Item IDs

**Initiative:** `INIT-VISIONSTUDIO-013`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Proposed
**Date:** 2026-10-09

## Current state

- `pkg/service/numbering.go`: `NextRMIID(ctx, repoID)` derives the slug with
  `rmi.SlugFromRepoID`, calls `Store.ListAllRMIs`, and returns
  `rmi.FormatID(slug, max+1)` over every ID that `rmi.ParseID` attributes to
  that slug. `NextInitiativeID(ctx, slug)` does the same over
  `Store.ListInitiatives` with `initiative.NormalizeSlug`/`ParseID`/`FormatID`.
  Its own comment says allocation "is not transactional against concurrent
  creates; the store's unique-ID constraint remains the backstop".
- `pkg/rmi` and `pkg/initiative`: `ParseID` accepts three or more digits
  (`^RMI-([A-Z0-9]+)-(\d{3,})$`, `^INIT-([A-Z0-9]+)-(\d{3,})$`); `FormatID`
  zero-pads to three. `rmi.ValidID` (exactly three digits) has no non-test
  callers. Nothing records which numbers have been issued.
- `pkg/store/store.go`: `RMIStore.CreateRMI` and
  `InitiativeStore.CreateInitiative` take a fully-formed entity whose `ID` the
  caller chose. There is no `DeleteRMI`/`DeleteInitiative`; `PhaseStore` has
  `DeletePhase`. Dependencies (`RMIDependency`) store `source_rmi_id` and
  `target_rmi_id` as plain string columns with a unique index on
  `(source, target, relationship)` and no existence check — a dependency on
  an ID that does not exist yet is accepted.
- `pkg/store/doltstore`: `CreateRMI`/`CreateInitiative` are single Ent
  `Create().Save()` calls; the string primary key (`rmi_id`,
  `initiative_id`, `MaxLen(64)`, both carrying `TenantMixin`) is the only
  guard against a duplicate. `DoltUnitOfWork.Execute` begins an Ent
  transaction but invokes `fn(ctx, u.store)` with the non-transactional store,
  and no caller in `pkg/service` or `cmd/visionstudio` uses `UnitOfWork` at
  all — the "insert transaction" that a mark must advance inside does not
  exist yet. `Migrate` is Ent auto-migration (`Schema.Create`). The same
  package serves Postgres (`d.dolt == nil` skips the `DOLT_COMMIT` step).
- `pkg/remote`: `Store` implements `store.Store` over the cloud tenant API
  (`/rmis`, `/initiatives`, …; `unsupported.go` returns not-supported errors
  for areas without endpoints). It has no `UnitOfWork`, no next-id and no
  alias calls, so in remote mode `Service.NextRMIID` runs over
  `remote.Store.ListAllRMIs` and computes the number on the client from the
  whole tenant's RMI list.
- `cmd/visionstudio`: `rmi next-id --repo [--count N]` calls
  `svc.NextRMIID` once and prints `n..n+count-1`; `initiative next-id
  --slug|--home-repo` calls `svc.NextInitiativeID`. `rmi create` and
  `initiative create` allocate only when `--id` is omitted; an explicit
  `--id` goes straight to `CreateRMI`/`CreateInitiative`. `roadmap import`
  (`pkg/roadmap.Importer`) always passes the IDs parsed from `ROADMAP.md`.
  `rmi get`/`initiative get`/`rmi dep add`/`work claim` look the ID up
  verbatim via `Store.GetRMI`/`GetInitiative`.

## TD-1: One `IDCounter` entity keyed by (kind, slug)

A new Ent entity `IDCounter` with fields `kind` (`rmi` | `initiative`),
`slug`, `high_water_mark int` and `updated_at`, unique on `(kind, slug)`,
carrying `TenantMixin` like every other entity so the cloud can scope it per
tenant without a second schema. One entity rather than two columns on
`Repository`/`Initiative` because initiative slugs are not repositories (an
initiative may use a project slug) and RMI numbering is per slug across
initiatives, so neither parent is the right owner. Additive migration: new
table only.

## TD-2: `IDAllocator` store interface, allocation inside the insert

`store.Store` gains an `IDAllocatorStore`:

```go
type IDAllocatorStore interface {
	// AllocateID reserves and returns mark+1 for (kind, slug), advancing the mark.
	AllocateID(ctx context.Context, kind, slug string) (int, error)
	// RaiseMark advances the mark to n when n is above it; a no-op otherwise.
	RaiseMark(ctx context.Context, kind, slug string, n int) error
	// PeekMark returns the current mark without changing it (0 when absent).
	PeekMark(ctx context.Context, kind, slug string) (int, error)
}
```

`CreateRMI`/`CreateInitiative` in `pkg/service` stop accepting an empty ID
from the CLI's allocation branch; instead the Dolt store performs allocation
and insert in one Ent transaction: `SELECT … FOR UPDATE` on the counter row
(inserting it at 0 when absent), `high_water_mark + 1`, insert the item,
commit. The row lock serialises concurrent creates for the same slug so they
receive distinct numbers; a rollback (insert failure) leaves the mark where
it was, so a number is still never handed out twice. This is the first real
use of a transaction on the write path; `DoltUnitOfWork.Execute` is fixed to
hand `fn` a store bound to `tx.Client()` so other callers can use it.

The in-memory `MemStore` implements the same interface under its mutex for
service-level tests.

## TD-3: Explicit numbers raise the mark; gaps are never filled

Every path that inserts an item with a caller-supplied ID — `rmi create
--id`, `initiative create --id`, `roadmap import` — parses the number with
`rmi.ParseID`/`initiative.ParseID` and calls `RaiseMark` in the same
transaction as the insert. A number at or below the mark is accepted as long
as no item or alias holds it (reproducing an existing roadmap stays
possible), but never moves the mark down and never changes which number is
next. `NextRMIID`/`NextInitiativeID` become `PeekMark + 1` and stop listing
items; `rmi next-id --count N` keeps printing consecutive numbers from that
base. The unique-ID constraint remains as a backstop, not as the mechanism.

## TD-4: Migration initialises marks from existing data

`Migrate` is followed by a one-time `InitIDCounters` step (idempotent,
runnable by `db migrate` and on first connect): for every slug seen in
`roadmap_items.rmi_id` and `initiatives.initiative_id`, set the mark to the
highest parsed number if no counter row exists. Marks are only ever created
or raised by this step, never lowered, so running it on a database that
already has counters is harmless. After it runs, `next-id` returns exactly
what it returns today for every slug, so nothing visible changes until the
first create.

## TD-5: `IDAlias` entity and a single resolution point

A new Ent entity `IDAlias` with `kind`, `alias` (the former ID, unique per
kind), `target_id` (the current ID) and `created_at`, again with
`TenantMixin`. Aliases are written by a new `rmi reidentify <old> <new>` /
`initiative reidentify <old> <new>` (the new ID goes through TD-3, so it
raises the mark; the old ID is recorded as an alias and the dependency rows
naming it are rewritten to the new ID in the same transaction).

Resolution happens in one place, `service.ResolveRMIID` /
`ResolveInitiativeID`: try the ID as given; on not-found, look up the alias
and retry with its target. `GetRMI`, `GetInitiative`, `CreateDependency`
(both ends), `work claim`/`work complete` and the dashboard's detail routes
call the resolver rather than the store directly. `AllocateID` and
`RaiseMark` treat an alias as taken: a number whose formatted ID is an alias
is never returned by allocation, and an explicit `--id` equal to an alias is
refused with a message naming the current ID. Dependency rows keep storing
plain string IDs; the rewrite on re-identification plus resolution on write
is enough, so no edge or foreign key is introduced.

## TD-6: Remote store parity through cloud endpoints

`pkg/remote.Store` implements `IDAllocatorStore` and alias resolution over
two cloud endpoints rather than client-side scans: `POST
/t/{tenant}/ids/next` (`{kind, slug}` → `{id, number}`; the server allocates
in its own transaction and the client never computes a number) and `GET
/t/{tenant}/ids/resolve?kind=&id=` (`{id, canonical_id, alias: bool}`).
`PeekMark` and `RaiseMark` are served by `GET/PUT /t/{tenant}/ids/marks`.
Create requests that carry an explicit ID keep their shape; the server
raises the mark on its side. Until the endpoints exist, the remote store
returns the usual not-supported error for allocation rather than silently
falling back to the client-side scan, so a remote `rmi create` without
`--id` fails loudly instead of racing. The cloud's implementation of the
endpoints is tracked in the cloud repository (`RMI-VISIONSTUDIOCLOUD-022`).

## TD-7: CLI changes

- `rmi next-id` / `initiative next-id`: unchanged flags, read the mark; the
  long help describes the ID allocation policy.
- `rmi create` / `initiative create`: when `--id` is omitted the ID comes
  back from the store's allocation (the CLI no longer pre-computes it); when
  given, an alias is refused and a number above the mark raises it.
- `roadmap import`: IDs from the file raise marks; a file that names an
  alias is reported with the canonical ID in the import actions.
- `rmi reidentify` / `initiative reidentify`: new, see TD-5.
- `rmi get` / `initiative get`: when the argument was an alias, print
  `Alias of: <canonical>` on the first line of the detail.
- `db migrate` runs `InitIDCounters` after Ent migration.

## Testing

- `pkg/store`: round-trip tests for `IDCounter` and `IDAlias` on `MemStore`
  and `doltstore` (embedded Dolt), including a parallel-creates test that
  allocates from N goroutines for one slug and asserts N distinct consecutive
  numbers and a final mark of N.
- `pkg/service`: allocation after a removed highest-numbered item does not
  reissue the number; explicit `--id` above the mark raises it, below the
  mark does not lower it; `next-id` after `InitIDCounters` equals the old
  max+1 for a seeded database; alias resolution on get, dependency add and
  work claim; allocation skipping an alias; `reidentify` refusing a taken
  target.
- `pkg/remote`: `remotetest` server gains the `/ids/*` routes; tests cover
  allocation, mark peek/raise, alias resolution and the not-supported error
  when the server lacks the routes.
- `cmd/visionstudio`: CLI smoke tests for `next-id`, `create` without
  `--id`, `create --id` with an alias, `reidentify` and `get` on an alias,
  matching the existing `rmi_test.go`/`initiative_test.go` style.
