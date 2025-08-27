package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
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
	creds, err := c.newCredentials()
	if err != nil {
		return fmt.Errorf("failed to get AWS credentials: %w", err)
	}

	cfg := aws.NewConfig().WithCredentials(creds)
	if c.config.Region != "" {
		cfg = cfg.WithRegion(c.config.Region)
	}

	if c.config.Endpoint != "" {
		cfg = cfg.WithEndpoint(c.config.Endpoint).WithS3ForcePathStyle(true)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	stsClient := sts.New(sess)
	_, err = stsClient.GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
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
func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
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

// newCredentials returns credentials for connecting to AWS.
func (c *Connection) newCredentials() (*credentials.Credentials, error) {
	// If a role ARN is provided, assume the role and return the credentials.
	if c.config.RoleARN != "" {
		assumedCreds, err := c.assumeRole()
		if err != nil {
			return nil, fmt.Errorf("failed to assume role: %w", err)
		}
		return assumedCreds, nil
	}

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

// assumeRole returns a new credentials object that assumes the role specified by the ARN.
func (c *Connection) assumeRole() (*credentials.Credentials, error) {
	// Add session name if specified
	sessionName := c.config.RoleSessionName
	if sessionName == "" {
		sessionName = "rill-session"
	}

	sessOpts := session.Options{
		SharedConfigState: session.SharedConfigDisable, // Disable shared config to prevent loading default config
	}

	// Add region if specified
	if c.config.Region != "" {
		sessOpts.Config.Region = &c.config.Region
	}

	// Add credentials if provided
	if c.config.AccessKeyID != "" && c.config.SecretAccessKey != "" {
		sessOpts.Config.Credentials = credentials.NewStaticCredentials(
			c.config.AccessKeyID,
			c.config.SecretAccessKey,
			c.config.SessionToken,
		)
	}

	// Create session with explicit configuration
	s, err := session.NewSessionWithOptions(sessOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create STS client with explicit session
	stsClient := sts.New(s)

	// Create assume role input with explicit parameters
	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         &c.config.RoleARN,
		RoleSessionName: &sessionName,
	}

	// Add external ID if provided to mitigate confused deputy problem
	if c.config.ExternalID != "" {
		assumeRoleInput.ExternalId = &c.config.ExternalID
	}

	// Assume the role
	result, err := stsClient.AssumeRole(assumeRoleInput)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	// Return static credentials from the assumed role
	return credentials.NewStaticCredentials(
		*result.Credentials.AccessKeyId,
		*result.Credentials.SecretAccessKey,
		*result.Credentials.SessionToken,
	), nil
}
