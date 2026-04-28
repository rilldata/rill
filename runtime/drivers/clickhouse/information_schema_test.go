package clickhouse

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse/testclickhouse"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInformationSchema(t *testing.T) {
	testmode.Expensive(t)
	dsn := testclickhouse.Start(t)
	conn, err := drivers.Open("clickhouse", "", "default", map[string]any{"dsn": dsn, "mode": "readwrite"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	ctx := t.Context()
	olap, ok := conn.AsOLAP("")
	require.True(t, ok)
	infoSchema := olap.InformationSchema()
	require.True(t, ok)

	prepareConn(t, ctx, olap)
	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, ctx, infoSchema) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, ctx, infoSchema) })
	t.Run("testInformationSchemaSystemAllLike", func(t *testing.T) {
		conn, err := drivers.Open("clickhouse", "", "default", map[string]any{"dsn": dsn + "/system", "mode": "readwrite"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		olap, ok := conn.AsOLAP("")
		require.True(t, ok)
		infoSchema := olap.InformationSchema()
		testInformationSchemaSystemAllLike(t, ctx, infoSchema)
	})
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, ctx, infoSchema) })
	t.Run("testInformationSchemaAllPagination", func(t *testing.T) { testInformationSchemaAllPagination(t, ctx, infoSchema) })
	t.Run("testInformationSchemaAllPaginationWithLike", func(t *testing.T) { testInformationSchemaAllPaginationWithLike(t, ctx, infoSchema) })
	t.Run("testInformationSchemaListDatabaseSchemas", func(t *testing.T) { testInformationSchemaListDatabaseSchemas(t, ctx, infoSchema) })
	t.Run("testInformationSchemaListTables", func(t *testing.T) { testInformationSchemaListTables(t, ctx, infoSchema) })
	t.Run("testInformationSchemaListDatabaseSchemasPagination", func(t *testing.T) { testInformationSchemaListDatabaseSchemasPagination(t, ctx, infoSchema) })
	t.Run("testInformationSchemaListTablesPagination", func(t *testing.T) { testInformationSchemaListTablesPagination(t, ctx, infoSchema) })
	t.Run("testLoadDDL", func(t *testing.T) { testLoadDDL(t, ctx, infoSchema) })
}

func testInformationSchemaAll(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	tables, _, err := infoSchema.All(ctx, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 5, len(tables))

	err = infoSchema.LoadPhysicalSize(ctx, tables)
	require.NoError(t, err)

	require.Equal(t, "bar", tables[0].Name)
	require.Equal(t, "foo", tables[1].Name)
	require.Equal(t, "model", tables[2].Name)
	require.Equal(t, "other", tables[3].DatabaseSchema)
	require.Equal(t, "other", tables[4].DatabaseSchema)
	require.Equal(t, "bar", tables[3].Name)
	require.Equal(t, "foo", tables[4].Name)

	require.Equal(t, true, tables[0].IsDefaultDatabaseSchema)
	require.Equal(t, true, tables[1].IsDefaultDatabaseSchema)
	require.Equal(t, true, tables[2].IsDefaultDatabaseSchema)
	require.Equal(t, false, tables[3].IsDefaultDatabaseSchema)
	require.Equal(t, false, tables[4].IsDefaultDatabaseSchema)

	require.Equal(t, 2, len(tables[1].Schema.Fields))
	require.Equal(t, "bar", tables[1].Schema.Fields[0].Name)
	require.Equal(t, runtimev1.Type_CODE_STRING, tables[1].Schema.Fields[0].Type.Code)
	require.Equal(t, "baz", tables[1].Schema.Fields[1].Name)
	require.Equal(t, runtimev1.Type_CODE_INT32, tables[1].Schema.Fields[1].Type.Code)

	require.Equal(t, true, tables[2].View)
	require.Equal(t, int64(0), tables[2].PhysicalSizeBytes)
	require.Greater(t, tables[0].PhysicalSizeBytes, int64(0))
	require.Greater(t, tables[1].PhysicalSizeBytes, int64(0))
}

func testInformationSchemaAllLike(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	tables, _, err := infoSchema.All(ctx, "%odel", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "model", tables[0].Name)

	tables, _, err = infoSchema.All(ctx, "other.%ar", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "bar", tables[0].Name)
}

func testInformationSchemaSystemAllLike(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	tables, _, err := infoSchema.All(ctx, "query_log", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "query_log", tables[0].Name)

	tables, _, err = infoSchema.All(ctx, "other.%ar", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "bar", tables[0].Name)
}

func testInformationSchemaLookup(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	table, err := infoSchema.Lookup(ctx, "", "", "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", table.Name)
	require.Equal(t, true, table.IsDefaultDatabaseSchema)

	_, err = infoSchema.Lookup(ctx, "", "", "bad")
	require.Equal(t, drivers.ErrNotFound, err)

	table, err = infoSchema.Lookup(ctx, "", "", "model")
	require.NoError(t, err)
	require.Equal(t, "model", table.Name)
	require.Equal(t, true, table.IsDefaultDatabaseSchema)

	table, err = infoSchema.Lookup(ctx, "", "other", "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", table.Name)
	require.Equal(t, "other", table.DatabaseSchema)
	require.Equal(t, false, table.IsDefaultDatabaseSchema)
}

func testInformationSchemaAllPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 2

	// Test first page
	tables1, nextToken1, err := infoSchema.All(ctx, "", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := infoSchema.All(ctx, "", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.NotEmpty(t, nextToken2)

	// Test third page
	tables3, nextToken3, err := infoSchema.All(ctx, "", uint32(pageSize), nextToken2)
	require.NoError(t, err)
	require.Equal(t, 1, len(tables3))
	require.Empty(t, nextToken3)

	// Test with page size 0
	tables, nextToken, err := infoSchema.All(ctx, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 5, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = infoSchema.All(ctx, "", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 5, len(tables))
	require.Empty(t, nextToken)
}

func testInformationSchemaAllPaginationWithLike(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 1

	// Test first page
	tables1, nextToken1, err := infoSchema.All(ctx, "%ba%", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := infoSchema.All(ctx, "%ba%", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.Empty(t, nextToken2)

	// Test with page size 0
	tables, nextToken, err := infoSchema.All(ctx, "%ba%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = infoSchema.All(ctx, "%ba%", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)
}

func testInformationSchemaListDatabaseSchemas(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	databaseSchemaInfo, _, err := infoSchema.ListDatabaseSchemas(ctx, 0, "")
	require.NoError(t, err)
	require.Equal(t, 3, len(databaseSchemaInfo))

	require.Equal(t, "", databaseSchemaInfo[0].Database)
	require.Equal(t, "clickhouse", databaseSchemaInfo[0].DatabaseSchema)
	require.Equal(t, "", databaseSchemaInfo[1].Database)
	require.Equal(t, "default", databaseSchemaInfo[1].DatabaseSchema)
	require.Equal(t, "", databaseSchemaInfo[2].Database)
	require.Equal(t, "other", databaseSchemaInfo[2].DatabaseSchema)
}

func testInformationSchemaListTables(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	tables, _, err := infoSchema.ListTables(ctx, "", "default", 0, "")
	require.NoError(t, err)
	require.Equal(t, len(tables), 3)

	require.Equal(t, "bar", tables[0].Name)
	require.Equal(t, false, tables[0].View)
	require.Equal(t, "foo", tables[1].Name)
	require.Equal(t, false, tables[1].View)
	require.Equal(t, "model", tables[2].Name)
	require.Equal(t, true, tables[2].View)

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase)
		require.True(t, tbl.IsDefaultDatabaseSchema)
	}

	tables, _, err = infoSchema.ListTables(ctx, "", "other", 0, "")
	require.NoError(t, err)
	require.Equal(t, len(tables), 2)

	require.Equal(t, "bar", tables[0].Name)
	require.Equal(t, false, tables[0].View)
	require.Equal(t, "foo", tables[1].Name)
	require.Equal(t, false, tables[1].View)

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase)
		require.False(t, tbl.IsDefaultDatabaseSchema)
	}
}

func testInformationSchemaListDatabaseSchemasPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 2

	// First page
	page1, token1, err := infoSchema.ListDatabaseSchemas(ctx, uint32(pageSize), "")
	require.NoError(t, err)
	require.Len(t, page1, pageSize)
	require.NotEmpty(t, token1)

	// second page
	page2, token2, err := infoSchema.ListDatabaseSchemas(ctx, uint32(pageSize), token1)
	require.NoError(t, err)
	require.NotEmpty(t, page2)
	require.Empty(t, token2)

	// Page size 0
	all, token, err := infoSchema.ListDatabaseSchemas(ctx, 0, "")
	require.NoError(t, err)
	require.Equal(t, len(all), 3)
	require.Empty(t, token)
}

func testInformationSchemaListTablesPagination(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	pageSize := 2

	// First page
	page1, token1, err := infoSchema.ListTables(ctx, "", "default", uint32(pageSize), "")
	require.NoError(t, err)
	require.Len(t, page1, pageSize)
	require.NotEmpty(t, token1)

	// Second page
	page2, token2, err := infoSchema.ListTables(ctx, "", "default", uint32(pageSize), token1)
	require.NoError(t, err)
	require.NotEmpty(t, page2)
	require.Empty(t, token2)

	// Page size 0
	all, token, err := infoSchema.ListTables(ctx, "", "default", 0, "")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(all), 3)
	require.Empty(t, token)
}

func testLoadDDL(t *testing.T, ctx context.Context, infoSchema drivers.InformationSchema) {
	// Test DDL for a table
	table, err := infoSchema.Lookup(ctx, "", "", "foo")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, table)
	require.NoError(t, err)
	require.Contains(t, table.DDL, "CREATE TABLE")
	require.Contains(t, table.DDL, "foo")

	// Test DDL for a view
	view, err := infoSchema.Lookup(ctx, "", "", "model")
	require.NoError(t, err)
	err = infoSchema.LoadDDL(ctx, view)
	require.NoError(t, err)
	require.Contains(t, view.DDL, "CREATE VIEW")
	require.Contains(t, view.DDL, "model")
}

func prepareConn(t *testing.T, ctx context.Context, olap drivers.OLAPStore) {
	err := olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE OR REPLACE VIEW model as (select 1, 2, 3)",
	})
	require.NoError(t, err)

	err = olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE TABLE foo(bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()",
	})
	require.NoError(t, err)

	err = olap.Exec(ctx, &drivers.Statement{
		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE TABLE bar(bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()",
	})
	require.NoError(t, err)

	err = olap.Exec(ctx, &drivers.Statement{
		Query: "INSERT INTO bar VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE DATABASE other",
	})
	require.NoError(t, err)

	err = olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE TABLE other.foo(bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()",
	})
	require.NoError(t, err)

	err = olap.Exec(ctx, &drivers.Statement{
		Query: "INSERT INTO other.foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE TABLE other.bar(bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()",
	})
	require.NoError(t, err)

	// test dry run
	err = olap.Exec(ctx, &drivers.Statement{
		DryRun: true,
		Query: `WITH cte_numbers AS
			(
				SELECT num
				FROM generateRandom('num UInt64', NULL)
				LIMIT 10000000000
			)
		SELECT count()
		FROM cte_numbers
		WHERE num IN (
			SELECT num
			FROM cte_numbers
		)`,
	})
	require.NoError(t, err)
}
