# VisionStudio Evaluation System TRD

## Overview

Technical design for the 1-5 integer evaluation system with per-dimension scoring, reason codes, and confidence values.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        VisionStudio UI                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ EvalPanel   │  │ FindingsView│  │ DimensionScoreCard      │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      VisionStudio Daemon                        │
│  ┌─────────────────┐  ┌─────────────────┐                      │
│  │ GET /specs/:id  │  │ POST /evaluate  │                      │
│  │ /eval           │  │                 │                      │
│  └─────────────────┘  └─────────────────┘                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         VisionSpec                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ EvalEngine      │  │ RubricLoader    │  │ ReasonCodeDB    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         OmniLLM                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ Structured output with JSON schema enforcement              ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## Data Model

### EvalResult (v2 Schema)

```go
// EvalResult represents the complete evaluation of a specification.
type EvalResult struct {
    SchemaVersion string            `json:"schemaVersion"` // "v2"
    SpecType      string            `json:"specType"`
    EvaluatedAt   time.Time         `json:"evaluatedAt"`

    Overall       OverallScore      `json:"overall"`
    Quality       *AxisScore        `json:"quality,omitempty"`   // v0.4.0+
    Decision      *AxisScore        `json:"decision,omitempty"`  // v0.4.0+
    Dimensions    []DimensionScore  `json:"dimensions"`
    Blocking      []string          `json:"blocking"`
    Recommendations []string        `json:"recommendations"`

    // Metadata
    Model         string            `json:"model"`
    PromptVersion string            `json:"promptVersion"`
}

// OverallScore is the aggregate evaluation result.
type OverallScore struct {
    Score      int     `json:"score"`      // 1-5
    Pass       bool    `json:"pass"`
    Confidence float64 `json:"confidence"` // 0.0-1.0
}

// AxisScore represents Quality or Decision axis (v0.4.0+).
type AxisScore struct {
    Score      int     `json:"score"`      // 1-5
    Confidence float64 `json:"confidence"`
}

// DimensionScore is a single rubric dimension evaluation.
type DimensionScore struct {
    ID          string    `json:"id"`          // e.g., "customer_problem"
    Name        string    `json:"name"`        // e.g., "Customer Problem"
    Score       int       `json:"score"`       // 1-5
    Severity    Severity  `json:"severity"`    // none, minor, major, critical
    Confidence  float64   `json:"confidence"`
    ReasonCodes []string  `json:"reasonCodes"` // e.g., ["MISSING_USER_PERSONA"]
    Findings    []Finding `json:"findings"`
}

// Severity levels for dimension issues.
type Severity string

const (
    SeverityNone     Severity = "none"
    SeverityMinor    Severity = "minor"
    SeverityMajor    Severity = "major"
    SeverityCritical Severity = "critical"
)

// Finding is a specific issue within a dimension.
type Finding struct {
    Location string `json:"location,omitempty"` // e.g., "REQ-12", "Section 3.2"
    Message  string `json:"message"`
    Code     string `json:"code,omitempty"`     // reason code
}
```

### Reason Code Registry

```go
// ReasonCode defines a standardized evaluation finding code.
type ReasonCode struct {
    Code        string   `json:"code"`
    Category    string   `json:"category"`    // requirement, metric, security, etc.
    Description string   `json:"description"`
    Severity    Severity `json:"severity"`
    RepairHint  string   `json:"repairHint"`  // guidance for automated repair
    SpecTypes   []string `json:"specTypes"`   // applicable spec types
}

// Example registry entries
var ReasonCodeRegistry = []ReasonCode{
    {
        Code:        "AMBIGUOUS_REQUIREMENT",
        Category:    "requirement",
        Description: "Requirement lacks specificity or has multiple interpretations",
        Severity:    SeverityMajor,
        RepairHint:  "Add specific, measurable criteria to the requirement",
        SpecTypes:   []string{"prd", "mrd"},
    },
    {
        Code:        "MISSING_ACCEPTANCE_CRITERIA",
        Category:    "requirement",
        Description: "User story has no testable acceptance criteria",
        Severity:    SeverityCritical,
        RepairHint:  "Add Given/When/Then acceptance criteria",
        SpecTypes:   []string{"prd"},
    },
    // ... more codes
}
```

### Rubric Profile

```yaml
# profiles/big-tech-product/rubric.yaml
specType: prd

dimensions:
  - id: customer_problem
    name: Customer Problem
    weight: 1.5
    criteria:
      - Clear problem statement
      - Defined customer segment
      - Evidence of customer need

  - id: requirements_clarity
    name: Requirements Clarity
    weight: 1.0
    criteria:
      - Unambiguous language
      - Testable acceptance criteria
      - Prioritized requirements

passThreshold: 3  # minimum overall score to pass

blocking:
  - MISSING_ACCEPTANCE_CRITERIA
  - SECURITY_GAP
  - NO_CUSTOMER_PERSONA
```

