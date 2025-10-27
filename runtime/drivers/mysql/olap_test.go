package mysql_test

import (
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOLAP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestMySQL(t)
	t.Run("Test MapScan", func(t *testing.T) {
		testMapScan(t, olap)
	})
	t.Run("Test Scan", func(t *testing.T) {
		testScan(t, olap)
	})
	t.Run("Test Empty Rows", func(t *testing.T) {
		testEmptyRows(t, olap)
	})
	t.Run("Test Exec", func(t *testing.T) {
		testExec(t, olap)
	})
	t.Run("Test Scan Full Table", func(t *testing.T) {
		testFullTableScan(t, olap)
	})

}

func testMapScan(t *testing.T, olap drivers.OLAPStore) {
	tests := []struct {
		query  string
		args   []any
		result map[string]any
	}{
		// Integer types: TINYINT, SMALLINT, MEDIUMINT, INT, BIGINT
		{
			"SELECT tinyint_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"tinyint_col": int8(127)},
		},
		{
			"SELECT smallint_col FROM all_datatypes WHERE smallint_col = 32767",
			nil,
			map[string]any{"smallint_col": int16(32767)},
		},
		{
			"SELECT mediumint_col FROM all_datatypes WHERE mediumint_col = 8388607",
			nil,
			map[string]any{"mediumint_col": int32(8388607)},
		},
		{
			"SELECT int_col FROM all_datatypes WHERE int_col = ?",
			[]any{2147483647},
			map[string]any{"int_col": int32(2147483647)},
		},
		{
			"SELECT bigint_col FROM all_datatypes WHERE bigint_col = 999999999999999999",
			nil,
			map[string]any{"bigint_col": int64(999999999999999999)},
		},
		// Boolean type
		{
			"SELECT boolean_col FROM all_datatypes WHERE boolean_col = 1",
			nil,
			map[string]any{"boolean_col": int8(1)},
		},
		// Floating point types: FLOAT, DOUBLE
		{
			"SELECT float_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"float_col": float64(1.1)},
		},
		{
			"SELECT double_col FROM all_datatypes WHERE double_col = 2.2",
			nil,
			map[string]any{"double_col": float64(2.2)},
		},
		// Decimal type
		{
			"SELECT decimal_col FROM all_datatypes WHERE decimal_col = 3.3",
			nil,
			map[string]any{"decimal_col": "3.30"},
		},
		// String types: CHAR, VARCHAR, TINYTEXT, TEXT, MEDIUMTEXT, LONGTEXT
		{
			"SELECT char_col FROM all_datatypes WHERE char_col = 'C'",
			nil,
			map[string]any{"char_col": "C"},
		},
		{
			"SELECT varchar_col FROM all_datatypes WHERE varchar_col = 'VarChar'",
			nil,
			map[string]any{"varchar_col": "VarChar"},
		},
		{
			"SELECT tinytext_col FROM all_datatypes WHERE tinytext_col = 'Tiny Text'",
			nil,
			map[string]any{"tinytext_col": "Tiny Text"},
		},
		{
			"SELECT text_col FROM all_datatypes WHERE text_col = 'Text'",
			nil,
			map[string]any{"text_col": "Text"},
		},
		{
			"SELECT mediumtext_col FROM all_datatypes WHERE mediumtext_col = 'Medium Text'",
			nil,
			map[string]any{"mediumtext_col": "Medium Text"},
		},
		{
			"SELECT longtext_col FROM all_datatypes WHERE longtext_col = 'Long text content'",
			nil,
			map[string]any{"longtext_col": "Long text content"},
		},
		// Binary types: BINARY, VARBINARY, TINYBLOB, BLOB, MEDIUMBLOB, LONGBLOB
		{
			"SELECT binary_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"binary_col": "Binary\x00\x00\x00\x00"},
		},
		{
			"SELECT varbinary_col FROM all_datatypes WHERE varbinary_col = 'VarBinary'",
			nil,
			map[string]any{"varbinary_col": "VarBinary"},
		},
		{
			"SELECT tinyblob_col FROM all_datatypes WHERE tinyblob_col = 'Tiny Blob'",
			nil,
			map[string]any{"tinyblob_col": "Tiny Blob"},
		},
		{
			"SELECT blob_col FROM all_datatypes WHERE blob_col = 'Blob'",
			nil,
			map[string]any{"blob_col": "Blob"},
		},
		{
			"SELECT mediumblob_col FROM all_datatypes WHERE mediumblob_col = 'Medium Blob'",
			nil,
			map[string]any{"mediumblob_col": "Medium Blob"},
		},
		{
			"SELECT longblob_col FROM all_datatypes WHERE longblob_col = 'Long Blob'",
			nil,
			map[string]any{"longblob_col": "Long Blob"},
		},
		// Enum and Set types
		{
			"SELECT enum_col FROM all_datatypes WHERE enum_col = 'medium'",
			nil,
			map[string]any{"enum_col": "medium"},
		},
		{
			"SELECT set_col FROM all_datatypes WHERE set_col = 'a,b'",
			nil,
			map[string]any{"set_col": "a,b"},
		},
		// Date and Time types: DATE, TIME, DATETIME, TIMESTAMP, YEAR
		{
			"SELECT date_col FROM all_datatypes WHERE date_col = '2024-02-14'",
			nil,
			map[string]any{"date_col": "2024-02-14"},
		},
		{
			"SELECT time_col FROM all_datatypes WHERE time_col = '12:34:56'",
			nil,
			map[string]any{"time_col": "12:34:56"},
		},
		{
			"SELECT datetime_col FROM all_datatypes WHERE datetime_col = '2025-02-14 12:34:56'",
			nil,
			map[string]any{"datetime_col": "2025-02-14 12:34:56"},
		},
		{
			"SELECT timestamp_col FROM all_datatypes WHERE timestamp_col = '2025-02-14 12:34:56'",
			nil,
			map[string]any{"timestamp_col": "2025-02-14 12:34:56"},
		},
		{
			"SELECT year_col FROM all_datatypes WHERE year_col = 2024",
			nil,
			map[string]any{"year_col": int16(2024)},
		},
		// BIT type
		{
			"SELECT bit_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"bit_col": "1"},
		},
		// JSON type
		{
			"SELECT json_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"json_col": `{"key": "value"}`},
		},
		{
			"SELECT NULL",
			nil,
			map[string]any{"NULL": nil},
		},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			rows, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query, Args: test.args})
			require.NoError(t, err)
			defer rows.Close()
			require.True(t, rows.Next())
			res := make(map[string]any)
			err = rows.MapScan(res)
			require.NoError(t, err)
			require.Equal(t, test.result, res)
			require.False(t, rows.Next())
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
		})
	}
}

