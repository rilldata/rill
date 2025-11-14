package drivers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/timeutil"

	// Load IANA time zone data
	_ "time/tzdata"
)

var (
	// ErrUnsupportedConnector is returned from Ingest for unsupported connectors.
	ErrUnsupportedConnector = errors.New("drivers: connector not supported")
	// ErrOptimizationFailure is returned when an optimization fails.
	ErrOptimizationFailure = errors.New("drivers: optimization failure")

	DefaultQuerySchemaTimeout = 30 * time.Second

	dictPwdRegex = regexp.MustCompile(`PASSWORD\s+'[^']*'`)
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
}

// Dialect enumerates OLAP query languages.
type Dialect int

const (
	DialectUnspecified Dialect = iota
	DialectDuckDB
	DialectDruid
	DialectClickHouse
	DialectPinot

	// Below dialects are not fully supported dialects.
	DialectBigQuery
	DialectSnowflake
	DialectAthena
	DialectRedshift
	DialectMySQL
	DialectPostgres
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
	case DialectPinot:
		return "pinot"
	case DialectBigQuery:
		return "bigquery"
	case DialectSnowflake:
		return "snowflake"
	case DialectAthena:
		return "athena"
	case DialectRedshift:
		return "redshift"
	case DialectMySQL:
		return "mysql"
	case DialectPostgres:
		return "postgres"
	default:
		panic("not implemented")
	}
}

func (d Dialect) CanPivot() bool {
	return d == DialectDuckDB
}

// EscapeIdentifier returns an escaped SQL identifier in the dialect.
func (d Dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}

	switch d {
	case DialectMySQL, DialectBigQuery:
		// MySQL uses backticks for quoting identifiers
		// Replace any backticks inside the identifier with double backticks.
		return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))

	default:
		// Most other dialects follow ANSI SQL: use double quotes.
		// Replace any internal double quotes with escaped double quotes.
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(ident, `"`, `""`)) // nolint:gocritic
	}
}

func (d Dialect) EscapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

func (d Dialect) ConvertToDateTruncSpecifier(grain runtimev1.TimeGrain) string {
	var str string
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		str = "MILLISECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		str = "SECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		str = "MINUTE"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		str = "HOUR"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		str = "DAY"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		str = "WEEK"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		str = "MONTH"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		str = "QUARTER"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		str = "YEAR"
	}

	if d == DialectClickHouse {
		return strings.ToLower(str)
	}
	return str
}

func (d Dialect) SupportsILike() bool {
	return d != DialectDruid && d != DialectPinot
}

// RequiresCastForLike returns true if the dialect requires an expression used in a LIKE or ILIKE condition to explicitly be cast to type TEXT.
func (d Dialect) RequiresCastForLike() bool {
	return d == DialectClickHouse
}

func (d Dialect) SupportsRegexMatch() bool {
	return d == DialectDruid
}

func (d Dialect) GetRegexMatchFunction() string {
	switch d {
	case DialectDruid:
		return "REGEXP_LIKE"
	default:
		panic(fmt.Sprintf("unsupported dialect %q for regex match", d))
	}
}

// EscapeTable returns an escaped table name with database, schema and table.
func (d Dialect) EscapeTable(db, schema, table string) string {
	if d == DialectDuckDB {
		return d.EscapeIdentifier(table)
	}
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

// EscapeMember returns an escaped member name with table alias and column name.
func (d Dialect) EscapeMember(tbl, name string) string {
	if tbl == "" {
		return d.EscapeIdentifier(name)
	}
	return fmt.Sprintf("%s.%s", d.EscapeIdentifier(tbl), d.EscapeIdentifier(name))
}

func (d Dialect) DimensionSelect(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_Dimension) (dimSelect, unnestClause string, err error) {
	colName := d.EscapeIdentifier(dim.Name)
	if !dim.Unnest || d == DialectDruid {
		expr, err := d.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return fmt.Sprintf(`(%s) as %s`, expr, colName), "", nil
	}
	if dim.Unnest && d == DialectClickHouse {
		expr, err := d.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return fmt.Sprintf(`arrayJoin(%s) as %s`, expr, colName), "", nil
	}

	unnestColName := d.EscapeIdentifier(tempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := tempName("tbl")
	sel := fmt.Sprintf(`%s as %s`, unnestColName, colName)
	if dim.Expression == "" {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl_name("unnested_colName") ...
		return sel, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, d.EscapeTable(db, dbSchema, table), colName, unnestTableName, unnestColName), nil
	}

	return sel, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName), nil
}

