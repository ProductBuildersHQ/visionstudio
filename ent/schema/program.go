package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Program holds the schema definition for the Program entity.
type Program struct {
	ent.Schema
}

func (Program) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("program_id").MaxLen(64),
		field.String("name").MaxLen(128),
		field.String("organization").MaxLen(128),
		field.Text("description").Optional(),
		field.Bool("hidden").Default(false),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}

func (Program) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("initiatives", Initiative.Type),
	}
}
