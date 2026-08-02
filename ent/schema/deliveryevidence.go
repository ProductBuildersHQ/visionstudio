package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// DeliveryEvidence holds the schema definition for the DeliveryEvidence entity.
type DeliveryEvidence struct {
	ent.Schema
}

func (DeliveryEvidence) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("evidence_id").MaxLen(128),
		field.String("evidence_type").MaxLen(32),
		field.String("reference").MaxLen(512),
		field.String("commit_type").MaxLen(32).Optional(),
		field.String("commit_scope").MaxLen(64).Optional(),
		field.Time("occurred_at").Optional().Nillable(),
		field.Time("created_at"),
	}
}

func (DeliveryEvidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roadmap_item", RoadmapItem.Type).Ref("evidence").Unique().Required(),
	}
}
