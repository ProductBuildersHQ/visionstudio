# PLAN — INIT-VISIONSTUDIO-007: Org, Person, and Visibility Modeling in the Registry

## Sequencing Rationale

Single-phase, four RMIs, schema-then-data-then-consumers:

1. **Entities and backfill first** (Organization, Person, visibility
   field): pure schema + migration work; backfill from existing registry
   strings and one hand-authored Person row.
2. **Visibility ingest second**: GitHub-derived, refresh-on-ingest;
   `unknown` semantics proven before any consumer trusts the field.
3. **Safety rails third**: the two-filter export guard lands *before*
   VS-006's public export ships (coordinate ordering — the guard should
   exist by the time RMI-VISIONSTUDIO-309/310 go live).
4. **Slices last**: per-org rollups, private-repo focus list, ACTS
   practitioner-lens query.

## Prerequisites (read before starting any RMI)

- **UI-facing deliverables in RMI-404 (private-repo focus list, per-org
  rollup views) are blocked on INIT-VISIONSTUDIO-001 Phases 2–3** — the
  unified web app they render in is not built yet (that initiative is at
  Phase 1 in progress). Deliver the entities, ingest, export rail, and
  query/CLI forms of the slices first; panel rendering follows the web
  foundation.
- Schema and backfill work (RMI-401..403) has no UI dependency and can
  start immediately against the current Ent schema
  (`ent/schema/repository.go` holds `organization` as a plain string
  today).

## Working Agreements

- Commits carry `Refs: RMI-VISIONSTUDIO-40N` (4xx block; 3xx belongs to
  INIT-VISIONSTUDIO-006).
- `Repository.organization` string is not removed — additive edge, no
  breaking migration.
- Settle prism-core Person import-vs-mirror before writing the schema
  (TRD risk 3).

## Dependencies and Coordination

- **INIT-VISIONSTUDIO-006:** the repo-visibility export guard (RMI-403)
  should land before or with the public export (RMI-VISIONSTUDIO-309/310).
  The internal releases panel and board are unaffected.
- **INIT-PBHQSITE-001:** per-org coverage enumeration can use the DB once
  orgs exist; not blocking — releaselog works from GitHub directly.
- **ACTS (INIT-ACTS-001):** the practitioner-period rollup consumes the
  Person → orgs → repos resolution (its TRD T3 portfolio lens); coordinate
  when ACTS Phase 2 lands.
- **prism-core:** Person primitive alignment.
