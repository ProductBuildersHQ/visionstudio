# Performance

**Route:** `/performance`

Token spend and cost tracking, sourced from ingested OmniDevX usage data (`visionstudio db ingest-tokens`). A time-range selector at the top (Week / Month / Quarter / All Time) filters everything below except the Monthly History and Cost by Initiative sections, which always show their own full history.

If no token data is available, the page shows an empty state instead.

## Cost Summary

Five metric cards: total tokens, input tokens, output tokens, cache-read tokens, and total cost (highlighted).

## Model Breakdown

Three donut charts, shown once there's per-model data: tokens by model, cost by model, and tokens by category (input/output/cache read/cache write).

## Spend Over Time

Stacked bar charts — cost by model and tokens by category — bucketed by week or month (toggle at the top of this section). Bucket labels are `W<n>` for weeks or the month abbreviation for months.

## Monthly History

A month picker (dropdown, defaulting to the most recent month) shows that month's summary (total cost, total tokens, and the input/output/cache read/cache write breakdown) alongside a cost-by-model donut for that month. Below that, a full month-over-month comparison table — tokens, cost, and percent change vs. the previous month (red for an increase, green for a decrease) — click any row to select that month above.

## Accomplishments

RMIs completed within the selected time range, most recent first: a weekly velocity bar chart (when there's more than one week of data), then the completed RMIs grouped by week, each row showing its type icon, title, ID, initiative title, per-RMI cost (if attributable), and completion date.

## Cost by Initiative

A table of every initiative with attributed spend, sorted by token count descending: initiative ID and title, total tokens, and total cost.

## Ingesting Token Data

```bash
visionstudio db ingest-tokens --since 2026-01-01 --until 2026-12-31
```

Reads events from the local OmniDevX JSONL store (`~/.plexusone/omnidevx/data` by default, override with `--omnidevx-dir`) into the database. The ingest is idempotent, so re-running it is safe.
