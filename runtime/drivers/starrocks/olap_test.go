package starrocks

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/starrocks/teststarrocks"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStarRocksOLAP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dsn := teststarrocks.StartWithData(t)

	conn, err := driver{}.Open("default", map[string]any{
		"dsn": dsn,
	}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	// Basic type tests
	t.Run("VarcharNotBinary", func(t *testing.T) {
		testVarcharNotBinary(t, olap)
	})

	t.Run("NullHandling", func(t *testing.T) {
		testNullHandling(t, olap)
	})

	t.Run("NumericTypes", func(t *testing.T) {
		testNumericTypes(t, olap)
	})

	// All types tests
	t.Run("AllBasicTypes", func(t *testing.T) {
		testAllBasicTypes(t, olap)
	})

	t.Run("DateTimeTypes", func(t *testing.T) {
		testDateTimeTypes(t, olap)
	})

	t.Run("StringTypes", func(t *testing.T) {
		testStringTypes(t, olap)
	})

	t.Run("BinaryTypes", func(t *testing.T) {
		testBinaryTypes(t, olap)
	})

	t.Run("AggregateTypes", func(t *testing.T) {
		testAggregateTypes(t, olap)
	})

	t.Run("UnicodeStrings", func(t *testing.T) {
		testUnicodeStrings(t, olap)
	})

	t.Run("JSONType", func(t *testing.T) {
		testJSONType(t, olap)
	})

	// API tests
	t.Run("DryRun", func(t *testing.T) {
		testDryRun(t, olap)
	})

	t.Run("Exec", func(t *testing.T) {
		testExec(t, olap)
	})

	t.Run("QuerySchema", func(t *testing.T) {
		testQuerySchema(t, olap)
	})

	// Result set tests
	t.Run("EmptyResultSet", func(t *testing.T) {
		testEmptyResultSet(t, olap)
	})

	t.Run("MultipleRows", func(t *testing.T) {
		testMultipleRows(t, olap)
	})

	// Boundary and special cases
	t.Run("BoundaryValues", func(t *testing.T) {
		testBoundaryValues(t, olap)
	})

	t.Run("NullHandlingDetailed", func(t *testing.T) {
		testNullHandlingDetailed(t, olap)
	})

	t.Run("NegativeValues", func(t *testing.T) {
		testNegativeValues(t, olap)
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		testSpecialCharacters(t, olap)
	})

	// Complex types
	t.Run("ComplexTypes", func(t *testing.T) {
		testComplexTypes(t, olap)
	})

	// Error cases
	t.Run("ErrorCases", func(t *testing.T) {
		testErrorCases(t, olap)
	})

	// Other tests
	t.Run("ParameterBinding", func(t *testing.T) {
		testParameterBinding(t, olap)
	})

	t.Run("SchemaValidation", func(t *testing.T) {
		testSchemaValidation(t, olap)
	})

	t.Run("AggregateFunctions", func(t *testing.T) {
		testAggregateFunctions(t, olap)
	})

	// Output all types and values
	t.Run("AllTypesOutput", func(t *testing.T) {
		testAllTypesOutput(t, olap)
	})

	// High-precision DECIMAL test (DECIMAL32, DECIMAL64, DECIMAL128)
	t.Run("DecimalPrecision", func(t *testing.T) {
		testDecimalPrecision(t, olap)
	})

	t.Run("LoadDDL", func(t *testing.T) {
		testLoadDDL(t, conn)
	})
}

func testVarcharNotBinary(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT 'hello' AS str_col, 'world' AS str_col2",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// VARCHAR should be string, not []byte
	strVal, ok := row["str_col"].(string)
	require.True(t, ok, "expected string type, got %T", row["str_col"])
	require.Equal(t, "hello", strVal)
}

func testNullHandling(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT NULL AS null_col, 'value' AS str_col",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	require.Nil(t, row["null_col"])
	require.Equal(t, "value", row["str_col"])
}

func testNumericTypes(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT 42 AS int_col, 3.14 AS float_col, TRUE AS bool_col",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// Check types are correct (not []byte)
	// StarRocks returns small integers as TINYINT/SMALLINT via MySQL protocol
	intVal := row["int_col"]
	_, isByte := intVal.([]byte)
	require.False(t, isByte, "int_col should not be []byte, got %T", intVal)

	// Accept any integer type (int16, int32, int64)
	switch intVal.(type) {
	case int16, int32, int64:
		// OK - valid integer type
	default:
		t.Errorf("expected int type, got %T", intVal)
	}
}

