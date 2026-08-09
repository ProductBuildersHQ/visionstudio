# Evaluation

VisionStudio uses LLM-as-a-Judge to evaluate spec quality against profile-specific rubrics.

## Evaluation Process

1. Open a spec in the editor
2. Ask the LLM panel to evaluate
3. Review scores and findings
4. Iterate on the spec
5. Re-evaluate until passing

## Scoring

Specs are scored on a 1-5 integer scale:

| Score | Decision | Meaning |
|-------|----------|---------|
| 4-5 | Pass | Ready for next step |
| 3 | Conditional | Needs improvements |
| 1-2 | Fail | Needs significant work |

## Decision Status

Each evaluation produces a decision:

| Status | Meaning |
|--------|---------|
| `pass` | All criteria met, no blocking issues |
| `conditional` | Some issues need addressing before approval |
| `fail` | Blocking issues prevent approval |
| `human_review` | Low confidence, needs human review |

## Evaluation Report

Reports use the `rubric.Rubric` schema from `structured-evaluation`:

```json
{
  "schemaVersion": "v2",
  "reviewType": "prd",
  "intScore": 4,
  "confidence": 0.85,
  "pass": true,
  "categories": [
    {
      "category": "Problem Definition",
      "intScore": 5,
      "score": "pass",
      "confidence": 0.9,
      "reasoning": "Clear problem statement with specific pain points"
    }
  ],
  "findings": [
    {
      "id": "finding-1",
      "category": "Requirements",
      "severity": "medium",
      "title": "Missing requirement IDs",
      "description": "Add numbered requirement IDs for traceability",
      "recommendation": "Prefix each requirement with REQ-NNN"
    }
  ],
  "decision": {
    "status": "pass",
    "passed": true,
    "rationale": "All categories pass or partial, no blocking findings"
  },
  "nextSteps": {
    "rerunCommand": "vistudio eval prd",
    "immediate": [],
    "recommended": [
      {
        "action": "Add numbered requirement IDs",
        "category": "Requirements",
        "severity": "medium"
      }
    ]
  }
}
```

## Rubric Categories

Each spec type has specific evaluation criteria:

### MRD Rubric

- Problem Definition
- Target Users
- Business Goals
- Market Context
- Document Quality

### PRD Rubric

- Problem Definition
- Goals and Non-Goals
- User Stories
- Requirements
- Success Metrics

### TRD Rubric

- Architecture
- Data Model
- APIs
- Security
- Testing Strategy

### PLAN Rubric

- Phases
- Milestones
- Dependencies
- Risks
- Resources

### ROADMAP Rubric

- RMI Structure
- Phase Organization
- Acceptance Criteria
- Priorities

## Findings

Evaluation results include findings with severity levels:

| Severity | Blocking | Action |
|----------|----------|--------|
| Critical | Yes | Must fix before proceeding |
| High | Yes | Must fix before approval |
| Medium | No | Should fix |
| Low | No | Consider fixing |
| Info | No | Optional improvements |

## Next Steps

Reports include actionable next steps:

- **Immediate**: Blocking actions that must be completed before approval
- **Recommended**: Suggested improvements that would raise the score

## Eval Files

Evaluations are stored as `*.eval.json` files:

```
docs/specs/initiatives/INIT-*/evaluations/
├── prd.eval.json
├── trd.eval.json
├── plan.eval.json
└── roadmap.eval.json
```

These files use the `rubric.Rubric` schema (v2) from `structured-evaluation`.
