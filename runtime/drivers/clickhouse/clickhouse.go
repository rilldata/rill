package clickhouse

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
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
}

type driver struct{}

var maxOpenConnections = 20

// Open connects to Clickhouse using std API.
// Connection string format : https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#dsn
func (d driver) Open(config map[string]any, shared bool, client activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("clickhouse driver can't be shared")
	}
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open clickhouse connection")
	}

	db, err := sqlx.Open("clickhouse", dsn)
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

	conn := &connection{
		db:      db,
		config:  config,
		logger:  logger,
		metaSem: semaphore.NewWeighted(1),
		olapSem: priorityqueue.NewSemaphore(maxOpenConnections - 1),
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return drivers.Spec{}
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type connection struct {
	db       *sqlx.DB
	config   map[string]any
	logger   *zap.Logger
	activity activity.Client

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
	return c.config
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

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
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
		if from.Driver() == "s3" {
			return NewS3Transporter(from, olap, c.logger), true
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

func (c *connection) EstimateSize() (int64, bool) {
	return 0, false
}

func (c *connection) AcquireLongRunning(ctx context.Context) (func(), error) {
	return nil, fmt.Errorf("not implemented")
}
