package clickhouse_test

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse/testclickhouse"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestClickhouseModels(t *testing.T) {
	dsn := testclickhouse.Start(t)

	t.Run("MaterializeType", func(t *testing.T) { testMaterializeType(t, dsn) })
	t.Run("PartitionOverwrite", func(t *testing.T) { testPartitionOverwrite(t, dsn) })
	t.Run("StagedPostExecRunsAgainstFinalTable", func(t *testing.T) { testStagedPostExecRunsAgainstFinalTable(t, dsn) })
	t.Run("StagedDictionaryRefresh", func(t *testing.T) { testStagedDictionaryRefresh(t, dsn) })
	t.Run("DictionaryModelRename", func(t *testing.T) { testDictionaryModelRename(t, dsn) })
}

// newInstance creates a runtime instance on the shared ClickHouse instance started by TestClickhouseModels.
// Each instance gets its own database so that tests do not observe tables created by the others.
func newInstance(t *testing.T, dsn string, opts testruntime.InstanceOptions) (*runtime.Runtime, string) {
	t.Helper()

	database := nonAlphanumeric.ReplaceAllString(t.Name(), "_")
	conn, err := drivers.Open("clickhouse", "", "default", map[string]any{"dsn": dsn, "mode": "readwrite"}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()
	olap, ok := conn.AsOLAP("")
	require.True(t, ok)
	require.NoError(t, olap.Exec(t.Context(), &drivers.Statement{Query: fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)}))

	opts.Variables = map[string]string{
		"connector.clickhouse.dsn":  fmt.Sprintf("%s/%s", dsn, database),
		"connector.clickhouse.mode": "readwrite",
	}
	return testruntime.NewInstanceWithOptions(t, opts)
}

var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]`)

func testMaterializeType(t *testing.T, dsn string) {
	truth, falsity := true, false
	cases := []struct {
		name        string
		materialize *bool
		typ         string
		wantType    string
		wantErr     bool
	}{
		{"plain", nil, "", "view", false},
		{"materialize-false", &falsity, "", "view", false},
		{"materialize-true", &truth, "", "table", false},
		{"materialize-false-view", &falsity, "view", "view", false},
		{"materialize-true-view", &truth, "view", "", true},
		{"materialize-false-table", &falsity, "table", "", true},
		{"materialize-true-table", &truth, "table", "table", false},
		{"materialize-false-dictionary", &falsity, "dictionary", "", true},
		{"materialize-true-dictionary", &truth, "dictionary", "dictionary", false},
		{"unknown-type", nil, "unknown", "", true},
		{"unknown-type-materialize-false", &falsity, "unknown", "", true},
		{"unknown-type-materialize-true", &truth, "unknown", "", true},
	}

	files := map[string]string{"rill.yaml": "olap_connector: clickhouse\n"}
	for _, c := range cases {
		data := "type: model\nsql: SELECT 1 AS id\n"
		if c.materialize != nil {
			data += fmt.Sprintf("materialize: %v\n", *c.materialize)
		}
		if c.typ != "" {
			data += fmt.Sprintf("output:\n  type: %s\n", c.typ)
			if c.typ == "dictionary" {
				data += "  primary_key: id\n"
			}
		}
		files[fmt.Sprintf("%s.yaml", c.name)] = data
	}

	rt, id := newInstance(t, dsn, testruntime.InstanceOptions{Files: files})
	testruntime.ReconcileParserAndWait(t, rt, id)

	for _, c := range cases {
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, c.name)
		require.NotNil(t, r, c.name)
		if c.wantErr {
			require.NotEmpty(t, r.Meta.ReconcileError, c.name)
		} else {
			require.Empty(t, r.Meta.ReconcileError, c.name)
		}
		if c.wantType != "" {
			resultProps := r.GetModel().State.ResultProperties.AsMap()
			typ := strings.ToLower(resultProps["type"].(string))
			require.Equal(t, c.wantType, typ, c.name)
		}
	}
}

func testPartitionOverwrite(t *testing.T, dsn string) {
	files := map[string]string{
		"rill.yaml": "olap_connector: clickhouse",
		// Model that creates 10 distinct partitions with 10 rows each.
		// We'll expect the output to have 100 rows.
		"partition_overwrite1.yaml": `
type: model
incremental: true
partitions:
  sql: SELECT number as id, now() as watermark FROM numbers(0, 10)
partitions_watermark: watermark
sql: SELECT {{.partition.id}} as partition_id, number as num FROM numbers(10)
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
  sql: SELECT number as id, now() as watermark FROM numbers(0, 10)
partitions_watermark: watermark
sql: SELECT 1 as partition_id, number as num FROM numbers(10)
output:
  incremental_strategy: partition_overwrite
  partition_by: partition_id
