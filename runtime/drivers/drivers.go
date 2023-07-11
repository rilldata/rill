package drivers

import (
	"context"
	"errors"
	"fmt"
	"math"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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
	// Open opens a new connection to an underlying store.
	Open(config map[string]any, logger *zap.Logger) (Connection, error)

	// Drop tears down a store. Drivers that do not support it return ErrDropNotSupported.
	Drop(config map[string]any, logger *zap.Logger) error
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

	// AsConnector returns a Connector if a driver can server as such, otherwise returns false.
	AsConnector() (Connector, bool)
}

type ObjectStore interface {
	// DownloadFiles provides an iterator for downloading and consuming files
	DownloadFiles(ctx context.Context, src *BucketSource) (FileIterator, error)
}

type FileStore interface {
	// FilePaths returns local paths where files are stored
	FilePaths(ctx context.Context, src *FilesSource) ([]string, error)
}

// FileIterator provides ways to iteratively downloade files from external sources
// Clients should call close once they are done with iterator to release any resources
type FileIterator interface {
	// Close do cleanup and release resources
	Close() error
	// NextBatch returns a list of file downloaded from external sources
	// NextBatch cleanups file created in previous batch
	NextBatch(limit int) ([]string, error)
	// HasNext can be utlisied to check if iterator has more elements left
	HasNext() bool
	// Returns size in unit. The numbers may not be 100% accurate
	// returns 0,false if not able to compute size in given unit
	Size(unit ProgressUnit) (int64, bool)
}

// Transporter implements logic for moving data between two connectors
// (the actual connector objects are provided in AsTransporter)
type Transporter interface {
	Transfer(ctx context.Context, source Source, sink Sink, t *TransferOpts, p Progress) error
}

// A Source is expected to only return ok=true for one of the source types.
// The caller will know which type based on the connector type.
type Source interface {
	BucketSource() (*BucketSource, bool)
	DatabaseSource() (*DatabaseSource, bool)
	FilesSource() (*FilesSource, bool)
}

// A Sink is expected to only return ok=true for one of the sink types.
// The caller will know which type based on the connector type.
type Sink interface {
	BucketSink() (*BucketSink, bool)
	DatabaseSink() (*DatabaseSink, bool)
}

type BucketSource struct {
	Paths         []string // May be globs
	ExtractPolicy *runtimev1.Source_ExtractPolicy
	Properties    map[string]any // TODO :: this should also be part of connection open
}

var _ Source = &BucketSource{}

func (b *BucketSource) BucketSource() (*BucketSource, bool) {
	return b, true
}

func (b *BucketSource) DatabaseSource() (*DatabaseSource, bool) {
	return nil, false
}

func (b *BucketSource) FilesSource() (*FilesSource, bool) {
	return nil, false
}

type BucketSink struct {
	Path string
	// Format FileFormat
	// NOTE: In future, may add file name and output partitioning config here
}

var _ Sink = &BucketSink{}

func (b *BucketSink) BucketSink() (*BucketSink, bool) {
	return b, true
}

func (b *BucketSink) DatabaseSink() (*DatabaseSink, bool) {
	return nil, false
}

type DatabaseSource struct {
	// Pass only Query OR Table
	Query    string
	Table    string
	Database string
	Limit    int
}

var _ Source = &DatabaseSource{}

func (d *DatabaseSource) BucketSource() (*BucketSource, bool) {
	return nil, false
}

func (d *DatabaseSource) DatabaseSource() (*DatabaseSource, bool) {
	return d, true
}

func (d *DatabaseSource) FilesSource() (*FilesSource, bool) {
	return nil, false
}

type DatabaseSink struct {
	Table  string
	Append bool
}

var _ Sink = &DatabaseSink{}

func (d *DatabaseSink) BucketSink() (*BucketSink, bool) {
	return nil, false
}

func (d *DatabaseSink) DatabaseSink() (*DatabaseSink, bool) {
	return d, true
}

type FilesSource struct {
	Name       string
	Properties map[string]any
}

var _ Source = &FilesSource{}

func (f *FilesSource) BucketSource() (*BucketSource, bool) {
	return nil, false
}

func (f *FilesSource) DatabaseSource() (*DatabaseSource, bool) {
	return nil, false
}

func (f *FilesSource) FilesSource() (*FilesSource, bool) {
	return f, true
}

type TransferOpts struct {
	IteratorBatch int
	LimitInBytes  int64
}

func NewTransferOpts(opts ...TransferOption) *TransferOpts {
	t := &TransferOpts{
		IteratorBatch: _iteratorBatch,
		LimitInBytes:  math.MaxInt64,
	}

	for _, opt := range opts {
		opt(t)
	}
	return t
}

type TransferOption func(*TransferOpts)

func WithIteratorBatch(b int) TransferOption {
	return func(t *TransferOpts) {
		t.IteratorBatch = b
	}
}

func WithLimitInBytes(limit int64) TransferOption {
	return func(t *TransferOpts) {
		t.LimitInBytes = limit
	}
}

// Progress is an interface for communicating progress info
type Progress interface {
	Target(val int64, unit ProgressUnit)
	Observe(val int64, unit ProgressUnit)
}

type NoOpProgress struct{}

func (n NoOpProgress) Target(val int64, unit ProgressUnit)  {}
func (n NoOpProgress) Observe(val int64, unit ProgressUnit) {}

var _ Progress = NoOpProgress{}

type ProgressUnit int

const (
	ProgressUnitByte ProgressUnit = iota
	ProgressUnitFile
	ProgressUnitRecord
)
