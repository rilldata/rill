package clickhouse

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"go.uber.org/zap"
)

func TestInformationSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("clickhouse: skipping test in short mode")
	}

	ctx := context.Background()
	clickHouseContainer, err := clickhouse.RunContainer(ctx,
		testcontainers.WithImage("clickhouse/clickhouse-server:latest"),
		clickhouse.WithUsername("clickhouse"),
		clickhouse.WithPassword("clickhouse"),
		clickhouse.WithConfigFile("../../testruntime/testdata/clickhouse-config.xml"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := clickHouseContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	host, err := clickHouseContainer.Host(ctx)
	require.NoError(t, err)
	port, err := clickHouseContainer.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)

	conn, err := driver{}.Open("default", map[string]any{"dsn": fmt.Sprintf("clickhouse://clickhouse:clickhouse@%v:%v", host, port.Port())}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	prepareConn(t, conn)
	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, conn) })
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, conn) })
}

func testInformationSchemaAll(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("")
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE VIEW model as (select 1, 2, 3)",
	})
	require.NoError(t, err)

	tables, err := olap.InformationSchema().All(context.Background())
	require.NoError(t, err)
	require.Equal(t, 5, len(tables))

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
}

func testInformationSchemaLookup(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("")
	ctx := context.Background()

	err := olap.Exec(ctx, &drivers.Statement{
		Query: "CREATE OR REPLACE VIEW model as (select 1, 2, 3)",
	})
	require.NoError(t, err)

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

func prepareConn(t *testing.T, conn drivers.Handle) {

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	err := olap.Exec(context.Background(), &drivers.Statement{
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
