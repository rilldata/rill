package redshift

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("redshift", driver{})
	drivers.RegisterAsConnector("redshift", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Amazon Redshift",
	Description: "Connect to Amazon Redshift database.",
	DocsURL:     "",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "aws_access_key_id",
			Secret: true,
		},
		{
			Key:    "aws_secret_access_key",
			Secret: true,
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from Redshift.",
			Placeholder: "select * from public.table;",
		},
		{
			Key:         "output_location",
			DisplayName: "S3 output location",
			Description: "Output location in S3 for temporary data.",
			Placeholder: "s3://bucket-name/path/",
			Type:        drivers.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "workgroup",
			DisplayName: "AWS Redshift workgroup",
			Description: "AWS Redshift workgroup",
			Placeholder: "default-workgroup",
			Type:        drivers.StringPropertyType,
			Required:    false,
		},
		{
			Key:         "region",
			DisplayName: "AWS region",
			Description: "AWS region",
			Placeholder: "us-east-1",
			Type:        drivers.StringPropertyType,
			Required:    false,
		},
		{
			Key:         "database",
			DisplayName: "Redshift database",
			Description: "Redshift database",
			Placeholder: "dev",
			Type:        drivers.StringPropertyType,
			Required:    true,
		},
		{
			Key:         "cluster_identifier",
			DisplayName: "Redshift cluster identifier",
			Description: "Redshift cluster identifier",
			Placeholder: "redshift-cluster-1",
			Type:        drivers.StringPropertyType,
			Required:    false,
		},
		{
			Key:         "role_arn",
			DisplayName: "Redshift role ARN",
			Description: "Redshift role ARN",
			Placeholder: "arn:aws:iam::03214372:role/service-role/AmazonRedshift-CommandsAccessRole-20240307T203902",
			Type:        drivers.StringPropertyType,
			Required:    true,
		},
	},
	ImplementsSQLStore: true,
}

type driver struct{}

type configProperties struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_access_token"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(config map[string]any, shared bool, _ *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("redshift driver can't be shared")
	}
	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
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

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type Connection struct {
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Handle = &Connection{}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "redshift"
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

// AsAI implements drivers.Handle.
func (c *Connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}
