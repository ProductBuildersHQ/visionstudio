# Workflows

Workflows define the sequence of specs for a project based on its profile.

## Workflow Diagram

Click **Workflow** in the sidebar to see the visual workflow diagram:

- Specs are grouped by category (Source, GTM, Product, Technical)
- Arrows show dependencies
- Colors indicate status (green=pass, yellow=needs work, gray=not started)
- Click any spec to open it in the editor

## Spec Categories

### Source Specs (Human-Authored)

| Spec | Purpose |
|------|---------|
| MRD | Market Requirements - problem, audience, goals |
| OpportunitySpec | Feature-level opportunity canvas |
| PRD | Product Requirements - user stories, features |
| UXD | User Experience - flows, interactions |

### GTM Specs (LLM-Generated)

| Spec | Purpose |
|------|---------|
| Press Release | Working Backwards announcement |
| FAQ | Challenges the vision, surfaces gaps |
| 6-Pager | Executive narrative with data |

### Technical Specs (LLM-Generated)

| Spec | Purpose |
|------|---------|
| TRD | Technical Requirements - architecture, APIs |
| TPD | Test Plan - testing strategy |
| IRD | Infrastructure Requirements - deployment |

## Workflow Progression

1. Complete source specs first (MRD or OpportunitySpec)
2. Evaluate each spec before moving on
3. Synthesize GTM specs from source specs
4. Iterate until specs pass evaluation
5. Generate technical specs from approved product specs
