package athena_test

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
	database       = "awsdatacatalog"
	databaseSchema = "integration_test"
)

var knownTestTables = []string{"all_datatypes", "bar", "baz", "foo", "foz", "model"}

const numKnown = 6

func TestInformationSchema(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestAthena(t)
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
	t.Run("testListTablesForAll", func(t *testing.T) {
		testListTablesForAll(t, ctx, infoSchema)
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
		require.False(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=false", tbl.Name)
	}
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
}

func testListTablesForAll(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	all, _, err := infoSchema.ListTables(ctx, "", "", "", 0, "")
	require.NoError(t, err)
	tables := filterOLAP(all)
	require.Equal(t, numKnown, len(tables))

	require.Equal(t, "all_datatypes", tables[0].Name)
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
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.False(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=false", tbl.Name)
	}
}

func testLookup(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	bar, err := infoSchema.Lookup(ctx, database, databaseSchema, "bar")
	require.NoError(t, err)
	require.Equal(t, "bar", bar.Name)
	require.Equal(t, database, bar.Database)
	require.Equal(t, databaseSchema, bar.DatabaseSchema)
	require.True(t, bar.IsDefaultDatabase)
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
		{"id", runtimev1.Type_CODE_INT32, "integer", false},
		{"boolean_col", runtimev1.Type_CODE_BOOL, "boolean", false},
		{"int32_col", runtimev1.Type_CODE_INT32, "integer", false},
		{"int64_col", runtimev1.Type_CODE_INT64, "bigint", false},
		{"float_col", runtimev1.Type_CODE_FLOAT32, "real", false},
		{"double_col", runtimev1.Type_CODE_FLOAT64, "double", false},
		{"byte_array_col", runtimev1.Type_CODE_STRING, "varbinary", false},
		{"fixed_len_byte_array_col", runtimev1.Type_CODE_STRING, "varbinary", false},
		{"string_col", runtimev1.Type_CODE_STRING, "varchar", false},
		{"decimal_col", runtimev1.Type_CODE_STRING, "decimal(10,2)", false},
		{"date_col", runtimev1.Type_CODE_TIMESTAMP, "date", false},
		{"time_millis_col", runtimev1.Type_CODE_INT32, "integer", false},
		{"time_micros_col", runtimev1.Type_CODE_INT64, "bigint", false},
		{"timestamp_millis_col", runtimev1.Type_CODE_TIMESTAMP, "timestamp(3)", false},
		{"timestamp_micros_col", runtimev1.Type_CODE_TIMESTAMP, "timestamp(3)", false},
		{"uuid_col", runtimev1.Type_CODE_STRING, "varchar", false},
		{"list_int_col", runtimev1.Type_CODE_STRING, "array(integer)", false},
		{"list_string_col", runtimev1.Type_CODE_STRING, "array(varchar)", false},
		{"map_col", runtimev1.Type_CODE_STRING, "map(varchar, integer)", false},
		{"struct_col", runtimev1.Type_CODE_STRING, "row(field_int_col integer, field_float_col real, field_string_col varchar)", false},
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
	require.Empty(t, table.DDL)

	view, err := infoSchema.Lookup(ctx, database, databaseSchema, "model")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, view)
	require.NoError(t, err)
	require.Empty(t, view.DDL)
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
