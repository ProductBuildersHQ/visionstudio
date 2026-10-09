# PRD: Stable Work Item IDs

**Initiative:** `INIT-VISIONSTUDIO-013`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Proposed
**Date:** 2026-10-09

## Source

Found while fixing RMI and initiative IDs ahead of time for a set of
cross-repository initiatives, where a `Depends on:` line in one repository's
`ROADMAP.md` names an RMI that another repository's roadmap has not created
yet. That only works if a number, once printed by `rmi next-id` or written
into a roadmap, can never end up on a different item.

Today it can. `NextRMIID`/`NextInitiativeID` compute "highest existing number
for the slug plus one" over a full list of items. The number is derived from
whatever rows happen to exist, so it moves backwards as soon as the
highest-numbered item is gone (a Dolt revert, direct SQL, a cleaned-up failed
import), and two sessions that allocate at the same moment compute the same
number — the store's unique-ID constraint fails one of them after the fact
rather than handing out distinct numbers. In remote mode the same code runs
over the cloud tenant's complete RMI list downloaded to the client, so the
race spans every session on the tenant.

## Problem

An RMI or initiative ID is a reference, not a row: it is written into git
`Refs:` trailers, `ROADMAP.md` files, other repositories' dependency lines,
commit messages and conversation. VisionStudio treats the ID as the primary
key of a row, so it is only as stable as the row. There is no record of which
numbers have ever been issued, no rule that an issued number is retired with
its item, and no way to keep an old ID resolvable once an item has been
re-identified.

## Goals

- **Never reissue a number.** Every slug keeps a high-water mark per kind
  (`rmi`, `initiative`) that only increases. Allocation hands out mark + 1
  and advances the mark in the same transaction as the insert, so a deleted,
  moved or reverted item's number is never given to another item and
  concurrent creates get distinct numbers.
- **Explicit numbers advance the mark.** `--id` on `rmi create` /
  `initiative create` and IDs in `roadmap import` are still accepted; a
  number above the mark raises it. Gaps below the mark are expected and are
  never filled. `rmi next-id` and `initiative next-id` read the marks instead
  of scanning items.
- **Changed IDs resolve as aliases.** When an item is re-identified, its
  former ID becomes an alias. `rmi get`, `initiative get`, dependency edges
  and every other ID lookup resolve the alias to the current item, and an
  alias is never issued as a new ID.
- **Remote mode allocates the same way.** `visionstudio --remote` asks the
  cloud API for the next ID and for alias resolution instead of computing
  them from a downloaded list, so local and cloud behave identically.
- **Documented as the ID allocation policy.** The guides state the policy
  (never reissued, gaps expected, aliases) so that anyone reading a roadmap
  with holes in its numbering knows that is normal.

## Non-goals

- Changing the ID formats. `RMI-<REPOSLUG>-<NNN>` and `INIT-<SLUG>-<NNN>`
  stay as they are, including the three-digit zero padding and acceptance of
  longer numbers past 999.
- A `delete` command for RMIs or initiatives. This initiative makes deletion
  safe for numbering; it does not add one.
- Renumbering or back-filling existing roadmaps. Existing IDs are taken as
  the initial marks and nothing is rewritten.
- Automatic re-identification when an RMI changes repository. `rmi move`
  keeps the RMI's ID; aliases exist so that a deliberate re-identification can
  be recorded, not to trigger one.
- The cloud's own storage of marks and aliases, which is specified in the
  cloud repository; this initiative consumes its endpoints.
