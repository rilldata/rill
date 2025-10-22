package athena_test

import (
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestAthena(t)
	tests := []struct {
		name   string
		query  string
		args   []any
		result map[string]any
	}{
		{
			name:   "boolean true",
			query:  "SELECT true AS bool",
			args:   nil,
			result: map[string]any{"bool": true},
		},
		{
			name:   "boolean false",
			query:  "SELECT false AS bool",
			args:   nil,
			result: map[string]any{"bool": false},
		},
		{
			name:   "tinyint",
			query:  "SELECT CAST(127 AS TINYINT) AS val",
			args:   nil,
			result: map[string]any{"val": int8(127)},
		},
		{
			name:   "smallint",
			query:  "SELECT CAST(32767 AS SMALLINT) AS val",
			args:   nil,
			result: map[string]any{"val": int16(32767)},
		},
		{
			name:   "integer",
			query:  "SELECT CAST(2147483647 AS INTEGER) AS val",
			args:   nil,
			result: map[string]any{"val": int32(2147483647)},
		},
		{
			name:   "bigint",
			query:  "SELECT CAST(9223372036854775807 AS BIGINT) AS val",
			args:   nil,
			result: map[string]any{"val": int64(9223372036854775807)},
		},
		{
			name:   "float",
			query:  "SELECT CAST(3.14 AS REAL) AS val",
			args:   nil,
			result: map[string]any{"val": float32(3.14)},
		},
		{
			name:   "double",
			query:  "SELECT CAST(3.14159265359 AS DOUBLE) AS val",
			args:   nil,
			result: map[string]any{"val": float64(3.14159265359)},
		},
		{
			name:   "string",
			query:  "SELECT 'hello world' AS val",
			args:   nil,
			result: map[string]any{"val": "hello world"},
		},
		{
			name:   "varchar",
			query:  "SELECT CAST('test' AS VARCHAR) AS val",
			args:   nil,
			result: map[string]any{"val": "test"},
		},
		{
			name:   "date",
			query:  "SELECT DATE '2021-01-01' AS val",
			args:   nil,
			result: map[string]any{"val": mustParseTime(t, "2006-01-02", "2021-01-01")},
		},
		{
			name:   "timestamp",
			query:  "SELECT TIMESTAMP '2025-01-31 23:59:59.999' AS val",
			args:   nil,
			result: map[string]any{"val": mustParseTime(t, "2006-01-02 15:04:05.000", "2025-01-31 23:59:59.999")},
		},
		{
			name:   "decimal",
			query:  "SELECT CAST(123.45 AS DECIMAL(10,2)) AS val",
			args:   nil,
			result: map[string]any{"val": "123.45"},
		},
		{
			name:   "array of integers",
			query:  "SELECT ARRAY[1, 2, 3] AS val",
			args:   nil,
			result: map[string]any{"val": "[1, 2, 3]"},
		},
		{
			name:   "array of strings",
			query:  "SELECT ARRAY['a', 'b', 'c'] AS val",
			args:   nil,
			result: map[string]any{"val": "[a, b, c]"},
		},
		{
			name:   "map",
			query:  "SELECT MAP(ARRAY['key1', 'key2'], ARRAY[1, 2]) AS val",
			args:   nil,
			result: map[string]any{"val": "{key1=1, key2=2}"},
		},
		{
			name:   "struct",
			query:  "SELECT CAST(ROW(1, 'abc') AS ROW(a INTEGER, b VARCHAR)) AS val",
			args:   nil,
			result: map[string]any{"val": "{a=1, b=abc}"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query, Args: test.args})
			require.NoError(t, err)
			defer result.Close()

			require.True(t, result.Next(), "expected at least one row")
			res := make(map[string]any)
			err = result.MapScan(res)
			require.NoError(t, err)
			require.Equal(t, test.result, res)
			require.False(t, result.Next(), "expected only one row")
			require.NoError(t, result.Err())
		})
	}
}

func TestQueryWithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestAthena(t)

	t.Run("query with parameter", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT int32_col, float_col FROM integration_test.all_datatypes WHERE int32_col = ?",
			Args:  []any{1},
		})
		require.NoError(t, err)
		defer result.Close()

		hasRows := false
		for result.Next() {
			hasRows = true
			res := make(map[string]any)
			err = result.MapScan(res)
			require.NoError(t, err)
			require.Equal(t, int32(1), res["int32_col"])
			require.NotNil(t, res["float_col"])
		}
		require.True(t, hasRows, "expected at least one row")
		require.NoError(t, result.Err())
	})
}

func TestEmptyRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestAthena(t)

	result, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT int32_col, float_col FROM integration_test.all_datatypes LIMIT 0",
	})
	require.NoError(t, err)
	defer result.Close()

	// For empty result sets, we should still be able to iterate (but get no rows)
	require.False(t, result.Next())
	require.NoError(t, result.Err())
}

func TestQueryScan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestAthena(t)

	t.Run("scan values", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT 42 AS num, 'test' AS str, true AS flag",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var num int32
		var str string
		var flag bool
		err = result.Scan(&num, &str, &flag)
		require.NoError(t, err)
		require.Equal(t, int32(42), num)
		require.Equal(t, "test", str)
		require.Equal(t, true, flag)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})
}

func acquireTestAthena(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "athena")
	cfg["output_location"] = "s3://integration-test.rilldata.com/athena/"
	conn, err := drivers.Open("athena", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}

func mustParseTime(t *testing.T, layout, value string) time.Time {
	t.Helper()
	tm, err := time.ParseInLocation(layout, value, time.Local)
	require.NoError(t, err)
	return tm
}
