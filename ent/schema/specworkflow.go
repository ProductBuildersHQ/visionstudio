package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SpecWorkflow defines a specification workflow template that determines
// which specs are required or optional for initiatives of certain types.
type SpecWorkflow struct {
	ent.Schema
}

func (SpecWorkflow) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("workflow_id").MaxLen(64),
		field.String("name").MaxLen(128),
		field.Text("description").Optional(),
		field.JSON("specs_required", []string{}).Optional(),
		field.JSON("specs_optional", []string{}).Optional(),
		field.JSON("init_types", []string{}).Optional(),
	}
}

func (SpecWorkflow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("initiatives", Initiative.Type),
		edge.To("rubrics", JudgeRubric.Type),
	}
}
