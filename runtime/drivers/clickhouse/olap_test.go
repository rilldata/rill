package clickhouse_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClickhouseCrudOps(t *testing.T) {
	// t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	if testing.Short() {
		t.Skip("clickhouse: skipping test in short mode")
	}

	dsn, cluster := testruntime.ClickhouseCluster(t)
	t.Run("SingleHost", func(t *testing.T) { testClickhouseSingleHost(t, dsn) })
	t.Run("Cluster", func(t *testing.T) { testClickhouseCluster(t, dsn, cluster) })
}

func testClickhouseSingleHost(t *testing.T, dsn string) {
	conn, err := drivers.Open("clickhouse", "default", map[string]any{"dsn": dsn}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()
	prepareConn(t, conn)

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)
	t.Run("RenameView", func(t *testing.T) {
		testRenameView(t, olap)
	})
	t.Run("RenameTable", func(t *testing.T) { testRenameTable(t, olap) })
	t.Run("CreateTableAsSelect", func(t *testing.T) { testCreateTableAsSelect(t, olap) })
	t.Run("TestDictionary", func(t *testing.T) { testDictionary(t, olap) })

}

func testClickhouseCluster(t *testing.T, dsn, cluster string) {
	conn, err := drivers.Open("clickhouse", "default", map[string]any{"dsn": dsn, "cluster": cluster}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	prepareClusterConn(t, olap, cluster)

	t.Run("RenameView", func(t *testing.T) {
		testRenameView(t, olap)
	})
	t.Run("RenameTable", func(t *testing.T) { testRenameTable(t, olap) })
	t.Run("CreateTableAsSelect", func(t *testing.T) { testCreateTableAsSelect(t, olap) })
	t.Run("TestDictionary", func(t *testing.T) { testDictionary(t, olap) })
}

func testRenameView(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	err := olap.CreateTableAsSelect(ctx, "foo_view", true, "SELECT 1 AS id", map[string]any{"type": "VIEW"})
	require.NoError(t, err)

	err = olap.CreateTableAsSelect(ctx, "bar_view", true, "SELECT 'city' AS name", map[string]any{"type": "VIEW"})
	require.NoError(t, err)

	// rename to unknown view
	err = olap.RenameTable(ctx, "foo_view", "foo_view1", true)
	require.NoError(t, err)

	// rename to existing view
	err = olap.RenameTable(ctx, "foo_view1", "bar_view", true)
	require.NoError(t, err)

	// check that views no longer exist
	notExists(t, olap, "foo_view")
	notExists(t, olap, "foo_view1")

	res, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT id FROM bar_view"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var id int
	require.NoError(t, res.Scan(&id))
	require.Equal(t, 1, id)
}

func testRenameTable(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	err := olap.RenameTable(ctx, "foo", "foo1", false)
	require.NoError(t, err)

	err = olap.RenameTable(ctx, "foo1", "bar", false)
	require.NoError(t, err)

	notExists(t, olap, "foo")
	notExists(t, olap, "foo1")
}

func notExists(t *testing.T, olap drivers.OLAPStore, tbl string) {
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

func testDictionary(t *testing.T, olap drivers.OLAPStore) {
	err := olap.CreateTableAsSelect(context.Background(), "dict", false, "SELECT 1 AS id, 'Earth' AS planet", map[string]any{"table": "Dictionary", "primary_key": "id"})
	require.NoError(t, err)

	err = olap.RenameTable(context.Background(), "dict", "dict1", false)
	require.NoError(t, err)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT id, planet FROM dict1"})
	require.NoError(t, err)

	require.True(t, res.Next())
	var id int
	var planet string
	require.NoError(t, res.Scan(&id, &planet))
	require.Equal(t, 1, id)
	require.Equal(t, "Earth", planet)

	require.NoError(t, olap.DropTable(context.Background(), "dict1", false))
}

func prepareClusterConn(t *testing.T, olap drivers.OLAPStore, cluster string) {
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE OR REPLACE TABLE foo_local ON CLUSTER %s (bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()", cluster),
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE OR REPLACE TABLE foo ON CLUSTER %s AS foo_local engine=Distributed(%s, currentDatabase(), foo_local, rand())", cluster, cluster),
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE OR REPLACE TABLE bar_local ON CLUSTER %s (bar VARCHAR, baz INTEGER) engine=MergeTree ORDER BY tuple()", cluster),
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE OR REPLACE TABLE bar ON CLUSTER %s AS foo_local engine=Distributed(%s, currentDatabase(), foo_local, rand())", cluster, cluster),
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO bar VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)
}