`,
		// Model similar to partition_overwrite1, but testing the implicit default partition overwrite strategy.
		"partition_overwrite3.yaml": `
type: model
incremental: true
partitions:
  sql: SELECT number as id, now() as watermark FROM numbers(0, 10)
partitions_watermark: watermark
sql: SELECT number as num FROM numbers(10)
`,
	}

	rt, id := newInstance(t, dsn, testruntime.InstanceOptions{Files: files})
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

func testStagedPostExecRunsAgainstFinalTable(t *testing.T, dsn string) {
	rt, id := newInstance(t, dsn, testruntime.InstanceOptions{
		StageChanges: true,
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"staged_ch.yaml": `
type: model
materialize: true
sql: SELECT 1 AS id UNION ALL SELECT 2 AS id
post_exec: CREATE TABLE staged_ch_marker ENGINE=Memory AS SELECT count() AS c FROM staged_ch
`,
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)

	// The post_exec saw the final table (2 rows), proving it ran after the rename against the final name.
	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": `SELECT c FROM staged_ch_marker`},
		Result:     []map[string]any{{"c": 2}},
	})
}

func testStagedDictionaryRefresh(t *testing.T, dsn string) {
	model := func(sql string) map[string]string {
		return map[string]string{"campaign_name_dict.yaml": fmt.Sprintf(`
type: model
materialize: true
sql: %s
output:
  type: dictionary
  primary_key: id
  dictionary_source_user: default
  dictionary_source_password: default
`, sql)}
	}

	rt, id := newInstance(t, dsn, testruntime.InstanceOptions{
		StageChanges: true,
		Files: map[string]string{
			"rill.yaml":               "olap_connector: clickhouse",
			"campaign_name_dict.yaml": model(`SELECT toUInt64(1) AS id, 'a' AS name`)["campaign_name_dict.yaml"],
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	requireDictionary(t, rt, id, []map[string]any{{"name": "a"}})

	// Refreshing must repoint the dictionary at a fresh source table without leaving the old one behind.
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "campaign_name_dict"})
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	requireDictionary(t, rt, id, []map[string]any{{"name": "a"}})

	// A refresh that changes both the data and the schema must be picked up too.
	testruntime.PutFiles(t, rt, id, model(`SELECT toUInt64(1) AS id, 'b' AS name, 'x' AS extra`))
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	requireDictionary(t, rt, id, []map[string]any{{"name": "b", "extra": "x"}})
}

func testDictionaryModelRename(t *testing.T, dsn string) {
	rt, id := newInstance(t, dsn, testruntime.InstanceOptions{
		StageChanges: true,
		Files: map[string]string{
			"rill.yaml": "olap_connector: clickhouse",
			"dict_a.yaml": `
type: model
materialize: true
sql: SELECT toUInt64(1) AS id, 'a' AS name
output:
  type: dictionary
  primary_key: id
  dictionary_source_user: default
  dictionary_source_password: default
`,
		},
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)

	// Renaming the model renames the dictionary but not the table it sources from, so the source table must not be
	// identified by the dictionary's name, and the rename must not strand it.
	testruntime.RenameFile(t, rt, id, "dict_a.yaml", "dict_b.yaml")
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)

	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": `SELECT name FROM dict_b`},
		Result:     []map[string]any{{"name": "a"}},
	})

	// Refreshing after the rename must still find and drop the old source table rather than accumulating one.
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "dict_b"})
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver: "sql",
		Properties: map[string]any{"sql": `
			SELECT count() AS c FROM system.tables
			WHERE database = currentDatabase() AND position(name, '_dict_temp_') > 0`},
		Result: []map[string]any{{"c": 1}},
	})
}

// requireDictionary asserts the contents of the campaign_name_dict dictionary, and that it is backed by exactly one
// source table with no staging leftovers. ClickHouse forbids dropping a table that a dictionary depends on, so a
// refresh that leaves the dictionary sourcing from the wrong table strands that table permanently.
func requireDictionary(t *testing.T, rt *runtime.Runtime, id string, want []map[string]any) {
	t.Helper()

	cols := make([]string, 0, len(want[0]))
	for col := range want[0] {
		cols = append(cols, col)
	}
	sort.Strings(cols)

	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": fmt.Sprintf("SELECT %s FROM campaign_name_dict", strings.Join(cols, ", "))},
		Result:     want,
	})

	testruntime.RequireResolve(t, rt, id, &testruntime.RequireResolveOptions{
		Resolver: "sql",
		Properties: map[string]any{"sql": `
			SELECT
				countIf(startsWith(name, '__rill_tmp_model_')) AS staged,
				countIf(position(name, '_dict_temp_') > 0) AS sources
			FROM system.tables
			WHERE database = currentDatabase()`},
		Result: []map[string]any{{"staged": 0, "sources": 1}},
	})
}
