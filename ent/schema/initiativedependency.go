package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InitiativeDependency holds the schema definition for directed
// edges between initiatives within a program.
type InitiativeDependency struct {
	ent.Schema
}

// Mixin adds the optional tenant_id discriminator for cloud pool + RLS
// tenancy; inert for the single-tenant local app. See TenantMixin.
func (InitiativeDependency) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (InitiativeDependency) Fields() []ent.Field {
	return []ent.Field{
		field.String("source_initiative_id").MaxLen(64),
		field.String("target_initiative_id").MaxLen(64),
		field.String("relationship").MaxLen(32),
	}
}

func (InitiativeDependency) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_initiative_id", "target_initiative_id", "relationship").Unique(),
	}
}

func (InitiativeDependency) Edges() []ent.Edge {
	return nil
}