func testScan(t *testing.T, olap drivers.OLAPStore) {
	tests := []struct {
		name     string
		query    string
		args     []any
		scanFunc func(rows drivers.Rows) error
	}{
		{
			name:  "Null scan",
			query: "SELECT NULL AS null_col",
			args:  nil,
			scanFunc: func(rows drivers.Rows) error {
				var nullCol any

				require.True(t, rows.Next())
				err := rows.Scan(&nullCol)
				require.NoError(t, err)
				require.Nil(t, nullCol)
				return nil
			},
		},
		{
			name:  "Integer types scan",
			query: "SELECT tinyint_col, smallint_col, mediumint_col, int_col, bigint_col FROM all_datatypes WHERE int_col = ?",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var tinyintCol int8
				var smallintCol int16
				var mediumintCol int32
				var intCol int32
				var bigintCol int64

				require.True(t, rows.Next())
				err := rows.Scan(&tinyintCol, &smallintCol, &mediumintCol, &intCol, &bigintCol)
				require.NoError(t, err)
				require.Equal(t, int8(127), tinyintCol)
				require.Equal(t, int16(32767), smallintCol)
				require.Equal(t, int32(8388607), mediumintCol)
				require.Equal(t, int32(2147483647), intCol)
				require.Equal(t, int64(999999999999999999), bigintCol)
				return nil
			},
		},
		{
			name:  "Floating point types scan",
			query: "SELECT float_col, double_col, decimal_col FROM all_datatypes WHERE int_col = ?",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var floatCol float64
				var doubleCol float64
				var decimalCol string

				require.True(t, rows.Next())
				err := rows.Scan(&floatCol, &doubleCol, &decimalCol)
				require.NoError(t, err)
				require.Equal(t, float64(1.1), floatCol)
				require.Equal(t, float64(2.2), doubleCol)
				require.Equal(t, "3.30", decimalCol)
				return nil
			},
		},
		{
			name:  "String types scan",
			query: "SELECT char_col, varchar_col, tinytext_col, text_col, mediumtext_col, longtext_col FROM all_datatypes WHERE int_col = ?",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var charCol string
				var varcharCol string
				var tinytextCol string
				var textCol string
				var mediumtextCol string
				var longtextCol string

				require.True(t, rows.Next())
				err := rows.Scan(&charCol, &varcharCol, &tinytextCol, &textCol, &mediumtextCol, &longtextCol)
				require.NoError(t, err)
				require.Equal(t, "C", charCol)
				require.Equal(t, "VarChar", varcharCol)
				require.Equal(t, "Tiny Text", tinytextCol)
				require.Equal(t, "Text", textCol)
				require.Equal(t, "Medium Text", mediumtextCol)
				require.Equal(t, "Long text content", longtextCol)
				return nil
			},
		},
		{
			name:  "Binary types scan",
			query: "SELECT binary_col, varbinary_col, tinyblob_col, blob_col, mediumblob_col, longblob_col FROM all_datatypes WHERE int_col = ?",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var binaryCol string
				var varbinaryCol string
				var tinyblobCol string
				var blobCol string
				var mediumblobCol string
				var longblobCol string

				require.True(t, rows.Next())
				err := rows.Scan(&binaryCol, &varbinaryCol, &tinyblobCol, &blobCol, &mediumblobCol, &longblobCol)
				require.NoError(t, err)
				require.Equal(t, "Binary", strings.TrimRight(binaryCol, "\x00")) // binary is padded with null bytes
				require.Equal(t, "VarBinary", varbinaryCol)
				require.Equal(t, "Tiny Blob", tinyblobCol)
				require.Equal(t, "Blob", blobCol)
				require.Equal(t, "Medium Blob", mediumblobCol)
				require.Equal(t, "Long Blob", longblobCol)
				return nil
			},
		},
		{
			name:  "Enum and Set types scan",
			query: "SELECT enum_col, set_col FROM all_datatypes WHERE int_col = 2147483647",
			args:  nil,
			scanFunc: func(rows drivers.Rows) error {
				var enumCol string
				var setCol string

				require.True(t, rows.Next())
				err := rows.Scan(&enumCol, &setCol)
				require.NoError(t, err)
				require.Equal(t, "medium", enumCol)
				require.Equal(t, "a,b", setCol)
				return nil
			},
		},
		{
			name:  "Date and Time types scan",
			query: "SELECT date_col, datetime_col, timestamp_col, time_col, year_col FROM all_datatypes WHERE int_col = 2147483647",
			args:  nil,
			scanFunc: func(rows drivers.Rows) error {
				var dateCol string
				var datetimeCol string
				var timestampCol string
				var timeCol string
				var yearCol int64

				require.True(t, rows.Next())
				err := rows.Scan(&dateCol, &datetimeCol, &timestampCol, &timeCol, &yearCol)
				require.NoError(t, err)
				require.Equal(t, "2024-02-14", dateCol)
				require.Equal(t, "2025-02-14 12:34:56", datetimeCol)
				require.Equal(t, "2025-02-14 12:34:56", timestampCol)
				require.Equal(t, "12:34:56", timeCol)
				require.Equal(t, int64(2024), yearCol)
				return nil
			},
		},
		{
			name:  "Boolean and Bit types scan",
			query: "SELECT boolean_col, bit_col FROM all_datatypes WHERE int_col = 2147483647",
			args:  nil,
			scanFunc: func(rows drivers.Rows) error {
				var booleanCol int8
				var bitCol string

				require.True(t, rows.Next())
				err := rows.Scan(&booleanCol, &bitCol)
				require.NoError(t, err)
				require.Equal(t, int8(1), booleanCol)
				require.Equal(t, "1", bitCol)
				return nil
			},
		},
		{
			name:  "JSON type scan",
			query: "SELECT json_col FROM all_datatypes WHERE int_col = 2147483647",
			args:  nil,
			scanFunc: func(rows drivers.Rows) error {
				var jsonCol string

				require.True(t, rows.Next())
				err := rows.Scan(&jsonCol)
				require.NoError(t, err)
				require.Equal(t, `{"key": "value"}`, jsonCol)
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query, Args: test.args})
			require.NoError(t, err)
			defer rows.Close()

			err = test.scanFunc(rows)
			require.NoError(t, err)
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
		})
	}
}

