package sqlserver_test

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

func TestOLAP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestSQLServer(t)
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
		// BIT type (maps to bool)
		{
			"SELECT bit_col FROM all_datatypes WHERE bit_col = 1",
			nil,
			map[string]any{"bit_col": true},
		},
		// Integer types: TINYINT, SMALLINT, INT, BIGINT
		{
			"SELECT tinyint_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"tinyint_col": int64(127)},
		},
		{
			"SELECT smallint_col FROM all_datatypes WHERE smallint_col = 32767",
			nil,
			map[string]any{"smallint_col": int64(32767)},
		},
		{
			"SELECT int_col FROM all_datatypes WHERE int_col = @p1",
			[]any{2147483647},
			map[string]any{"int_col": int64(2147483647)},
		},
		{
			"SELECT bigint_col FROM all_datatypes WHERE bigint_col = 999999999999999999",
			nil,
			map[string]any{"bigint_col": int64(999999999999999999)},
		},
		// Floating point types: REAL, FLOAT
		{
			"SELECT real_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"real_col": float64(1.5)},
		},
		{
			"SELECT float_col FROM all_datatypes WHERE float_col = 2.5",
			nil,
			map[string]any{"float_col": float64(2.5)},
		},
		// String types: CHAR, VARCHAR, TEXT, NCHAR, NVARCHAR, NTEXT
		{
			"SELECT varchar_col FROM all_datatypes WHERE varchar_col = 'VarChar'",
			nil,
			map[string]any{"varchar_col": "VarChar"},
		},
		{
			"SELECT nvarchar_col FROM all_datatypes WHERE nvarchar_col = N'NVarChar'",
			nil,
			map[string]any{"nvarchar_col": "NVarChar"},
		},
		// UNIQUEIDENTIFIER
		{
			"SELECT uniqueidentifier_col FROM all_datatypes WHERE uniqueidentifier_col = '6F9619FF-8B86-D011-B42D-00CF4FC964FF'",
			nil,
			map[string]any{"uniqueidentifier_col": []byte{0xFF, 0x19, 0x96, 0x6F, 0x86, 0x8B, 0x11, 0xD0, 0xB4, 0x2D, 0x00, 0xCF, 0x4F, 0xC9, 0x64, 0xFF}},
		},
		// XML type
		{
			"SELECT xml_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"xml_col": "<root><element>value</element></root>"},
		},
		// NULL
		{
			"SELECT NULL AS null_col",
			nil,
			map[string]any{"null_col": nil},
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
			query: "SELECT tinyint_col, smallint_col, int_col, bigint_col FROM all_datatypes WHERE int_col = @p1",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var tinyintCol, smallintCol, intCol, bigintCol int64
				require.True(t, rows.Next())
				err := rows.Scan(&tinyintCol, &smallintCol, &intCol, &bigintCol)
				require.NoError(t, err)
				require.Equal(t, int64(127), tinyintCol)
				require.Equal(t, int64(32767), smallintCol)
				require.Equal(t, int64(2147483647), intCol)
				require.Equal(t, int64(999999999999999999), bigintCol)
				return nil
			},
		},
		{
			name:  "Floating point types scan",
			query: "SELECT real_col, float_col FROM all_datatypes WHERE int_col = @p1",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var realCol, floatCol float64
				require.True(t, rows.Next())
				err := rows.Scan(&realCol, &floatCol)
				require.NoError(t, err)
				require.InDelta(t, float64(1.5), realCol, 0.01)
				require.Equal(t, float64(2.5), floatCol)
				return nil
			},
		},
		{
			name:  "String types scan",
			query: "SELECT varchar_col, nvarchar_col FROM all_datatypes WHERE int_col = @p1",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var varcharCol, nvarcharCol string
				require.True(t, rows.Next())
				err := rows.Scan(&varcharCol, &nvarcharCol)
				require.NoError(t, err)
				require.Equal(t, "VarChar", varcharCol)
				require.Equal(t, "NVarChar", nvarcharCol)
				return nil
			},
		},
		{
			name:  "Date and Time types scan",
			query: "SELECT date_col, time_col, datetime_col FROM all_datatypes WHERE int_col = @p1",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var dateCol, timeCol, datetimeCol time.Time
				require.True(t, rows.Next())
				err := rows.Scan(&dateCol, &timeCol, &datetimeCol)
				require.NoError(t, err)
				require.Equal(t, 2024, dateCol.Year())
				require.Equal(t, time.February, dateCol.Month())
				require.Equal(t, 14, dateCol.Day())
				require.Equal(t, 12, timeCol.Hour())
				require.Equal(t, 34, timeCol.Minute())
				require.Equal(t, 56, timeCol.Second())
				return nil
			},
		},
		{
			name:  "Boolean (BIT) type scan",
			query: "SELECT bit_col FROM all_datatypes WHERE int_col = @p1",
			args:  []any{2147483647},
			scanFunc: func(rows drivers.Rows) error {
				var bitCol bool
				require.True(t, rows.Next())
				err := rows.Scan(&bitCol)
				require.NoError(t, err)
				require.True(t, bitCol)
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
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT TOP 0 int_col, float_col FROM all_datatypes"})
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
	err = olap.Exec(t.Context(), &drivers.Statement{Query: "CREATE TABLE exec_test (id INT, name NVARCHAR(255))"})
	require.NoError(t, err)

	// drop table
	err = olap.Exec(t.Context(), &drivers.Statement{Query: "DROP TABLE exec_test"})
	require.NoError(t, err)
}

func testFullTableScan(t *testing.T, olap drivers.OLAPStore) {
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: `SELECT
		bit_col, tinyint_col, smallint_col, int_col, bigint_col,
		real_col, float_col,
		decimal_col, numeric_col, money_col, smallmoney_col,
		char_col, varchar_col, text_col, nchar_col, nvarchar_col, ntext_col,
		binary_col, varbinary_col,
		date_col, time_col, datetime_col, datetime2_col, smalldatetime_col, datetimeoffset_col,
		uniqueidentifier_col, xml_col
		FROM all_datatypes`})
	require.NoError(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		res := make(map[string]any)
		err := rows.MapScan(res)
		require.NoError(t, err)

		expectedCols := []string{
			"bit_col", "tinyint_col", "smallint_col", "int_col", "bigint_col",
			"real_col", "float_col",
			"decimal_col", "numeric_col", "money_col", "smallmoney_col",
			"char_col", "varchar_col", "text_col", "nchar_col", "nvarchar_col", "ntext_col",
			"binary_col", "varbinary_col",
			"date_col", "time_col", "datetime_col", "datetime2_col", "smalldatetime_col", "datetimeoffset_col",
			"uniqueidentifier_col", "xml_col",
		}

		for _, col := range expectedCols {
			_, exists := res[col]
			require.True(t, exists, "Column %s should be present in result", col)
		}

		count++
	}
	require.NoError(t, rows.Err())
	require.Equal(t, 3, count)
}

func acquireTestSQLServer(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "sqlserver")
	conn, err := drivers.Open("sqlserver", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
