// Package speceval provides LLM-as-judge evaluation for spec documents.
package speceval

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/specification-workflow-spec/pkg/workflows"
	"github.com/plexusone/structured-evaluation/rubric"
)

// EvaluationResult is the outcome of evaluating a spec document.
type EvaluationResult struct {
	SpecType    string           `json:"spec_type"`
	Score       int              `json:"score"`
	Verdict     string           `json:"verdict"`
	Findings    []Finding        `json:"findings,omitempty"`
	Categories  []CategoryResult `json:"categories,omitempty"`
	Rationale   string           `json:"rationale,omitempty"`
	Model       string           `json:"model,omitempty"`
	EvaluatedAt time.Time        `json:"evaluated_at"`
}

// Finding is a specific issue found during evaluation.
type Finding struct {
	Severity string `json:"severity"`
	Section  string `json:"section,omitempty"`
	Message  string `json:"message"`
}

// CategoryResult is the evaluation of a single rubric category.
type CategoryResult struct {
	Name      string `json:"name"`
	Score     int    `json:"score"`
	Verdict   string `json:"verdict"`
	Rationale string `json:"rationale,omitempty"`
}

// Evaluator evaluates spec documents against rubrics using an LLM.
type Evaluator struct {
	llmClient LLMClient
	rubrics   map[string]*rubric.RubricSet
}

// LLMClient is the interface for making LLM calls.
type LLMClient interface {
	Complete(ctx context.Context, prompt string, model string) (string, error)
}

// NewEvaluator creates an Evaluator with the given LLM client.
func NewEvaluator(client LLMClient) *Evaluator {
	return &Evaluator{
		llmClient: client,
		rubrics:   make(map[string]*rubric.RubricSet),
	}
}

// NewEvaluatorWithWorkflow creates an Evaluator pre-loaded with rubrics from a workflow.
func NewEvaluatorWithWorkflow(client LLMClient, workflowID string) (*Evaluator, error) {
	e := NewEvaluator(client)
	loaded, err := workflows.DefaultLoader().Load(workflowID)
	if err != nil {
		return nil, fmt.Errorf("load workflow %q: %w", workflowID, err)
	}
	for specType, r := range loaded.Rubrics {
		e.rubrics[specType] = r
	}
	return e, nil
}

// RegisterRubric registers a rubric for a spec type.
func (e *Evaluator) RegisterRubric(specType string, r *rubric.RubricSet) {
	e.rubrics[specType] = r
}

// Evaluate evaluates a spec document against its rubric.
func (e *Evaluator) Evaluate(ctx context.Context, specType, content, model string) (*EvaluationResult, error) {
	r, ok := e.rubrics[specType]
	if !ok {
		return nil, fmt.Errorf("no rubric registered for spec type: %s", specType)
	}

	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	prompt := buildEvaluationPrompt(r, content)
	response, err := e.llmClient.Complete(ctx, prompt, model)
	if err != nil {
		return nil, fmt.Errorf("llm evaluation failed: %w", err)
	}

	result, err := parseEvaluationResponse(specType, response)
	if err != nil {
		return nil, fmt.Errorf("parse evaluation response: %w", err)
	}

	result.Model = model
	result.EvaluatedAt = time.Now()
	return result, nil
}

