# Repositories

## Repository Catalog

**Route:** `/repositories`

Lists every registered repository, grouped by organization (alphabetical), with a header showing total repository and RMI counts plus a small donut chart of overall RMI status. Each repository card shows: name, organization, domain (if set), a status badge, its RMI count, and a progress bar with percentage (share of its RMIs that are `completed`, `released`, or `done`). Click a card to open its detail view.

If no repositories are registered, the panel shows an empty state instead.

### Registering a repository

```bash
visionstudio registry add --org <github-org> --name <repo-name> --path /path/to/your/repo
```

See `visionstudio registry --help` for the full set of registry commands (`add`, `list`, `deps`, `scan`, `unpushed`).

## Repository Detail

**Route:** `/repository/*` (matches nested paths, since repository IDs can contain slashes)

The header shows the repository name and status badge, with its organization underneath. Four stat cards follow: **RMIs**, **Initiatives** (distinct count of initiatives with at least one RMI in this repo), **Progress** (percentage of this repo's RMIs completed/released/done), and **Completed**.

A **Repository Info** card shows whichever of these fields are set: Go module path, default branch, domain, and local filesystem path.

Below that, a two-column layout:

- **Initiatives** — every initiative touching this repository, each row showing ID, status badge, title, a progress bar (colored green/blue/yellow by percentage), percentage, and an "N/M" completed-count. Click a row to open that initiative.
- **RMI Status** — a donut chart of this repository's RMIs by status.

Finally, a full **Roadmap Items** table: ID, title, initiative (clickable, links to initiative detail), status badge, and priority.
