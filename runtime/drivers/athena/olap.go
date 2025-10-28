package athena

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

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
	client, err := c.acquireClient(ctx)
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

	schema, err := rows.runtimeSchema()
	if err != nil {
		return nil, err
	}
	return &drivers.Result{
		Rows:   rows,
		Schema: schema,
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
func (c *Connection) Lookup(ctx context.Context, db string, schema string, name string) (*drivers.OlapTable, error) {
	meta, err := c.GetTable(ctx, db, schema, name)
	if err != nil {
		return nil, err
	}
	runtimeSchema := &runtimev1.StructType{
		Fields: make([]*runtimev1.StructType_Field, 0, len(meta.Schema)),
	}
	for name, typ := range meta.Schema {
		rtType, err := athenaTypeToRuntimeType(typ)
		if err != nil {
			return nil, err
		}
		runtimeSchema.Fields = append(runtimeSchema.Fields, &runtimev1.StructType_Field{
			Name: name,
			Type: rtType,
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
	r.currentRow = 1 // todo: check if there is header in next page also
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

func (r *rows) runtimeSchema() (*runtimev1.StructType, error) {
	fields := make([]*runtimev1.StructType_Field, len(r.columnInfo))
	for i, col := range r.columnInfo {
		fields[i] = &runtimev1.StructType_Field{
			Name: *col.Name,
		}
		colType, err := athenaTypeToRuntimeType(*col.Type)
		if err != nil {
			return nil, err
		}
		fields[i].Type = colType
	}
	res := &runtimev1.StructType{
		Fields: fields,
	}
	return res, nil
}

func athenaTypeToRuntimeType(colType string) (*runtimev1.Type, error) {
	t := &runtimev1.Type{}
	switch colType {
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
		return nil, fmt.Errorf("unknown type %q", colType)
	}
	return t, nil
}

func convertValue(colType string, val string) (any, error) {
	var (
		err error
		i   int64
		f   float64
	)
	switch colType {
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
	case "date", "time", "time with time zone", "timestamp", "timestamp with time zone":
		vv, err := scanTime(val)
		if !vv.Valid {
			return nil, fmt.Errorf("invalid time value")
		}
		return vv.Time, err
	// for all other datatypes return string value directly
	case "json", "char", "varchar", "varbinary", "row", "string", "binary",
		"struct", "interval year to month", "interval day to second", "decimal",
		"ipaddress", "array", "map", "unknown":
		return val, nil
	default:
		return nil, fmt.Errorf("unknown type %q with value %q", colType, val)
	}
}

// athenaTime represents a time.Time value that can be null.
// The athenaTime supports Athena's Date, Time and Timestamp data types,
// with or without time zone.
type athenaTime struct {
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

func scanTime(vv string) (athenaTime, error) {
	parts := strings.Split(vv, " ")
	if len(parts) > 1 && !unicode.IsDigit(rune(parts[len(parts)-1][0])) {
		return parseAthenaTimeWithLocation(vv)
	}
	return parseAthenaTime(vv)
}

func parseAthenaTime(v string) (athenaTime, error) {
	var t time.Time
	var err error
	for _, layout := range timeLayouts {
		t, err = time.ParseInLocation(layout, v, time.Local)
		if err == nil {
			return athenaTime{Valid: true, Time: t}, nil
		}
	}
	return athenaTime{}, err
}

func parseAthenaTimeWithLocation(v string) (athenaTime, error) {
	idx := strings.LastIndex(v, " ")
	if idx == -1 {
		return athenaTime{}, fmt.Errorf("cannot convert %v (%T) to time+zone", v, v)
	}
	stamp, location := v[:idx], v[idx+1:]
	loc, err := time.LoadLocation(location)
	if err != nil {
		return athenaTime{}, fmt.Errorf("cannot load timezone %q: %v", location, err)
	}
	var t time.Time
	for _, layout := range timeLayouts {
		t, err = time.ParseInLocation(layout, stamp, loc)
		if err == nil {
			return athenaTime{Valid: true, Time: t}, nil
		}
	}
	return athenaTime{}, err
}
