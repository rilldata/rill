package duckdb_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestInformationSchema(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/all_datatypes.sql": "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/foo.sql":           "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/bar.sql":           "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/foz.sql":           "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/baz.sql":           "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/model.sql":         "SELECT 1,2,3",
		},
	})

	conn, release, err := rt.AcquireHandle(t.Context(), instanceID, "duckdb")
	require.NoError(t, err)
	t.Cleanup(func() { release() })
	olap, ok := conn.AsOLAP("")
	require.True(t, ok)
	infoSchema := olap.InformationSchema()
	require.True(t, ok)
	database := "main_db"
	databaseSchema := "main"
	ctx := t.Context()
	t.Run("testListDatabaseSchemas", func(t *testing.T) {
		testListDatabaseSchemas(t, ctx, infoSchema, database, databaseSchema)
	})
	t.Run("testListTables", func(t *testing.T) { testListTables(t, ctx, infoSchema, database, databaseSchema) })
	t.Run("testListTablesPagination", func(t *testing.T) {
		testListTablesPagination(t, ctx, infoSchema, database, databaseSchema)
	})
	t.Run("testListTablesForAll", func(t *testing.T) { testListTablesForAll(t, ctx, infoSchema) })
	t.Run("testListTablesForAllLike", func(t *testing.T) { testListTablesForAllLike(t, ctx, infoSchema) })
	t.Run("testListTablesForAllPagination", func(t *testing.T) { testListTablesForAllPagination(t, ctx, infoSchema) })
	t.Run("testListTablesForAllPaginationWithLike", func(t *testing.T) { testListTablesForAllPaginationWithLike(t, ctx, infoSchema) })
	t.Run("testLookup", func(t *testing.T) { testLookup(t, ctx, infoSchema, database, databaseSchema) })
	t.Run("testLoadDDL", func(t *testing.T) { testLoadDDL(t, ctx, infoSchema) })
}

func TestInformationSchemaMotherduck(t *testing.T) {
	testmode.Expensive(t)

	cfg := testruntime.AcquireConnector(t, "motherduck")
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"motherduck"},
		Variables: map[string]string{
			"motherduck_token": cfg["token"].(string),
		},
		Files: map[string]string{
			"connectors/motherduck.yaml": `
type: connector
driver: motherduck
token: "{{ .env.motherduck_token }}"
path: md:rilldata
schema_name: integration_test
`,
		},
	})

	conn, release, err := rt.AcquireHandle(t.Context(), instanceID, "motherduck")
	require.NoError(t, err)
	t.Cleanup(func() { release() })
	olap, ok := conn.AsOLAP("")
	require.True(t, ok)
	infoSchema := olap.InformationSchema()
	database := "rilldata"
	databaseSchema := "integration_test"
	ctx := t.Context()
	t.Run("testListDatabaseSchemas", func(t *testing.T) {
		testListDatabaseSchemas(t, ctx, infoSchema, database, databaseSchema)
	})
	t.Run("testListTables", func(t *testing.T) { testListTables(t, ctx, infoSchema, database, databaseSchema) })
	t.Run("testListTablesPagination", func(t *testing.T) {
		testListTablesPagination(t, ctx, infoSchema, database, databaseSchema)
	})
	t.Run("testListTablesForAll", func(t *testing.T) { testListTablesForAll(t, ctx, infoSchema) })
	t.Run("testListTablesForAllLike", func(t *testing.T) { testListTablesForAllLike(t, ctx, infoSchema) })
	t.Run("testListTablesForAllPagination", func(t *testing.T) { testListTablesForAllPagination(t, ctx, infoSchema) })
	t.Run("testListTablesForAllPaginationWithLike", func(t *testing.T) { testListTablesForAllPaginationWithLike(t, ctx, infoSchema) })
	t.Run("testLookup", func(t *testing.T) { testLookup(t, ctx, infoSchema, database, databaseSchema) })

}

func testListDatabaseSchemas(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	databaseSchemaInfo, _, err := infoSchema.ListDatabaseSchemas(ctx, 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(databaseSchemaInfo))

	require.Equal(t, database, databaseSchemaInfo[0].Database)
	require.Equal(t, databaseSchema, databaseSchemaInfo[0].DatabaseSchema)
}

func testListTables(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	tables, _, err := infoSchema.ListTables(ctx, database, databaseSchema, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))

	require.Equal(t, "all_datatypes", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)

	model := tables[5]
	require.Equal(t, true, model.View)

	for _, tbl := range tables {
		require.Equal(t, database, tbl.Database, "table %s: expected Database=%s", tbl.Name, database)
		require.Equal(t, databaseSchema, tbl.DatabaseSchema, "table %s: expected DatabaseSchema=%s", tbl.Name, databaseSchema)
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected Database=%s", tbl.Name, database)
		require.True(t, tbl.IsDefaultDatabaseSchema, "table %s: expected DatabaseSchema=%s", tbl.Name, databaseSchema)
	}
}

func testListTablesPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	pageSize := 2

	// Test first page
	tables1, nextToken1, err := infoSchema.ListTables(ctx, database, databaseSchema, "", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := infoSchema.ListTables(ctx, database, databaseSchema, "", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.NotEmpty(t, nextToken2)

	// Test third page
	tables3, nextToken3, err := infoSchema.ListTables(ctx, database, databaseSchema, "", uint32(pageSize), nextToken2)
	require.NoError(t, err)
	require.Equal(t, 2, len(tables3))
	require.Empty(t, nextToken3)

	// Test with page size 0
	tables, nextToken, err := infoSchema.ListTables(ctx, database, databaseSchema, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = infoSchema.ListTables(ctx, database, databaseSchema, "", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))
	require.Empty(t, nextToken)
}

func testListTablesForAll(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {

	tables, _, err := infoSchema.ListTables(ctx, "", "", "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))

	require.Equal(t, "all_datatypes", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)

	model := tables[5]
	require.Equal(t, true, model.View)
}

func testListTablesForAllLike(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	tables, _, err := infoSchema.ListTables(ctx, "", "", "%odel", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "model", tables[0].Name)

	tables, _, err = infoSchema.ListTables(ctx, "", "", "%model%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "model", tables[0].Name)

	tables, _, err = infoSchema.ListTables(ctx, "", "", "%nonexistent_table%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(tables))
}

func testListTablesForAllPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 2

	// Test first page
	tables1, nextToken1, err := infoSchema.ListTables(ctx, "", "", "", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := infoSchema.ListTables(ctx, "", "", "", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.NotEmpty(t, nextToken2)

	// Test third page
	tables3, nextToken3, err := infoSchema.ListTables(ctx, "", "", "", uint32(pageSize), nextToken2)
	require.NoError(t, err)
	require.Equal(t, 2, len(tables3))
	require.Empty(t, nextToken3)

	// Test with page size 0
	tables, nextToken, err := infoSchema.ListTables(ctx, "", "", "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = infoSchema.ListTables(ctx, "", "", "", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))
	require.Empty(t, nextToken)
}

func testListTablesForAllPaginationWithLike(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 1
	// Test first page
	tables1, nextToken1, err := infoSchema.ListTables(ctx, "", "", "%ba%", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := infoSchema.ListTables(ctx, "", "", "%ba%", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.Empty(t, nextToken2)

	// Test with page size 0
	tables, nextToken, err := infoSchema.ListTables(ctx, "", "", "%ba%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = infoSchema.ListTables(ctx, "", "", "%ba%", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)
}

func testLookup(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema, database, databaseSchema string) {
	bar, err := infoSchema.Lookup(ctx, "", "", "bar")
	require.NoError(t, err)
	require.Equal(t, "bar", bar.Name)
	require.Equal(t, database, bar.Database)
	require.Equal(t, databaseSchema, bar.DatabaseSchema)
	require.True(t, bar.IsDefaultDatabase)
	require.True(t, bar.IsDefaultDatabaseSchema)

	require.Equal(t, 2, len(bar.Schema.Fields))
	require.Equal(t, "bar", bar.Schema.Fields[0].Name)
	require.Equal(t, runtimev1.Type_CODE_STRING, bar.Schema.Fields[0].Type.Code)
	require.Equal(t, "baz", bar.Schema.Fields[1].Name)
	require.Equal(t, runtimev1.Type_CODE_INT32, bar.Schema.Fields[1].Type.Code)
	require.Equal(t, false, bar.View)

	_, err = infoSchema.Lookup(ctx, "", "", "nonexistent_table")
	require.Equal(t, drivers.ErrNotFound, err)

	table, err := infoSchema.Lookup(ctx, "", "", "model")
	require.NoError(t, err)
	require.Equal(t, "model", table.Name)
}

func testLoadDDL(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	// Test DDL for a materialized table
	table, err := infoSchema.Lookup(ctx, "", "", "bar")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, table)
	require.NoError(t, err)
	require.Contains(t, table.DDL, "CREATE TABLE")
	require.Contains(t, table.DDL, "bar")

	// Test DDL for a view
	view, err := infoSchema.Lookup(ctx, "", "", "model")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, view)
	require.NoError(t, err)
	require.Contains(t, view.DDL, "CREATE VIEW")
	require.Contains(t, view.DDL, "model")
}
