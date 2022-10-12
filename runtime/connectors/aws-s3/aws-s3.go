package s3

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
		Placeholder: "s3://<bucket>/<file>",
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
		Key:         "aws.region",
		DisplayName: "AWS Region for the bucket.",
		Description: "",
		Placeholder: "",
		Type:        api.Connector_Property_TYPE_STRING,
		Required:    true,
	},
	{
		Key:         "aws.access.key",
		DisplayName: "AWS Access Key",
		Description: "",
		Placeholder: "",
		Type:        api.Connector_Property_TYPE_STRING,
		Required:    false,
	},
	{
		Key:         "aws.access.secret",
		DisplayName: "AWS Access Secret",
		Description: "",
		Placeholder: "",
		Type:        api.Connector_Property_TYPE_STRING,
		Required:    false,
	},
	{
		Key:         "aws.access.session",
		DisplayName: "AWS Access Session Token",
		Description: "",
		Placeholder: "",
		Type:        api.Connector_Property_TYPE_STRING,
		Required:    false,
	},
}

func init() {
	connectors.Register(sources.AWSS3ConnectorName, connector{})
}

type connector struct{}

type AWSS3Config struct {
	Path       string `key:"path"`
	Format     string `key:"format"`
	AwsRegion  string `key:"aws.region"`
	AwsKey     string `key:"aws.access.key"`
	AwsSecret  string `key:"aws.access.secret"`
	AwsSession string `key:"aws.access.session"`
}

func (c connector) Ingest(ctx context.Context, source sources.Source, olap drivers.OLAPStore) error {
	var conf AWSS3Config
	err := connectors.ValidatePropertiesAndExtract(source, c.Spec(), &conf)
	if err != nil {
		return err
	}

	supported, _, err := olap.Ingest(ctx, source, conf)
	if supported {
		return err
	}
	// TODO: download the file and ingest as local file
	return errors.New("OLAP doesnt support s3 file")
}

func (c connector) Validate(source sources.Source) error {
	return nil
}

func (c connector) Spec() []sources.Property {
	return spec
}
