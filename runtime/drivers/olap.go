package drivers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
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
	DialectNameAthena     = "athena"
	DialectNameBigQuery   = "bigquery"
	DialectNameClickHouse = "clickhouse"
	DialectNameDuckDB     = "duckdb"
	DialectNameDruid      = "druid"
	DialectNameMySQL      = "mysql"
	DialectNamePinot      = "pinot"
	DialectNamePostgres   = "postgres"
	DialectNameRedshift   = "redshift"
	DialectNameSnowflake  = "snowflake"
	DialectNameStarRocks  = "starrocks"
)

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
	GetRegexMatchFunction() (string, error)
	RequiresArrayContainsForInOperator() bool
	GetArrayContainsFunction() (string, error)
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
	ColumnCardinality(db, dbSchema, table, column string) (string, error)
	ColumnDescriptiveStatistics(db, dbSchema, table, column string) (string, error)
	IsNonNullFinite(floatColumn string) string
	ColumnNullCount(db, dbSchema, table, column string) (string, error)
	ColumnNumericHistogramBucket(db, dbSchema, table, column string) (string, error)
}

// BaseDialect provides default implementations for the Dialect interface.
// Embed it in a concrete dialect struct and call InitBase to wire up virtual dispatch.
type BaseDialect struct {
	self Dialect
}

// InitBase wires up virtual dispatch. Must be called in the concrete dialect's constructor.
func (b *BaseDialect) InitBase(self Dialect) {
	b.self = self
}

func (b *BaseDialect) CanPivot() bool {
	return false
}

func (b *BaseDialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// Most other dialects follow ANSI SQL: use double quotes.
	// Replace any internal double quotes with escaped double quotes.
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(ident, `"`, `""`)) // nolint:gocritic
}

func (b *BaseDialect) EscapeAlias(alias string) string {
	return b.self.EscapeIdentifier(alias)
}

// EscapeQualifiedIdentifier escapes a dot-separated qualified name (e.g. "schema.table") by escaping each part individually.
// Use this instead of EscapeIdentifier when the input may contain dots that represent schema/table separators.
// WARNING: Only use it for edge features where it is an acceptable trade-off to NOT support tables with a dot in their name (which we occasionally see in real-world use cases).
func (b *BaseDialect) EscapeQualifiedIdentifier(name string) string {
	if name == "" {
		return name
	}
	parts := strings.Split(name, ".")
	for i, part := range parts {
		parts[i] = b.self.EscapeIdentifier(part)
	}
	return strings.Join(parts, ".")
}

func (b *BaseDialect) EscapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

func (b *BaseDialect) ConvertToDateTruncSpecifier(grain runtimev1.TimeGrain) string {
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

func (b *BaseDialect) SupportsILike() bool { return true }

// GetCastExprForLike returns the cast expression for use in a LIKE or ILIKE condition, or an empty string if no cast is necessary.
func (b *BaseDialect) GetCastExprForLike() string {
	return ""
}

func (b *BaseDialect) SupportsRegexMatch() bool {
	return false
}

func (b *BaseDialect) GetRegexMatchFunction() (string, error) {
	return "", fmt.Errorf("regex match not supported for %s dialect", b.self.String())
}

// EscapeTable returns an escaped table name with database, schema and table.
func (b *BaseDialect) EscapeTable(db, schema, table string) string {
	var sb strings.Builder
	if db != "" {
		sb.WriteString(b.self.EscapeIdentifier(db))
		sb.WriteString(".")
	}
	if schema != "" {
		sb.WriteString(b.self.EscapeIdentifier(schema))
		sb.WriteString(".")
	}
	sb.WriteString(b.self.EscapeIdentifier(table))
	return sb.String()
}

// EscapeMember returns an escaped member name with table alias and column name.
func (b *BaseDialect) EscapeMember(tbl, name string) string {
	if tbl == "" {
		return b.self.EscapeIdentifier(name)
	}
	return fmt.Sprintf("%s.%s", b.self.EscapeIdentifier(tbl), b.self.EscapeIdentifier(name))
}

// EscapeMemberAlias is like EscapeMember but uses EscapeAlias for the column name.
func (b *BaseDialect) EscapeMemberAlias(tbl, alias string) string {
	if tbl == "" {
		return b.self.EscapeAlias(alias)
	}
	return fmt.Sprintf("%s.%s", b.self.EscapeIdentifier(tbl), b.self.EscapeAlias(alias))
}

func (b *BaseDialect) DimensionSelect(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_Dimension) (dimSelect, unnestClause string, err error) {
	colName := b.self.EscapeIdentifier(dim.Name)
	alias := b.self.EscapeAlias(dim.Name)
	if !dim.Unnest {
		expr, err := b.self.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return fmt.Sprintf(`(%s) AS %s`, expr, alias), "", nil
	}

	unnestColName := b.self.EscapeIdentifier(TempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := TempName("tbl")
	sel := fmt.Sprintf(`%s AS %s`, unnestColName, alias)
	if dim.Expression == "" {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl_name("unnested_colName") ...
		return sel, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, b.self.EscapeTable(db, dbSchema, table), colName, unnestTableName, unnestColName), nil
	}
	return sel, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName), nil
}

