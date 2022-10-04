package drivers

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ErrClosed indicates the connection is closed
var ErrClosed = errors.New("driver: connection is closed")

// ErrNotFound indicates the resource wasn't found
var ErrNotFound = errors.New("driver: not found")

// Drivers is a registry of drivers
var Drivers = make(map[string]Driver)

// Register registers a new driver
func Register(name string, driver Driver) {
	if Drivers[name] != nil {
		panic(fmt.Errorf("already registered infra driver with name '%s'", name))
	}
	Drivers[name] = driver
}

// Open opens a new connection
func Open(driver string, dsn string) (Connection, error) {
	d, ok := Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}

	conn, err := d.Open(dsn)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Driver represents an underlying DB
type Driver interface {
	Open(dsn string) (Connection, error)
}

// Connection represents a connection to an underlying DB.
// It should implement one or more of Registry, Catalog, Repo and OLAP.
type Connection interface {
	// Close closes the connection
	Close() error

	// Registry returns a Registry if the driver can serve as such, otherwise returns false
	Registry() (Registry, bool)

	// Catalog returns a Catalog if the driver can serve as such, otherwise returns false
	Catalog() (Catalog, bool)

	// Repo returns a Repo if the driver can serve as such, otherwise returns false
	Repo() (Repo, bool)

	// OLAP returns an OLAP if the driver can serve as such, otherwise returns false
	OLAP() (OLAP, bool)

	// Migrate prepares the connection for use. It will be called before the connection is first used.
	Migrate(ctx context.Context) error

	// MigrationStatus returns the connection's current and desired migration version
	MigrationStatus(ctx context.Context) (current int, desired int, err error)
}

// Registry is implemented by drivers capable of storing and looking up instances and repos
type Registry interface {
}

// Catalog is implemented by drivers capable of storing catalog info for a specific instance
type Catalog interface {
}

// Repo is implemented by drivers capable of storing SQL file artifacts
type Repo interface {
}

// OLAP is implemented by drivers that are capable of storing, transforming and serving analytical queries
type OLAP interface {
	Execute(ctx context.Context, stmt *Statement) (*sqlx.Rows, error)
	InformationSchema() InformationSchema
}

// Statement wraps a query to execute against an OLAP driver
type Statement struct {
	Query    string
	Args     []any
	DryRun   bool
	Priority int
}

// InformationSchema contains information about existing tables in an OLAP driver
type InformationSchema interface {
	All(ctx context.Context) ([]*Table, error)
	Lookup(ctx context.Context, name string) (*Table, error)
}

// Table represents a table in an information schema
type Table struct {
	Database string
	Schema   string
	Name     string
	Type     string
	Columns  []Column
}

// Column represents a column in a table
type Column struct {
	Name     string
	Type     string
	Nullable bool
}
