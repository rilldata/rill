package gcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	rillblob "github.com/rilldata/rill/runtime/connectors/blob"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultPageSize = 20

func init() {
	connectors.Register("gcs", Connector{})
}

var errNoCredentials = errors.New("empty credentials: set `google_application_credentials` env variable")

var spec = connectors.Spec{
	DisplayName:        "Google Cloud Storage",
	Description:        "Connect to Google Cloud Storage.",
	ServiceAccountDocs: "https://docs.rilldata.com/deploy/credentials/gcs",
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
			Href:        "https://docs.rilldata.com/develop/import-data#configure-credentials-for-gcs",
		},
	},
	ConnectorVariables: []connectors.VariableSchema{
		{
			Key:  "google_application_credentials",
			Help: "Enter path of file to load from.",
			ValidateFunc: func(any interface{}) error {
				val := any.(string)
				if val == "" {
					// user can chhose to leave empty for public sources
					return nil
				}

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
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

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
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

type Connector struct{}

func (c Connector) Spec() connectors.Spec {
	return spec
}

// ConsumeAsIterator returns a file iterator over objects stored in gcs.
// The credential json is read from a env variable google_application_credentials.
// Additionally in case `env.AllowHostCredentials` is true it looks for "Application Default Credentials" as well
func (c Connector) ConsumeAsIterator(ctx context.Context, env *connectors.Env, source *connectors.Source, l *zap.Logger) (connectors.FileIterator, error) {
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

	iter, err := rillblob.NewIterator(ctx, bucketObj, opts, l)
	if err != nil {
		apiError := &googleapi.Error{}
		// in cases when no creds are passed
		if errors.As(err, &apiError) && apiError.Code == http.StatusUnauthorized {
			return nil, connectors.NewPermissionDeniedError(fmt.Sprintf("can't access remote source %q err: %v", source.Name, apiError))
		}

		// StatusUnauthorized when incorrect key is passsed
		// StatusBadRequest when key doesn't have a valid credentials file
		retrieveError := &oauth2.RetrieveError{}
		if errors.As(err, &retrieveError) && (retrieveError.Response.StatusCode == http.StatusUnauthorized || retrieveError.Response.StatusCode == http.StatusBadRequest) {
			return nil, connectors.NewPermissionDeniedError(fmt.Sprintf("can't access remote source %q err: %v", source.Name, retrieveError))
		}
	}

	return iter, err
}

func (c Connector) HasAnonymousAccess(ctx context.Context, env *connectors.Env, source *connectors.Source) (bool, error) {
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

func (c Connector) ListBuckets(ctx context.Context, req *runtimev1.GCSListBucketsRequest, env *connectors.Env) ([]string, string, error) {
	credentials, err := resolvedCredentials(ctx, env)
	if err != nil {
		return nil, "", err
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, "", err
	}
	defer client.Close()

	projectID := credentials.ProjectID
	if projectID == "" {
		f := &credentialsFile{}
		if err := json.Unmarshal(credentials.JSON, f); err != nil {
			return nil, "", err
		}

		projectID = f.getProjectID()
	}

	pageSize := int(req.GetPageSize())
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	pager := iterator.NewPager(client.Buckets(ctx, projectID), pageSize, req.GetPageToken())
	buckets := make([]*storage.BucketAttrs, 0)
	next, err := pager.NextPage(&buckets)
	if err != nil {
		return nil, "", err
	}

	names := make([]string, len(buckets))
	for i := 0; i < len(buckets); i++ {
		names[i] = buckets[i].Name
	}
	return names, next, nil
}

func (c Connector) ListObjects(ctx context.Context, req *runtimev1.GCSListObjectsRequest, env *connectors.Env) ([]*runtimev1.GCSObject, string, error) {
	client, err := createClient(ctx, env)
	if err != nil {
		return nil, "", err
	}

	bucket, err := gcsblob.OpenBucket(ctx, client, req.GetBucket(), nil)
	if err != nil {
		return nil, "", err
	}
	defer bucket.Close()

	pageSize := int(req.GetPageSize())
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	var pageToken []byte
	if req.GetPageToken() == "" {
		pageToken = blob.FirstPageToken
	} else {
		pageToken = []byte(req.GetPageToken())
	}

	objects, nextToken, err := bucket.ListPage(ctx, pageToken, pageSize, &blob.ListOptions{
		Prefix:    req.Prefix,
		Delimiter: req.Delimitter,
		BeforeList: func(as func(interface{}) bool) error {
			var q *storage.Query
			if as(&q) {
				q.StartOffset = req.GetStartOffset()
				q.EndOffset = req.GetEndOffset()
			} else {
				panic("Listobjects failed")
			}
			return nil
		},
	})
	if err != nil {
		return nil, "", err
	}

	gcsObjects := make([]*runtimev1.GCSObject, len(objects))
	for i, object := range objects {
		gcsObjects[i] = &runtimev1.GCSObject{
			Name:       object.Key,
			ModifiedOn: timestamppb.New(object.ModTime),
			Size:       object.Size,
			IsDir:      object.IsDir,
		}
	}
	return gcsObjects, string(nextToken), nil
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
		creds, err := gcp.DefaultCredentials(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "google: could not find default credentials") {
				return nil, errNoCredentials
			}

			return nil, err
		}
		return creds, nil
	}
	return nil, errNoCredentials
}

// credentialsFile is the unmarshalled representation of a credentials file.
type credentialsFile struct {
	Type string `json:"type"`

	// Service Account fields
	ProjectID string `json:"project_id"`

	// External Account fields
	QuotaProjectID string `json:"quota_project_id"`

	// Service account impersonation
	SourceCredentials *credentialsFile `json:"source_credentials"`
}

func (c *credentialsFile) getProjectID() string {
	if c.Type == "impersonated_service_account" {
		return c.SourceCredentials.getProjectID()
	}
	if c.ProjectID != "" {
		return c.ProjectID
	}
	return c.QuotaProjectID
}
