# Token Attribution & Portfolio Reporting — Roadmap

**Initiative:** `INIT-TOKENATTRIB-001`
**Repository:** `github.com/ProductBuildersHQ/prism-build`
**Program:** `PROG-PRISM-CONTROL`
**Status:** Proposed

Map token spend to the planning graph: initiative and quarterly reports (initiatives → RMIs → tokens by category, cost by model) as a consistent subset of devfolio's overall spend — same omnidevx events, same pricing, explicit coverage. Design: TRD §16. Cross-repo: prism-control owns attribution and reports; devfolio owns the `devx` Dolt ingest on the shared server.

**Exit criteria:** `report tokens --quarter` over a real quarter lists every PRISM-tracked initiative with per-RMI token/cost figures; the three attribution buckets (RMI-attributed, repository-level, out-of-management) sum to devfolio's overall Token Spend for the same window, and managed spend matches devfolio filtered to the same workspaces; coverage (managed ÷ overall) is stated, not implied.

## Phase 1 — Attribution & Reports (prism-control)

**Theme:** Join token spend to assignments, RMIs, and initiatives
**Status:** Planned — 0 of 5 items completed

- [ ] `RMI-PRISMCONTROL-036` `pkg/tokens`: `TokenSource` interface + omnidevx event-JSONL reader
  - Acceptance: reads `ai.message.completed` events from the omnidevx store layout (`events/YYYY/MM/DD/*.jsonl`, path configurable); yields session ID, workspace, model, timestamps, and all four token categories; unit-tested against fixture JSONL
- [ ] `RMI-PRISMCONTROL-037` Attribution join: session→assignment (time-window scoped) primary, workspace→repository fallback, unattributed residual; cost via `omnidevx-core` embedded pricing (no local pricing table)
  - Depends on: `RMI-PRISMCONTROL-036`
- [ ] `RMI-PRISMCONTROL-038` `prismctl report tokens`: initiative and period (`--quarter`/`--since`/`--until`) modes, JSON + Markdown
  - Depends on: `RMI-PRISMCONTROL-037`
  - Acceptance: initiative mode shows per-RMI and per-model breakdown, totals, residual; period mode lists initiatives → RMIs → token/cost sorted by spend plus coverage of overall spend (managed ÷ overall; three buckets sum to devfolio's total per TRD §16); `docs/integrations/omnidevx.md` refreshed for worker auto-detection
- [ ] `RMI-PRISMCONTROL-039` Dashboard token/cost panels: initiative summary column + detail-view per-phase/per-RMI spend
  - Depends on: `RMI-PRISMCONTROL-038`
- [ ] `RMI-PRISMCONTROL-040` Program/initiative summaries: surface `description` on dashboard summary and detail views; backfill descriptions for existing programs and initiatives

## Phase 2 — Shared devx Database Integration

**Theme:** devx.token_events on the shared Dolt server
**Status:** Planned — 0 of 1 items completed

- [ ] `RMI-PRISMCONTROL-041` `devx` Dolt `TokenSource`: SQL implementation against `devx.token_events` on the shared Dolt server + cross-database views (`v_initiative_tokens`, `v_unattributed_tokens`)
  - Depends on: `RMI-PRISMCONTROL-037`; blocked externally on devfolio Dolt ingest (`RMI-DEVFOLIO-050`+)
