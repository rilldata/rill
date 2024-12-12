package clickhouse_test

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClickhouseSingle(t *testing.T) {
	cfg := testruntime.AcquireConnector(t, "clickhouse")

	conn, err := drivers.Open("clickhouse", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()
	prepareConn(t, conn)

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)
	t.Run("RenameView", func(t *testing.T) { testRenameView(t, olap) })
	t.Run("RenameTable", func(t *testing.T) { testRenameTable(t, olap) })
	t.Run("CreateTableAsSelect", func(t *testing.T) { testCreateTableAsSelect(t, olap) })
	t.Run("InsertTableAsSelect_WithAppend", func(t *testing.T) { testInsertTableAsSelect_WithAppend(t, olap) })
	t.Run("InsertTableAsSelect_WithMerge", func(t *testing.T) { testInsertTableAsSelect_WithMerge(t, olap) })
	t.Run("InsertTableAsSelect_WithPartitionOverwrite", func(t *testing.T) { testInsertTableAsSelect_WithPartitionOverwrite(t, olap) })
	t.Run("TestDictionary", func(t *testing.T) { testDictionary(t, olap) })
	t.Run("TestIntervalType", func(t *testing.T) { testIntervalType(t, olap) })
}

func TestClickhouseCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("clickhouse: skipping test in short mode")
	}

	dsn, cluster := testruntime.ClickhouseCluster(t)

	conn, err := drivers.Open("clickhouse", "default", map[string]any{"dsn": dsn, "cluster": cluster}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	prepareClusterConn(t, olap, cluster)

	t.Run("RenameView", func(t *testing.T) { testRenameView(t, olap) })
	t.Run("RenameTable", func(t *testing.T) { testRenameTable(t, olap) })
	t.Run("CreateTableAsSelect", func(t *testing.T) { testCreateTableAsSelect(t, olap) })
	t.Run("InsertTableAsSelect_WithAppend", func(t *testing.T) { testInsertTableAsSelect_WithAppend(t, olap) })
	t.Run("InsertTableAsSelect_WithMerge", func(t *testing.T) { testInsertTableAsSelect_WithMerge(t, olap) })
	t.Run("InsertTableAsSelect_WithPartitionOverwrite", func(t *testing.T) { testInsertTableAsSelect_WithPartitionOverwrite(t, olap) })
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

func testInsertTableAsSelect_WithAppend(t *testing.T, olap drivers.OLAPStore) {
	err := olap.CreateTableAsSelect(context.Background(), "append_tbl", false, "SELECT 1 AS id, 'Earth' AS planet", map[string]any{
		"engine":                   "MergeTree",
		"table":                    "tbl",
		"distributed.sharding_key": "rand()",
		"incremental_strategy":     drivers.IncrementalStrategyAppend,
	})
	require.NoError(t, err)

	err = olap.InsertTableAsSelect(context.Background(), "append_tbl", "SELECT 2 AS id, 'Mars' AS planet", false, true, drivers.IncrementalStrategyAppend, nil)
	require.NoError(t, err)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT id, planet FROM append_tbl ORDER BY id"})
	require.NoError(t, err)

	var result []struct {
		ID     int
		Planet string
	}

	for res.Next() {
		var r struct {
			ID     int
			Planet string
		}
		require.NoError(t, res.Scan(&r.ID, &r.Planet))
		result = append(result, r)
	}

	expected := []struct {
		ID     int
		Planet string
	}{
		{1, "Earth"},
		{2, "Mars"},
	}

	// Convert the result set to a map to represent the set
	resultSet := make(map[int]string)
	for _, r := range result {
		resultSet[r.ID] = r.Planet
	}

	// Check if the expected values are present in the result set
	for _, e := range expected {
		value, exists := resultSet[e.ID]
		require.True(t, exists, "Expected ID %d to be present in the result set", e.ID)
		require.Equal(t, e.Planet, value, "Expected planet for ID %d to be %s, but got %s", e.ID, e.Planet, value)
	}
}

func testInsertTableAsSelect_WithMerge(t *testing.T, olap drivers.OLAPStore) {
	err := olap.CreateTableAsSelect(context.Background(), "merge_tbl", false, "SELECT generate_series AS id, 'insert' AS value FROM generate_series(0, 4)", map[string]any{
		"typs":                     "TABLE",
		"engine":                   "ReplacingMergeTree",
		"table":                    "tbl",
		"distributed.sharding_key": "rand()",
		"incremental_strategy":     drivers.IncrementalStrategyMerge,
		"order_by":                 "id",
		"primary_key":              "id",
		"unique_key":               "id",
	})
	require.NoError(t, err)

	err = olap.InsertTableAsSelect(context.Background(), "merge_tbl", "SELECT generate_series AS id, 'merge' AS value FROM generate_series(2, 5)", false, true, drivers.IncrementalStrategyMerge, []string{"id"})
	require.NoError(t, err)

	var result []struct {
		ID    int
		Value string
	}

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT id, value FROM merge_tbl ORDER BY id"})
	require.NoError(t, err)

	for res.Next() {
		var r struct {
			ID    int
			Value string
		}
		require.NoError(t, res.Scan(&r.ID, &r.Value))
		result = append(result, r)
	}

	expected := map[int]string{
		0: "insert",
		1: "insert",
		2: "merge",
		3: "merge",
		4: "merge",
	}

	// Convert the result set to a map to represent the set
	resultSet := make(map[int]string)
	for _, r := range result {
		if v, ok := resultSet[r.ID]; !ok {
			resultSet[r.ID] = r.Value
		} else {
			if v == "merge" {
				resultSet[r.ID] = v
			}
		}

	}

	// Check if the expected values are present in the result set
	for id, expected := range expected {
		actual, exists := resultSet[id]
		require.True(t, exists, "Expected ID %d to be present in the result set", id)
		require.Equal(t, expected, actual, "Expected value for ID %d to be %s, but got %s", id, expected, actual)
	}
}

func testInsertTableAsSelect_WithPartitionOverwrite(t *testing.T, olap drivers.OLAPStore) {
	err := olap.CreateTableAsSelect(context.Background(), "replace_tbl", false, "SELECT generate_series AS id, 'insert' AS value FROM generate_series(0, 4)", map[string]any{
		"typs":                     "TABLE",
		"engine":                   "MergeTree",
		"table":                    "tbl",
		"distributed.sharding_key": "rand()",
		"incremental_strategy":     drivers.IncrementalStrategyPartitionOverwrite,
		"partition_by":             "id",
		"order_by":                 "value",
		"primary_key":              "value",
	})
	require.NoError(t, err)

	err = olap.InsertTableAsSelect(context.Background(), "replace_tbl", "SELECT generate_series AS id, 'replace' AS value FROM generate_series(2, 5)", false, true, drivers.IncrementalStrategyPartitionOverwrite, nil)
	require.NoError(t, err)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT id, value FROM replace_tbl ORDER BY id"})
	require.NoError(t, err)

	var result []struct {
		ID    int
		Value string
	}

	for res.Next() {
		var r struct {
			ID    int
			Value string
		}
		require.NoError(t, res.Scan(&r.ID, &r.Value))
		result = append(result, r)
	}

	expected := []struct {
		ID    int
		Value string
	}{
		{0, "insert"},
		{1, "insert"},
		{2, "replace"},
		{3, "replace"},
		{4, "replace"},
	}

	// Convert the result set to a map to represent the set
	resultSet := make(map[int]string)
	for _, r := range result {
		resultSet[r.ID] = r.Value
	}

	// Check if the expected values are present in the result set
	for _, e := range expected {
		value, exists := resultSet[e.ID]
		require.True(t, exists, "Expected ID %d to be present in the result set", e.ID)
		require.Equal(t, e.Value, value, "Expected value for ID %d to be %s, but got %s", e.ID, e.Value, value)
	}
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

func testIntervalType(t *testing.T, olap drivers.OLAPStore) {
	cases := []struct {
		query string
		ms    int64
	}{
		{query: "SELECT INTERVAL '1' SECOND", ms: 1000},
		{query: "SELECT INTERVAL '2' MINUTES", ms: 2 * 60 * 1000},
		{query: "SELECT INTERVAL '3' HOURS", ms: 3 * 60 * 60 * 1000},
		{query: "SELECT INTERVAL '4' DAYS", ms: 4 * 24 * 60 * 60 * 1000},
		{query: "SELECT INTERVAL '5' MONTHS", ms: 5 * 30 * 24 * 60 * 60 * 1000},
		{query: "SELECT INTERVAL '6' YEAR", ms: 6 * 365 * 24 * 60 * 60 * 1000},
	}
	for _, c := range cases {
		rows, err := olap.Execute(context.Background(), &drivers.Statement{Query: c.query})
		require.NoError(t, err)
		require.Equal(t, runtimev1.Type_CODE_INTERVAL, rows.Schema.Fields[0].Type.Code)

		require.True(t, rows.Next())
		var s string
		require.NoError(t, rows.Scan(&s))
		ms, ok := clickhouse.ParseIntervalToMillis(s)
		require.True(t, ok)
		require.Equal(t, c.ms, ms)
		require.NoError(t, rows.Close())
	}
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
