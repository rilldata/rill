package drivers

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
)

// ErrUnsupportedConnector is returned from Ingest for unsupported connectors.
var ErrUnsupportedConnector = errors.New("drivers: connector not supported")

// WithConnectionFunc is a callback function that provides a context to be used in further OLAP store calls to enforce affinity to a single connection.
// It's called with two contexts: wrappedCtx wraps the input context (including cancellation),
// and ensuredCtx wraps a background context (ensuring it can never be cancelled).
type WithConnectionFunc func(wrappedCtx context.Context, ensuredCtx context.Context) error

// OLAPStore is implemented by drivers that are capable of storing, transforming and serving analytical queries.
type OLAPStore interface {
	Dialect() Dialect
	WithConnection(ctx context.Context, priority int, fn WithConnectionFunc) error
	Exec(ctx context.Context, stmt *Statement) error
	Execute(ctx context.Context, stmt *Statement) (*Result, error)
	Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) (*IngestionSummary, error)
	InformationSchema() InformationSchema
}

// Statement wraps a query to execute against an OLAP driver.
type Statement struct {
	Query            string
	Args             []any
	DryRun           bool
	Priority         int
	ExecutionTimeout time.Duration
}

// Result wraps the results of query.
type Result struct {
	*sqlx.Rows
	Schema    *runtimev1.StructType
	cleanupFn func() error
}

// SetCleanupFunc sets a function, which will be called when the Result is closed.
func (r *Result) SetCleanupFunc(fn func() error) {
	if r.cleanupFn != nil {
		panic("cleanup function already set")
	}
	r.cleanupFn = fn
}

// Close wraps rows.Close and calls the Result's cleanup function (if it is set).
// Close should be idempotent.
func (r *Result) Close() error {
	firstErr := r.Rows.Close()
	if r.cleanupFn != nil {
		err := r.cleanupFn()
		if firstErr == nil {
			firstErr = err
		}

		// Prevent cleanupFn from being called multiple times.
		// NOTE: Not idempotent for error returned from cleanupFn.
		r.cleanupFn = nil
	}
	return firstErr
}

// InformationSchema contains information about existing tables in an OLAP driver.
type InformationSchema interface {
	All(ctx context.Context) ([]*Table, error)
	Lookup(ctx context.Context, name string) (*Table, error)
}

// Table represents a table in an information schema.
type Table struct {
	Database       string
	DatabaseSchema string
	Name           string
	Schema         *runtimev1.StructType
}

// Dialect enumerates OLAP query languages.
type Dialect int

const (
	DialectUnspecified Dialect = iota
	DialectDuckDB
	DialectDruid
)

func (d Dialect) String() string {
	switch d {
	case DialectUnspecified:
		return ""
	case DialectDuckDB:
		return "duckdb"
	case DialectDruid:
		return "druid"
	default:
		panic("not implemented")
	}
}

// IngestionSummary is details about ingestion
type IngestionSummary struct {
	BytesIngested int64
}
