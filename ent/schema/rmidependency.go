package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RMIDependency holds the schema definition for the RMIDependency entity.
type RMIDependency struct {
	ent.Schema
}

func (RMIDependency) Fields() []ent.Field {
	return []ent.Field{
		field.String("source_rmi_id").MaxLen(64),
		field.String("target_rmi_id").MaxLen(64),
		field.String("relationship").MaxLen(32),
	}
}

func (RMIDependency) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_rmi_id", "target_rmi_id", "relationship").Unique(),
	}
}

func (RMIDependency) Edges() []ent.Edge {
	return nil
}
