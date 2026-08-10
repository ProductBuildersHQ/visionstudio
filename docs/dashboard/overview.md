# Dashboard Tour

The web dashboard is a single-page app served by the `visionstudio` binary (`visionstudio app start` or `visionstudio ui`). This page is a tour of the shell that wraps every panel; see the other Dashboard Guide pages for what each section shows.

## Layout

The dashboard is a sidebar plus a main content area. On load, it fetches everything it needs up front — initiatives/RMIs/repositories, spec workflows, maturity data, and token spend — and shows a full-page loading state until all four calls resolve. If any of them fail, the whole app shows a full-page error state with a retry button instead of partially rendering.

## Sidebar

The sidebar has four sections, each a top-level nav target:

- **Initiatives** 📋 — expands to show every program (each expandable to its initiatives) and a **Standalone** group for initiatives with no program. Each initiative row shows its ID, title, and completion percentage.
- **Repositories** 📦 — expands to show up to 10 repositories, sorted by RMI count descending, each with its RMI count. A "See all N repositories →" link appears if there are more than 10.
- **Maturity** 📈 — no sub-items; links straight to the Maturity page.
- **Performance** 📊 — no sub-items; links straight to the Performance page.

Click the section header (icon + label) to navigate to that section's default view; click the ▶/▼ chevron next to a program or the Standalone group to expand/collapse it without navigating.

The **collapse button** (« / ») at the top of the sidebar shrinks it to icon-only width — useful on smaller screens or when you want more room for a wide table.

### Connection status dot

Next to the "VisionStudio" wordmark at the top of the sidebar (hidden when collapsed) is a small status dot:

| Color | Meaning |
|-------|---------|
| 🟢 Green | Connected — data loaded successfully |
| 🟡 Yellow | Loading |
| 🔴 Red | Error — the API call failed |

## Common visual patterns

These appear throughout the dashboard, so it's worth knowing them once:

- **Status badges** — a colored pill showing an initiative/RMI/repository's status (`completed`, `in_progress`/`executing`, `ready`, `planned`, `proposed`, `cancelled`, or any other string, which falls back to the same style as `proposed`). Underscores are shown as spaces.
- **Progress bars** — a horizontal bar colored green at ≥100%, blue at ≥50%, yellow below that.
- **RMI type icons** — a single-character glyph shown next to each roadmap item: ★ capability, ↺ refactor, ✓ quality, ⚠ fix, ⚙ chore, ⚡ spike, • anything else.
- **Pie/donut charts** — used for status and cost/token breakdowns; hover a slice for its exact value.

## Navigating

Every panel that lists initiatives, repositories, or RMIs is clickable — clicking a row/tile navigates to its detail view (initiative detail, repository detail, or the spec viewer). Every detail view has a "← Back" link at the top rather than relying on the browser back button, though the browser back button also works since navigation goes through normal routes.
