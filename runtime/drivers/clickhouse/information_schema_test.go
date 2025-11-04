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
	conn, err := drivers.Open("clickhouse", "default", map[string]any{"dsn": dsn, "mode": "readwrite"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	prepareConn(t, conn)
	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, conn) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, conn) })
	t.Run("testInformationSchemaSystemAllLike", func(t *testing.T) {
		conn, err := drivers.Open("clickhouse", "default", map[string]any{"dsn": dsn + "/system", "mode": "readwrite"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		testInformationSchemaSystemAllLike(t, conn)
	})
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, conn) })
	t.Run("testInformationSchemaPagination", func(t *testing.T) { testInformationSchemaAllPagination(t, conn) })
	t.Run("testInformationSchemaPaginationWithLike", func(t *testing.T) { testInformationSchemaAllPaginationWithLike(t, conn) })
}

func testInformationSchemaAll(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("")
	tables, _, err := olap.InformationSchema().All(context.Background(), "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 5, len(tables))

	err = olap.InformationSchema().LoadPhysicalSize(context.Background(), tables)
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

func testInformationSchemaAllLike(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("")
	tables, _, err := olap.InformationSchema().All(context.Background(), "%odel", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "model", tables[0].Name)

	tables, _, err = olap.InformationSchema().All(context.Background(), "other.%ar", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "bar", tables[0].Name)
}

func testInformationSchemaSystemAllLike(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("")

	tables, _, err := olap.InformationSchema().All(context.Background(), "query_log", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "query_log", tables[0].Name)

	tables, _, err = olap.InformationSchema().All(context.Background(), "other.%ar", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "bar", tables[0].Name)
}

func testInformationSchemaLookup(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("")
	ctx := context.Background()
	table, err := olap.InformationSchema().Lookup(ctx, "", "", "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", table.Name)
	require.Equal(t, true, table.IsDefaultDatabaseSchema)

	_, err = olap.InformationSchema().Lookup(ctx, "", "", "bad")
	require.Equal(t, drivers.ErrNotFound, err)

	table, err = olap.InformationSchema().Lookup(ctx, "", "", "model")
	require.NoError(t, err)
	require.Equal(t, "model", table.Name)
	require.Equal(t, true, table.IsDefaultDatabaseSchema)

	table, err = olap.InformationSchema().Lookup(ctx, "", "other", "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", table.Name)
	require.Equal(t, "other", table.DatabaseSchema)
	require.Equal(t, false, table.IsDefaultDatabaseSchema)
}

func testInformationSchemaAllPagination(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("")
	ctx := context.Background()

	pageSize := 2

	// Test first page
	tables1, nextToken1, err := olap.InformationSchema().All(ctx, "", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := olap.InformationSchema().All(ctx, "", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.NotEmpty(t, nextToken2)

	// Test third page
	tables3, nextToken3, err := olap.InformationSchema().All(ctx, "", uint32(pageSize), nextToken2)
	require.NoError(t, err)
	require.Equal(t, 1, len(tables3))
	require.Empty(t, nextToken3)

	// Test with page size 0
	tables, nextToken, err := olap.InformationSchema().All(ctx, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 5, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = olap.InformationSchema().All(ctx, "", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 5, len(tables))
	require.Empty(t, nextToken)
}

func testInformationSchemaAllPaginationWithLike(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("")
	ctx := context.Background()

	pageSize := 1

	// Test first page
	tables1, nextToken1, err := olap.InformationSchema().All(ctx, "%ba%", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := olap.InformationSchema().All(ctx, "%ba%", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.Empty(t, nextToken2)

	// Test with page size 0
	tables, nextToken, err := olap.InformationSchema().All(ctx, "%ba%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = olap.InformationSchema().All(ctx, "%ba%", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)
}

func prepareConn(t *testing.T, conn drivers.Handle) {
	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE OR REPLACE VIEW model as (select 1, 2, 3)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE foo(bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE bar(bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO bar VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE DATABASE other",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE other.foo(bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO other.foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE other.bar(bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()",
	})
	require.NoError(t, err)

	// test dry run
	err = olap.Exec(context.Background(), &drivers.Statement{
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