func testAllBasicTypes(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.all_types WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// Verify each type is correctly converted
	require.Equal(t, int32(1), row["id"])

	// Boolean - MySQL protocol returns BOOLEAN as TINYINT
	// So it might be bool or int16 depending on driver behavior
	boolVal := row["bool_col"]
	_, isBoolByte := boolVal.([]byte)
	require.False(t, isBoolByte, "bool_col should not be []byte, got %T", boolVal)
	switch v := boolVal.(type) {
	case bool:
		require.True(t, v)
	case int16:
		require.Equal(t, int16(1), v) // 1 = true
	default:
		t.Errorf("expected bool or int16 for BOOLEAN, got %T", boolVal)
	}

	// Integer types
	_, ok := row["tinyint_col"].(int16) // TINYINT maps to int16 (via NullInt16)
	require.True(t, ok, "expected int16 for tinyint, got %T", row["tinyint_col"])

	_, ok = row["smallint_col"].(int16)
	require.True(t, ok, "expected int16 type, got %T", row["smallint_col"])

	_, ok = row["int_col"].(int32)
	require.True(t, ok, "expected int32 type, got %T", row["int_col"])

	_, ok = row["bigint_col"].(int64)
	require.True(t, ok, "expected int64 type, got %T", row["bigint_col"])

	// Float types
	_, ok = row["float_col"].(float64)
	require.True(t, ok, "expected float64 type, got %T", row["float_col"])

	_, ok = row["double_col"].(float64)
	require.True(t, ok, "expected float64 type, got %T", row["double_col"])

	// Decimal type - stored as string to preserve precision (same as MySQL driver)
	decimalVal, ok := row["decimal_col"].(string)
	require.True(t, ok, "expected string type for DECIMAL, got %T", row["decimal_col"])
	require.Equal(t, "12345.6789", decimalVal, "decimal value mismatch")

	// String types - should NOT be []byte
	charVal, ok := row["char_col"].(string)
	require.True(t, ok, "expected string type for CHAR, got %T", row["char_col"])
	require.Contains(t, charVal, "char_val")

	varcharVal, ok := row["varchar_col"].(string)
	require.True(t, ok, "expected string type for VARCHAR, got %T", row["varchar_col"])
	require.Equal(t, "varchar_value", varcharVal)

	stringVal, ok := row["string_col"].(string)
	require.True(t, ok, "expected string type for STRING, got %T", row["string_col"])
	require.Equal(t, "string_value", stringVal)
}

func testDateTimeTypes(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT date_col, datetime_col FROM test_db.all_types WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// DATE - should be string (StarRocks returns as string via MySQL protocol)
	dateVal, ok := row["date_col"].(string)
	require.True(t, ok, "expected string type for DATE, got %T", row["date_col"])
	require.Contains(t, dateVal, "2024-01-15")

	// DATETIME - should be string (MySQL driver returns as string without parseTime=true)
	datetimeVal, ok := row["datetime_col"].(string)
	require.True(t, ok, "expected string type for DATETIME, got %T", row["datetime_col"])
	require.Contains(t, datetimeVal, "2024-01-15")
	require.Contains(t, datetimeVal, "10:30:00")
}

func testStringTypes(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT char_col, varchar_col, string_col FROM test_db.all_types WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// All string types should be Go string, not []byte
	for _, col := range []string{"char_col", "varchar_col", "string_col"} {
		val := row[col]
		_, isByte := val.([]byte)
		require.False(t, isByte, "%s should not be []byte, got %T", col, val)

		_, isString := val.(string)
		require.True(t, isString, "%s should be string, got %T", col, val)
	}
}

func testBinaryTypes(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, binary_col, blob_col FROM test_db.binary_types WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// Binary types might be []byte or base64 string depending on driver
	// Just verify they're not nil for non-null values
	require.NotNil(t, row["binary_col"])
	require.NotNil(t, row["blob_col"])
}

func testAggregateTypes(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	// Query aggregate table with HLL and BITMAP
	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, dt, hll_cardinality(hll_col) as hll_count, bitmap_count(bitmap_col) as bitmap_count, count_col FROM test_db.aggregate_types WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// Verify aggregate results
	require.Equal(t, int32(1), row["id"])
	require.NotNil(t, row["hll_count"])
	require.NotNil(t, row["bitmap_count"])
}

func testUnicodeStrings(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT ascii_col, unicode_col, emoji_col, korean_col, chinese_col, japanese_col FROM test_db.string_encoding_test WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// Verify Unicode strings are correctly handled
	asciiVal, ok := row["ascii_col"].(string)
	require.True(t, ok, "expected string type, got %T", row["ascii_col"])
	require.Equal(t, "Hello World", asciiVal)

	unicodeVal, ok := row["unicode_col"].(string)
	require.True(t, ok, "expected string type, got %T", row["unicode_col"])
	require.Equal(t, "Héllo Wörld", unicodeVal)

	emojiVal, ok := row["emoji_col"].(string)
	require.True(t, ok, "expected string type, got %T", row["emoji_col"])
	require.Equal(t, "😀🎉🚀", emojiVal)

	koreanVal, ok := row["korean_col"].(string)
	require.True(t, ok, "expected string type, got %T", row["korean_col"])
	require.Equal(t, "안녕하세요", koreanVal)

	chineseVal, ok := row["chinese_col"].(string)
	require.True(t, ok, "expected string type, got %T", row["chinese_col"])
	require.Equal(t, "你好世界", chineseVal)

	japaneseVal, ok := row["japanese_col"].(string)
	require.True(t, ok, "expected string type, got %T", row["japanese_col"])
	require.Equal(t, "こんにちは", japaneseVal)
}