func (d Dialect) DimensionSelectPair(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_Dimension) (expr, alias, unnestClause string, err error) {
	colName := d.EscapeIdentifier(dim.Name)
	if !dim.Unnest || d == DialectDruid {
		ex, err := d.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return ex, colName, "", nil
	}

	unnestColName := d.EscapeIdentifier(tempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := tempName("tbl")
	if dim.Expression == "" {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl_name("unnested_colName") ...
		return unnestColName, colName, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, d.EscapeTable(db, dbSchema, table), colName, unnestTableName, unnestColName), nil
	}

	return unnestColName, colName, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName), nil
}

func (d Dialect) LateralUnnest(expr, tableAlias, colName string) (tbl string, tupleStyle, auto bool, err error) {
	if d == DialectDruid || d == DialectPinot {
		return "", false, true, nil
	}
	if d == DialectClickHouse {
		// using `LEFT ARRAY JOIN` instead of just `ARRAY JOIN` as it includes empty arrays in the result set with zero values
		return fmt.Sprintf("LEFT ARRAY JOIN %s as %s", expr, d.EscapeIdentifier(colName)), false, false, nil
	}
	return fmt.Sprintf(`LATERAL UNNEST(%s) %s(%s)`, expr, tableAlias, d.EscapeIdentifier(colName)), true, false, nil
}

func (d Dialect) UnnestSQLSuffix(tbl string) string {
	if d == DialectDruid || d == DialectPinot {
		panic("Druid and Pinot auto unnests")
	}
	if d == DialectClickHouse {
		return fmt.Sprintf(" %s", tbl)
	}
	return fmt.Sprintf(", %s", tbl)
}

func (d Dialect) MetricsViewDimensionExpression(dimension *runtimev1.MetricsViewSpec_Dimension) (string, error) {
	if dimension.LookupTable != "" {
		var keyExpr string
		if dimension.Column != "" {
			keyExpr = d.EscapeIdentifier(dimension.Column)
		} else if dimension.Expression != "" {
			keyExpr = dimension.Expression
		} else {
			return "", fmt.Errorf("dimension %q has a lookup table but no column or expression defined", dimension.Name)
		}
		return d.LookupExpr(dimension.LookupTable, dimension.LookupValueColumn, keyExpr, dimension.LookupDefaultExpression)
	}

	if dimension.Expression != "" {
		return dimension.Expression, nil
	}
	if dimension.Column != "" {
		return d.EscapeIdentifier(dimension.Column), nil
	}
	// Backwards compatibility for older projects that have not run reconcile on this metrics view.
	// In that case `column` will not be present.
	return d.EscapeIdentifier(dimension.Name), nil
}

// AnyValueExpression applies the ANY_VALUE aggregation function (or equivalent) to the given expression.
func (d Dialect) AnyValueExpression(expr string) string {
	return fmt.Sprintf("ANY_VALUE(%s)", expr)
}

func (d Dialect) GetTimeDimensionParameter() string {
	if d == DialectPinot {
		return "CAST(? AS TIMESTAMP)"
	}
	return "?"
}

func (d Dialect) SafeDivideExpression(numExpr, denExpr string) string {
	switch d {
	case DialectDruid:
		return fmt.Sprintf("SAFE_DIVIDE(%s, CAST(%s AS DOUBLE))", numExpr, denExpr)
	default:
		return fmt.Sprintf("(%s)/CAST(%s AS DOUBLE)", numExpr, denExpr)
	}
}

func (d Dialect) OrderByExpression(name string, desc bool) string {
	res := d.EscapeIdentifier(name)
	if desc {
		res += " DESC"
	}
	if d == DialectDuckDB {
		res += " NULLS LAST"
	}
	return res
}

func (d Dialect) JoinOnExpression(lhs, rhs string) string {
	if d == DialectClickHouse {
		return fmt.Sprintf("isNotDistinctFrom(%s, %s)", lhs, rhs)
	}
	return fmt.Sprintf("%s IS NOT DISTINCT FROM %s", lhs, rhs)
}

