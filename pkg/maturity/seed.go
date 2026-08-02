// Package maturity provides capability maturity model definitions and assessment logic.
//
//nolint:dupl // Model builder functions intentionally share structure
package maturity

import (
	"context"
	"fmt"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store"
)

// BuiltInModels returns the predefined capability maturity models.
func BuiltInModels() []*store.CapabilityModel {
	return []*store.CapabilityModel{
		bigTechEssentialsModel(),
		bigTechFullModel(),
		continuousDiscoveryModel(),
		apiFirstModel(),
	}
}

// SeedBuiltIn inserts all built-in capability models into the store.
// It skips models that already exist (idempotent).
func SeedBuiltIn(ctx context.Context, s store.MaturityStore) (int, error) {
	models := BuiltInModels()
	created := 0
	for _, m := range models {
		if _, err := s.GetCapabilityModel(ctx, m.ID); err == nil {
			continue
		}
		if err := s.CreateCapabilityModel(ctx, m); err != nil {
			return created, fmt.Errorf("create model %s: %w", m.ID, err)
		}
		created++
	}
	return created, nil
}

func bigTechEssentialsModel() *store.CapabilityModel {
	return &store.CapabilityModel{
		ID:          "big-tech-essentials",
		Name:        "Big Tech Essentials",
		Description: "Streamlined maturity model focusing on Amazon, Google, and Stripe practices.",
		MaxLevel:    5,
		Dimensions: []store.Dimension{
			{
				Key:     "customer-obsession",
				Name:    "Customer Obsession",
				Sources: []string{"Amazon"},
				Levels: []store.Level{
					{Level: 1, Name: "Ad-hoc", Description: "Customer feedback is gathered sporadically"},
					{Level: 2, Name: "Reactive", Description: "Respond to customer issues as they arise"},
					{Level: 3, Name: "Proactive", Description: "Regular customer research informs decisions"},
					{Level: 4, Name: "Integrated", Description: "Customer metrics drive roadmap priorities"},
					{Level: 5, Name: "Optimized", Description: "Working backwards from customer press release"},
				},
			},
			{
				Key:     "okr-quality",
				Name:    "OKR Quality",
				Sources: []string{"Google"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No objectives or key results defined"},
					{Level: 2, Name: "Basic", Description: "Objectives exist but lack measurable results"},
					{Level: 3, Name: "Measurable", Description: "KRs are quantifiable with clear targets"},
					{Level: 4, Name: "Ambitious", Description: "Stretch goals with 70% achievement norm"},
					{Level: 5, Name: "Aligned", Description: "OKRs cascade and align across teams"},
				},
			},
			{
				Key:     "api-first",
				Name:    "API-First Design",
				Sources: []string{"Stripe"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No API considerations in design"},
					{Level: 2, Name: "Afterthought", Description: "APIs added after implementation"},
					{Level: 3, Name: "Designed", Description: "API contracts defined before coding"},
					{Level: 4, Name: "Developer-centric", Description: "APIs optimized for DX with examples"},
					{Level: 5, Name: "Platform", Description: "APIs are the product; external consumers"},
				},
			},
			{
				Key:     "explicit-tradeoffs",
				Name:    "Explicit Tradeoffs",
				Sources: []string{"Google"},
				Levels: []store.Level{
					{Level: 1, Name: "Hidden", Description: "Tradeoffs not documented or discussed"},
					{Level: 2, Name: "Acknowledged", Description: "Tradeoffs mentioned but not analyzed"},
					{Level: 3, Name: "Documented", Description: "Alternatives considered in design docs"},
					{Level: 4, Name: "Quantified", Description: "Tradeoffs have cost/benefit estimates"},
					{Level: 5, Name: "Reversible", Description: "Two-way door decisions distinguished"},
				},
			},
			{
				Key:     "documentation",
				Name:    "Documentation Quality",
				Sources: []string{"Stripe"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No documentation exists"},
					{Level: 2, Name: "Basic", Description: "README with setup instructions"},
					{Level: 3, Name: "Complete", Description: "API reference and usage examples"},
					{Level: 4, Name: "Exceptional", Description: "Tutorials, guides, and error docs"},
					{Level: 5, Name: "Living", Description: "Docs tested as code; always current"},
				},
			},
		},
	}
}

func bigTechFullModel() *store.CapabilityModel {
	return &store.CapabilityModel{
		ID:          "big-tech-full",
		Name:        "Big Tech Full",
		Description: "Comprehensive maturity model combining practices from 10 leading methodologies.",
		MaxLevel:    5,
		Dimensions: []store.Dimension{
			{
				Key:     "customer-obsession",
				Name:    "Customer Obsession",
				Sources: []string{"Amazon"},
			},
			{
				Key:     "okr-quality",
				Name:    "OKR Quality",
				Sources: []string{"Google"},
			},
			{
				Key:     "api-first",
				Name:    "API-First Design",
				Sources: []string{"Stripe"},
			},
			{
				Key:     "explicit-tradeoffs",
				Name:    "Explicit Tradeoffs",
				Sources: []string{"Google"},
			},
			{
				Key:     "documentation",
				Name:    "Documentation Quality",
				Sources: []string{"Stripe"},
			},
			{
				Key:     "freedom-responsibility",
				Name:    "Freedom & Responsibility",
				Sources: []string{"Netflix"},
			},
			{
				Key:     "autonomous-squads",
				Name:    "Autonomous Squads",
				Sources: []string{"Spotify"},
			},
			{
				Key:     "fixed-time-variable-scope",
				Name:    "Fixed Time, Variable Scope",
				Sources: []string{"Basecamp"},
			},
			{
				Key:     "weekly-touchpoints",
				Name:    "Weekly Customer Touchpoints",
				Sources: []string{"Torres"},
			},
			{
				Key:     "assumption-testing",
				Name:    "Assumption Testing",
				Sources: []string{"Torres"},
			},
		},
	}
}

