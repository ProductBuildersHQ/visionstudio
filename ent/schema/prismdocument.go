package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// PRISMDocument holds prism-maturity domain/stage models.
// TRD T3: two maturity IRs coexist — CapabilityModel (simple) and PRISMDocument (full).
type PRISMDocument struct {
	ent.Schema
}

// Mixin adds the optional tenant_id discriminator for cloud pool + RLS
// tenancy; inert for the single-tenant local app. See TenantMixin.
func (PRISMDocument) Mixin() []ent.Mixin {
	return []ent.Mixin{TenantMixin{}}
}

func (PRISMDocument) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("organization").Optional(),
		field.String("repository_id").Optional(),
		field.String("name"),
		field.String("description").Optional(),
		field.String("version").Optional(),
		field.JSON("domains", []any{}).Optional(),
		field.JSON("layers", []any{}).Optional(),
		field.JSON("metrics", []any{}).Optional(),
		field.JSON("maturity", map[string]any{}).Optional(),
		field.JSON("sli_state", map[string]any{}).Optional(),
		field.JSON("maturity_state", map[string]any{}).Optional(),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}
