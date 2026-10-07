# Deterministic Context Assembly — Roadmap

**Initiative:** `INIT-PRISMCONTROL-002`
**Repository:** `github.com/ProductBuildersHQ/prism-build`
**Program:** `PROG-PRISM-CONTROL`
**Status:** Proposed

Specs are the cache, not the chat: assemble agent working context deterministically from the PRISM graph and repository specs, not accumulated conversation history. Design: TRD §15; origin: `IDEATION_CHAT_CACHE-OPTIMIZATION.md`. Sessions keep working context for the duration of a phase and re-ground from authoritative sources at each phase/RMI start.

**Exit criteria:** a fresh Claude Code session, given only `prismctl context build` output for a phase of a real initiative, claims and completes an RMI end-to-end without the prior transcript; `context build` output is byte-identical across two runs at the same Dolt/git revisions.

## Phase 1 — Context Assembly

**Theme:** Specs are the cache, not the chat
**Status:** Planned — 0 of 6 items completed

- [ ] `RMI-PRISMCONTROL-030` `pkg/contextbuild`: deterministic context-package assembly for a phase or RMI
  - Acceptance: package contains graph data (program/initiative/phase/RMI, acceptance criteria, decisions), derived repo set (RMI repo + dependency-RMI repos + registered repo dependencies), spec-file references with revisions (not copies), sections ordered stable→volatile with stability classes; byte-identical output across two runs at the same revisions; Go structs → generated JSON Schema (invopop + schemago, `//go:embed`)
- [ ] `RMI-PRISMCONTROL-031` Phase handoff projection: aggregate member-RMI assignment handoffs, evidence, and derived status into one structured artifact (JSON + Markdown)
  - Depends on: `RMI-PRISMCONTROL-030`
- [ ] `RMI-PRISMCONTROL-032` `prismctl context build <rmi-id|phase-id>` with `--format json|markdown` and `--out`
  - Depends on: `RMI-PRISMCONTROL-030`, `RMI-PRISMCONTROL-031`
- [ ] `RMI-PRISMCONTROL-033` MCP `context_build` tool (single tool; same service method as the CLI)
  - Depends on: `RMI-PRISMCONTROL-032`
- [ ] `RMI-PRISMCONTROL-034` Session-protocol update: phase-start/RMI-start re-grounding via `context build`; phase-boundary context lifecycle (continue/compact/reset) documented in `AGENTS.md` + product-repo pointer
  - Depends on: `RMI-PRISMCONTROL-032`, `RMI-PRISMCONTROL-033`
- [ ] `RMI-PRISMCONTROL-035` Explicit context overrides: optional `context_spec` on RMIs (extra repos, included/excluded spec docs) with validation
  - Depends on: `RMI-PRISMCONTROL-030`