func continuousDiscoveryModel() *store.CapabilityModel {
	return &store.CapabilityModel{
		ID:          "continuous-discovery",
		Name:        "Continuous Discovery",
		Description: "Teresa Torres continuous discovery habits maturity assessment.",
		MaxLevel:    5,
		Dimensions: []store.Dimension{
			{
				Key:     "weekly-touchpoints",
				Name:    "Weekly Customer Touchpoints",
				Sources: []string{"Torres"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No regular customer contact"},
					{Level: 2, Name: "Monthly", Description: "Customer contact at least monthly"},
					{Level: 3, Name: "Weekly", Description: "Talking to customers every week"},
					{Level: 4, Name: "Systematic", Description: "Automated recruiting and scheduling"},
					{Level: 5, Name: "Embedded", Description: "Product trio engages together weekly"},
				},
			},
			{
				Key:     "opportunity-trees",
				Name:    "Opportunity Solution Trees",
				Sources: []string{"Torres"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No outcome-based thinking"},
					{Level: 2, Name: "Outcomes", Description: "Clear outcome defined"},
					{Level: 3, Name: "Opportunities", Description: "Multiple opportunities mapped"},
					{Level: 4, Name: "Solutions", Description: "Multiple solutions per opportunity"},
					{Level: 5, Name: "Living", Description: "OST updated weekly with learnings"},
				},
			},
			{
				Key:     "assumption-testing",
				Name:    "Assumption Testing",
				Sources: []string{"Torres"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "Build without validation"},
					{Level: 2, Name: "Late", Description: "Test after building"},
					{Level: 3, Name: "Early", Description: "Test before building"},
					{Level: 4, Name: "Prioritized", Description: "Test riskiest assumptions first"},
					{Level: 5, Name: "Continuous", Description: "Parallel testing across solutions"},
				},
			},
			{
				Key:     "story-interviews",
				Name:    "Story-Based Interviews",
				Sources: []string{"Torres"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No interviews conducted"},
					{Level: 2, Name: "Opinion", Description: "Ask what customers want"},
					{Level: 3, Name: "Story", Description: "Collect stories of past behavior"},
					{Level: 4, Name: "Structured", Description: "Consistent story mining protocol"},
					{Level: 5, Name: "Synthesized", Description: "Stories mapped to opportunities"},
				},
			},
		},
	}
}

func apiFirstModel() *store.CapabilityModel {
	return &store.CapabilityModel{
		ID:          "api-first",
		Name:        "API-First",
		Description: "Stripe-inspired API-first design maturity assessment.",
		MaxLevel:    5,
		Dimensions: []store.Dimension{
			{
				Key:     "api-design",
				Name:    "API Design",
				Sources: []string{"Stripe"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No API design process"},
					{Level: 2, Name: "Informal", Description: "Ad-hoc endpoint creation"},
					{Level: 3, Name: "Contract", Description: "OpenAPI spec before code"},
					{Level: 4, Name: "Reviewed", Description: "API reviews before shipping"},
					{Level: 5, Name: "Evolved", Description: "Versioning strategy and deprecation policy"},
				},
			},
			{
				Key:     "error-handling",
				Name:    "Error Handling",
				Sources: []string{"Stripe"},
				Levels: []store.Level{
					{Level: 1, Name: "Opaque", Description: "Generic error messages"},
					{Level: 2, Name: "Typed", Description: "Error codes defined"},
					{Level: 3, Name: "Helpful", Description: "Error messages guide resolution"},
					{Level: 4, Name: "Consistent", Description: "Uniform error format across APIs"},
					{Level: 5, Name: "Documented", Description: "All errors documented with examples"},
				},
			},
			{
				Key:     "developer-experience",
				Name:    "Developer Experience",
				Sources: []string{"Stripe"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No DX consideration"},
					{Level: 2, Name: "Basic", Description: "API reference exists"},
					{Level: 3, Name: "Good", Description: "SDKs and quickstarts available"},
					{Level: 4, Name: "Great", Description: "Interactive docs and playground"},
					{Level: 5, Name: "Delightful", Description: "Time-to-first-success optimized"},
				},
			},
			{
				Key:     "consistency",
				Name:    "API Consistency",
				Sources: []string{"Stripe"},
				Levels: []store.Level{
					{Level: 1, Name: "None", Description: "No style guide"},
					{Level: 2, Name: "Partial", Description: "Some conventions followed"},
					{Level: 3, Name: "Documented", Description: "Style guide exists"},
					{Level: 4, Name: "Enforced", Description: "Linters enforce style"},
					{Level: 5, Name: "Predictable", Description: "Developers can guess the API shape"},
				},
			},
		},
	}
}
