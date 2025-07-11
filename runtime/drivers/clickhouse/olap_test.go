package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse/testclickhouse"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClickhouseSingle(t *testing.T) {
	dsn := testclickhouse.Start(t)

	conn, err := driver{}.Open("default", map[string]any{"dsn": dsn, "mode": "readwrite"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()
	prepareConn(t, conn)

	c := conn.(*Connection)
	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	t.Run("WithConnection", func(t *testing.T) { testWithConnection(t, olap) })
	t.Run("RenameView", func(t *testing.T) { testRenameView(t, c, olap) })
	t.Run("RenameTable", func(t *testing.T) { testRenameTable(t, c, olap) })
	t.Run("CreateTableAsSelect", func(t *testing.T) { testCreateTableAsSelect(t, c) })
	t.Run("InsertTableAsSelect_WithAppend", func(t *testing.T) { testInsertTableAsSelect_WithAppend(t, c, olap) })
	t.Run("InsertTableAsSelect_WithMerge", func(t *testing.T) { testInsertTableAsSelect_WithMerge(t, c, olap) })
	t.Run("InsertTableAsSelect_WithPartitionOverwrite", func(t *testing.T) { testInsertTableAsSelect_WithPartitionOverwrite(t, c, olap) })
	t.Run("InsertTableAsSelect_WithPartitionOverwrite_DatePartition", func(t *testing.T) { testInsertTableAsSelect_WithPartitionOverwrite_DatePartition(t, c, olap) })
	t.Run("TestDictionary", func(t *testing.T) { testDictionary(t, c, olap) })
	t.Run("TestIntervalType", func(t *testing.T) { testIntervalType(t, olap) })
}

func TestClickhouseCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("clickhouse: skipping test in short mode")
	}

	dsn, cluster := testclickhouse.StartCluster(t)

	conn, err := drivers.Open("clickhouse", "default", map[string]any{"dsn": dsn, "cluster": cluster, "mode": "readwrite"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()

	c := conn.(*Connection)
	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	prepareClusterConn(t, olap, cluster)

	t.Run("WithConnection", func(t *testing.T) { testWithConnection(t, olap) })
	t.Run("RenameView", func(t *testing.T) { testRenameView(t, c, olap) })
	t.Run("RenameTable", func(t *testing.T) { testRenameTable(t, c, olap) })
	t.Run("CreateTableAsSelect", func(t *testing.T) { testCreateTableAsSelect(t, c) })
	t.Run("InsertTableAsSelect_WithAppend", func(t *testing.T) { testInsertTableAsSelect_WithAppend(t, c, olap) })
	t.Run("InsertTableAsSelect_WithMerge", func(t *testing.T) { testInsertTableAsSelect_WithMerge(t, c, olap) })
	t.Run("InsertTableAsSelect_WithPartitionOverwrite", func(t *testing.T) { testInsertTableAsSelect_WithPartitionOverwrite(t, c, olap) })
	t.Run("InsertTableAsSelect_WithPartitionOverwrite_DatePartition", func(t *testing.T) { testInsertTableAsSelect_WithPartitionOverwrite_DatePartition(t, c, olap) })
	t.Run("TestDictionary", func(t *testing.T) { testDictionary(t, c, olap) })
}

func testWithConnection(t *testing.T, olap drivers.OLAPStore) {
	err := olap.WithConnection(context.Background(), 1, func(ctx, ensuredCtx context.Context, conn *sql.Conn) error {
		err := olap.Exec(ctx, &drivers.Statement{
			Query: "CREATE table tbl engine=Memory AS SELECT 1 AS id, 'Earth' AS planet",
		})
		require.NoError(t, err)

		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT id, planet FROM tbl",
		})
		require.NoError(t, err)
		var (
			id     int
			planet string
		)
		for res.Next() {
			err = res.Scan(&id, &planet)
			require.NoError(t, err)
			require.Equal(t, 1, id)
		}
		require.NoError(t, res.Err())
		require.NoError(t, res.Close())
		return nil
	})
	require.NoError(t, err)
}

