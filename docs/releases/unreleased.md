# Unreleased

**Status:** In progress — spec workflow single source of truth (`INIT-VISIONSTUDIO-005`).

## Highlights

- **One workflow catalog** — all default spec workflows are defined in [`specification-workflow-spec`](https://github.com/ProductBuildersHQ/specification-workflow-spec) (~25 profiles); VisionStudio's divergent local catalog was retired, with `workflow sync` migrating the database index and remapping retired IDs (`pbhq-standard` → `pbhq-lite`)
- **Workflow switching** — `visionstudio initiative update <id> --workflow <new-id>` changes an initiative's spec workflow after creation; the Initiative page, spec expectations, and evaluation rubrics all follow
- **Workflow-aware Initiative page** — dynamic spec sets, progress denominators, tab ordering, and a per-workflow diagram replace the hardcoded PBHQ Lite rendering; files not in the selected workflow are badged **Extra**
- **AWS Working Backwards fidelity** — `aws-one-way-door` and `aws-two-way-door` profiles now match the authoritative visionspec D2 flows (execution phases, corrected synthesis DAGs), and a silent YAML parse bug that dropped pbhq-lite's execution ordering was fixed upstream
- **Reversible initiative lifecycle** — `initiative transition` now supports backwards transitions: an initiative can reopen to any earlier pipeline status as its scope evolves (e.g. `delivery_complete` → `executing` when new phases land), with the lifecycle timestamps of undone stages cleared so history stays truthful; `cancelled` reopens only to pre-release statuses, and `closed` is still only entered going forward. A persistence bug where cleared timestamps silently survived in the database (Ent's `SetNillable*(nil)` is a no-op) was fixed alongside

## Full Changelog

See [CHANGELOG.md](../../CHANGELOG.md) for the complete list of changes.
