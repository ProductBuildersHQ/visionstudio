# VisionStudio Roadmap

## Overview

This roadmap outlines the planned features and enhancements for VisionStudio, organized by priority and implementation order.

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