func testJSONType(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT json_col FROM test_db.all_types WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// JSON should be returned as string
	jsonVal, ok := row["json_col"].(string)
	require.True(t, ok, "expected string type for JSON, got %T", row["json_col"])
	require.Contains(t, jsonVal, "key")
	require.Contains(t, jsonVal, "value")
}

// ============================================================
// DryRun tests
// ============================================================
func testDryRun(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	// Query DryRun - valid query
	t.Run("QueryDryRunValid", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query:  "SELECT * FROM test_db.all_types",
			DryRun: true,
		})
		require.NoError(t, err)
		require.Nil(t, res, "DryRun should return nil result")
	})

	// Query DryRun - invalid query
	t.Run("QueryDryRunInvalid", func(t *testing.T) {
		_, err := olap.Query(ctx, &drivers.Statement{
			Query:  "SELECT * FROM nonexistent_table",
			DryRun: true,
		})
		require.Error(t, err)
	})

	// Exec DryRun
	t.Run("ExecDryRun", func(t *testing.T) {
		err := olap.Exec(ctx, &drivers.Statement{
			Query:  "SELECT 1",
			DryRun: true,
		})
		require.NoError(t, err)
	})
}

// ============================================================
// Exec tests
// ============================================================
func testExec(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	// Basic exec
	t.Run("BasicExec", func(t *testing.T) {
		err := olap.Exec(ctx, &drivers.Statement{
			Query: "SELECT 1",
		})
		require.NoError(t, err)
	})

	// Error case
	t.Run("ExecError", func(t *testing.T) {
		err := olap.Exec(ctx, &drivers.Statement{
			Query: "INVALID SQL SYNTAX",
		})
		require.Error(t, err)
	})
}

// ============================================================
// QuerySchema tests
// ============================================================
func testQuerySchema(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	schema, err := olap.QuerySchema(ctx, "SELECT id, varchar_col, int_col FROM test_db.all_types", nil)
	require.NoError(t, err)
	require.NotNil(t, schema)
	require.Len(t, schema.Fields, 3)

	// Verify schema types
	require.Equal(t, "id", schema.Fields[0].Name)
	require.Equal(t, runtimev1.Type_CODE_INT32, schema.Fields[0].Type.Code)

	require.Equal(t, "varchar_col", schema.Fields[1].Name)
	require.Equal(t, runtimev1.Type_CODE_STRING, schema.Fields[1].Type.Code)

	require.Equal(t, "int_col", schema.Fields[2].Name)
	require.Equal(t, runtimev1.Type_CODE_INT32, schema.Fields[2].Type.Code)
}

// ============================================================
// Empty result set test
// ============================================================
func testEmptyResultSet(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.all_types WHERE id = -999",
	})
	require.NoError(t, err)
	defer res.Close()

	// Schema should exist
	require.NotNil(t, res.Schema)
	require.Greater(t, len(res.Schema.Fields), 0)

	// No rows
	require.False(t, res.Next())
	require.NoError(t, res.Err())
}

// ============================================================
// Multiple rows test
// ============================================================
func testMultipleRows(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, varchar_col FROM test_db.all_types ORDER BY id",
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]any
	for res.Next() {
		row := make(map[string]any)
		err := res.MapScan(row)
		require.NoError(t, err)
		rows = append(rows, row)
	}
	require.NoError(t, res.Err())

	require.Len(t, rows, 3, "expected 3 rows")
	require.Equal(t, int32(1), rows[0]["id"])
	require.Equal(t, int32(2), rows[1]["id"])
	require.Equal(t, int32(3), rows[2]["id"])
}

// ============================================================
// Boundary values test
// ============================================================
func testBoundaryValues(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.boundary_values WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())
	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// TINYINT boundary
	require.Equal(t, int16(-128), row["tinyint_min"])
	require.Equal(t, int16(127), row["tinyint_max"])

	// SMALLINT boundary
	require.Equal(t, int16(-32768), row["smallint_min"])
	require.Equal(t, int16(32767), row["smallint_max"])

	// INT boundary
	require.Equal(t, int32(-2147483648), row["int_min"])
	require.Equal(t, int32(2147483647), row["int_max"])

	// BIGINT boundary
	require.Equal(t, int64(-9223372036854775808), row["bigint_min"])
	require.Equal(t, int64(9223372036854775807), row["bigint_max"])

	// Empty string
	require.Equal(t, "", row["empty_string"])

	// Whitespace string
	require.Equal(t, "   ", row["whitespace_string"])
}

