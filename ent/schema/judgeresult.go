package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// JudgeResult stores an LLM-as-a-Judge evaluation of a spec document.
// The report field contains the full structured-evaluation rubric.Rubric
// serialized as JSON, enabling rich evaluation data without schema changes.
type JudgeResult struct {
	ent.Schema
}

func (JudgeResult) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("result_id").MaxLen(64),
		field.String("initiative_id").MaxLen(64),
		field.String("spec_path").MaxLen(512),
		field.String("spec_type").MaxLen(32).Optional(),
		field.Time("evaluated_at"),
		// Full structured-evaluation report stored as JSON.
		// Contains categories, findings, scores, confidence, decision.
		field.JSON("report", map[string]any{}).Optional(),
		// Legacy fields for backwards compatibility and quick queries.
		// These are derived from report but kept for indexing/filtering.
		field.Int("int_score").Optional().Comment("1-5 integer score from report.IntScore"),
		field.Bool("pass").Default(false).Comment("Pass/fail from report.Pass"),
		field.String("model").MaxLen(128).Optional().Comment("Judge model from report.Judge.Model"),
	}
}

func (JudgeResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("rubric", JudgeRubric.Type).Ref("results").Unique(),
		edge.From("initiative", Initiative.Type).Ref("judge_results").Unique().Required().Field("initiative_id"),
	}
}
