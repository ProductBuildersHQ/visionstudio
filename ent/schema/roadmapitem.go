package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// RoadmapItem holds the schema definition for the RoadmapItem entity.
type RoadmapItem struct {
	ent.Schema
}

func (RoadmapItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("rmi_id").MaxLen(64),
		field.String("title").MaxLen(255),
		field.Text("description").Optional(),
		field.String("item_type").MaxLen(32),
		field.String("status").MaxLen(32),
		field.String("priority").MaxLen(32).Optional(),
		field.Bool("required").Default(true),
		field.Int("sequence_number").Optional(),
		field.JSON("acceptance_criteria", []string{}).Optional(),
		field.Time("created_at"),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("updated_at"),
	}
}

func (RoadmapItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("initiative", Initiative.Type).Ref("roadmap_items").Unique(),
		edge.From("phase", Phase.Type).Ref("roadmap_items").Unique(),
		edge.From("repository", Repository.Type).Ref("roadmap_items").Unique().Required(),
		edge.To("assignments", Assignment.Type),
		edge.To("evidence", DeliveryEvidence.Type),
	}
}
