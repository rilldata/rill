package resolvers_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProjectStorage(t *testing.T) {
	testmode.Expensive(t)

	// Discover a table from the Druid test cluster to use in the metrics view.
	druidCfg := testruntime.AcquireConnector(t, "druid")
	druidTable, druidDim := discoverDruidTable(t, druidCfg)

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"clickhouse", "druid"},
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
			"models/duckdb_model.sql": "SELECT 1 AS id, 'hello' AS name",
			"models/ch_model.yaml": `
type: model
connector: clickhouse
sql: SELECT 1 AS id, 'hello' AS name
`,
			"dashboards/mv_duckdb.yaml": `
type: metrics_view
model: duckdb_model
dimensions:
  - name: name
    column: name
measures:
  - name: count
    expression: count(*)
`,
			"dashboards/mv_ch.yaml": `
type: metrics_view
connector: clickhouse
model: ch_model
dimensions:
  - name: name
    column: name
measures:
  - name: count
    expression: count(*)
`,
			"dashboards/mv_druid.yaml": fmt.Sprintf(`
type: metrics_view
connector: druid
table: %s
dimensions:
  - name: %s
    column: %s
measures:
  - name: count
    expression: count(*)
`, druidTable, druidDim, druidDim),
		},
	})

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
	}
	entries := make(map[string]entry)
	for {
		row, err := res.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		var e entry
		e.Connector = row["connector"].(string)
		e.Driver = row["driver"].(string)
		e.IsDefaultOLAP = row["is_default_olap"].(bool)
		e.Managed = row["managed"].(bool)
		e.SizeBytes = row["size_bytes"].(int64)
		entries[e.Connector] = e
	}

	require.Len(t, entries, 3)

	// DuckDB: default OLAP, managed, size > 0
	duckdb := entries["duckdb"]
	require.True(t, duckdb.IsDefaultOLAP)
	require.True(t, duckdb.Managed)
	require.Greater(t, duckdb.SizeBytes, int64(0))

	// ClickHouse: not default OLAP, managed, size > 0
	ch := entries["clickhouse"]
	require.False(t, ch.IsDefaultOLAP)
	require.True(t, ch.Managed)
	require.Greater(t, ch.SizeBytes, int64(0))

	// Druid: not default OLAP, not managed, size == -1 (not supported)
	druid := entries["druid"]
	require.False(t, druid.IsDefaultOLAP)
	require.False(t, druid.Managed)
	require.Equal(t, int64(-1), druid.SizeBytes)
}

// discoverDruidTable queries a Druid connector to find an available table and column name for use in tests.
func discoverDruidTable(t *testing.T, cfg map[string]any) (table, dim string) {
	conn, err := drivers.Open("druid", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	// Find the first table in the druid schema.
	rows, err := olap.Query(context.Background(), &drivers.Statement{
		Query: "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'druid' ORDER BY TABLE_NAME LIMIT 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&table))
	require.NotEmpty(t, table)

	// Find the first string column in that table.
	cols, err := olap.Query(context.Background(), &drivers.Statement{
		Query: fmt.Sprintf("SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = 'druid' AND TABLE_NAME = '%s' AND DATA_TYPE = 'VARCHAR' LIMIT 1", table),
	})
	require.NoError(t, err)
	defer cols.Close()

	require.True(t, cols.Next())
	require.NoError(t, cols.Scan(&dim))
	require.NotEmpty(t, dim)

	return table, dim
}
