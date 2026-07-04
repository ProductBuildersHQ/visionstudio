# Editor

The spec editor provides a focused writing environment with live preview.

## View Modes

Toggle between views using the toolbar:

- **Source** - Edit raw Markdown
- **Rendered** - Preview formatted output

## Toolbar

| Control | Action |
|---------|--------|
| Source/Rendered | Toggle view mode |
| Save | Save changes to file |
| Dirty indicator (•) | Unsaved changes |

## Markdown Support

The editor supports standard Markdown:

- Headings (`#`, `##`, `###`)
- Bold (`**text**`)
- Italic (`*text*`)
- Lists (`-` or `1.`)
- Code blocks (triple backticks)
- Links (`[text](url)`)
- Tables

## Auto-Save

Changes are not auto-saved. Click **Save** or use keyboard shortcut to persist changes.

## LLM Assistance

Use the LLM panel to:

- Ask questions about your spec
- Request improvements
- Get evaluation feedback

Example prompts:

- "Evaluate this MRD against the rubric"
- "Improve the problem statement section"
- "Add more specific user personas"
