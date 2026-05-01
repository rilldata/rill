package pinot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/c2h5oh/datasize"
	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.schemaURL+"/databases", http.NoBody)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var databaseSchemas []string
	if err := json.NewDecoder(resp.Body).Decode(&databaseSchemas); err != nil {
		return nil, "", err
	}

	sort.Strings(databaseSchemas)

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	startIndex := 0
	if pageToken != "" {
		startIndex, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}

	endIndex := startIndex + limit
	if endIndex > len(databaseSchemas) {
		endIndex = len(databaseSchemas)
	}
	if startIndex >= len(databaseSchemas) {
		return []*drivers.DatabaseSchemaInfo{}, "", nil
	}

	result := make([]*drivers.DatabaseSchemaInfo, 0, endIndex-startIndex)
	for _, s := range databaseSchemas[startIndex:endIndex] {
		result = append(result, &drivers.DatabaseSchemaInfo{Database: "", DatabaseSchema: s})
	}

	next := ""
	if endIndex < len(databaseSchemas) {
		next = strconv.Itoa(endIndex)
	}
	return result, next, nil
}

func (c *connection) ListTables(ctx context.Context, database, databaseSchema, like string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.schemaURL+"/tables", http.NoBody)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("database", databaseSchema)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tablesResp pinotTables
	err = json.NewDecoder(resp.Body).Decode(&tablesResp)
	if err != nil {
		return nil, "", err
	}

	sort.Strings(tablesResp.Tables)

	if like != "" {
		filtered := tablesResp.Tables[:0]
		for _, name := range tablesResp.Tables {
			if matchLike(name, like) {
				filtered = append(filtered, name)
			}
		}
		tablesResp.Tables = filtered
	}

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	startIndex := 0
	if pageToken != "" {
		startIndex, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}
	endIndex := startIndex + limit
	if endIndex > len(tablesResp.Tables) {
		endIndex = len(tablesResp.Tables)
	}
	if startIndex >= len(tablesResp.Tables) {
		return []*drivers.TableInfo{}, "", nil
	}

	names := tablesResp.Tables[startIndex:endIndex]
	result := make([]*drivers.TableInfo, 0, len(names))
	for _, n := range names {
		result = append(result, &drivers.TableInfo{Database: "", DatabaseSchema: databaseSchema, Name: n, View: false, IsDefaultDatabase: true, IsDefaultDatabaseSchema: true})
	}

	next := ""
	if endIndex < len(tablesResp.Tables) {
		next = strconv.Itoa(endIndex)
	}
	return result, next, nil
}

// All implements drivers.InformationSchema.
func (c *connection) All(ctx context.Context, like string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.schemaURL+"/tables", http.NoBody)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tablesResp pinotTables
	if err := json.NewDecoder(resp.Body).Decode(&tablesResp); err != nil {
		return nil, "", err
	}

	sort.Strings(tablesResp.Tables)

	if like != "" {
		filtered := tablesResp.Tables[:0]
		for _, name := range tablesResp.Tables {
			if matchLike(name, like) {
				filtered = append(filtered, name)
			}
		}
		tablesResp.Tables = filtered
	}

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	startIndex := 0
	if pageToken != "" {
		startIndex, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	}

	endIndex := startIndex + limit
	if endIndex >= len(tablesResp.Tables) {
		endIndex = len(tablesResp.Tables)
	}
	if startIndex >= len(tablesResp.Tables) {
		return []*drivers.TableInfo{}, "", nil
	}

	tables := make([]*drivers.TableInfo, 0, endIndex-startIndex)
	for _, name := range tablesResp.Tables[startIndex:endIndex] {
		tables = append(tables, &drivers.TableInfo{
			Database:       "",
			DatabaseSchema: "",
			Name:           name,
		})
	}

	next := ""
	if endIndex < len(tablesResp.Tables) {
		next = strconv.Itoa(endIndex)
	}

	return tables, next, nil
}

func (c *connection) Lookup(ctx context.Context, database, databaseSchema, table string) (*drivers.TableInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.schemaURL+"/tables/"+table+"/schema", http.NoBody)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, drivers.ErrNotFound
	}
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
		if field.DataType != "TIMESTAMP" && field.DataType != "LONG" {
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

	return &drivers.TableInfo{
		Database:                database,
		DatabaseSchema:          databaseSchema,
		Name:                    table,
		View:                    false,
		IsDefaultDatabase:       true,
		IsDefaultDatabaseSchema: true,
		Schema:                  &runtimev1.StructType{Fields: schemaFields},
		UnsupportedCols:         unsupportedCols,
		PhysicalSizeBytes:       -1,
	}, nil
}

// LoadPhysicalSize populates the PhysicalSizeBytes field of the tables.
// This was not tested when implemented so should be tested when pinot becomes a fairly used connector.
func (c *connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.TableInfo) error {
	if len(tables) == 0 {
		return nil
	}
	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(5)
	for _, table := range tables {
		table := table
		wg.Go(func() error {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.schemaURL+"/debug/tables/"+table.Name+"?type=OFFLINE", http.NoBody)
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				c.logger.Warn("failed to fetch table size", zap.String("table", table.Name), zap.Error(err), observability.ZapCtx(ctx))
				return nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				c.logger.Warn("unexpected status code", zap.String("table", table.Name), zap.Int("status", resp.StatusCode), observability.ZapCtx(ctx))
				return nil
			}

			var data []struct {
				TableSize struct {
					ReportedSize string `json:"reportedSize"`
				} `json:"tableSize"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				c.logger.Warn("failed to decode response", zap.String("table", table.Name), zap.Error(err), observability.ZapCtx(ctx))
				return nil
			}

			var size int64
			for _, d := range data {
				if d.TableSize.ReportedSize != "" && d.TableSize.ReportedSize != "-1 bytes" {
					// Reported size is in bytes
					sz, err := datasize.ParseString(d.TableSize.ReportedSize)
					if err != nil {
						c.logger.Warn("failed to parse reported size", zap.String("table", table.Name), zap.String("size", d.TableSize.ReportedSize), zap.Error(err), observability.ZapCtx(ctx))
						return nil
					}
					size += int64(sz.Bytes())
				}
			}
			table.PhysicalSizeBytes = size
			return nil
		})
	}
	return wg.Wait()
}

// LoadDDL implements drivers.InformationSchema.
func (c *connection) LoadDDL(ctx context.Context, table *drivers.TableInfo) error {
	return nil // Not implemented
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
	t := &runtimev1.Type{Nullable: nullable, RawType: dbt}
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

// matchLike performs a case-insensitive SQL LIKE match where % matches any sequence and _ matches any single character.
func matchLike(name, pattern string) bool {
	name = strings.ToLower(name)
	// Convert SQL LIKE pattern to glob pattern
	var sb strings.Builder
	for _, ch := range strings.ToLower(pattern) {
		switch ch {
		case '%':
			sb.WriteRune('*')
		case '_':
			sb.WriteRune('?')
		case '*', '?', '[', '\\':
			sb.WriteRune('\\')
			sb.WriteRune(ch)
		default:
			sb.WriteRune(ch)
		}
	}
	matched, _ := path.Match(sb.String(), name)
	return matched
}