func (b *BaseDialect) DimensionSelectPair(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_Dimension) (expr, alias, unnestClause string, err error) {
	colAlias := b.self.EscapeAlias(dim.Name)
	if !dim.Unnest {
		ex, err := b.self.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return ex, colAlias, "", nil
	}

	unnestColName := b.self.EscapeIdentifier(TempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := TempName("tbl")
	if dim.Expression == "" {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl_name("unnested_colName") ...
		return unnestColName, colAlias, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, b.self.EscapeTable(db, dbSchema, table), colAlias, unnestTableName, unnestColName), nil
	}
	return unnestColName, colAlias, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName), nil
}

func (b *BaseDialect) LateralUnnest(expr, tableAlias, colName string) (tbl string, tupleStyle, auto bool, err error) {
	return fmt.Sprintf(`LATERAL UNNEST(%s) %s(%s)`, expr, tableAlias, b.self.EscapeIdentifier(colName)), true, false, nil
}

func (b *BaseDialect) UnnestSQLSuffix(tbl string) string {
	return fmt.Sprintf(", %s", tbl)
}

func (b *BaseDialect) RequiresArrayContainsForInOperator() bool {
	return false
}

func (b *BaseDialect) GetArrayContainsFunction() (string, error) {
	return "", fmt.Errorf("array contains not supported for %s dialect", b.self.String())
}

func (b *BaseDialect) MetricsViewDimensionExpression(dimension *runtimev1.MetricsViewSpec_Dimension) (string, error) {
	if dimension.LookupTable != "" {
		var keyExpr string
		if dimension.Column != "" {
			keyExpr = b.self.EscapeIdentifier(dimension.Column)
		} else if dimension.Expression != "" {
			keyExpr = dimension.Expression
		} else {
			return "", fmt.Errorf("dimension %q has a lookup table but no column or expression defined", dimension.Name)
		}
		return b.self.LookupExpr(dimension.LookupTable, dimension.LookupValueColumn, keyExpr, dimension.LookupDefaultExpression)
	}
	if dimension.Expression != "" {
		return dimension.Expression, nil
	}
	if dimension.Column != "" {
		return b.self.EscapeIdentifier(dimension.Column), nil
	}
	// Backwards compatibility for older projects that have not run reconcile on this metrics view.
	// In that case `column` will not be present.
	return b.self.EscapeIdentifier(dimension.Name), nil
}

// AnyValueExpression applies the ANY_VALUE aggregation function (or equivalent) to the given expression.
func (b *BaseDialect) AnyValueExpression(expr string) string {
	return fmt.Sprintf("ANY_VALUE(%s)", expr)
}

func (b *BaseDialect) MinDimensionExpression(expr string) string {
	return fmt.Sprintf("MIN(%s)", expr)
}

func (b *BaseDialect) MaxDimensionExpression(expr string) string {
	return fmt.Sprintf("MAX(%s)", expr)
}

func (b *BaseDialect) GetTimeDimensionParameter() string {
	return "?"
}

