package localfile

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.RegisterConnector("local_file", &connection{})
}

var spec = drivers.Spec{
	DisplayName: "Local file",
	Description: "Import Locally Stored File.",
	Properties: []drivers.PropertySchema{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path or URL to file",
			Placeholder: "/path/to/file",
		},
		{
			Key:         "format",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Format",
			Description: "Either CSV or Parquet. Inferred if not set.",
			Placeholder: "csv",
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
