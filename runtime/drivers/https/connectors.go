package https

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.RegisterConnector("https", &connection{})
}

var spec = drivers.Spec{
	DisplayName: "http(s)",
	Description: "Connect to a remote file.",
	Properties: []drivers.PropertySchema{
		{
			Key:         "path",
			DisplayName: "Path",
			Description: "Path to the remote file.",
			Placeholder: "https://example.com/file.csv",
			Type:        drivers.StringPropertyType,
			Required:    true,
		},
	},
}

// ConnectorSpec implements drivers.Connection.
func (c *connection) Spec() drivers.Spec {
	return spec
}

func (c *connection) HasAnonymousAccess(ctx context.Context, props map[string]any) (bool, error) {
	return true, nil
}
