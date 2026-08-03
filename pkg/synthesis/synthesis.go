// Package synthesis generates spec documents from sources using LLM.
package synthesis

import (
	"context"
	"fmt"
	"strings"
)

// LLMClient is the interface for making LLM calls.
type LLMClient interface {
	Complete(ctx context.Context, prompt string, model string) (string, error)
}

// SourceDocument represents an input document for synthesis.
type SourceDocument struct {
	Type    string `json:"type"`
	Path    string `json:"path,omitempty"`
	Content string `json:"content"`
}

// SynthesisRequest contains parameters for generating a spec.
type SynthesisRequest struct {
	TargetSpecType string           `json:"target_spec_type"`
	Sources        []SourceDocument `json:"sources"`
	WorkflowID     string           `json:"workflow_id,omitempty"`
	InitiativeID   string           `json:"initiative_id,omitempty"`
	Model          string           `json:"model,omitempty"`
	DryRun         bool             `json:"dry_run,omitempty"`
}

// SynthesisResult contains the generated spec and metadata.
type SynthesisResult struct {
	TargetSpecType string           `json:"target_spec_type"`
	Content        string           `json:"content"`
	Sources        []SourceDocument `json:"sources"`
	Model          string           `json:"model"`
	TokensUsed     int              `json:"tokens_used,omitempty"`
	DryRun         bool             `json:"dry_run"`
}

// Executor generates specs from source documents using an LLM.
type Executor struct {
	llmClient LLMClient
	templates map[string]string
}

// NewExecutor creates a synthesis executor with the given LLM client.
func NewExecutor(client LLMClient) *Executor {
	return &Executor{
		llmClient: client,
		templates: DefaultTemplates(),
	}
}

// RegisterTemplate registers a custom synthesis template for a spec type.
func (e *Executor) RegisterTemplate(specType, template string) {
	e.templates[specType] = template
}

// Synthesize generates a spec document from source documents.
func (e *Executor) Synthesize(ctx context.Context, req *SynthesisRequest) (*SynthesisResult, error) {
	if req.TargetSpecType == "" {
		return nil, fmt.Errorf("target_spec_type is required")
	}
	if len(req.Sources) == 0 {
		return nil, fmt.Errorf("at least one source document is required")
	}

	template, ok := e.templates[req.TargetSpecType]
	if !ok {
		return nil, fmt.Errorf("no synthesis template for spec type: %s", req.TargetSpecType)
	}

	model := req.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	prompt := e.buildPrompt(template, req)

	if req.DryRun {
		return &SynthesisResult{
			TargetSpecType: req.TargetSpecType,
			Content:        fmt.Sprintf("[DRY RUN] Would generate %s from %d source(s)\n\nPrompt preview:\n%s", req.TargetSpecType, len(req.Sources), truncate(prompt, 2000)),
			Sources:        req.Sources,
			Model:          model,
			DryRun:         true,
		}, nil
	}

	content, err := e.llmClient.Complete(ctx, prompt, model)
	if err != nil {
		return nil, fmt.Errorf("llm synthesis failed: %w", err)
	}

	content = extractMarkdown(content)

	return &SynthesisResult{
		TargetSpecType: req.TargetSpecType,
		Content:        content,
		Sources:        req.Sources,
		Model:          model,
	}, nil
}

