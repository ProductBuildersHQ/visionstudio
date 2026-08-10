# Spec Viewer and Evaluation

## Spec Viewer

**Routes:** `/initiative/:initiativeId/spec/:specType`, `/initiative/:initiativeId/spec` (redirects to the first spec in workflow order)

Tabs across the top list every spec file found for the initiative, ordered PRD → TRD → PLAN → ROADMAP first, then any other spec types. Selecting a tab navigates to that spec's URL, so links to a specific spec are shareable.

### Display vs. Markdown

A toggle in the header switches between:

- **Display** — the spec rendered as formatted Markdown.
- **Markdown** — the raw Markdown source.

The current mode is reflected in the URL (`?mode=markdown`), so a shared link preserves it.

### Actions

- **Copy Link** — copies the current page's full URL (including the spec type and mode) to your clipboard.
- **Download PDF** — opens a print-formatted version of the currently displayed spec in a new tab and triggers your browser's print dialog; choose "Save as PDF" there to export. The exported page includes the initiative ID, spec type, generation date, file path, and last-modified date as a header.

Below the tabs, the file's path (last 3 path segments) and last-modified date are shown, followed by the rendered/source content in a scrollable panel.

If the initiative has no spec files on disk yet, the panel shows an empty state pointing at `docs/specs/initiatives/<id>/`.

## LLM-as-a-Judge Evaluation

Specs are evaluated against workflow rubrics (`structured-evaluation/rubric.Rubric`): a 1–5 integer score, an overall pass/fail decision, per-category scores and reasoning, findings with severity, and next steps. Results are stored as `*.eval.json` files under `docs/specs/initiatives/<id>/evaluations/` and shown:

- Inline in the [initiative detail Definition tab](initiatives.md#definition-details-tab) — as color-coded boxes in the PBHQ Lite workflow diagram, and as an expandable list of every judge result with score and rationale.
- Score color-coding is consistent throughout the dashboard: green for a score ≥4, yellow for ≥3, red below 3 (out of 5).

### Recording an evaluation

```bash
visionstudio spec judge record <initiative-id> <spec-file> \
  --score <1-5> --rationale "<why>" --model <model-id>
```

This is typically run by an agent or reviewer after evaluating a spec against its workflow rubric, not by hand-editing the `*.eval.json` file directly.
