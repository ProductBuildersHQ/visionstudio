package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Assignment holds the schema definition for the Assignment entity.
type Assignment struct {
	ent.Schema
}

// Mixin adds the optional tenant_id discriminator for cloud pool + RLS
// tenancy; inert for the single-tenant local app. See TenantMixin.
func (Assignment) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (Assignment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("assignment_id").MaxLen(64),
		field.String("worker").MaxLen(128),
		field.String("status").MaxLen(32),
		field.Time("lease_expires_at"),
		field.String("workspace").MaxLen(512).Optional(),
		field.JSON("handoff", map[string]any{}).Optional(),
		field.Time("created_at"),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("updated_at"),
	}
}

func (Assignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roadmap_item", RoadmapItem.Type).Ref("assignments").Unique().Required(),
	}
}
