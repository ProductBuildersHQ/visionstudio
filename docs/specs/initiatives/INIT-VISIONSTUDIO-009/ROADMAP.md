# ROADMAP: Authentic AWS Working Backwards + Web Creation Surface

**Initiative:** `INIT-VISIONSTUDIO-009`
**Repository:** `github.com/ProductBuildersHQ/visionstudio`
**Status:** Delivery Complete (retroactive log)
**Date:** 2026-08-14

Retroactive log of the follow-on work delivered 2026-08-10 through 2026-08-13,
after INIT-VISIONSTUDIO-005 closed. Cross-repo: specification-workflow-spec
(profiles, templates, rubrics), visionspec (D2 diagrams, aws.md), visionstudio
(API, UI, migration machinery).

## Phase 1 — Web Creation Surface

**Theme:** The dashboard's first mutation: create an initiative with catalog workflow selection

- [x] `RMI-VISIONSTUDIO-510` POST /api/initiatives — first mutation endpoint (apitypes request/response, loader-validated workflow, 409 on duplicate, handler unit tests)
- [x] `RMI-VISIONSTUDIO-511` CreateInitiativeModal + "+ New Initiative" buttons on all overview surfaces (workflow dropdown with required-docs preview, type-driven workflow defaults, program pre-selection)
  - Depends on: `RMI-VISIONSTUDIO-510`

## Phase 2 — PR-First Authentic Flows

**Theme:** The human-authored Press Release becomes the founding artifact in both AWS profiles

- [x] `RMI-VISIONSTUDIO-512` Reorder both AWS profiles PR-first: press human-authored (explicit empty-sources synthesis override of enterprise's press←mrd), faq←press, MRD/OpportunitySpec demoted to optional post-FAQ deepening
- [x] `RMI-VISIONSTUDIO-513` Rewrite visionspec D2 flows + regenerate SVGs + aws.md (PR/FAQ start, decision-gate diamond, iterate back-edge)
  - Depends on: `RMI-VISIONSTUDIO-512`
- [x] `RMI-VISIONSTUDIO-514` Iteration + gates: iteration_trigger on the PR/FAQ, prfaq_review gates in both profiles, required decision_meeting gate for the one-way door
- [x] `RMI-VISIONSTUDIO-515` AI-native evidence machinery: evidence-appendix section in press templates (artifacts labeled by what they prove), spike-backed feasibility prompts in FAQ templates, "AI-Native Working Backwards" division-of-labor doc section

## Phase 3 — Door Naming & Scale Decoupling

**Theme:** Profiles named by Amazon's actual selection criterion — decision reversibility

- [x] `RMI-VISIONSTUDIO-516` Rename aws-product/aws-feature → aws-one-way-door/aws-two-way-door across specification-workflow-spec (dirs, pipeline, spectype, loops, tests), visionspec (diagrams, decision tree, docs), and visionstudio (CLI help, docs, changelog)
- [x] `RMI-VISIONSTUDIO-517` Decouple scale from door: both deepening docs (MRD + OpportunitySpec) optional in both profiles — the door sets the ceremony, the scale picks the tool
  - Depends on: `RMI-VISIONSTUDIO-516`
- [x] `RMI-VISIONSTUDIO-518` Door-rename migration: retiredRemap entries + regression test (guarding remap keys against bulk-edit clobbering) + live DB remap via workflow sync
  - Depends on: `RMI-VISIONSTUDIO-516`

## Phase 4 — Leadership Principles Infusion

**Theme:** The 16 Amazon Leadership Principles grounded in templates and judge rubrics from the Jassy transcripts

- [x] `RMI-VISIONSTUDIO-519` Full 16-principle set with canonical amazon.jobs reference in both profiles' methodology metadata (transcript-derived teachings in each description)
- [x] `RMI-VISIONSTUDIO-520` LP template guidance: canonical FAQ vetting questions, disconfirmation discipline, Kozmo economic-sustainability lesson, personal-money test, hard-to-fake-details narrative rationale
- [x] `RMI-VISIONSTUDIO-521` LP rubric categories: press economic_sustainability, FAQ disconfirmation_rigor (required), 6-pager ownership_horizon; door-differentiated decision_reversibility criteria in PRD rubrics; TestLeadershipPrincipleGrounding conformance suite
  - Depends on: `RMI-VISIONSTUDIO-519`

## Phase 5 — Template & Rubric Viewing

**Theme:** Definition tab answers "what should this document contain and how will it be judged?"

- [x] `RMI-VISIONSTUDIO-522` GET /api/workflows/{id}/specs/{type} — template + rubric from the embedded catalog (WorkflowSpecDetail apitype, case-insensitive spec type, handler tests)
- [x] `RMI-VISIONSTUDIO-523` SpecTypeDetailModal (Template | Judge Rubric tabs, structured rubric renderer) wired to workflow-diagram box clicks and spec-viewer buttons
  - Depends on: `RMI-VISIONSTUDIO-522`
