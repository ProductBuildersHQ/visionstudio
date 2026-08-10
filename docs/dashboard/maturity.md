# Maturity

**Route:** `/maturity`

The Maturity page has three view modes, switched via buttons at the top: **SCALE Platform**, **Leverage Graph**, and **Capability Models**. Each covers a different kind of maturity signal.

## SCALE Platform

Platform-adoption metrics from the SCALE framework, organized into five aspects (rendered as colored tiles at the top: Standards, Consumption, Automation, Leverage, Effectiveness). Each tile shows a letter grade, the aspect name, a percentage score, and how many of its eligible metrics have been observed. Tile color reflects score: green ≥80%, blue ≥60%, yellow ≥40%, orange for any positive score below that, gray for zero.

Below the tiles, an assessment summary (period, as-of date, observation count) and the SCALE domain's capabilities, each an expandable card. Expanding a capability shows its metrics grouped by aspect, each metric row showing its current value (or "No data"), a target value if one is set, and an attainment bar (green ≥80% of target, yellow ≥50%, red below that).

If no SCALE catalog is configured, this view shows an empty state instead.

## Leverage Graph

Code reuse metrics across your Go module ecosystem — which internal modules are depended on, and which aren't depended on by anything (reuse opportunities).

Four summary cards up top: total modules scanned, internal module count, internal ratio (percentage of dependencies that stay within your organizations vs. external), and orphan count. Below that, two side-by-side lists:

- **Top Leveraged** — modules with the most dependents (most reused).
- **Top Consumers** — modules with the most internal dependencies of their own, split into direct/indirect counts.

A filter toggles between:

- **By Organization** — every internal module grouped by GitHub org, each group expandable into a table (module, dependents, dependencies, leverage score — a ratio above 1x means the module is depended on more than it depends on others).
- **Orphans** — internal modules with zero dependents, i.e. modules nothing else in the ecosystem currently reuses.

The footer shows when the graph was generated, plus its ecosystem (e.g. `go`) and scope.

## Capability Models

Capability maturity assessments against a defined model (dimensions, levels per dimension, and a max level). Switch between models with the buttons at the top if more than one is defined.

For the selected model:

- **Radar chart** — one line per selected initiative, one axis per dimension (only shown once a model has 3+ dimensions and at least one initiative is selected). Toggle which initiatives are plotted via the chips above the chart.
- **Dimension cards** — one per dimension, showing the average level across selected initiatives, a level-by-level bar (filled segments up to the initiative's current level), the most recent level's name/description, and any notes recorded with that score.
- **Assessments table** — every assessment, sorted newest first: initiative ID, assessed date, how many dimensions were scored, and an average-level badge (green ≥80% of max, blue ≥50%, yellow ≥30%, gray below that).

If the model has no assessments yet, this view shows an empty state instead.
