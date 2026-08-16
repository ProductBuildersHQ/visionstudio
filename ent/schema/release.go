package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Release is a per-repository release (a git tag that shipped): the unit
// that initiatives associate with to answer "shipped in what release?".
// A multi-repo initiative accumulates one release per repo per ship; a
// release lists every initiative and RMI whose work it carried.
//
// ID is "<repository-id>@<tag>" (e.g.
// "github.com/ProductBuildersHQ/visionstudio@v0.3.0") — the full
// repository ID rather than a bare slug, so releases stay unique across
// organizations.
type Release struct {
	ent.Schema
}

// Fields of the Release.
func (Release) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("release_id").MaxLen(256),
		field.String("tag").MaxLen(128),
		field.Time("released_at"),
		field.String("url").MaxLen(512).Optional(),
		field.String("notes_ref").MaxLen(256).Optional(),
		// Body is release-notes text (e.g. a GitHub Release body),
		// truncated. Captured as match evidence for AI-assisted
		// historical backfill (RMI-VISIONSTUDIO-315) — never as fact,
		// never auto-interpreted; it exists so a human review step
		// doesn't need to re-fetch from GitHub.
		field.String("body").MaxLen(4096).Optional(),
		field.String("repository_id").MaxLen(128),
		field.Time("created_at"),
		field.Time("updated_at"),
	}
}

// Edges of the Release.
func (Release) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("releases").
			Field("repository_id").
			Unique().
			Required(),
		edge.To("initiatives", Initiative.Type),
		edge.To("roadmap_items", RoadmapItem.Type),
	}
}