func (d Dialect) DateTruncExpr(dim *runtimev1.MetricsViewSpec_Dimension, grain runtimev1.TimeGrain, tz string, firstDayOfWeek, firstMonthOfYear int) (string, error) {
	if tz == "UTC" || tz == "Etc/UTC" {
		tz = ""
	}

	if tz != "" {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return "", fmt.Errorf("invalid time zone %q: %w", tz, err)
		}
	}

	var specifier string
	if tz != "" && d == DialectDruid {
		specifier = druidTimeFloorSpecifier(grain)
	} else {
		specifier = d.ConvertToDateTruncSpecifier(grain)
	}

	var expr string
	if dim.Expression != "" {
		expr = fmt.Sprintf("(%s)", dim.Expression)
	} else {
		expr = d.EscapeIdentifier(dim.Column)
	}

	switch d {
	case DialectDuckDB:
		var shift string
		if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
			offset := 8 - firstDayOfWeek
			shift = fmt.Sprintf("%d DAY", offset)
		} else if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
			offset := 13 - firstMonthOfYear
			shift = fmt.Sprintf("%d MONTH", offset)
		}

		if tz == "" {
			if shift == "" {
				return fmt.Sprintf("date_trunc('%s', %s::TIMESTAMP)::TIMESTAMP", specifier, expr), nil
			}
			return fmt.Sprintf("date_trunc('%s', %s::TIMESTAMP + INTERVAL %s)::TIMESTAMP - INTERVAL %s", specifier, expr, shift, shift), nil
		}

		// Optimization: date_trunc is faster for day+ granularity
		switch grain {
		case runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_WEEK, runtimev1.TimeGrain_TIME_GRAIN_MONTH, runtimev1.TimeGrain_TIME_GRAIN_QUARTER, runtimev1.TimeGrain_TIME_GRAIN_YEAR:
			if shift == "" {
				return fmt.Sprintf("timezone('%s', date_trunc('%s', timezone('%s', %s::TIMESTAMPTZ)))::TIMESTAMP", tz, specifier, tz, expr), nil
			}
			return fmt.Sprintf("timezone('%s', date_trunc('%s', timezone('%s', %s::TIMESTAMPTZ) + INTERVAL %s) - INTERVAL %s)::TIMESTAMP", tz, specifier, tz, expr, shift, shift), nil
		}

		if shift == "" {
			return fmt.Sprintf("time_bucket(INTERVAL '1 %s', %s::TIMESTAMPTZ, '%s')", specifier, expr, tz), nil
		}
		return fmt.Sprintf("time_bucket(INTERVAL '1 %s', %s::TIMESTAMPTZ + INTERVAL %s, '%s') - INTERVAL %s", specifier, expr, shift, tz, shift), nil
	case DialectDruid:
		var shift int
		var shiftPeriod string
		if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
			shift = 8 - firstDayOfWeek
			shiftPeriod = "P1D"
		} else if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
			shift = 13 - firstMonthOfYear
			shiftPeriod = "P1M"
		}

		if tz == "" {
			if shift == 0 {
				return fmt.Sprintf("date_trunc('%s', %s)", specifier, expr), nil
			}
			return fmt.Sprintf("time_shift(date_trunc('%s', time_shift(%s, '%s', %d)), '%s', -%d)", specifier, expr, shiftPeriod, shift, shiftPeriod, shift), nil
		}

		if shift == 0 {
			return fmt.Sprintf("time_floor(%s, '%s', null, '%s')", expr, specifier, tz), nil
		}
		return fmt.Sprintf("time_shift(time_floor(time_shift(%s, '%s', %d), '%s', null, '%s'), '%s', -%d)", expr, shiftPeriod, shift, specifier, tz, shiftPeriod, shift), nil
	case DialectClickHouse:
		var shift string
		if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
			offset := 8 - firstDayOfWeek
			shift = fmt.Sprintf("%d DAY", offset)
		} else if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
			offset := 13 - firstMonthOfYear
			shift = fmt.Sprintf("%d MONTH", offset)
		}

		if tz == "" {
			if shift == "" {
				return fmt.Sprintf("date_trunc('%s', %s, 'UTC')::DateTime64", specifier, expr), nil
			}
			return fmt.Sprintf("date_trunc('%s', %s + INTERVAL %s, 'UTC')::DateTime64 - INTERVAL %s", specifier, expr, shift, shift), nil
		}

		if shift == "" {
			return fmt.Sprintf("date_trunc('%s', %s::DateTime64(6, '%s'))::DateTime64(6, '%s')", specifier, expr, tz, tz), nil
		}
		return fmt.Sprintf("date_trunc('%s', %s::DateTime64(6, '%s') + INTERVAL %s)::DateTime64(6, '%s') - INTERVAL %s", specifier, expr, tz, shift, tz, shift), nil
	case DialectPinot:
		// TODO: Handle tz instead of ignoring it.
		// TODO: Handle firstDayOfWeek and firstMonthOfYear. NOTE: We currently error when configuring these for Pinot in runtime/validate.go.
		// adding a cast to timestamp to get the the output type as TIMESTAMP otherwise it returns a long
		if tz == "" {
			return fmt.Sprintf("CAST(date_trunc('%s', %s, 'MILLISECONDS') AS TIMESTAMP)", specifier, expr), nil
		}
		return fmt.Sprintf("CAST(date_trunc('%s', %s, 'MILLISECONDS', '%s') AS TIMESTAMP)", specifier, expr, tz), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", d)
	}
}

