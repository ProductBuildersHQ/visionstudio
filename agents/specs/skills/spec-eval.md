---
name: spec-eval
description: Quick spec evaluation against PBHQ Lite rubrics
triggers: [spec-eval, evaluate spec, eval spec, judge spec]
---

# Spec Evaluation Skill

Evaluate a spec document against PBHQ Lite workflow rubrics.

## Usage

```
/spec-eval [initiative_id] [spec_type]
```

Examples:
- `/spec-eval INIT-VISIONSTUDIO-003` - Evaluate all specs
- `/spec-eval INIT-VISIONSTUDIO-003 prd` - Evaluate just the PRD
- `/spec-eval` - Detect initiative from current directory

## Quick Evaluation Flow

1. **Find specs** in `docs/specs/initiatives/{ID}/`
2. **Read each spec** (PRD, TRD, PLAN, ROADMAP)
3. **Score against rubric** categories
4. **Report findings** with actionable fixes

## Rubric Summary

| Spec | Key Categories |
|------|----------------|
| PRD | Problem, Goals, Stories, Requirements, Metrics |
| TRD | Architecture, Data, APIs, Security, Testing |
| PLAN | Phases, Milestones, Dependencies, Risks |
| ROADMAP | RMIs, Phases, Acceptance Criteria, Priorities |

## Scoring

- **Pass** (70+): Ready for next phase
- **Partial** (50-69): Address findings before proceeding  
- **Fail** (<50): Major rework needed

## Output

Brief summary with:
- Overall score per spec
- Top 3 findings to address
- Recommended next actions
