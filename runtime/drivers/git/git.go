package git

import (
	"context"
	"fmt"
	"os"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("git", driver{})
}

type driver struct{}

func (d driver) Open(dsn string, logger *zap.Logger) (drivers.Connection, error) {
	c := &connection{root: dsn}
	if err := c.checkRoot(); err != nil {
		err = os.RemoveAll("tempdir")
		if err != nil {
			return nil, err
		}

		split := strings.Split(dsn, "|")
		if len(split) > 1 {
			dsn = split[1]
		}

		_, err := gogit.PlainClone("tempdir", false, &gogit.CloneOptions{
			URL:      dsn,
			Progress: os.Stdout,
		})
		if err != nil {
			return nil, err
		}
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
	info, err := os.Stat("tempdir")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("repo: directory does not exist at '%s'", c.root)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("repo: file is not a directory '%s'", c.root)
	}

	_, err = gogit.PlainOpen("tempdir")
	if err != nil {
		return err
	}

	return nil
}