// ============================================================
// Detailed NULL handling test
// ============================================================
func testNullHandlingDetailed(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	// Row with all columns NULL (id=3)
	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.all_types WHERE id = 3",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())
	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// Only id is non-null
	require.Equal(t, int32(3), row["id"])

	// All other columns should be NULL
	nullColumns := []string{
		"bool_col", "tinyint_col", "smallint_col", "int_col", "bigint_col",
		"float_col", "double_col", "decimal_col",
		"char_col", "varchar_col", "string_col", "date_col", "datetime_col", "json_col",
	}
	for _, col := range nullColumns {
		require.Nil(t, row[col], "expected %s to be nil", col)
	}
}

// ============================================================
// Negative values test
// ============================================================
func testNegativeValues(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.all_types WHERE id = 2",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())
	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// Negative integers
	require.Equal(t, int16(-128), row["tinyint_col"])
	require.Equal(t, int16(-32768), row["smallint_col"])
	require.Equal(t, int32(-2147483648), row["int_col"])
	require.Equal(t, int64(-9223372036854775808), row["bigint_col"])

	// Negative floats
	floatVal, _ := row["float_col"].(float64)
	require.Less(t, floatVal, float64(0))

	doubleVal, _ := row["double_col"].(float64)
	require.Less(t, doubleVal, float64(0))

	// DECIMAL is returned as string to preserve precision
	decimalVal, ok := row["decimal_col"].(string)
	require.True(t, ok, "expected string type for DECIMAL, got %T", row["decimal_col"])
	require.True(t, len(decimalVal) > 0 && decimalVal[0] == '-', "expected negative decimal value")

	// false boolean
	boolVal := row["bool_col"]
	switch v := boolVal.(type) {
	case bool:
		require.False(t, v)
	case int16:
		require.Equal(t, int16(0), v)
	}
}

// ============================================================
// Special characters test
// ============================================================
func testSpecialCharacters(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.special_chars WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())
	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// Quotes
	quoteVal, ok := row["quote_col"].(string)
	require.True(t, ok)
	require.Contains(t, quoteVal, "'")
	require.Contains(t, quoteVal, "\"")

	// Emoji
	emojiVal, ok := row["emoji_col"].(string)
	require.True(t, ok)
	require.Contains(t, emojiVal, "😀")

	// SQL injection string (should be stored as-is)
	sqlVal, ok := row["sql_injection_col"].(string)
	require.True(t, ok)
	require.Contains(t, sqlVal, "DROP TABLE")
}

// ============================================================
// Complex types test (ARRAY, MAP, STRUCT)
// ============================================================
func testComplexTypes(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, array_col, map_col, struct_col FROM test_db.complex_types WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())
	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	// ARRAY - handled by default case (*any)
	require.NotNil(t, row["array_col"])
	t.Logf("array_col type: %T, value: %v", row["array_col"], row["array_col"])

	// MAP - handled by default case
	require.NotNil(t, row["map_col"])
	t.Logf("map_col type: %T, value: %v", row["map_col"], row["map_col"])

	// STRUCT - handled by default case
	require.NotNil(t, row["struct_col"])
	t.Logf("struct_col type: %T, value: %v", row["struct_col"], row["struct_col"])
}

// ============================================================
// Error cases test
// ============================================================
func testErrorCases(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	t.Run("NonexistentTable", func(t *testing.T) {
		_, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT * FROM nonexistent_db.nonexistent_table",
		})
		require.Error(t, err)
	})

	t.Run("SyntaxError", func(t *testing.T) {
		_, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELEC * FROM test_db.all_types",
		})
		require.Error(t, err)
	})

	t.Run("InvalidColumn", func(t *testing.T) {
		_, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT nonexistent_column FROM test_db.all_types",
		})
		require.Error(t, err)
	})
}

// ============================================================
// Parameter binding test
// ============================================================
func testParameterBinding(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, varchar_col FROM test_db.all_types WHERE id = ?",
		Args:  []any{1},
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())
	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)

	require.Equal(t, int32(1), row["id"])
	require.Equal(t, "varchar_value", row["varchar_col"])
}