## API Changes

### GET /api/projects/{project}/specs/{specType}/eval

Returns evaluation result in v2 schema format.

**Response:**

```json
{
  "schemaVersion": "v2",
  "specType": "prd",
  "evaluatedAt": "2026-07-06T10:30:00Z",
  "overall": {
    "score": 4,
    "pass": true,
    "confidence": 0.91
  },
  "dimensions": [
    {
      "id": "customer_problem",
      "name": "Customer Problem",
      "score": 5,
      "severity": "none",
      "confidence": 0.96,
      "reasonCodes": [],
      "findings": []
    },
    {
      "id": "requirements_clarity",
      "name": "Requirements Clarity",
      "score": 3,
      "severity": "major",
      "confidence": 0.82,
      "reasonCodes": ["AMBIGUOUS_REQUIREMENT", "MISSING_ACCEPTANCE_CRITERIA"],
      "findings": [
        {
          "location": "REQ-12",
          "message": "Acceptance criteria missing",
          "code": "MISSING_ACCEPTANCE_CRITERIA"
        }
      ]
    }
  ],
  "blocking": ["MISSING_ACCEPTANCE_CRITERIA"],
  "recommendations": [
    "Define measurable acceptance criteria for REQ-12"
  ],
  "model": "claude-sonnet-4-20250514",
  "promptVersion": "v2.1"
}
```

### POST /api/projects/{project}/specs/{specType}/evaluate

Triggers evaluation and returns result.

**Request:**

```json
{
  "rubric": "big-tech-product",  // optional, uses project profile default
  "forceRefresh": false
}
```

## UI Components

### DimensionScoreCard

Displays a single dimension with score visualization.

```tsx
interface DimensionScoreCardProps {
  dimension: DimensionScore;
  onFindingClick?: (finding: Finding) => void;
}

function DimensionScoreCard({ dimension, onFindingClick }: DimensionScoreCardProps) {
  return (
    <div className={`dimension-card severity-${dimension.severity}`}>
      <div className="dimension-header">
        <span className="dimension-name">{dimension.name}</span>
        <ScoreBadge score={dimension.score} />
      </div>
      {dimension.reasonCodes.length > 0 && (
        <div className="reason-codes">
          {dimension.reasonCodes.map(code => (
            <ReasonCodeChip key={code} code={code} />
          ))}
        </div>
      )}
      {dimension.findings.length > 0 && (
        <FindingsList findings={dimension.findings} onClick={onFindingClick} />
      )}
    </div>
  );
}
```

### ScoreBadge

Visual representation of 1-5 score.

```tsx
function ScoreBadge({ score }: { score: number }) {
  const labels = ['', 'Unacceptable', 'Major Revisions', 'Acceptable', 'Good', 'Excellent'];
  const colors = ['', 'red', 'orange', 'yellow', 'green', 'blue'];

  return (
    <span className={`score-badge score-${score}`} style={{ backgroundColor: colors[score] }}>
      {score} - {labels[score]}
    </span>
  );
}
```

## Migration

### v1 to v2 Schema Migration

```go
func MigrateEvalResultV1ToV2(v1 *EvalResultV1) *EvalResult {
    // Map 1-10 float to 1-5 integer
    score := int(math.Round(v1.Score / 2.0))
    if score < 1 {
        score = 1
    }
    if score > 5 {
        score = 5
    }

    // Map decision to pass boolean
    pass := v1.Decision == "pass"

    // Convert findings to dimensions (best effort)
    dimensions := convertFindingsToDimensions(v1.Findings)

    return &EvalResult{
        SchemaVersion: "v2",
        Overall: OverallScore{
            Score:      score,
            Pass:       pass,
            Confidence: 0.5, // unknown confidence for migrated data
        },
        Dimensions: dimensions,
        // ...
    }
}
```

## Testing

### Score Consistency Tests

```go
func TestEvalConsistency(t *testing.T) {
    spec := loadTestSpec("prd-sample.md")

    results := make([]int, 10)
    for i := 0; i < 10; i++ {
        result := evaluator.Evaluate(spec)
        results[i] = result.Overall.Score
    }

    // All scores should be identical or within 1 point
    for i := 1; i < len(results); i++ {
        diff := abs(results[i] - results[0])
        if diff > 1 {
            t.Errorf("Score variance too high: %v", results)
        }
    }
}
```

## Security Considerations

- Evaluation prompts should not be exposed to users (prompt injection risk)
- Reason codes are from controlled vocabulary, not LLM-generated
- Confidence values should be calibrated, not trusted blindly

## Dependencies

- visionspec v0.13.0+ (eval engine changes)
- omnillm-core (structured output support)
