package localfile

import (
	"context"
	"fmt"
	"path"

	"github.com/mitchellh/mapstructure"
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

type Config struct {
	Path         string `mapstructure:"path"`
	Format       string `mapstructure:"format"`
	CSVDelimiter string `mapstructure:"csv.delimiter"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Format == "" {
		conf.Format = path.Ext(conf.Path)
	}

	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

// local file connectors should directly use glob patterns
// keeping it for reference
func (c connector) ConsumeAsFile(ctx context.Context, env *connectors.Env, source *connectors.Source) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
