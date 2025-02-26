package pinot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/c2h5oh/datasize"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var _ drivers.OLAPStore = &connection{}

// AddTableColumn implements drivers.OLAPStore.
func (c *connection) AddTableColumn(ctx context.Context, tableName, columnName, typ string) error {
	return fmt.Errorf("pinot: data transformation not yet supported")
}

// AlterTableColumn implements drivers.OLAPStore.
func (c *connection) AlterTableColumn(ctx context.Context, tableName, columnName, newType string) error {
	return fmt.Errorf("pinot: data transformation not yet supported")
}

// CreateTableAsSelect implements drivers.OLAPStore.
func (c *connection) CreateTableAsSelect(ctx context.Context, name, sql string, opts *drivers.CreateTableOptions) error {
	return fmt.Errorf("pinot: data transformation not yet supported")
}

// DropTable implements drivers.OLAPStore.
func (c *connection) DropTable(ctx context.Context, name string) error {
	return fmt.Errorf("pinot: data transformation not yet supported")
}

// InsertTableAsSelect implements drivers.OLAPStore.
func (c *connection) InsertTableAsSelect(ctx context.Context, name, sql string, opts *drivers.InsertTableOptions) error {
	return fmt.Errorf("pinot: data transformation not yet supported")
}

// RenameTable implements drivers.OLAPStore.
func (c *connection) RenameTable(ctx context.Context, name, newName string) error {
	return fmt.Errorf("pinot: data transformation not yet supported")
}

func (c *connection) Dialect() drivers.Dialect {
	return drivers.DialectPinot
}

func (c *connection) WithConnection(ctx context.Context, priority int, longRunning bool, fn drivers.WithConnectionFunc) error {
	return fmt.Errorf("pinot: WithConnection not supported")
}

func (c *connection) Exec(ctx context.Context, stmt *drivers.Statement) error {
	res, err := c.Execute(ctx, stmt)
	if err != nil {
		return err
	}
	if stmt.DryRun {
		return nil
	}
	return res.Close()
}

func (c *connection) Execute(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	if c.logQueries {
		c.logger.Info("pinot query", zap.String("sql", stmt.Query), zap.Any("args", stmt.Args), observability.ZapCtx(ctx))
	}
	if stmt.DryRun {
		rows, err := c.db.QueryxContext(ctx, "EXPLAIN PLAN FOR "+stmt.Query, stmt.Args...)
		if err != nil {
			return nil, err
		}

		return nil, rows.Close()
	}

	var cancelFunc context.CancelFunc
	if stmt.ExecutionTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, stmt.ExecutionTimeout)
	}

	rows, err := c.db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		if cancelFunc != nil {
			cancelFunc()
		}
		return nil, err
	}

	schema, err := rowsToSchema(rows)
	if err != nil {
		rows.Close()
		if cancelFunc != nil {
			cancelFunc()
		}
		return nil, err
	}

	r := &drivers.Result{Rows: rows, Schema: schema}
	r.SetCleanupFunc(func() error {
		if cancelFunc != nil {
			cancelFunc()
		}
		return nil
	})

	return r, nil
}

func (c *connection) MayBeScaledToZero(ctx context.Context) bool {
	return false
}

type informationSchema struct {
	c *connection
}

func (c *connection) InformationSchema() drivers.InformationSchema {
	return informationSchema{c: c}
}

