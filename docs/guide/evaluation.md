# Evaluation

VisionStudio uses LLM-as-a-Judge to evaluate spec quality against profile-specific rubrics.

## Evaluation Process

1. Open a spec in the editor
2. Ask the LLM panel to evaluate
3. Review scores and findings
4. Iterate on the spec
5. Re-evaluate until passing

## Scoring

Specs are scored on a 0-10 scale:

| Score | Decision | Meaning |
|-------|----------|---------|
| 8.0+ | Pass | Ready for next step |
| 6.0-7.9 | Conditional | Needs minor improvements |
| < 6.0 | Fail | Needs significant work |

## Rubric Categories

Each spec type has specific evaluation criteria:

### MRD Rubric

- Problem Definition (25%)
- Target Users (20%)
- Business Goals (20%)
- Market Context (15%)
- Document Quality (20%)

### PRD Rubric

- User Stories (25%)
- Requirements Clarity (25%)
- Acceptance Criteria (20%)
- Traceability (15%)
- Document Quality (15%)

### TRD Rubric

- Architecture Clarity (25%)
- API Design (20%)
- Data Model (20%)
- Security (15%)
- Document Quality (20%)

## Findings

Evaluation results include findings with severity levels:

| Severity | Action |
|----------|--------|
| Critical | Must fix before proceeding |
| High | Should fix |
| Medium | Consider fixing |
| Low/Info | Optional improvements |
