package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

var spec = drivers.Spec{
	DisplayName: "Amazon S3",
	Description: "Connect to AWS S3 Storage.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/data-source/s3",
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
		{
			Key:         "region",
			Type:        drivers.StringPropertyType,
			DisplayName: "Region",
			Description: "AWS Region for the bucket.",
			Placeholder: "us-east-1",
			Required:    false,
			Hint:        "Rill will use the default region in your local AWS config, unless set here.",
		},
		{
			Key:         "endpoint",
			Type:        drivers.StringPropertyType,
			DisplayName: "Endpoint",
			Description: "Override S3 endpoint URL",
			Placeholder: "https://s3.example.com",
			Required:    false,
			Hint:        "Overrides the S3 endpoint to connect to. This should only be used to connect to S3 compatible services, such as Cloudflare R2 or MinIO.",
		},
		{
			Key:         "aws_role_arn",
			Type:        drivers.StringPropertyType,
			Secret:      true,
			Description: "AWS Role ARN to assume",
		},
		{
			Key:         "aws_role_session_name",
			Type:        drivers.StringPropertyType,
			Secret:      true,
			Description: "Optional session name to use when assuming an AWS role. Defaults to 'rill-session'.",
		},
		{
			Key:         "aws_external_id",
			Type:        drivers.StringPropertyType,
			Secret:      true,
			Description: "Optional external ID to use when assuming an AWS role for cross-account access.",
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
			Key:         "name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Source name",
			Description: "The name of the source",
			Placeholder: "my_new_source",
			Required:    true,
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

type ConfigProperties struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_access_token"`
	Region          string `mapstructure:"region"`
	Endpoint        string `mapstructure:"endpoint"`
	RoleARN         string `mapstructure:"aws_role_arn"`
	RoleSessionName string `mapstructure:"aws_role_session_name"`
	ExternalID      string `mapstructure:"aws_external_id"`
	// A list of bucket path prefixes that this connector is allowed to access.
	// Useful when different buckets or bucket prefixes use different credentials,
	// allowing the system to select the appropriate connector based on the bucket path.
	// Example formats: `s3://my-bucket/` `s3://my-bucket/path/` `s3://my-bucket/path/prefix`
	PathPrefixes    []string `mapstructure:"path_prefixes"`
	AllowHostAccess bool     `mapstructure:"allow_host_access"`
}

// Open implements drivers.Driver
func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("s3 driver can't be shared")
	}

	cfg := &ConfigProperties{}
	err := mapstructure.WeakDecode(config, cfg)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config:  cfg,
		storage: st,
		logger:  logger,
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type Connection struct {
	config  *ConfigProperties
	storage *storage.Client
	logger  *zap.Logger
}

var _ drivers.Handle = &Connection{}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	if c.config.Endpoint == "" {
		stsClient, err := getSTSClient(ctx, c.config)
		if err != nil {
			return err
		}
		_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			return fmt.Errorf("GetCallerIdentity failed: %w", err)
		}
	}
	return nil
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "s3"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any)
	err := mapstructure.Decode(c.config, &m)
	if err != nil {
		c.logger.Warn("error in generating s3 config", zap.Error(err))
	}
	return m
}

// ParsedConfig returns the parsed S3 configuration.
func (c *Connection) ParsedConfig() *ConfigProperties {
	cpy := *c.config
	return &cpy
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

// AsInformationSchema implements drivers.Connection.
func (c *Connection) AsInformationSchema() (drivers.InformationSchema, bool) {
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

// AsModelExecutor implements drivers.Handle.
func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return c, true
}

// AsFileStore implements drivers.Connection.
func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *Connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// BucketRegion returns the region to use for the given bucket.
func BucketRegion(ctx context.Context, confProp *ConfigProperties, bucket string) (string, error) {
	cfg, err := getAWSConfig(ctx, confProp)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config for bucket region detection (set `region` in s3 connector yaml): %w", err)
	}
	return bucketRegionFromConfig(ctx, cfg, confProp, bucket)
}

func bucketRegionFromConfig(ctx context.Context, cfg aws.Config, confProp *ConfigProperties, bucket string) (string, error) {
	// If S3Endpoint is set, we assume we're targeting an S3 compatible API (but not AWS)
	if confProp.Endpoint != "" {
		if confProp.Region == "" {
			// Set the default region for bwd compatibility reasons
			// cloudflare and minio ignore if us-east-1 is set, not tested for others
			return "us-east-1", nil
		}
	}
	if confProp.Region != "" {
		return confProp.Region, nil
	}

	// default region is required to even to detect bucket region
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	// Use the manager utility to detect the correct bucket region.
	region, err := manager.GetBucketRegion(ctx, s3.NewFromConfig(cfg), bucket)
	if err != nil {
		return "", fmt.Errorf(
			"failed to detect bucket region for %s (set `region` in s3 connector yaml): %w",
			bucket, err,
		)
	}

	return region, nil
}

