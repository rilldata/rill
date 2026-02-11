package duckdb_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/duckdb/duckdb-go/v2"
	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestDuckDBToDuckDBTransfer(t *testing.T) {
	tempDir := t.TempDir()
	dbFile := filepath.Join(tempDir, "transfer.db")
	db, err := sql.Open("duckdb", dbFile)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), "CREATE TABLE foo(bar VARCHAR, baz INTEGER)")
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	duckDB, err := drivers.Open("duckdb", "default", map[string]any{"data_dir": t.TempDir()}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	opts := &drivers.ModelExecutorOptions{
		InputHandle:     duckDB,
		InputConnector:  "duckdb",
		OutputHandle:    duckDB,
		OutputConnector: "duckdb",
		Env: &drivers.ModelEnv{
			AllowHostAccess: false,
			StageChanges:    true,
		},
		PreliminaryInputProperties: map[string]any{
			"sql": "SELECT * FROM foo;",
			"db":  dbFile,
		},
		PreliminaryOutputProperties: map[string]any{
			"table": "sink",
		},
	}

	me, err := duckDB.AsModelExecutor("default", opts)
	require.NoError(t, err)

	execOpts := &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      opts.PreliminaryInputProperties,
		OutputProperties:     opts.PreliminaryOutputProperties,
	}
	_, err = me.Execute(context.Background(), execOpts)
	require.NoError(t, err)

	olap, _ := duckDB.AsOLAP("")
	rows, err := olap.Query(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM sink"})
	require.NoError(t, err)

	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())

	// transfer again
	_, err = me.Execute(context.Background(), execOpts)
	require.NoError(t, err)

	rows, err = olap.Query(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM sink"})
	require.NoError(t, err)

	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())
}

func TestPartitionOverwrite(t *testing.T) {
	files := map[string]string{
		"rill.yaml": "olap_connector: duckdb",
		// Model that creates 10 distinct partitions with 10 rows each.
		// We'll expect the output to have 100 rows.
		"partition_overwrite1.yaml": `
type: model
incremental: true
partitions:
  sql: SELECT range as id, now() as watermark FROM range(0, 10)
partitions_watermark: watermark
sql: SELECT {{.partition.id}} as partition_id, range as num FROM range(10)
output:
  incremental_strategy: partition_overwrite
  partition_by: partition_id
`,
		// Model that creates 10 partitions that are inserted with the same partition_id. Each partition has 10 rows.
		// We'll expect the partitions to keep overwriting each other, so the output has 10 rows.
		"partition_overwrite2.yaml": `
type: model
incremental: true
partitions:
  sql: SELECT range as id, now() as watermark FROM range(0, 10)
partitions_watermark: watermark
sql: SELECT 1 as partition_id, range as num FROM range(10)
output:
  incremental_strategy: partition_overwrite
  partition_by: partition_id
`,
		// Model similar to partition_overwrite1, but testing the implicit default partition overwrite strategy.
		"partition_overwrite3.yaml": `
type: model
incremental: true
partitions:
  sql: SELECT range as id, now() as watermark FROM range(0, 10)
partitions_watermark: watermark
sql: SELECT range as num FROM range(10)
`,
	}

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: files,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Wait a second for the current_timestamp watermark to advance, then refresh the models.
	// This causes all partitions to be re-processed enabling more rigourous testing of partition overwrites.
	time.Sleep(time.Second)
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "partition_overwrite1"})
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "partition_overwrite2"})
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "partition_overwrite3"})

	// partition_overwrite should have 100 rows
	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": `SELECT COUNT(*) AS count, MIN(num) AS min, MAX(num) AS max FROM partition_overwrite1`},
		Result:     []map[string]any{{"count": 100, "min": 0, "max": 9}},
	})

	// partition_overwrite2 should have 10 rows
	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": `SELECT COUNT(*) AS count, MIN(num) AS min, MAX(num) AS max FROM partition_overwrite2`},
		Result:     []map[string]any{{"count": 10, "min": 0, "max": 9}},
	})

	// partition_overwrite3 should have 100 rows and a __rill_partition column
	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": `SELECT COUNT(*) AS count, COUNT(DISTINCT __rill_partition) AS partitions, MIN(num) AS min_num, MAX(num) AS max_num FROM partition_overwrite3`},
		Result:     []map[string]any{{"count": 100, "partitions": 10, "min_num": 0, "max_num": 9}},
	})
}
