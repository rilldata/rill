package git

import (
	"context"
	"os"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	gogit "github.com/go-git/go-git/v5"
	"github.com/hashicorp/go-multierror"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("git", driver{})
}

type driver struct{}

func (d driver) Open(dsn string, logger *zap.Logger) (drivers.Connection, error) {
	r := retrier.New(retrier.ExponentialBackoff(3, 100*time.Millisecond), nil)

	var c *connection
	err := r.Run(func() error {
		tempdir, err := os.MkdirTemp("", "git_repo_driver")
		if err != nil {
			return err
		}

		c = &connection{root: dsn, tempdir: tempdir}

		_, err = gogit.PlainClone(tempdir, false, &gogit.CloneOptions{
			URL: dsn,
		})
		if err != nil {
			removeError := os.RemoveAll(tempdir)
			if removeError != nil {
				var combinedError error
				combinedError = multierror.Append(combinedError, err)
				combinedError = multierror.Append(combinedError, removeError)
				return combinedError
			}

			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

type connection struct {
	root    string
	tempdir string
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	err := os.RemoveAll(c.tempdir)
	if err != nil {
		return err
	}

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
