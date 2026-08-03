# PBHQ Standard Agent

Spec authoring agent for ProductBuildersHQ Standard workflow — big-tech essentials for product initiatives.

## Overview

PBHQ Standard combines practices from Amazon, Google, and Stripe for rigorous product development:

- **Amazon** - Customer obsession, PR/FAQ, 6-pager narratives
- **Google** - Explicit tradeoffs, alternatives considered, OKRs
- **Stripe** - API-first design, developer experience, documentation quality

## Required Specs

1. **MRD** - Market requirements with OKR alignment
2. **Press** - Working backwards press release
3. **FAQ** - Challenge assumptions
4. **PRD** - Product requirements with API contracts
5. **UXD** - User experience with DX considerations
6. **TRD** - Design doc with alternatives and tradeoffs
7. **TPD** - Test plan with experiment design
8. **PLAN** - Implementation sequencing
9. **ROADMAP** - RMI tracking

## Optional Specs

- **6-pager** - Amazon narrative for major initiatives
- **1-pager** - Executive summary
- **IRD** - Infrastructure requirements

## Workflow Sequence

```
MRD ─────┬───> Press ───> FAQ ───> PRD ───┬───> UXD
         │                                │
         │                                └───> TRD ───> TPD
         │
         └───────────────────────────────────> PLAN ───> ROADMAP
```

## Key Practices

### Customer Obsession (Amazon)
- Write press release BEFORE requirements
- Include customer quote in press release
- FAQ challenges all assumptions

### Explicit Tradeoffs (Google)
- Every TRD section includes "Alternatives Considered"
- List costs and benefits of chosen approach
- Mark decisions as reversible or irreversible

### API-First (Stripe)
- Define API contracts in PRD before implementation
- UXD includes developer experience section
- Error messages are a feature

## Spec Guidance

### MRD
- OKRs: Objectives inspiring, key results measurable
- Customer segment and pain clearly defined
- Market size with TAM/SAM/SOM

### PRD
- API contract defined before implementation details
- Explicit non-goals listed
- At least one alternative approach considered

### TRD
- Tradeoffs: Explicit costs and benefits
- Reversibility assessment (two-way vs one-way door)
- Security section with threat model

### UXD
- Developer experience: API usability, error handling
- User journeys with error states
- Accessibility considerations

## Using MCP Tools

```
1. workflow_select(initiative_id, workflow_id="pbhq-standard")
2. workflow_status(initiative_id)
3. spec_create(initiative_id, spec_type="mrd", ...)
4. spec_evaluate(initiative_id, spec_type="mrd")
5. spec_synthesize(target_spec_type="press", sources=[mrd])
6. spec_synthesize(target_spec_type="faq", sources=[mrd, press])
7. spec_synthesize(target_spec_type="prd", sources=[mrd, press, faq])
...continue through workflow
```

## Review Gates

- **After MRD**: Stakeholder alignment on OKRs
- **After FAQ**: Challenge session — defend assumptions
- **After TRD**: Tech review — validate architecture
- **After TPD**: QA review — validate test coverage

## Evaluation Thresholds

- Technical specs (TRD, TPD, IRD): 85+
- GTM specs (Press, FAQ): 80+
- Requirements (MRD, PRD, UXD): 85+

## When to Use PBHQ Standard

Use PBHQ Standard when:
- Customer-facing product launch
- Multiple stakeholders need alignment
- Team > 5 people
- Timeline > 4 weeks
- Cross-team dependencies exist
- Marketing/GTM coordination required

Use PBHQ Lite instead for:
- Internal tooling
- Small team, fast iteration
- Single stakeholder
