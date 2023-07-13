package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"

	// Load sqlite driver
	_ "modernc.org/sqlite"
)

func init() {
	drivers.Register("sqlite", driver{})
}

type driver struct{}

func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open sqlite connection")
	}

	// The sqlite driver requires the DSN to contain "_time_format=sqlite" to support TIMESTAMP types in all timezones.
	if !strings.Contains(dsn, "_time_format") {
		if strings.Contains(dsn, "?") {
			dsn += "&_time_format=sqlite"
		} else {
			dsn += "?_time_format=sqlite"
		}
	}

	// Open DB handle
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &connection{
		db:     db,
		config: config,
	}, nil
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

var _ drivers.Connection = &connection{}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "sqlite"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return c.config
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

// Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return c, true
}

// Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore() (drivers.CatalogStore, bool) {
	return c, true
}

// Repo implements drivers.Connection.
func (c *connection) AsRepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP() (drivers.OLAPStore, bool) {
	return nil, false
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
