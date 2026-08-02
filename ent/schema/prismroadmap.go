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
