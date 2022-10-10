package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name"),
		field.String("UserName").
			Immutable().
			Unique().
			NotEmpty(),
		field.String("Description").
			Default("Unknown"),
		field.Time("created_on").
			Default(time.Now),
		field.Time("updated_on").
			Default(time.Now()),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("projects", Project.Type),
		edge.To("role", Role.Type),
		// create an inverse-edge called "groups" of type `Groups`
		// and reference it to the "users" edge (in Group schema)
		// explicitly using the `Ref` method.
		edge.From("groups", Groups.Type).
			Ref("users"),
		edge.From("organization", Organization.Type).
			Ref("users"),
	}
}