func (b *BaseDialect) CastToDataType(typ runtimev1.Type_Code) (string, error) {
	switch typ {
	case runtimev1.Type_CODE_TIMESTAMP:
		return "TIMESTAMP", nil
	default:
		return "", fmt.Errorf("unsupported cast type %q for %s dialect", b.self.String(), typ.String())
	}
}

func (b *BaseDialect) SafeDivideExpression(numExpr, denExpr string) string {
	return fmt.Sprintf("(%s)/CAST(%s AS DOUBLE)", numExpr, denExpr)
}

func (b *BaseDialect) OrderByExpression(name string, desc bool) string {
	res := b.self.EscapeIdentifier(name)
	if desc {
		res += " DESC"
	}
	return res
}

func (b *BaseDialect) OrderByAliasExpression(name string, desc bool) string {
	res := b.self.EscapeAlias(name)
	if desc {
		res += " DESC"
	}
	return res
}

func (b *BaseDialect) JoinOnExpression(lhs, rhs string) string {
	return fmt.Sprintf("%s IS NOT DISTINCT FROM %s", lhs, rhs)
}

func (b *BaseDialect) DateTruncExpr(_ *runtimev1.MetricsViewSpec_Dimension, _ runtimev1.TimeGrain, _ string, _, _ int) (string, error) {
	return "", fmt.Errorf("DateTruncExpr not implemented for %s dialect", b.self.String())
}

func (b *BaseDialect) DateDiff(_ runtimev1.TimeGrain, _, _ time.Time) (string, error) {
	return "", fmt.Errorf("DateDiff not implemented for %s dialect", b.self.String())
}

func (b *BaseDialect) IntervalSubtract(_, _ string, _ runtimev1.TimeGrain) (string, error) {
	return "", fmt.Errorf("IntervalSubtract not implemented for %s dialect", b.self.String())
}

func (b *BaseDialect) SelectTimeRangeBins(_, _ time.Time, _ runtimev1.TimeGrain, _ string, _ *time.Location, _, _ int) (string, []any, error) {
	return "", nil, fmt.Errorf("SelectTimeRangeBins not implemented for %s dialect", b.self.String())
}

func (b *BaseDialect) SelectInlineResults(result *Result) (string, []any, []any, error) {
	// check schema field type for compatibility
	for _, f := range result.Schema.Fields {
		if !CheckTypeCompatibility(f) {
			return "", nil, nil, fmt.Errorf("select inline: schema field type not supported %q: %w", f.Type.Code, ErrOptimizationFailure)
		}
	}

	values := make([]any, len(result.Schema.Fields))
	valuePtrs := make([]any, len(result.Schema.Fields))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var dimVals []any
	var args []any

	prefix := ""
	suffix := ""
	// creating inline query for all dialects in one loop, accumulating field exprs first and then creating the query can be more cleaner
	for result.Next() {
		if err := result.Scan(valuePtrs...); err != nil {
			return "", nil, nil, fmt.Errorf("select inline: failed to scan value: %w", err)
		}
		// format: SELECT ? AS a, ? AS b UNION ALL SELECT ...
		if prefix != "" {
			prefix += " UNION ALL "
		}
		prefix += "SELECT "
		dimVals = append(dimVals, values[0])
		for i, v := range values {
			if i > 0 {
				prefix += ", "
			}
			prefix += fmt.Sprintf("%s AS %s", "?", b.self.EscapeIdentifier(result.Schema.Fields[i].Name))
			args = append(args, v)
		}
	}
	if err := result.Err(); err != nil {
		return "", nil, nil, err
	}
	return prefix + suffix, args, dimVals, nil
}

func (b *BaseDialect) GetArgExpr(val any, typ runtimev1.Type_Code) (string, any, error) {
	// handle date types especially otherwise they get sent as time.Time args which will be treated as datetime/timestamp types in olap
	if typ == runtimev1.Type_CODE_DATE {
		t, ok := val.(time.Time)
		if !ok {
			return "", nil, fmt.Errorf("could not cast value %v to time.Time for date type", val)
		}
		return "CAST(? AS DATE)", t.Format(time.DateOnly), nil
	}
	return "?", val, nil
}

