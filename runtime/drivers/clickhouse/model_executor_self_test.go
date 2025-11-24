package clickhouse_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestMaterializeType(t *testing.T) {
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

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files:          files,
	})
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

func TestPartitionOverwrite(t *testing.T) {
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

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse"},
		Files:          files,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Wait a second for the current_timestamp watermark to advance, then refresh the models.
	// This causes all partitions to be re-processed enabling more rigourous testing of partition overwrites.
	time.Sleep(time.Second)
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "partition_overwrite1"}, nil)
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "partition_overwrite2"}, nil)
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "partition_overwrite3"}, nil)

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
