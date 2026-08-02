package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SpecDocument is a visionspec document registry entry.
// RMI-121: Spec docs/rubrics discoverable per repo/initiative.
type SpecDocument struct {
	ent.Schema
}

func (SpecDocument) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("organization").Optional(),
		field.String("repository_id"),
		field.String("initiative_id").Optional(),
		field.String("spec_type"),
		field.String("file_path"),
		field.String("title").Optional(),
		field.String("status").Optional(),
		field.String("content_hash").Optional(),
		field.Time("synced_at"),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}

func (SpecDocument) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("initiative", Initiative.Type).Ref("spec_documents").Unique().Field("initiative_id"),
		edge.From("repository", Repository.Type).Ref("spec_documents").Unique().Required().Field("repository_id"),
	}
}
