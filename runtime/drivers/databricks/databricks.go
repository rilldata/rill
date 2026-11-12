package databricks

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	dbsqllog "github.com/databricks/databricks-sql-go/logger"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	// Load Databricks SQL driver
	_ "github.com/databricks/databricks-sql-go"
)

func init() {
	drivers.Register("databricks", driver{})
	drivers.RegisterAsConnector("databricks", driver{})

	_ = dbsqllog.SetLogLevel("disabled")
}

var spec = drivers.Spec{
	DisplayName: "Databricks",
	Description: "Connect to Databricks.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connectors/data-source/databricks",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			DisplayName: "Databricks Connection String",
			Placeholder: "token:<token>@<host>:443/<http_path>?catalog=<catalog>&schema=<schema>",
			Hint:        "Can be configured here or by setting the 'connector.databricks.dsn' environment variable (using '.env' or '--env').",
			Secret:      true,
		},
		{
			Key:         "host",
			Type:        drivers.StringPropertyType,
			DisplayName: "Host",
			Required:    true,
			Placeholder: "dbc-xxxxxxxx-xxxx.cloud.databricks.com",
			Hint:        "Databricks SQL warehouse hostname",
		},
		{
			Key:         "http_path",
			Type:        drivers.StringPropertyType,
			DisplayName: "HTTP Path",
			Required:    true,
			Placeholder: "/sql/1.0/warehouses/xxxxxxxxxxxxxxxx",
			Hint:        "HTTP path for the SQL warehouse",
		},
		{
			Key:         "token",
			Type:        drivers.StringPropertyType,
			DisplayName: "Access Token",
			Required:    true,
			Placeholder: "dapi...",
			Hint:        "Databricks personal access token",
			Secret:      true,
		},
		{
			Key:         "auth",
			Type:        drivers.StringPropertyType,
			DisplayName: "Authentication",
			Hint:        "Set to 'oauth' to run dashboard queries as each viewer with their own Databricks OAuth token (Rill Cloud only)",
		},
		{
			Key:         "oauth_client_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "OAuth Client ID",
			Hint:        "Client ID of the Databricks OAuth app connection for 'auth: oauth' (Rill Cloud only)",
		},
		{
			Key:         "oauth_client_secret",
			Type:        drivers.StringPropertyType,
			DisplayName: "OAuth Client Secret",
			Hint:        "Client secret of the Databricks OAuth app connection for 'auth: oauth' (Rill Cloud only)",
			Secret:      true,
		},
		{
			Key:         "oauth_scopes",
			Type:        drivers.StringPropertyType,
			DisplayName: "OAuth Scopes",
			Hint:        "List of OAuth scopes for 'auth: oauth' (Rill Cloud only; defaults to 'sql' and 'offline_access')",
		},
		{
			Key:         "catalog",
			Type:        drivers.StringPropertyType,
			DisplayName: "Catalog",
			Placeholder: "main",
			Hint:        "Unity Catalog name (optional; defaults to the workspace default)",
		},
		{
			Key:         "schema",
			Type:        drivers.StringPropertyType,
			DisplayName: "Schema",
			Placeholder: "default",
			Hint:        "Schema within the catalog (optional; defaults to the workspace default)",
		},
	},
	ImplementsOLAP:      true,
	ImplementsWarehouse: true,
}

type driver struct{}

type configProperties struct {
	DSN      string `mapstructure:"dsn"`
	Host     string `mapstructure:"host"`
	HTTPPath string `mapstructure:"http_path"`
	Token    string `mapstructure:"token"`
	// Auth is "oauth" to run queries on behalf of users with their own Databricks OAuth tokens (Rill Cloud only).
	// The token above remains the service credential for work with no user, such as reconciling resources.
	Auth              string   `mapstructure:"auth"`
	OAuthClientID     string   `mapstructure:"oauth_client_id"`
	OAuthClientSecret string   `mapstructure:"oauth_client_secret"`
	OAuthScopes       []string `mapstructure:"oauth_scopes"`
	Catalog           string   `mapstructure:"catalog"`
	Schema            string   `mapstructure:"schema"`
	// CatalogWhitelist is a comma separated list of catalogs to list in information schema calls.
	CatalogWhitelist string `mapstructure:"catalog_whitelist"`
	LogQueries       bool   `mapstructure:"log_queries"`
}

