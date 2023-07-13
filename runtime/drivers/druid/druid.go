package druid

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"

	// Load calcite avatica driver for druid
	_ "github.com/apache/calcite-avatica-go/v5"
)

func init() {
	drivers.Register("druid", driver{})
}

type driver struct{}

// Open connects to Druid using Avatica.
// Note that the Druid connection string must have the form "http://host/druid/v2/sql/avatica-protobuf/".
func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open druid connection")
	}

	db, err := sqlx.Open("avatica", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(40)

	conn := &connection{
		db:     db,
		config: config,
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return drivers.Spec{}
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src drivers.Source, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

type connection struct {
	db     *sqlx.DB
	config map[string]any
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "druid"
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
func (c *connection) AsCatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) AsRepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP() (drivers.OLAPStore, bool) {
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
func (c *connection) AsTransporter(from, to drivers.Connection) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

func (c *connection) EstimateSize() (int64, bool) {
	return 0, false
}