func buildEvaluationPrompt(r *rubric.RubricSet, content string) string {
	if r.JudgePromptTemplate != "" {
		prompt := r.JudgePromptTemplate
		prompt = strings.ReplaceAll(prompt, "{content}", content)
		prompt = strings.ReplaceAll(prompt, "{rubric_name}", r.Name)
		return prompt
	}

	var sb strings.Builder
	sb.WriteString("You are an expert evaluator assessing a specification document.\n\n")
	sb.WriteString("## Rubric: ")
	sb.WriteString(r.Name)
	sb.WriteString("\n\n")
	sb.WriteString(r.Description)
	sb.WriteString("\n\n## Categories to Evaluate\n\n")

	for _, cat := range r.Categories {
		sb.WriteString("### ")
		sb.WriteString(cat.Name)
		sb.WriteString("\n")
		sb.WriteString(cat.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Document to Evaluate\n\n")
	sb.WriteString(content)
	sb.WriteString("\n\n## Instructions\n\n")
	sb.WriteString("Evaluate the document against each category. For each category, provide:\n")
	sb.WriteString("- score (0-100)\n")
	sb.WriteString("- verdict (pass/partial/fail)\n")
	sb.WriteString("- brief rationale\n\n")
	sb.WriteString("Also list any specific findings (issues or recommendations).\n\n")
	sb.WriteString("Respond with JSON in this exact format:\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{
  "score": <overall 0-100>,
  "verdict": "<pass|partial|fail>",
  "rationale": "<overall rationale>",
  "categories": [
    {"name": "<category>", "score": <0-100>, "verdict": "<pass|partial|fail>", "rationale": "<brief>"}
  ],
  "findings": [
    {"severity": "<critical|high|medium|low>", "section": "<optional section>", "message": "<description>"}
  ]
}
`)
	sb.WriteString("```\n")

	return sb.String()
}

func parseEvaluationResponse(specType, response string) (*EvaluationResult, error) {
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}
	jsonStr := response[jsonStart : jsonEnd+1]

	var raw struct {
		Score      int              `json:"score"`
		Verdict    string           `json:"verdict"`
		Rationale  string           `json:"rationale"`
		Categories []CategoryResult `json:"categories"`
		Findings   []Finding        `json:"findings"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	return &EvaluationResult{
		SpecType:   specType,
		Score:      raw.Score,
		Verdict:    raw.Verdict,
		Rationale:  raw.Rationale,
		Categories: raw.Categories,
		Findings:   raw.Findings,
	}, nil
}

// DefaultRubrics returns built-in rubrics for common spec types.
func DefaultRubrics() map[string]*rubric.RubricSet {
	categoricalScale := rubric.Scale{
		Type: rubric.ScaleTypeCategorical,
		Options: []rubric.ScaleOption{
			{Value: "pass", Label: "Pass"},
			{Value: "partial", Label: "Partial"},
			{Value: "fail", Label: "Fail"},
		},
	}

	return map[string]*rubric.RubricSet{
		"prd": {
			ID:             "prd-rubric",
			Name:           "Product Requirements Document",
			Version:        "1.0.0",
			Description:    "Evaluates PRDs for completeness, clarity, and technical feasibility.",
			EvaluationType: rubric.EvaluationTypeAnalytic,
			PassCriteria: rubric.RubricPassCriteria{
				MinCategoriesPassing: "all_required",
				ScoreThresholds:      &rubric.ScoreThresholds{Pass: 85, Partial: 70},
			},
			Categories: []rubric.Category{
				{ID: "problem", Name: "Problem Statement", Description: "Clear articulation of the problem being solved", Required: true, Weight: 20, Scale: categoricalScale},
				{ID: "stories", Name: "User Stories", Description: "Complete user stories covering all personas", Required: true, Weight: 25, Scale: categoricalScale},
				{ID: "reqs", Name: "Requirements", Description: "Specific, measurable, achievable requirements", Required: true, Weight: 25, Scale: categoricalScale},
				{ID: "metrics", Name: "Success Metrics", Description: "Defined success criteria and KPIs", Required: true, Weight: 15, Scale: categoricalScale},
				{ID: "edge", Name: "Edge Cases", Description: "Coverage of edge cases and error scenarios", Required: false, Weight: 15, Scale: categoricalScale},
			},
		},
		"trd": {
			ID:             "trd-rubric",
			Name:           "Technical Requirements Document",
			Version:        "1.0.0",
			Description:    "Evaluates TRDs for technical completeness and implementation clarity.",
			EvaluationType: rubric.EvaluationTypeAnalytic,
			PassCriteria: rubric.RubricPassCriteria{
				MinCategoriesPassing: "all_required",
				ScoreThresholds:      &rubric.ScoreThresholds{Pass: 85, Partial: 70},
			},
			Categories: []rubric.Category{
				{ID: "arch", Name: "Architecture", Description: "Clear system architecture and component design", Required: true, Weight: 25, Scale: categoricalScale},
				{ID: "data", Name: "Data Model", Description: "Complete data model with relationships", Required: true, Weight: 20, Scale: categoricalScale},
				{ID: "api", Name: "APIs", Description: "Well-defined API contracts and interfaces", Required: true, Weight: 20, Scale: categoricalScale},
				{ID: "security", Name: "Security", Description: "Security considerations and mitigations", Required: true, Weight: 20, Scale: categoricalScale},
				{ID: "testing", Name: "Testing Strategy", Description: "Test plan covering unit, integration, E2E", Required: false, Weight: 15, Scale: categoricalScale},
			},
		},
		"press": {
			ID:             "press-release-rubric",
			Name:           "Press Release",
			Version:        "1.0.0",
			Description:    "Evaluates press releases for clarity and customer focus.",
			EvaluationType: rubric.EvaluationTypeAnalytic,
			PassCriteria: rubric.RubricPassCriteria{
				MinCategoriesPassing: "all_required",
				ScoreThresholds:      &rubric.ScoreThresholds{Pass: 80, Partial: 65},
			},
			Categories: []rubric.Category{
				{ID: "benefit", Name: "Customer Benefit", Description: "Clear articulation of customer value", Required: true, Weight: 30, Scale: categoricalScale},
				{ID: "narrative", Name: "Problem-Solution", Description: "Compelling problem-solution narrative", Required: true, Weight: 25, Scale: categoricalScale},
				{ID: "clarity", Name: "Clarity", Description: "Clear, jargon-free language", Required: true, Weight: 25, Scale: categoricalScale},
				{ID: "cta", Name: "Call to Action", Description: "Clear next steps for the reader", Required: false, Weight: 20, Scale: categoricalScale},
			},
		},
	}
}
