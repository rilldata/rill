package redshift

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	redshift_types "github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	"github.com/aws/smithy-go/tracing/smithyoteltracing"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

func init() {
	drivers.Register("redshift", driver{})
	drivers.RegisterAsConnector("redshift", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Amazon Redshift",
	Description: "Connect to Amazon Redshift database.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/data-source/redshift",
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
			Key:         "workgroup",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS Redshift workgroup",
			Description: "AWS Redshift workgroup",
			Placeholder: "default-workgroup",
			Required:    false,
		},
		{
			Key:         "region",
			Type:        drivers.StringPropertyType,
			DisplayName: "AWS region",
			Description: "AWS region",
			Placeholder: "us-east-1",
			Required:    false,
		},
		{
			Key:         "database",
			Type:        drivers.StringPropertyType,
			DisplayName: "Redshift database",
			Description: "Redshift database",
			Placeholder: "dev",
			Required:    true,
		},
	},
	ImplementsWarehouse: true,
}

type driver struct{}

type configProperties struct {
	AccessKeyID       string `mapstructure:"aws_access_key_id"`
	SecretAccessKey   string `mapstructure:"aws_secret_access_key"`
	SessionToken      string `mapstructure:"aws_access_token"`
	AWSRegion         string `mapstructure:"region"`
	Database          string `mapstructure:"database"`
	Workgroup         string `mapstructure:"workgroup"`
	ClusterIdentifier string `mapstructure:"cluster_identifier"`
	AllowHostAccess   bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("redshift driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config:   conf,
		logger:   logger,
		storage:  st,
		clientMu: semaphore.NewWeighted(1),
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

	client    *redshiftdata.Client
	clientErr error
	clientMu  *semaphore.Weighted
}

var _ drivers.Handle = &Connection{}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	client, err := c.getClient(ctx)
	if err != nil {
		return err
	}

	_, err = c.executeQuery(ctx, client, "SELECT 1", c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier, nil)
	return err
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "redshift"
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

// AsOLAP implements drivers.Connection.
func (c *Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
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

// AsAI implements drivers.Handle.
func (c *Connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
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
	if c.config.AccessKeyID != "" || c.config.SecretAccessKey != "" {
		p := credentials.NewStaticCredentialsProvider(c.config.AccessKeyID, c.config.SecretAccessKey, c.config.SessionToken)
		loadOptions = append(loadOptions, config.WithCredentialsProvider(p))
	} else if !c.config.AllowHostAccess {
		return aws.Config{}, fmt.Errorf("static creds are not provided, and host access is not allowed")
	}

	return config.LoadDefaultConfig(ctx, loadOptions...)
}

func (c *Connection) getClient(ctx context.Context) (*redshiftdata.Client, error) {
	if err := c.clientMu.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer c.clientMu.Release(1)

	if c.client != nil || c.clientErr != nil {
		return c.client, c.clientErr
	}

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		c.clientErr = fmt.Errorf("failed to get AWS config: %w", err)
		return nil, c.clientErr
	}

	c.client = redshiftdata.NewFromConfig(awsConfig, func(o *redshiftdata.Options) {
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
	})
	return c.client, nil
}

// executeQuery executes a query with optional parameters and waits for it to complete
func (c *Connection) executeQuery(ctx context.Context, client *redshiftdata.Client, sql, database, workgroup, clusterIdentifier string, params []redshift_types.SqlParameter) (*redshiftdata.DescribeStatementOutput, error) {
	executeParams := &redshiftdata.ExecuteStatementInput{
		Sql:      aws.String(sql),
		Database: aws.String(database),
	}

	// Only set Parameters if there are any (Redshift Data API requires non-empty array)
	if len(params) > 0 {
		executeParams.Parameters = params
	}

	// Set either ClusterIdentifier or WorkgroupName, but not both
	// WorkgroupName is preferred for serverless
	if workgroup != "" {
		executeParams.WorkgroupName = aws.String(workgroup)
	} else if clusterIdentifier != "" {
		executeParams.ClusterIdentifier = aws.String(clusterIdentifier)
	} else {
		return nil, fmt.Errorf("either workgroup or cluster_identifier is required")
	}

	queryExecutionOutput, err := client.ExecuteStatement(ctx, executeParams)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			cancelCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			_, err = client.CancelStatement(cancelCtx, &redshiftdata.CancelStatementInput{
				Id: queryExecutionOutput.Id,
			})
			cancel()
			return nil, errors.Join(ctx.Err(), err)
		case <-ticker.C:
			status, err := client.DescribeStatement(ctx, &redshiftdata.DescribeStatementInput{
				Id: queryExecutionOutput.Id,
			})
			if err != nil {
				return nil, err
			}

			state := status.Status

			if status.Error != nil {
				return nil, fmt.Errorf("Redshift query execution failed %s", *status.Error)
			}

			if state != redshift_types.StatusStringSubmitted && state != redshift_types.StatusStringStarted && state != redshift_types.StatusStringPicked {
				return status, nil
			}
		}
	}
}