// ============================================================
// Schema validation test
// ============================================================
func testSchemaValidation(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, bool_col, tinyint_col, smallint_col, int_col, bigint_col, float_col, double_col, decimal_col, varchar_col, date_col, datetime_col, json_col FROM test_db.all_types WHERE id = 1",
	})
	require.NoError(t, err)
	defer res.Close()

	schema := res.Schema
	require.NotNil(t, schema)

	// Debug: Log actual schema types returned by StarRocks
	t.Log("=== Actual Schema Types from StarRocks ===")
	for _, field := range schema.Fields {
		t.Logf("Column: %-15s Type.Code: %v", field.Name, field.Type.Code)
	}

	// Also log the raw DatabaseTypeName from the driver
	t.Log("=== Raw DatabaseTypeName from MySQL Driver ===")
	if starrocksRes, ok := res.Rows.(*starrocksRows); ok {
		for _, ct := range starrocksRes.colTypes {
			t.Logf("Column: %-15s DatabaseTypeName: %s", ct.Name(), ct.DatabaseTypeName())
		}
	}

	expectedTypes := map[string]runtimev1.Type_Code{
		"id":           runtimev1.Type_CODE_INT32,
		"bool_col":     runtimev1.Type_CODE_BOOL, // Note: MySQL protocol may report as TINYINT
		"tinyint_col":  runtimev1.Type_CODE_INT8,
		"smallint_col": runtimev1.Type_CODE_INT16,
		"int_col":      runtimev1.Type_CODE_INT32,
		"bigint_col":   runtimev1.Type_CODE_INT64,
		"float_col":    runtimev1.Type_CODE_FLOAT32,
		"double_col":   runtimev1.Type_CODE_FLOAT64,
		"decimal_col":  runtimev1.Type_CODE_STRING, // DECIMAL returns as string to preserve precision
		"varchar_col":  runtimev1.Type_CODE_STRING,
		"date_col":     runtimev1.Type_CODE_DATE,
		"datetime_col": runtimev1.Type_CODE_TIMESTAMP,
		"json_col":     runtimev1.Type_CODE_JSON,
	}

	for _, field := range schema.Fields {
		expectedCode, exists := expectedTypes[field.Name]
		if exists {
			// bool_col may be reported as TINYINT (CODE_INT8) by MySQL protocol
			if field.Name == "bool_col" {
				if field.Type.Code != runtimev1.Type_CODE_BOOL && field.Type.Code != runtimev1.Type_CODE_INT8 {
					t.Errorf("expected BOOL or INT8 for bool_col, got %v", field.Type.Code)
				}
				continue
			}
			// json_col may be reported as STRING by MySQL protocol
			if field.Name == "json_col" {
				if field.Type.Code != runtimev1.Type_CODE_JSON && field.Type.Code != runtimev1.Type_CODE_STRING {
					t.Errorf("expected JSON or STRING for json_col, got %v", field.Type.Code)
				}
				continue
			}
			require.Equal(t, expectedCode, field.Type.Code, "type mismatch for %s", field.Name)
		}
	}
}

// ============================================================
// Aggregate functions test
// ============================================================
func testAggregateFunctions(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	t.Run("COUNT", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT COUNT(*) as cnt FROM test_db.all_types",
		})
		require.NoError(t, err)
		defer res.Close()

		require.True(t, res.Next())
		row := make(map[string]any)
		err = res.MapScan(row)
		require.NoError(t, err)

		cnt, ok := row["cnt"].(int64)
		require.True(t, ok, "expected int64 for COUNT, got %T", row["cnt"])
		require.Equal(t, int64(3), cnt)
	})

	t.Run("SUM", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT SUM(tinyint_col) as sum_val FROM test_db.all_types",
		})
		require.NoError(t, err)
		defer res.Close()

		require.True(t, res.Next())
		row := make(map[string]any)
		err = res.MapScan(row)
		require.NoError(t, err)
		require.NotNil(t, row["sum_val"])
	})

	t.Run("AVG", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT AVG(int_col) as avg_val FROM test_db.all_types WHERE int_col IS NOT NULL",
		})
		require.NoError(t, err)
		defer res.Close()

		require.True(t, res.Next())
		row := make(map[string]any)
		err = res.MapScan(row)
		require.NoError(t, err)

		_, ok := row["avg_val"].(float64)
		require.True(t, ok, "expected float64 for AVG, got %T", row["avg_val"])
	})

	t.Run("MIN_MAX", func(t *testing.T) {
		res, err := olap.Query(ctx, &drivers.Statement{
			Query: "SELECT MIN(int_col) as min_val, MAX(int_col) as max_val FROM test_db.all_types",
		})
		require.NoError(t, err)
		defer res.Close()

		require.True(t, res.Next())
		row := make(map[string]any)
		err = res.MapScan(row)
		require.NoError(t, err)

		minVal, ok := row["min_val"].(int32)
		require.True(t, ok)
		require.Equal(t, int32(-2147483648), minVal)

		maxVal, ok := row["max_val"].(int32)
		require.True(t, ok)
		require.Equal(t, int32(2147483647), maxVal)
	})
}

