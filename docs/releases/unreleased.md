# Unreleased

**Status:** Changes pending since [v0.7.0](v0.7.0.md).

## Added

- **`registry rename`** moves a repository to a new ID after the repository itself was renamed, in one transaction. It registers the new ID (or reuses an existing record), repoints the RMIs, spec documents, initiative home repositories, repository dependencies, PRISM roadmaps, goals, and documents, and DevX reports, and archives the old record as superseded by the new one. RMI and initiative IDs keep their names. Releases stay on the old ID unless `--move-releases` is given, since they were published under that name. `--dry-run` reports what would move, running it again finishes a rename done partly by hand, and on Dolt the rename is one commit. See [Repositories](../dashboard/repositories.md).

## Full Changelog

See [CHANGELOG.md](../../CHANGELOG.md) for the complete list of changes.
