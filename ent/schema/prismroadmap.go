package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// PRISMRoadmap holds prism-roadmap artifacts keyed by repo.
// Stores the full roadmap JSON for prism-roadmap/roadmap.Roadmap.
type PRISMRoadmap struct {
	ent.Schema
}

// Mixin adds the optional tenant_id discriminator for cloud pool + RLS
// tenancy; inert for the single-tenant local app. See TenantMixin.
func (PRISMRoadmap) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (PRISMRoadmap) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("organization").Optional(),
		field.String("repository_id"),
		field.String("name").Optional(),
		field.JSON("phases", []any{}).Optional(),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}

// PRISMGoal holds prism-roadmap goals artifacts.
type PRISMGoal struct {
	ent.Schema
}

func (PRISMGoal) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("organization").Optional(),
		field.String("repository_id"),
		field.String("goal_type"),
		field.JSON("document", map[string]any{}).Optional(),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}
