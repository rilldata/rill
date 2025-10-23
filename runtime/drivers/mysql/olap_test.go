package mysql_test

import (
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
	tests := []struct {
		query   string
		args    []any
		result  map[string]any
		scanVal any
	}{
		// Integer types: TINYINT, SMALLINT, MEDIUMINT, INT, BIGINT
		{
			"SELECT tinyint_col FROM all_datatypes WHERE tinyint_col = 127",
			nil,
			map[string]any{"tinyint_col": int64(127)},
			ptr(int64(0)),
		},
		{
			"SELECT smallint_col FROM all_datatypes WHERE smallint_col = 32767",
			nil,
			map[string]any{"smallint_col": int64(32767)},
			ptr(int64(0)),
		},
		{
			"SELECT mediumint_col FROM all_datatypes WHERE mediumint_col = 8388607",
			nil,
			map[string]any{"mediumint_col": int64(8388607)},
			ptr(int64(0)),
		},
		{
			"SELECT int_col FROM all_datatypes WHERE int_col = ?",
			[]any{2147483647},
			map[string]any{"int_col": int64(2147483647)},
			ptr(int64(0)),
		},
		{
			"SELECT bigint_col FROM all_datatypes WHERE bigint_col = 999999999999999999",
			nil,
			map[string]any{"bigint_col": int64(999999999999999999)},
			ptr(int64(0)),
		},
		// Boolean type
		{
			"SELECT boolean_col FROM all_datatypes WHERE boolean_col = 1",
			nil,
			map[string]any{"boolean_col": int64(1)},
			ptr(int64(0)),
		},
		// Floating point types: FLOAT, DOUBLE
		{
			"SELECT float_col FROM all_datatypes WHERE float_col = 1.1",
			nil,
			map[string]any{"float_col": float64(1.1)},
			ptr(float64(0)),
		},
		{
			"SELECT double_col FROM all_datatypes WHERE double_col = 2.2",
			nil,
			map[string]any{"double_col": float64(2.2)},
			ptr(float64(0)),
		},
		// Decimal type
		{
			"SELECT decimal_col FROM all_datatypes WHERE decimal_col = 3.3",
			nil,
			map[string]any{"decimal_col": []uint8{51, 46, 51, 48}},
			ptr([]uint8{}),
		},
		// String types: CHAR, VARCHAR, TINYTEXT, TEXT, MEDIUMTEXT, LONGTEXT
		{
			"SELECT char_col FROM all_datatypes WHERE char_col = 'C'",
			nil,
			map[string]any{"char_col": []uint8{67}},
			ptr([]uint8{}),
		},
		// {
		// 	"SELECT varchar_col FROM all_datatypes WHERE varchar_col = 'VarChar'",
		// 	nil,
		// 	map[string]any{"varchar_col": []uint8{86, 97, 114, 67, 104, 97, 114}},
		// },
		// {
		// 	"SELECT tinytext_col FROM all_datatypes WHERE tinytext_col = 'Tiny Text'",
		// 	nil,
		// 	map[string]any{"tinytext_col": []uint8{84, 105, 110, 121, 32, 84, 101, 120, 116}},
		// },
		// {
		// 	"SELECT text_col FROM all_datatypes WHERE text_col = 'Text'",
		// 	nil,
		// 	map[string]any{"text_col": []uint8{84, 101, 120, 116}},
		// },
		// {
		// 	"SELECT mediumtext_col FROM all_datatypes WHERE mediumtext_col = 'Medium Text'",
		// 	nil,
		// 	map[string]any{"mediumtext_col": []uint8{77, 101, 100, 105, 117, 109, 32, 84, 101, 120, 116}},
		// },
		// {
		// 	"SELECT longtext_col FROM all_datatypes WHERE longtext_col = 'Long text content'",
		// 	nil,
		// 	map[string]any{"longtext_col": []uint8{76, 111, 110, 103, 32, 116, 101, 120, 116, 32, 99, 111, 110, 116, 101, 110, 116}},
		// },
		// // Binary types: BINARY, VARBINARY, TINYBLOB, BLOB, MEDIUMBLOB, LONGBLOB
		// {
		// 	"SELECT binary_col FROM all_datatypes WHERE binary_col = 'Binary'",
		// 	nil,
		// 	map[string]any{"binary_col": []uint8{66, 105, 110, 97, 114, 121, 0, 0, 0, 0}},
		// },
		// {
		// 	"SELECT varbinary_col FROM all_datatypes WHERE varbinary_col = 'VarBinary'",
		// 	nil,
		// 	map[string]any{"varbinary_col": []uint8{86, 97, 114, 66, 105, 110, 97, 114, 121}},
		// },
		// {
		// 	"SELECT tinyblob_col FROM all_datatypes WHERE tinyblob_col = 'Tiny Blob'",
		// 	nil,
		// 	map[string]any{"tinyblob_col": []uint8{84, 105, 110, 121, 32, 66, 108, 111, 98}},
		// },
		// {
		// 	"SELECT blob_col FROM all_datatypes WHERE blob_col = 'Blob'",
		// 	nil,
		// 	map[string]any{"blob_col": []uint8{66, 108, 111, 98}},
		// },
		// {
		// 	"SELECT mediumblob_col FROM all_datatypes WHERE mediumblob_col = 'Medium Blob'",
		// 	nil,
		// 	map[string]any{"mediumblob_col": []uint8{77, 101, 100, 105, 117, 109, 32, 66, 108, 111, 98}},
		// },
		// {
		// 	"SELECT longblob_col FROM all_datatypes WHERE longblob_col = 'Long Blob'",
		// 	nil,
		// 	map[string]any{"longblob_col": []uint8{76, 111, 110, 103, 32, 66, 108, 111, 98}},
		// },
		// // Enum and Set types
		// {
		// 	"SELECT enum_col FROM all_datatypes WHERE enum_col = 'medium'",
		// 	nil,
		// 	map[string]any{"enum_col": []uint8{109, 101, 100, 105, 117, 109}},
		// },
		// {
		// 	"SELECT set_col FROM all_datatypes WHERE set_col = 'a,b'",
		// 	nil,
		// 	map[string]any{"set_col": []uint8{97, 44, 98}},
		// },
		// // Date and Time types: DATE, TIME, DATETIME, TIMESTAMP, YEAR
		// {
		// 	"SELECT date_col FROM all_datatypes WHERE date_col = '2024-02-14'",
		// 	nil,
		// 	map[string]any{"date_col": []uint8{50, 48, 50, 52, 45, 48, 50, 45, 49, 52}},
		// },
		// {
		// 	"SELECT time_col FROM all_datatypes WHERE time_col = '12:34:56'",
		// 	nil,
		// 	map[string]any{"time_col": []uint8{49, 50, 58, 51, 52, 58, 53, 54}},
		// },
		// {
		// 	"SELECT datetime_col FROM all_datatypes WHERE datetime_col = '2025-02-14 12:34:56'",
		// 	nil,
		// 	map[string]any{"datetime_col": []uint8{50, 48, 50, 53, 45, 48, 50, 45, 49, 52, 32, 49, 50, 58, 51, 52, 58, 53, 54}},
		// },
		// {
		// 	"SELECT timestamp_col FROM all_datatypes WHERE timestamp_col = '2025-02-14 12:34:56'",
		// 	nil,
		// 	map[string]any{"timestamp_col": []uint8{50, 48, 50, 53, 45, 48, 50, 45, 49, 52, 32, 49, 50, 58, 51, 52, 58, 53, 54}},
		// },
		// {
		// 	"SELECT year_col FROM all_datatypes WHERE year_col = 2024",
		// 	nil,
		// 	map[string]any{"year_col": []uint8{50, 48, 50, 52}},
		// },
		// // BIT type
		// {
		// 	"SELECT bit_col FROM all_datatypes WHERE bit_col = B'1'",
		// 	nil,
		// 	map[string]any{"bit_col": []uint8{1}},
		// },
		// // JSON type
		// {
		// 	"SELECT json_col FROM all_datatypes WHERE JSON_EXTRACT(json_col, '$.key') = 'value'",
		// 	nil,
		// 	map[string]any{"json_col": []uint8{123, 34, 107, 101, 121, 34, 58, 32, 34, 118, 97, 108, 117, 101, 34, 125}},
		// },
		// // Geometry types
		// {
		// 	"SELECT geometry_col FROM all_datatypes WHERE geometry_col IS NOT NULL",
		// 	nil,
		// 	map[string]any{"geometry_col": []uint8{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63}},
		// },
		// {
		// 	"SELECT linestring_col FROM all_datatypes WHERE linestring_col IS NOT NULL",
		// 	nil,
		// 	map[string]any{"linestring_col": []uint8{0, 0, 0, 0, 1, 2, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63}},
		// },
		// {
		// 	"SELECT polygon_col FROM all_datatypes WHERE polygon_col IS NOT NULL",
		// 	nil,
		// 	map[string]any{"polygon_col": []uint8{0, 0, 0, 0, 1, 3, 0, 0, 0, 1, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		// },
		// {
		// 	"SELECT multipoint_col FROM all_datatypes WHERE multipoint_col IS NOT NULL",
		// 	nil,
		// 	map[string]any{"multipoint_col": []uint8{0, 0, 0, 0, 1, 4, 0, 0, 0, 2, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 64}},
		// },
		// {
		// 	"SELECT multilinestring_col FROM all_datatypes WHERE multilinestring_col IS NOT NULL",
		// 	nil,
		// 	map[string]any{"multilinestring_col": []uint8{0, 0, 0, 0, 1, 5, 0, 0, 0, 2, 0, 0, 0, 1, 2, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63, 1, 2, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 64}},
		// },
		// {
		// 	"SELECT multipolygon_col FROM all_datatypes WHERE multipolygon_col IS NOT NULL",
		// 	nil,
		// 	map[string]any{"multipolygon_col": []uint8{0, 0, 0, 0, 1, 6, 0, 0, 0, 1, 0, 0, 0, 1, 3, 0, 0, 0, 1, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		// },
		// {
		// 	"SELECT geometrycollection_col FROM all_datatypes WHERE geometrycollection_col IS NOT NULL",
		// 	nil,
		// 	map[string]any{"geometrycollection_col": []uint8{0, 0, 0, 0, 1, 7, 0, 0, 0, 2, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63, 1, 2, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 240, 63, 0, 0, 0, 0, 0, 0, 240, 63}},
		// },
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
			require.NoError(t, rows.Close())
		})
	}

	// Same tests using Scan instead of MapScan
	for _, test := range tests {
		t.Run(test.query+"_scan", func(t *testing.T) {
			rows, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query, Args: test.args})
			require.NoError(t, err)
			defer rows.Close()

			// Get column name from schema
			require.Len(t, rows.Schema.Fields, 1)
			colName := rows.Schema.Fields[0].Name

			for rows.Next() {
				err = rows.Scan(&test.scanVal)
				require.NoError(t, err)
				require.Equal(t, test.result[colName], test.scanVal)
			}
			require.NoError(t, rows.Err())
		})
	}
}

func TestEmptyRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestMySQL(t)
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

func TestExec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	_, olap := acquireTestMySQL(t)

	// dry run with SELECT query
	err := olap.Exec(t.Context(), &drivers.Statement{Query: "SELECT 1", DryRun: true})
	require.NoError(t, err)

	// create table
	err = olap.Exec(t.Context(), &drivers.Statement{Query: "CREATE TABLE IF NOT EXISTS exec_test (id INT, name VARCHAR(255))"})
	require.NoError(t, err)

	// drop table
	err = olap.Exec(t.Context(), &drivers.Statement{Query: "DROP TABLE IF EXISTS exec_test"})
	require.NoError(t, err)
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

func ptr[T any](v T) *T {
	return &v
}
