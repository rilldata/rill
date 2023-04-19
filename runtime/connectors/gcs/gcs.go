package gcs

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

func init() {
	connectors.Register("gcs", connector{})
}

var errNoCredentials = errors.New("empty credentials: set `GOOGLE_APPLICATION_CREDENTIALS` env variable")

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
	ConnectorVariables: []connectors.VariableSchema{
		{
			Key:  "GOOGLE_APPLICATION_CREDENTIALS",
			Help: "Enter path of file to load from. Leave blank if public access enabled.",
			ValidateFunc: func(any interface{}) error {
				val := any.(string)
				if val == "" {
					// user can chhose to leave empty for public sources
					return nil
				}

				path, err := fileutil.ExpandHome(val)
				if err != nil {
					return err
				}

				_, err = os.Stat(path)
				return err
			},
			TransformFunc: func(any interface{}) interface{} {
				val := any.(string)
				if val == "" {
					return ""
				}

				path, err := fileutil.ExpandHome(val)
				if err != nil {
					return err
				}

				// ignoring error since PathError is already validated
				content, _ := os.ReadFile(path)
				return string(content)
			},
		},
	},
}

type Config struct {
	Path                  string `key:"path"`
	GlobMaxTotalSize      int64  `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int    `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64  `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int    `mapstructure:"glob.page_size"`
	url                   *globutil.URL
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if !doublestar.ValidatePattern(conf.Path) {
		// ideally this should be validated at much earlier stage
		// keeping it here to have gcs specific validations
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}
	url, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", conf.Path, err)
	}

	if url.Scheme != "gs" {
		return nil, fmt.Errorf("invalid gcs path %q, should start with gs://", conf.Path)
	}

	conf.url = url
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

// ConsumeAsIterator returns a file iterator over objects stored in gcs.
// The credential json is read from a env variable GOOGLE_APPLICATION_CREDENTIALS.
// Additionally in case `env.AllowHostCredentials` is true it looks for "Application Default Credentials" as well
func (c connector) ConsumeAsIterator(ctx context.Context, env *connectors.Env, source *connectors.Source) (connectors.FileIterator, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := createClient(ctx, env)
	if err != nil {
		return nil, err
	}

	bucketObj, err := gcsblob.OpenBucket(ctx, client, conf.url.Host, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}

	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         source.ExtractPolicy,
		StorageLimitInBytes:   env.StorageLimitInBytes,
	}
	return rillblob.NewIterator(ctx, bucketObj, opts)
}

func (c connector) HasAnonymousAccess(ctx context.Context, env *connectors.Env, source *connectors.Source) (bool, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}

	client := gcp.NewAnonymousHTTPClient(gcp.DefaultTransport())
	bucketObj, err := gcsblob.OpenBucket(ctx, client, conf.url.Host, nil)
	if err != nil {
		return false, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}

	return bucketObj.IsAccessible(ctx)
}

func createClient(ctx context.Context, env *connectors.Env) (*gcp.HTTPClient, error) {
	creds, err := resolvedCredentials(ctx, env)
	if err != nil {
		if !errors.Is(err, errNoCredentials) {
			return nil, err
		}

		// no credentials set, we try with a anonymous client in case user is trying to access public buckets
		return gcp.NewAnonymousHTTPClient(gcp.DefaultTransport()), nil
	}
	// the token source returned from credentials works for all kind of credentials like serviceAccountKey, credentialsKey etc.
	return gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
}

func resolvedCredentials(ctx context.Context, env *connectors.Env) (*google.Credentials, error) {
	if secretJSON := env.Variables["GOOGLE_APPLICATION_CREDENTIALS"]; secretJSON != "" {
		// GOOGLE_APPLICATION_CREDENTIALS is set, use credentials from json string provided by user
		return google.CredentialsFromJSON(ctx, []byte(secretJSON), "https://www.googleapis.com/auth/cloud-platform")
	}
	// GOOGLE_APPLICATION_CREDENTIALS is not set
	if env.AllowHostAccess {
		// use host credentials
		return gcp.DefaultCredentials(ctx)
	}
	return nil, errNoCredentials
}
