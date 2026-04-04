package oracle_test

import (
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

	_ "github.com/sijms/go-ora/v2"
)

func TestOracleOLAP(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestOracle(t)

	t.Run("test map scan", func(t *testing.T) {
		testMapScan(t, olap)
	})

	t.Run("test empty rows", func(t *testing.T) {
		testEmptyRows(t, olap)
	})

	t.Run("test null values", func(t *testing.T) {
		testNullValues(t, olap)
	})

	t.Run("test numeric types", func(t *testing.T) {
		testNumericTypes(t, olap)
	})

	t.Run("test string types", func(t *testing.T) {
		testStringTypes(t, olap)
	})

	t.Run("test timestamp types", func(t *testing.T) {
		testTimestampTypes(t, olap)
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

	t.Run("test LoadDDL", func(t *testing.T) {
		testLoadDDL(t, olap)
	})
}

// toFloat64 converts a numeric value to float64, handling both float32 and float64.
// Oracle's go-ora driver returns float32 for BINARY_FLOAT and float64 for NUMBER/BINARY_DOUBLE.
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

func testMapScan(t *testing.T, olap drivers.OLAPStore) {
	tests := []struct {
		name     string
		query    string
		args     []any
		expected float64
		col      string
	}{
		{
			name:     "NUMBER column",
			query:    "SELECT number_col FROM all_datatypes WHERE id = 1",
			col:      "NUMBER_COL",
			expected: 42.5,
		},
		{
			name:     "BINARY_FLOAT column",
			query:    "SELECT binary_float_col FROM all_datatypes WHERE id = 1",
			col:      "BINARY_FLOAT_COL",
			expected: 3.14,
		},
		{
			name:     "BINARY_DOUBLE column",
			query:    "SELECT binary_double_col FROM all_datatypes WHERE id = 1",
			col:      "BINARY_DOUBLE_COL",
			expected: 2.718281828,
		},
		{
			name:     "INTEGER column",
			query:    "SELECT integer_col FROM all_datatypes WHERE id = 1",
			col:      "INTEGER_COL",
			expected: 1234567,
		},
		{
			name:     "BOOLEAN column (NUMBER(1))",
			query:    "SELECT boolean_col FROM all_datatypes WHERE id = 1",
			col:      "BOOLEAN_COL",
			expected: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query, Args: test.args})
			require.NoError(t, err)
			defer rows.Close()
			require.True(t, rows.Next())
			res := make(map[string]any)
			err = rows.MapScan(res)
			require.NoError(t, err)
			actual, ok := toFloat64(res[test.col])
			require.True(t, ok, "expected numeric for %s, got %T", test.col, res[test.col])
			require.InDelta(t, test.expected, actual, 0.01)
			require.False(t, rows.Next())
			require.NoError(t, rows.Err())
		})
	}

	// Test string column separately
	t.Run("VARCHAR2 column", func(t *testing.T) {
		rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT varchar2_col FROM all_datatypes WHERE id = 1"})
		require.NoError(t, err)
		defer rows.Close()
		require.True(t, rows.Next())
		res := make(map[string]any)
		err = rows.MapScan(res)
		require.NoError(t, err)
		require.Equal(t, "Hello World", res["VARCHAR2_COL"])
	})
}

func testEmptyRows(t *testing.T, olap drivers.OLAPStore) {
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT number_col, varchar2_col FROM all_datatypes WHERE 1=0"})
	require.NoError(t, err)
	defer rows.Close()

	sc := rows.Schema
	require.Len(t, sc.Fields, 2)
	require.Equal(t, "NUMBER_COL", sc.Fields[0].Name)
	require.Equal(t, "VARCHAR2_COL", sc.Fields[1].Name)
	require.False(t, rows.Next())
	require.Nil(t, rows.Err())
}

