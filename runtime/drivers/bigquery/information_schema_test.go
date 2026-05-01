package bigquery_test

import (
	"context"
	"sort"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

const (
	database       = "rilldata"
	databaseSchema = "integration_test"
)

var knownTestTables = []string{"all_datatypes", "bar", "baz", "foo", "foz", "model"}

const numKnown = 6

func TestInformationSchemaBigQuery(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestBigQuery(t)
	ctx := t.Context()
	infoSchema := olap.InformationSchema()
	t.Run("testListDatabaseSchemas", func(t *testing.T) {
		testListDatabaseSchemas(t, ctx, infoSchema)
	})
	t.Run("testListTables", func(t *testing.T) {
		testListTables(t, ctx, infoSchema)
	})
	t.Run("testListTablesPagination", func(t *testing.T) {
		testListTablesPagination(t, ctx, infoSchema)
	})
	t.Run("testAll", func(t *testing.T) {
		testAll(t, ctx, infoSchema)
	})
	t.Run("testLookup", func(t *testing.T) {
		testLookup(t, ctx, infoSchema)
	})
	t.Run("testLoadDDL", func(t *testing.T) {
		testLoadDDL(t, ctx, infoSchema)
	})
}

func testListDatabaseSchemas(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	schemas, _, err := infoSchema.ListDatabaseSchemas(ctx, 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, schemas)

	found := false
	for _, s := range schemas {
		if s.Database == database && s.DatabaseSchema == databaseSchema {
			found = true
			break
		}
	}
	require.True(t, found, "expected schema %s.%s in ListDatabaseSchemas", database, databaseSchema)
}

func testListTables(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	all, _, err := infoSchema.ListTables(ctx, database, databaseSchema, "", 0, "")
	require.NoError(t, err)
	tables := filterTableInfos(all)
	require.Equal(t, numKnown, len(tables))

	require.Equal(t, "all_datatypes", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)
	require.True(t, tables[5].View)

	for _, tbl := range tables {
		require.Equal(t, database, tbl.Database, "table %s: expected Database=%s", tbl.Name, database)
		require.Equal(t, databaseSchema, tbl.DatabaseSchema, "table %s: expected DatabaseSchema=%s", tbl.Name, databaseSchema)
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		// BigQuery has no default dataset concept
		require.False(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=false", tbl.Name)
	}

	// like filter: %ba% should match bar and baz only
	liked, _, err := infoSchema.ListTables(ctx, database, databaseSchema, "%ba%", 0, "")
	require.NoError(t, err)
	liked = filterTableInfos(liked)
	require.Equal(t, 2, len(liked))
	require.Equal(t, "bar", liked[0].Name)
	require.Equal(t, "baz", liked[1].Name)
}

func testListTablesPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 2

	var collectedAll []*drivers.TableInfo
	var nextToken string
	for {
		tables, token, err := infoSchema.ListTables(ctx, database, databaseSchema, "", uint32(pageSize), nextToken)
		require.NoError(t, err)
		require.LessOrEqual(t, len(tables), pageSize)
		collectedAll = append(collectedAll, tables...)
		nextToken = token
		if token == "" {
			break
		}
	}
	require.Equal(t, numKnown, len(filterTableInfos(collectedAll)))

	// All at once
	tables, nextToken, err := infoSchema.ListTables(ctx, database, databaseSchema, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, numKnown, len(filterTableInfos(tables)))
	require.Empty(t, nextToken)

	// Paginate with like=%ba% (matches bar, baz): collect all pages
	var liked []*drivers.TableInfo
	var likedToken string
	for {
		page, tok, err := infoSchema.ListTables(ctx, database, databaseSchema, "%ba%", 1, likedToken)
		require.NoError(t, err)
		liked = append(liked, page...)
		likedToken = tok
		if tok == "" {
			break
		}
	}
	liked = filterTableInfos(liked)
	require.Equal(t, 2, len(liked))
	require.Equal(t, "bar", liked[0].Name)
	require.Equal(t, "baz", liked[1].Name)
}

func testLookup(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	bar, err := infoSchema.Lookup(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	require.Equal(t, "bar", bar.Name)
	require.Equal(t, database, bar.Database)
	require.Equal(t, databaseSchema, bar.DatabaseSchema)
	require.True(t, bar.IsDefaultDatabase)
	// BigQuery has no default dataset concept
	require.False(t, bar.IsDefaultDatabaseSchema)
	require.Equal(t, 2, len(bar.Schema.Fields))
	fieldNames := make(map[string]bool)
	for _, f := range bar.Schema.Fields {
		fieldNames[f.Name] = true
	}
	require.True(t, fieldNames["bar"])
	require.True(t, fieldNames["baz"])
	require.False(t, bar.View)

	model, err := infoSchema.Lookup(ctx, database, databaseSchema, "model")
	require.NoError(t, err)
	require.True(t, model.View)

	_, err = infoSchema.Lookup(ctx, database, databaseSchema, "nonexistent_table")
	require.Error(t, err)

	allDatatypes, err := infoSchema.Lookup(ctx, database, databaseSchema, "all_datatypes")
	require.NoError(t, err)
	require.False(t, allDatatypes.View)

	type fieldSpec struct {
		name     string
		code     runtimev1.Type_Code
		rawType  string
		nullable bool
	}
	expectedFields := []fieldSpec{
		{"int_col", runtimev1.Type_CODE_INT64, "INTEGER", true},
		{"float_col", runtimev1.Type_CODE_FLOAT64, "FLOAT", true},
		{"numeric_col", runtimev1.Type_CODE_STRING, "NUMERIC", true},
		{"bignumeric_col", runtimev1.Type_CODE_STRING, "BIGNUMERIC", true},
		{"bool_col", runtimev1.Type_CODE_BOOL, "BOOLEAN", true},
		{"string_col", runtimev1.Type_CODE_STRING, "STRING", true},
		{"bytes_col", runtimev1.Type_CODE_BYTES, "BYTES", true},
		{"date_col", runtimev1.Type_CODE_DATE, "DATE", true},
		{"datetime_col", runtimev1.Type_CODE_TIMESTAMP, "DATETIME", true},
		{"time_col", runtimev1.Type_CODE_STRING, "TIME", true},
		{"timestamp_col", runtimev1.Type_CODE_TIMESTAMP, "TIMESTAMP", true},
		{"json_col", runtimev1.Type_CODE_JSON, "JSON", true},
		{"geography_col", runtimev1.Type_CODE_STRING, "GEOGRAPHY", true},
		{"range_date_col", runtimev1.Type_CODE_STRING, "RANGE", true},
		{"range_datetime_col", runtimev1.Type_CODE_STRING, "RANGE", true},
		{"range_timestamp_col", runtimev1.Type_CODE_STRING, "RANGE", true},
		{"array_int_col", runtimev1.Type_CODE_ARRAY, "ARRAY<INTEGER>", true},
		{"array_float_col", runtimev1.Type_CODE_ARRAY, "ARRAY<FLOAT>", true},
		{"array_numeric_col", runtimev1.Type_CODE_ARRAY, "ARRAY<NUMERIC>", true},
		{"array_bignumeric_col", runtimev1.Type_CODE_ARRAY, "ARRAY<BIGNUMERIC>", true},
		{"array_bool_col", runtimev1.Type_CODE_ARRAY, "ARRAY<BOOLEAN>", true},
		{"array_string_col", runtimev1.Type_CODE_ARRAY, "ARRAY<STRING>", true},
		{"array_bytes_col", runtimev1.Type_CODE_ARRAY, "ARRAY<BYTES>", true},
		{"array_date_col", runtimev1.Type_CODE_ARRAY, "ARRAY<DATE>", true},
		{"array_datetime_col", runtimev1.Type_CODE_ARRAY, "ARRAY<DATETIME>", true},
		{"array_time_col", runtimev1.Type_CODE_ARRAY, "ARRAY<TIME>", true},
		{"array_timestamp_col", runtimev1.Type_CODE_ARRAY, "ARRAY<TIMESTAMP>", true},
		{"array_json_col", runtimev1.Type_CODE_ARRAY, "ARRAY<JSON>", true},
		{"array_geography_col", runtimev1.Type_CODE_ARRAY, "ARRAY<GEOGRAPHY>", true},
		{"array_range_date_col", runtimev1.Type_CODE_ARRAY, "ARRAY<RANGE>", true},
		{"array_range_datetime_col", runtimev1.Type_CODE_ARRAY, "ARRAY<RANGE>", true},
		{"array_range_timestamp_col", runtimev1.Type_CODE_ARRAY, "ARRAY<RANGE>", true},
		{"array_struct_col", runtimev1.Type_CODE_ARRAY, "ARRAY<RECORD>", true},
		{"struct_col", runtimev1.Type_CODE_JSON, "RECORD", true},
	}
	actualFields := make([]fieldSpec, len(allDatatypes.Schema.Fields))
	for i, f := range allDatatypes.Schema.Fields {
		actualFields[i] = fieldSpec{f.Name, f.Type.Code, f.Type.RawType, f.Type.Nullable}
	}
	require.Equal(t, expectedFields, actualFields)
}

func testLoadDDL(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	table, err := infoSchema.Lookup(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, table)
	require.NoError(t, err)
	require.NotEmpty(t, table.DDL)

	view, err := infoSchema.Lookup(ctx, database, databaseSchema, "model")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, view)
	require.NoError(t, err)
	require.NotEmpty(t, view.DDL)
}

func testAll(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	all, _, err := infoSchema.All(ctx, "", 0, "")
	require.NoError(t, err)
	tables := filterTableInfos(all)
	require.Equal(t, numKnown, len(tables))

	require.Equal(t, "all_datatypes", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)
	require.True(t, tables[5].View)

	for _, tbl := range tables {
		require.Equal(t, database, tbl.Database, "table %s: expected Database=%s", tbl.Name, database)
		require.Equal(t, databaseSchema, tbl.DatabaseSchema, "table %s: expected DatabaseSchema=%s", tbl.Name, databaseSchema)
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		// BigQuery has no default dataset concept
		require.False(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=false", tbl.Name)
	}

	// like filter: %ba% should match bar and baz only
	liked, _, err := infoSchema.All(ctx, "%ba%", 0, "")
	require.NoError(t, err)
	liked = filterTableInfos(liked)
	require.Equal(t, 2, len(liked))
	require.Equal(t, "bar", liked[0].Name)
	require.Equal(t, "baz", liked[1].Name)
}

func filterTableInfos(tables []*drivers.TableInfo) []*drivers.TableInfo {
	known := make(map[string]bool, numKnown)
	for _, n := range knownTestTables {
		known[n] = true
	}
	var out []*drivers.TableInfo
	for _, t := range tables {
		if known[t.Name] {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
