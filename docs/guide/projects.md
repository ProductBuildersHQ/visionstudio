# Projects

A project in VisionStudio represents a product or feature you're specifying.

## Project Structure

Each project is a directory containing:

```
docs/specs/{project-name}/
├── visionspec.yaml      # Project configuration
├── source/              # Human-authored specs
│   ├── mrd.md
│   ├── prd.md
│   └── uxd.md
├── gtm/                 # LLM-generated GTM docs
│   ├── press.md
│   ├── faq.md
│   └── narrative.md
├── technical/           # LLM-generated technical docs
│   ├── trd.md
│   └── tpd.md
└── eval/                # Evaluation results
```

## Project Configuration

The `visionspec.yaml` file defines project settings:

```yaml
name: my-project
description: Project description
profile: big-tech-product

llm:
  provider: anthropic
  model: claude-sonnet-4-20250514
```

## Creating a Project

1. Create the directory structure
2. Add `visionspec.yaml` with your profile
3. The project appears in VisionStudio's sidebar

## Profiles

Profiles determine which specs are required and their workflow:

| Profile | Starting Spec | Workflow |
|---------|--------------|----------|
| big-tech-product | MRD | MRD → Press → FAQ → 6-Pager → PRD → TRD |
| big-tech-feature | OpportunitySpec | OpportunitySpec → Press → FAQ → PRD → TRD |
| aws-product | MRD | MRD → Press → FAQ → 6-Pager |
