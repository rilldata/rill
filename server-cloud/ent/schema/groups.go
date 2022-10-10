package schema

import (
	"regexp"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Groups holds the schema definition for the Groups entity.
type Groups struct {
	ent.Schema
}

// Fields of the Groups.
func (Groups) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").
			// regexp validation for group name.
			Match(regexp.MustCompile("[a-zA-Z_]+$")),
		field.String("Description").
			Default("Unknown"),
		field.Time("created_on").
			Default(time.Now),
		field.Time("updated_on").
			Default(time.Now),
	}
}

// Edges of the Groups.
func (Groups) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
		edge.To("projects", Project.Type),
		// create an inverse-edge called "organization" of type `Organization`
		// and reference it to the "users" edge (in Group schema)
		// explicitly using the `Ref` method.
		edge.From("organization", Organization.Type).
			Ref("groups"),
	}
}
