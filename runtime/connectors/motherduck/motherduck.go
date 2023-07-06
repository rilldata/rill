package motherduck

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/connectors"
	"go.uber.org/zap"
)

func init() {
	connectors.Register("motherduck", connector{})
}

var spec = connectors.Spec{
	DisplayName: "MotherDuck",
	Description: "Import data from MotherDuck.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "query",
			Type:        connectors.StringPropertyType,
			Required:    true,
			DisplayName: "Query",
			Description: "Query to extract data from MotherDuck.",
			Placeholder: "select * from my_db.my_table;",
		},
	},
	ConnectorVariables: []connectors.VariableSchema{
		{
			Key:    "token",
			Secret: true,
		},
	},
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

// duckdb driver directly ingests data from motherduck service
func (c connector) ConsumeAsIterator(ctx context.Context, env *connectors.Env, source *connectors.Source, logger *zap.Logger) (connectors.FileIterator, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c connector) HasAnonymousAccess(ctx context.Context, env *connectors.Env, source *connectors.Source) (bool, error) {
	return false, nil
}
