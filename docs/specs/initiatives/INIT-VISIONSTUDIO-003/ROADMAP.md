# ROADMAP: Workflow-Scoped Spec MCP Server

**Initiative:** `INIT-VISIONSTUDIO-003`  
**Repository:** `github.com/ProductBuildersHQ/visionstudio`  
**Status:** Draft  
**Date:** 2026-08-02

## Phase 1 — Foundation

**Theme:** Core workflow selection and spec CRUD

- [x] `RMI-VISIONSTUDIO-030` Add spec_documents and initiative_workflow Ent schemas
- [x] `RMI-VISIONSTUDIO-031` Implement SpecService in service layer
  - Depends on: `RMI-VISIONSTUDIO-030`
- [x] `RMI-VISIONSTUDIO-032` Add workflow_list MCP tool
  - Depends on: `RMI-VISIONSTUDIO-031`
- [x] `RMI-VISIONSTUDIO-033` Add workflow_select MCP tool
  - Depends on: `RMI-VISIONSTUDIO-031`
- [x] `RMI-VISIONSTUDIO-034` Add workflow_status MCP tool
  - Depends on: `RMI-VISIONSTUDIO-031`
- [x] `RMI-VISIONSTUDIO-035` Add spec_list MCP tool
  - Depends on: `RMI-VISIONSTUDIO-031`
- [x] `RMI-VISIONSTUDIO-036` Add spec_create MCP tool
  - Depends on: `RMI-VISIONSTUDIO-031`
- [x] `RMI-VISIONSTUDIO-037` Add spec_read MCP tool
  - Depends on: `RMI-VISIONSTUDIO-031`

## Phase 2 — Evaluation

**Theme:** LLM-as-judge quality gates

- [x] `RMI-VISIONSTUDIO-038` Integrate structured-evaluation for spec judging
- [x] `RMI-VISIONSTUDIO-039` Add spec_evaluate MCP tool
  - Depends on: `RMI-VISIONSTUDIO-038`
- [x] `RMI-VISIONSTUDIO-040` Persist evaluation results to spec_documents
  - Depends on: `RMI-VISIONSTUDIO-039`
- [x] `RMI-VISIONSTUDIO-041` Implement gate enforcement in workflow_status
  - Depends on: `RMI-VISIONSTUDIO-040`

## Phase 3 — Synthesis

**Theme:** Generate specs from sources following DAG

- [x] `RMI-VISIONSTUDIO-042` Implement synthesis executor with LLM generation
- [x] `RMI-VISIONSTUDIO-043` Add spec_synthesize MCP tool
  - Depends on: `RMI-VISIONSTUDIO-042`
- [x] `RMI-VISIONSTUDIO-044` Add spec_add MCP tool for custom specs
  - Depends on: `RMI-VISIONSTUDIO-043`
- [x] `RMI-VISIONSTUDIO-045` Add dry-run preview mode for synthesis
  - Depends on: `RMI-VISIONSTUDIO-043`

## Phase 4 — Agent Workflows

**Theme:** Packaged agent rules for methodologies

- [x] `RMI-VISIONSTUDIO-046` Create vs-spec-rules core workflow
- [x] `RMI-VISIONSTUDIO-047` Create aws-working-backwards rule details
  - Depends on: `RMI-VISIONSTUDIO-046`
- [x] `RMI-VISIONSTUDIO-048` Create pbhq-lite rule details
  - Depends on: `RMI-VISIONSTUDIO-046`
- [x] `RMI-VISIONSTUDIO-049` Create pbhq-standard rule details
  - Depends on: `RMI-VISIONSTUDIO-046`
- [x] `RMI-VISIONSTUDIO-050` Create Claude Code workflow scripts
  - Depends on: `RMI-VISIONSTUDIO-047`

## Phase 5 — Dashboard Spec Viewer

**Theme:** Enhanced spec viewing in visionstudio dashboard

- [x] `RMI-VISIONSTUDIO-051` Add deep links to individual specs (PRD/TRD/PLAN/ROADMAP)
- [x] `RMI-VISIONSTUDIO-052` Add Display vs Markdown toggle with HTML rendering
  - Depends on: `RMI-VISIONSTUDIO-051`
- [x] `RMI-VISIONSTUDIO-053` Add PDF download for specs
  - Depends on: `RMI-VISIONSTUDIO-052`
- [x] `RMI-VISIONSTUDIO-054` Integrate @grokify/markdown-editor for reusable markdown components
  - Depends on: `RMI-VISIONSTUDIO-052`

## Notes

- RMI IDs continue from existing visionstudio sequence (assuming 029 is last)
- Phase 1 is MVP — enables basic workflow-scoped spec authoring
- Phase 2 adds quality gates — critical for methodology compliance
- Phase 3 enables automation — agents can generate specs from sources
- Phase 4 packages the experience — drop-in agent rules like aidlc-workflows
- Phase 5 RMIs 051-053 completed in this session
