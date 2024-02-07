package iceberg

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
	"os"
)

func init() {
	drivers.Register("iceberg", driver{})
	drivers.RegisterAsConnector("iceberg", driver{})
}

var spec = drivers.Spec{
	DisplayName:        "Iceberg + S3",
	Description:        "Read from an Iceberg table in AWS S3",
	ServiceAccountDocs: "",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "warehouse",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Iceberg Warehouse",
			Description: "The location of the iceberg warehouse - Must be S3",
			Placeholder: "<warehouse uri>",
		},
		{
			Key:         "database",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Iceberg Database",
			Description: "The database to read from",
			Placeholder: "<database name>",
		},
		{
			Key:         "table",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Iceberg Table Name",
			Description: "The iceberg table to read from",
			Placeholder: "<table name>",
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
		{
			Key:    "aws_region",
			Secret: false,
		},
	},
}

type driver struct{}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) Open(config map[string]any, shared bool, client activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("iceberg driver can't be shared")
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

type Connection struct {
	config *configProperties
	logger *zap.Logger
}

type configProperties struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_access_token"`
	AwsRegion       string `mapstructure:"aws_region"`
}

func (c Connection) Driver() string {
	return "iceberg"
}

func (c Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	err := mapstructure.Decode(c.config, m)
	if err != nil {
		panic(err)
	}
	return m
}

func (c Connection) Close() error {
	return nil
}

func (c Connection) awsConfig() *aws.Config {

	cfg := aws.NewConfig()
	// TODO why aren't these being passed with the connection config?
	//cfg.Region = c.config.AwsRegion
	//cfg.Credentials = credentials.NewStaticCredentialsProvider(
	//	c.config.AccessKeyID, c.config.SecretAccessKey, c.config.SessionToken,
	//)
	cfg.Region = "us-east-1"
	accessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	cfg.Credentials = credentials.NewStaticCredentialsProvider(
		accessKeyId, secretAccessKey, sessionToken,
	)
	return cfg
}
