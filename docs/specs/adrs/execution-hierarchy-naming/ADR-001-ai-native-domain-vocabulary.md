# ADR-001: Keep Program / Initiative / Phase / RMI as Native Vocabulary

**Status:** Accepted
**Date:** 2026-08-10
**Deciders:** John Wang

## Context

VisionStudio's execution hierarchy — Program → Initiative → Phase → RMI —
emerged from the team's actual way of working: AI-agent-driven execution with
a Scrum-like weekly cadence, where a "release" is whatever turned out to be
shippable that week, not a forward-planned commitment. The question came up
whether this vocabulary should instead adopt terms from established PM tools
(Jira, Aha) for legibility to people already fluent in those systems.

To answer it honestly rather than by assumption, the hierarchy was mapped
against both tools using real evidence, not general tool knowledge:

- **Jira** (Advanced Roadmaps' 4-tier model: Initiative → Epic → Story/Task →
  Subtask), checked against actual `visionstudio` CLI output for
  `INIT-PRISMCONTROL-003` (phases with derived rollup status, RMIs with
  `acceptance_criteria`, single-repo/single-assignee leaf units) and the Ent
  schemas (`ent/schema/roadmapitem.go`, `ent/schema/initiative.go`).
- **Aha** (Product → Initiative → {Epic, Feature} → Requirement, with Goal /
  Release / Idea outside the containment chain), checked against the real
  struct definitions in `github.com/grokify/aha-go` (`initiative.go`,
  `epic.go`, `feature.go`, `requirement.go`, `release.go`) rather than the
  AQL-query schema metadata alone.

The mapping came out consistent and well-grounded:

| VisionStudio | Jira | Aha |
|---|---|---|
| Program | Theme (Premium-only, optional) | Product / Product Line |
| Initiative | Initiative (Premium tier, not Epic) | Initiative |
| Phase | Epic | Epic |
| RMI | Story / Task | Feature |

This confirmed the granularity boundaries are sane — VisionStudio's levels
land in roughly the same places Jira and Aha draw theirs. But converging on
the *names* would import baggage that doesn't fit:

- **Epic** conventionally carries a human-authored title. VisionStudio's
  Phases are deliberately unnamed — "Phase 1, 2, 3, 4" — and that's a feature,
  not a gap: the theme is descriptive text, not an identity.
- **Story** carries Scrum's narrative/user-value framing. A large share of
  RMIs are `item_type: capability` — pure engineering work ("Add TokenSpend
  entity") with no user story to tell.
- **Feature** carries the same product-surface implication as Story and is
  wrong for the same reason.
- **Task** collides in-repo: this toolchain already uses "Task" for a
  different concept (session work tracking, cron/OS tasks). Reusing it for
  RMI would make "task" ambiguous in every conversation about the system.
- Neither Jira nor Aha has a concept for what `pkg/assignment`'s lease-based
  work-claim model does — RMI exists as a unit an **agent** claims and
  executes, not a unit a human picks up off a board. No borrowed term
  captures that; "Roadmap Item" at least stays neutral instead of overclaiming
  a narrative (Story) or a product surface (Feature) it doesn't have.
- Aha's **Release** is a forward-planning, date-driven scheduling container
  spanning initiatives. VisionStudio's actual practice inverts this: a
  release is an *emergent*, after-the-fact fact about what shipped in a given
  week, not a forward commitment. Modeling it as a plannable container would
  misrepresent how delivery actually happens here.

## Decision

Keep **Program, Initiative, Phase, RMI** as the permanent, canonical
vocabulary. Do not rename toward Jira or Aha terms.

The Jira/Aha mapping table above is retained as a **translation reference**
for onboarding or fielding questions from people coming from those tools —
not as a target the vocabulary should converge on.

No first-class `Release` entity is added. Initiative's existing lifecycle
timestamps (`executing_at`, `delivery_complete_at`, `released_at`,
`closed_at`) are sufficient to answer "what shipped when" without a separate
forward-planning scheduling container.

## Consequences

**Positive**

- Preserves the properties specific to how this team actually works:
  unnamed, sequential Phases; RMI as an agent-leasable unit of work; no
  forced step-by-step gate ceremony (consistent with
  [ADR-001 dual-loop](../dual-loop-ratification/ADR-001-dual-loop-ratification.md)'s
  D1, which rejected a "Jira turnstile" for the same reason — matching an
  external tool's process shape adds ceremony without adding information).
- Avoids term collisions and false-connotation risk (Task vs. this toolchain's
  own task-tracking; Story/Feature vs. capability-type RMIs that have no
  product narrative).

**Negative / costs**

- Newcomers fluent in Jira or Aha need a short translation pass — this ADR
  and its mapping table — to map onto VisionStudio's terms. No tooling,
  schema, or ID-format change follows from this decision.

## References

- Cross-repo ID convention already in production use: `RMI-<REPOSLUG>-<NNN>`,
  `Refs: RMI-<REPOSLUG>-<NNN>` commit trailer (`prism-roadmap`, `prism-build`,
  this repo) — a concrete cost that any rename would have had to account for.
- `ent/schema/roadmapitem.go`, `ent/schema/initiative.go`, `ent/schema/phase.go`
- External evidence: `github.com/grokify/aha-go` (`initiative.go`, `epic.go`,
  `feature.go`, `requirement.go`, `release.go`)
- [ADR-001: Dual-Loop Lifecycle and Waterline Ratification](../dual-loop-ratification/ADR-001-dual-loop-ratification.md) — prior precedent for rejecting borrowed-tool process shape in favor of the team's actual working model
