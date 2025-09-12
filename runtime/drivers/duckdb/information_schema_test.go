package duckdb

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	goruntime "runtime"

	"github.com/joho/godotenv"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInformationSchema(t *testing.T) {
	conn, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	prepareData(t, olap)
	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, olap) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, olap) })
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, olap) })
	t.Run("testInformationSchemaPagination", func(t *testing.T) { testInformationSchemaPagination(t, olap) })
	t.Run("testInformationSchemaPaginationWithLike", func(t *testing.T) { testInformationSchemaPaginationWithLike(t, olap) })
}

func TestInformationSchemaMotherduck(t *testing.T) {
	if testing.Short() {
		t.Skip("motherduck: skipping test in short mode")
	}

	_, currentFile, _, _ := goruntime.Caller(0)
	envPath := filepath.Join(currentFile, "..", "..", "..", "..", ".env")
	_, err := os.Stat(envPath)
	if err == nil {
		require.NoError(t, godotenv.Load(envPath))
	}
	token := os.Getenv("RILL_RUNTIME_MOTHERDUCK_TEST_TOKEN")
	require.NotEmpty(t, token, "RILL_RUNTIME_MOTHERDUCK_TEST_TOKEN not configured")

	cfg := map[string]any{
		"token":       token,
		"path":        "md:rilldata",
		"schema_name": "integration_test",
	}
	conn, err := Driver{name: "motherduck"}.Open("default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, olap) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, olap) })
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, olap) })
	t.Run("testInformationSchemaPagination", func(t *testing.T) { testInformationSchemaPagination(t, olap) })
	t.Run("testInformationSchemaPaginationWithLike", func(t *testing.T) { testInformationSchemaPaginationWithLike(t, olap) })
}

func prepareData(t *testing.T, olap drivers.OLAPStore) {

	_, err := olap.(*connection).createTableAsSelect(context.Background(), "all_datatypes", "SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)", &createTableOptions{})
	require.NoError(t, err)

	_, err = olap.(*connection).createTableAsSelect(context.Background(), "foo", "SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)", &createTableOptions{})
	require.NoError(t, err)

	_, err = olap.(*connection).createTableAsSelect(context.Background(), "bar", "SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)", &createTableOptions{})
	require.NoError(t, err)

	_, err = olap.(*connection).createTableAsSelect(context.Background(), "foz", "SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)", &createTableOptions{})
	require.NoError(t, err)

	_, err = olap.(*connection).createTableAsSelect(context.Background(), "baz", "SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)", &createTableOptions{})
	require.NoError(t, err)

	_, err = olap.(*connection).createTableAsSelect(context.Background(), "model", "SELECT 1,2,3", &createTableOptions{view: true})
	require.NoError(t, err)

}

func testInformationSchemaAll(t *testing.T, olap drivers.OLAPStore) {

	tables, _, err := olap.InformationSchema().All(context.Background(), "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))

	require.Equal(t, "all_datatypes", tables[0].Name)
	require.Equal(t, "bar", tables[1].Name)
	require.Equal(t, "baz", tables[2].Name)
	require.Equal(t, "foo", tables[3].Name)
	require.Equal(t, "foz", tables[4].Name)
	require.Equal(t, "model", tables[5].Name)

	bar := tables[1]
	require.Equal(t, 2, len(bar.Schema.Fields))
	require.Equal(t, "bar", bar.Name)
	require.Equal(t, "bar", bar.Schema.Fields[0].Name)
	require.Equal(t, runtimev1.Type_CODE_STRING, bar.Schema.Fields[0].Type.Code)
	require.Equal(t, "baz", bar.Schema.Fields[1].Name)
	require.Equal(t, runtimev1.Type_CODE_INT32, bar.Schema.Fields[1].Type.Code)
	require.Equal(t, false, bar.View)
	// add this condition to prevent size check for motherduck connector
	if bar.DatabaseSchema != "integration_test" {
		require.Greater(t, bar.PhysicalSizeBytes, int64(0))
	}

	model := tables[5]
	require.Equal(t, 3, len(model.Schema.Fields))
	require.Equal(t, true, model.View)
	require.Equal(t, int64(0), model.PhysicalSizeBytes)
}

func testInformationSchemaAllLike(t *testing.T, olap drivers.OLAPStore) {
	tables, _, err := olap.InformationSchema().All(context.Background(), "%odel", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "model", tables[0].Name)

	tables, _, err = olap.InformationSchema().All(context.Background(), "%model%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "model", tables[0].Name)

	tables, _, err = olap.InformationSchema().All(context.Background(), "%nonexistent_table%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(tables))
}

func testInformationSchemaLookup(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	table, err := olap.InformationSchema().Lookup(ctx, "", "", "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", table.Name)

	_, err = olap.InformationSchema().Lookup(ctx, "", "", "nonexistent_table")
	require.Equal(t, drivers.ErrNotFound, err)

	table, err = olap.InformationSchema().Lookup(ctx, "", "", "model")
	require.NoError(t, err)
	require.Equal(t, "model", table.Name)
}

func testInformationSchemaPagination(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	pageSize := 2

	// Test first page
	tables1, nextToken1, err := olap.InformationSchema().All(ctx, "", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := olap.InformationSchema().All(ctx, "", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.NotEmpty(t, nextToken2)

	// Test third page
	tables3, nextToken3, err := olap.InformationSchema().All(ctx, "", uint32(pageSize), nextToken2)
	require.NoError(t, err)
	require.Equal(t, 2, len(tables3))
	require.Empty(t, nextToken3)

	// Test with page size 0
	tables, nextToken, err := olap.InformationSchema().All(ctx, "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = olap.InformationSchema().All(ctx, "", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 6, len(tables))
	require.Empty(t, nextToken)
}

func testInformationSchemaPaginationWithLike(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	pageSize := 1
	// Test first page
	tables1, nextToken1, err := olap.InformationSchema().All(ctx, "%ba%", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := olap.InformationSchema().All(ctx, "%ba%", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.Empty(t, nextToken2)

	// Test with page size 0
	tables, nextToken, err := olap.InformationSchema().All(ctx, "%ba%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = olap.InformationSchema().All(ctx, "%ba%", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)
}

func TestDatabaseTypeToPB(t *testing.T) {
	tests := []struct {
		input  string
		output *runtimev1.Type
	}{
		{
			input:  "DECIMAL(10,20)",
			output: &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL, Nullable: true},
		},
		{
			input: `STRUCT(foo HUGEINT, "bar" STRUCT(a INTEGER, b MAP(INTEGER, BOOLEAN)), baz VARCHAR[])`,
			output: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
				{Name: "foo", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT128, Nullable: true}},
				{Name: "bar", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
					{Name: "a", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32, Nullable: true}},
					{Name: "b", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_MAP, Nullable: true, MapType: &runtimev1.MapType{
						KeyType:   &runtimev1.Type{Code: runtimev1.Type_CODE_INT32, Nullable: true},
						ValueType: &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL, Nullable: true},
					}}},
				}}}},
				{Name: "baz", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY, Nullable: true, ArrayElementType: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING, Nullable: true}}},
			}}},
		},
		{
			input: `STRUCT("foo ""("" bar" STRUCT("baz ,, \ \"" "" )" INTEGER))`,
			output: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
				{Name: `foo "(" bar`, Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
					{Name: `baz ,, \ \" " )`, Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32, Nullable: true}},
				}}}},
			}}},
		},
	}

	for _, test := range tests {
		output, err := databaseTypeToPB(test.input, true)
		require.NoError(t, err)
		require.Equal(t, test.output, output)
	}
}
