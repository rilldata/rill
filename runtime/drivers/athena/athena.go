package athena

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

const defaultPageSize = 20

func init() {
	drivers.Register("athena", driver{})
	drivers.RegisterAsConnector("athena", driver{})
}

var spec = drivers.Spec{
	DisplayName:        "Amazon Athena",
	Description:        "Connect to Amazon Athena database.",
	ServiceAccountDocs: "",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from Athena.",
			Placeholder: "select * from catalog.table;",
		},
		{
			Key:         "output.location",
			DisplayName: "Output location",
			Description: "Oputut location for query results in S3.",
			Placeholder: "bucket-name",
			Type:        drivers.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "profile.name",
			DisplayName: "AWS profile",
			Description: "AWS profile for credentials.",
			Type:        drivers.StringPropertyType,
			Required:    true,
		},
	},
	ConfigProperties: []drivers.PropertySchema{},
}

type driver struct{}

type configProperties struct {
	// SecretJSON      string `mapstructure:"google_application_credentials"`
	// AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(config map[string]any, shared bool, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("athena driver can't be shared")
	}
	conf := &configProperties{}
	err := mapstructure.Decode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config: conf,
		logger: logger,
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src drivers.Source, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

type sourceProperties struct {
	SQL            string `mapstructure:"sql"`
	OutputLocation string `mapstructure:"output.location"`
	ProfileName    string `mapstructure:"profile.name"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

type Connection struct {
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Handle = &Connection{}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "athena"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, m)
	return m
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return nil
}

// Registry implements drivers.Connection.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *Connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
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

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
func (c *Connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// DownloadFiles returns a file iterator over objects stored in gcs.
// The credential json is read from config google_application_credentials.
// Additionally in case `allow_host_credentials` is true it looks for "Application Default Credentials" as well
func (c *Connection) DownloadFiles(ctx context.Context, source *drivers.BucketSource) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	prefix := "parquet_output_" + uuid.New().String()
	bucketName := strings.TrimPrefix(strings.TrimRight(conf.OutputLocation, "/"), "s3://")
	unloadPath := bucketName + "/" + prefix
	err = c.unload(ctx, conf, "s3://"+unloadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unload: %w", err)
	}

	bucketObj, err := c.openBucket(ctx, conf, bucketName)
	if err != nil {
		return nil, fmt.Errorf("cannot open bucket %q: %w", unloadPath, err)
	}

	opts := rillblob.Options{
		ExtractPolicy: &runtimev1.Source_ExtractPolicy{
			// FilesStrategy: runtimev1.Source_ExtractPolicy_STRATEGY_HEAD,
		},
		GlobPattern: prefix + "/**",
	}

	it, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		// TODO :: fix this for single file access. for single file first call only happens during download
		var failureErr awserr.RequestFailure
		if !errors.As(err, &failureErr) {
			return nil, fmt.Errorf("failed to create the iterator %q %w", unloadPath, err)
		}

		// check again
		if errors.As(err, &failureErr) && (failureErr.StatusCode() == http.StatusForbidden || failureErr.StatusCode() == http.StatusBadRequest) {
			return nil, drivers.NewPermissionDeniedError(fmt.Sprintf("can't access remote err: %v", failureErr))
		}
	}

	return it, err
}

func (c *Connection) openBucket(ctx context.Context, conf *sourceProperties, bucket string) (*blob.Bucket, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		// o.Region = conf.Region
		return nil
	}, config.WithSharedConfigProfile(conf.ProfileName))
	if err != nil {
		return nil, err
	}

	s3client := s3v2.NewFromConfig(cfg)
	return s3blob.OpenBucketV2(ctx, s3client, bucket, nil)
}

func (c *Connection) unload(ctx context.Context, conf *sourceProperties, path string) error {
	finalSQL := fmt.Sprintf("UNLOAD (%s) TO '%s' WITH (format = 'PARQUET')", conf.SQL, path)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(conf.ProfileName))
	if err != nil {
		return err
	}

	client := athena.NewFromConfig(cfg)

	resultConfig := &types.ResultConfiguration{
		OutputLocation: aws.String("s3://" + strings.TrimPrefix(strings.TrimRight(conf.OutputLocation, "/"), "s3://") + "/output/"),
	}

	executeParams := &athena.StartQueryExecutionInput{
		QueryString:         aws.String(finalSQL),
		ResultConfiguration: resultConfig,
	}

	// Start Query Execution
	athenaExecution, err := client.StartQueryExecution(ctx, executeParams)

	if err != nil {
		return err
	}

	// Get Query execution and check for the Query state constantly every 2 second
	executionID := *athenaExecution.QueryExecutionId

	r := retrier.New(retrier.LimitedExponentialBackoff(20, 100*time.Millisecond, 1*time.Second), nil) // 100 200 400 800 1000 1000 1000 1000 1000 1000 ... < 20 sec

	return r.Run(func() error {
		status, stateErr := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
			QueryExecutionId: &executionID,
		})

		if stateErr != nil {
			return stateErr
		}

		state := status.QueryExecution.Status.State

		if state == types.QueryExecutionStateSucceeded || state == types.QueryExecutionStateCancelled {
			return nil
		} else if state == types.QueryExecutionStateFailed {
			return fmt.Errorf("Athena query execution failed %s", *status.QueryExecution.Status.AthenaError.ErrorMessage)
		}
		return fmt.Errorf("Execution is not completed yet, current state: %s", state)
	})
}