func (e *Executor) buildPrompt(template string, req *SynthesisRequest) string {
	var sb strings.Builder
	sb.WriteString(template)
	sb.WriteString("\n\n## Source Documents\n\n")

	for i, src := range req.Sources {
		sb.WriteString(fmt.Sprintf("### Source %d: %s", i+1, src.Type))
		if src.Path != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", src.Path))
		}
		sb.WriteString("\n\n")
		sb.WriteString(src.Content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Instructions\n\n")
	sb.WriteString("Generate the specification document based on the source documents above. ")
	sb.WriteString("Output only the markdown content, no explanations or preamble.\n")

	return sb.String()
}

func extractMarkdown(response string) string {
	if idx := strings.Index(response, "```markdown"); idx != -1 {
		start := idx + len("```markdown\n")
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}
	if idx := strings.Index(response, "```"); idx != -1 {
		start := idx + 3
		for start < len(response) && response[start] != '\n' {
			start++
		}
		start++
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}
	return strings.TrimSpace(response)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// DefaultTemplates returns built-in synthesis templates for common spec types.
func DefaultTemplates() map[string]string {
	return map[string]string{
		"prd": `# Product Requirements Document Synthesis

You are a product manager synthesizing a PRD from input documents.

Generate a complete PRD with these sections:
1. **Executive Summary** - Brief overview of the product/feature
2. **Problem Statement** - What problem are we solving and for whom
3. **Goals and Success Metrics** - Measurable outcomes
4. **User Stories** - As a [user], I want [goal] so that [benefit]
5. **Requirements** - Functional and non-functional requirements
6. **Out of Scope** - What we're explicitly not doing
7. **Dependencies and Risks** - Known blockers and risk mitigation
8. **Timeline** - High-level milestones

Derive all content from the source documents. Do not invent requirements.`,

		"trd": `# Technical Requirements Document Synthesis

You are a senior engineer synthesizing a TRD from a PRD and other sources.

Generate a complete TRD with these sections:
1. **Overview** - Technical approach summary
2. **Architecture** - System components and interactions
3. **Data Model** - Schemas, relationships, migrations
4. **API Design** - Endpoints, request/response formats
5. **Security Considerations** - Auth, encryption, compliance
6. **Testing Strategy** - Unit, integration, E2E approaches
7. **Deployment** - Infrastructure, rollout, rollback
8. **Observability** - Logging, metrics, alerts

Map each technical decision to a requirement from the PRD.`,

		"plan": `# Implementation Plan Synthesis

You are a tech lead synthesizing an implementation plan from PRD and TRD.

Generate an implementation plan with:
1. **Phases** - Logical groupings of work
2. **Tasks** - Specific deliverables within each phase
3. **Dependencies** - What blocks what
4. **Estimates** - T-shirt sizes (S/M/L/XL) per task
5. **Risk Mitigation** - Highest-risk items first

Each task should map to specific TRD sections.`,

		"roadmap": "# Roadmap Synthesis\n\n" +
			"You are a product owner synthesizing a roadmap from a plan.\n\n" +
			"Generate a ROADMAP.md with:\n" +
			"1. **Phases** with themes (Phase 1 — Foundation, etc.)\n" +
			"2. **RMI items** per phase with format: - [ ] `RMI-XXX-NNN` Title\n" +
			"3. **Dependencies** as sub-bullets: - Depends on: `RMI-XXX-001`\n\n" +
			"Use the repository slug for RMI IDs (derive from context if available).",

		"press": `# Press Release Synthesis (Working Backwards)

You are a product leader writing a press release for an unreleased product.

Write a customer-focused press release with:
1. **Headline** - Attention-grabbing, benefit-focused
2. **Subheadline** - One sentence expanding on the headline
3. **Date and Location** - "[City] — [Date]"
4. **Opening Paragraph** - Who, what, why in 2-3 sentences
5. **Problem Paragraph** - The customer pain being solved
6. **Solution Paragraph** - How the product solves it
7. **Quote from Leadership** - Vision and commitment
8. **Customer Quote** - Testimonial-style benefit statement
9. **Call to Action** - How to get started

Write for customers, not engineers. No jargon.`,

		"faq": `# FAQ Synthesis

You are a product manager synthesizing an FAQ from other docs.

Generate an FAQ with:
1. **Customer-facing questions** - What users will actually ask
2. **Internal questions** - What the team needs to know
3. **Technical questions** - Implementation details for engineers

Each answer should be 2-4 sentences. Cite the source doc where relevant.`,
	}
}
