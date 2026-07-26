# Unreleased

**Status:** In progress, not yet tagged

Work in progress since [v0.3.0](v0.3.0.md): DevX period reports. Today the
DevX dashboard only renders the latest generated snapshot; this adds the
ability to browse historical weekly/monthly/quarterly reports from the
same view, plus two new chart types needed to render them.

## Highlights

- **DevX period reports** - browse and view historical weekly/monthly/quarterly dashboards, not just the latest snapshot
- **New chart types** - pie (model breakdown) and stacked bar (period breakdowns), alongside the existing line chart
- **Path-traversal-safe daemon endpoints** - period type and label are validated before touching the filesystem

## What's New

### Features

- Period selector dropdown (weekly/monthly/quarterly) in the DevX Usage Dashboard view
- `PieChart` renderer for model-breakdown donuts and `BarChart` renderer for stacked period breakdowns; `ChartWidget` now dispatches by mark geometry instead of assuming a line chart
- `api.getDevXPeriods()` and `api.getDevXPeriodDashboard()`, plus a `DevXPeriodEntry` type and `stack`/`value`/`name` encode fields on chart marks

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/devx/periods` | List available period reports |
| GET | `/api/devx/reports/{periodType}/{label}` | Serve a specific period report |

Both `periodType` and `label` are validated against path traversal before
being used to build a filesystem path.

### Tests

- `validPeriodType` and `validPeriodLabel` path-traversal-prevention tests

## Breaking Changes

None so far.

## Full Changelog

See [CHANGELOG.md](../../CHANGELOG.md) for the complete list of changes.
