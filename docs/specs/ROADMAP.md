# VisionStudio Roadmap

## Overview

This roadmap outlines the planned features and enhancements for VisionStudio, organized by priority and implementation order.

## Q3 2026 - Extension System & Marketplace

### 0. Extension System Architecture (Priority: Critical)

**Status:** Complete (all 7 phases done)

Refactor existing workflows into a VS Code-style extension interface to enable pluggable workflows and a marketplace for third-party extensions.

**Goals:**

- Extract a general-purpose extension interface from existing VisionSpec, AIDLC, V2MOM, and Analytics workflows
- Enable loading extensions from a GitHub-backed marketplace
- Pressure-test the interface by refactoring all existing workflows into extensions first

**Milestones:**

#### Phase 1: Extension Types & Registry ✓

- [x] Define `ExtensionManifest`, `Extension`, `ExtensionContext` types
- [x] Build `ExtensionRegistry` singleton (register, activate, deactivate, view lookup)
- [x] Create scoped `ExtensionContext` with API proxy, project data, evaluation, events, UI notifications

#### Phase 2: Built-in Extension Wrappers ✓

- [x] Wrap VisionSpec workflow as `visionstudio.visionspec` extension
- [x] Wrap AIDLC workflow as `visionstudio.aidlc` extension
- [x] Wrap V2MOM workflow as `visionstudio.v2mom` extension
- [x] Wrap Analytics (maturity, capabilities, roadmap, devx) as `visionstudio.analytics` extension

#### Phase 3: Shared UI Toolkit ✓

- [x] Extract `SummaryCard`, `SeverityBadge`, `ScoreBadge`, `StatusBadge` from duplicated code
- [x] Extract `IssueCard`, `DimensionBar`, `QualityBadge`
- [x] Extract `LoadingState`, `ErrorState`, `EmptyState` lifecycle components
- [x] Extract `ViewHeader`, `RefreshButton` view chrome

#### Phase 4: Wire Registry into App ✓

- [x] Replace hardcoded `ActiveView` union type with dynamic dispatch from `extensionRegistry.getView()`
- [x] Create `AppContext` for shared state (activeProject, navigateToSpec, navigateToView)
- [x] Create adapter components in each extension wrapper (no more `as unknown` casts)
- [x] Lazy-load view components via `React.lazy()` in extension wrappers
- [x] Initialize and activate all extensions on app mount
- [x] Generate sidebar sections dynamically from registered extensions
- [x] Activate/deactivate extensions on project switch

#### Phase 5: Marketplace & Extension Management ✓

- [x] Create `marketplace/extensions.json` manifest in ProductBuildersHQ repo
- [x] Build `MarketplaceClient` with caching, listing, search
- [x] Add extension management UI panel (installed vs. available, view links, marketplace browser)
- [x] Extension install/uninstall flow (localStorage persistence, confirmation UI, state refresh)

#### Phase 6: api-style-spec Extension ✓

- [x] Dashboard view: evaluation scores, category breakdown, severity dots, next steps
- [x] Findings explorer: tree view grouped by category/severity, expandable groups with IssueCard
- [x] Lint results view: violation list with severity mapping, summary cards, spec/profile metadata
- [x] Rules browser: category filter, expandable rules with rationale, good/bad examples
- [x] Register as `plexusone.api-style-spec` marketplace extension
- [x] TypeScript types matching Go `EvaluationReport`, `LintReport`, `StyleProfile` structs
- [x] Inline editor: OpenAPI spec editor with line-numbered gutter, severity markers, violation panel
- [x] AI-assisted fix: suggest fixes via `api-style lint --suggest-fixes`, preview diff, apply changes
- [x] Live linting: debounced re-lint on content/profile changes (800ms)
- [x] Backend API endpoints: POST /lint, GET /profiles, GET /profiles/{name}, POST /suggest-fixes

#### Phase 7: Refactor Existing Components ✓

- [x] Migrate FindingsView to use toolkit `SummaryCard`, `SeverityDot`, `IssueCard`
- [x] Migrate V2MOMView to use toolkit `LoadingState`, `ErrorState`, `EmptyState`
- [x] Migrate AIDLC views (WorkflowView, SyncPanel) to use toolkit `LoadingState`, `ErrorState`
- [x] Migrate analytics views (MaturityModel, Capabilities, Roadmap, DevX) to use toolkit `LoadingState`, `ErrorState`
- [x] Update existing views to use `useApp()` context — adapter pattern removed from extension wrappers

## Q3 2026 - Strategic Planning & Workflow Integration

### 1. Organization-Level V2MOMs (Priority: High)

**Status:** In Progress

Add workspace/organization concept with top-level V2MOMs that cascade to projects.

**Goals:**

- Enable company-wide strategic alignment via V2MOM cascade
- Support multi-level hierarchy: Organization → Department → Team → Project
- Allow projects to inherit and align with parent V2MOMs

**Key Features:**

- Workspace/Organization entity with metadata
- Organization-level V2MOM storage (`~/.visionstudio/org/v2moms/`)
- Project linking to organization V2MOMs
- Cascade visualization across organization → projects
- Alignment scoring between org methods and project methods

**Technical Changes:**

- Add `/api/organization/v2moms` endpoints
- Extend V2MOM types with organization scope
- Update frontend with organization selector
- Add alignment validation between levels

### 2. AIDLC Workflow Enhancements (Priority: High)

**Status:** Planned

Improve the AWS AI DLC workflow integration.

**Goals:**

- Enhance phase transition UX
- Add document templates and scaffolding
- Improve sync between visionspec specs and AIDLC docs
- Add quality gates for phase advancement

**Key Features:**

- Interactive phase transition wizard
- Auto-scaffolding of AIDLC documents from spec content
- Bidirectional sync with conflict resolution UI
- Phase readiness dashboard
- Quality score thresholds for phase gates

### 3. SpecKit Implementation Methodology (Priority: Medium)

**Status:** Planned

Implement the SpecKit workflow methodology option.

**Goals:**

- Support GitHub SpecKit spec structure
- Enable SpecKit export/sync
- Integrate SpecKit validation

**Key Features:**

- SpecKit document templates
- Export to SpecKit format
- SpecKit-specific workflow phases
- Integration with GitHub Issues/Projects

## Q4 2026 - Enterprise Features

### 4. Multi-Project Dashboard

**Status:** Backlog

Unified view across all projects in a workspace.

### 5. Team Collaboration

**Status:** Backlog

Real-time collaboration features for spec editing.

### 6. CI/CD Integration

**Status:** Backlog

Integrate spec validation into CI/CD pipelines.

## Completed

### Dual-Methodology Selection (July 2026)

- Requirements methodology selection (AWS WB, Lean Startup, etc.)
- Implementation methodology selection (AIDLC, SpecKit, None)
- Dynamic sidebar menu based on methodology
- Project methodology configuration persistence