func (d Dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	switch d {
	case DialectClickHouse:
		return fmt.Sprintf("DATEDIFF('%s', parseDateTimeBestEffort('%s'), parseDateTimeBestEffort('%s'))", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
	case DialectDruid:
		return fmt.Sprintf("TIMESTAMPDIFF(%q, TIME_PARSE('%s'), TIME_PARSE('%s'))", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
	case DialectDuckDB:
		return fmt.Sprintf("DATEDIFF('%s', TIMESTAMP '%s', TIMESTAMP '%s')", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
	case DialectPinot:
		return fmt.Sprintf("DATEDIFF('%s', %d, %d)", unit, t1.UnixMilli(), t2.UnixMilli()), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", d)
	}
}

func (d Dialect) IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error) {
	switch d {
	case DialectClickHouse, DialectDruid, DialectDuckDB:
		return fmt.Sprintf("(%s - INTERVAL (%s) %s)", tsExpr, unitExpr, d.ConvertToDateTruncSpecifier(grain)), nil
	case DialectPinot:
		return fmt.Sprintf("CAST((dateAdd('%s', -1 * %s, %s)) AS TIMESTAMP)", d.ConvertToDateTruncSpecifier(grain), unitExpr, tsExpr), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", d)
	}
}

func (d Dialect) SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location) (string, []any, error) {
	var args []any
	switch d {
	case DialectDuckDB:
		return fmt.Sprintf("SELECT range AS %s FROM range('%s'::TIMESTAMP, '%s'::TIMESTAMP, INTERVAL '1 %s')", d.EscapeIdentifier(alias), start.Format(time.RFC3339), end.Format(time.RFC3339), d.ConvertToDateTruncSpecifier(grain)), nil, nil
	case DialectClickHouse:
		// format - SELECT c1 AS "alias" FROM VALUES(toDateTime('2021-01-01 00:00:00'), toDateTime('2021-01-01 00:00:00'),...)
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("SELECT c1 AS %s FROM VALUES(", d.EscapeIdentifier(alias)))
		for t := start; t.Before(end); t = timeutil.OffsetTime(t, timeutil.TimeGrainFromAPI(grain), 1, tz) {
			if t != start {
				sb.WriteString(", ")
			}
			sb.WriteString("?")
			args = append(args, t)
		}
		sb.WriteString(")")
		return sb.String(), args, nil
	case DialectDruid, DialectPinot:
		// generate select like - SELECT * FROM (
		//  VALUES
		//  (CAST('2006-01-02T15:04:05Z' AS TIMESTAMP)),
		//  (CAST('2006-01-02T15:04:05Z' AS TIMESTAMP))
		// ) t (time)
		var sb strings.Builder
		sb.WriteString("SELECT * FROM (VALUES ")
		for t := start; t.Before(end); t = timeutil.OffsetTime(t, timeutil.TimeGrainFromAPI(grain), 1, tz) {
			if t != start {
				sb.WriteString(", ")
			}
			sb.WriteString("(CAST(? AS TIMESTAMP))")
			args = append(args, t)
		}
		sb.WriteString(fmt.Sprintf(") t (%s)", d.EscapeIdentifier(alias)))
		return sb.String(), args, nil
	default:
		return "", nil, fmt.Errorf("unsupported dialect %q", d)
	}
}

