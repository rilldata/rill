package gcs

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
)

func init() {
	connectors.Register("gcs", connector{})
}

var spec = connectors.Spec{
	DisplayName: "Google Cloud Storage",
	Description: "Connect to Google Cloud Storage.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "GS URI",
			Description: "Path to file on the disk.",
			Placeholder: "gs://bucket-name/path/to/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
			Hint:        "Note that gzipped files & glob patterns aren't yet supported",
		},
		{
			Key:         "gcp.credentials",
			DisplayName: "GCP credentials",
			Description: "GCP credentials inferred from your local environment.",
			Type:        connectors.InformationalPropertyType,
			Hint:        "Set your local credentials: <p><code>gcloud auth application-default login</code></p> Click to learn more.",
			Href:        "https://docs.rilldata.com/import-data#setting-google-gcs-credentials",
		},
	},
}

type Config struct {
	Path string `key:"path"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

func (c connector) ConsumeAsFile(ctx context.Context, source *connectors.Source, callback func(filename string) error) error {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bucket, object, extension, err := getGcsUrlParts(conf.Path)
	if err != nil {
		return fmt.Errorf("failed to parse path %s, %v", conf.Path, err)
	}

	f, err := os.CreateTemp(
		os.TempDir(),
		fmt.Sprintf("%s*.%s", source.Name, extension),
	)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}
	defer rc.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	err = callback(f.Name())
	if err != nil {
		return fmt.Errorf("failed to ingest f, %v", err)
	}

	return nil
}

func getGcsUrlParts(path string) (string, string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", "", "", err
	}

	p := strings.Split(u.Path, ".")

	return u.Host, strings.Replace(u.Path, "/", "", 1), p[len(p)-1], nil
}
