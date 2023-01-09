package gcs

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"google.golang.org/api/option"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob" // blank import required for bucket functions
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
			Hint:        "Note that glob patterns aren't yet supported",
		},
		{
			Key:         "gcp.credentials",
			DisplayName: "GCP credentials",
			Description: "GCP credentials inferred from your local environment.",
			Type:        connectors.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>gcloud auth application-default login</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/using-rill/import-data#setting-google-gcs-credentials",
		},
	},
}

type Config struct {
	Path          string `key:"path"`
	MaxSize       int64  `mapstructure:"glob.max_size"`
	MaxDownload   int    `mapstructure:"glob.max_download"`
	MaxIterations int64  `mapstructure:"glob.max_iterations"`
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

func (c connector) ConsumeAsFile(ctx context.Context, env *connectors.Env, source *connectors.Source) (string, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return "", fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := getGcsClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	bucket, object, extension, err := gcsURLParts(conf.Path)
	if err != nil {
		return "", fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("Object(%q).NewReader: %w", object, err)
	}
	defer rc.Close()

	return fileutil.CopyToTempFile(rc, source.Name, extension)
}

func gcsURLParts(path string) (string, string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", "", "", err
	}
	return u.Host, strings.Replace(u.Path, "/", "", 1), fileutil.FullExt(u.Path), nil
}

func getGcsClient(ctx context.Context) (*storage.Client, error) {
	client, err := storage.NewClient(ctx)
	if err == nil || !strings.Contains(err.Error(), "google: could not find default credentials") {
		return client, err
	}
	client, err = storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Responsibility of caller to close bucketObject
func (c connector) PrepareBlob(ctx context.Context, source *connectors.Source) (*rillblob.BlobHandler, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if !doublestar.ValidatePattern(conf.Path) {
		// ideally this should be validated at much earlier stage
		// keeping it here to have gcs specific validations
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}

	bucket, glob, _, err := gcsURLParts(conf.Path)
	bucket = fmt.Sprintf("gs://%s", bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}
	bucketObj, err := blob.OpenBucket(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %s, %w", bucket, err)
	}
	fetchConfigs := rillblob.FetchConfigs{
		MaxSize:       conf.MaxSize,
		MaxDownload:   conf.MaxDownload,
		MaxIterations: conf.MaxIterations,
	}
	return rillblob.FetchBlobHandler(ctx, bucketObj, fetchConfigs, glob, bucket)
}
