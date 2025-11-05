package redshift

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	redshift_types "github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/sqlconvert"
)

var _ drivers.OLAPStore = &Connection{}

// Dialect implements drivers.OLAPStore.
func (c *Connection) Dialect() drivers.Dialect {
	return drivers.DialectRedshift
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

	// Convert args to parameters (Redshift Data API uses string parameters)
	var params []redshift_types.SqlParameter
	if len(stmt.Args) > 0 {
		params = make([]redshift_types.SqlParameter, len(stmt.Args))
		for i, v := range stmt.Args {
			params[i] = redshift_types.SqlParameter{
				Name:  aws.String(fmt.Sprintf("param%d", i+1)),
				Value: aws.String(fmt.Sprint(v)),
			}
		}
	}

	out, err := c.executeQuery(ctx, client, stmt.Query, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier, params)
	if err != nil {
		return nil, err
	}

	noResult := true
	if out.HasResultSet != nil && *out.HasResultSet {
		noResult = false
	}
	rows, err := newRows(ctx, client, *out.Id, noResult)
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
			Type: redshiftTypeToRuntimeType(typ),
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
	client     *redshiftdata.Client
	results    *redshiftdata.GetStatementResultOutput
	columnMeta []redshift_types.ColumnMetadata
	noResult   bool

	currentRow int
	scannedRow []any
	scanErr    error
}

var _ drivers.Rows = &rows{}

func newRows(ctx context.Context, client *redshiftdata.Client, queryID string, noResult bool) (*rows, error) {
	if noResult {
		return &rows{
			queryID:  queryID,
			client:   client,
			noResult: true,
		}, nil
	}
	results, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{Id: aws.String(queryID)})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var colMeta []redshift_types.ColumnMetadata
	if results.ColumnMetadata != nil {
		colMeta = results.ColumnMetadata
	} else {
		return nil, fmt.Errorf("no column metadata found in query results")
	}

	return &rows{
		queryID:    queryID,
		client:     client,
		results:    results,
		columnMeta: colMeta,
		currentRow: 0,
		scannedRow: make([]any, len(colMeta)),
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

	for i, col := range r.columnMeta {
		dest[*col.Name] = r.scannedRow[i]
	}
	return nil
}

// Next implements drivers.Rows.
func (r *rows) Next() bool {
	if r.noResult {
		return false
	}
	if r.results == nil {
		r.scanErr = sql.ErrConnDone
		return false
	}
	if r.scanErr != nil {
		return false
	}

	// see if we have more rows in current page
	if r.currentRow < len(r.results.Records) {
		record := r.results.Records[r.currentRow]
		var err error
		for i, field := range record {
			if field == nil {
				r.scannedRow[i] = nil
				continue
			}
			r.scannedRow[i], err = convertFieldValue(r.columnMeta[i], field)
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
		r.results = nil
		return false
	}

	nextResults, err := r.client.GetStatementResult(context.Background(), &redshiftdata.GetStatementResultInput{
		Id:        aws.String(r.queryID),
		NextToken: r.results.NextToken,
	})
	if err != nil {
		r.scanErr = fmt.Errorf("failed to get next page of query results: %w", err)
		return false
	}
	r.results = nextResults
	r.currentRow = 0
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
	fields := make([]*runtimev1.StructType_Field, len(r.columnMeta))
	for i, col := range r.columnMeta {
		fields[i] = &runtimev1.StructType_Field{
			Name: *col.Name,
			Type: redshiftTypeToRuntimeType(*col.TypeName),
		}
	}
	res := &runtimev1.StructType{
		Fields: fields,
	}
	return res
}

func redshiftTypeToRuntimeType(colType string) *runtimev1.Type {
	t := &runtimev1.Type{}
	typeLower := strings.ToLower(colType)

	// Handle types with parameters (e.g., "numeric(18,2)", "varchar(255)")
	baseType := typeLower
	if idx := strings.Index(typeLower, "("); idx != -1 {
		baseType = typeLower[:idx]
	}

	switch baseType {
	case "smallint", "int2":
		t.Code = runtimev1.Type_CODE_INT16
	case "integer", "int", "int4":
		t.Code = runtimev1.Type_CODE_INT32
	case "bigint", "int8":
		t.Code = runtimev1.Type_CODE_INT64
	case "real", "float4":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "double precision", "float8", "float":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "boolean", "bool":
		t.Code = runtimev1.Type_CODE_BOOL
	case "date", "time", "timetz", "timestamp", "timestamptz":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "char", "varchar", "text", "bpchar", "nchar", "nvarchar",
		"character varying", "character", "json", "jsonb",
		"numeric", "decimal", "bytea", "super":
		t.Code = runtimev1.Type_CODE_STRING
	default:
		t.Code = runtimev1.Type_CODE_UNSPECIFIED
	}
	return t
}

func convertFieldValue(colMeta redshift_types.ColumnMetadata, field redshift_types.Field) (any, error) {
	// Handle null values
	if field == nil {
		return nil, nil
	}

	typeLower := strings.ToLower(*colMeta.TypeName)
	baseType := typeLower
	if idx := strings.Index(typeLower, "("); idx != -1 {
		baseType = typeLower[:idx]
	}

	// Extract the value from the field based on its type
	switch v := field.(type) {
	case *redshift_types.FieldMemberBooleanValue:
		return v.Value, nil
	case *redshift_types.FieldMemberDoubleValue:
		// For float types, convert to float32 if needed
		if baseType == "real" || baseType == "float4" {
			return float32(v.Value), nil
		}
		return v.Value, nil
	case *redshift_types.FieldMemberLongValue:
		// Convert to appropriate int type based on the column type
		switch baseType {
		case "smallint", "int2":
			return int16(v.Value), nil
		case "integer", "int", "int4":
			return int32(v.Value), nil
		default:
			return v.Value, nil
		}
	case *redshift_types.FieldMemberStringValue:
		return convertStringValue(baseType, v.Value)
	case *redshift_types.FieldMemberBlobValue:
		// Binary data
		return v.Value, nil
	case *redshift_types.FieldMemberIsNull:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported field type: %T", field)
	}
}

func convertStringValue(colType, val string) (any, error) {
	switch colType {
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
	case "timetz":
		t, err := time.Parse("15:04:05-07", val)
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
	case "timestamptz":
		t, err := time.Parse("2006-01-02 15:04:05-07", val)
		if err != nil {
			return nil, err
		}
		return t, nil
	default:
		// For all other types (varchar, char, text, numeric, decimal, json, etc.), return as string
		return val, nil
	}
}
