package file

import (
	"context"

	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.Register("file", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (drivers.Connection, error) {
	return &connection{}, nil
}

type connection struct{}

// Close implements drivers.Connection
func (c *connection) Close() error {
	return nil
}

// Registry implements drivers.Connection
func (c *connection) Registry() (drivers.Registry, bool) {
	return nil, false
}

// Catalog implements drivers.Connection
func (c *connection) Catalog() (drivers.Catalog, bool) {
	return nil, false
}

// Repo implements drivers.Connection
func (c *connection) Repo() (drivers.Repo, bool) {
	return nil, false
}

// OLAP implements drivers.Connection
func (c *connection) OLAP() (drivers.OLAP, bool) {
	return nil, false
}

// Migrate implements drivers.Connection
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection
func (c *connection) MigrationStatus(ctx context.Context) (current int, desired int, err error) {
	return 0, 0, nil
}
