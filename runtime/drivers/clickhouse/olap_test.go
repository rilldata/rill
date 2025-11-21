package clickhouse

import (
	"context"
	"fmt"
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
	t.Run("TestOptimizeTable", func(t *testing.T) { testOptimizeTable(t, c, olap) })
}

func TestClickhouseCluster(t *testing.T) {
	testmode.Expensive(t)

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
	t.Run("TestOptimizeTable", func(t *testing.T) { testOptimizeTable(t, c, olap) })
}

func testWithConnection(t *testing.T, olap drivers.OLAPStore) {
	err := olap.WithConnection(context.Background(), 1, func(ctx, ensuredCtx context.Context) error {
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
		DictionarySourceUser:     "default",
		DictionarySourcePassword: "default",
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

func testOptimizeTable(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	ctx := context.Background()
	tempTableName := "optimize_basic_test"

	// Create table with MergeTree engine - handle cluster mode
	var err error
	if c.config.Cluster != "" {
		localTableName := tempTableName + "_local"
		localCreateQuery := fmt.Sprintf("CREATE TABLE %s ON CLUSTER %s (id INT, value VARCHAR) ENGINE=MergeTree ORDER BY id", localTableName, c.config.Cluster)
		err = olap.Exec(ctx, &drivers.Statement{Query: localCreateQuery})
		require.NoError(t, err)
	} else {
		createQuery := fmt.Sprintf("CREATE TABLE %s (id INT, value VARCHAR) ENGINE=MergeTree ORDER BY id", tempTableName)
		err = olap.Exec(ctx, &drivers.Statement{Query: createQuery})
		require.NoError(t, err)
	}

	// Insert test data
	err = olap.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf("INSERT INTO %s VALUES (1, 'test1'), (2, 'test2'), (3, 'test3')", tempTableName),
	})
	require.NoError(t, err)

	// Run OPTIMIZE
	err = c.optimizeTable(ctx, tempTableName)
	require.NoError(t, err)

	// Verify data integrity after optimization
	res, err := olap.Query(ctx, &drivers.Statement{
		Query: fmt.Sprintf("SELECT id, value FROM %s ORDER BY id", tempTableName),
	})
	require.NoError(t, err)

	var results []struct {
		ID    int
		Value string
	}
	for res.Next() {
		var r struct {
			ID    int
			Value string
		}
		require.NoError(t, res.Scan(&r.ID, &r.Value))
		results = append(results, r)
	}
	require.NoError(t, res.Close())

	expected := []struct {
		ID    int
		Value string
	}{
		{1, "test1"},
		{2, "test2"},
		{3, "test3"},
	}
	require.Equal(t, expected, results)
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
		executor, err := conn.AsModelExecutor("default", opts)
		require.ErrorContains(t, err, "model execution is disabled")
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
		executor, err := conn.AsModelExecutor("default", opts)
		require.ErrorContains(t, err, "model execution is disabled")
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
		executor, err := conn.AsModelExecutor("default", opts)
		require.NoError(t, err, "Model executor should be available in readwrite mode")
		require.NotNil(t, executor)

		// Should be able to get model manager in readwrite mode
		manager, ok := conn.AsModelManager("default")
		require.True(t, ok, "Model manager should be available in readwrite mode")
		require.NotNil(t, manager)
	})
}

