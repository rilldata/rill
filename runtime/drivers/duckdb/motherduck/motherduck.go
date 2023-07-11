package motherduck

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.RegisterConnector("motherduck", connector{})
}

var spec = drivers.Spec{
	DisplayName: "Motherduck",
	Description: "Import data from Motherduck.",
	Properties: []drivers.PropertySchema{
		{
			Key:         "query",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Query",
			Description: "Query to extract data from Motherduck.",
			Placeholder: "select * from my_db.my_table;",
		},
	},
	ConnectorVariables: []drivers.VariableSchema{
		{
			Key:    "token",
			Secret: true,
		},
	},
}

type connector struct{}

func (c connector) Spec() drivers.Spec {
	return spec
}

func (c connector) HasAnonymousAccess(ctx context.Context, props map[string]any) (bool, error) {
	return false, nil
}