func (i informationSchema) All(ctx context.Context, like string, includeSize bool) ([]*drivers.Table, error) {
	// query /tables endpoint, for each table name, query /tables/{tableName}/schema
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, i.c.schemaURL+"/tables", http.NoBody)
	for k, v := range i.c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tablesResp pinotTables
	err = json.NewDecoder(resp.Body).Decode(&tablesResp)
	if err != nil {
		return nil, err
	}

	// Poor man's conversion of a SQL ILIKE pattern to a Go regexp.
	var likeRegexp *regexp.Regexp
	if like != "" {
		likeRegexp, err = regexp.Compile(fmt.Sprintf("(?i)^%s$", strings.ReplaceAll(like, "%", ".*")))
		if err != nil {
			return nil, fmt.Errorf("failed to convert like pattern to regexp: %w", err)
		}
	}

	tables := make([]*drivers.Table, 0, len(tablesResp.Tables))
	// fetch table schemas in parallel with concurrency of 5
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(5)
	for _, tableName := range tablesResp.Tables {
		if likeRegexp != nil && !likeRegexp.MatchString(tableName) {
			continue
		}

		tableName := tableName
		g.Go(func() error {
			table, err := i.Lookup(gctx, "", "", tableName)
			if err != nil {
				fmt.Printf("Error fetching schema for table %s: %v\n", tableName, err)
				return nil
			}
			tables = append(tables, table)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	g, gctx = errgroup.WithContext(ctx)
	g.SetLimit(5)
	if includeSize {
		for j := range tables {
			g.Go(func() error {
				table := tables[j]
				size, err := i.TableSize(gctx, table.Name)
				if err != nil {
					fmt.Printf("Error fetching size for table %s: %v\n", table.Name, err)
					return nil
				}
				table.BytesOnDisk = size
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return tables, nil
}

func (i informationSchema) Lookup(ctx context.Context, db, schema, name string) (*drivers.Table, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, i.c.schemaURL+"/tables/"+name+"/schema", http.NoBody)
	for k, v := range i.c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var schemaResponse pinotSchema
	err = json.NewDecoder(resp.Body).Decode(&schemaResponse)
	if err != nil {
		return nil, err
	}

	unsupportedCols := make(map[string]string)
	var schemaFields []*runtimev1.StructType_Field
	for _, field := range schemaResponse.DateTimeFieldSpecs {
		if field.DataType != "TIMESTAMP" {
			unsupportedCols[field.Name] = field.DataType + "_(DATE_TIME_FIELD)"
			continue
		}
		schemaFields = append(schemaFields, &runtimev1.StructType_Field{Name: field.Name, Type: databaseTypeToPB(field.DataType, !field.NotNull, true)})
	}
	for _, field := range schemaResponse.DimensionFieldSpecs {
		singleValueField := true
		if field.SingleValueField != nil {
			singleValueField = *field.SingleValueField
		}
		if !singleValueField {
			// Skip array fields for now
			unsupportedCols[field.Name] = field.DataType + "_ARRAY"
			continue
		}
		schemaFields = append(schemaFields, &runtimev1.StructType_Field{Name: field.Name, Type: databaseTypeToPB(field.DataType, !field.NotNull, singleValueField)})
	}
	for _, field := range schemaResponse.MetricFieldSpecs {
		singleValueField := true
		if field.SingleValueField != nil {
			singleValueField = *field.SingleValueField
		}
		if !singleValueField {
			// Skip array fields for now
			unsupportedCols[field.Name] = field.DataType + "_ARRAY"
			continue
		}
		schemaFields = append(schemaFields, &runtimev1.StructType_Field{Name: field.Name, Type: databaseTypeToPB(field.DataType, !field.NotNull, singleValueField)})
	}

	table := &drivers.Table{
		Database:        "",
		DatabaseSchema:  "",
		Name:            name,
		View:            false,
		Schema:          &runtimev1.StructType{Fields: schemaFields},
		UnsupportedCols: unsupportedCols,
		BytesOnDisk:     -1,
	}

	return table, nil
}

func (i informationSchema) TableSize(ctx context.Context, name string) (int64, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, i.c.schemaURL+"/debug/tables/"+name+"?type=OFFLINE", http.NoBody)
	for k, v := range i.c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data []struct {
		TableSize struct {
			ReportedSize string `json:"reportedSize"`
		} `json:"tableSize"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	var size int64
	for _, d := range data {
		if d.TableSize.ReportedSize != "" && d.TableSize.ReportedSize != "-1" {
			// Reported size is in bytes
			sz, err := datasize.ParseString(d.TableSize.ReportedSize)
			if err != nil {
				return 0, err
			}
			size += int64(sz.Bytes())
		}
	}
	return size, nil
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: databaseTypeToPB(ct.DatabaseTypeName(), nullable, true),
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

func databaseTypeToPB(dbt string, nullable, singleValueField bool) *runtimev1.Type {
	t := &runtimev1.Type{Nullable: nullable}
	if !singleValueField {
		// currently we don't support array fields, so unreachable code
		t.Code = runtimev1.Type_CODE_ARRAY
		t.ArrayElementType = databaseTypeToPB(dbt, false, true)
		return t
	}
	switch dbt {
	case "INT":
		t.Code = runtimev1.Type_CODE_INT32
	case "LONG":
		t.Code = runtimev1.Type_CODE_INT64
	case "FLOAT":
		t.Code = runtimev1.Type_CODE_FLOAT32
	case "DOUBLE":
		t.Code = runtimev1.Type_CODE_FLOAT64
	case "BIG_DECIMAL":
		t.Code = runtimev1.Type_CODE_STRING
	case "BOOLEAN":
		t.Code = runtimev1.Type_CODE_BOOL
	case "STRING":
		t.Code = runtimev1.Type_CODE_STRING
	case "TIMESTAMP":
		t.Code = runtimev1.Type_CODE_TIMESTAMP
	case "JSON":
		t.Code = runtimev1.Type_CODE_JSON
	case "BYTES":
		t.Code = runtimev1.Type_CODE_BYTES
	default:
		t.Code = runtimev1.Type_CODE_STRING
	}

	return t
}

type pinotTables struct {
	Tables []string `json:"tables"`
}

type pinotSchema struct {
	SchemaName                    string           `json:"schemaName"`
	EnableColumnBasedNullHandling bool             `json:"enableColumnBasedNullHandling"`
	DimensionFieldSpecs           []pinotFieldSpec `json:"dimensionFieldSpecs"`
	MetricFieldSpecs              []pinotFieldSpec `json:"metricFieldSpecs"`
	DateTimeFieldSpecs            []pinotFieldSpec `json:"dateTimeFieldSpecs"`
}

type pinotFieldSpec struct {
	Name             string      `json:"name"`
	DataType         string      `json:"dataType"`
	SingleValueField *bool       `json:"singleValueField"`
	NotNull          bool        `json:"notNull"`
	DefaultNullValue interface{} `json:"defaultNullValue"`
	Format           string      `json:"format"`      // only for timeFieldSpec
	Granularity      string      `json:"granularity"` // only for timeFieldSpec
}
