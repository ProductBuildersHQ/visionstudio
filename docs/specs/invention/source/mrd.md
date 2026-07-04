# Market Requirements Document: VisionApp

## 1. Executive Summary

VisionApp is a desktop application that provides an LLM-powered workspace for authoring, evaluating, and iterating on product specifications using the VisionSpec methodology. It combines structured spec workflows with AI-assisted writing and LLM-as-a-Judge evaluation to help product teams create high-quality specifications efficiently.

## 2. Problem Statement

### Current State

Product teams creating specifications face several challenges:

1. **Fragmented tooling** - Specs live in Google Docs, Notion, Confluence with no workflow awareness
2. **Quality inconsistency** - No objective evaluation of spec completeness or quality
3. **Manual iteration** - Teams manually review specs without structured feedback
4. **Methodology drift** - Teams start with a methodology (Working Backwards, Shape Up) but abandon rigor under time pressure
5. **Context switching** - Writing specs requires jumping between editor, research, and review

### Pain Points

- "I wrote a PRD but I'm not sure if it's good enough"
- "We started using Working Backwards but stopped doing the FAQ"
- "I have to copy-paste between my spec and ChatGPT to get feedback"
- "Our specs are inconsistent across teams"

## 3. Target Users

### Primary: Product Managers

- Write MRDs, PRDs, and coordinate spec workflows
- Need structured guidance on methodology
- Want objective quality feedback before stakeholder review

### Secondary: Technical Writers / Technical PMs

- Author TRDs and technical specifications
- Need templates and evaluation for technical completeness
- Want to ensure alignment with PRD requirements

### Tertiary: Engineering Leads

- Review and approve TRDs
- Need visibility into spec status and quality
- Want to trace technical decisions to product requirements

## 4. Market Context

### Existing Solutions

| Solution | Gap |
|----------|-----|
| Google Docs / Notion | No workflow, no evaluation, no LLM integration |
| ChatGPT / Claude.ai | No structure, no persistence, no methodology |
| Linear / Jira | Task tracking, not spec authoring |
| Coda / Airtable | Flexible but no spec-specific workflows |

### Why Now

- LLMs capable of meaningful evaluation (Claude 3.5+, GPT-4o+)
- Teams adopting AI for writing assistance
- Remote work increased need for clear written specs
- VisionSpec library provides foundation (profiles, rubrics, synthesis)

## 5. Business Goals

### OKRs

**Objective:** Establish VisionApp as the standard for AI-assisted spec authoring

- **KR1:** 100 active users within 6 months of launch
- **KR2:** 80% of users complete full workflow (MRD → TRD) at least once
- **KR3:** Average spec evaluation score improves 20% from first to final draft
- **KR4:** NPS > 50 among active users

### Success Metrics

| Metric | Target | Rationale |
|--------|--------|-----------|
| Specs completed per user/month | > 2 | Demonstrates regular usage |
| Evaluation-to-revision rate | > 60% | Users act on LLM feedback |
| Profile adoption | > 3 profiles used | Validates methodology library value |
| Session duration | 15-45 min | Focused work sessions |

## 6. Solution Overview

### Core Capabilities

1. **Project Management**
   - Create/manage multiple spec projects
   - Each project is a directory with structured spec files
   - Projects can be product-level (MRD start) or feature-level (OpportunitySpec start)

2. **Profile-Driven Workflows**
   - Select from profiles: aws-product, aws-feature, big-tech-product, big-tech-feature, etc.
   - Visual workflow diagram showing spec sequence and dependencies
   - Status indicators: not started, draft, evaluated, approved

3. **Spec Authoring**
   - Markdown editor with profile-specific templates
   - Toggle between Source (Markdown) and Rendered view
   - Template variables auto-populated from project context

4. **LLM-as-a-Judge Evaluation**
   - Evaluate spec against profile rubric
   - Detailed scoring by category (e.g., Problem Clarity, User Stories, Technical Completeness)
   - Actionable feedback with specific improvement suggestions

5. **LLM Writing Assistant**
   - Chat panel for spec assistance
   - Context-aware: knows current spec, project, profile
   - Can draft sections, improve clarity, expand details

6. **Synthesis**
   - Auto-generate downstream specs from source specs
   - Example: Generate Press Release from MRD
   - Human review and edit before finalization

## 7. User Experience

### Information Architecture

```
Left Panel (Navigation + LLM)
├── Project selector
├── Profile dropdown
├── Workflow link (opens diagram)
└── Spec list (ordered by workflow)
    ├── MRD ✓ (9.2)
    ├── Press Release ✓ (8.5)
    ├── FAQ ⚠ (6.1)
    └── ...
─────────────────────────────
LLM Chat Panel
├── Message input
└── Conversation history

Main Panel (Content)
├── Workflow diagram (when Workflow selected)
└── Spec editor (when spec selected)
    ├── Toolbar: [Source | Rendered]
    └── Content area
```

### Key Interactions

1. **Start new project** → Choose profile → Project created with empty specs
2. **Open spec** → See template if empty, content if exists
3. **Edit spec** → Autosave, Markdown editing
4. **Evaluate** → Click evaluate or ask LLM → See scores and feedback
5. **Act on feedback** → Edit spec → Re-evaluate → Iterate until passing
6. **Synthesize next** → When spec passes, synthesize downstream specs

## 8. Technical Constraints

### Must Have

- **Offline-capable** - Core editing works without internet
- **Local-first** - Specs stored as local Markdown files
- **Cross-platform** - macOS, Windows, Linux
- **LLM-agnostic** - Support Claude, GPT, local models via omniagent

### Architecture Decisions

- Desktop app built with Wails (Go + TypeScript/React)
- Import visionspec for profiles, templates, rubrics, evaluation
- Import omniagent for LLM interface
- Specs stored in user's filesystem, not cloud

## 9. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| LLM evaluation quality varies | Users lose trust | Use structured rubrics, show reasoning, allow override |
| Users don't complete workflows | Low engagement | Gamification (progress indicators), guided prompts |
| Too complex for casual users | Low adoption | Sensible defaults, progressive disclosure |
| Wails platform issues | Release delays | Test on all platforms early, have Electron fallback |

## 10. Non-Goals (Out of Scope)

- Cloud sync / collaboration (future phase)
- Mobile app
- Integration with Jira/Linear (future phase)
- Custom rubric editor (use visionspec CLI)
- Multi-language support

## 11. Open Questions

1. Should we support custom profiles created in the app, or only visionspec-provided profiles?
2. How do we handle LLM API key management securely?
3. Should evaluation history be persisted for trend analysis?
4. Do we need a "team" concept, or is this purely single-user initially?

## 12. Timeline Considerations

### Phase 1: Core Experience

- Project management
- Single profile (big-tech-product)
- Markdown editor with Source/Rendered toggle
- LLM chat panel
- Basic evaluation display

### Phase 2: Full Workflow

- All profiles
- Workflow diagram
- Synthesis
- Detailed evaluation with category scores

### Phase 3: Polish

- Evaluation history
- Export to visionspec formats
- Multiple LLM provider support
