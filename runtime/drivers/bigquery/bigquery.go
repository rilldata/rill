package bigquery

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

func init() {
	drivers.Register("bigquery", driver{})
	drivers.RegisterAsConnector("bigquery", driver{})
}

// spec for bigquery connector
var spec = drivers.Spec{
	DisplayName:        "BigQuery",
	Description:        "Import data from BigQuery.",
	ServiceAccountDocs: "https://docs.rilldata.com/reference/connectors/bigquery",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from BigQuery.",
			Placeholder: "select * from project.dataset.table;",
		},
		{
			Key:         "project_id",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Project ID",
			Description: "Google project ID.",
			Placeholder: "my-project",
			Hint:        "Rill will use the project ID from your local credentials, unless set here. Set this if no project ID configured in credentials.",
		},
		{
			Key:         "google_application_credentials",
			DisplayName: "GCP credentials",
			Description: "GCP credentials inferred from your local environment.",
			Type:        drivers.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>gcloud auth application-default login</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/reference/connectors/gcs#local-credentials",
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:  "google_application_credentials",
			Hint: "Enter path of file to load from.",
			ValidateFunc: func(any interface{}) error {
				val := any.(string)
				if val == "" {
					// user can chhose to leave empty for public sources
					return nil
				}

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
				if err != nil {
					return err
				}

				_, err = os.Stat(path)
				return err
			},
			TransformFunc: func(any interface{}) interface{} {
				val := any.(string)
				if val == "" {
					return ""
				}

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
				if err != nil {
					return err
				}
				// ignoring error since PathError is already validated
				content, _ := os.ReadFile(path)
				return string(content)
			},
		},
	},
}

type driver struct{}

type configProperties struct {
	SecretJSON      string `mapstructure:"google_application_credentials"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("bigquery driver can't be shared")
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
	// gcp provides public access to the data via a project
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

var _ drivers.SQLStore = &Connection{}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "gcs"
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

// AsAdmin implements drivers.Handle.
func (c *Connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *Connection) AsAI(instanceID string) (drivers.AIService, bool) {
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

// AsSQLStore implements drivers.Connection.
func (c *Connection) AsSQLStore() (drivers.SQLStore, bool) {
	return c, true
}

// AsTransporter implements drivers.Connection.
func (c *Connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *Connection) AsNotifier() (drivers.Notifier, bool) {
	return nil, false
}

type sourceProperties struct {
	ProjectID string `mapstructure:"project_id"`
	SQL       string `mapstructure:"sql"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory for connector \"bigquery\"")
	}
	if conf.ProjectID == "" {
		conf.ProjectID = bigquery.DetectProjectID
	}
	return conf, err
}

func (c *Connection) clientOption(ctx context.Context) ([]option.ClientOption, error) {
	creds, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess)
	if err != nil {
		return nil, err
	}
	return []option.ClientOption{option.WithCredentials(creds)}, nil
}
