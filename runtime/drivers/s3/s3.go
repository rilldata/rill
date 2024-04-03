package s3

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

var spec = drivers.Spec{
	DisplayName: "Amazon S3",
	Description: "Connect to AWS S3 Storage.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/s3",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "aws_access_key_id",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
		{
			Key:    "aws_secret_access_key",
			Type:   drivers.StringPropertyType,
			Secret: true,
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			DisplayName: "S3 URI",
			Description: "Path to file on the disk.",
			Placeholder: "s3://bucket-name/path/to/file.csv",
			Required:    true,
			Hint:        "Glob patterns are supported",
		},
		{
			Key:         "region",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS region",
			Description: "AWS Region for the bucket.",
			Placeholder: "us-east-1",
			Required:    false,
			Hint:        "Rill will use the default region in your local AWS config, unless set here.",
		},
		{
			Key:         "endpoint",
			Type:        drivers.StringPropertyType,
			DisplayName: "Endpoint URL",
			Description: "Override S3 Endpoint URL",
			Placeholder: "https://my.s3.server.com",
			Required:    false,
			Hint:        "Overrides the S3 endpoint to connect to. This should only be used to connect to S3-compatible services, such as Cloudflare R2 or MinIO.",
		},
		{
			Key:         "aws.credentials",
			Type:        drivers.InformationalPropertyType,
			DisplayName: "AWS credentials",
			Description: "AWS credentials inferred from your local environment.",
			Hint:        "Set your local credentials: <code>aws configure</code> Click to learn more.",
			DocsURL:     "https://docs.rilldata.com/reference/connectors/s3#local-credentials",
		},
	},
	ImplementsObjectStore: true,
}

const defaultPageSize = 20

func init() {
	drivers.Register("s3", driver{})
	drivers.RegisterAsConnector("s3", driver{})
}

type driver struct{}

var _ drivers.Driver = driver{}

type configProperties struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_access_token"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
	RetainFiles     bool   `mapstructure:"retain_files"`
}

// Open implements drivers.Driver
func (d driver) Open(cfgMap map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("s3 driver can't be shared")
	}

	cfg := &configProperties{}
	err := mapstructure.WeakDecode(cfgMap, cfg)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config: cfg,
		logger: logger,
	}
	return conn, nil
}

// Drop implements drivers.Driver
func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	conf, err := parseSourceProperties(props)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}

	c, err := d.Open(map[string]any{}, false, activity.NewNoopClient(), logger)
	if err != nil {
		return false, err
	}

	conn := c.(*Connection)
	bucketObj, err := conn.openBucket(ctx, conf, conf.url.Host, credentials.AnonymousCredentials)
	if err != nil {
		return false, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}
	defer bucketObj.Close()

	return bucketObj.IsAccessible(ctx)
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type Connection struct {
	// config is input configs passed to driver.Open
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Handle = &Connection{}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "s3"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return nil
}

// AsRegistry implements drivers.Connection.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
func (c *Connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *Connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *Connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Connection.
func (c *Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
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
func (c *Connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
func (c *Connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

type sourceProperties struct {
	Path                  string         `mapstructure:"path"`
	URI                   string         `mapstructure:"uri"`
	AWSRegion             string         `mapstructure:"region"`
	GlobMaxTotalSize      int64          `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int            `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64          `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int            `mapstructure:"glob.page_size"`
	S3Endpoint            string         `mapstructure:"endpoint"`
	Extract               map[string]any `mapstructure:"extract"`
	BatchSize             string         `mapstructure:"batch_size"`
	url                   *globutil.URL
	extractPolicy         *rillblob.ExtractPolicy
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.WeakDecode(props, conf)
	if err != nil {
		return nil, err
	}

	// Backwards compatibility for "uri" renamed to "path"
	if conf.URI != "" {
		conf.Path = conf.URI
	}

	if !doublestar.ValidatePattern(conf.Path) {
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}

	url, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", conf.Path, err)
	}
	conf.url = url

	if url.Scheme != "s3" {
		return nil, fmt.Errorf("invalid s3 path %q, should start with s3://", conf.Path)
	}

	conf.extractPolicy, err = rillblob.ParseExtractPolicy(conf.Extract)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extract config: %w", err)
	}

	return conf, nil
}

// DownloadFiles returns a file iterator over objects stored in s3.
//
// The credentials are read from following configs
//   - aws_access_key_id
//   - aws_secret_access_key
//   - aws_session_token
//
// Additionally in case allow_host_credentials is true it looks for credentials stored on host machine as well
func (c *Connection) DownloadFiles(ctx context.Context, src map[string]any) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(src)
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

	var batchSize datasize.ByteSize
	if conf.BatchSize == "-1" {
		batchSize = math.MaxInt64 // download everything in one batch
	} else {
		batchSize, err = datasize.ParseString(conf.BatchSize)
		if err != nil {
			return nil, err
		}
	}
	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         conf.extractPolicy,
		BatchSizeBytes:        int64(batchSize.Bytes()),
		KeepFilesUntilClose:   conf.BatchSize == "-1",
		RetainFiles:           c.config.RetainFiles,
	}

	it, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		// TODO :: fix this for single file access. for single file first call only happens during download
		var failureErr awserr.RequestFailure
		if !errors.As(err, &failureErr) {
			return nil, err
		}

		// aws returns StatusForbidden in cases like no creds passed, wrong creds passed and incorrect bucket
		// r2 returns StatusBadRequest in all cases above
		// we try again with anonymous credentials in case bucket is public
		if (failureErr.StatusCode() == http.StatusForbidden || failureErr.StatusCode() == http.StatusBadRequest) && creds != credentials.AnonymousCredentials {
			c.logger.Debug("s3 list objects failed, re-trying with anonymous credential", zap.Error(err), observability.ZapCtx(ctx))
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

func (c *Connection) openBucket(ctx context.Context, conf *sourceProperties, bucket string, creds *credentials.Credentials) (*blob.Bucket, error) {
	sess, err := c.getAwsSessionConfig(ctx, conf, bucket, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return s3blob.OpenBucket(ctx, sess, bucket, nil)
}

func (c *Connection) getAwsSessionConfig(ctx context.Context, conf *sourceProperties, bucket string, creds *credentials.Credentials) (*session.Session, error) {
	// If S3Endpoint is set, we assume we're targeting an S3 compatible API (but not AWS)
	if conf.S3Endpoint != "" {
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
	if c.config.AllowHostAccess {
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
	staticProvider.AccessKeyID = c.config.AccessKeyID
	staticProvider.SecretAccessKey = c.config.SecretAccessKey
	staticProvider.SessionToken = c.config.SessionToken
	staticProvider.ProviderName = credentials.StaticProviderName
	// in case user doesn't set access key id and secret access key the credentials retreival will fail
	// the credential lookup will proceed to next provider in chain
	providers = append(providers, staticProvider)

	if c.config.AllowHostAccess {
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
