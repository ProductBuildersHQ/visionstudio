# Repositories

## Repository Catalog

**Route:** `/repositories`

Lists every registered repository, grouped by organization (alphabetical), with a header showing total repository and RMI counts plus a small donut chart of overall RMI status. Each repository card shows: name, organization, domain (if set), a status badge, its RMI count, and a progress bar with percentage (share of its RMIs that are `completed`, `released`, or `done`). Click a card to open its detail view.

If no repositories are registered, the panel shows an empty state instead.

### Registering a repository

```bash
visionstudio registry add --org <github-org> --name <repo-name> --path /path/to/your/repo
```

`registry add` warns (without blocking) if `--path` is missing, isn't a directory, isn't a git working tree, or collides with an already-registered repository's path or git remote.

### Managing repository entries

```bash
visionstudio registry update <repo-id> [--path --org --branch --name]
visionstudio registry rename <old-id> <new-id> [--path --go-module --move-releases --dry-run]
visionstudio registry archive <repo-id> --superseded-by <new-id> [--reassign-rmis] [--reason "..."]
visionstudio registry remove <repo-id> --force
visionstudio registry doctor
```

- `update` repoints or corrects an entry's metadata. `--org`/`--name` edit metadata only — they don't rename the repository ID.
- `rename` moves a repository to a new ID after the repository itself was renamed, in one transaction: it registers the new ID (or reuses an existing record), repoints the RMIs, spec documents, initiative home repositories, repository dependencies, PRISM roadmaps, goals, and documents, and DevX reports, and archives the old record as superseded by the new one. RMI and initiative IDs do not change. Releases stay on the old ID unless `--move-releases` is given, since they were published under that name. `--dry-run` reports what would move. Running it again finishes a rename done partly by hand. On Dolt the rename is one commit.
- `archive` marks a repository archived without deleting its record (or the RMIs/releases/spec documents that reference it) — the preferred way to handle a merge. For a rename, use `rename`, which moves everything that references the repository and not only its RMIs. `--superseded-by` records the replacement repository; add `--reassign-rmis` to also repoint every RMI on this repository to the replacement in the same operation (equivalent to running `rmi bulk-update` first).
- `remove` hard-deletes a repository record — for true mistakes only (e.g. a typo'd `add`). It always refuses while any RMI, release, or spec document still references the repository, and otherwise requires `--force`.
- `doctor` walks every registered repository's local path and reports missing directories, directories that aren't git working trees, and git remotes that no longer match the registered ID. Report-only.

Repository references accepted by `--repo` flags across the CLI (`rmi list`, `rmi create`, `rmi update`, `rmi bulk-update`, `initiative list`) resolve short names and `org/name`, not just the full `github.com/org/name` ID — an unrecognized value returns a did-you-mean suggestion instead of silently matching nothing.

See `visionstudio registry --help` for the full set of registry commands (`add`, `list`, `update`, `rename`, `archive`, `remove`, `doctor`, `deps`, `scan`, `unpushed`, `org`, `person`, `visibility`, `focus`).

## Repository Detail

**Route:** `/repository/*` (matches nested paths, since repository IDs can contain slashes)

The header shows the repository name and status badge, with its organization underneath. Four stat cards follow: **RMIs**, **Initiatives** (distinct count of initiatives with at least one RMI in this repo), **Progress** (percentage of this repo's RMIs completed/released/done), and **Completed**.

A **Repository Info** card shows whichever of these fields are set: Go module path, default branch, domain, and local filesystem path.

Below that, a two-column layout:

- **Initiatives** — every initiative touching this repository, each row showing ID, status badge, title, a progress bar (colored green/blue/yellow by percentage), percentage, and an "N/M" completed-count. Click a row to open that initiative.
- **RMI Status** — a donut chart of this repository's RMIs by status.

Finally, a full **Roadmap Items** table: ID, title, initiative (clickable, links to initiative detail), status badge, and priority.
