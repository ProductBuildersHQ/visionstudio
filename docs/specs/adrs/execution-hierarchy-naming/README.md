# ADR: Execution Hierarchy Naming

Architecture decision records for VisionStudio's Program / Initiative / Phase
/ RMI vocabulary.

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-001](ADR-001-ai-native-domain-vocabulary.md) | Keep Program / Initiative / Phase / RMI as Native Vocabulary | Accepted |

## Summary

VisionStudio's execution hierarchy names emerged from how the team actually
works — AI-agent-driven, weekly-shippable, releases as an emergent fact
rather than a forward plan — not from adopting an existing PM tool's schema.
A deliberate comparison against Jira's and Aha's real data models (including
Aha's actual Go client structs) confirmed the granularity is sound but showed
the *names* (Epic, Story, Feature, Release) would import assumptions that
don't fit: named theme entities, human-narrative framing, product-surface
framing, and forward-planned scheduling. The decision is to keep the native
vocabulary and use the Jira/Aha mapping only as an onboarding translation
reference, not a convergence target.
