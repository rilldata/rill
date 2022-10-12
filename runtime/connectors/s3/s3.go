package s3

import (
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
)

func init() {
	connectors.Register("s3", connector{})
}

var spec = []connectors.PropertySchema{
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
}

type Config struct {
	Path       string `key:"path"`
	AWSRegion  string `key:"aws.region"`
	AWSKey     string `key:"aws.access.key"`
	AWSSecret  string `key:"aws.access.secret"`
	AWSSession string `key:"aws.access.session"`
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

func (c connector) Spec() []connectors.PropertySchema {
	return spec
}
