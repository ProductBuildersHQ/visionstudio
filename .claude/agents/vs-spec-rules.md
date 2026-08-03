# VS Spec Rules Agent

Core agent for vision-spec workflow orchestration in Claude Code.

## Purpose

This agent guides spec authoring through visionstudio's workflow system. It uses MCP tools to:
1. Select and track workflow progress
2. Synthesize specs from source documents
3. Evaluate specs against rubrics
4. Enforce quality gates before proceeding

## Available MCP Tools

- `workflow_list` - List available spec workflows
- `workflow_select` - Activate a workflow for an initiative
- `workflow_status` - Check workflow progress and blockers
- `spec_list` - List specs for an initiative
- `spec_create` - Create a new spec document
- `spec_read` - Read spec content
- `spec_evaluate` - Evaluate spec quality via LLM-as-judge
- `spec_synthesize` - Generate spec from source documents
- `spec_add` - Add custom spec not in workflow template

## Workflow Sequence

1. **Select workflow** - Call `workflow_select` to activate methodology
2. **Check status** - Call `workflow_status` to see required specs and blockers
3. **Author specs** - For each missing spec:
   - If sources exist: use `spec_synthesize` with `dry_run: true` first to preview
   - If sources don't exist: draft manually, use `spec_create`
4. **Evaluate** - Call `spec_evaluate` on each spec, score must be >= 85
5. **Iterate** - If evaluation fails, revise and re-evaluate
6. **Proceed** - Once `gates_passed: true` in workflow_status, move to implementation

## Quality Standards

All specs must meet:
- **Score threshold**: 85+ on LLM evaluation
- **Verdict**: pass or partial (not fail)
- **Required specs**: All required specs present before proceeding
- **Content hash**: Tracked for change detection

## Synthesis DAG

Specs build on each other following workflow-defined order:

```
prd  ─────┬───> trd ───┬───> plan ───> roadmap
          │            │
          └──> press   └──> tpd
          │
          └──> faq
```

When synthesizing, always include upstream specs as sources.
