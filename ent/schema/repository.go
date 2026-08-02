package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Repository holds the schema definition for the Repository entity.
type Repository struct {
	ent.Schema
}

func (Repository) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("repository_id").MaxLen(128),
		field.String("organization").MaxLen(128),
		field.String("repository_name").MaxLen(128),
		field.String("default_branch").MaxLen(128).Default("main"),
		field.String("local_path").MaxLen(512).Optional(),
		field.String("go_module").MaxLen(256).Optional(),
		field.String("domain").MaxLen(128).Optional(),
		field.String("status").MaxLen(32),
		field.String("ingest_high_water").MaxLen(128).Optional(),
	}
}

func (Repository) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roadmap_items", RoadmapItem.Type),
	}
}
