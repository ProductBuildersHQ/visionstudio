package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// DevXPeriodReport holds developer experience metrics for a period.
// Mirrors devfolio's PeriodReport for DB persistence.
type DevXPeriodReport struct {
	ent.Schema
}

func (DevXPeriodReport) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("organization").Optional(),
		field.String("repository_id").Optional(),
		field.String("person_id"),
		field.String("period_type"),
		field.String("period_label"),
		field.Time("period_start"),
		field.Time("period_end"),
		field.JSON("metrics", map[string]any{}).Optional(),
		field.JSON("by_model", map[string]any{}).Optional(),
		field.Float("coverage_score").Optional(),
		field.Time("created_at"),
	}
}
