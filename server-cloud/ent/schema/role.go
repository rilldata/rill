package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").
			Immutable().
			Unique().
			NotEmpty(),
		field.Time("created_on").
			Default(time.Now),
		field.Time("updated_on").
			Default(time.Now),
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("permission", Permission.Type),
		edge.From("user", User.Type).
			Ref("role"),
	}
}
