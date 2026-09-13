//go:build databricks_kernel

package databricks_test

import (
	"os"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestOLAP_LakehouseRT runs the Rill Databricks OLAP path against a live Lakehouse//RT
// warehouse. RT speaks only SEA, reached via the SEA-via-kernel backend, so this file is
// tagged databricks_kernel (out of the default build and `go test -short` CI). Config
// passes only the DSN (no use_kernel): a passing query proves Thrift->SEA autodetect.
// Run: CGO_ENABLED=1 RILL_RUNTIME_TEST_MODE=expensive RILL_RUNTIME_DATABRICKS_RT_TEST_DSN=...
// go test -tags databricks_kernel -run TestOLAP_LakehouseRT ./runtime/drivers/databricks/...
func TestOLAP_LakehouseRT(t *testing.T) {
	t.Skip("skipping Lakehouse//RT live test; needs a databricks_kernel build and an RT warehouse (set RILL_RUNTIME_DATABRICKS_RT_TEST_DSN)")
	testmode.Expensive(t)

	dsn := os.Getenv("RILL_RUNTIME_DATABRICKS_RT_TEST_DSN")
	if dsn == "" {
		t.Skip("RILL_RUNTIME_DATABRICKS_RT_TEST_DSN not configured")
	}

	_, olap := acquireTestDatabricksRT(t, dsn)

	// Only assert queries RT supports (e.g. current_version() is UNRESOLVED_ROUTINE on RT).
	t.Run("scalar_values", func(t *testing.T) {
		tests := []struct {
			query  string
			result map[string]any
		}{
			{"SELECT TRUE AS bool_val", map[string]any{"bool_val": true}},
			{"SELECT FALSE AS bool_val", map[string]any{"bool_val": false}},
			{"SELECT 'hello' AS string_val", map[string]any{"string_val": "hello"}},
		}
		for _, test := range tests {
			t.Run(test.query, func(t *testing.T) {
				rows, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query})
				require.NoError(t, err)
				defer rows.Close()
				for rows.Next() {
					res := make(map[string]any)
					require.NoError(t, rows.MapScan(res))
					require.Equal(t, test.result, res)
				}
				require.NoError(t, rows.Err())
			})
		}
	})

	t.Run("session_routines", func(t *testing.T) {
		for _, q := range []string{
			"SELECT current_catalog() AS v",
			"SELECT current_schema() AS v",
			"SELECT current_user() AS v",
		} {
			t.Run(q, func(t *testing.T) {
				rows, err := olap.Query(t.Context(), &drivers.Statement{Query: q})
				require.NoError(t, err)
				defer rows.Close()
				require.True(t, rows.Next(), "expected one row")
				res := make(map[string]any)
				require.NoError(t, rows.MapScan(res))
				require.NotEmpty(t, res["v"], "expected a non-empty value from %s", q)
				require.NoError(t, rows.Err())
			})
		}
	})

	t.Run("query_schema", func(t *testing.T) {
		schema, err := olap.QuerySchema(t.Context(), "SELECT 1 AS int_col, 'x' AS str_col", nil)
		require.NoError(t, err)
		require.Len(t, schema.Fields, 2)
		require.Equal(t, "int_col", schema.Fields[0].Name)
		require.Equal(t, "str_col", schema.Fields[1].Name)
	})

	t.Run("dry_run", func(t *testing.T) {
		_, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT 1", DryRun: true})
		require.NoError(t, err)
	})
}

func acquireTestDatabricksRT(t *testing.T, dsn string) (drivers.Handle, drivers.OLAPStore) {
	cfg := map[string]any{"dsn": dsn}
	conn, err := drivers.Open("databricks", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
