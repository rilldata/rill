package postgres

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"

	// Load postgres driver
	_ "github.com/jackc/pgx/v5/stdlib"
)

func init() {
	drivers.Register("postgres", driver{})
	drivers.RegisterAsConnector("postgres", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Postgres",
	Description: "Connect to Postgres.",
	DocsURL:     "https://docs.rilldata.com/reference/connectors/postgres",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:    "database_url",
			Secret: true,
		},
	},
	// Important: Any edits to the below properties must be accompanied by changes to the client-side form validation schemas.
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "sql",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "SQL",
			Description: "Query to extract data from Postgres.",
			Placeholder: "select * from table;",
		},
		{
			Key:         "database_url",
			Type:        drivers.StringPropertyType,
			DisplayName: "Postgres Connection String",
			Required:    false,
			DocsURL:     "https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING",
			Placeholder: "postgresql://postgres:postgres@localhost:5432/postgres",
			Hint:        "Can be configured here or by setting the 'connector.postgres.database_url' environment variable (using '.env' or '--env')",
			Secret:      true,
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
	ImplementsSQLStore: true,
}

type driver struct{}

type ConfigProperties struct {
	DatabaseURL string `mapstructure:"database_url"`
	DSN         string `mapstructure:"dsn"`
}

func (c *ConfigProperties) ResolveDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	return c.DatabaseURL
}

func (d driver) Open(instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("postgres driver can't be shared")
	}

	return &connection{
		config: config,
	}, nil
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
	config map[string]any
}

// Ping implements drivers.Handle.
func (c *connection) Ping(ctx context.Context) error {
	// Open DB handle
	db, err := c.getDB("")
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()
	return db.PingContext(ctx)
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "postgres"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return maps.Clone(c.config)
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return nil
}

// AsRegistry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
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

// AsOLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Connection.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (c *connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, bool) {
	return nil, false
}

// AsModelManager implements drivers.Handle.
func (c *connection) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (c *connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// getDB opens a new sqlx.DB connection to the specified database; if empty, connects to the default database from the DSN.
func (c *connection) getDB(database string) (*sqlx.DB, error) {
	conf := &ConfigProperties{}
	var err error
	if err = mapstructure.WeakDecode(c.config, conf); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	dsn := conf.ResolveDSN()
	if dsn == "" {
		return nil, fmt.Errorf("database_url or dsn not provided")
	}
	if database != "" {
		dsn, err = updateDatabaseInDSN(dsn, database)
	}

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}
	return db, nil
}

// updateDSNDatabase sets or replaces the database name in a PostgreSQL DSN.
func updateDatabaseInDSN(dsn, database string) (string, error) {
	// Handle URL-style DSN
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "", err
		}
		u.Path = "/" + database
		return u.String(), nil
	}

	// Handle DSN as key=value pairs (e.g., user=foo password=bar dbname=mydb)
	parts := strings.Fields(dsn)
	found := false
	for i, part := range parts {
		if strings.HasPrefix(part, "dbname=") {
			parts[i] = "dbname=" + database
			found = true
			break
		}
	}
	if !found {
		parts = append(parts, "dbname="+database)
	}
	return strings.Join(parts, " "), nil
}
