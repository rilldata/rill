package clickhouse

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	// import clickhouse driver
	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func init() {
	drivers.Register("clickhouse", driver{})
	drivers.RegisterAsConnector("clickhouse", driver{})
}

var spec = drivers.Spec{
	DisplayName: "ClickHouse",
	Description: "Connect to ClickHouse.",
	DocsURL:     "https://docs.rilldata.com/reference/olap-engines/clickhouse",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Connection string",
			Placeholder: "clickhouse://localhost:9000?username=default&password=",
			Secret:      true,
		},
	},
	// This spec is intentionally missing a source schema, as the frontend provides
	// custom instructions for how to connect Clickhouse as the OLAP driver.
	SourceProperties: nil,
	ImplementsOLAP:   true,
}

var maxOpenConnections = 20

type driver struct{}

type configProperties struct {
	// DSN is the connection string
	DSN string `mapstructure:"dsn"`
	// EnableCache controls whether to enable cache for Clickhouse queries.
	EnableCache bool `mapstructure:"enable_cache"`
	// LogQueries controls whether to log the raw SQL passed to OLAP.Execute.
	LogQueries bool `mapstructure:"log_queries"`
}

// Open connects to Clickhouse using std API.
// Connection string format : https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#dsn
func (d driver) Open(instanceID string, config map[string]any, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("clickhouse driver can't be shared")
	}

	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	if conf.DSN == "" {
		return nil, fmt.Errorf("no DSN provided to open the connection")
	}

	db, err := sqlx.Open("clickhouse", conf.DSN)
	if err != nil {
		return nil, err
	}

	// very roughly approximating num queries required for a typical page load
	// TODO: copied from druid reevaluate
	db.SetMaxOpenConns(maxOpenConnections)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("connection: %w", err)
	}

	// group by positional args are supported post 22.7 and we use them heavily in our queries
	row := db.QueryRow(`
	WITH
    	splitByChar('.', version()) AS parts,
    	toInt32(parts[1]) AS major,
    	toInt32(parts[2]) AS minor
	SELECT (major > 22) OR ((major = 22) AND (minor >= 7)) AS is_supported
`)
	var isSupported bool
	if err := row.Scan(&isSupported); err != nil {
		return nil, err
	}
	if !isSupported {
		return nil, fmt.Errorf("clickhouse version must be 22.7 or higher")
	}

	conn := &connection{
		db:      db,
		config:  conf,
		logger:  logger,
		metaSem: semaphore.NewWeighted(1),
		olapSem: priorityqueue.NewSemaphore(maxOpenConnections - 1),
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type connection struct {
	db         *sqlx.DB
	config     *configProperties
	logger     *zap.Logger
	activity   *activity.Client
	instanceID string

	// logic around this copied from duckDB driver
	// This driver may issue both OLAP and "meta" queries (like catalog info) against DuckDB.
	// Meta queries are usually fast, but OLAP queries may take a long time. To enable predictable parallel performance,
	// we gate queries with semaphores that limits the number of concurrent queries of each type.
	// The metaSem allows 1 query at a time and the olapSem allows cfg.PoolSize-1 queries at a time.
	// When cfg.PoolSize is 1, we set olapSem to still allow 1 query at a time.
	// This creates contention for the same connection in database/sql's pool, but its locks will handle that.
	metaSem *semaphore.Weighted
	olapSem *priorityqueue.Semaphore
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "clickhouse"
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, m)
	return m
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

// Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
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

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	c.instanceID = instanceID
	return c, true
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	olap, _ := to.(*connection)
	if c == to {
		switch from.Driver() {
		case "s3":
			return NewS3Transporter(from, olap, c.logger), true
		case "https":
			return NewHTTPTransporter(from, olap, c.logger), true
		}
	}
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
// Use OLAPStore instead.
func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// AsNotifier implements drivers.Connection.
func (c *connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

func (c *connection) EstimateSize() (int64, bool) {
	return 0, false
}

func (c *connection) AcquireLongRunning(ctx context.Context) (func(), error) {
	return nil, fmt.Errorf("not implemented")
}
