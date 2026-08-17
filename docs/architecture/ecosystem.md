# Ecosystem

VisionStudio sits at the top of the ProductBuildersHQ spec stack. This page
describes how it consumes the layer beneath it and the rules that keep the
layering intact. The canonical cross-repo reference lives in the
[org architecture doc](https://github.com/ProductBuildersHQ/.github/blob/main/ARCHITECTURE.md).

```
visionstudio  ──▶  visionspec
 (studio)           (contract + engine)
```

specification-workflow-spec, the former standalone home for the contract
layer, was merged into visionspec in v0.16.0 and archived. What used to be
a three-repo stack is now two.

## The Two Layers

| Layer | Repository | Role |
|-------|-----------|------|
| Contract + Engine | [visionspec](https://github.com/ProductBuildersHQ/visionspec) | Go types, JSON Schemas, and the embedded library of 25 default workflows — configurations, markdown templates, and structured-evaluation rubrics — with composable loaders, **plus** everything that acts on them: project scaffolding, LLM synthesis, LLM-as-Judge evaluation, lint/drift/status/reconcile, MCP server, execution-target export, and the organization CLI framework. |
| Studio | visionstudio (this repo) | LLM-powered application for authoring, evaluating, and managing spec portfolios, with daemon/web/desktop surfaces and ent + Dolt persistence. |

## Interaction Planes

### 1. Data Plane — visionspec's workflow catalog, directly

Workflow browsing and rendering loads embedded data as Go structs directly
from visionspec's `pkg/workflows`:

```go
import "github.com/ProductBuildersHQ/visionspec/pkg/workflows"

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
of truth is the on-disk project layout defined by visionspec's `pkg/layout`;
the daemon's fsnotify watchers pick up changes agents make.

## Version Discipline

VisionStudio pins tagged releases of visionspec — never local `replace`
directives on `main`.

## Boundary Rules (Summary)

- Workflow data and definition types live in visionspec's `pkg/workflows` —
  contributions of new default workflows go there, not into VisionStudio.
- Execution behavior lives in visionspec — VisionStudio does not reimplement
  synthesis, evaluation, or linting.
- VisionStudio owns the experience: UI, persistence, jobs, portfolio views,
  and anything spanning multiple projects over time.
