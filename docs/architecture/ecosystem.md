# Ecosystem

VisionStudio sits at the top of the ProductBuildersHQ spec stack. This page
describes how it consumes the two layers beneath it and the rules that keep
the layering intact. The canonical cross-repo reference lives in the
[org architecture doc](https://github.com/ProductBuildersHQ/.github/blob/main/ARCHITECTURE.md).

```
visionstudio  ──▶  visionspec  ──▶  specification-workflow-spec
 (studio)           (engine)           (contract)
```

## The Three Layers

| Layer | Repository | Role |
|-------|-----------|------|
| Contract | [specification-workflow-spec](https://github.com/ProductBuildersHQ/specification-workflow-spec) | Go types, JSON Schemas, and the embedded library of 24 default workflows — configurations, markdown templates, and structured-evaluation rubrics — with composable loaders. No execution logic, no LLM dependencies. |
| Engine | [visionspec](https://github.com/ProductBuildersHQ/visionspec) | Everything that acts on the contract: project scaffolding, LLM synthesis, LLM-as-Judge evaluation, lint/drift/status/reconcile, MCP server, execution-target export, and the organization CLI framework. |
| Studio | visionstudio (this repo) | LLM-powered application for authoring, evaluating, and managing spec portfolios, with daemon/web/desktop surfaces and ent + Dolt persistence. |

## Interaction Planes

### 1. Data Plane — specification-workflow-spec, directly

Workflow browsing and rendering never touches visionspec. The methodology
picker, template editor, and rubric viewer load embedded data as Go structs:

```go
import "github.com/ProductBuildersHQ/specification-workflow-spec/pkg/workflows"

w, err := workflows.DefaultLoader().Load("aws-two-way-door")
// w.Workflow  — configuration with extends-inheritance resolved
// w.Templates — raw markdown per spec type
// w.Rubrics   — *rubric.RubricSet per spec type (structured-evaluation)
```

Organization workflow sets chain over the embedded defaults with
`workflows.NewChainLoader(workflows.NewFileLoader(dir), workflows.DefaultLoader())`
wrapped in a `ResolvingLoader`.

### 2. Execution Plane — visionspec as an imported SDK

When the user acts — scaffold a project, synthesize a spec, run an
evaluation — the daemon calls visionspec's importable packages in-process
(`pkg/eval`, `pkg/synth`, `pkg/lint`, `pkg/status`, `pkg/aidlc`, ...).

Rules of engagement:

- LLM-backed operations (synthesis, evaluation) run as **async jobs** with
  results persisted to the ent/Dolt store; HTTP handlers never block on them.
- The daemon **never shells out** to the `visionspec` binary. If a capability
  exists only inside a visionspec cobra handler, the fix is to lift it into an
  importable package there — not to exec the CLI.
- Rubrics arrive as structured-evaluation's `rubric.RubricSet` and flow into
  evaluation without conversion.

### 3. Agent Plane — MCP for external assistants

VisionStudio does not use MCP for its own calls (it is in-process Go).
visionspec's MCP server exists so external AI assistants (Claude Code, Claude
Tag) can operate on the same projects VisionStudio manages. The shared source
of truth is the on-disk project layout defined by
specification-workflow-spec's `pkg/layout`; the daemon's fsnotify watchers
pick up changes agents make.

## Version Discipline

VisionStudio pins tagged releases of both libraries — never local `replace`
directives on `main`. Release order flows up the stack: tag
specification-workflow-spec, then visionspec, then adopt here.

## Boundary Rules (Summary)

- Workflow data and definition types live in specification-workflow-spec —
  contributions of new default workflows go there.
- Execution behavior lives in visionspec — VisionStudio does not reimplement
  synthesis, evaluation, or linting.
- VisionStudio owns the experience: UI, persistence, jobs, portfolio views,
  and anything spanning multiple projects over time.
