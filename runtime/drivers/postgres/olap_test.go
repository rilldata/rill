package postgres_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPgxOLAP(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestPostgres(t)

	t.Run("test map scan", func(t *testing.T) {
		testOLAP(t, olap)
	})

	t.Run("test empty rows", func(t *testing.T) {
		testEmptyRows(t, olap)
	})

	t.Run("test complex types", func(t *testing.T) {
		testComplexTypes(t, olap)
	})

	t.Run("test null values", func(t *testing.T) {
		testNullValues(t, olap)
	})

	t.Run("test timestamp with time zone", func(t *testing.T) {
		testTimestampWithTimeZone(t, olap)
	})

	t.Run("test numeric types", func(t *testing.T) {
		testNumericTypes(t, olap)
	})

	t.Run("test string types", func(t *testing.T) {
		testStringTypes(t, olap)
	})

	t.Run("test enum type", func(t *testing.T) {
		testEnumType(t, olap)
	})

	t.Run("test dry run", func(t *testing.T) {
		testDryRun(t, olap)
	})

	t.Run("test information schema", func(t *testing.T) {
		testInformationSchema(t, olap)
	})

	t.Run("test exec", func(t *testing.T) {
		testExec(t, olap)
	})
}

func testOLAP(t *testing.T, olap drivers.OLAPStore) {
	tests := []struct {
		query  string
		args   []any
		result map[string]any
	}{
		{
			"SELECT TRUE AS bool",
			nil,
			map[string]any{"bool": true},
		},
		{
			"SELECT '2021-01-01'::DATE AS date",
			nil,
			map[string]any{"date": time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			"SELECT '2025-01-31 23:59:59.999999'::TIMESTAMP AS datetime",
			nil,
			map[string]any{"datetime": time.Date(2025, 1, 31, 23, 59, 59, 999999000, time.UTC)},
		},
		{
			"SELECT 123::int2 AS smallint",
			nil,
			map[string]any{"smallint": int64(123)},
		},
		{
			"SELECT 123456789 AS integer",
			nil,
			map[string]any{"integer": int64(123456789)},
		},
		{
			"SELECT 99999999999999999999999999999.999999999::NUMERIC(38,9) AS number",
			nil,
			map[string]any{"number": "99999999999999999999999999999.999999999"},
		},
		{
			"SELECT 0.1::NUMERIC(10,1) AS number",
			nil,
			map[string]any{"number": "0.1"},
		},
		{
			"SELECT 3.14::FLOAT AS number",
			nil,
			map[string]any{"number": float64(3.14)},
		},
		{
			"SELECT ARRAY[1, 2, 3] AS arr",
			nil,
			map[string]any{"arr": "{1,2,3}"},
		},
		{
			"SELECT '23:59:59.999999'::TIME AS t",
			nil,
			map[string]any{"t": "23:59:59.999999"}, // TIME is returned as string
		},
		{
			"SELECT weight FROM all_datatypes WHERE age = $1",
			[]any{30},
			map[string]any{"weight": float64(75.4000015258789)}, // FLOAT4 returned as FLOAT64
		},
		{
			"SELECT name FROM all_datatypes WHERE uuid = $1",
			[]any{"8a25ac46-8ad6-4415-9a2e-12aa3962c144"},
			map[string]any{"name": "John Doe"},
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

func testEmptyRows(t *testing.T, olap drivers.OLAPStore) {
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT age, weight FROM all_datatypes LIMIT 0"})
	require.NoError(t, err)
	defer rows.Close()

	sc := rows.Schema
	require.Len(t, sc.Fields, 2)
	require.Equal(t, "age", sc.Fields[0].Name)
	require.Equal(t, "weight", sc.Fields[1].Name)
	require.False(t, rows.Next())
	require.Nil(t, rows.Err())
}

func testComplexTypes(t *testing.T, olap drivers.OLAPStore) {
	// Test complex data types (json, jsonb, array)
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT personal_info, personal_info2, salary_history FROM all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify JSON values (returned as []uint8 byte slices)
	var jsonCol map[string]string
	jsonBytes, ok := res["personal_info"].([]uint8)
	require.True(t, ok, "personal_info should be []uint8")
	err = json.Unmarshal(jsonBytes, &jsonCol)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"hobbies": "Travel, Tech"}, jsonCol)

	var jsonbCol map[string]string
	jsonbBytes, ok := res["personal_info2"].([]uint8)
	require.True(t, ok, "personal_info2 should be []uint8")
	err = json.Unmarshal(jsonbBytes, &jsonbCol)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"job": "Software Engineer"}, jsonbCol)

	// Verify array value (Postgres returns arrays as strings in the format "{val1,val2}")
	require.Equal(t, "{1234567,7654312}", res["salary_history"])

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testNullValues(t *testing.T, olap drivers.OLAPStore) {
	// Test NULL handling
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT is_married, emp_salary FROM all_datatypes WHERE id = 3",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify NULL values
	require.Nil(t, res["is_married"])
	require.Nil(t, res["emp_salary"])

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testTimestampWithTimeZone(t *testing.T, olap drivers.OLAPStore) {
	// Test timestamp with time zone
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT last_login FROM all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify the timestamp value exists (exact value may vary based on timezone)
	require.NotNil(t, res["last_login"])
	_, ok := res["last_login"].(time.Time)
	require.True(t, ok, "last_login should be a time.Time value")

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testNumericTypes(t *testing.T, olap drivers.OLAPStore) {
	// Test various numeric types
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT num_of_dependents, age, net_worth, weight, height FROM all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify numeric types (all integers return as int64)
	require.Equal(t, int64(2), res["num_of_dependents"])    // smallint -> int64
	require.Equal(t, int64(30), res["age"])                 // integer -> int64
	require.Equal(t, int64(1234567), res["net_worth"])      // bigint -> int64
	require.InDelta(t, float64(75.4), res["weight"], 0.1)   // float4 -> float64
	require.InDelta(t, float64(180.5), res["height"], 0.01) // float8 -> float64

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testStringTypes(t *testing.T, olap drivers.OLAPStore) {
	// Test various string/char types
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT name, gender, gender_full, nickname, biography FROM all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify string types
	require.Equal(t, "John Doe", res["name"])                                                                     // text
	require.Equal(t, "M", res["gender"])                                                                          // character
	require.Equal(t, "Male", res["gender_full"])                                                                  // character varying
	require.Equal(t, "abcd      ", res["nickname"])                                                               // bpchar(10) - padded with spaces
	require.Equal(t, "John is a software engineer who loves to travel and explore new places.", res["biography"]) // text

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testEnumType(t *testing.T, olap drivers.OLAPStore) {
	// Test ENUM type
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT country FROM all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify enum value (should be returned as string)
	require.Equal(t, "IND", res["country"])

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testDryRun(t *testing.T, olap drivers.OLAPStore) {
	// Dry run query
	_, err := olap.Query(t.Context(), &drivers.Statement{
		Query:  "SELECT * FROM all_datatypes WHERE age = $1",
		Args:   []any{30},
		DryRun: true,
	})
	require.NoError(t, err)
}

func testInformationSchema(t *testing.T, olap drivers.OLAPStore) {
	// Test All() method to list tables
	tables, _, err := olap.InformationSchema().All(t.Context(), "", 100, "")
	require.NoError(t, err)
	require.NotEmpty(t, tables)

	// Find our test table
	var foundTable *drivers.OlapTable
	for _, table := range tables {
		if table.Name == "all_datatypes" {
			foundTable = table
			break
		}
	}
	require.NotNil(t, foundTable, "all_datatypes table should be in the list")
	require.False(t, foundTable.View)

	// Test Lookup() method
	table, err := olap.InformationSchema().Lookup(t.Context(), foundTable.Database, foundTable.DatabaseSchema, "all_datatypes")
	require.NoError(t, err)
	require.NotNil(t, table)
	require.Equal(t, "all_datatypes", table.Name)
	require.NotNil(t, table.Schema)
	require.NotEmpty(t, table.Schema.Fields)

	// Verify some fields exist
	fieldNames := make(map[string]bool)
	for _, field := range table.Schema.Fields {
		fieldNames[field.Name] = true
	}
	require.True(t, fieldNames["id"])
	require.True(t, fieldNames["name"])
	require.True(t, fieldNames["age"])
	require.True(t, fieldNames["uuid"])
}

func testExec(t *testing.T, olap drivers.OLAPStore) {
	// Test Exec method - create a regular table instead of temp table
	// (temp tables are session-specific and won't work with connection pooling)
	tableName := "test_exec_" + time.Now().Format("20060102150405")

	err := olap.Exec(t.Context(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE TABLE %s (id INT, name TEXT)", tableName),
	})
	require.NoError(t, err)

	// Clean up at the end
	defer func() {
		_ = olap.Exec(t.Context(), &drivers.Statement{
			Query: fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName),
		})
	}()

	// Insert data
	err = olap.Exec(t.Context(), &drivers.Statement{
		Query: fmt.Sprintf("INSERT INTO %s (id, name) VALUES ($1, $2)", tableName),
		Args:  []any{1, "test"},
	})
	require.NoError(t, err)

	// Verify data was inserted
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: fmt.Sprintf("SELECT id, name FROM %s WHERE id = $1", tableName),
		Args:  []any{1},
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)
	require.Equal(t, int64(1), res["id"]) // INT returns as int64
	require.Equal(t, "test", res["name"])
}

func acquireTestPostgres(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "postgres")
	conn, err := drivers.Open("postgres", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
