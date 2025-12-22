package athena_test

import (
	"context"
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
			result: map[string]any{"val": mustParseTime(t, "2006-01-02", "2021-01-01", time.UTC)},
		},
		{
			name:   "time",
			query:  "SELECT TIME '10:11:12.345 -06:30' AS tpz",
			args:   nil,
			result: map[string]any{"tpz": mustParseTime(t, time.TimeOnly, "10:11:12.345", time.FixedZone("", -6*60*60-30*60))},
		},
		{
			name:   "timestamp",
			query:  "SELECT TIMESTAMP '2025-01-31 23:59:59.999 +06:30' AS val",
			args:   nil,
			result: map[string]any{"val": mustParseTime(t, time.DateTime, "2025-01-31 23:59:59.999", time.FixedZone("", 6*60*60+30*60))},
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

			require.True(t, result.Next(), "expected at least one row, scan error: %v", result.Err())
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
			Query: "SELECT int32_col, float_col FROM integration_test.all_datatypes WHERE id = ?",
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
			require.Equal(t, int32(123), res["int32_col"])
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

func TestScanAllRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestAthena(t)

	result, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT int32_col FROM integration_test.all_datatypes",
	})
	require.NoError(t, err)
	defer result.Close()

	var rowCount int
	for result.Next() {
		var int32Col *int32
		err = result.Scan(&int32Col)
		require.NoError(t, err)
		rowCount++
	}
	require.NoError(t, result.Err())
	require.Equal(t, rowCount, 3)
}

func TestQueryScan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestAthena(t)

	t.Run("scan basic types", func(t *testing.T) {
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

	t.Run("scan integer types", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT CAST(127 AS TINYINT) AS tiny, CAST(32767 AS SMALLINT) AS small, CAST(2147483647 AS INTEGER) AS int, CAST(9223372036854775807 AS BIGINT) AS big",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var tiny int8
		var small int16
		var intVal int32
		var big int64
		err = result.Scan(&tiny, &small, &intVal, &big)
		require.NoError(t, err)
		require.Equal(t, int8(127), tiny)
		require.Equal(t, int16(32767), small)
		require.Equal(t, int32(2147483647), intVal)
		require.Equal(t, int64(9223372036854775807), big)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan float types", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT CAST(3.14 AS REAL) AS float_val, CAST(3.14159265359 AS DOUBLE) AS double_val",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var floatVal float32
		var doubleVal float64
		err = result.Scan(&floatVal, &doubleVal)
		require.NoError(t, err)
		require.Equal(t, float32(3.14), floatVal)
		require.Equal(t, float64(3.14159265359), doubleVal)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan string types", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT 'hello world' AS str, CAST('test' AS VARCHAR) AS varchar_val",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var str string
		var varcharVal string
		err = result.Scan(&str, &varcharVal)
		require.NoError(t, err)
		require.Equal(t, "hello world", str)
		require.Equal(t, "test", varcharVal)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan date and timestamp", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT DATE '2021-01-01' AS date_val, TIMESTAMP '2025-01-31 23:59:59.999' AS timestamp_val",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var dateVal time.Time
		var timestampVal time.Time
		err = result.Scan(&dateVal, &timestampVal)
		require.NoError(t, err)
		require.Equal(t, mustParseTime(t, "2006-01-02", "2021-01-01", time.UTC), dateVal)
		require.Equal(t, mustParseTime(t, "2006-01-02 15:04:05.000", "2025-01-31 23:59:59.999", time.UTC), timestampVal)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan decimal", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT CAST(123.45 AS DECIMAL(10,2)) AS decimal_val",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var decimalVal string
		err = result.Scan(&decimalVal)
		require.NoError(t, err)
		require.Equal(t, "123.45", decimalVal)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan complex types", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT ARRAY[1, 2, 3] AS array_val, MAP(ARRAY['key1', 'key2'], ARRAY[1, 2]) AS map_val, CAST(ROW(1, 'abc') AS ROW(a INTEGER, b VARCHAR)) AS struct_val",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var arrayVal string
		var mapVal string
		var structVal string
		err = result.Scan(&arrayVal, &mapVal, &structVal)
		require.NoError(t, err)
		require.Equal(t, "[1, 2, 3]", arrayVal)
		require.Equal(t, "{key1=1, key2=2}", mapVal)
		require.Equal(t, "{a=1, b=abc}", structVal)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan boolean", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT true AS bool_true, false AS bool_false",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var boolTrue bool
		var boolFalse bool
		err = result.Scan(&boolTrue, &boolFalse)
		require.NoError(t, err)
		require.Equal(t, true, boolTrue)
		require.Equal(t, false, boolFalse)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan multiple rows", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT * FROM (VALUES (1, 'first'), (2, 'second'), (3, 'third')) AS t(id, name)",
		})
		require.NoError(t, err)
		defer result.Close()

		expectedRows := []struct {
			id   int32
			name string
		}{
			{1, "first"},
			{2, "second"},
			{3, "third"},
		}

		rowCount := 0
		for result.Next() {
			var id int32
			var name string
			err = result.Scan(&id, &name)
			require.NoError(t, err)
			require.Less(t, rowCount, len(expectedRows))
			require.Equal(t, expectedRows[rowCount].id, id)
			require.Equal(t, expectedRows[rowCount].name, name)
			rowCount++
		}

		require.Equal(t, len(expectedRows), rowCount)
		require.NoError(t, result.Err())
	})
}

func TestExec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestAthena(t)

	t.Run("simple exec", func(t *testing.T) {
		// Create a view
		err := olap.Exec(t.Context(), &drivers.Statement{
			Query: "CREATE OR REPLACE VIEW integration_test.all_datatypes_view AS SELECT * FROM integration_test.all_datatypes",
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			// Drop the view
			err = olap.Exec(context.Background(), &drivers.Statement{
				Query: "DROP TABLE integration_test.all_datatypes_view",
			})
			require.NoError(t, err)
		})

		// Verify the view was created by querying it
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT id FROM integration_test.all_datatypes_view limit 0",
		})
		require.NoError(t, err)
		require.NoError(t, result.Close())
		require.Equal(t, len(result.Schema.Fields), 1)
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

func mustParseTime(t *testing.T, layout, value string, loc *time.Location) time.Time {
	t.Helper()
	tm, err := time.ParseInLocation(layout, value, loc)
	require.NoError(t, err)
	return tm
}
