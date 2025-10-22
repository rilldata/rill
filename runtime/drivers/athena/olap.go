package athena

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var _ drivers.OLAPStore = &Connection{}

// Dialect implements drivers.OLAPStore.
func (c *Connection) Dialect() drivers.Dialect {
	return drivers.DialectAthena
}

// Exec implements drivers.OLAPStore.
func (c *Connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	panic("unimplemented")
}

// InformationSchema implements drivers.OLAPStore.
func (c *Connection) InformationSchema() drivers.OLAPInformationSchema {
	panic("unimplemented")
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *Connection) MayBeScaledToZero(ctx context.Context) bool {
	return true
}

// Query implements drivers.OLAPStore.
func (c *Connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	client, err := c.acquireClient(ctx)
	if err != nil {
		return nil, err
	}

	params := make([]string, len(stmt.Args))
	for i, v := range stmt.Args {
		params[i] = fmt.Sprint(v)
	}
	queryID, err := c.executeQuery(ctx, client, stmt.Query, c.config.Workgroup, c.config.OutputLocation, params)
	if err != nil {
		return nil, err
	}

	rows, err := newRows(ctx, client, *queryID)
	if err != nil {
		return nil, err
	}

	return &drivers.Result{
		Rows: rows,
	}, nil
}

// QuerySchema implements drivers.OLAPStore.
func (c *Connection) QuerySchema(ctx context.Context, query string, args []any) (*runtimev1.StructType, error) {
	return nil, drivers.ErrNotImplemented
}

// WithConnection implements drivers.OLAPStore.
func (c *Connection) WithConnection(ctx context.Context, priority int, fn drivers.WithConnectionFunc) error {
	return drivers.ErrNotImplemented
}

type rows struct {
	queryID    string
	client     *athena.Client
	results    *athena.GetQueryResultsOutput
	columnInfo []types.ColumnInfo

	currentRow int
	scannedRow []any
	scanErr    error
}

var _ drivers.Rows = &rows{}

func newRows(ctx context.Context, client *athena.Client, queryID string) (*rows, error) {
	results, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{QueryExecutionId: aws.String(queryID)})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var colInfo []types.ColumnInfo
	if results.ResultSet != nil && results.ResultSet.ResultSetMetadata != nil {
		colInfo = results.ResultSet.ResultSetMetadata.ColumnInfo
	} else {
		return nil, fmt.Errorf("no column info found in query results")
	}
	return &rows{
		queryID:    queryID,
		client:     client,
		results:    results,
		columnInfo: colInfo,
		currentRow: 1, // first row is header
		scannedRow: make([]any, len(colInfo)),
	}, nil
}

// Close implements drivers.Rows.
func (r *rows) Close() error {
	r.results = nil
	r.scannedRow = nil
	return nil
}

// Err implements drivers.Rows.
func (r *rows) Err() error {
	return r.scanErr
}

// MapScan implements drivers.Rows.
func (r *rows) MapScan(dest map[string]any) error {
	if r.scannedRow == nil {
		return fmt.Errorf("must call Next before Scan")
	}
	if dest == nil {
		return fmt.Errorf("nil destination map in MapScan")
	}

	for i, col := range r.columnInfo {
		dest[*col.Name] = r.scannedRow[i]
	}
	return nil
}

// Next implements drivers.Rows.
func (r *rows) Next() bool {
	if r.results == nil {
		r.scanErr = sql.ErrConnDone
		return false
	}
	if r.results.UpdateCount != nil && *r.results.UpdateCount != 0 {
		return false
	}

	// see if we have more rows in current page
	if r.currentRow < len(r.results.ResultSet.Rows) {
		row := r.results.ResultSet.Rows[r.currentRow]
		var err error
		for i, col := range row.Data {
			r.scannedRow[i], err = convertValue(*r.columnInfo[i].Type, *col.VarCharValue)
			if err != nil {
				r.scanErr = err
				return false
			}
		}
		r.currentRow++
		return true
	}
	// Fetch next page if available
	if r.results.NextToken == nil {
		return false
	}
	nextResults, err := r.client.GetQueryResults(context.Background(), &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(r.queryID),
		NextToken:        r.results.NextToken,
	})
	if err != nil {
		r.scanErr = fmt.Errorf("failed to get next page of query results: %w", err)
		return false
	}
	r.results = nextResults
	r.currentRow = 1 // todo: check if there is header in next page also
	// call Next to process the newly fetched page
	return r.Next()
}

