package redshift_test

import (
	"context"
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

func TestQuery(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestRedshift(t)
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
			name:   "real",
			query:  "SELECT CAST(3.14 AS REAL) AS val",
			args:   nil,
			result: map[string]any{"val": float32(3.14)},
		},
		{
			name:   "double precision",
			query:  "SELECT CAST(3.14159265359 AS DOUBLE PRECISION) AS val",
			args:   nil,
			result: map[string]any{"val": 3.14159265359},
		},
		{
			name:   "varchar",
			query:  "SELECT CAST('hello' AS VARCHAR(10)) AS val",
			args:   nil,
			result: map[string]any{"val": "hello"},
		},
		{
			name:   "date",
			query:  "SELECT DATE '2021-01-01' AS val",
			args:   nil,
			result: map[string]any{"val": time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			name:   "timestamp",
			query:  "SELECT TIMESTAMP '2021-01-01 12:30:45' AS val",
			args:   nil,
			result: map[string]any{"val": time.Date(2021, 1, 1, 12, 30, 45, 0, time.UTC)},
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

	// Test timestamp with timezone separately since it returns a time.Time with FixedZone location
	t.Run("timestamp with timezone", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT TIMESTAMPTZ '2021-01-01 12:30:45.666666-08:00' AS val",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next(), "expected at least one row, scan error: %v", result.Err())
		res := make(map[string]any)
		err = result.MapScan(res)
		require.NoError(t, err)

		// Check that we got a time.Time value
		actualTime, ok := res["val"].(time.Time)
		require.True(t, ok, "expected time.Time value")

		// The time should be converted to UTC: 12:30:45-08:00 = 20:30:45 UTC
		expectedTime := time.Date(2021, 1, 1, 20, 30, 45, 666666000, time.UTC)
		require.True(t, actualTime.Equal(expectedTime), "expected %v, got %v", expectedTime, actualTime)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})
}

func TestQueryWithParameters(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestRedshift(t)

	t.Run("simple parameter", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT :param1 AS val",
			Args:  []any{42},
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())
		var val int32
		err = result.Scan(&val)
		require.NoError(t, err)
		require.Equal(t, int32(42), val)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})
}

func TestEmptyRows(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestRedshift(t)
	result, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT 1 AS val LIMIT 0",
	})
	require.NoError(t, err)
	defer result.Close()

	// For empty result sets, we should still be able to iterate (but get no rows)
	require.False(t, result.Next())
	require.NoError(t, result.Err())
}

func TestScanAllRows(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestRedshift(t)
	result, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT generate_series(1, 10) AS val",
	})
	require.NoError(t, err)
	defer result.Close()

	var rowCount int
	for result.Next() {
		var val int32
		err = result.Scan(&val)
		require.NoError(t, err)
		rowCount++
	}
	require.NoError(t, result.Err())
	require.Equal(t, 10, rowCount)
}

func TestQueryScan(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestRedshift(t)

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
			Query: "SELECT CAST(32767 AS SMALLINT) AS small, CAST(2147483647 AS INTEGER) AS int, CAST(9223372036854775807 AS BIGINT) AS big",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var small int16
		var intVal int32
		var big int64
		err = result.Scan(&small, &intVal, &big)
		require.NoError(t, err)
		require.Equal(t, int16(32767), small)
		require.Equal(t, int32(2147483647), intVal)
		require.Equal(t, int64(9223372036854775807), big)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan null values", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT NULL AS val",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())

		var val *string
		err = result.Scan(&val)
		require.NoError(t, err)
		require.Nil(t, val)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("scan multiple rows", func(t *testing.T) {
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT * FROM (SELECT 1 AS id, 'first' AS name UNION ALL SELECT 2, 'second' UNION ALL SELECT 3, 'third') t ORDER BY id",
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

		require.NoError(t, result.Err())
		require.Equal(t, len(expectedRows), rowCount)
	})
}

func TestExec(t *testing.T) {
	testmode.Expensive(t)

	_, olap := acquireTestRedshift(t)

	t.Run("simple exec", func(t *testing.T) {
		// Create a temp table
		err := olap.Exec(t.Context(), &drivers.Statement{
			Query: "CREATE TABLE test_table (id INTEGER, name VARCHAR(50))",
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			err := olap.Exec(context.Background(), &drivers.Statement{Query: "DROP TABLE test_table"})
			require.NoError(t, err)
		})

		// Insert data
		err = olap.Exec(t.Context(), &drivers.Statement{
			Query: "INSERT INTO test_table VALUES (1, 'test')",
		})
		require.NoError(t, err)

		// Verify the data was inserted by querying it
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT id, name FROM test_table",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())
		var id int32
		var name string
		err = result.Scan(&id, &name)
		require.NoError(t, err)
		require.Equal(t, int32(1), id)
		require.Equal(t, "test", name)

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})

	t.Run("dry run exec", func(t *testing.T) {
		// Create a temp table for testing
		err := olap.Exec(t.Context(), &drivers.Statement{
			Query: "CREATE TABLE test_dry_run (id INTEGER, value VARCHAR(50))",
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			err := olap.Exec(context.Background(), &drivers.Statement{Query: "DROP TABLE test_dry_run"})
			require.NoError(t, err)
		})

		// Dry run insert should not error but also not insert data
		err = olap.Exec(t.Context(), &drivers.Statement{
			Query:  "INSERT INTO test_dry_run VALUES (1, 'should not be inserted')",
			DryRun: true,
		})
		require.NoError(t, err)

		// Verify no data was actually inserted
		result, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT COUNT(*) AS count FROM test_dry_run",
		})
		require.NoError(t, err)
		defer result.Close()

		require.True(t, result.Next())
		var count int32
		err = result.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, int32(0), count, "dry run should not insert data")

		require.False(t, result.Next())
		require.NoError(t, result.Err())
	})
}

func acquireTestRedshift(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "redshift")
	// Ensure database is set (required for Redshift Data API)
	if cfg["database"] == nil || cfg["database"] == "" {
		cfg["database"] = "test_db"
	}
	cfg["workgroup"] = "integration-test-wg"
	conn, err := drivers.Open("redshift", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