func testNullValues(t *testing.T, olap drivers.OLAPStore) {
	// Row 2 has all NULLs
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT number_col, varchar2_col, date_col FROM all_datatypes WHERE id = 2",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	require.Nil(t, res["NUMBER_COL"])
	require.Nil(t, res["VARCHAR2_COL"])
	require.Nil(t, res["DATE_COL"])

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testNumericTypes(t *testing.T, olap drivers.OLAPStore) {
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT number_col, binary_float_col, binary_double_col, integer_col, smallint_col FROM all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Oracle maps all numeric types through NUMBER internally; go-ora returns float32 for BINARY_FLOAT
	numberVal, ok := toFloat64(res["NUMBER_COL"])
	require.True(t, ok)
	require.InDelta(t, 42.5, numberVal, 0.01)

	floatVal, ok := toFloat64(res["BINARY_FLOAT_COL"])
	require.True(t, ok)
	require.InDelta(t, 3.14, floatVal, 0.01)

	doubleVal, ok := toFloat64(res["BINARY_DOUBLE_COL"])
	require.True(t, ok)
	require.InDelta(t, 2.718281828, doubleVal, 0.001)

	intVal, ok := toFloat64(res["INTEGER_COL"])
	require.True(t, ok)
	require.InDelta(t, 1234567, intVal, 0.01)

	smallVal, ok := toFloat64(res["SMALLINT_COL"])
	require.True(t, ok)
	require.InDelta(t, 123, smallVal, 0.01)

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testStringTypes(t *testing.T, olap drivers.OLAPStore) {
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT varchar2_col, nvarchar2_col, char_col, clob_col FROM all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	require.Equal(t, "Hello World", res["VARCHAR2_COL"])
	require.Equal(t, "Unicode Text", res["NVARCHAR2_COL"])
	// CHAR(10) is padded with spaces
	require.Equal(t, "ABCD      ", res["CHAR_COL"])
	require.Equal(t, "This is a CLOB text field for testing.", res["CLOB_COL"])

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testTimestampTypes(t *testing.T, olap drivers.OLAPStore) {
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT date_col, timestamp_col, timestamp_tz_col FROM all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify timestamp values are returned as time.Time
	dateVal, ok := res["DATE_COL"].(time.Time)
	require.True(t, ok, "DATE_COL should be time.Time, got %T", res["DATE_COL"])
	require.Equal(t, 2024, dateVal.Year())
	require.Equal(t, time.February, dateVal.Month())
	require.Equal(t, 14, dateVal.Day())

	tsVal, ok := res["TIMESTAMP_COL"].(time.Time)
	require.True(t, ok, "TIMESTAMP_COL should be time.Time, got %T", res["TIMESTAMP_COL"])
	require.Equal(t, 2025, tsVal.Year())
	require.Equal(t, 12, tsVal.Hour())

	tsTzVal, ok := res["TIMESTAMP_TZ_COL"].(time.Time)
	require.True(t, ok, "TIMESTAMP_TZ_COL should be time.Time, got %T", res["TIMESTAMP_TZ_COL"])
	require.NotNil(t, tsTzVal)

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func testDryRun(t *testing.T, olap drivers.OLAPStore) {
	_, err := olap.Query(t.Context(), &drivers.Statement{
		Query:  "SELECT * FROM all_datatypes WHERE id = :1",
		Args:   []any{1},
		DryRun: true,
	})
	require.NoError(t, err)
}

func testInformationSchema(t *testing.T, olap drivers.OLAPStore) {
	// Test Lookup() method directly; All() may fail for SYSTEM user with too many tables.
	// The SYSTEM user owns the table we created.
	table, err := olap.InformationSchema().Lookup(t.Context(), "", "SYSTEM", "ALL_DATATYPES")
	require.NoError(t, err)
	require.NotNil(t, table)
	require.Equal(t, "ALL_DATATYPES", table.Name)
	require.NotNil(t, table.Schema)
	require.NotEmpty(t, table.Schema.Fields)

	// Verify some fields exist
	fieldNames := make(map[string]bool)
	for _, field := range table.Schema.Fields {
		fieldNames[field.Name] = true
	}
	require.True(t, fieldNames["ID"])
	require.True(t, fieldNames["VARCHAR2_COL"])
	require.True(t, fieldNames["NUMBER_COL"])
}

func testExec(t *testing.T, olap drivers.OLAPStore) {
	tableName := fmt.Sprintf("TEST_EXEC_%d", time.Now().UnixNano())

	// CREATE TABLE (no bind params)
	err := olap.Exec(t.Context(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE TABLE %s (id NUMBER, name VARCHAR2(100))", tableName),
	})
	require.NoError(t, err)

	defer func() {
		_ = olap.Exec(t.Context(), &drivers.Statement{
			Query: fmt.Sprintf("DROP TABLE %s", tableName),
		})
	}()

	// INSERT with literal values (avoid bind param issues with go-ora QueryxContext on DML)
	err = olap.Exec(t.Context(), &drivers.Statement{
		Query: fmt.Sprintf("INSERT INTO %s (id, name) VALUES (1, 'test')", tableName),
	})
	require.NoError(t, err)

	// Verify data was inserted
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: fmt.Sprintf("SELECT id, name FROM %s WHERE id = 1", tableName),
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)
	idVal, ok := toFloat64(res["ID"])
	require.True(t, ok)
	require.InDelta(t, 1, idVal, 0.01)
	require.Equal(t, "test", res["NAME"])
}

func testLoadDDL(t *testing.T, olap drivers.OLAPStore) {
	table, err := olap.InformationSchema().Lookup(t.Context(), "", "SYSTEM", "ALL_DATATYPES")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(t.Context(), table)
	require.NoError(t, err)
	require.Contains(t, table.DDL, "ALL_DATATYPES")

	// Create a view and test DDL for it
	viewName := fmt.Sprintf("TEST_DDL_VIEW_%d", time.Now().UnixNano())
	err = olap.Exec(t.Context(), &drivers.Statement{
		Query: fmt.Sprintf("CREATE VIEW %s AS SELECT id, varchar2_col FROM all_datatypes", viewName),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = olap.Exec(t.Context(), &drivers.Statement{Query: fmt.Sprintf("DROP VIEW %s", viewName)})
	})

	view, err := olap.InformationSchema().Lookup(t.Context(), "", "SYSTEM", viewName)
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(t.Context(), view)
	require.NoError(t, err)
	require.Contains(t, view.DDL, viewName)
}

func acquireTestOracle(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "oracle")
	conn, err := drivers.Open("oracle", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
