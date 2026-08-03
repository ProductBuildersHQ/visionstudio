# PLAN: Workflow-Scoped Spec MCP Server

**Initiative:** INIT-VISIONSTUDIO-003  
**Status:** Draft  
**Date:** 2026-08-02

## Objective

Extend visionstudio's MCP server with workflow-scoped specification operations, enabling AI agents to follow structured product methodologies (AWS Working Backwards, PBHQ, etc.) when authoring specs.

## Approach

### Phase 1: Foundation (Core Tools)

**Goal:** Basic workflow selection and spec CRUD operations.

1. **Database schema**
   - Add `spec_documents` table for spec metadata
   - Add `initiative_workflow` table for workflow selection
   - Ent schema generation

2. **Service layer**
   - `pkg/service/spec_service.go` — spec operations
   - `pkg/specworkflow/resolver.go` — workflow resolution

3. **MCP tools (read-only first)**
   - `workflow_list` — list available workflows
   - `workflow_status` — show initiative's workflow state
   - `spec_list` — list specs with status

4. **MCP tools (write)**
   - `workflow_select` — activate workflow for initiative
   - `spec_create` — create spec from template
   - `spec_read` — read spec content

### Phase 2: Evaluation

**Goal:** LLM-as-judge quality gates.

1. **Integration**
   - Connect to `plexusone/structured-evaluation`
   - Rubric loading from profiles

2. **MCP tools**
   - `spec_evaluate` — run evaluation, persist results

3. **Gate enforcement**
   - Block dependent specs until source passes threshold

### Phase 3: Synthesis

**Goal:** Generate specs from sources following DAG.

1. **Synthesis engine**
   - `pkg/specworkflow/executor.go` — LLM synthesis
   - Source validation against workflow DAG

2. **MCP tools**
   - `spec_synthesize` — generate spec from sources
   - Dry-run preview mode

3. **Custom specs**
   - `spec_add` — add specs beyond workflow

### Phase 4: Agent Workflows

**Goal:** Packaged agent rules for methodologies.

1. **Rule files**
   - `workflows/vs-spec-rules/core-workflow.md`
   - `workflows/vs-spec-rule-details/{methodology}/`

2. **Claude Code workflows**
   - `workflows/spec-review.js` — orchestration scripts

## Dependencies

| Dependency | Type | Status |
|------------|------|--------|
| specification-workflow-spec | Go module | Available |
| structured-evaluation | Go module | Available |
| visionspec (templates/rubrics) | Go module | Future integration |
| Dolt store | Infrastructure | Running |

## Risks

| Risk | Mitigation |
|------|------------|
| LLM evaluation cost | Rate limiting, caching, model tier options |
| Synthesis quality | Structured prompts, human review gates |
| Workflow complexity | Start with 3 built-in workflows, extend later |

## Success Criteria

1. **Phase 1 complete**: Agents can select workflow, create/read specs
2. **Phase 2 complete**: Specs are evaluated before proceeding
3. **Phase 3 complete**: FAQ can be synthesized from MRD + Press
4. **Phase 4 complete**: AWS Working Backwards flow is fully automated

## Timeline

- Phase 1: 1 week
- Phase 2: 1 week  
- Phase 3: 1 week
- Phase 4: 1 week

Total: ~4 weeks
