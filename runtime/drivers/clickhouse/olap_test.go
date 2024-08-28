package clickhouse_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"go.uber.org/zap"
)

func TestClickhouseSingleHost(t *testing.T) {
	if testing.Short() {
		t.Skip("clickhouse: skipping test in short mode")
	}

	ctx := context.Background()
	clickHouseContainer, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:latest",
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

	conn, err := drivers.Open("clickhouse", "default", map[string]any{"dsn": fmt.Sprintf("clickhouse://clickhouse:clickhouse@%v:%v", host, port.Port())}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	prepareConn(t, conn)

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	t.Run("RenameTable", func(t *testing.T) { testRenameTable(t, olap) })
	t.Run("CreateTableAsSelect", func(t *testing.T) { testCreateTableAsSelect(t, olap) })
}

func TestClickhouseCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("clickhouse: skipping test in short mode")
	}

	dsn, cluster := testruntime.ClickhouseCluster(t)

	conn, err := drivers.Open("clickhouse", "default", map[string]any{"dsn": dsn, "cluster": cluster}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	prepareClusterConn(t, olap, cluster)

	t.Run("RenameTable", func(t *testing.T) { testRenameTable(t, olap) })
	t.Run("CreateTableAsSelect", func(t *testing.T) { testCreateTableAsSelect(t, olap) })
}

func testRenameTable(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	err := olap.RenameTable(ctx, "foo", "foo1", false)
	require.NoError(t, err)

	err = olap.RenameTable(ctx, "foo1", "bar", false)
	require.NoError(t, err)

	tableNotExists(t, olap, "foo")
	tableNotExists(t, olap, "foo1")
}

func tableNotExists(t *testing.T, olap drivers.OLAPStore, tbl string) {
	result, err := olap.Execute(context.Background(), &drivers.Statement{
		Query: "EXISTS " + tbl,
	})
	require.NoError(t, err)
	require.True(t, result.Next())
	var exist bool
	require.NoError(t, result.Scan(&exist))
	require.False(t, exist)
}

func testCreateTableAsSelect(t *testing.T, olap drivers.OLAPStore) {
	err := olap.CreateTableAsSelect(context.Background(), "tbl", false, "SELECT 1 AS id, 'Earth' AS planet", map[string]any{
		"engine":                   "MergeTree",
		"table":                    "tbl",
		"distributed.sharding_key": "rand()",
	})
	require.NoError(t, err)
}

func prepareClusterConn(t *testing.T, olap drivers.OLAPStore, cluster string) {
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE TABLE foo_local ON CLUSTER %s (bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()", cluster),
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE TABLE foo ON CLUSTER %s AS foo_local engine=Distributed(%s, currentDatabase(), foo_local, rand())", cluster, cluster),
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE TABLE bar_local ON CLUSTER %s (bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()", cluster),
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE TABLE bar ON CLUSTER %s AS foo_local engine=Distributed(%s, currentDatabase(), foo_local, rand())", cluster, cluster),
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO bar VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)
}
