package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// TenantMixin adds an OPTIONAL tenant_id discriminator to an entity, enabling
// the cloud's pool + Row-Level Security multi-tenancy (visionstudio-cloud
// architecture ADR-002).
//
// It is deliberately optional and carries no privacy policy: the single-tenant
// local app leaves it null and is entirely unaffected, while the cloud
// populates and isolates on it at the Postgres layer — a DEFAULT drawn from a
// per-request tenant GUC plus RLS policies, both added by a cloud-side
// migration on top of the Ent-generated schema. Isolation lives in Postgres
// RLS, not in Ent, so this mixin stays inert for local by construction.
//
// Approved additive change to the frozen local schema (architecture ADR-001
// exception log); see RMI-VISIONSTUDIO-557.
type TenantMixin struct {
	mixin.Schema
}

// Fields of the TenantMixin.
func (TenantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("tenant_id").Optional().MaxLen(64),
	}
}

// Indexes of the TenantMixin. tenant_id leads every tenant-scoped lookup in
// the cloud, so it is indexed.
func (TenantMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
	}
}
