package drivers

import (
	"context"
	"errors"
	"fmt"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
)

// ErrNotFound indicates the resource wasn't found.
var ErrNotFound = errors.New("driver: not found")

// ErrNotImplemented indicates the driver doesn't support the requested operation.
var ErrNotImplemented = errors.New("driver: not implemented")

// ErrStorageLimitExceeded indicates the driver's storage limit was exceeded.
var ErrStorageLimitExceeded = fmt.Errorf("connectors: exceeds storage limit")

// ErrNotNotifier indicates the driver cannot be used as a Notifier.
var ErrNotNotifier = errors.New("driver: not a notifier")

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
func Open(driver string, config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (Handle, error) {
	d, ok := Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown driver: %s", driver)
	}

	conn, err := d.Open(config, shared, client, logger)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Driver represents an external service that Rill can connect to.
type Driver interface {
	// Spec returns metadata about the driver, such as which configuration properties it supports.
	Spec() Spec

	// Open opens a new handle.
	Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (Handle, error)

	// HasAnonymousSourceAccess returns true if the driver can access the data identified by srcProps without any additional configuration.
	HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error)

	// TertiarySourceConnectors returns a list of drivers required to access the data identified by srcProps, excluding the driver itself.
	TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error)
}

// Handle represents a connection to an external service, such as a database, object store, etc.
// It should implement one or more of the As...() functions.
type Handle interface {
	// Driver name used to open the handle.
	Driver() string

	// Config used to open the handle.
	Config() map[string]any

	// Migrate prepares the handle for use. It will always be called before any of the As...() functions.
	Migrate(ctx context.Context) error

	// MigrationStatus returns the handle's current and desired migration version (if applicable).
	MigrationStatus(ctx context.Context) (current int, desired int, err error)

	// Close closes the handle.
	Close() error

	// AsRegistry returns a RegistryStore if the handle can serve as such, otherwise returns false.
	// A registry is responsible for tracking runtime metadata, namely instances and their configuration.
	AsRegistry() (RegistryStore, bool)

	// AsCatalogStore returns a CatalogStore if the handle can serve as such, otherwise returns false.
	// A catalog stores the state of an instance's resources (such as sources, models, metrics views, alerts, etc.)
	AsCatalogStore(instanceID string) (CatalogStore, bool)

	// AsRepoStore returns a RepoStore if the handle can serve as such, otherwise returns false.
	// A repo stores an instance's file artifacts (mostly YAML and SQL files).
	AsRepoStore(instanceID string) (RepoStore, bool)

	// AsAdmin returns an AdminService if the handle can serve as such, otherwise returns false.
	// An admin service enables the runtime to interact with the control plane that deployed it.
	AsAdmin(instanceID string) (AdminService, bool)

	// AsAI returns an AIService if the driver can serve as such, otherwise returns false.
	// An AI service enables an instance to request prompt-based text inference.
	AsAI(instanceID string) (AIService, bool)

	// AsSQLStore returns a SQLStore if the driver can serve as such, otherwise returns false.
	// A SQL store represents a service that can execute SQL statements and return the resulting rows.
	AsSQLStore() (SQLStore, bool)

	// AsOLAP returns an OLAPStore if the driver can serve as such, otherwise returns false.
	// An OLAP store is used to serve interactive, low-latency, analytical queries.
	// NOTE: We should consider merging the OLAPStore and SQLStore interfaces.
	AsOLAP(instanceID string) (OLAPStore, bool)

	// AsObjectStore returns an ObjectStore if the driver can serve as such, otherwise returns false.
	// An object store can store, list and retrieve files on a remote server.
	AsObjectStore() (ObjectStore, bool)

	// AsFileStore returns a FileStore if the driver can serve as such, otherwise returns false.
	// A file store can store, list and retrieve local files.
	// NOTE: The file store can probably be merged with the ObjectStore interface.
	AsFileStore() (FileStore, bool)

	// AsTransporter returns a Transporter for moving data between two other handles. One of the input handles may be the Handle itself.
	// Examples: duckdb.AsTransporter(gcs, duckdb), beam.AsTransporter(gcs, s3).
	AsTransporter(from Handle, to Handle) (Transporter, bool)

	// AsNotifier returns a Notifier (if the driver can serve as such) to send notifications: alerts, reports, etc.
	// Examples: email notifier, slack notifier.
	AsNotifier(properties map[string]any) (Notifier, error)
}

// PermissionDeniedError is returned when a driver cannot access some data due to insufficient permissions.
type PermissionDeniedError struct {
	msg string
}

func NewPermissionDeniedError(msg string) error {
	return &PermissionDeniedError{msg: msg}
}

func (e *PermissionDeniedError) Error() string {
	return e.msg
}
