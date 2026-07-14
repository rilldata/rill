package resolvers_test

import (
	"context"
	"io"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestProjectStorage(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"druid"},
		Files: map[string]string{
			"rill.yaml": "olap_connector: duckdb",
			"connectors/duckdb.yaml": `
type: connector
driver: duckdb
managed: true
`,
			"connectors/clickhouse.yaml": `
type: connector
driver: clickhouse
managed: true
`,
			"connectors/druid.yaml": `
type: connector
driver: druid
dsn: "{{ .env.connector.druid.dsn }}"
`,
			"models/model_duckdb.yaml": `
type: model
materialize: true
sql: SELECT 1 AS id, 'hello' AS name
`,
			"models/model_ch.yaml": `
type: model
materialize: true
connector: clickhouse
sql: SELECT 1 AS id, 'hello' AS name
output:
  connector: clickhouse
`,
			"metrics/mv_duckdb.yaml": `
type: metrics_view
model: model_duckdb
dimensions:
  - name: name
    column: name
measures:
  - name: count
    expression: count(*)
explore:
  skip: true
`,
			"metrics/mv_ch.yaml": `
type: metrics_view
connector: clickhouse
model: model_ch
dimensions:
  - name: name
    column: name
measures:
  - name: count
    expression: count(*)
explore:
  skip: true
`,
			"metrics/mv_druid.yaml": `
type: metrics_view
connector: druid
model: AdBids
dimensions:
  - name: publisher
    column: publisher
measures:
  - name: count
    expression: count(*)
explore:
  skip: true
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 9, 0, 0)

	// Resolve project_storage.
	res, _, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: id,
		Resolver:   "project_storage",
		Claims:     &runtime.SecurityClaims{SkipChecks: true},
	})
	require.NoError(t, err)
	defer res.Close()

	// Collect all rows.
	type entry struct {
		Connector     string `mapstructure:"connector"`
		Driver        string `mapstructure:"driver"`
		IsDefaultOLAP bool   `mapstructure:"is_default_olap"`
		Managed       bool   `mapstructure:"managed"`
		SizeBytes     int64  `mapstructure:"size_bytes"`
		Error         string `mapstructure:"error"`
	}
	entries := make(map[string]entry)
	for {
		row, err := res.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		var e entry
		err = mapstructureutil.WeakDecode(row, &e)
		require.NoError(t, err)
		entries[e.Connector] = e
	}
	require.Len(t, entries, 3)

	// DuckDB: default OLAP, managed, size > 0
	duckdb := entries["duckdb"]
	require.True(t, duckdb.IsDefaultOLAP)
	require.True(t, duckdb.Managed)
	require.Empty(t, duckdb.Error)
	require.Greater(t, duckdb.SizeBytes, int64(0))

	// ClickHouse: not default OLAP, managed, size > 0
	ch := entries["clickhouse"]
	require.False(t, ch.IsDefaultOLAP)
	require.True(t, ch.Managed)
	require.Empty(t, ch.Error)
	require.Greater(t, ch.SizeBytes, int64(0))

	// Druid: not default OLAP, not managed, size == -1 (not supported)
	druid := entries["druid"]
	require.False(t, druid.IsDefaultOLAP)
	require.False(t, druid.Managed)
	require.Empty(t, druid.Error)
	require.Equal(t, int64(-1), druid.SizeBytes)
}
