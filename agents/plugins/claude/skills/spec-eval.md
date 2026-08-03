---
name: spec-eval
description: Quick spec evaluation against PBHQ Lite rubrics
---

# Spec Evaluation Skill

Evaluate spec documents against PBHQ Lite workflow rubrics.

## Usage

Parse the user's input to extract:
- `initiative_id`: e.g., INIT-VISIONSTUDIO-003 (or detect from cwd)
- `spec_type`: prd, trd, plan, roadmap, or "all" (default)

## Steps

1. Find specs at `docs/specs/initiatives/{ID}/`
2. For each spec type requested:
   - Read the markdown content
   - Evaluate against rubric categories
   - Calculate weighted score
3. Report findings

## Rubric Categories

### PRD (5 categories, all required)
- Problem Definition (25%): Clear problem, user impact
- Goals/Non-Goals (20%): Measurable, scoped
- User Stories (20%): Proper format, testable AC
- Requirements (20%): Numbered, traceable
- Success Metrics (15%): Quantitative targets

### TRD (5 categories, 4 required)
- Architecture (25%): System design
- Data Model (20%): Schema, relationships
- APIs (20%): Contracts defined
- Security (20%): Auth, data protection
- Testing (15%): Coverage plan (optional)

### PLAN (5 categories, 4 required)
- Phases (25%): Logical breakdown
- Milestones (20%): Checkpoints
- Dependencies (20%): Identified
- Risks (20%): With mitigations
- Resources (15%): Estimates (optional)

### ROADMAP (4 categories, all required)
- RMI Structure (30%): IDs, titles
- Phase Organization (25%): Grouping
- Acceptance Criteria (25%): Testable
- Priorities (20%): Required/optional

## Output Format

```
## Spec Evaluation: {INITIATIVE_ID}

### PRD: 78/100 🟡 PARTIAL
- ✅ Problem Definition (90)
- ✅ Goals/Non-Goals (85)
- ⚠️ User Stories (65) - Missing AC for US-3
- ✅ Requirements (88)
- ❌ Success Metrics (40) - No quantitative targets

### Top Findings
1. 🔴 Add numeric targets to success metrics
2. 🟡 Complete acceptance criteria for US-3
3. 🟡 Clarify REQ-7 ("fast enough" → specific latency)

### Next Steps
- Address high-severity findings before TRD
- Consider peer review for requirements section
```

## Scoring Guide

| Score | Verdict | Meaning |
|-------|---------|---------|
| 90-100 | PASS | Ready for implementation |
| 70-89 | PASS | Minor polish needed |
| 50-69 | PARTIAL | Address findings first |
| <50 | FAIL | Major rework required |
