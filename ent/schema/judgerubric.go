package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// JudgeRubric defines scoring criteria for evaluating a specific spec type
// within a workflow. Used by LLM-as-a-Judge for spec quality assessment.
type JudgeRubric struct {
	ent.Schema
}

func (JudgeRubric) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("rubric_id").MaxLen(64),
		field.String("spec_type").MaxLen(64),
		field.JSON("criteria", map[string]any{}).Optional(),
		field.Text("prompt_template").Optional(),
	}
}

func (JudgeRubric) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", SpecWorkflow.Type).Ref("rubrics").Unique(),
		edge.To("results", JudgeResult.Type),
	}
}
