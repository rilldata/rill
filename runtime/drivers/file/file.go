package file

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/rilldata/rill/runtime/drivers"
)

func init() {
	drivers.Register("file", driver{})
}

type driver struct{}

func (d driver) Open(dsn string, logger *zap.Logger) (drivers.Connection, error) {
	c := &connection{root: dsn}
	if err := c.checkRoot(); err != nil {
		return nil, err
	}
	return c, nil
}

type connection struct {
	root string
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return nil
}

// Registry implements drivers.Connection.
func (c *connection) RegistryStore() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) CatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) RepoStore() (drivers.RepoStore, bool) {
	return c, true
}

// OLAP implements drivers.Connection.
func (c *connection) OLAPStore() (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// checkPath checks that the connection's root is a valid directory.
func (c *connection) checkRoot() error {
	info, err := os.Stat(c.root)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("repo: directory does not exist at '%s'", c.root)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("repo: file is not a directory '%s'", c.root)
	}

	return nil
}
