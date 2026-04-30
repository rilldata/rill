package snowflake_test

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
	database       = "INTEGRATION_TEST"
	databaseSchema = "PUBLIC"
)

var knownTestTables = []string{"ALL_DATATYPES", "BAR", "BAZ", "FOO", "FOZ", "MODEL"}

const numKnown = 6

func TestInformationSchema(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestSnowflake(t)
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

	require.Equal(t, "ALL_DATATYPES", tables[0].Name)
	require.Equal(t, "BAR", tables[1].Name)
	require.Equal(t, "BAZ", tables[2].Name)
	require.Equal(t, "FOO", tables[3].Name)
	require.Equal(t, "FOZ", tables[4].Name)
	require.Equal(t, "MODEL", tables[5].Name)
	require.True(t, tables[5].View)

	for _, tbl := range tables {
		require.Equal(t, database, tbl.Database, "table %s: expected Database=%s", tbl.Name, database)
		require.Equal(t, databaseSchema, tbl.DatabaseSchema, "table %s: expected DatabaseSchema=%s", tbl.Name, databaseSchema)
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
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

	require.Equal(t, "ALL_DATATYPES", tables[0].Name)
	require.Equal(t, "BAR", tables[1].Name)
	require.Equal(t, "BAZ", tables[2].Name)
	require.Equal(t, "FOO", tables[3].Name)
	require.Equal(t, "FOZ", tables[4].Name)
	require.Equal(t, "MODEL", tables[5].Name)
	require.True(t, tables[5].View)
	for _, tbl := range tables[:5] {
		require.False(t, tbl.View, "table %s should not be a view", tbl.Name)
	}

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=true", tbl.Name)
	}
}

func testLookup(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	bar, err := infoSchema.Lookup(ctx, database, databaseSchema, "BAR")
	require.NoError(t, err)
	require.Equal(t, "BAR", bar.Name)
	require.Equal(t, database, bar.Database)
	require.Equal(t, databaseSchema, bar.DatabaseSchema)
	require.True(t, bar.IsDefaultDatabase)
	require.True(t, bar.IsDefaultDatabaseSchema)
	require.Equal(t, 2, len(bar.Schema.Fields))
	fieldNames := make(map[string]bool)
	for _, f := range bar.Schema.Fields {
		fieldNames[f.Name] = true
	}
	require.True(t, fieldNames["BAR"])
	require.True(t, fieldNames["BAZ"])
	require.False(t, bar.View)

	model, err := infoSchema.Lookup(ctx, database, databaseSchema, "MODEL")
	require.NoError(t, err)
	require.True(t, model.View)

	_, err = infoSchema.Lookup(ctx, database, databaseSchema, "nonexistent_table")
	require.Error(t, err)

	allDatatypes, err := infoSchema.Lookup(ctx, database, databaseSchema, "ALL_DATATYPES")
	require.NoError(t, err)
	require.False(t, allDatatypes.View)

	type fieldSpec struct {
		name     string
		code     runtimev1.Type_Code
		rawType  string
		nullable bool
	}

	expectedFields := []fieldSpec{
		{"ID", runtimev1.Type_CODE_DECIMAL, "NUMBER", true},
		{"BOOLEAN_COL", runtimev1.Type_CODE_BOOL, "BOOLEAN", true},
		{"TINYINT_COL", runtimev1.Type_CODE_DECIMAL, "NUMBER", true},
		{"SMALLINT_COL", runtimev1.Type_CODE_DECIMAL, "NUMBER", true},
		{"INT32_COL", runtimev1.Type_CODE_DECIMAL, "NUMBER", true},
		{"INT64_COL", runtimev1.Type_CODE_DECIMAL, "NUMBER", true},
		{"NUMBER_COL", runtimev1.Type_CODE_DECIMAL, "NUMBER", true},
		{"FLOAT_COL", runtimev1.Type_CODE_FLOAT64, "FLOAT", true},
		{"DOUBLE_COL", runtimev1.Type_CODE_FLOAT64, "FLOAT", true},
		{"DECIMAL_COL", runtimev1.Type_CODE_DECIMAL, "NUMBER", true},
		{"STRING_COL", runtimev1.Type_CODE_STRING, "TEXT", true},
		{"TEXT_COL", runtimev1.Type_CODE_STRING, "TEXT", true},
		{"DATE_COL", runtimev1.Type_CODE_DATE, "DATE", true},
		{"TIME_COL", runtimev1.Type_CODE_TIME, "TIME", true},
		{"TIMESTAMP_NTZ_COL", runtimev1.Type_CODE_TIMESTAMP, "TIMESTAMP_NTZ", true},
		{"TIMESTAMP_LTZ_COL", runtimev1.Type_CODE_TIMESTAMP, "TIMESTAMP_LTZ", true},
		{"TIMESTAMP_TZ_COL", runtimev1.Type_CODE_TIMESTAMP, "TIMESTAMP_TZ", true},
		{"VARIANT_COL", runtimev1.Type_CODE_JSON, "VARIANT", true},
		{"ARRAY_COL", runtimev1.Type_CODE_JSON, "ARRAY", true},
		{"OBJECT_COL", runtimev1.Type_CODE_JSON, "OBJECT", true},
		{"BINARY_COL", runtimev1.Type_CODE_BYTES, "BINARY", true},
		{"GEOGRAPHY_COL", runtimev1.Type_CODE_STRING, "GEOGRAPHY", true},
		{"GEOMETRY_COL", runtimev1.Type_CODE_STRING, "GEOMETRY", true},
	}
	actualFields := make([]fieldSpec, len(allDatatypes.Schema.Fields))
	for i, f := range allDatatypes.Schema.Fields {
		actualFields[i] = fieldSpec{f.Name, f.Type.Code, f.Type.RawType, f.Type.Nullable}
	}
	require.Equal(t, expectedFields, actualFields)
}

func testLoadDDL(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	table, err := infoSchema.Lookup(ctx, database, databaseSchema, "BAR")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, table)
	require.NoError(t, err)
	require.NotEmpty(t, table.DDL)

	view, err := infoSchema.Lookup(ctx, database, databaseSchema, "MODEL")
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
