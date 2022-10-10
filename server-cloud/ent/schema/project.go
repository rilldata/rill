package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").
			Immutable().
			NotEmpty().
			Unique(),
		field.String("Description").
			Default("Unknown"),
		field.Time("created_on").
			Default(time.Now),
		field.Time("updated_on").
			Default(time.Now),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", Groups.Type).
			Ref("projects"),
		edge.From("organization", Organization.Type).
			Ref("projects").Unique(),
	}
}
