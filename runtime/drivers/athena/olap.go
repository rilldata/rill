package athena

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/sqlconvert"
)

var _ drivers.OLAPStore = &Connection{}

// Dialect implements drivers.OLAPStore.
func (c *Connection) Dialect() drivers.Dialect {
	return drivers.DialectAthena
}

// Exec implements drivers.OLAPStore.
func (c *Connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Query(ctx, stmt)
	if err != nil {
		return err
	}
	return res.Rows.Close()
}

// InformationSchema implements drivers.OLAPStore.
func (c *Connection) InformationSchema() drivers.OLAPInformationSchema {
	return c
}

// MayBeScaledToZero implements drivers.OLAPStore.
func (c *Connection) MayBeScaledToZero(ctx context.Context) bool {
	return true
}

// Query implements drivers.OLAPStore.
func (c *Connection) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, err
	}

	if stmt.DryRun {
		stmt.Query = fmt.Sprintf("EXPLAIN %s", stmt.Query)
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
		Rows:   rows,
		Schema: rows.runtimeSchema(),
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

// All implements drivers.OLAPInformationSchema.
func (c *Connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.OlapTable, string, error) {
	return drivers.AllFromInformationSchema(ctx, like, pageSize, pageToken, c)
}

// LoadPhysicalSize implements drivers.OLAPInformationSchema.
func (c *Connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.OlapTable) error {
	return nil
}

// Lookup implements drivers.OLAPInformationSchema.
func (c *Connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}
	runtimeSchema := &runtimev1.StructType{
		Fields: make([]*runtimev1.StructType_Field, 0, len(meta.Schema)),
	}
	for name, typ := range meta.Schema {
		runtimeSchema.Fields = append(runtimeSchema.Fields, &runtimev1.StructType_Field{
			Name: name,
			Type: athenaTypeToRuntimeType(typ),
		})
	}
	return &drivers.OlapTable{
		Database:          db,
		DatabaseSchema:    schema,
		Name:              name,
		View:              meta.View,
		Schema:            runtimeSchema,
		UnsupportedCols:   nil,
		PhysicalSizeBytes: 0,
	}, nil
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
	if r.scanErr != nil {
		return false
	}

	// see if we have more rows in current page
	if r.currentRow < len(r.results.ResultSet.Rows) {
		row := r.results.ResultSet.Rows[r.currentRow]
		var err error
		for i, col := range row.Data {
			if col.VarCharValue == nil {
				r.scannedRow[i] = nil
				continue
			}
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
	r.currentRow = 0 // there is no header in next page
	// call Next to process the newly fetched page
	return r.Next()
}

// Scan implements drivers.Rows.
func (r *rows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	if r.scannedRow == nil {
		return fmt.Errorf("must call Next before Scan")
	}
	if len(dest) != len(r.scannedRow) {
		return fmt.Errorf("expected %d destination arguments in Scan, got %d", len(r.scannedRow), len(dest))
	}

	for i := range dest {
		err := sqlconvert.ConvertAssign(dest[i], r.scannedRow[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rows) runtimeSchema() *runtimev1.StructType {
	fields := make([]*runtimev1.StructType_Field, len(r.columnInfo))
	for i, col := range r.columnInfo {
		fields[i] = &runtimev1.StructType_Field{
			Name: *col.Name,
			Type: athenaTypeToRuntimeType(*col.Type),
		}
	}
	res := &runtimev1.StructType{
		Fields: fields,
	}
	return res
}

func athenaTypeToRuntimeType(colType string) *runtimev1.Type {
	t := &runtimev1.Type{}
	switch strings.ToLower(colType) {
	case "tinyint":
		t.Code = runtimev1.Type_CODE_INT8
	case "smallint":
		t.Code = runtimev1.Type_CODE_INT16
	case "integer":
		t.Code = runtimev1.Type_CODE_INT32
	case "bigint":
		t.Code = runtimev1.Type_CODE_INT64
	case "float", "real":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "double":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "boolean":
		t.Code = runtimev1.Type_CODE_BOOL
	case "date", "time", "time with time zone", "timestamp", "timestamp with time zone":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "json", "char", "varchar", "varbinary", "row", "string", "binary",
		"struct", "interval year to month", "interval day to second", "decimal",
		"ipaddress", "array", "map", "unknown":
		t.Code = runtimev1.Type_CODE_STRING
	default:
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	}
	return t
}

var timeTZLayout = "15:04:05-07:00"

func convertValue(colType, val string) (any, error) {
	var (
		err error
		i   int64
		f   float64
	)
	switch strings.ToLower(colType) {
	case "tinyint":
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
	case "boolean":
		return strconv.ParseBool(val)
	case "date":
		t, err := time.Parse(time.DateOnly, val)
		if err != nil {
			return nil, err
		}
		return t, nil
	case "time":
		t, err := time.Parse(time.TimeOnly, val)
		if err != nil {
			return nil, err
		}
		return t, nil
	case "time with time zone":
		t, err := time.Parse(timeTZLayout, val)
		if err != nil {
			return nil, err
		}
		return t, nil
	case "timestamp":
		t, err := time.Parse(time.DateTime, val)
		if err != nil {
			return nil, err
		}
		return t, nil
	case "timestamp with time zone":
		t, err := parseAthenaTimeWithLocation(time.DateTime, val)
		if err != nil {
			return nil, err
		}
		return t, err
	// for all other datatypes return string value directly
	case "json", "char", "varchar", "varbinary", "row", "string", "binary",
		"struct", "interval year to month", "interval day to second", "decimal",
		"ipaddress", "array", "map", "unknown":
		return val, nil
	default:
		return val, nil
	}
}

func parseAthenaTimeWithLocation(layout, v string) (time.Time, error) {
	idx := strings.LastIndexAny(v, " ")
	if idx == -1 {
		return time.Parse(layout, v)
	}
	stamp, location := strings.TrimSpace(v[:idx]), v[idx+1:]
	var loc *time.Location
	var err error
	if isTimezoneOffset(location) {
		loc, err = parseOffSet(location)
		if err != nil {
			return time.Time{}, err
		}
	} else {
		loc, err = time.LoadLocation(location)
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot load timezone %q: %w", location, err)
		}
	}
	t, err := time.ParseInLocation(layout, stamp, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse time %q with layout %q: %w", stamp, layout, err)
	}
	return t, nil
}

func parseOffSet(offset string) (*time.Location, error) {
	sign := 1
	if strings.HasPrefix(offset, "-") {
		sign = -1
		offset = offset[1:] // remove sign
	} else if strings.HasPrefix(offset, "+") {
		offset = offset[1:]
	}

	hours, mins := 0, 0
	_, err := fmt.Sscanf(offset, "%d:%d", &hours, &mins)
	if err != nil {
		return nil, err
	}
	totalMinutes := sign * (hours*60 + mins)
	loc := time.FixedZone("", totalMinutes*60)
	return loc, nil
}

// isTimezoneOffset checks if the string is a timezone offset like +05:30 or -06:00
func isTimezoneOffset(s string) bool {
	if s == "" {
		return false
	}

	if s[0] != '+' && s[0] != '-' {
		return false
	}

	for _, c := range s[1:] {
		if c != ':' && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}