// Scan implements drivers.Rows.
func (r *rows) Scan(dest ...any) error {
	if r.scannedRow == nil {
		return fmt.Errorf("must call Next before Scan")
	}
	if len(dest) != len(r.scannedRow) {
		return fmt.Errorf("expected %d destination arguments in Scan, got %d", len(r.scannedRow), len(dest))
	}

	for i := range dest {
		// Use reflection to set the value through the pointer
		rv := reflect.ValueOf(dest[i])
		if rv.Kind() != reflect.Ptr || rv.IsNil() {
			return fmt.Errorf("destination argument %d must be a non-nil pointer", i)
		}
		if r.scannedRow[i] == nil {
			// Set zero value for nil
			rv.Elem().Set(reflect.Zero(rv.Elem().Type()))
		} else {
			rv.Elem().Set(reflect.ValueOf(r.scannedRow[i]))
		}
	}
	return nil
}

func convertValue(colType string, val string) (any, error) {
	// https://stackoverflow.com/questions/30299649/parse-string-to-specific-type-of-int-int8-int16-int32-int64
	// https://prestodb.io/docs/current/language/types.html#integer
	var err error
	var i int64
	var f float64
	switch colType {
	case "tinyint":
		// strconv.ParseInt() behavior is to return (int64(0), err)
		// which is not as good as just return (nil, err)
		if i, err = strconv.ParseInt(val, 10, 8); err != nil {
			return nil, err
		}
		return int8(i), nil
	case "smallint":
		if i, err = strconv.ParseInt(val, 10, 16); err != nil {
			return nil, err
		}
		return int16(i), nil
	case "integer":
		if i, err = strconv.ParseInt(val, 10, 32); err != nil {
			return nil, err
		}
		return int32(i), nil
	case "bigint":
		if i, err = strconv.ParseInt(val, 10, 64); err != nil {
			return nil, err
		}
		return i, nil
	case "float", "real":
		if f, err = strconv.ParseFloat(val, 32); err != nil {
			return nil, err
		}
		return float32(f), nil
	case "double":
		if f, err = strconv.ParseFloat(val, 64); err != nil {
			return nil, err
		}
		return f, nil
	// for binary, we assume all chars are 0 or 1; for json,
	// we assume the json syntax is correct. Leave to caller to verify it.
	case "json", "char", "varchar", "varbinary", "row", "string", "binary",
		"struct", "interval year to month", "interval day to second", "decimal",
		"ipaddress", "array", "map", "unknown":
		return val, nil
	case "boolean":
		val, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return val, nil
	case "date", "time", "time with time zone", "timestamp", "timestamp with time zone":
		vv, err := scanTime(val)
		if !vv.Valid {
			return nil, fmt.Errorf("invalid time value")
		}
		return vv.Time, err
	default:
		return nil, fmt.Errorf("unknown type %q with value %q", colType, val)
	}
}

// AthenaTime represents a time.Time value that can be null.
// The AthenaTime supports Athena's Date, Time and Timestamp data types,
// with or without time zone.
type AthenaTime struct {
	Time  time.Time
	Valid bool
}

var timeLayouts = []string{
	"2006-01-02",
	"15:04:05.000",
	"2006-01-02 15:04:05.000000000",
	"2006-01-02 15:04:05.000000",
	"2006-01-02 15:04:05.000",
}

func scanTime(vv string) (AthenaTime, error) {
	parts := strings.Split(vv, " ")
	if len(parts) > 1 && !unicode.IsDigit(rune(parts[len(parts)-1][0])) {
		return parseAthenaTimeWithLocation(vv)
	}
	return parseAthenaTime(vv)
}

func parseAthenaTime(v string) (AthenaTime, error) {
	var t time.Time
	var err error
	for _, layout := range timeLayouts {
		t, err = time.ParseInLocation(layout, v, time.Local)
		if err == nil {
			return AthenaTime{Valid: true, Time: t}, nil
		}
	}
	return AthenaTime{}, err
}

func parseAthenaTimeWithLocation(v string) (AthenaTime, error) {
	idx := strings.LastIndex(v, " ")
	if idx == -1 {
		return AthenaTime{}, fmt.Errorf("cannot convert %v (%T) to time+zone", v, v)
	}
	stamp, location := v[:idx], v[idx+1:]
	loc, err := time.LoadLocation(location)
	if err != nil {
		return AthenaTime{}, fmt.Errorf("cannot load timezone %q: %v", location, err)
	}
	var t time.Time
	for _, layout := range timeLayouts {
		t, err = time.ParseInLocation(layout, stamp, loc)
		if err == nil {
			return AthenaTime{Valid: true, Time: t}, nil
		}
	}
	return AthenaTime{}, err
}
