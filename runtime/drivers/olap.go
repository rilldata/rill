package drivers

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	// QuerySchema returns the schema of the sql without trying not to run the actual query.
	QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error)
	// InformationSchema enables introspecting the tables and views available in the OLAP driver.
	InformationSchema() OLAPInformationSchema
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

// OLAPInformationSchema contains information about existing tables in an OLAP driver.
// Table lookups should be case insensitive.
type OLAPInformationSchema interface {
	// All returns metadata about all tables and views.
	// The like argument can optionally be passed to filter the tables by name.
	All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*OlapTable, string, error)
	// Lookup returns metadata about a specific tables and views.
	Lookup(ctx context.Context, db, schema, name string) (*OlapTable, error)
	// LoadPhysicalSize populates the PhysicalSizeBytes field of table metadata.
	// It should be called after All or Lookup and not on manually created tables.
	LoadPhysicalSize(ctx context.Context, tables []*OlapTable) error
	// LoadDDL populates the DDL field of a single table's metadata.
	// Drivers that don't support DDL retrieval should return nil (leaving DDL empty).
	LoadDDL(ctx context.Context, table *OlapTable) error
}

// OlapTable represents a table in an information schema.
type OlapTable struct {
	Database                string
	DatabaseSchema          string
	IsDefaultDatabase       bool
	IsDefaultDatabaseSchema bool
	Name                    string
	View                    bool
	// Schema is the table schema. It is only set when only single table is looked up. It is not set when listing all tables.
	Schema            *runtimev1.StructType
	UnsupportedCols   map[string]string
	PhysicalSizeBytes int64
	DDL               string
}

// DialectName constants identify SQL dialects by name.
// Use Dialect.String() == DialectNameDuckDB for comparisons.
const (
	DialectNameDuckDB     = "duckdb"
	DialectNameDruid      = "druid"
	DialectNameClickHouse = "clickhouse"
	DialectNamePinot      = "pinot"
	DialectNameStarRocks  = "starrocks"
	DialectNameBigQuery   = "bigquery"
	DialectNameSnowflake  = "snowflake"
	DialectNameAthena     = "athena"
	DialectNameRedshift   = "redshift"
	DialectNameMySQL      = "mysql"
	DialectNamePostgres   = "postgres"
)

// EscapeIdentifierDuckDB escapes an identifier using DuckDB/ANSI SQL double-quote syntax.
// This is a convenience helper for use in non-OLAP contexts that deal only with DuckDB.
func EscapeIdentifierDuckDB(ident string) string {
	if ident == "" {
		return ident
	}
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"` // nolint:gocritic
}

// ConvertToDateTruncSpecifierDuckDB converts a time grain to a DuckDB date_trunc specifier.
// This is a convenience helper for use in non-OLAP contexts that deal only with DuckDB.
func ConvertToDateTruncSpecifierDuckDB(grain runtimev1.TimeGrain) string {
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "MILLISECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "SECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "MINUTE"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "HOUR"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "DAY"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "WEEK"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "MONTH"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "QUARTER"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "YEAR"
	}
	return ""
}

// Dialect is the SQL dialect used by an OLAP driver.
type Dialect interface {
	String() string
	CanPivot() bool
	EscapeIdentifier(ident string) string
	EscapeAlias(alias string) string
	EscapeQualifiedIdentifier(name string) string
	EscapeStringValue(s string) string
	EscapeTable(db, schema, table string) string
	EscapeMember(tbl, name string) string
	EscapeMemberAlias(tbl, alias string) string
	ConvertToDateTruncSpecifier(grain runtimev1.TimeGrain) string
	SupportsILike() bool
	GetCastExprForLike() string
	SupportsRegexMatch() bool
	GetRegexMatchFunction() string
	RequiresArrayContainsForInOperator() bool
	GetArrayContainsFunction() string
	DimensionSelect(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_Dimension) (dimSelect, unnestClause string, err error)
	DimensionSelectPair(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_Dimension) (expr, alias, unnestClause string, err error)
	LateralUnnest(expr, tableAlias, colName string) (tbl string, tupleStyle, auto bool, err error)
	UnnestSQLSuffix(tbl string) string
	MetricsViewDimensionExpression(dimension *runtimev1.MetricsViewSpec_Dimension) (string, error)
	AnyValueExpression(expr string) string
	MinDimensionExpression(expr string) string
	MaxDimensionExpression(expr string) string
	GetTimeDimensionParameter() string
	CastToDataType(typ runtimev1.Type_Code) (string, error)
	SafeDivideExpression(numExpr, denExpr string) string
	OrderByExpression(name string, desc bool) string
	OrderByAliasExpression(name string, desc bool) string
	JoinOnExpression(lhs, rhs string) string
	DateTruncExpr(dim *runtimev1.MetricsViewSpec_Dimension, grain runtimev1.TimeGrain, tz string, firstDayOfWeek, firstMonthOfYear int) (string, error)
	DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error)
	IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error)
	SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location, firstDay, firstMonth int) (string, []any, error)
	SelectInlineResults(result *Result) (string, []any, []any, error)
	GetArgExpr(val any, typ runtimev1.Type_Code) (string, any, error)
	GetValExpr(val any, typ runtimev1.Type_Code) (bool, string, error)
	GetNullExpr(typ runtimev1.Type_Code) (bool, string)
	GetDateTimeExpr(t time.Time) (bool, string)
	GetDateExpr(t time.Time) (bool, string)
	LookupExpr(lookupTable, lookupValueColumn, lookupKeyExpr, lookupDefaultExpression string) (string, error)
	LookupSelectExpr(lookupTable, lookupKeyColumn string) (string, error)
	SanitizeQueryForLogging(sql string) string
}