// SelectInlineResults returns a SQL query which inline results from the result set supplied along with the positional arguments and dimension values.
func (d Dialect) SelectInlineResults(result *Result) (string, []any, []any, error) {
	// check schema field type for compatibility
	for _, f := range result.Schema.Fields {
		if !d.checkTypeCompatibility(f) {
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

	rows := 0
	prefix := ""
	suffix := ""
	// creating inline query for all dialects in one loop, accumulating field exprs first and then creating the query can be more cleaner
	for result.Next() {
		if err := result.Scan(valuePtrs...); err != nil {
			return "", nil, nil, fmt.Errorf("select inline: failed to scan value: %w", err)
		}
		if d == DialectDruid || d == DialectDuckDB || d == DialectPinot {
			// format - select * from (values (1, 2), (3, 4)) t(a, b)
			if rows == 0 {
				prefix = "SELECT * FROM (VALUES "
				suffix = "t("
			}
			if rows > 0 {
				prefix += ", "
			}
		} else if d == DialectClickHouse {
			// format - SELECT c1 AS a, c2 AS b FROM VALUES((1, 2), (3, 4))
			if rows == 0 {
				prefix = "SELECT "
				suffix = " FROM VALUES ("
			}
			if rows > 0 {
				suffix += ", "
			}
		} else {
			// format - select 1 as a, 2 as b union all select 3 as a, 4 as b
			if rows > 0 {
				prefix += " UNION ALL "
			}
			prefix += "SELECT "
		}

		dimVals = append(dimVals, values[0])
		for i, v := range values {
			if d == DialectDruid || d == DialectDuckDB || d == DialectPinot {
				if i == 0 {
					prefix += "("
				} else {
					prefix += ", "
				}
				if rows == 0 {
					suffix += d.EscapeIdentifier(result.Schema.Fields[i].Name)
					if i != len(result.Schema.Fields)-1 {
						suffix += ", "
					}
				}
			} else if d == DialectClickHouse {
				if i == 0 {
					suffix += "("
				} else {
					suffix += ", "
				}
				if rows == 0 {
					prefix += fmt.Sprintf("c%d AS %s", i+1, d.EscapeIdentifier(result.Schema.Fields[i].Name))
					if i != len(result.Schema.Fields)-1 {
						prefix += ", "
					}
				}
			} else if i > 0 {
				prefix += ", "
			}

			if d == DialectDuckDB {
				prefix += "?"
				args = append(args, v)
			} else if d == DialectClickHouse {
				suffix += "?"
				args = append(args, v)
			} else if d == DialectDruid || d == DialectPinot {
				ok, expr, err := d.GetValExpr(v, result.Schema.Fields[i].Type.Code)
				if err != nil {
					return "", nil, nil, fmt.Errorf("select inline: failed to get value expression: %w", err)
				}
				if !ok {
					return "", nil, nil, fmt.Errorf("select inline: unsupported value type %q: %w", result.Schema.Fields[i].Type.Code, ErrOptimizationFailure)
				}
				prefix += expr
			} else {
				prefix += fmt.Sprintf("%s AS %s", "?", d.EscapeIdentifier(result.Schema.Fields[i].Name))
				args = append(args, v)
			}
		}

		if d == DialectDruid || d == DialectDuckDB || d == DialectPinot {
			prefix += ")"
			if rows == 0 {
				suffix += ")"
			}
		} else if d == DialectClickHouse {
			suffix += ")"
		}

		rows++
	}
	err := result.Err()
	if err != nil {
		return "", nil, nil, err
	}

	if d == DialectDruid || d == DialectDuckDB || d == DialectPinot {
		prefix += ") "
	} else if d == DialectClickHouse {
		suffix += ")"
	}

	return prefix + suffix, args, dimVals, nil
}

func (d Dialect) GetValExpr(val any, typ runtimev1.Type_Code) (bool, string, error) {
	if val == nil {
		ok, expr := d.GetNullExpr(typ)
		if ok {
			return true, expr, nil
		}
		return false, "", fmt.Errorf("could not get null expr for type %q", typ)
	}
	switch typ {
	case runtimev1.Type_CODE_STRING:
		if s, ok := val.(string); ok {
			return true, d.EscapeStringValue(s), nil
		}
		return false, "", fmt.Errorf("could not cast value %v to string type", val)
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64, runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
		// check NaN and Inf
		if f, ok := val.(float64); ok && (math.IsNaN(f) || math.IsInf(f, 0)) {
			return true, "NULL", nil
		}

		return true, fmt.Sprintf("%v", val), nil
	case runtimev1.Type_CODE_BOOL:
		return true, fmt.Sprintf("%v", val), nil
	case runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
		if t, ok := val.(time.Time); ok {
			if ok, expr := d.GetTimeExpr(t); ok {
				return true, expr, nil
			}
			return false, "", fmt.Errorf("cannot get time expr for dialect %q", d)
		}
		return false, "", fmt.Errorf("unsupported time type %q", typ)
	default:
		return false, "", fmt.Errorf("unsupported type %q", typ)
	}
}

func (d Dialect) GetNullExpr(typ runtimev1.Type_Code) (bool, string) {
	if d == DialectDruid {
		switch typ {
		case runtimev1.Type_CODE_STRING:
			return true, "CAST(NULL AS VARCHAR)"
		case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256, runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64, runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
			return true, "CAST(NULL AS INTEGER)"
		case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64, runtimev1.Type_CODE_DECIMAL:
			return true, "CAST(NULL AS DOUBLE)"
		case runtimev1.Type_CODE_BOOL:
			return true, "CAST(NULL AS BOOLEAN)"
		case runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
			return true, "CAST(NULL AS TIMESTAMP)"
		default:
			return false, ""
		}
	}
	return true, "NULL"
}

func (d Dialect) GetTimeExpr(t time.Time) (bool, string) {
	switch d {
	case DialectClickHouse:
		return true, fmt.Sprintf("parseDateTimeBestEffort('%s')", t.Format(time.RFC3339Nano))
	case DialectDuckDB, DialectDruid:
		return true, fmt.Sprintf("CAST('%s' AS TIMESTAMP)", t.Format(time.RFC3339Nano))
	case DialectPinot:
		return true, fmt.Sprintf("CAST(%d AS TIMESTAMP)", t.UnixMilli())
	default:
		return false, ""
	}
}

func (d Dialect) LookupExpr(lookupTable, lookupValueColumn, lookupKeyExpr, lookupDefaultExpression string) (string, error) {
	switch d {
	case DialectClickHouse:
		if lookupDefaultExpression != "" {
			return fmt.Sprintf("dictGetOrDefault('%s', '%s', %s, %s)", lookupTable, lookupValueColumn, lookupKeyExpr, lookupDefaultExpression), nil
		}
		return fmt.Sprintf("dictGet('%s', '%s', %s)", lookupTable, lookupValueColumn, lookupKeyExpr), nil
	default:
		// Druid already does reverse lookup inherently so defining lookup expression directly as dimension expression should be ok.
		// For Duckdb I think we should just avoid going into this complexity as it should not matter much at that scale.
		return "", fmt.Errorf("lookup tables are not supported for dialect %q", d)
	}
}

func (d Dialect) LookupSelectExpr(lookupTable, lookupKeyColumn string) (string, error) {
	switch d {
	case DialectClickHouse:
		return fmt.Sprintf("SELECT %s FROM dictionary(%s)", d.EscapeIdentifier(lookupKeyColumn), d.EscapeIdentifier(lookupTable)), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", d)
	}
}

func (d Dialect) SanitizeQueryForLogging(sql string) string {
	if d == DialectClickHouse {
		// replace inline "PASSWORD 'pwd'" for dict source with "PASSWORD '***'"
		sql = dictPwdRegex.ReplaceAllString(sql, "PASSWORD '***'")
	}
	return sql
}

func (d Dialect) checkTypeCompatibility(f *runtimev1.StructType_Field) bool {
	switch f.Type.Code {
	// types that align with native go types are supported
	case runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64, runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64, runtimev1.Type_CODE_BOOL, runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
		return true
	default:
		return false
	}
}

func druidTimeFloorSpecifier(grain runtimev1.TimeGrain) string {
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "PT0.001S"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "PT1S"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "PT1M"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "PT1H"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "P1D"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "P1W"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "P1M"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "P3M"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "P1Y"
	}
	panic(fmt.Errorf("invalid time grain enum value %d", int(grain)))
}

func tempName(prefix string) string {
	return prefix + strings.ReplaceAll(uuid.New().String(), "-", "")
}
