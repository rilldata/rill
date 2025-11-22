package bigquery

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/option"
)

func init() {
	drivers.Register("bigquery", driver{})
	drivers.RegisterAsConnector("bigquery", driver{})
}

// spec for bigquery connector
var spec = drivers.Spec{
	DisplayName: "BigQuery",
	Description: "Import data from BigQuery.",
	DocsURL:     "https://docs.rilldata.com/build/connectors/data-source/bigquery",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "project_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "Project ID",
			Description: "Google project ID.",
			Placeholder: "my-project",
			Hint:        "Rill will use the project ID from your local credentials, unless set here. Set this if no project ID configured in credentials.",
		},
		{
			Key:         "google_application_credentials",
			Type:        drivers.FilePropertyType,
			DisplayName: "GCP Credentials",
			Description: "GCP credentials as JSON string",
			Placeholder: "Paste your GCP service account JSON here",
			Secret:      true,
			Required:    true,
		},
	},
	ImplementsWarehouse: true,
}

type driver struct{}

type configProperties struct {
	SecretJSON      string `mapstructure:"google_application_credentials"`
	ProjectID       string `mapstructure:"project_id"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.
	LogQueries bool `mapstructure:"log_queries"`
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("bigquery driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config:   conf,
		storage:  st,
		logger:   logger,
		clientMu: semaphore.NewWeighted(1),
	}
	return conn, nil
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
	config  *configProperties
	storage *storage.Client
	logger  *zap.Logger

	client    *bigquery.Client // lazily populated using getClient
	clientErr error
	clientMu  *semaphore.Weighted
}

var _ drivers.Handle = &Connection{}

// Ping implements drivers.Handle.
func (c *Connection) Ping(ctx context.Context) error {
	client, err := c.getClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Run a simple query to verify connection
	q := client.Query("SELECT 1")
	_, err = q.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute test query: %w", err)
	}

	return nil
}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "bigquery"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, &m)
	return m
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
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
	if opts.InputHandle == c {
		store, ok := opts.OutputHandle.AsObjectStore()
		if ok && opts.OutputHandle.Driver() == "gcs" {
			return &selfToGCSExecutor{
				c:     c,
				store: store,
			}, nil
		}
	}
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

// AsNotifier implements drivers.Connection.
func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// getClient initializes and caches a BigQuery client
func (c *Connection) getClient(ctx context.Context) (*bigquery.Client, error) {
	err := c.clientMu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.clientMu.Release(1)

	if c.client != nil || c.clientErr != nil {
		return c.client, c.clientErr
	}
	client, err := c.createClient(ctx, "")
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			// don't cache context errors
			return nil, err
		}
		c.clientErr = err
		return nil, err
	}
	c.client = client
	return c.client, nil
}

// createClient initializes a BigQuery client using the provided context and project ID.
// If no project ID is given, it attempts to use the one from the config or auto-detect it.
func (c *Connection) createClient(ctx context.Context, projectID string) (*bigquery.Client, error) {
	opts, err := c.clientOption(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Google API client options: %w", err)
	}
	if projectID == "" {
		if c.config.ProjectID != "" {
			projectID = c.config.ProjectID
		} else {
			projectID = bigquery.DetectProjectID
		}
	}
	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		if strings.Contains(err.Error(), "unable to detect projectID") {
			return nil, fmt.Errorf("projectID not detected in credentials. Please set `project_id` in the connector YAML")
		}
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}
	return client, nil
}

func (c *Connection) clientOption(ctx context.Context) ([]option.ClientOption, error) {
	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/drive.readonly",
	}
	creds, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess, scopes...)
	if err != nil {
		return nil, err
	}
	return []option.ClientOption{option.WithCredentials(creds)}, nil
}

type sourceProperties struct {
	ProjectID string `mapstructure:"project_id"`
	SQL       string `mapstructure:"sql"`
}

func (c *Connection) parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory for connector \"bigquery\"")
	}
	if conf.ProjectID == "" {
		if c.config.ProjectID != "" {
			conf.ProjectID = c.config.ProjectID
		} else {
			conf.ProjectID = bigquery.DetectProjectID
		}
	}
	return conf, err
}
