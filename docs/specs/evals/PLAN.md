# VisionStudio Evaluation System Implementation Plan

## Overview

Implementation plan for the 1-5 integer evaluation system, targeting v0.2.0 release.

## Phases

### Phase 1: Schema & Backend (visionspec)

**Duration:** 1 week

**Tasks:**

1. Define v2 EvalResult types in `pkg/types/eval.go`
   - OverallScore struct
   - DimensionScore struct
   - Finding struct
   - Severity enum

2. Create reason code registry in `pkg/eval/codes.go`
   - ReasonCode struct
   - Registry with initial codes
   - Lookup functions

3. Update rubric profiles in `pkg/profiles/`
   - Add dimension definitions to each profile
   - Define pass thresholds
   - Define blocking codes

4. Update evaluation engine in `pkg/eval/`
   - Modify LLM prompt for 1-5 scoring
   - Add structured output schema
   - Implement dimension extraction

5. Add migration utility
   - v1 to v2 schema conversion
   - Batch migration command

**Deliverables:**

- `pkg/types/eval.go` - v2 types
- `pkg/eval/codes.go` - reason code registry
- `pkg/profiles/*/rubric.yaml` - updated rubrics
- `cmd/visionspec/migrate-evals` - migration command

### Phase 2: API Updates (visionstudio daemon)

**Duration:** 3 days

**Tasks:**

1. Update `pkg/api/types.go`
   - Add v2 EvalResult types
   - Add DimensionScore types

2. Update daemon handlers
   - `handleGetSpec` - return v2 eval format
   - `handleEvaluateSpec` - use v2 evaluation

3. Add reason code endpoint
   - `GET /api/reason-codes` - list all codes
   - `GET /api/reason-codes/{code}` - code details

**Deliverables:**

- Updated API types
- v2-compatible handlers
- Reason code API

### Phase 3: UI Components (visionstudio frontend)

**Duration:** 1 week

**Tasks:**

1. Create evaluation components
   - `DimensionScoreCard.tsx`
   - `ScoreBadge.tsx`
   - `ReasonCodeChip.tsx`
   - `FindingsList.tsx`
   - `ConfidenceIndicator.tsx`

2. Update EvalPanel
   - Display per-dimension scores
   - Show blocking issues prominently
   - Add recommendations section

3. Update FindingsView
   - Group by dimension
   - Filter by severity
   - Filter by reason code

4. Add score visualization
   - Dimension score bar chart
   - Overall score gauge
   - Pass/fail banner

**Deliverables:**

- New UI components
- Updated EvalPanel
- Updated FindingsView

### Phase 4: Testing & Documentation

**Duration:** 3 days

**Tasks:**

1. Backend tests
   - Score consistency tests
   - Migration tests
   - Reason code validation

2. Frontend tests
   - Component unit tests
   - Integration tests

3. Documentation
   - Update API docs
   - Add reason code reference
   - Update user guide

**Deliverables:**

- Test coverage >80%
- Updated documentation

## Timeline

```
Week 1: Phase 1 (Schema & Backend)
├── Day 1-2: Types and reason codes
├── Day 3-4: Rubric updates
└── Day 5: Evaluation engine

Week 2: Phase 2 + Phase 3 Start
├── Day 1-3: API updates
└── Day 4-5: UI components start

Week 3: Phase 3 + Phase 4
├── Day 1-3: UI components finish
└── Day 4-5: Testing & docs
```

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| LLM scoring inconsistency | Add retry logic, temperature=0, structured output |
| Migration breaks existing evals | Keep v1 support, gradual migration |
| UI complexity increase | Progressive disclosure, collapsible sections |

## Success Criteria

- [ ] All existing evals can be viewed in v2 format
- [ ] New evaluations produce v2 schema
- [ ] Per-dimension scores visible in UI
- [ ] Reason codes displayed with descriptions
- [ ] Pass/fail clearly indicated
- [ ] Score consistency >90% on repeated evals

## Dependencies

- visionspec must release before visionstudio v0.2.0
- omnillm-core structured output support
