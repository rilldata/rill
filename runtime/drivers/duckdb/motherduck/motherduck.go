package motherduck

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.RegisterConnector("motherduck", connector{})
}

var spec = drivers.Spec{
	DisplayName: "MotherDuck",
	Description: "Import data from MotherDuck.",
	Properties: []drivers.PropertySchema{
		{
			Key:         "query",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Query",
			Description: "Query to extract data from MotherDuck.",
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
