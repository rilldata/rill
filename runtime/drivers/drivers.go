package drivers

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

const _iteratorBatch = 8

var ErrIngestionLimitExceeded = fmt.Errorf("connectors: source ingestion exceeds limit")

type PermissionDeniedError struct {
	msg string
}

func NewPermissionDeniedError(msg string) error {
	return &PermissionDeniedError{msg: msg}
}

func (e *PermissionDeniedError) Error() string {
	return e.msg
}

// ErrNotFound indicates the resource wasn't found.
var ErrNotFound = errors.New("driver: not found")

// ErrDropNotSupported indicates the driver doesn't support dropping its underlying store.
var ErrDropNotSupported = errors.New("driver: drop not supported")

// Drivers is a registry of drivers.
var Drivers = make(map[string]Driver)

// Register registers a new driver.
func Register(name string, driver Driver) {
	if Drivers[name] != nil {
		panic(fmt.Errorf("already registered infra driver with name '%s'", name))
	}
	Drivers[name] = driver
}

// Open opens a new connection
func Open(driver string, config map[string]any, logger *zap.Logger) (Connection, error) {
	d, ok := Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown driver: %s", driver)
	}

	conn, err := d.Open(config, logger)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Drop tears down a store. Drivers that do not support it return ErrDropNotSupported.
func Drop(driver string, config map[string]any, logger *zap.Logger) error {
	d, ok := Drivers[driver]
	if !ok {
		return fmt.Errorf("unknown driver: %s", driver)
	}

	return d.Drop(config, logger)
}

// Driver represents an underlying DB.
type Driver interface {
	Spec() Spec

	// Open opens a new connection to an underlying store.
	Open(config map[string]any, logger *zap.Logger) (Connection, error)

	// Drop tears down a store. Drivers that do not support it return ErrDropNotSupported.
	Drop(config map[string]any, logger *zap.Logger) error

	// HasAnonymousSourceAccess returns true if external system can be accessed without credentials
	HasAnonymousSourceAccess(ctx context.Context, src Source, logger *zap.Logger) (bool, error)
}

// Connection represents a connection to an underlying DB.
// It should implement one or more of RegistryStore, CatalogStore, RepoStore, and OLAPStore.
type Connection interface {
	// Driver type (like "duckdb")
	Driver() string

	// Config used to open the Connection
	Config() map[string]any

	// Migrate prepares the connection for use. It will be called before the connection is first used.
	// (Not to be confused with migrating artifacts, which is handled by the runtime and tracked in the catalog.)
	Migrate(ctx context.Context) error

	// MigrationStatus returns the connection's current and desired migration version (if applicable)
	MigrationStatus(ctx context.Context) (current int, desired int, err error)

	// Close closes the connection
	Close() error

	// AsRegistry returns a AsRegistry if the driver can serve as such, otherwise returns false.
	// The registry is responsible for tracking instances and repos.
	AsRegistry() (RegistryStore, bool)

	// AsCatalogStore returns a AsCatalogStore if the driver can serve as such, otherwise returns false.
	// A catalog is used to store state about migrated/deployed objects (such as sources and metrics views).
	AsCatalogStore() (CatalogStore, bool)

	// AsRepoStore returns a AsRepoStore if the driver can serve as such, otherwise returns false.
	// A repo stores file artifacts (either in a folder or virtualized in a database).
	AsRepoStore() (RepoStore, bool)

	// AsOLAP returns an AsOLAP if the driver can serve as such, otherwise returns false.
	// OLAP stores are where we actually store, transform, and query users' data.
	AsOLAP() (OLAPStore, bool)

	// AsObjectStore returns an ObjectStore if the driver can serve as such, otherwise returns false.
	AsObjectStore() (ObjectStore, bool)

	// AsFileStore returns a Filetore if the driver can serve as such, otherwise returns false.
	AsFileStore() (FileStore, bool)

	// AsTransporter optionally returns an implementation for moving data between two connectors.
	// One of the input connections may be the Connection itself.
	// Examples:
	// a) myDuckDB.AsTransporter(myGCS, myDuckDB)
	// b) myBeam.AsTransporter(myGCS, myS3) // In the future
	AsTransporter(from Connection, to Connection) (Transporter, bool)
}
