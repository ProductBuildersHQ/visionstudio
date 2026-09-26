package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Phase holds the schema definition for the Phase entity.
type Phase struct {
	ent.Schema
}

// Mixin adds the optional tenant_id discriminator for cloud pool + RLS
// tenancy; inert for the single-tenant local app. See TenantMixin.
func (Phase) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (Phase) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("phase_id").MaxLen(64),
		field.Int("sequence_number"),
		field.String("title").MaxLen(255),
		field.String("theme").MaxLen(255).Optional(),
	}
}

func (Phase) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("initiative", Initiative.Type).Ref("phases").Unique().Required(),
		edge.To("roadmap_items", RoadmapItem.Type),
	}
}
