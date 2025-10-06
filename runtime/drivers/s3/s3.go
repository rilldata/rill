package s3

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
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
	DocsURL:     "https://docs.rilldata.com/connect/data-source/s3",
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
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
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
	cfg, err := c.GetAWSConfig(ctx)
	if err != nil {
		return err
	}
	stsClient := c.GetSTSClient(cfg)
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("GetCallerIdentity failed: %w", err)
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

func (c *Connection) GetS3Client(cfg aws.Config) *s3.Client {
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		if c.config.Endpoint != "" {
			o.BaseEndpoint = aws.String(c.config.Endpoint)
			o.UsePathStyle = true
		}
	})
}

func (c *Connection) GetSTSClient(cfg aws.Config) *sts.Client {
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	return sts.NewFromConfig(cfg, func(o *sts.Options) {
		if c.config.Endpoint != "" {
			o.BaseEndpoint = aws.String(c.config.Endpoint)
		}
	})
}

func (c *Connection) GetAWSConfig(ctx context.Context) (aws.Config, error) {
	provider, err := c.newCredentialsProvider(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to get AWS credentials: %w", err)
	}

	opts := []func(*config.LoadOptions) error{
		config.WithCredentialsProvider(provider),
	}
	if c.config.Region != "" {
		opts = append(opts, config.WithRegion(c.config.Region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return cfg, nil
}

// newCredentialsProvider returns credentials for connecting to AWS.
func (c *Connection) newCredentialsProvider(ctx context.Context) (aws.CredentialsProvider, error) {
	// 1. If a role ARN is provided, assume it.
	if c.config.RoleARN != "" {
		return c.assumeRole(ctx)
	}

	// 1. Explicit static credentials
	if c.config.AccessKeyID != "" && c.config.SecretAccessKey != "" {
		return aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			c.config.AccessKeyID,
			c.config.SecretAccessKey,
			c.config.SessionToken,
		)), nil
	}

	// 3. Allow host-based credentials, but only local (env + shared files)
	if c.config.AllowHostAccess {
		os.Setenv("AWS_EC2_METADATA_DISABLED", "true") // Disable remote lookups
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
func (c *Connection) assumeRole(ctx context.Context) (aws.CredentialsProvider, error) {
	// Add session name if specified
	sessionName := c.config.RoleSessionName
	if sessionName == "" {
		sessionName = "rill-session"
	}

	loadOpts := []func(*config.LoadOptions) error{
		config.WithSharedConfigFiles([]string{}), // Disable shared config (~/.aws/config, ~/.aws/credentials)
	}

	// Add region if specified, otherwise default to us-east-1
	if c.config.Region != "" {
		loadOpts = append(loadOpts, config.WithRegion(c.config.Region))
	} else {
		loadOpts = append(loadOpts, config.WithRegion("us-east-1"))
	}

	// Add static credentials if explicitly provided (AccessKeyID, SecretAccessKey, SessionToken)
	if c.config.AccessKeyID != "" && c.config.SecretAccessKey != "" {
		loadOpts = append(loadOpts,
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				c.config.AccessKeyID,
				c.config.SecretAccessKey,
				c.config.SessionToken,
			)),
		)
	}

	// Create AWS config with explicit configuration
	cfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create STS client with explicit configuration
	stsClient := sts.NewFromConfig(cfg)

	// Create an assume role provider that automatically refreshes credentials before expiration
	assumeRoleProvider := stscreds.NewAssumeRoleProvider(stsClient, c.config.RoleARN, func(o *stscreds.AssumeRoleOptions) {
		o.RoleSessionName = sessionName
		// Add external ID if provided to mitigate confused deputy problem
		if c.config.ExternalID != "" {
			o.ExternalID = &c.config.ExternalID
		}
	})

	// Wrap in a credentials cache so multiple calls share refreshed credentials
	return aws.NewCredentialsCache(assumeRoleProvider), nil
}

func GetAnonymousS3Client(region, endpoint string) *s3.Client {
	if region == "" {
		region = "us-east-1"
	}
	cfg := aws.Config{
		Region:      region,
		Credentials: aws.AnonymousCredentials{},
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})
}
