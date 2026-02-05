package bigquery_test

import (
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
	_, olap := acquireTestBigQuery(t)
	tests := []struct {
		query  string
		args   []any
		result map[string]any
	}{
		{
			"SELECT [true, false, true] AS booleans;",
			nil,
			map[string]any{"booleans": "[true,false,true]"},
		},
		{
			"SELECT GENERATE_ARRAY(21, 14, -1) AS countdown;",
			nil,
			map[string]any{"countdown": "[21,20,19,18,17,16,15,14]"},
		},
		{
			"SELECT true AS bool",
			nil,
			map[string]any{"bool": true},
		},
		{
			"SELECT CAST('2021-01-01' AS DATE) AS date",
			nil,
			map[string]any{"date": "2021-01-01"},
		},
		{
			"SELECT CAST('2025-01-31 23:59:59.999999' AS DATETIME) AS datetime;",
			nil,
			map[string]any{"datetime": "2025-01-31T23:59:59.999999000"},
		},
		{
			`select JSON '{  "id": 10,  "type": "fruit",  "name": "apple",  "on_menu": true,  "recipes":    {      "salads":      [        { "id": 2001, "type": "Walnut Apple Salad" },        { "id": 2002, "type": "Apple Spinach Salad" }      ],      "desserts":      [        { "id": 3001, "type": "Apple Pie" },        { "id": 3002, "type": "Apple Scones" },        { "id": 3003, "type": "Apple Crumble" }      ]    }}' AS json`,
			nil,
			map[string]any{"json": `{"id":10,"name":"apple","on_menu":true,"recipes":{"desserts":[{"id":3001,"type":"Apple Pie"},{"id":3002,"type":"Apple Scones"},{"id":3003,"type":"Apple Crumble"}],"salads":[{"id":2001,"type":"Walnut Apple Salad"},{"id":2002,"type":"Apple Spinach Salad"}]},"type":"fruit"}`},
		},
		{
			"SELECT 9223372036854775807 AS integer",
			nil,
			map[string]any{"integer": int64(9223372036854775807)},
		},
		{
			"SELECT cast(9.9999999999999999999999999999999999999E+28 as NUMERIC) as number",
			nil,
			map[string]any{"number": "99999999999999999999999999999.999999999"},
		},
		{
			"SELECT cast(0.1 as NUMERIC) as number",
			nil,
			map[string]any{"number": "0.1"},
		},
		{
			"SELECT cast(5.7896044618658097711785492504343953926634992332820282019728792003956564819967E+38 as BIGNUMERIC) as number",
			nil,
			map[string]any{"number": "578960446186580977117854925043439539266.34992332820282019728792003956564819967"},
		},
		{
			"SELECT cast(3.14 as FLOAT64) as number",
			nil,
			map[string]any{"number": 3.14},
		},
		{
			"SELECT RANGE(Date'2020-01-01', Date'2025-01-01') AS date_range",
			nil,
			map[string]any{"date_range": "[2020-01-01, 2025-01-01)"},
		},
		{
			"SELECT STRUCT(1 AS a, 'abc' AS b) as str",
			nil,
			map[string]any{"str": `{"a":1,"b":"abc"}`},
		},
		{
			"SELECT TIME'23:59:59.999999' AS t",
			nil,
			map[string]any{"t": "23:59:59.999999000"},
		},
		{
			"SELECT TIMESTAMP'2025-01-01 23:59:59.999999 UTC' AS t",
			nil,
			map[string]any{"t": time.Date(2025, 1, 1, 23, 59, 59, 999999000, time.UTC)},
		},
		{
			"SELECT float_col FROM `rilldata.integration_test.all_datatypes` where int_col = ?",
			[]any{1},
			map[string]any{"float_col": 1.1},
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
	_, olap := acquireTestBigQuery(t)
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT int_col, float_col FROM `rilldata.integration_test.all_datatypes` LIMIT 0"})
	require.NoError(t, err)
	defer rows.Close()

	sc := rows.Schema
	require.Len(t, sc.Fields, 2)
	require.Equal(t, "int_col", sc.Fields[0].Name)
	require.Equal(t, "float_col", sc.Fields[1].Name)
	require.False(t, rows.Next())
	require.Nil(t, rows.Err())

}

func TestExec(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestBigQuery(t)

	// create table with dry run
	err := olap.Exec(t.Context(), &drivers.Statement{Query: "CREATE TABLE `rilldata.integration_test.exec_test` (id INT64, name STRING)", DryRun: true})
	require.NoError(t, err)

	// create table actually
	err = olap.Exec(t.Context(), &drivers.Statement{Query: "CREATE OR REPLACE TABLE `rilldata.integration_test.exec_test` (id INT64, name STRING)"})
	require.NoError(t, err)

	// drop table
	err = olap.Exec(t.Context(), &drivers.Statement{Query: "DROP TABLE `rilldata.integration_test.exec_test`"})
	require.NoError(t, err)
}

func TestScan(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestBigQuery(t)

	t.Run("successful scan with various types", func(t *testing.T) {
		rows, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT CAST(42 AS INT64) AS int_col, 'hello' AS str_col, TRUE AS bool_col, 3.14 AS float_col",
		})
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())

		var intVal int64
		var strVal string
		var boolVal bool
		var floatVal float64

		err = rows.Scan(&intVal, &strVal, &boolVal, &floatVal)
		require.NoError(t, err)
		require.Equal(t, int64(42), intVal)
		require.Equal(t, "hello", strVal)
		require.Equal(t, true, boolVal)
		require.Equal(t, 3.14, floatVal)

		require.False(t, rows.Next())
		require.NoError(t, rows.Err())
	})

	t.Run("scan with wrong number of destinations", func(t *testing.T) {
		rows, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT 1 AS col1, 2 AS col2",
		})
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())

		// Try to scan 2 columns into 1 destination
		var val int64
		err = rows.Scan(&val)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected 2 destination arguments in Scan, got 1")

		// Try to scan 2 columns into 3 destinations
		var val1, val2, val3 int64
		err = rows.Scan(&val1, &val2, &val3)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected 2 destination arguments in Scan, got 3")
	})

	t.Run("scan without calling Next first", func(t *testing.T) {
		rows, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT 1 AS col1",
		})
		require.NoError(t, err)
		defer rows.Close()

		// Try to scan without calling Next first
		var val int64
		err = rows.Scan(&val)
		require.Error(t, err)
		require.Contains(t, err.Error(), "must call Next before Scan")
	})

	t.Run("scan with NULL values", func(t *testing.T) {
		rows, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT CAST(NULL AS INT64) AS null_int, CAST(NULL AS STRING) AS null_str",
		})
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())

		var nullInt *int64
		var nullStr *string

		err = rows.Scan(&nullInt, &nullStr)
		require.NoError(t, err)
		require.Nil(t, nullInt)
		require.Nil(t, nullStr)
	})

	t.Run("scan multiple rows", func(t *testing.T) {
		rows, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT * FROM UNNEST([1, 2, 3]) AS val",
		})
		require.NoError(t, err)
		defer rows.Close()

		expectedVals := []int64{1, 2, 3}
		actualVals := []int64{}

		for rows.Next() {
			var val int64
			err = rows.Scan(&val)
			require.NoError(t, err)
			actualVals = append(actualVals, val)
		}

		require.NoError(t, rows.Err())
		require.Equal(t, expectedVals, actualVals)
	})

	t.Run("scan with type conversion", func(t *testing.T) {
		rows, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT CAST(123 AS INT64) AS int_col",
		})
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())

		// Test scanning int64 into different compatible types
		var intVal int
		err = rows.Scan(&intVal)
		require.NoError(t, err)
		require.Equal(t, 123, intVal)
	})

	t.Run("scan complex types as JSON strings", func(t *testing.T) {
		rows, err := olap.Query(t.Context(), &drivers.Statement{
			Query: "SELECT STRUCT(1 AS a, 'test' AS b) AS struct_col, [1, 2, 3] AS array_col",
		})
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())

		var structVal, arrayVal string
		err = rows.Scan(&structVal, &arrayVal)
		require.NoError(t, err)
		require.Contains(t, structVal, `"a":1`)
		require.Contains(t, structVal, `"b":"test"`)
		require.Equal(t, "[1,2,3]", arrayVal)
	})
}

func acquireTestBigQuery(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "bigquery")
	conn, err := drivers.Open("bigquery", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
