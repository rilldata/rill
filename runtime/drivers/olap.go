package drivers

import (
	"context"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"

	// Load IANA time zone data
	_ "time/tzdata"
)

var (
	// ErrUnsupportedConnector is returned from Ingest for unsupported connectors.
	ErrUnsupportedConnector = errors.New("drivers: connector not supported")
	// ErrOptimizationFailure is returned when an optimization fails.
	ErrOptimizationFailure = errors.New("drivers: optimization failure")

	DefaultQuerySchemaTimeout = 30 * time.Second
)

// WithConnectionFunc is a callback function that provides a context to be used in further OLAP store calls to enforce affinity to a single connection.
// It also provides pointers to the actual database/sql and database/sql/driver connections.
// It's called with two contexts: wrappedCtx wraps the input context (including cancellation),
// and ensuredCtx wraps a background context (ensuring it can never be cancelled).
type WithConnectionFunc func(wrappedCtx context.Context, ensuredCtx context.Context) error

// OLAPStore is implemented by drivers that are capable of storing, transforming and serving analytical queries.
type OLAPStore interface {
	// Dialect is the SQL dialect that the driver uses.
	Dialect() Dialect
	// MayBeScaledToZero returns true if the driver might currently be scaled to zero.
	MayBeScaledToZero(ctx context.Context) bool
	// WithConnection acquires a connection from the pool and keeps it open until the callback returns.
	WithConnection(ctx context.Context, priority int, fn WithConnectionFunc) error
	// Exec executes a query against the OLAP driver.
	Exec(ctx context.Context, stmt *Statement) error
	// Query executes a query against the OLAP driver and returns an iterator for the resulting rows and schema.
	// The result MUST be closed after use.
	Query(ctx context.Context, stmt *Statement) (*Result, error)
	// Head executes a query with a limit of N and returns the resulting rows and schema.
	// It is separate from Query to allow drivers like BigQuery to optimize table previews and not incur huge costs of running a full query with limit.
	// The result MUST be closed after use.
	Head(ctx context.Context, db, schema, table string, limit int64) (*Result, error)
	// QuerySchema returns the schema of the sql without trying not to run the actual query.
	QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error)
	// InformationSchema enables introspecting the tables and views available in the OLAP driver.
	InformationSchema() InformationSchema
	// EstimateSize returns an estimate of the total data size in bytes.
	// Returns -1 if size estimation is not supported by the driver.
	EstimateSize(ctx context.Context) (int64, error)
}

// Statement wraps a query to execute against an OLAP driver.
type Statement struct {
	// Query is the SQL query to execute.
	Query string
	// Args are positional arguments to bind to the query.
	Args []any
	// DryRun indicates if the query should be parsed and validated, but not actually executed.
	DryRun bool
	// Priority provides a query priority if the driver supports it (a higher value indicates a higher priority).
	Priority int
	// UseCache explicitly enables/disables reading from database-level caches (if supported by the driver).
	// If not set, the driver will use its default behavior.
	UseCache *bool
	// PopulateCache explicitly enables/disables writing to database-level caches (if supported by the driver).
	// If not set, the driver will use its default behavior.
	PopulateCache *bool
	// ExecutionTimeout provides a timeout for query execution.
	// Unlike a timeout on ctx, it will be enforced only for query execution, not for time spent waiting in queues.
	// It may not be supported by all drivers.
	ExecutionTimeout time.Duration
	// QueryAttributes provides additional attributes for the query (if supported by the driver).
	// These can be used to customize the behavior of the query "{{ .user.partnerId }}"
	QueryAttributes map[string]string
}

// Rows is an iterator for rows returned by a query. It mimics the behavior of sqlx.Rows.
type Rows interface {
	Next() bool
	Err() error
	Close() error
	Scan(dest ...any) error
	MapScan(dest map[string]any) error
}

// Result is the result of a query. It wraps a Rows iterator with additional functionality.
type Result struct {
	Rows
	Schema    *runtimev1.StructType
	cleanupFn func() error
	cap       int64
	rows      int64
}

// SetCleanupFunc sets a function, which will be called when the Result is closed.
func (r *Result) SetCleanupFunc(fn func() error) {
	if r.cleanupFn == nil {
		r.cleanupFn = fn
		return
	}

	prevFn := r.cleanupFn
	r.cleanupFn = func() error {
		err1 := prevFn()
		err2 := fn()
		return errors.Join(err1, err2)
	}
}

// SetCap caps the number of rows to return. If the number is exceeded, an error is returned.
func (r *Result) SetCap(n int64) {
	if r.cap > 0 {
		panic("cap already set")
	}
	r.cap = n
}

// Next wraps rows.Next and enforces the cap set by SetCap.
func (r *Result) Next() bool {
	res := r.Rows.Next()
	if !res {
		return false
	}

	r.rows++
	if r.cap > 0 && r.rows > r.cap {
		return false
	}

	return true
}

// Err returns the error of the underlying rows.
func (r *Result) Err() error {
	err := r.Rows.Err()
	if err != nil {
		return err
	}

	if r.cap > 0 && r.rows > r.cap {
		return fmt.Errorf("result cap exceeded: returned more than %d rows", r.cap)
	}

	return nil
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
