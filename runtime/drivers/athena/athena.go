package athena

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	types2 "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/tracing/smithyoteltracing"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("athena", driver{})
	drivers.RegisterAsConnector("athena", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Amazon Athena",
	Description: "Connect to Amazon Athena database.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/data-source/athena",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "aws_access_key_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS access key ID",
			Description: "AWS access key ID",
			Placeholder: "your_access_key_id",
			Required:    true,
			Secret:      true,
		},
		{
			Key:         "aws_secret_access_key",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS secret access key",
			Description: "AWS secret access key",
			Placeholder: "your_secret_access_key",
			Required:    true,
			Secret:      true,
		},
		{
			Key:         "output_location",
			Type:        drivers.StringPropertyType,
			DisplayName: "S3 output location",
			Description: "An output location for query result is required either through the workgroup result configuration setting or set here.",
			Placeholder: "s3://bucket-name/path/",
			Required:    true,
		},
	},
	ImplementsWarehouse: true,
}

type driver struct{}

type configProperties struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_access_token"`
	RoleARN         string `mapstructure:"role_arn"`
	RoleSessionName string `mapstructure:"role_session_name"`
	ExternalID      string `mapstructure:"external_id"`
	AWSRegion       string `mapstructure:"region"`
	Workgroup       string `mapstructure:"workgroup"`
	OutputLocation  string `mapstructure:"output_location"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("athena driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config:  conf,
		logger:  logger,
		storage: st,
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type Connection struct {
	config  *configProperties
	logger  *zap.Logger
	storage *storage.Client
}

var _ drivers.Handle = &Connection{}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	client, err := c.getClient(ctx)
	if err != nil {
		return err
	}

	// Execute a simple query to verify connection
	_, err = c.executeQuery(ctx, client, "SELECT 1", c.config.Workgroup, c.config.OutputLocation)
	return err
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "athena"
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

// AsInformationSchema implements drivers.Connection.
func (c *Connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
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
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *Connection) AsWarehouse() (drivers.Warehouse, bool) {
	return c, true
}

// AsNotifier implements drivers.Handle.
func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

func (c *Connection) awsConfig(ctx context.Context, awsRegion string) (aws.Config, error) {
	loadOptions := []func(*config.LoadOptions) error{
		// Setting the default region to an empty string, will result in the default region value being ignored
		config.WithDefaultRegion("us-east-1"),
		// Setting the region to an empty string, will result in the region value being ignored
		config.WithRegion(awsRegion),
	}

	// If one of the static properties is specified: access key, secret key, or session token, use static credentials,
	// Else fallback to the SDK's default credential chain (environment, instance, etc) unless AllowHostAccess is false
	if c.config.AccessKeyID != "" || c.config.SecretAccessKey != "" || c.config.SessionToken != "" {
		p := credentials.NewStaticCredentialsProvider(c.config.AccessKeyID, c.config.SecretAccessKey, c.config.SessionToken)
		loadOptions = append(loadOptions, config.WithCredentialsProvider(p))
	} else if !c.config.AllowHostAccess {
		return aws.Config{}, fmt.Errorf("static creds are not provided, and host access is not allowed")
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return aws.Config{}, err
	}

	if c.config.RoleARN != "" {
		stsClient := sts.NewFromConfig(awsConfig)
		assumeRoleOptions := []func(*stscreds.AssumeRoleOptions){}
		if c.config.RoleSessionName != "" {
			assumeRoleOptions = append(assumeRoleOptions, func(o *stscreds.AssumeRoleOptions) {
				o.RoleSessionName = c.config.RoleSessionName
			})
		}
		if c.config.ExternalID != "" {
			assumeRoleOptions = append(assumeRoleOptions, func(o *stscreds.AssumeRoleOptions) {
				o.ExternalID = &c.config.ExternalID
			})
		}
		provider := stscreds.NewAssumeRoleProvider(stsClient, c.config.RoleARN, assumeRoleOptions...)
		awsConfig.Credentials = aws.NewCredentialsCache(provider)
	}

	return awsConfig, nil
}

func (c *Connection) getClient(ctx context.Context) (*athena.Client, error) {
	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := athena.NewFromConfig(awsConfig, func(o *athena.Options) {
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	})
	return client, nil
}

func (c *Connection) executeQuery(ctx context.Context, client *athena.Client, sql, workgroup, outputLocation string) (*string, error) {
	executeParams := &athena.StartQueryExecutionInput{
		QueryString: aws.String(sql),
	}

	// this is not required be can be infer auto from workgroup if configure in it.
	if outputLocation != "" {
		executeParams.ResultConfiguration = &types2.ResultConfiguration{
			OutputLocation: aws.String(outputLocation),
		}
	}

	if workgroup != "" { // primary is used if nothing is set
		executeParams.WorkGroup = aws.String(workgroup)
	}

	queryExecutionOutput, err := client.StartQueryExecution(ctx, executeParams)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			ctx, cancel := graceful.WithMinimumDuration(ctx, 15*time.Second)
			_, stopErr := client.StopQueryExecution(ctx, &athena.StopQueryExecutionInput{
				QueryExecutionId: queryExecutionOutput.QueryExecutionId,
			})
			cancel()
			return nil, errors.Join(ctx.Err(), stopErr)
		default:
			status, err := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
				QueryExecutionId: queryExecutionOutput.QueryExecutionId,
			})
			if err != nil {
				return nil, err
			}

			switch status.QueryExecution.Status.State {
			case types2.QueryExecutionStateSucceeded:
				return queryExecutionOutput.QueryExecutionId, nil
			case types2.QueryExecutionStateCancelled:
				return nil, fmt.Errorf("Athena query execution cancelled")
			case types2.QueryExecutionStateFailed:
				return nil, fmt.Errorf("Athena query execution failed: %s", aws.ToString(status.QueryExecution.Status.AthenaError.ErrorMessage))
			}
		}
		time.Sleep(time.Second)
	}
}
