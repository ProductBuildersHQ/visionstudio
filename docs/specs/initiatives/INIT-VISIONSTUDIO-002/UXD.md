# UXD — VisionStudio Cloud

**Status: deferred by design — author before the Phase 3 gate.**

Phases 1–2 have almost no novel UX surface: the public pages reuse the
existing `@grokify/releaselog` widget and the VS-006 board layout, and the
sync experience is two CLI commands (`visionstudio cloud login`,
`visionstudio sync --tenant <slug>`), whose ergonomics are specced in
TRD T2.

This document must be written before Phase 3 (hosted web shell) begins —
the first surface with real UX decisions: login flow, tenant switcher,
panel parity with local, empty states for freshly-synced tenants, and the
free-tier onboarding path (M3 gate requires one-command onboarding
without hand-holding).