func getAnonymousS3Client(ctx context.Context, confProp *ConfigProperties, bucket string) (*s3.Client, error) {
	cfg := aws.Config{
		Credentials: aws.AnonymousCredentials{},
	}
	region := confProp.Region
	// If the region is not explicitly provided in the config,
	// try to automatically detect it from the bucket
	if region == "" && bucket != "" {
		var err error
		region, err = bucketRegionFromConfig(ctx, cfg, confProp, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to detect bucket region: %w", err)
		}
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		if confProp.Endpoint != "" {
			o.BaseEndpoint = aws.String(confProp.Endpoint)
			o.UsePathStyle = true
		}
		o.Region = region
	}), nil
}

func getS3Client(ctx context.Context, confProp *ConfigProperties, bucket string) (*s3.Client, error) {
	cfg, err := getAWSConfig(ctx, confProp)
	if err != nil {
		return nil, err
	}
	region := confProp.Region
	// If the region is not explicitly provided in the config,
	// try to automatically detect it from the bucket
	if region == "" && bucket != "" {
		var err error
		region, err = bucketRegionFromConfig(ctx, cfg, confProp, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to detect bucket region: %w", err)
		}
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		if confProp.Endpoint != "" {
			o.BaseEndpoint = aws.String(confProp.Endpoint)
			o.UsePathStyle = true
		}
		o.Region = region
	}), nil
}

func getSTSClient(ctx context.Context, confProp *ConfigProperties) (*sts.Client, error) {
	cfg, err := getAWSConfig(ctx, confProp)
	if err != nil {
		return nil, err
	}
	return sts.NewFromConfig(cfg, func(o *sts.Options) {
		if confProp.Endpoint != "" {
			o.BaseEndpoint = aws.String(confProp.Endpoint)
		}
		// set default region to "us-east-1"
		if cfg.Region == "" {
			o.Region = "us-east-1"
		}
	}), nil
}

func getAWSConfig(ctx context.Context, confProp *ConfigProperties) (aws.Config, error) {
	provider, err := newCredentialsProvider(ctx, confProp)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to get AWS credentials: %w", err)
	}

	opts := []func(*config.LoadOptions) error{
		config.WithCredentialsProvider(provider),
	}
	if confProp.Region != "" {
		opts = append(opts, config.WithRegion(confProp.Region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return cfg, nil
}

// newCredentialsProvider returns credentials for connecting to AWS.
func newCredentialsProvider(ctx context.Context, confProp *ConfigProperties) (aws.CredentialsProvider, error) {
	// 1. If a role ARN is provided, assume it.
	if confProp.RoleARN != "" {
		return assumeRole(ctx, confProp)
	}

	// 1. Explicit static credentials
	if confProp.AccessKeyID != "" && confProp.SecretAccessKey != "" {
		return aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			confProp.AccessKeyID,
			confProp.SecretAccessKey,
			confProp.SessionToken,
		)), nil
	}

	// 3. Allow host-based credentials
	if confProp.AllowHostAccess {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		// Optional: pre-fetch credentials to ensure valid keys exist
		if creds, err := cfg.Credentials.Retrieve(ctx); err != nil || !creds.HasKeys() {
			// fallback to anonymous if nothing found
			return aws.AnonymousCredentials{}, nil
		}

		return cfg.Credentials, nil
	}

	// 3. Fallback to anonymous credentials
	return aws.AnonymousCredentials{}, nil
}

// assumeRole returns a credentials provider that assumes the role specified by the ARN using AWS SDK v2.
// It uses stscreds.NewAssumeRoleProvider so credentials are automatically refreshed before expiration.
func assumeRole(ctx context.Context, confProp *ConfigProperties) (aws.CredentialsProvider, error) {
	// Add session name if specified
	sessionName := confProp.RoleSessionName
	if sessionName == "" {
		sessionName = "rill-session"
	}

	region := confProp.Region
	if region == "" {
		region = "us-east-1"
	}

	var credsProvider aws.CredentialsProvider

	if confProp.AccessKeyID != "" && confProp.SecretAccessKey != "" {
		// Use explicit static credentials
		credsProvider = credentials.NewStaticCredentialsProvider(
			confProp.AccessKeyID,
			confProp.SecretAccessKey,
			confProp.SessionToken,
		)
	} else if confProp.AllowHostAccess {
		hostCfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load host credentials: %w", err)
		}
		credsProvider = hostCfg.Credentials
	} else {
		// No valid credentials to assume role
		return nil, fmt.Errorf("cannot assume role: no base credentials available")
	}

	// Load AWS config with the chosen provider
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credsProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create STS client with explicit configuration
	stsClient := sts.NewFromConfig(cfg)

	// Create an assume role provider that automatically refreshes credentials before expiration
	assumeRoleProvider := stscreds.NewAssumeRoleProvider(stsClient, confProp.RoleARN, func(o *stscreds.AssumeRoleOptions) {
		o.RoleSessionName = sessionName
		// Add external ID if provided to mitigate confused deputy problem
		if confProp.ExternalID != "" {
			o.ExternalID = &confProp.ExternalID
		}
	})

	// Wrap in a credentials cache so multiple calls share refreshed credentials
	return aws.NewCredentialsCache(assumeRoleProvider), nil
}