// ============================================================
// All Types Output - prints all types and actual values
// ============================================================
func testAllTypesOutput(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	t.Log("================================================================================")
	t.Log("                    StarRocks Type → Go Return Value Mapping")
	t.Log("================================================================================")

	// 1. Basic types from all_types table
	t.Log("")
	t.Log("=== 1. Basic Types (test_db.all_types, id=1) ===")
	t.Log("--------------------------------------------------------------------------------")
	t.Logf("%-20s | %-20s | %-15s | %s", "Column", "StarRocks Type", "Go Type", "Value")
	t.Log("--------------------------------------------------------------------------------")

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.all_types WHERE id = 1",
	})
	require.NoError(t, err)

	// Get schema
	schema := res.Schema
	schemaMap := make(map[string]string)
	for _, field := range schema.Fields {
		schemaMap[field.Name] = field.Type.Code.String()
	}

	require.True(t, res.Next())
	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)
	res.Close()

	// Define column order and StarRocks types
	basicColumns := []struct {
		name   string
		srType string
	}{
		{"id", "INT"},
		{"bool_col", "BOOLEAN"},
		{"tinyint_col", "TINYINT"},
		{"smallint_col", "SMALLINT"},
		{"int_col", "INT"},
		{"bigint_col", "BIGINT"},
		{"largeint_col", "LARGEINT"},
		{"float_col", "FLOAT"},
		{"double_col", "DOUBLE"},
		{"decimal_col", "DECIMAL(18,4)"},
		{"char_col", "CHAR(10)"},
		{"varchar_col", "VARCHAR(255)"},
		{"string_col", "STRING"},
		{"date_col", "DATE"},
		{"datetime_col", "DATETIME"},
		{"json_col", "JSON"},
	}

	for _, col := range basicColumns {
		val := row[col.name]
		goType := fmt.Sprintf("%T", val)
		valStr := formatValue(val)
		t.Logf("%-20s | %-20s | %-15s | %s", col.name, col.srType, goType, valStr)
	}

	// 2. Complex types
	t.Log("")
	t.Log("=== 2. Complex Types (test_db.complex_types, id=1) ===")
	t.Log("--------------------------------------------------------------------------------")
	t.Logf("%-20s | %-20s | %-15s | %s", "Column", "StarRocks Type", "Go Type", "Value")
	t.Log("--------------------------------------------------------------------------------")

	res2, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, array_col, map_col, struct_col FROM test_db.complex_types WHERE id = 1",
	})
	require.NoError(t, err)
	require.True(t, res2.Next())
	row2 := make(map[string]any)
	err = res2.MapScan(row2)
	require.NoError(t, err)
	res2.Close()

	complexColumns := []struct {
		name   string
		srType string
	}{
		{"array_col", "ARRAY<INT>"},
		{"map_col", "MAP<STRING,INT>"},
		{"struct_col", "STRUCT<...>"},
	}

	for _, col := range complexColumns {
		val := row2[col.name]
		goType := fmt.Sprintf("%T", val)
		valStr := formatValue(val)
		t.Logf("%-20s | %-20s | %-15s | %s", col.name, col.srType, goType, valStr)
	}

	// 3. Binary types
	t.Log("")
	t.Log("=== 3. Binary Types (test_db.binary_types, id=1) ===")
	t.Log("--------------------------------------------------------------------------------")
	t.Logf("%-20s | %-20s | %-15s | %s", "Column", "StarRocks Type", "Go Type", "Value")
	t.Log("--------------------------------------------------------------------------------")

	res3, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, binary_col, blob_col FROM test_db.binary_types WHERE id = 1",
	})
	require.NoError(t, err)
	require.True(t, res3.Next())
	row3 := make(map[string]any)
	err = res3.MapScan(row3)
	require.NoError(t, err)
	res3.Close()

	binaryColumns := []struct {
		name   string
		srType string
	}{
		{"binary_col", "VARBINARY(255)"},
		{"blob_col", "VARBINARY(65535)"},
	}

	for _, col := range binaryColumns {
		val := row3[col.name]
		goType := fmt.Sprintf("%T", val)
		valStr := formatValue(val)
		t.Logf("%-20s | %-20s | %-15s | %s", col.name, col.srType, goType, valStr)
	}

	// 4. Aggregate types (with functions)
	t.Log("")
	t.Log("=== 4. Aggregate Types (test_db.aggregate_types) ===")
	t.Log("--------------------------------------------------------------------------------")
	t.Logf("%-20s | %-20s | %-15s | %s", "Column", "StarRocks Type", "Go Type", "Value")
	t.Log("--------------------------------------------------------------------------------")

	res4, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT id, dt, hll_cardinality(hll_col) as hll_result, bitmap_count(bitmap_col) as bitmap_result, count_col FROM test_db.aggregate_types WHERE id = 1",
	})
	require.NoError(t, err)
	require.True(t, res4.Next())
	row4 := make(map[string]any)
	err = res4.MapScan(row4)
	require.NoError(t, err)
	res4.Close()

	aggColumns := []struct {
		name   string
		srType string
	}{
		{"id", "INT"},
		{"dt", "DATE"},
		{"hll_result", "HLL→BIGINT"},
		{"bitmap_result", "BITMAP→BIGINT"},
		{"count_col", "BIGINT SUM"},
	}

	for _, col := range aggColumns {
		val := row4[col.name]
		goType := fmt.Sprintf("%T", val)
		valStr := formatValue(val)
		t.Logf("%-20s | %-20s | %-15s | %s", col.name, col.srType, goType, valStr)
	}

	// 5. NULL values
	t.Log("")
	t.Log("=== 5. NULL Values (test_db.all_types, id=3) ===")
	t.Log("--------------------------------------------------------------------------------")
	t.Logf("%-20s | %-20s | %-15s | %s", "Column", "StarRocks Type", "Go Type", "Value")
	t.Log("--------------------------------------------------------------------------------")

	res5, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.all_types WHERE id = 3",
	})
	require.NoError(t, err)
	require.True(t, res5.Next())
	row5 := make(map[string]any)
	err = res5.MapScan(row5)
	require.NoError(t, err)
	res5.Close()

	nullColumns := []struct {
		name   string
		srType string
	}{
		{"bool_col", "BOOLEAN"},
		{"int_col", "INT"},
		{"varchar_col", "VARCHAR"},
		{"date_col", "DATE"},
		{"json_col", "JSON"},
	}

	for _, col := range nullColumns {
		val := row5[col.name]
		goType := fmt.Sprintf("%T", val)
		valStr := formatValue(val)
		t.Logf("%-20s | %-20s | %-15s | %s", col.name, col.srType, goType, valStr)
	}

	// 6. Boundary values
	t.Log("")
	t.Log("=== 6. Boundary Values (test_db.boundary_values) ===")
	t.Log("--------------------------------------------------------------------------------")
	t.Logf("%-20s | %-20s | %-15s | %s", "Column", "StarRocks Type", "Go Type", "Value")
	t.Log("--------------------------------------------------------------------------------")

	res6, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.boundary_values WHERE id = 1",
	})
	require.NoError(t, err)
	require.True(t, res6.Next())
	row6 := make(map[string]any)
	err = res6.MapScan(row6)
	require.NoError(t, err)
	res6.Close()

	boundaryColumns := []struct {
		name   string
		srType string
	}{
		{"tinyint_min", "TINYINT"},
		{"tinyint_max", "TINYINT"},
		{"smallint_min", "SMALLINT"},
		{"smallint_max", "SMALLINT"},
		{"int_min", "INT"},
		{"int_max", "INT"},
		{"bigint_min", "BIGINT"},
		{"bigint_max", "BIGINT"},
		{"empty_string", "VARCHAR"},
		{"whitespace_string", "VARCHAR"},
	}

	for _, col := range boundaryColumns {
		val := row6[col.name]
		goType := fmt.Sprintf("%T", val)
		valStr := formatValue(val)
		t.Logf("%-20s | %-20s | %-15s | %s", col.name, col.srType, goType, valStr)
	}

	// 7. Unicode strings
	t.Log("")
	t.Log("=== 7. Unicode/Encoding (test_db.string_encoding_test) ===")
	t.Log("--------------------------------------------------------------------------------")
	t.Logf("%-20s | %-20s | %-15s | %s", "Column", "StarRocks Type", "Go Type", "Value")
	t.Log("--------------------------------------------------------------------------------")

	res7, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.string_encoding_test WHERE id = 1",
	})
	require.NoError(t, err)
	require.True(t, res7.Next())
	row7 := make(map[string]any)
	err = res7.MapScan(row7)
	require.NoError(t, err)
	res7.Close()

	unicodeColumns := []struct {
		name   string
		srType string
	}{
		{"ascii_col", "VARCHAR (ASCII)"},
		{"unicode_col", "VARCHAR (Unicode)"},
		{"emoji_col", "VARCHAR (Emoji)"},
		{"korean_col", "VARCHAR (Korean)"},
		{"chinese_col", "VARCHAR (Chinese)"},
		{"japanese_col", "VARCHAR (Japanese)"},
	}

	for _, col := range unicodeColumns {
		val := row7[col.name]
		goType := fmt.Sprintf("%T", val)
		valStr := formatValue(val)
		t.Logf("%-20s | %-20s | %-15s | %s", col.name, col.srType, goType, valStr)
	}

	// 8. Special characters
	t.Log("")
	t.Log("=== 8. Special Characters (test_db.special_chars) ===")
	t.Log("--------------------------------------------------------------------------------")
	t.Logf("%-20s | %-20s | %-15s | %s", "Column", "StarRocks Type", "Go Type", "Value")
	t.Log("--------------------------------------------------------------------------------")

	res8, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.special_chars WHERE id = 1",
	})
	require.NoError(t, err)
	require.True(t, res8.Next())
	row8 := make(map[string]any)
	err = res8.MapScan(row8)
	require.NoError(t, err)
	res8.Close()

	specialColumns := []struct {
		name   string
		srType string
	}{
		{"newline_col", "VARCHAR"},
		{"tab_col", "VARCHAR"},
		{"quote_col", "VARCHAR"},
		{"emoji_col", "VARCHAR"},
		{"sql_injection_col", "VARCHAR"},
	}

	for _, col := range specialColumns {
		val := row8[col.name]
		goType := fmt.Sprintf("%T", val)
		valStr := formatValue(val)
		t.Logf("%-20s | %-20s | %-15s | %s", col.name, col.srType, goType, valStr)
	}

	// Summary table
	t.Log("")
	t.Log("================================================================================")
	t.Log("                           Type Mapping Summary")
	t.Log("================================================================================")
	t.Logf("%-20s | %-20s | %-15s", "StarRocks Type", "Schema Code", "Go Return Type")
	t.Log("--------------------------------------------------------------------------------")

	summaryTypes := []struct {
		srType     string
		schemaCode string
		goType     string
	}{
		{"BOOLEAN", "CODE_INT8", "int16"},
		{"TINYINT", "CODE_INT8", "int16"},
		{"SMALLINT", "CODE_INT16", "int16"},
		{"INT", "CODE_INT32", "int32"},
		{"BIGINT", "CODE_INT64", "int64"},
		{"LARGEINT", "CODE_INT128", "string (>64bit auto)"},
		{"FLOAT", "CODE_FLOAT32", "float64"},
		{"DOUBLE", "CODE_FLOAT64", "float64"},
		{"DECIMAL", "CODE_STRING", "string (precision)"},
		{"CHAR/VARCHAR/STRING", "CODE_STRING", "string"},
		{"DATE", "CODE_DATE", "string"},
		{"DATETIME", "CODE_TIMESTAMP", "string"},
		{"JSON", "CODE_STRING", "string"},
		{"ARRAY<T>", "CODE_ARRAY", "string (JSON)"},
		{"MAP<K,V>", "CODE_MAP", "string (JSON)"},
		{"STRUCT<...>", "CODE_STRUCT", "string (JSON)"},
		{"VARBINARY", "CODE_STRING", "string"},
		{"HLL", "N/A", "use hll_cardinality()"},
		{"BITMAP", "N/A", "use bitmap_count()"},
		{"NULL", "N/A", "<nil>"},
	}

	for _, s := range summaryTypes {
		t.Logf("%-20s | %-20s | %-15s", s.srType, s.schemaCode, s.goType)
	}

	t.Log("================================================================================")
}

