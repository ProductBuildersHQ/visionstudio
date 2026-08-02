package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// MaturityAssessment captures a point-in-time capability assessment
// for an initiative or organization against a CapabilityModel.
type MaturityAssessment struct {
	ent.Schema
}

func (MaturityAssessment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("assessment_id").MaxLen(64),
		field.String("initiative_id").MaxLen(64).Optional(),
		field.String("organization").MaxLen(128).Optional(),
		field.JSON("scores", map[string]DimensionScore{}).Optional(),
		field.Float("overall_score").Optional().Nillable(),
		field.Text("summary").Optional(),
		field.String("assessed_by").MaxLen(128).Optional(),
		field.String("model").MaxLen(64).Optional(),
		field.Time("assessed_at"),
	}
}

func (MaturityAssessment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("capability_model", CapabilityModel.Type).Ref("assessments").Unique(),
	}
}

// DimensionScore captures the assessment for a single dimension.
type DimensionScore struct {
	Level     int    `json:"level"`
	Rationale string `json:"rationale,omitempty"`
	Evidence  string `json:"evidence,omitempty"`
}
