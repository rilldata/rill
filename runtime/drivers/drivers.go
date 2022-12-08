package drivers

import (
	"context"
	"errors"
	"fmt"
)

// ErrClosed indicates the connection is closed.
var ErrClosed = errors.New("driver: connection is closed")

// ErrNotFound indicates the resource wasn't found.
var ErrNotFound = errors.New("driver: not found")

// Drivers is a registry of drivers.
var Drivers = make(map[string]Driver)

// Register registers a new driver.
func Register(name string, driver Driver) {
	if Drivers[name] != nil {
		panic(fmt.Errorf("already registered infra driver with name '%s'", name))
	}
	Drivers[name] = driver
}

// Open opens a new connection.
func Open(driver string, dsn string, poolSize int) (Connection, error) {
	d, ok := Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}

	conn, err := d.Open(dsn, poolSize)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Driver represents an underlying DB.
type Driver interface {
	Open(dsn string, poolSize int) (Connection, error)
}

// Connection represents a connection to an underlying DB.
// It should implement one or more of RegistryStore, CatalogStore, RepoStore, and OLAPStore.
type Connection interface {
	// Migrate prepares the connection for use. It will be called before the connection is first used.
	// (Not to be confused with migrating artifacts, which is handled by the runtime and tracked in the catalog.)
	Migrate(ctx context.Context) error

	// MigrationStatus returns the connection's current and desired migration version (if applicable)
	MigrationStatus(ctx context.Context) (current int, desired int, err error)

	// Close closes the connection
	Close() error

	// RegistryStore returns a RegistryStore if the driver can serve as such, otherwise returns false.
	// The registry is responsible for tracking instances and repos.
	RegistryStore() (RegistryStore, bool)

	// CatalogStore returns a CatalogStore if the driver can serve as such, otherwise returns false.
	// A catalog is used to store state about migrated/deployed objects (such as sources and metrics views).
	CatalogStore() (CatalogStore, bool)

	// RepoStore returns a RepoStore if the driver can serve as such, otherwise returns false.
	// A repo stores file artifacts (either in a folder or virtualized in a database).
	RepoStore() (RepoStore, bool)

	// OLAPStore returns an OLAPStore if the driver can serve as such, otherwise returns false.
	// OLAP stores are where we actually store, transform, and query users' data.
	OLAPStore() (OLAPStore, bool)
}
