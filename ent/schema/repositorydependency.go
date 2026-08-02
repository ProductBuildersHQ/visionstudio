package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RepositoryDependency holds the schema definition for the RepositoryDependency entity.
type RepositoryDependency struct {
	ent.Schema
}

func (RepositoryDependency) Fields() []ent.Field {
	return []ent.Field{
		field.String("source_repository_id").MaxLen(128),
		field.String("target_repository_id").MaxLen(128),
		field.String("dependency_type").MaxLen(32),
	}
}

func (RepositoryDependency) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_repository_id", "target_repository_id", "dependency_type").Unique(),
	}
}

func (RepositoryDependency) Edges() []ent.Edge {
	return nil
}
