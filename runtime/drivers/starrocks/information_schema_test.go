package starrocks

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
	database       = "default_catalog"
	databaseSchema = "test_db"
)

var knownTestTables = []string{"all_types", "bar", "baz", "foo", "foz", "model"}

const numKnown = 6

func TestInformationSchema(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestStarRocks(t)
	ctx := t.Context()
	infoSchema := olap.InformationSchema()

	t.Run("testAll", func(t *testing.T) {
		testAll(t, ctx, infoSchema)
	})
	t.Run("testLookup", func(t *testing.T) {
		testLookup(t, ctx, infoSchema)
	})
	t.Run("testListDatabaseSchemas", func(t *testing.T) {
		testListDatabaseSchemas(t, ctx, infoSchema)
	})
	t.Run("testListTables", func(t *testing.T) {
		testListTables(t, ctx, infoSchema)
	})
	t.Run("testListTablesPagination", func(t *testing.T) {
		testListTablesPagination(t, ctx, infoSchema)
	})
	t.Run("testLoadDDL", func(t *testing.T) {
		testLoadDDL(t, ctx, infoSchema)
	})
}

func testAll(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	all, _, err := infoSchema.All(ctx, "", 0, "")
	require.NoError(t, err)
	tables := filterOLAP(all)
	require.Equal(t, numKnown, len(tables))

	require.Equal(t, "all_types", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)
	require.True(t, tables[5].View)
	for _, tbl := range tables[:5] {
		require.False(t, tbl.View, "table %s should not be a view", tbl.Name)
	}

	for _, tbl := range tables {
		require.False(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=false", tbl.Name)
		require.False(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=false", tbl.Name)
	}
}

func testLookup(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	bar, err := infoSchema.Lookup(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	require.Equal(t, "bar", bar.Name)
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

	allTypes, err := infoSchema.Lookup(ctx, database, databaseSchema, "all_types")
	require.NoError(t, err)
	require.False(t, allTypes.View)

	type fieldSpec struct {
		name     string
		code     runtimev1.Type_Code
		rawType  string
		nullable bool
	}

	expectedFields := []fieldSpec{
		{"id", runtimev1.Type_CODE_INT32, "int", false},
		{"bool_col", runtimev1.Type_CODE_INT8, "tinyint", false},
		{"tinyint_col", runtimev1.Type_CODE_INT8, "tinyint", false},
		{"smallint_col", runtimev1.Type_CODE_INT16, "smallint", false},
		{"int_col", runtimev1.Type_CODE_INT32, "int", false},
		{"bigint_col", runtimev1.Type_CODE_INT64, "bigint", false},
		// largeint_col: LARGEINT is not exposed via information_schema; it is skipped in Lookup.
		{"float_col", runtimev1.Type_CODE_FLOAT32, "float", false},
		{"double_col", runtimev1.Type_CODE_FLOAT64, "double", false},
		{"decimal_col", runtimev1.Type_CODE_STRING, "decimal", false},
		{"char_col", runtimev1.Type_CODE_STRING, "char", false},
		{"varchar_col", runtimev1.Type_CODE_STRING, "varchar", false},
		{"string_col", runtimev1.Type_CODE_STRING, "varchar", false},
		{"date_col", runtimev1.Type_CODE_DATE, "date", false},
		{"datetime_col", runtimev1.Type_CODE_TIMESTAMP, "datetime", false},
		{"json_col", runtimev1.Type_CODE_JSON, "json", false},
		{"array_col", runtimev1.Type_CODE_ARRAY, "array", false},
		{"map_col", runtimev1.Type_CODE_MAP, "map", false},
		{"struct_col", runtimev1.Type_CODE_STRUCT, "struct", false},
	}
	actualFields := make([]fieldSpec, len(allTypes.Schema.Fields))
	for i, f := range allTypes.Schema.Fields {
		actualFields[i] = fieldSpec{f.Name, f.Type.Code, f.Type.RawType, f.Type.Nullable}
	}
	require.Equal(t, expectedFields, actualFields)
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
	all, _, err := infoSchema.ListTables(ctx, database, databaseSchema, 0, "")
	require.NoError(t, err)
	tables := filterTableInfos(all)
	require.Equal(t, numKnown, len(tables))

	require.Equal(t, "all_types", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)
	require.True(t, tables[5].View)

	for _, tbl := range tables {
		require.False(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=false", tbl.Name)
		require.False(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=false", tbl.Name)
	}
}

func testListTablesPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 2

	var collectedAll []*drivers.TableInfo
	var nextToken string
	for {
		tables, token, err := infoSchema.ListTables(ctx, database, databaseSchema, uint32(pageSize), nextToken)
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
	tables, nextToken, err := infoSchema.ListTables(ctx, database, databaseSchema, 0, "")
	require.NoError(t, err)
	require.Equal(t, numKnown, len(filterTableInfos(tables)))
	require.Empty(t, nextToken)
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

func filterOLAP(tables []*drivers.TableInfo) []*drivers.TableInfo {
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
