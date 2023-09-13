package athena

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	s3v2types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

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
			Key:         "athena_output_location",
			DisplayName: "S3 output location",
			Description: "Oputut location for query results in S3.",
			Placeholder: "s3://bucket-name/path/",
			Type:        drivers.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "athena_workgroup",
			DisplayName: "AWS Athena workgroup",
			Description: "AWS Athena workgroup to use for queries.",
			Type:        drivers.StringPropertyType,
			Required:    false,
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:    "aws_access_key_id",
			Secret: true,
		},
		{
			Key:    "aws_secret_access_key",
			Secret: true,
		},
	},
}

type driver struct{}

type configProperties struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_access_token"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(config map[string]any, shared bool, client activity.Client, logger *zap.Logger) (drivers.Handle, error) {
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

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

type sourceProperties struct {
	SQL            string `mapstructure:"sql"`
	OutputLocation string `mapstructure:"athena_output_location"`
	WorkGroup      string `mapstructure:"athena_workgroup"`
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
	return nil, false
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
	return c, true
}

func cleanPath(ctx context.Context, cfg aws.Config, bucketName, prefix string) error {
	s3client := s3v2.NewFromConfig(cfg)
	out, err := s3client.ListObjectsV2(ctx, &s3v2.ListObjectsV2Input{
		Bucket: &bucketName,
		Prefix: &prefix,
	})
	if err != nil {
		return err
	}

	ids := make([]s3v2types.ObjectIdentifier, 0, len(out.Contents))
	for _, o := range out.Contents {
		ids = append(ids, s3v2types.ObjectIdentifier{
			Key: o.Key,
		})
	}
	_, err = s3client.DeleteObjects(ctx, &s3v2.DeleteObjectsInput{
		Delete: &s3v2types.Delete{
			Objects: ids,
		},
	})
	return err
}

type janitorIterator struct {
	drivers.FileIterator
	ctx        context.Context
	cfg        aws.Config
	bucketName string
	unloadPath string
}

func (ci janitorIterator) Close() error {
	err := ci.FileIterator.Close()
	if err != nil {
		return err
	}

	return cleanPath(ci.ctx, ci.cfg, ci.bucketName, ci.unloadPath)
}

func (c *Connection) newCfg(ctx context.Context) (aws.Config, error) {
	var cfg aws.Config
	var err error
	if c.config.AllowHostAccess {
		cfg, err = awsconfig.LoadDefaultConfig(
			ctx,
		)
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(
			ctx,
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.config.AccessKeyID, c.config.SecretAccessKey, c.config.SessionToken)),
		)
	}
	if err != nil {
		return aws.Config{}, err
	}

	return cfg, nil
}

func (c *Connection) openBucket(ctx context.Context, conf *sourceProperties, bucket string) (*blob.Bucket, error) {
	cfg, err := c.newCfg(ctx)
	if err != nil {
		return nil, err
	}

	s3client := s3v2.NewFromConfig(cfg)
	return s3blob.OpenBucketV2(ctx, s3client, bucket, nil)
}

func resolveOutputLocation(ctx context.Context, client *athena.Client, conf *sourceProperties) (string, error) {
	if conf.OutputLocation != "" {
		return conf.OutputLocation, nil
	} else {
		wo, err := client.GetWorkGroup(ctx, &athena.GetWorkGroupInput{
			WorkGroup: aws.String(conf.WorkGroup),
		})
		if err != nil {
			return "", err
		}
		return *wo.WorkGroup.Configuration.ResultConfiguration.OutputLocation, nil
	}
}

func (c *Connection) unload(client *athena.Client, ctx context.Context, cfg aws.Config, conf *sourceProperties, unloadLocation string) error {
	finalSQL := fmt.Sprintf("UNLOAD (%s) TO '%s' WITH (format = 'PARQUET')", conf.SQL, unloadLocation)

	resultConfig := &types.ResultConfiguration{
		OutputLocation: aws.String(conf.OutputLocation),
	}

	executeParams := &athena.StartQueryExecutionInput{
		QueryString:         aws.String(finalSQL),
		ResultConfiguration: resultConfig,
	}

	if conf.WorkGroup != "" {
		executeParams = &athena.StartQueryExecutionInput{
			QueryString: aws.String(finalSQL),
			WorkGroup:   aws.String(conf.WorkGroup),
		}
	}

	athenaExecution, err := client.StartQueryExecution(ctx, executeParams)
	if err != nil {
		return err
	}

	r := retrier.New(retrier.ConstantBackoff(int(5*time.Minute/time.Second), time.Second), nil) // 5 minutes timeout
	return r.RunCtx(ctx, func(ctx context.Context) error {
		status, err := client.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
			QueryExecutionId: athenaExecution.QueryExecutionId,
		})
		if err != nil {
			return err
		}

		state := status.QueryExecution.Status.State

		if state == types.QueryExecutionStateSucceeded || state == types.QueryExecutionStateCancelled {
			return nil
		} else if state == types.QueryExecutionStateFailed {
			return fmt.Errorf("Athena query execution failed %s", *status.QueryExecution.Status.AthenaError.ErrorMessage)
		}
		return fmt.Errorf("Athena ingestion timeout")
	})
}
