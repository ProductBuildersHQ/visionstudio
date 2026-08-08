# PRD: Dual-Loop Lifecycle with Waterline Ratification

**Initiative:** INIT-VISIONSTUDIO-004
**Status:** Draft
**Author:** John Wang
**Date:** 2026-08-07

## Problem Statement

VisionStudio records an initiative for the first time at the **tail** of the
product-development lifecycle. In the real working flow, an initiative is *born
fully planned*: a human and a coding agent finish Define → Approve → Accept →
Plan in a single grounded session, the agent writes the specs, initiative,
phases, and RMIs in one batch, and only then does anything land in VisionStudio.
Everything upstream — the entire [Product Loop](https://productbuildershq.com/frameworks/product-loop-builder-loop)
plus the first Builder-Loop stations — happened before the tool ever saw a row.

This produces four concrete gaps:

1. **The Product Baseline is never captured.** The versioned artifact the human
   approved (`product.approve → builder.accept`, the "Product Baseline" seam) is
   destroyed the moment specs are edited during Build. It is the one thing that
   cannot be reconstructed after the fact.
2. **VisionStudio models only the Builder Loop.** The `Initiative` lifecycle
   (`created → planned → executing → delivery_complete → released → closed`) is
   pure Builder Loop. The tool embodies the exact asymmetry the two-loop
   framework was written to expose: a task that "appears from nowhere."
3. **There is no batch author path.** No `phase_create` MCP tool exists at all;
   there is no single call that writes initiative + phases + RMIs together, so
   the agent's one-pass authoring is not representable as one operation.
4. **There is no ratification gate.** The single human review that stands
   between agent-authored artifacts and Build kickoff is not modeled. Approval
   is invisible, so "the spec is the contract" is unenforceable.

## Solution

Add a **dual-loop lifecycle** to initiatives, authored in one batch and gated by
a **single human ratification** rendered as a **waterline**.

The design is governed by two hard constraints drawn from how the flow actually
works:

- **Token economics.** The upstream Product Loop runs on a free web-tier agent;
  the metered coding agent is reserved for coding. Therefore provenance capture
  must cost **zero coding-agent tokens** — capture is a mechanical byte-copy off
  disk, never an agent summarization pass.
- **Compression, not skipping.** The stations are traversed in order, but
  AI-DLC velocity means they are authored together and reviewed **once**. The
  model compresses *human gates* (many → one), not *phases* (every station gets
  a real artifact).

### The lifecycle

```
Agent batch-authors every pre-build station  →  Human ratifies the bundle once  →  Build kicks off
        (all stations: authored)                  (waterline: authored → ratified)     (executing)
```

- All pre-build stations land in **`authored`** state in a single MCP call.
- One human review sets a **waterline**: "ratified through station X." Every
  station at or before the waterline flips to **`ratified`**; the rest stay
  `authored` (i.e. bounced for rework).
- **Build unlocks iff the waterline reaches `builder.approve`.**
- The **Product Baseline is snapshotted at the moment the waterline crosses
  `product.approve`** — the versioned frozen artifact Build is measured against.

### Why a waterline, not per-station gates or checkboxes

In a linear loop, approval is **monotonic**: approving Plan is meaningless unless
Define/Approve/Accept are also approved. So a single high-water mark
(`ratified_station`) encodes the entire approval state — every station's status
is *derived*, never set individually (consistent with the existing "phase status
is derived from RMIs, never set directly" convention). This is the opposite of a
Jira turnstile: free entry at any station, no forced cycling, integrity enforced
by **artifact invariants** (you cannot ratify through `builder.approve` unless a
Product Baseline exists), not by transition order.

Explicit per-station **checkboxes** are retained only as an advanced escape hatch
for the rare genuinely non-contiguous case; the waterline is the default and
handles both approve-all and kickback with one gesture.

## User Stories

### As the coding agent...

1. **I want one batch call** that writes the initiative, its phases/stations,
   and its RMIs together in `authored` state, so my single authoring pass is one
   operation.
2. **I want to attach the ideation transcript** (`IDEATION_CHAT*.md`) and the
   finalized define specs as origin artifacts, by reference/byte-copy, without
   spending tokens re-summarizing them.

### As the human governor...

1. **I want to review the authored bundle once** and ratify through a station
   with a single gesture, so I am not clicking a gate per phase.
2. **I want to kick back a bad station** by dropping the waterline below it, so
   Build stays locked until the rework is re-ratified.
3. **I want Build to be impossible** until a Product Baseline has been approved,
   so drift is detectable and the spec is genuinely the contract.
4. **I want to see the two loops** — Product and Builder — with the ratification
   line drawn across them, so the state of an initiative is legible at a glance.

### As a product-metrics consumer...

1. **I want time-in-Define vs time-in-Build** and Product-Loop duration
   (ideation date → ratify) as an autonomy proxy.
2. **I want spec-drift detection** — the approved baseline vs. current specs.

## Scope

### In scope

| Area | Capability |
|------|-----------|
| Loop catalog | Consume `pkg/loop` + `pbhq-two-loop.yaml` from specification-workflow-spec (stations, gates, seams) |
| Schema | Station typing on `Phase`; `ratified_station` / `ratified_at` / `ratified_by` high-water mark on `Initiative`; Product Baseline snapshot |
| Write path | `phase_create` MCP tool; `initiative_author` batch tool; `initiative_ratify` waterline action |
| Gate | Build (`executing`) gated on ratified reaching `builder.approve` |
| UI | Dual-loop waterline view in the `web/` SPA, extending `InitiativeDetail` |
| Metrics | Loop-duration, time-in-Define-vs-Build, spec-drift |

### Out of scope

- Multi-user concurrent ratification / role separation (single-governor first).
- Automating the `product.approve` gate (framework: it stays human at every
  autonomy level — deliberately never automated).
- Live-logging the free web-tier session (it has no MCP; the file bridge stands).
- A visual loop *editor* (stations come from the upstream YAML catalog).

## Success Metrics

1. **Baseline capture:** 100% of initiatives that reach Build have a versioned
   Product Baseline snapshot taken at ratify.
2. **One-gesture ratification:** median human ratification actions per initiative
   before Build = 1.
3. **Drift visibility:** spec-drift (baseline vs. current) is computable for any
   initiative that has shipped.
4. **Legibility:** the dual-loop view renders both loops and the waterline for
   any initiative without additional agent calls.

## Dependencies

- **INIT-VISIONSTUDIO-001** — unified backend / Dolt store.
- **INIT-VISIONSTUDIO-003** — spec MCP tools and `SpecDocument`/evaluation model
  (the Product Baseline reuses this).
- **specification-workflow-spec** — module bump to a version exposing
  `pkg/loop` and `pkg/loops/default/pbhq-two-loop.yaml`.

## Related Decisions

See [ADR-001: Dual-Loop Lifecycle and Waterline Ratification](../../adrs/dual-loop-ratification/ADR-001-dual-loop-ratification.md)
for the architectural decisions (high-water-mark encoding, extend-Phase vs.
new-entity, AIDLC unification, mechanical snapshot).
