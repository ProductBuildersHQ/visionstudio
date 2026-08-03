package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// InitiativeWorkflow records which workflow is selected for an initiative.
// RMI-030: Join table for initiative workflow selection with timestamp.
type InitiativeWorkflow struct {
	ent.Schema
}

func (InitiativeWorkflow) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("initiative_id"),
		field.String("workflow_id"),
		field.Time("selected_at"),
	}
}

func (InitiativeWorkflow) Edges() []ent.Edge {
	return nil
}
