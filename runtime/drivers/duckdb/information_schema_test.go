package duckdb

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestInformationSchemaAll(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()

	rows, err := olap.Execute(context.Background(), &drivers.Statement{
		Query: "CREATE TEMPORARY VIEW model as (select 1, 2, 3)",
	})
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	tables, err := olap.InformationSchema().All(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, len(tables))

	require.Equal(t, "bar", tables[0].Name)
	require.Equal(t, "foo", tables[1].Name)
	require.Equal(t, "model", tables[2].Name)

	require.Equal(t, 2, len(tables[1].Schema.Fields))
	require.Equal(t, "bar", tables[1].Schema.Fields[0].Name)
	require.Equal(t, api.Type_CODE_STRING, tables[1].Schema.Fields[0].Type.Code)
	require.Equal(t, "baz", tables[1].Schema.Fields[1].Name)
	require.Equal(t, api.Type_CODE_INT32, tables[1].Schema.Fields[1].Type.Code)
}

func TestInformationSchemaLookup(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()
	ctx := context.Background()

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query: "CREATE TEMPORARY VIEW model as (select 1, 2, 3)",
	})
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	table, err := olap.InformationSchema().Lookup(ctx, "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", table.Name)

	_, err = olap.InformationSchema().Lookup(ctx, "bad")
	require.Equal(t, drivers.ErrNotFound, err)

	table, err = olap.InformationSchema().Lookup(ctx, "model")
	require.NoError(t, err)
	require.Equal(t, "model", table.Name)
}

func TestDatabaseTypeToPB(t *testing.T) {
	tests := []struct {
		input  string
		output *api.Type
	}{
		{
			input:  "DECIMAL(10,20)",
			output: &api.Type{Code: api.Type_CODE_DECIMAL, Nullable: true},
		},
		{
			input: "STRUCT(foo HUGEINT, bar STRUCT(a INTEGER, b MAP(INTEGER, BOOLEAN)), baz VARCHAR[])",
			output: &api.Type{Code: api.Type_CODE_STRUCT, Nullable: true, StructType: &api.StructType{Fields: []*api.StructType_Field{
				{Name: "foo", Type: &api.Type{Code: api.Type_CODE_INT128, Nullable: true}},
				{Name: "bar", Type: &api.Type{Code: api.Type_CODE_STRUCT, Nullable: true, StructType: &api.StructType{Fields: []*api.StructType_Field{
					{Name: "a", Type: &api.Type{Code: api.Type_CODE_INT32, Nullable: true}},
					{Name: "b", Type: &api.Type{Code: api.Type_CODE_MAP, Nullable: true, MapType: &api.MapType{
						KeyType:   &api.Type{Code: api.Type_CODE_INT32, Nullable: true},
						ValueType: &api.Type{Code: api.Type_CODE_BOOL, Nullable: true},
					}}},
				}}}},
				{Name: "baz", Type: &api.Type{Code: api.Type_CODE_ARRAY, Nullable: true, ArrayElementType: &api.Type{Code: api.Type_CODE_STRING, Nullable: true}}},
			}}},
		},
	}

	for _, test := range tests {
		output, err := databaseTypeToPB(test.input, true)
		require.NoError(t, err)
		require.Equal(t, test.output, output)
	}
}
