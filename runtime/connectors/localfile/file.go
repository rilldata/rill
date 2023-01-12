package localfile

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/connectors"
)

func init() {
	connectors.Register("local_file", connector{})
}

var spec = connectors.Spec{
	DisplayName: "Local file",
	Description: "Import Locally Stored File.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			Type:        connectors.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path or URL to file",
			Placeholder: "/path/to/file",
		},
		{
			Key:         "format",
			Type:        connectors.StringPropertyType,
			Required:    false,
			DisplayName: "Format",
			Description: "Either CSV or Parquet. Inferred if not set.",
			Placeholder: "csv",
		},
		{
			Key:         "csv.delimiter",
			Type:        connectors.StringPropertyType,
			Required:    false,
			DisplayName: "CSV Delimiter",
			Description: "Force delimiter for a CSV file.",
			Placeholder: ",",
		},
	},
}

// type config struct {
// 	connectors.Config `mapstructure:",squash"`
// }

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

// local file connectors should directly use glob patterns
// keeping it for reference
func (c connector) ConsumeAsFiles(ctx context.Context, env *connectors.Env, source *connectors.Source) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
