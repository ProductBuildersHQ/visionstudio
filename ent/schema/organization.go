package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Organization is a first-class GitHub organization, or a user account
// acting as one (e.g. grokify). Repository.organization remains as a
// legacy string; the edge from Repository is the queryable relation.
type Organization struct {
	ent.Schema
}

// Fields of the Organization.
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("org_entity_id").MaxLen(128),
		field.String("login").MaxLen(128),
		field.String("kind").MaxLen(32).Default("organization"),
		field.String("display_name").MaxLen(256).Optional(),
		field.String("website").MaxLen(512).Optional(),
		field.String("release_page_url").MaxLen(512).Optional(),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}

// Edges of the Organization.
func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("repositories", Repository.Type),
		edge.From("members", Person.Type).Ref("organizations"),
	}
}