func testRenameView(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	ctx := context.Background()
	opts := &ModelOutputProperties{Typ: "VIEW"}
	_, err := c.createTableAsSelect(ctx, "foo_view", "SELECT 1 AS id", opts)
	require.NoError(t, err)

	_, err = c.createTableAsSelect(ctx, "bar_view", "SELECT 'city' AS name", opts)
	require.NoError(t, err)

	// rename to unknown view
	err = c.renameEntity(ctx, "foo_view", "foo_view1")
	require.NoError(t, err)

	// rename to existing view
	err = c.renameEntity(ctx, "foo_view1", "bar_view")
	require.NoError(t, err)

	// check that views no longer exist
	notExists(t, olap, "foo_view")
	notExists(t, olap, "foo_view1")

	res, err := olap.Query(ctx, &drivers.Statement{Query: "SELECT id FROM bar_view"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var id int
	require.NoError(t, res.Scan(&id))
	require.Equal(t, 1, id)
	require.NoError(t, res.Close())
}

func testRenameTable(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	ctx := context.Background()
	err := c.renameEntity(ctx, "foo", "foo1")
	require.NoError(t, err)

	err = c.renameEntity(ctx, "foo1", "bar")
	require.NoError(t, err)

	notExists(t, olap, "foo")
	notExists(t, olap, "foo1")
}

func notExists(t *testing.T, olap drivers.OLAPStore, tbl string) {
	result, err := olap.Query(context.Background(), &drivers.Statement{
		Query: "EXISTS " + tbl,
	})
	require.NoError(t, err)
	require.True(t, result.Next())
	var exist bool
	require.NoError(t, result.Scan(&exist))
	require.False(t, exist)
	require.NoError(t, result.Close())
}

func testCreateTableAsSelect(t *testing.T, c *Connection) {
	_, err := c.createTableAsSelect(context.Background(), "tbl", "SELECT 1 AS id, 'Earth' AS planet", &ModelOutputProperties{
		Engine:                 "MergeTree",
		Table:                  "tbl",
		DistributedShardingKey: "rand()",
	})
	require.NoError(t, err)
}

func testInsertTableAsSelect_WithAppend(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	props := &ModelOutputProperties{
		Engine:                 "MergeTree",
		Table:                  "append_tbl",
		DistributedShardingKey: "rand()",
		IncrementalStrategy:    drivers.IncrementalStrategyAppend,
	}

	_, err := c.createTableAsSelect(context.Background(), "append_tbl", "SELECT 1 AS id, 'Earth' AS planet", props)
	require.NoError(t, err)

	insertOpts := &InsertTableOptions{Strategy: drivers.IncrementalStrategyAppend}
	_, err = c.insertTableAsSelect(context.Background(), "append_tbl", "SELECT 2 AS id, 'Mars' AS planet", insertOpts, props)
	require.NoError(t, err)

	res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT id, planet FROM append_tbl ORDER BY id"})
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
	require.NoError(t, err)
	require.NoError(t, res.Close())

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

func testInsertTableAsSelect_WithMerge(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	props := &ModelOutputProperties{
		Typ:                    "TABLE",
		Engine:                 "ReplacingMergeTree",
		Table:                  "tbl",
		DistributedShardingKey: "rand()",
		IncrementalStrategy:    drivers.IncrementalStrategyMerge,
		OrderBy:                "id",
	}
	_, err := c.createTableAsSelect(context.Background(), "merge_tbl", "SELECT generate_series AS id, 'insert' AS value FROM generate_series(0, 4)", props)
	require.NoError(t, err)

	insertOpts := &InsertTableOptions{Strategy: drivers.IncrementalStrategyMerge}
	_, err = c.insertTableAsSelect(context.Background(), "merge_tbl", "SELECT generate_series AS id, 'merge' AS value FROM generate_series(2, 5)", insertOpts, props)
	require.NoError(t, err)

	var result []struct {
		ID    int
		Value string
	}

	res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT id, value FROM merge_tbl ORDER BY id"})
	require.NoError(t, err)

	for res.Next() {
		var r struct {
			ID    int
			Value string
		}
		require.NoError(t, res.Scan(&r.ID, &r.Value))
		result = append(result, r)
	}
	require.NoError(t, err)
	require.NoError(t, res.Close())

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

func testInsertTableAsSelect_WithPartitionOverwrite(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	props := &ModelOutputProperties{
		Engine:                 "MergeTree",
		Table:                  "tbl",
		DistributedShardingKey: "rand()",
		IncrementalStrategy:    drivers.IncrementalStrategyPartitionOverwrite,
		OrderBy:                "id",
		PartitionBy:            "id",
		PrimaryKey:             "id",
	}
	_, err := c.createTableAsSelect(context.Background(), "replace_tbl", "SELECT generate_series AS id, 'insert' AS value FROM generate_series(0, 4)", props)
	require.NoError(t, err)

	insertOpts := &InsertTableOptions{
		Strategy: drivers.IncrementalStrategyPartitionOverwrite,
	}
	_, err = c.insertTableAsSelect(context.Background(), "replace_tbl", "SELECT generate_series AS id, 'replace' AS value FROM generate_series(2, 5)", insertOpts, props)
	require.NoError(t, err)

	res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT id, value FROM replace_tbl ORDER BY id"})
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
	require.NoError(t, err)
	require.NoError(t, res.Close())

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

func testInsertTableAsSelect_WithPartitionOverwrite_DatePartition(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	props := &ModelOutputProperties{
		Engine:                 "MergeTree",
		Table:                  "tbl",
		DistributedShardingKey: "rand()",
		IncrementalStrategy:    drivers.IncrementalStrategyPartitionOverwrite,
		OrderBy:                "dt",
		PartitionBy:            "dt",
		PrimaryKey:             "dt",
	}
	_, err := c.createTableAsSelect(context.Background(), "replace_tbl", "SELECT date_add(hour, generate_series, toDate('2024-12-01')) AS dt, 'insert' AS value FROM generate_series(0, 4)", props)
	require.NoError(t, err)

	insertOpts := &InsertTableOptions{
		Strategy: drivers.IncrementalStrategyPartitionOverwrite,
	}
	_, err = c.insertTableAsSelect(context.Background(), "replace_tbl", "SELECT date_add(hour, generate_series, toDate('2024-12-01')) AS dt, 'replace' AS value FROM generate_series(2, 5)", insertOpts, props)
	require.NoError(t, err)

	res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT dt, value FROM replace_tbl ORDER BY dt"})
	require.NoError(t, err)

	var result []struct {
		DT    string
		Value string
	}

	for res.Next() {
		var r struct {
			DT    string
			Value string
		}
		require.NoError(t, res.Scan(&r.DT, &r.Value))
		result = append(result, r)
	}
	require.NoError(t, err)
	require.NoError(t, res.Close())

	expected := []struct {
		DT    string
		Value string
	}{
		{"2024-12-01T00:00:00Z", "insert"},
		{"2024-12-01T01:00:00Z", "insert"},
		{"2024-12-01T02:00:00Z", "replace"},
		{"2024-12-01T03:00:00Z", "replace"},
		{"2024-12-01T04:00:00Z", "replace"},
	}

	// Convert the result set to a map to represent the set
	resultSet := make(map[string]string)
	for _, r := range result {
		resultSet[r.DT] = r.Value
	}

	// Check if the expected values are present in the result set
	for _, e := range expected {
		value, exists := resultSet[e.DT]
		require.True(t, exists, "Expected DateTime %s to be present in the result set", e.DT)
		require.Equal(t, e.Value, value, "Expected value for DateTime %s to be %s, but got %s", e.DT, e.Value, value)
	}
}

func testDictionary(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	_, err := c.createTableAsSelect(context.Background(), "dict", "SELECT 1 AS id, 'Earth' AS planet", &ModelOutputProperties{
		Typ:                      "DICTIONARY",
		PrimaryKey:               "id",
		DictionarySourceUser:     "clickhouse",
		DictionarySourcePassword: "clickhouse",
	})
	require.NoError(t, err)

	err = c.renameEntity(context.Background(), "dict", "dict1")
	require.NoError(t, err)

	res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT id, planet FROM dict1"})
	require.NoError(t, err)

	require.True(t, res.Next())
	var id int
	var planet string
	require.NoError(t, res.Scan(&id, &planet))
	require.Equal(t, 1, id)
	require.Equal(t, "Earth", planet)
	require.NoError(t, res.Close())

	require.NoError(t, c.dropTable(context.Background(), "dict1"))
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
		rows, err := olap.Query(context.Background(), &drivers.Statement{Query: c.query})
		require.NoError(t, err)
		require.Equal(t, runtimev1.Type_CODE_INTERVAL, rows.Schema.Fields[0].Type.Code)

		require.True(t, rows.Next())
		var s string
		require.NoError(t, rows.Scan(&s))
		ms, ok := ParseIntervalToMillis(s)
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

func TestClickhouseReadWriteMode(t *testing.T) {
	dsn := testclickhouse.Start(t)

	t.Run("ReadOnlyMode_DisablesModelExecution", func(t *testing.T) {
		// Test default mode (read-only) with BYODB
		conn, err := driver{}.Open("default", map[string]any{"dsn": dsn}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		// Should not be able to get model executor in read-only mode
		opts := &drivers.ModelExecutorOptions{
			InputHandle:     conn,
			InputConnector:  "clickhouse",
			OutputHandle:    conn,
			OutputConnector: "clickhouse",
		}
		executor, ok := conn.AsModelExecutor("default", opts)
		require.False(t, ok, "Model executor should not be available in read-only mode")
		require.Nil(t, executor)

		// Should not be able to get model manager in read-only mode
		manager, ok := conn.AsModelManager("default")
		require.False(t, ok, "Model manager should not be available in read-only mode")
		require.Nil(t, manager)
	})

	t.Run("ExplicitReadOnlyMode_DisablesModelExecution", func(t *testing.T) {
		// Test explicit read-only mode
		conn, err := driver{}.Open("default", map[string]any{
			"dsn":  dsn,
			"mode": "read",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		// Should not be able to get model executor
		opts := &drivers.ModelExecutorOptions{
			InputHandle:     conn,
			InputConnector:  "clickhouse",
			OutputHandle:    conn,
			OutputConnector: "clickhouse",
		}
		executor, ok := conn.AsModelExecutor("default", opts)
		require.False(t, ok, "Model executor should not be available in explicit read-only mode")
		require.Nil(t, executor)

		// Should not be able to get model manager
		manager, ok := conn.AsModelManager("default")
		require.False(t, ok, "Model manager should not be available in explicit read-only mode")
		require.Nil(t, manager)
	})

	t.Run("ReadWriteMode_EnablesModelExecution", func(t *testing.T) {
		// Test readwrite mode for BYODB
		conn, err := driver{}.Open("default", map[string]any{
			"dsn":  dsn,
			"mode": "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		// Should be able to get model executor in readwrite mode
		opts := &drivers.ModelExecutorOptions{
			InputHandle:     conn,
			InputConnector:  "clickhouse",
			OutputHandle:    conn,
			OutputConnector: "clickhouse",
		}
		executor, ok := conn.AsModelExecutor("default", opts)
		require.True(t, ok, "Model executor should be available in readwrite mode")
		require.NotNil(t, executor)

		// Should be able to get model manager in readwrite mode
		manager, ok := conn.AsModelManager("default")
		require.True(t, ok, "Model manager should be available in readwrite mode")
		require.NotNil(t, manager)
	})

	t.Run("ManagedMode_EnablesModelExecution", func(t *testing.T) {
		// Test that managed mode automatically enables readwrite
		conn, err := driver{}.Open("default", map[string]any{
			"dsn":     dsn,
			"managed": true,
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		// Should be able to get model executor when managed=true
		opts := &drivers.ModelExecutorOptions{
			InputHandle:     conn,
			InputConnector:  "clickhouse",
			OutputHandle:    conn,
			OutputConnector: "clickhouse",
		}
		executor, ok := conn.AsModelExecutor("default", opts)
		require.True(t, ok, "Model executor should be available when managed=true")
		require.NotNil(t, executor)

		// Should be able to get model manager when managed=true
		manager, ok := conn.AsModelManager("default")
		require.True(t, ok, "Model manager should be available when managed=true")
		require.NotNil(t, manager)
	})

	t.Run("ManagedOverridesMode", func(t *testing.T) {
		// Test that managed=true overrides mode setting
		conn, err := driver{}.Open("default", map[string]any{
			"dsn":     dsn,
			"managed": true,
			"mode":    "read", // This should be overridden
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		// Should still be able to get model executor because managed=true overrides mode
		opts := &drivers.ModelExecutorOptions{
			InputHandle:     conn,
			InputConnector:  "clickhouse",
			OutputHandle:    conn,
			OutputConnector: "clickhouse",
		}
		executor, ok := conn.AsModelExecutor("default", opts)
		require.True(t, ok, "Model executor should be available when managed=true overrides read mode")
		require.NotNil(t, executor)
	})
}
