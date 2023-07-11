package s3

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

const defaultPageSize = 20

func init() {
	drivers.Register("s3", driver{})
}

type driver struct{}

// Open implements drivers.Driver
// TODO :: should it open connection here ? The bucket obj returned from go cdk doesn't make network call either till we make actual call
func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	conn := &Connection{
		config: config,
		logger: logger,
	}
	return conn, nil
}

// Drop implements drivers.Driver
func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

type Connection struct {
	// config holds aws_access_key_id, aws_secret_access_key, aws_secret_access_key, allow_host_access
	config map[string]any
	logger *zap.Logger
}

var _ drivers.Connection = &Connection{}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "s3"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	return c.config
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return nil
}

// Registry implements drivers.Connection.
func (c *Connection) RegistryStore() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *Connection) CatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *Connection) RepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *Connection) OLAPStore() (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (c *Connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *Connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return c, true
}

// AsTransporter implements drivers.Connection.
func (c *Connection) AsTransporter(from, to drivers.Connection) (drivers.Transporter, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsConnector implements drivers.Connection.
func (c *Connection) AsConnector() (drivers.Connector, bool) {
	return c, true
}

type config struct {
	Path                  string `mapstructure:"path"`
	AWSRegion             string `mapstructure:"region"`
	GlobMaxTotalSize      int64  `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int    `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64  `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int    `mapstructure:"glob.page_size"`
	S3Endpoint            string `mapstructure:"endpoint"`
	url                   *globutil.URL
}

func parseConfig(props map[string]any) (*config, error) {
	conf := &config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}

	if !doublestar.ValidatePattern(conf.Path) {
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}

	url, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", conf.Path, err)
	}

	if url.Scheme != "s3" {
		return nil, fmt.Errorf("invalid s3 path %q, should start with s3://", conf.Path)
	}
	conf.url = url
	return conf, nil
}

// DownloadFiles returns a file iterator over objects stored in s3.
//
// The credentials are read from following configs
//   - aws_access_key_id
//   - aws_secret_access_key
//   - aws_session_token
//
// Additionally in case ALLOW_HOST_CREDENTIALS is true it looks for credentials stored on host machine as well
func (c *Connection) DownloadFiles(ctx context.Context, src *drivers.BucketSource) (drivers.FileIterator, error) {
	conf, err := parseConfig(src.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	creds, err := c.getCredentials()
	if err != nil {
		return nil, err
	}

	bucketObj, err := c.openBucket(ctx, conf, conf.url.Host, creds)
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
		ExtractPolicy:         src.ExtractPolicy,
	}

	it, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		var failureErr awserr.RequestFailure
		if !errors.As(err, &failureErr) {
			return nil, err
		}

		// aws returns StatusForbidden in cases like no creds passed, wrong creds passed and incorrect bucket
		// r2 returns StatusBadRequest in all cases above
		// we try again with anonymous credentials in case bucket is public
		if (failureErr.StatusCode() == http.StatusForbidden || failureErr.StatusCode() == http.StatusBadRequest) && creds != credentials.AnonymousCredentials {
			c.logger.Info("s3 list objects failed, re-trying with anonymous credential", zap.Error(err), observability.ZapCtx(ctx))
			creds = credentials.AnonymousCredentials
			bucketObj, bucketErr := c.openBucket(ctx, conf, conf.url.Host, creds)
			if bucketErr != nil {
				return nil, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, bucketErr)
			}

			it, err = rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
		}

		// check again
		if errors.As(err, &failureErr) && (failureErr.StatusCode() == http.StatusForbidden || failureErr.StatusCode() == http.StatusBadRequest) {
			return nil, drivers.NewPermissionDeniedError(fmt.Sprintf("can't access remote err: %v", failureErr))
		}
	}

	return it, err
}

func (c *Connection) openBucket(ctx context.Context, conf *config, bucket string, creds *credentials.Credentials) (*blob.Bucket, error) {
	sess, err := c.getAwsSessionConfig(ctx, conf, bucket, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return s3blob.OpenBucket(ctx, sess, bucket, nil)
}

func (c *Connection) getAwsSessionConfig(ctx context.Context, conf *config, bucket string, creds *credentials.Credentials) (*session.Session, error) {
	// If S3Endpoint is set, we assume we're targeting an S3 compatible API (but not AWS)
	if len(conf.S3Endpoint) > 0 {
		region := conf.AWSRegion
		if region == "" {
			// Set the default region for bwd compatibility reasons
			// cloudflare and minio ignore if us-east-1 is set, not tested for others
			region = "us-east-1"
		}
		return session.NewSession(&aws.Config{
			Region:           aws.String(region),
			Endpoint:         &conf.S3Endpoint,
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      creds,
		})
	}
	// The logic below is AWS-specific, so we ignore it when conf.S3Endpoint is set
	// The complexity below relates to AWS being pretty strict about regions (probably to avoid unexpected cross-region traffic).

	// If the user explicitly set a region, we use that
	if conf.AWSRegion != "" {
		return session.NewSession(&aws.Config{
			Region:      aws.String(conf.AWSRegion),
			Credentials: creds,
		})
	}

	sharedConfigState := session.SharedConfigDisable
	if val, ok := c.config["allow_host_access"]; ok && val.(bool) {
		sharedConfigState = session.SharedConfigEnable // Tells to look for default region set with `aws configure`
	}
	// Create a session that tries to infer the region from the environment
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: sharedConfigState,
		Config: aws.Config{
			Credentials: creds,
		},
	})
	if err != nil {
		return nil, err
	}

	// If no region was found, we default to us-east-1 (which will be used to resolve the lookup in the next step)
	if sess.Config.Region == nil || *sess.Config.Region == "" {
		sess = sess.Copy(&aws.Config{Region: aws.String("us-east-1")})
	}

	// Bucket names are globally unique, but requests will fail if their region doesn't match the one configured in the session.
	// So we do a lookup for the bucket's region and configure the session to use that.
	reg, err := s3manager.GetBucketRegion(ctx, sess, bucket, "")
	if err != nil {
		return nil, err
	}
	if reg != "" {
		sess = sess.Copy(&aws.Config{Region: aws.String(reg)})
	}

	return sess, nil
}

func (c *Connection) getCredentials() (*credentials.Credentials, error) {
	providers := make([]credentials.Provider, 0)

	staticProvider := &credentials.StaticProvider{}
	staticProvider.AccessKeyID = c.config["aws_access_key_id"].(string)
	staticProvider.SecretAccessKey = c.config["aws_secret_access_key"].(string)
	staticProvider.SessionToken = c.config["aws_session_token"].(string)
	staticProvider.ProviderName = credentials.StaticProviderName
	// in case user doesn't set access key id and secret access key the credentials retreival will fail
	// the credential lookup will proceed to next provider in chain
	providers = append(providers, staticProvider)

	if val, ok := c.config["allow_host_access"]; ok && val.(bool) {
		// allowed to access host credentials so we add them in chain
		// The chain used here is a duplicate of defaults.CredProviders(), but without the remote credentials lookup (since they resolve too slowly).
		providers = append(providers, &credentials.EnvProvider{}, &credentials.SharedCredentialsProvider{Filename: "", Profile: ""})
	}
	// Find credentials to use.
	creds := credentials.NewChainCredentials(providers)
	if _, err := creds.Get(); err != nil {
		if !errors.Is(err, credentials.ErrNoValidProvidersFoundInChain) {
			return nil, err
		}
		// If no local credentials are found, you must explicitly set AnonymousCredentials to fetch public objects.
		// AnonymousCredentials can't be chained, so we try to resolve local creds, and use anon if none were found.
		creds = credentials.AnonymousCredentials
	}

	return creds, nil
}
