# Programs, Initiatives, and RMIs

## Initiatives Overview

**Routes:** `/` (all initiatives, grouped by program), `/program/:programId` (one program), `/standalone` (initiatives with no program).

The header shows a count ("N initiatives, M RMIs") and, when there's any RMI status data, a small donut chart with the top 4 statuses and their counts alongside it.

- **All Initiatives** (`/`) groups initiative tiles under each program's name, with a per-program progress bar showing the average progress across that program's initiatives, followed by a **Standalone** group for initiatives with no `programId`. Hidden programs and hidden initiatives are excluded — see [Hiding Programs and Initiatives](#hiding-programs-and-initiatives).
- **Program view** (`/program/:programId`) shows only that program's initiatives, ungrouped.
- **Standalone view** (`/standalone`) shows only initiatives with no program, ungrouped.

Each initiative tile shows: the initiative ID (monospace), a status badge, the title (clamped to 2 lines), and a progress bar with percentage. Click a tile to open its detail view.

## Initiative Detail

**Route:** `/initiative/:initiativeId`

The header shows the initiative ID, status badge, title, description (if any), and an overall completion percentage. Below that, four summary cards: **Definition** (how many of the 4 core PBHQ Lite specs — PRD/TRD/PLAN/ROADMAP — exist on disk, as a fraction and percentage), **Phases**, **RMIs**, and **Repos** (distinct repository count across the initiative's RMIs).

Two tabs follow — **Definition Details** and **Execution Details**. The initiative opens on whichever tab has data (Execution if the initiative has any phases or RMIs, otherwise Definition); the Execution tab carries an "empty" badge if there's nothing in it yet.

### Definition Details tab

- **PBHQ Lite Workflow diagram** — PRD → TRD → PLAN → ROADMAP shown as connected boxes. Each box is colored by its most-advanced known state: gray "not created" (no file, no evaluation), blue "spec exists (not evaluated)", or green/yellow/red for an evaluated spec scoring ≥4, ≥3, or below 3 out of 5 respectively. An average score badge appears next to the diagram once any spec has been evaluated, using the same ≥4/≥3/below-3 color scale as the individual boxes.
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
visionstudio rmi list --initiative <initiative-id>
visionstudio rmi list --repo <repository-id>
```

Roadmap items and phases are populated by syncing `ROADMAP.md` into the database — see [Quick Start](../getting-started/quickstart.md#working-with-initiatives-from-the-cli).
