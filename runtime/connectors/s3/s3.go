package s3

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
)

func init() {
	connectors.Register("s3", connector{})
}

var spec = connectors.Spec{
	DisplayName: "S3",
	Description: "Connect to CSV or Parquet files in an Amazon S3 bucket. For private buckets, provide an <a href=https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html target='_blank'>access key</a>.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "S3 URI",
			Description: "Path to file on the disk.",
			Placeholder: "s3://bucket-name/path/to/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
			Hint:        "Tip: use glob patterns to select multiple files",
		},
		{
			Key:         "aws.region",
			DisplayName: "AWS region",
			Description: "AWS Region for the bucket.",
			Placeholder: "us-east-1",
			Type:        connectors.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "aws.access.key",
			DisplayName: "AWS access key",
			Description: "",
			Placeholder: "...",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
		{
			Key:         "aws.access.secret",
			DisplayName: "AWS access secret",
			Description: "",
			Placeholder: "...",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
	},
}

type Config struct {
	Path       string `mapstructure:"path" ignored:"true"`
	AWSRegion  string `mapstructure:"aws.region" envconfig:"AWS_DEFAULT_REGION"`
	AWSKey     string `mapstructure:"aws.access.key" envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecret  string `mapstructure:"aws.access.secret" envconfig:"AWS_SECRET_ACCESS_KEY"`
	AWSSession string `mapstructure:"aws.access.session" ignored:"true"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := envconfig.Process("aws", conf)
	if err != nil {
		return nil, err
	}
	err = mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}
