package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Initiative holds the schema definition for the Initiative entity.
type Initiative struct {
	ent.Schema
}

func (Initiative) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("initiative_id").MaxLen(64),
		field.String("organization").MaxLen(128),
		field.String("title").MaxLen(255),
		field.Text("description").Optional(),
		field.String("status").MaxLen(32),
		field.String("init_type").MaxLen(32).Default("feature"),
		field.String("priority").MaxLen(32).Optional(),
		field.String("home_repo").MaxLen(255).Optional(),
		field.String("workspace").MaxLen(128).Optional(),
		field.JSON("specs", map[string]string{}).Optional(),
		field.Time("created_at"),
		field.Time("planned_at").Optional().Nillable(),
		field.Time("executing_at").Optional().Nillable(),
		field.Time("delivery_complete_at").Optional().Nillable(),
		field.Time("released_at").Optional().Nillable(),
		field.Time("closed_at").Optional().Nillable(),
		field.Time("updated_at"),
	}
}

func (Initiative) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("phases", Phase.Type),
		edge.To("roadmap_items", RoadmapItem.Type),
		edge.From("program", Program.Type).Ref("initiatives").Unique(),
		edge.From("workflow", SpecWorkflow.Type).Ref("initiatives").Unique(),
	}
}
