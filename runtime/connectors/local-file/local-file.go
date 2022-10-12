package file

import (
	"context"
	"errors"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/sources"
	"github.com/rilldata/rill/runtime/drivers"
)

var spec = []sources.Property{
	{
		Key:         "path",
		DisplayName: "Path",
		Description: "Path to file on the disk.",
		Placeholder: "/path/to/file",
		Type:        api.Connector_Property_TYPE_STRING,
		Required:    true,
	},
	{
		Key:         "format",
		DisplayName: "Format",
		Description: "Either CSV or Parquet. Inferred if not set.",
		Placeholder: "csv",
		Type:        api.Connector_Property_TYPE_STRING,
		Required:    false,
	},
	{
		Key:         "delimiter",
		DisplayName: "Delimiter",
		Description: "Forced delimiter for csv file.",
		Placeholder: ",",
		Type:        api.Connector_Property_TYPE_STRING,
		Required:    false,
	},
}

func init() {
	connectors.Register(sources.LocalFileConnectorName, connector{})
}

type LocalFileConfig struct {
	Path      string `key:"path"`
	Format    string `key:"format"`
	Delimiter string `key:"path"`
}

type connector struct{}

func (c connector) Ingest(ctx context.Context, source sources.Source, olap drivers.OLAPStore) error {
	var localFileConfig LocalFileConfig
	err := connectors.ValidatePropertiesAndExtract(source, c.Spec(), &localFileConfig)
	if err != nil {
		return err
	}

	supported, _, err := olap.Ingest(ctx, source, localFileConfig)
	if supported {
		return err
	}
	return errors.New("OLAP doesnt support local file")
}

func (c connector) Validate(source sources.Source) error {
	return nil
}

func (c connector) Spec() []sources.Property {
	return spec
}
