# Programs, Initiatives, and RMIs

## Initiatives Overview

**Routes:** `/` (all initiatives, grouped by program), `/status` (all initiatives, grouped by lifecycle status), `/program/:programId` (one program), `/standalone` (initiatives with no program).

The header shows a count ("N initiatives, M RMIs") and, when there's any RMI status data, a small donut chart with the top 4 statuses and their counts alongside it.

- **All Initiatives** (`/`) groups initiative tiles under each program's name, with a per-program progress bar showing the average progress across that program's initiatives, followed by a **Standalone** group for initiatives with no `programId`. Hidden programs and hidden initiatives are excluded — see [Hiding Programs and Initiatives](#hiding-programs-and-initiatives).
- **By Status** (`/status`, sidebar entry above the program list) groups every visible initiative into pipeline-ordered columns — `proposed` through `cancelled` — for managing WIP at a glance: how many initiatives are actually `executing` right now vs. sitting `delivery_complete` unreleased vs. still `proposed`. Reuses the same `/api/execution` data as every other initiative view; no separate endpoint.
- **Program view** (`/program/:programId`) shows only that program's initiatives, ungrouped.
- **Standalone view** (`/standalone`) shows only initiatives with no program, ungrouped.

Each initiative tile shows: the initiative ID (monospace), a status badge, the title (clamped to 2 lines), and a progress bar with percentage. Click a tile to open its detail view.

### Creating an initiative

A **+ New Initiative** button in the header (also shown on the empty state) opens a creation form: ID, title, description, type, priority, program, and — required — a **spec workflow** chosen from the [`specification-workflow-spec` catalog](specs-and-evaluation.md#spec-workflows). Selecting a workflow previews its required document sequence and optional documents; changing the initiative type updates the suggested workflow (maintenance/refactor/migration → `quick-fix`, otherwise `pbhq-lite`) until you pick one explicitly. In a program view, the program field is pre-selected.

On create you land on the new initiative's **Definition** tab, which renders the selected workflow's document layout. Mutations remain CLI/API-first — this form is backed by `POST /api/initiatives`, the API's only write endpoint; everything else (workflow switching, hiding, transitions) is still done via the CLI.

## Initiative Detail

**Route:** `/initiative/:initiativeId`

The header shows the initiative ID, status badge, a chip with the initiative's spec workflow (read-only; hover for the full name, and whether it's a type default), title, description (if any), and an overall completion percentage. Below that, four summary cards: **Definition** (how many of the assigned workflow's *required* spec documents exist on disk, as a fraction and percentage — the denominator follows the workflow, e.g. 4 for `pbhq-lite`, 7 for `aws-two-way-door`), **Phases**, **RMIs**, and **Repos** (distinct repository count across the initiative's RMIs).

Two tabs follow — **Definition Details** and **Execution Details**. The initiative opens on whichever tab has data (Execution if the initiative has any phases or RMIs, otherwise Definition); the Execution tab carries an "empty" badge if there's nothing in it yet.

### Definition Details tab

- **Workflow diagram** — the assigned workflow's required documents shown as connected boxes in flow order (e.g. PRD → TRD → PLAN → ROADMAP for `pbhq-lite`; PRESS → FAQ → … for the AWS door workflows). Each box is colored by its most-advanced known state: gray "not created" (no file, no evaluation), blue "spec exists (not evaluated)", or green/yellow/red for an evaluated spec scoring ≥4, ≥3, or below 3 out of 5 respectively. An average score badge appears next to the diagram once any spec has been evaluated, using the same ≥4/≥3/below-3 color scale as the individual boxes. **Click any box** to open that document type's authoring **template** (including its Leadership Principle guidance, for the AWS workflows) and its **LLM-as-a-Judge rubric** (categories, weights, and pass/partial/fail criteria) — most useful for gray boxes, answering "what should this document contain and how will it be judged?" before writing it. The same Template/Rubric buttons appear in the spec files viewer for the currently selected document.
- **Spec file viewer** — tabs across the initiative's spec files (in workflow order first: PRD, TRD, PLAN, ROADMAP, then anything else), rendering the selected file's Markdown inline (max height, scrollable) along with its file path and last-modified date. "Open Full View" links to the standalone [Spec Viewer](specs-and-evaluation.md) for the selected spec.
- **LLM-as-a-Judge results** — a collapsible list of every judge evaluation for this initiative, newest first, each row showing spec type, filename, model (if recorded), date, and a pass/fail or score-colored badge (green ≥4, yellow ≥3, red below 3, out of 5). Click a row to expand its rationale text.
- **Initiative dependencies**, if any — see [below](#dependencies).

### Execution Details tab

If the initiative has no phases and no RMIs, this tab shows an empty state pointing at `ROADMAP.md`. Otherwise:

- **RMI status counts** — inline counts per status (not a chart here).
- **Repository chips**, if RMIs span more than one repository.
- **Initiative dependencies**, if any — see [below](#dependencies).
- **Phases**, each a collapsible card (expanded by default) showing the phase title, RMI count, and a progress bar/percentage. Expanding a phase lists its RMIs in sequence order, each row showing: type icon, RMI ID, title, a "→ N" indicator with a tooltip listing what it depends on (if it has dependencies), who claimed it and when (if claimed), completion date (if completed), and a status badge.

### Dependencies

Initiative-level dependencies show as chips reading either "requires `OTHER-ID` (Other Title)" (this initiative depends on the other) or "`OTHER-ID` requires this" (the other depends on this initiative). RMI-level dependencies show inline on each RMI row as a "→ N" badge — hover it to see which RMI IDs it requires.

## Initiative Lifecycle

An initiative moves through a status pipeline, managed from the CLI:

```
proposed → planned → executing → delivery_complete → releasing → released → closed
```

```bash
visionstudio initiative transition <initiative-id> <status>
```

Three transition rules apply:

- **Forward** — one step at a time through the pipeline (no skipping). From `delivery_complete` an initiative can also go straight to `closed` (delivered but never released), and any pre-release status can go to `cancelled`.
- **Backwards (reopen)** — an initiative can return to *any earlier* pipeline status. Scope evolves: when new phases land in a delivered or released initiative's roadmap, its status can follow the work back (e.g. `delivery_complete → executing`). `closed` can be reopened too, but is only ever *entered* going forward.
- **Cancelled** — reopens to any pre-release status; it can never jump to `released` or `closed`, which would fabricate lifecycle history that never happened.

Each pipeline stage stamps a timestamp when entered (`planned_at`, `executing_at`, `delivery_complete_at`, `released_at`, `closed_at`). A backwards transition **clears the stamps of the stages it undoes** — an initiative reopened to `executing` is no longer delivery-complete, and its record won't claim otherwise; re-entering a stage later re-stamps it. Cancellation preserves all stamps.

### Finding initiatives ready to advance

```bash
visionstudio initiative sweep [--format json]
```

Lists non-terminal initiatives (`proposed`/`planned`/`executing`) where **every** RMI is `completed` — a signal that recorded status has fallen behind actual progress. For each candidate, `sweep` resolves every distinct repository referenced by its RMIs (not just the initiative's home repo — an initiative's work often spans more than one) and reports local git state per repo: clean and in sync, dirty (uncommitted changes), unpushed commits, behind upstream, or not found/not registered locally. Read-only checks against cached refs, same posture as [`registry doctor`](repositories.md#managing-repository-entries) — no network fetch.

`sweep` never transitions anything itself. It's a starting point for review, not an auto-closer: a `completed` RMI status doesn't guarantee the shipped code actually matches what the RMI describes — verify that by hand (or with an agent) before transitioning or recording a release, especially for multi-repo initiatives.

`sweep` is initiative-first. When you're instead releasing a specific repo and want to know which initiatives should ride along, flip the direction:

```bash
visionstudio release candidates --repo <repo-id> [--format json]
```

Lists every non-terminal initiative with at least one RMI in that repo, classified `ready` (every RMI across every repo it touches is done), `partial` (this repo's RMIs are done but the initiative has open work elsewhere — record the release, don't close yet), `not_ready`, or `already_attached` (a prior release of this repo already lists it). Same report-only discipline as `sweep`.

## Hiding Programs and Initiatives

Programs and initiatives can each be hidden from the dashboard independently:

```bash
visionstudio program hide <program-id>
visionstudio program show <program-id>

visionstudio initiative hide <initiative-id>
visionstudio initiative show <initiative-id>
```

**Hiding a program cascades to its initiatives** — every initiative attached to a hidden program is excluded everywhere initiatives are listed, not just from that program's own group. Individually hiding a single initiative has the same effect, scoped to just that initiative.

This is enforced consistently across:

- The sidebar's Initiatives nav (program groups and the Standalone group)
- Initiatives Overview (`/`, `/program/:id`, `/standalone`)
- A repository's linked-initiatives list
- Performance's Cost by Initiative table

Hiding is a *listing* concept, not access control: an initiative or program hidden this way is still reachable by navigating directly to its URL (`/initiative/:id` or `/program/:id`) if you already know the ID — it just won't appear in navigation or browsing. `initiative list` and `program list` both show a `HIDDEN` column so you can check status without opening the dashboard.

The [Maturity page's Capability Models initiative filter](maturity.md#capability-models) also excludes hidden initiatives. One narrow exception: Performance's [Accomplishments](performance.md#accomplishments) list shows completed RMIs (not initiatives) with their parent initiative's title for context, and doesn't currently filter by the parent initiative's hidden status.

## Working with RMIs from the CLI

```bash
visionstudio initiative list
visionstudio initiative list --repo <repo> --status <status> --program <program-id>
visionstudio rmi list --initiative <initiative-id>
visionstudio rmi list --repo <repo>
```

`--repo` flags (here and on `rmi create`/`rmi update`) accept a short repository name or `org/name`, not just the full `github.com/org/name` ID. Add `--format json` to `initiative list`, `rmi list`, `rmi get`, and `registry list` for scriptable output.

Roadmap items and phases are populated by syncing `ROADMAP.md` into the database — see [Quick Start](../getting-started/quickstart.md#working-with-initiatives-from-the-cli).

### RMI origin

Every RMI has an `origin` — how its scope was identified: `spec` (default, in the original PRD/ROADMAP), `implementation` (the agent found it necessary while building another RMI), `acceptance_testing` (a human found it using the shipped result), or `discussion` (proposed directly in conversation). Set it with `--origin` on `rmi create`/`rmi update`, filter with `rmi list --origin <value>`; non-default origins show in `rmi get`'s output and an ORIGIN column on `rmi list`. See CLAUDE.md's "RMI provenance" convention for when to file scope discovered mid-work as a new RMI rather than folding it into an existing one.

### Reassigning RMIs to a different repository

```bash
visionstudio rmi bulk-update --repo <old-repo> --set-repo <new-repo> [--initiative <id>] [--dry-run]
```

Repoints every matching RMI's repository field in one call — `--dry-run` previews the change first. This is folded into `registry archive --superseded-by <new-id> --reassign-rmis` as a single operation when the reassignment is part of retiring the old repository — see [Repositories](repositories.md#managing-repository-entries).
