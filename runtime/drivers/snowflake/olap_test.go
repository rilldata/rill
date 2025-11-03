package snowflake_test

import (
	"encoding/json"
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

	_, olap := acquireTestSnowflake(t)
	tests := []struct {
		query  string
		args   []any
		result map[string]any
	}{
		{
			"SELECT TRUE AS bool",
			nil,
			map[string]any{"BOOL": true},
		},
		{
			"SELECT FALSE AS bool",
			nil,
			map[string]any{"BOOL": false},
		},
		{
			"SELECT '2021-01-01'::DATE AS date",
			nil,
			map[string]any{"DATE": time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			"SELECT '2025-01-31 23:59:59.999999'::TIMESTAMP_NTZ AS datetime",
			nil,
			map[string]any{"DATETIME": time.Date(2025, 1, 31, 23, 59, 59, 999999000, time.UTC)},
		},
		{
			"SELECT 99999999999999999999999999999999999999 AS integer",
			nil,
			map[string]any{"INTEGER": "99999999999999999999999999999999999999"},
		},
		{
			"SELECT 99999999999999999999999999999.999999999::NUMBER(38,9) AS number",
			nil,
			map[string]any{"NUMBER": "99999999999999999999999999999.999999999"},
		},
		{
			"SELECT 0.1::NUMBER(10,1) AS number",
			nil,
			map[string]any{"NUMBER": "0.1"},
		},
		{
			"SELECT 3.14::FLOAT AS number",
			nil,
			map[string]any{"NUMBER": 3.14},
		},
		{
			"SELECT ARRAY_CONSTRUCT(1, 2, 3) AS arr",
			nil,
			map[string]any{"ARR": "[\n  1,\n  2,\n  3\n]"},
		},
		{
			"SELECT OBJECT_CONSTRUCT('a', 1, 'b', 'abc') AS obj",
			nil,
			map[string]any{"OBJ": "{\n  \"a\": 1,\n  \"b\": \"abc\"\n}"},
		},
		{
			"SELECT '23:59:59.999999'::TIME AS t",
			nil,
			map[string]any{"T": time.Date(1, 1, 1, 23, 59, 59, 999999000, time.UTC)},
		},
		{
			"SELECT float_col FROM integration_test.public.all_datatypes WHERE int32_col = ?",
			[]any{2147483647},
			map[string]any{"FLOAT_COL": 3.14},
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

func TestEmptyRows(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestSnowflake(t)
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT int32_col, float_col FROM integration_test.public.all_datatypes LIMIT 0"})
	require.NoError(t, err)
	defer rows.Close()

	sc := rows.Schema
	require.Len(t, sc.Fields, 2)
	require.Equal(t, "INT32_COL", sc.Fields[0].Name)
	require.Equal(t, "FLOAT_COL", sc.Fields[1].Name)
	require.False(t, rows.Next())
	require.Nil(t, rows.Err())
}

func TestComplexTypes(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestSnowflake(t)

	// Test complex data types (variant, array, object)
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT variant_col, array_col, object_col FROM integration_test.public.all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify values
	var varCol map[string]string
	err = json.Unmarshal([]byte(res["VARIANT_COL"].(string)), &varCol)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "value"}, varCol)

	var arrCol []int
	err = json.Unmarshal([]byte(res["ARRAY_COL"].(string)), &arrCol)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, arrCol)

	var objCol map[string]any
	err = json.Unmarshal([]byte(res["OBJECT_COL"].(string)), &objCol)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"city": "New York"}, objCol)

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func TestDryRun(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestSnowflake(t)
	// Dry run query
	_, err := olap.Query(t.Context(), &drivers.Statement{
		Query:  "SELECT * FROM integration_test.public.all_datatypes WHERE int32_col = ?",
		Args:   []any{2147483647},
		DryRun: true,
	})
	require.NoError(t, err)
}

func acquireTestSnowflake(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "snowflake")
	conn, err := drivers.Open("snowflake", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