func (c *configProperties) validate() error {
	var set []string
	if c.Host != "" {
		set = append(set, "host")
	}
	if c.HTTPPath != "" {
		set = append(set, "http_path")
	}
	if c.Token != "" {
		set = append(set, "token")
	}
	if c.Catalog != "" {
		set = append(set, "catalog")
	}
	if c.Schema != "" {
		set = append(set, "schema")
	}
	if c.DSN != "" && len(set) > 0 {
		return fmt.Errorf("databricks: only one of 'dsn' or [%s] can be set", strings.Join(set, ", "))
	}
	if c.DSN == "" && (c.Host == "" || c.HTTPPath == "" || c.Token == "") {
		return errors.New("databricks: either 'dsn' or 'host', 'http_path', and 'token' are required")
	}
	if c.Auth != "" && c.Auth != authOAuth {
		return fmt.Errorf("databricks: invalid 'auth': expected %q or empty, got %q", authOAuth, c.Auth)
	}
	if c.Auth == authOAuth && (c.DSN != "" || c.OAuthClientID == "" || c.OAuthClientSecret == "") {
		return errors.New("databricks: 'auth: oauth' requires 'host', 'oauth_client_id' and 'oauth_client_secret', and can't be used with 'dsn'")
	}
	// Otherwise queries would silently use the service token for every viewer.
	if c.Auth != authOAuth && (c.OAuthClientID != "" || c.OAuthClientSecret != "" || len(c.OAuthScopes) > 0) {
		return errors.New("databricks: 'oauth_client_id', 'oauth_client_secret' and 'oauth_scopes' require 'auth: oauth'")
	}
	return nil
}

func (c *configProperties) resolveDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	params := url.Values{}
	params.Set("timezone", "UTC")
	if c.Catalog != "" {
		params.Set("catalog", c.Catalog)
	}
	if c.Schema != "" {
		params.Set("schema", c.Schema)
	}
	// DSN format: https://token:<token>@<host>:443/<http_path>?catalog=<catalog>&schema=<schema>
	u := &url.URL{
		Scheme:   "https",
		User:     url.UserPassword("token", c.Token),
		Host:     c.Host + ":443",
		Path:     c.HTTPPath,
		RawQuery: params.Encode(),
	}
	return u.String()
}

func (d driver) Open(_ context.Context, connectorName, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("databricks driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	if err := conf.validate(); err != nil {
		return nil, err
	}

	c := &connection{
		name:       connectorName,
		instanceID: instanceID,
		config:     conf,
		storage:    st,
		logger:     logger,
		dbMu:       semaphore.NewWeighted(1),
	}
	if conf.Auth == authOAuth {
		c.oauthConfig = newOAuthConfig(conf)
		c.userDBs = make(map[string]*userDB)
	}
	return c, nil
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

type connection struct {
	name       string
	instanceID string
	config     *configProperties
	storage    *storage.Client
	logger     *zap.Logger

	// db authenticates with the service credential. It is lazily populated using getServiceDB.
	db    *sqlx.DB
	dbErr error
	dbMu  *semaphore.Weighted

	// oauthConfig is set if the connector uses per-user OAuth.
	oauthConfig *drivers.OAuthConfig
	// userDBs are the per-user connection pools, keyed by Rill user ID.
	// A Databricks session belongs to the identity that opened it, so pools are never shared across users.
	userDBs   map[string]*userDB
	userDBsMu sync.Mutex
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	err := c.withRetry(ctx, func(ctx context.Context, db *sqlx.DB) error {
		return db.PingContext(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to open databricks connection: %w", err)
	}
	return nil
}

// Migrate implements drivers.Handle.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Driver implements drivers.Handle.
func (c *connection) Driver() string {
	return "databricks"
}

// Config implements drivers.Handle.
func (c *connection) Config() map[string]any {
	var m map[string]any
	_ = mapstructure.WeakDecode(c.config, &m)
	return m
}

// Close implements drivers.Handle.
func (c *connection) Close() error {
	c.userDBsMu.Lock()
	for id, u := range c.userDBs {
		_ = u.db.Close()
		delete(c.userDBs, id)
	}
	c.userDBsMu.Unlock()
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// AsRegistry implements drivers.Handle.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Handle.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
}

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
}

// AsObjectStore implements drivers.Handle.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsFileStore implements drivers.Handle.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return c, true
}

// AsMCP implements drivers.Handle.
func (c *connection) AsMCP() (drivers.MCPService, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// getServiceDB returns the connection pool that authenticates with the service credential.
func (c *connection) getServiceDB(ctx context.Context) (*sqlx.DB, error) {
	err := c.dbMu.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.dbMu.Release(1)
	if c.db != nil || c.dbErr != nil {
		return c.db, c.dbErr
	}

	// The driver falls back to SEA when a Lakehouse//RT warehouse rejects Thrift.
	c.db, c.dbErr = sqlx.Open("databricks", c.config.resolveDSN())
	return c.db, c.dbErr
}
