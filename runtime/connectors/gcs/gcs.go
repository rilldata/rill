package gcs

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
)

func init() {
	connectors.Register("gcs", connector{})
}

var spec = connectors.Spec{
	DisplayName: "GCS",
	Description: "Connect to CSV or Parquet files in a Google Cloud Storage bucket. For private buckets, provide <a href=https://console.cloud.google.com/storage/settings;tab=interoperability target='_blank'>HMAC credentials</a>.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "Path",
			Description: "Path to file on the disk.",
			Placeholder: "gs://bucket-name/path/to/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
			Hint:        "Tip: use glob patterns to select multiple files",
		},
		{
			Key:         "gcp.region",
			DisplayName: "GCP region",
			Description: "GCP Region for the bucket.",
			Placeholder: "us-east-1",
			Type:        connectors.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "gcp.access.key",
			DisplayName: "GCP access Key",
			Description: "",
			Placeholder: "...",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
		{
			Key:         "gcp.access.secret",
			DisplayName: "GCP access secret",
			Description: "",
			Placeholder: "...",
			Type:        connectors.StringPropertyType,
			Required:    false,
		},
	},
}

type Config struct {
	Path      string `key:"path" ignored:"true"`
	GCPRegion string `key:"gcp.region" envconfig:"GCP_DEFAULT_REGION"`
	GCPKey    string `key:"gcp.access.key" envconfig:"GCP_ACCESS_KEY_ID"`
	GCPSecret string `key:"gcp.access.secret" envconfig:"GCP_SECRET_ACCESS_KEY"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := envconfig.Process("gcp", conf)
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
