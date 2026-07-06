# VisionStudio Evaluation System PRD

## Overview

Redesign the LLM-as-a-Judge evaluation system to use a 1-5 integer scale with binary pass/fail gates, replacing the current 1-10 float scale which produces unreliable results with excessive granularity.

## Problem Statement

Current LLM-as-a-Judge implementations struggle with fine-grained scoring distinctions. Research (MT-Bench, Chatbot Arena) documents position bias, verbosity bias, and self-enhancement bias that make 1-10 float scores unreliable. Judges cannot meaningfully distinguish 7.2 from 7.4.

## Goals

1. **Reliable Scoring**: Use 1-5 integer scale that LLMs can consistently apply
2. **Actionable Feedback**: Per-dimension scores tell users exactly what to improve
3. **Automation-Ready**: Reason codes enable automated spec repair workflows
4. **Clear Gates**: Binary pass/fail separate from quality scores

## Non-Goals

- Real-time evaluation during editing (v0.3.0+)
- Multi-judge consensus/inter-rater reliability (v0.3.0+)
- Custom rubric authoring UI (v0.4.0+)

## User Stories

### US-1: View Evaluation Results

As a product manager, I want to see my PRD evaluation with per-dimension scores so I can understand exactly which areas need improvement.

**Acceptance Criteria:**

- Each dimension displays 1-5 score with label
- Overall pass/fail status is prominent
- Blocking issues are highlighted
- Reason codes link to specific findings

### US-2: Understand What to Fix

As an author, I want actionable reason codes instead of prose so I can quickly address evaluation findings.

**Acceptance Criteria:**

- Reason codes use controlled vocabulary (e.g., `MISSING_ACCEPTANCE_CRITERIA`)
- Each code has a human-readable description
- Codes are filterable/searchable in the UI

### US-3: Trust Evaluation Confidence

As a team lead, I want to see confidence scores so I know when to trust auto-approval vs. request human review.

**Acceptance Criteria:**

- Confidence displayed as percentage or 0.0-1.0
- Low confidence (<0.7) triggers visual indicator
- Confidence threshold is configurable per organization

### US-4: Dual-Axis Evaluation

As a reviewer, I want separate Quality and Decision scores so I can distinguish well-written specs from good decisions.

**Acceptance Criteria:**

- Quality Score (1-5): Document clarity, completeness, consistency
- Decision Score (1-5): Solution soundness, strategic alignment
- Both scores visible in evaluation panel

## Scoring Scale

| Score | Label | Description |
|-------|-------|-------------|
| 1 | Unacceptable | Fundamental issues, requires complete rewrite |
| 2 | Major Revisions | Significant gaps, substantial rework needed |
| 3 | Acceptable | Meets minimum bar, minor improvements needed |
| 4 | Good | Solid quality, few refinements suggested |
| 5 | Excellent | Exemplary, ready for approval |

## Dimensions by Spec Type

### PRD Dimensions

- Customer Problem
- Requirements Clarity
- Success Metrics
- Dependencies
- Risks
- Engineering Readiness

### TRD Dimensions

- Architecture Clarity
- API Design
- Data Model
- Security
- Scalability
- Operational Readiness

### UXD Dimensions

- User Flows
- Accessibility
- Error Handling
- Design Rationale
- Consistency

## Reason Codes

| Code | Description |
|------|-------------|
| `AMBIGUOUS_REQUIREMENT` | Requirement lacks specificity |
| `MISSING_ACCEPTANCE_CRITERIA` | User story has no testable criteria |
| `UNMEASURABLE_SUCCESS_METRIC` | Metric cannot be objectively measured |
| `MISSING_USER_PERSONA` | No defined target user |
| `INCOMPLETE_ERROR_HANDLING` | Error cases not addressed |
| `SECURITY_GAP` | Security consideration missing |
| `SCALABILITY_CONCERN` | Scale requirements unclear |
| `MISSING_DEPENDENCY` | External dependency not documented |

## Success Metrics

| Metric | Target |
|--------|--------|
| Score consistency (same input, same output) | >90% agreement |
| User satisfaction with feedback actionability | >4.0/5.0 |
| Time to understand evaluation results | <30 seconds |
| Automated repair success rate | >70% for common issues |

## Dependencies

- visionspec evaluation engine changes
- Updated eval JSON schema
- UI components for dimension display

## Timeline

- v0.2.0: Core 1-5 scoring, per-dimension display, reason codes
- v0.3.0: Confidence values, multi-judge consensus
- v0.4.0: Dual-axis scoring, custom rubric UI
