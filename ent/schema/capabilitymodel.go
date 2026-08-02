package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// CapabilityModel defines a capability maturity framework.
// Each model contains a set of dimensions (capabilities) that can be assessed.
// Examples: "big-tech-product", "continuous-discovery", "api-first".
type CapabilityModel struct {
	ent.Schema
}

func (CapabilityModel) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("model_id").MaxLen(64),
		field.String("name").MaxLen(128),
		field.Text("description").Optional(),
		field.JSON("dimensions", []Dimension{}).Optional(),
		field.Int("max_level").Default(5),
	}
}

func (CapabilityModel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("assessments", MaturityAssessment.Type),
	}
}

// Dimension is a capability area within a maturity model.
type Dimension struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Levels      []Level  `json:"levels,omitempty"`
	Sources     []string `json:"sources,omitempty"`
}

// Level describes a maturity level within a dimension.
type Level struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