func testEmptyRows(t *testing.T, olap drivers.OLAPStore) {
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT int_col, float_col FROM all_datatypes LIMIT 0"})
	require.NoError(t, err)
	defer rows.Close()

	sc := rows.Schema
	require.Len(t, sc.Fields, 2)
	require.Equal(t, "int_col", sc.Fields[0].Name)
	require.Equal(t, "float_col", sc.Fields[1].Name)
	require.False(t, rows.Next())
	require.Nil(t, rows.Err())
}

func testExec(t *testing.T, olap drivers.OLAPStore) {
	// dry run with SELECT query
	_, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT 1", DryRun: true})
	require.NoError(t, err)

	// create table
	err = olap.Exec(t.Context(), &drivers.Statement{Query: "CREATE TABLE IF NOT EXISTS exec_test (id INT, name VARCHAR(255))"})
	require.NoError(t, err)

	// drop table
	err = olap.Exec(t.Context(), &drivers.Statement{Query: "DROP TABLE IF EXISTS exec_test"})
	require.NoError(t, err)
}

func testFullTableScan(t *testing.T, olap drivers.OLAPStore) {
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: `SELECT 
		tinyint_col, smallint_col, mediumint_col, int_col, bigint_col,
		boolean_col, float_col, double_col, decimal_col,
		char_col, varchar_col, tinytext_col, text_col, mediumtext_col, longtext_col,
		binary_col, varbinary_col, tinyblob_col, blob_col, mediumblob_col, longblob_col,
		enum_col, set_col,
		date_col, datetime_col, timestamp_col, time_col, year_col,
		bit_col, json_col
		FROM all_datatypes`})
	require.NoError(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		res := make(map[string]any)
		err := rows.MapScan(res)
		require.NoError(t, err)

		// Verify that all expected columns are present
		expectedCols := []string{
			"tinyint_col", "smallint_col", "mediumint_col", "int_col", "bigint_col",
			"boolean_col", "float_col", "double_col", "decimal_col",
			"char_col", "varchar_col", "tinytext_col", "text_col", "mediumtext_col", "longtext_col",
			"binary_col", "varbinary_col", "tinyblob_col", "blob_col", "mediumblob_col", "longblob_col",
			"enum_col", "set_col",
			"date_col", "datetime_col", "timestamp_col", "time_col", "year_col",
			"bit_col", "json_col",
		}

		for _, col := range expectedCols {
			_, exists := res[col]
			require.True(t, exists, "Column %s should be present in result", col)
		}

		count++
	}
	require.NoError(t, rows.Err())
	require.Equal(t, count, 3)
}

func acquireTestMySQL(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "mysql")
	conn, err := drivers.Open("mysql", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