func testLoadDDL(t *testing.T, conn drivers.Handle) {
	olap, _ := conn.AsOLAP("default")
	ctx := context.Background()

	// Test DDL for the all_types table
	table, err := olap.InformationSchema().Lookup(ctx, "", "test_db", "all_types")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(ctx, table)
	require.NoError(t, err)
	require.Contains(t, table.DDL, "CREATE TABLE")
	require.Contains(t, table.DDL, "all_types")
}

// formatValue formats a value for display, truncating long strings
func formatValue(val any) string {
	if val == nil {
		return "<nil>"
	}
	s := fmt.Sprintf("%v", val)
	if len(s) > 50 {
		return s[:47] + "..."
	}
	return s
}

// ============================================================
// High-precision DECIMAL test (DECIMAL32, DECIMAL64, DECIMAL128)
// ============================================================
func testDecimalPrecision(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	t.Log("=== Testing High-Precision DECIMAL Types ===")
	t.Log("StarRocks internal types based on precision:")
	t.Log("  DECIMAL(1-9, S)   → DECIMAL32")
	t.Log("  DECIMAL(10-18, S) → DECIMAL64")
	t.Log("  DECIMAL(19-38, S) → DECIMAL128")
	t.Log("")

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.decimal_precision_test ORDER BY id",
	})
	require.NoError(t, err)
	defer res.Close()

	// Log the DatabaseTypeName for each column
	if starrocksRes, ok := res.Rows.(*starrocksRows); ok {
		t.Log("=== Raw DatabaseTypeName from MySQL Driver ===")
		for _, ct := range starrocksRes.colTypes {
			t.Logf("Column: %-15s DatabaseTypeName: %s", ct.Name(), ct.DatabaseTypeName())
		}
	}

	t.Log("")
	t.Log("=== Values (returned as string to preserve precision) ===")
	t.Logf("%-5s | %-20s | %-25s | %s", "ID", "DECIMAL32 (9,4)", "DECIMAL64 (18,6)", "DECIMAL128 (38,10)")
	t.Log("------+----------------------+---------------------------+------------------------------------------")

	rowNum := 0
	for res.Next() {
		row := make(map[string]any)
		err := res.MapScan(row)
		require.NoError(t, err)
		rowNum++

		id := row["id"]
		d32 := row["decimal32_col"]
		d64 := row["decimal64_col"]
		d128 := row["decimal128_col"]

		// All DECIMAL types should return as string
		d32Str, ok := d32.(string)
		require.True(t, ok, "DECIMAL32 should be string, got %T", d32)

		d64Str, ok := d64.(string)
		require.True(t, ok, "DECIMAL64 should be string, got %T", d64)

		d128Str, ok := d128.(string)
		require.True(t, ok, "DECIMAL128 should be string, got %T", d128)

		t.Logf("%-5v | %-20s | %-25s | %s", id, d32Str, d64Str, d128Str)

		// Verify high-precision DECIMAL128 preserves all digits
		if rowNum == 1 {
			// 12345678901234567890123456.7890123456 - 26 digits before decimal, 10 after
			require.Contains(t, d128Str, "12345678901234567890123456")
			require.Contains(t, d128Str, "7890123456")
		}
	}

	t.Log("")
	t.Log("=== Precision Preservation Test ===")
	t.Log("DECIMAL128 value: 12345678901234567890123456.7890123456")
	t.Log("If this were float64, precision would be lost:")
	t.Log("  float64 max precision: ~15-17 significant digits")
	t.Log("  DECIMAL128 has 36 significant digits → string preserves all")
}
