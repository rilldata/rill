package s3

import (
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
)

func init() {
	connectors.Register("s3", connector{})
}

var spec = connectors.Spec{
	DisplayName: "S3",
	Description: "Connector for AWS S3",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "Path",
			Description: "Path to file on the disk.",
			Placeholder: "s3://<bucket>/<file>",
			Type:        connectors.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "aws.region",
			DisplayName: "AWS Region",
			Description: "AWS Region for the bucket.",
			Placeholder: "us-east-1",
			Type:        connectors.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "aws.access.key",
			DisplayName: "AWS Access Key",
			Description: "",
			Placeholder: "",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
		{
			Key:         "aws.access.secret",
			DisplayName: "AWS Access Secret",
			Description: "",
			Placeholder: "",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
		{
			Key:         "aws.access.session",
			DisplayName: "AWS Access Session Token",
			Description: "A session token is an alternative to an access key/secret pair",
			Placeholder: "",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
	},
}

type Config struct {
	Path       string `mapstructure:"path"`
	AWSRegion  string `mapstructure:"aws.region"`
	AWSKey     string `mapstructure:"aws.access.key"`
	AWSSecret  string `mapstructure:"aws.access.secret"`
	AWSSession string `mapstructure:"aws.access.session"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, &conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}
