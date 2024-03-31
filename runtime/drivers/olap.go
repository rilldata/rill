package drivers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ErrUnsupportedConnector is returned from Ingest for unsupported connectors.
var ErrUnsupportedConnector = errors.New("drivers: connector not supported")

// WithConnectionFunc is a callback function that provides a context to be used in further OLAP store calls to enforce affinity to a single connection.
// It also provides pointers to the actual database/sql and database/sql/driver connections.
// It's called with two contexts: wrappedCtx wraps the input context (including cancellation),
// and ensuredCtx wraps a background context (ensuring it can never be cancelled).
type WithConnectionFunc func(wrappedCtx context.Context, ensuredCtx context.Context, conn *sql.Conn) error

// OLAPStore is implemented by drivers that are capable of storing, transforming and serving analytical queries.
// NOTE crud APIs are not safe to be called with `WithConnection`
type OLAPStore interface {
	Dialect() Dialect
	WithConnection(ctx context.Context, priority int, longRunning, tx bool, fn WithConnectionFunc) error
	Exec(ctx context.Context, stmt *Statement) error
	Execute(ctx context.Context, stmt *Statement) (*Result, error)
	InformationSchema() InformationSchema
	EstimateSize() (int64, bool)

	CreateTableAsSelect(ctx context.Context, name string, view bool, sql string) error
	InsertTableAsSelect(ctx context.Context, name string, byName bool, sql string) error
	DropTable(ctx context.Context, name string, view bool) error
	// RenameTable is force rename
	RenameTable(ctx context.Context, name, newName string, view bool) error
	AddTableColumn(ctx context.Context, tableName, columnName string, typ string) error
	AlterTableColumn(ctx context.Context, tableName, columnName string, newType string) error
}

// Statement wraps a query to execute against an OLAP driver.
type Statement struct {
	Query            string
	Args             []any
	DryRun           bool
	Priority         int
	LongRunning      bool
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
// Table lookups should be case insensitive.
type InformationSchema interface {
	All(ctx context.Context) ([]*Table, error)
	Lookup(ctx context.Context, db, schema, name string) (*Table, error)
}

// Table represents a table in an information schema.
type Table struct {
	Database                string
	DatabaseSchema          string
	IsDefaultDatabase       bool
	IsDefaultDatabaseSchema bool
	Name                    string
	View                    bool
	Schema                  *runtimev1.StructType
	UnsupportedCols         map[string]string
}

// IngestionSummary is details about ingestion
type IngestionSummary struct {
	BytesIngested int64
}

// Dialect enumerates OLAP query languages.
type Dialect int

const (
	DialectUnspecified Dialect = iota
	DialectDuckDB
	DialectDruid
	DialectClickHouse
)

func (d Dialect) String() string {
	switch d {
	case DialectUnspecified:
		return ""
	case DialectDuckDB:
		return "duckdb"
	case DialectDruid:
		return "druid"
	case DialectClickHouse:
		return "clickhouse"
	default:
		panic("not implemented")
	}
}

// EscapeIdentifier returns an escaped SQL identifier in the dialect.
func (d Dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(ident, "\"", "\"\""))
}

// EscapeTable returns an esacped fully qualified table name
func (d Dialect) EscapeTable(db, schema, table string) string {
	var sb strings.Builder
	if db != "" {
		sb.WriteString(d.EscapeIdentifier(db))
		sb.WriteString(".")
	}
	if schema != "" {
		sb.WriteString(d.EscapeIdentifier(schema))
		sb.WriteString(".")
	}
	sb.WriteString(d.EscapeIdentifier(table))
	return sb.String()
}
