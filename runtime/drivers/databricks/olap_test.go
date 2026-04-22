package databricks_test

import (
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOLAP(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	tests := []struct {
		query  string
		args   []any
		result map[string]any
	}{
		{
			"SELECT TRUE AS bool_val",
			nil,
			map[string]any{"bool_val": true},
		},
		{
			"SELECT FALSE AS bool_val",
			nil,
			map[string]any{"bool_val": false},
		},
		{
			"SELECT CAST('2021-01-01' AS DATE) AS date_val",
			nil,
			map[string]any{"date_val": time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			"SELECT CAST(3.14 AS DOUBLE) AS float_val",
			nil,
			map[string]any{"float_val": 3.14},
		},
		{
			"SELECT 'hello' AS string_val",
			nil,
			map[string]any{"string_val": "hello"},
		},
		{
			"SELECT double_col FROM integration_test.all_datatypes WHERE int32_col = 2147483647",
			nil,
			map[string]any{"double_col": 2.718},
		},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			rows, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query, Args: test.args})
			require.NoError(t, err)
			defer rows.Close()
			for rows.Next() {
				res := make(map[string]any)
				err = rows.MapScan(res)
				require.NoError(t, err)
				require.Equal(t, test.result, res)
			}
			require.NoError(t, rows.Err())
		})
	}
}

func TestComplexTypes(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)

	// Test non-null complex types (id=1)
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT array_col, map_col, struct_col FROM integration_test.all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	require.Equal(t, "[1,2,3]", res["array_col"])
	require.Equal(t, `{"city":"New York"}`, res["map_col"])
	require.Equal(t, `{"city":"New York","zip":10001}`, res["struct_col"])

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())

	// Test null complex types (id=3)
	rows2, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT array_col, map_col, struct_col FROM integration_test.all_datatypes WHERE id = 3",
	})
	require.NoError(t, err)
	defer rows2.Close()

	require.True(t, rows2.Next())
	res2 := make(map[string]any)
	err = rows2.MapScan(res2)
	require.NoError(t, err)

	require.Nil(t, res2["array_col"])
	require.Nil(t, res2["map_col"])
	// Databricks expands NULL structs into their fields with null values
	require.Equal(t, `{"city":null,"zip":null}`, res2["struct_col"])

	require.False(t, rows2.Next())
	require.NoError(t, rows2.Err())
}

func TestEmptyRows(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT int32_col, double_col FROM integration_test.all_datatypes LIMIT 0"})
	require.NoError(t, err)
	defer rows.Close()

	sc := rows.Schema
	require.Len(t, sc.Fields, 2)
	require.Equal(t, "int32_col", sc.Fields[0].Name)
	require.Equal(t, "double_col", sc.Fields[1].Name)
	require.False(t, rows.Next())
	require.Nil(t, rows.Err())
}

func TestLoadDDL(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	table, err := olap.InformationSchema().Lookup(t.Context(), "", "integration_test", "all_datatypes")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(t.Context(), table)
	require.NoError(t, err)
	require.Contains(t, strings.ToUpper(table.DDL), "ALL_DATATYPES")
}

func TestDryRun(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	_, err := olap.Query(t.Context(), &drivers.Statement{
		Query:  "SELECT * FROM integration_test.all_datatypes WHERE int32_col = 2147483647",
		DryRun: true,
	})
	require.NoError(t, err)
}

func TestQuerySchema(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	schema, err := olap.QuerySchema(t.Context(), "SELECT int32_col, string_col FROM integration_test.all_datatypes", nil)
	require.NoError(t, err)
	require.Len(t, schema.Fields, 2)
	require.Equal(t, "int32_col", schema.Fields[0].Name)
	require.Equal(t, "string_col", schema.Fields[1].Name)
}

func acquireTestDatabricks(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "databricks")
	conn, err := drivers.Open("databricks", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
