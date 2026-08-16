# PLAN: Authentic AWS Working Backwards + Web Creation Surface

**Initiative:** `INIT-VISIONSTUDIO-009`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Delivered (retroactive record)
**Date:** 2026-08-14

> Retroactive record — the sequencing as it actually happened (2026-08-10 →
> 2026-08-13), which followed the dependency structure naturally.

## Sequencing Narrative

**Phase 1 (Web Creation Surface)** came first and independently — it needed
only 005's unified catalog. Precedent-setting (the API's first mutation), so
it was kept minimal: one endpoint, one modal.

**Phases 2–4 are an authenticity ratchet**, each step exposing the next: the
authenticity question ("is MRD-first right?") forced PR-first (Phase 2);
PR-first exposed that Amazon's real selection criterion is the door, not the
product/feature taxonomy, forcing the rename and the scale/door decoupling
(Phase 3); "as authentic as possible" then demanded the method's soul —
iteration, gates, and the Leadership Principles as living guidance rather
than a name-check (Phase 4). Each phase updated profile + D2 + docs + tests
together so the authoritative sources never diverged.

**Phase 5 (Template & Rubric Viewing)** closed the loop: once templates and
rubrics carried real LP guidance, hiding them behind the embedded FS wasted
them — surfacing them in the Definition tab makes the infusion visible at
the moment of authoring.

## Milestones (all reached)

| Milestone | Definition of done |
|-----------|--------------------|
| M1: Web create | Initiative created end-to-end from the dashboard on any catalog workflow |
| M2: PR-first | Both profiles PR-first; D2s regenerated; conformance tests pin the flows |
| M3: Doors | Renamed profiles; live DB migrated; zero stale references outside history |
| M4: LP infusion | 16 principles in metadata; LP template guidance; LP rubric categories; conformance suite |
| M5: Viewing | Template + rubric one click away from the workflow diagram |

## Risks (as managed)

- **Cross-repo drift** — every profile change landed with its D2/docs/test
  updates in the same pass; conformance tests are the drift alarm.
- **Rename fallout** — absorbed by `retiredRemap`; the sed-clobbered-keys
  incident was caught by migration-output review and is now regression-tested.
- **Release coupling** — all work sits on the dev-only `replace` directive;
  landing order is specification-workflow-spec v0.3.0 → visionstudio pin bump.
