# Unreleased

**Status:** Changes pending since [v0.7.0](v0.7.0.md).

## Added

- **`registry rename`** moves a repository to a new ID after the repository itself was renamed, in one transaction. It registers the new ID (or reuses an existing record), repoints the RMIs, spec documents, initiative home repositories, repository dependencies, PRISM roadmaps, goals, and documents, and DevX reports, and archives the old record as superseded by the new one. RMI and initiative IDs keep their names. Releases stay on the old ID unless `--move-releases` is given, since they were published under that name. `--dry-run` reports what would move, running it again finishes a rename done partly by hand, and on Dolt the rename is one commit. See [Repositories](../dashboard/repositories.md).

## Fixed

- **Units of work are transactions.** On the Dolt store, `UnitOfWork.Execute` opened a transaction but ran its function against the store outside it, so an error did not roll back the function's writes. The function now receives a store bound to the transaction. A `registry rename` inside a unit of work joins its transaction, and a Dolt commit or a nested unit of work inside one is refused.

- **Windows.** The check for whether a database or dashboard process is still running always answered yes on Windows, so `db stop` and `app stop` could not tell a stopped process from a running one. Process handling now goes through oscompat and works the same on every platform. The Windows test run passes again: tests isolate the home directory on every platform, and check file privacy only where the platform enforces modes.

## Changed

- **Credentials** are written atomically, with owner-only permissions where the platform enforces them.

- **The dashboard** opens the browser through oscompat instead of its own per-platform code, and now also opens it on the BSDs.

## Full Changelog

See [CHANGELOG.md](../../CHANGELOG.md) for the complete list of changes.