func (b *BaseDialect) GetValExpr(val any, typ runtimev1.Type_Code) (bool, string, error) {
	if val == nil {
		ok, expr := b.self.GetNullExpr(typ)
		if ok {
			return true, expr, nil
		}
		return false, "", fmt.Errorf("could not get null expr for type %q", typ)
	}
	switch typ {
	case runtimev1.Type_CODE_STRING:
		if s, ok := val.(string); ok {
			return true, b.self.EscapeStringValue(s), nil
		}
		return false, "", fmt.Errorf("could not cast value %v to string type", val)
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64,
		runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64,
		runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
		// check NaN and Inf
		if f, ok := val.(float64); ok && (math.IsNaN(f) || math.IsInf(f, 0)) {
			return true, "NULL", nil
		}
		return true, fmt.Sprintf("%v", val), nil
	case runtimev1.Type_CODE_BOOL:
		return true, fmt.Sprintf("%v", val), nil
	case runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_TIMESTAMP:
		if t, ok := val.(time.Time); ok {
			if ok, expr := b.self.GetDateTimeExpr(t); ok {
				return true, expr, nil
			}
			return false, "", fmt.Errorf("cannot get time expr for %s dialect", b.self.String())
		}
		return false, "", fmt.Errorf("unsupported time type %q", typ)
	case runtimev1.Type_CODE_DATE:
		if t, ok := val.(time.Time); ok {
			if ok, expr := b.self.GetDateExpr(t); ok {
				return true, expr, nil
			}
			return false, "", fmt.Errorf("cannot get date expr for %s dialect", b.self.String())
		}
		return false, "", fmt.Errorf("unsupported date type %q", typ)
	default:
		return false, "", fmt.Errorf("unsupported type %q", typ)
	}
}

func (b *BaseDialect) GetNullExpr(_ runtimev1.Type_Code) (bool, string) {
	return true, "NULL"
}

func (b *BaseDialect) GetDateTimeExpr(_ time.Time) (bool, string) {
	return false, ""
}

func (b *BaseDialect) GetDateExpr(_ time.Time) (bool, string) {
	return false, ""
}

func (b *BaseDialect) LookupExpr(_, _, _, _ string) (string, error) {
	return "", fmt.Errorf("lookup tables are not supported for %s dialect", b.self.String())
}

func (b *BaseDialect) LookupSelectExpr(_, _ string) (string, error) {
	return "", fmt.Errorf("lookup tables are not supported for %s dialect", b.self.String())
}

func (b *BaseDialect) SanitizeQueryForLogging(sql string) string { return sql }

func (b *BaseDialect) ColumnCardinality(db, dbSchema, table, column string) (string, error) {
	return "", fmt.Errorf("ColumnCardinality not implemented for %s dialect", b.self.String())
}

func (b *BaseDialect) ColumnDescriptiveStatistics(db, dbSchema, table, column string) (string, error) {
	return "", fmt.Errorf("ColumnDescriptiveStatistics not implemented for %s dialect", b.self.String())
}

func (b *BaseDialect) IsNonNullFinite(_ string) string {
	return "1=1"
}

func (b *BaseDialect) ColumnNullCount(db, dbSchema, table, column string) (string, error) {
	return fmt.Sprintf("SELECT count(*) AS count FROM %s WHERE %s IS NULL", b.self.EscapeTable(db, dbSchema, table), b.self.EscapeIdentifier(column)), nil
}

func (b *BaseDialect) ColumnNumericHistogram(db, dbSchema, table, column string) (string, error) {
	return "", fmt.Errorf("ColumnNumericHistogram not implemented for %s dialect", b.self.String())
}

func CheckTypeCompatibility(f *runtimev1.StructType_Field) bool {
	switch f.Type.Code {
	// types that align with native go types are supported
	case runtimev1.Type_CODE_STRING,
		runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64,
		runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64,
		runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64,
		runtimev1.Type_CODE_BOOL,
		runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
		return true
	default:
		return false
	}
}

func TempName(prefix string) string {
	return prefix + strings.ReplaceAll(uuid.New().String(), "-", "")
}