func TestClickhouseDualDSN(t *testing.T) {
	dsn := testclickhouse.Start(t)

	t.Run("SeparateReadWriteDSNs", func(t *testing.T) {
		// Test with both dsn and write_dsn specified
		conn, err := driver{}.Open("default", map[string]any{
			"dsn":       dsn,
			"write_dsn": dsn,
			"mode":      "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		c := conn.(*Connection)
		require.NotNil(t, c.readDB)
		require.NotNil(t, c.writeDB)

		// Verify that separate connections were created
		require.NotEqual(t, c.readDB, c.writeDB, "Read and write connections should be separate when using dual DSNs")

		// Test that both connections work
		olap, ok := conn.AsOLAP("default")
		require.True(t, ok)

		// Test read operation
		res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT 1 AS test"})
		require.NoError(t, err)
		require.True(t, res.Next())
		var testVal int
		require.NoError(t, res.Scan(&testVal))
		require.Equal(t, 1, testVal)
		require.NoError(t, res.Close())

		// Test write operation
		err = olap.Exec(context.Background(), &drivers.Statement{
			Query: "CREATE TABLE dual_dsn_test (id INT) ENGINE=Memory",
		})
		require.NoError(t, err)
	})

	t.Run("OnlyWriteDSN_ShouldFail", func(t *testing.T) {
		// Test that providing only write_dsn fails
		_, err := driver{}.Open("default", map[string]any{
			"write_dsn": dsn,
			"mode":      "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.Error(t, err)
		require.Contains(t, err.Error(), "no clickhouse connection configured")
	})

	t.Run("DualDSNWithRegularDSN_UsesDualDSN", func(t *testing.T) {
		// Test that dsn and write_dsn configuration works correctly
		conn, err := driver{}.Open("default", map[string]any{
			"dsn":       dsn,
			"write_dsn": dsn,
			"mode":      "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		c := conn.(*Connection)
		require.NotEqual(t, c.readDB, c.writeDB, "Should use separate connections when dual DSNs are provided")

		// Verify connection works (would fail if using invalid DSN)
		olap, ok := conn.AsOLAP("default")
		require.True(t, ok)

		res, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT 1"})
		require.NoError(t, err)
		require.NoError(t, res.Close())
	})

	t.Run("InvalidDSN_ShouldFail", func(t *testing.T) {
		// Test that invalid dsn causes failure
		_, err := driver{}.Open("default", map[string]any{
			"dsn":       "invalid-dsn",
			"write_dsn": dsn,
			"mode":      "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse DSN")
	})

	t.Run("InvalidWriteDSN_ShouldFail", func(t *testing.T) {
		// Test that invalid write_dsn causes failure
		_, err := driver{}.Open("default", map[string]any{
			"dsn":       dsn,
			"write_dsn": "invalid-dsn",
			"mode":      "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse write DSN")
	})

	t.Run("SingleDSN_SharedConnection", func(t *testing.T) {
		// Test that single DSN still uses shared connection (backward compatibility)
		conn, err := driver{}.Open("default", map[string]any{
			"dsn":  dsn,
			"mode": "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		c := conn.(*Connection)
		require.NotNil(t, c.readDB)
		require.NotNil(t, c.writeDB)

		// Should use the same connection for both read and write when using single DSN
		require.Equal(t, c.readDB, c.writeDB, "Read and write should share connection when using single DSN")
	})

	t.Run("NoConfiguration_ShouldFail", func(t *testing.T) {
		// Test that providing no valid configuration fails with appropriate error
		_, err := driver{}.Open("default", map[string]any{
			"mode": "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.Error(t, err)
		require.Contains(t, err.Error(), "no clickhouse connection configured")
	})
}

func TestClickhouseDualDSNFunctionality(t *testing.T) {
	dsn := testclickhouse.Start(t)

	t.Run("ReadWriteOperationsWithDualDSN", func(t *testing.T) {
		// Test that both read and write operations work with dual DSN setup
		conn, err := driver{}.Open("default", map[string]any{
			"dsn":       dsn,
			"write_dsn": dsn,
			"mode":      "readwrite",
		}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		defer conn.Close()

		c := conn.(*Connection)
		olap, ok := conn.AsOLAP("default")
		require.True(t, ok)

		// Create a test table using write connection
		err = olap.Exec(context.Background(), &drivers.Statement{
			Query: "CREATE TABLE dual_test (id INT, name VARCHAR) ENGINE=Memory",
		})
		require.NoError(t, err)

		// Insert data using write connection
		err = olap.Exec(context.Background(), &drivers.Statement{
			Query: "INSERT INTO dual_test VALUES (1, 'test1'), (2, 'test2')",
		})
		require.NoError(t, err)

		// Read data using read connection
		res, err := olap.Query(context.Background(), &drivers.Statement{
			Query: "SELECT id, name FROM dual_test ORDER BY id",
		})
		require.NoError(t, err)

		var results []struct {
			ID   int
			Name string
		}
		for res.Next() {
			var r struct {
				ID   int
				Name string
			}
			require.NoError(t, res.Scan(&r.ID, &r.Name))
			results = append(results, r)
		}
		require.NoError(t, res.Close())

		require.Len(t, results, 2)
		require.Equal(t, 1, results[0].ID)
		require.Equal(t, "test1", results[0].Name)
		require.Equal(t, 2, results[1].ID)
		require.Equal(t, "test2", results[1].Name)

		// Test model operations with dual DSN
		testDualDSNModelOperations(t, c, olap)
	})
}

func testDualDSNModelOperations(t *testing.T, c *Connection, olap drivers.OLAPStore) {
	// Test that model operations work correctly with dual DSN setup
	props := &ModelOutputProperties{
		Engine:              "Memory",
		Table:               "dual_dsn_model_test",
		IncrementalStrategy: drivers.IncrementalStrategyAppend,
	}

	// Create table using model operations (should use write connection)
	_, err := c.createTableAsSelect(context.Background(), "dual_dsn_model_test", "SELECT 1 AS id, 'initial' AS status", props)
	require.NoError(t, err)

	// Insert more data using model operations (should use write connection)
	insertOpts := &InsertTableOptions{Strategy: drivers.IncrementalStrategyAppend}
	_, err = c.insertTableAsSelect(context.Background(), "dual_dsn_model_test", "SELECT 2 AS id, 'added' AS status", insertOpts, props)
	require.NoError(t, err)

	// Query the results (should use read connection)
	res, err := olap.Query(context.Background(), &drivers.Statement{
		Query: "SELECT id, status FROM dual_dsn_model_test ORDER BY id",
	})
	require.NoError(t, err)

	var results []struct {
		ID     int
		Status string
	}
	for res.Next() {
		var r struct {
			ID     int
			Status string
		}
		require.NoError(t, res.Scan(&r.ID, &r.Status))
		results = append(results, r)
	}
	require.NoError(t, res.Close())

	require.Len(t, results, 2)
	require.Equal(t, 1, results[0].ID)
	require.Equal(t, "initial", results[0].Status)
	require.Equal(t, 2, results[1].ID)
	require.Equal(t, "added", results[1].Status)
}
