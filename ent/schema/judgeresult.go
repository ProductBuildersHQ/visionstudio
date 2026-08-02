package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// JudgeResult stores an LLM-as-a-Judge evaluation of a spec document.
type JudgeResult struct {
	ent.Schema
}

func (JudgeResult) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("result_id").MaxLen(64),
		field.String("initiative_id").MaxLen(64),
		field.String("spec_path").MaxLen(512),
		field.Float("score").Optional(),
		field.Text("rationale").Optional(),
		field.String("model").MaxLen(128).Optional(),
		field.Time("evaluated_at"),
	}
}

func (JudgeResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("rubric", JudgeRubric.Type).Ref("results").Unique(),
	}
}
