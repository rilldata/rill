package parser

import (
	"fmt"
	"net/mail"

	"google.golang.org/protobuf/types/known/structpb"
)

// QueryForYAML is the raw structure of a "for" clause defined in YAML.
// It specifies the user identity or attributes to use when evaluating security policies for a query.
type QueryForYAML struct {
	UserID     string         `yaml:"user_id"`
	UserEmail  string         `yaml:"user_email"`
	Attributes map[string]any `yaml:"attributes"`
}

// queryFor is the parsed representation of a QueryForYAML.
// At most one of the fields is set.
type queryFor struct {
	userID     string
	userEmail  string
	attributes *structpb.Struct
}

// parseQueryForYAML validates and parses a QueryForYAML.
// The prop parameter is the path of the property in the YAML, used in error messages (e.g. "for" or "query.for").
func parseQueryForYAML(raw *QueryForYAML, prop string) (*queryFor, error) {
	res := &queryFor{}
	n := 0
	if raw.UserID != "" {
		n++
		res.userID = raw.UserID
	}
	if raw.UserEmail != "" {
		n++
		_, err := mail.ParseAddress(raw.UserEmail)
		if err != nil {
			return nil, fmt.Errorf(`invalid value %q for property "%s.user_email"`, raw.UserEmail, prop)
		}
		res.userEmail = raw.UserEmail
	}
	if len(raw.Attributes) > 0 {
		n++
		var err error
		res.attributes, err = structpb.NewStruct(raw.Attributes)
		if err != nil {
			return nil, fmt.Errorf(`failed to serialize property "%s.attributes": %w`, prop, err)
		}
	}
	if n > 1 {
		return nil, fmt.Errorf(`only one of "%[1]s.user_id", "%[1]s.user_email", or "%[1]s.attributes" may be set`, prop)
	}
	return res, nil
}
