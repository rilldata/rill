package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Organization holds the schema definition for the Organization entity.
type Organization struct {
	ent.Schema
}

// Fields of the Organization.
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name"),
		field.String("Description").
			Default("Unknown"),
		field.Time("created_on").
			Default(time.Now),
		field.Time("updated_on").
			Default(time.Now),
	}
}

// Edges of the Organization.
func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("groups", Groups.Type),
		edge.To("projects", Project.Type),
		edge.To("users", User.Type),
	}
}
