package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Person is an identity: a human with a GitHub login and the commit-author
// email variants used to attribute work. Deliberately minimal (mirrors, not
// imports, prism-core's display-oriented Person): affiliation roles are
// deferred to the systemforge membership substrate (INIT-VISIONSTUDIO-002).
type Person struct {
	ent.Schema
}

// Mixin adds the optional tenant_id discriminator for cloud pool + RLS
// tenancy; inert for the single-tenant local app. See TenantMixin.
func (Person) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

// Fields of the Person.
func (Person) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("person_id").MaxLen(128),
		field.String("github_login").MaxLen(128),
		field.String("display_name").MaxLen(256).Optional(),
		field.Strings("email_identities").Optional(),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}

// Edges of the Person.
func (Person) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organizations", Organization.Type),
	}
}
