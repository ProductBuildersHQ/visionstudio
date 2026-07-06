# VisionStudio Evaluation System Roadmap

## Vision

Transform LLM-as-a-Judge evaluation into a reliable, actionable system that provides clear quality signals for specification documents while enabling automated repair workflows.

## Releases

### v0.2.0 - Core 1-5 Evaluation System

**Theme:** Reliable scoring with actionable feedback

**Features:**

- 1-5 integer scoring scale
- Per-dimension scoring (5-8 dimensions per spec type)
- Binary pass/fail gate
- Reason codes with controlled vocabulary
- Updated EvalPanel with dimension breakdown
- Updated FindingsView with filtering

**Schema:**

```json
{
  "schemaVersion": "v2",
  "overall": { "score": 4, "pass": true },
  "dimensions": [
    { "id": "requirements_clarity", "score": 3, "reasonCodes": ["AMBIGUOUS_REQUIREMENT"] }
  ]
}
```

---

### v0.3.0 - Confidence & Multi-Judge

**Theme:** Trustworthy evaluation with calibrated confidence

**Features:**

- Confidence scores (0.0-1.0) on overall and per-dimension
- Low-confidence visual indicators
- Multi-judge evaluation (3 runs, majority vote)
- Inter-rater reliability metrics
- Configurable confidence thresholds
- Auto-approve above confidence threshold

**Schema Additions:**

```json
{
  "overall": { "score": 4, "pass": true, "confidence": 0.91 },
  "dimensions": [
    { "id": "requirements_clarity", "score": 3, "confidence": 0.82 }
  ],
  "judgeRuns": 3,
  "agreement": 1.0
}
```

---

### v0.4.0 - Dual-Axis Evaluation

**Theme:** Separate document quality from decision quality

**Features:**

- Quality Score (1-5): Clarity, completeness, consistency
- Decision Score (1-5): Solution soundness, strategic alignment
- Independent evaluation prompts for each axis
- UI showing both axes
- Filter/sort by either axis

**Schema Additions:**

```json
{
  "quality": { "score": 5, "confidence": 0.94 },
  "decision": { "score": 3, "confidence": 0.71 }
}
```

---

### v0.5.0 - Automated Repair

**Theme:** AI-assisted spec improvement

**Features:**

- "Fix This" button on findings
- LLM-generated repair suggestions
- One-click apply repair
- Repair history tracking
- Before/after diff view

**User Flow:**

1. View evaluation with reason code `MISSING_ACCEPTANCE_CRITERIA`
2. Click "Fix This"
3. LLM suggests: "Add: Given X, When Y, Then Z"
4. User reviews and applies
5. Re-evaluate to confirm fix

---

### v0.6.0 - Custom Rubrics

**Theme:** Organization-specific evaluation criteria

**Features:**

- Rubric editor UI
- Custom dimension definitions
- Custom reason codes
- Custom pass thresholds
- Rubric versioning
- Rubric sharing/export

**Schema:**

```yaml
# custom-rubric.yaml
name: "Acme Corp PRD Rubric"
version: "1.0"
dimensions:
  - id: compliance
    name: Regulatory Compliance
    weight: 2.0
    criteria:
      - GDPR considerations documented
      - SOC2 requirements addressed
```

---

### v0.7.0 - Evaluation Analytics

**Theme:** Track evaluation quality over time

**Features:**

- Evaluation score trends
- Pass rate by spec type
- Common reason codes dashboard
- Team/author comparison
- Time-to-pass metrics
- Improvement velocity

---

## Future Considerations

### Cross-Spec Consistency Evaluation

Evaluate consistency across related specs:
- PRD ↔ UXD: Do user flows match requirements?
- PRD ↔ TRD: Does architecture support all features?
- MRD ↔ Press Release: Do claims match market analysis?

### Real-Time Evaluation

- Evaluate as you type (debounced)
- Inline suggestions
- Auto-complete for common patterns

### Evaluation Plugins

- Custom evaluation backends (not just LLM)
- Static analysis integration
- External validator hooks

---

## Version Mapping

| Version | Theme | Key Capability |
|---------|-------|----------------|
| v0.2.0 | Core 1-5 | Reliable integer scoring |
| v0.3.0 | Confidence | Calibrated trust signals |
| v0.4.0 | Dual-Axis | Quality vs Decision separation |
| v0.5.0 | Repair | Automated fix suggestions |
| v0.6.0 | Custom Rubrics | Organization-specific criteria |
| v0.7.0 | Analytics | Evaluation trends and metrics |
