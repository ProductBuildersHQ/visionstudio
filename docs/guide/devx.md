# DevX Usage Dashboard

The DevX view shows your own AI-assisted development activity — sessions,
prompts, commits, tool usage, cost — sourced from
[OmniDevX](https://github.com/plexusone/omnidevx-core), the PlexusOne
developer-experience telemetry domain.

## Overview

Unlike every other sidebar view, DevX is **not project-scoped**. It sits
alongside Organization in the sidebar because OmniDevX activity is
person/machine-scoped, not tied to a single tracked spec project.

VisionStudio also doesn't compute anything here — it renders a dashboard
file that [devfolio](https://github.com/plexusone/devfolio) already
generated. This is a deliberate boundary, not a missing feature:

- VisionStudio never reads the OmniDevX event store directly
- VisionStudio never computes metrics or reports
- The Go daemon only serves `~/.plexusone/omnidevx/dashboard.json` as-is

The producer (devfolio) decides what's safe to show; VisionStudio is a
read-only file consumer.

## Accessing the Dashboard

1. Click **DevX → Usage Dashboard** in the sidebar (below Organization)
2. If no dashboard has been generated yet, you'll see the exact command
   to run

## Generating the Dashboard

The dashboard file comes from devfolio, not VisionStudio:

```bash
devfolio devx dashboard --person person:jane --days 30 \
  -o ~/.plexusone/omnidevx/dashboard.json
```

This requires events already collected into the local OmniDevX store via
`omnidevx-core`'s providers (Claude Code, Codex CLI, git, GitHub) — see
[omnidevx-core's Getting Started](https://plexusone.github.io/omnidevx-core/getting-started/)
if you haven't set that up yet.

Refresh the dashboard in VisionStudio with the **Refresh** button after
regenerating the file.

## What's Shown

The dashboard file is [dashforge](https://github.com/plexusone/dashforge)
`dashboardir.Dashboard` JSON. VisionStudio renders three widget types:

- **Metric tiles** — sessions, prompts, commits (+ AI-assisted %), tool
  calls (+ failure rate), cost, coverage
- **Charts** — daily commits/prompts, daily cost, with hover tooltips
- **Table** — per-source coverage (which sources contributed, how many
  events, collection modes, confidence)

VisionStudio renders the subset of the dashforge schema devfolio's
exporter actually emits (metric tiles, line charts, tables) — not the
full dashforge spec.

## Data Quality Warnings

If the report has period-total events (like GitHub `devx.profile.snapshot`)
that aren't yet merged into the daily rollup, devfolio's export carries a
warning. VisionStudio doesn't currently surface this warning text in the
UI — check the JSON directly, or devfolio's own CLI output, if numbers
look incomplete.

## Related

- [devfolio](https://github.com/plexusone/devfolio) — generates the
  dashboard file
- [omnidevx-core](https://github.com/plexusone/omnidevx-core) — collects
  events into the local store this is built from
- [dashforge](https://github.com/plexusone/dashforge) — the dashboard-IR
  format
