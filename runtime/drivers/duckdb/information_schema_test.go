package duckdb_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestInformationSchema(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/all_datatypes.sql": "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/foo.sql":           "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/bar.sql":           "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/foz.sql":           "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/baz.sql":           "-- @materialize: true\n SELECT * FROM (VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)) AS t(bar, baz)",
			"models/model.sql":         "SELECT 1,2,3",
		},
	})

	olap, release, err := rt.OLAP(t.Context(), instanceID, "")
	require.NoError(t, err)
	t.Cleanup(func() { release() })

	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, olap) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, olap) })
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, olap) })
	t.Run("testInformationSchemaPagination", func(t *testing.T) { testInformationSchemaAllPagination(t, olap) })
	t.Run("testInformationSchemaPaginationWithLike", func(t *testing.T) { testInformationSchemaAllPaginationWithLike(t, olap) })
}

func TestInformationSchemaMotherduck(t *testing.T) {
	testmode.Expensive(t)

	cfg := testruntime.AcquireConnector(t, "motherduck")
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"motherduck"},
		Variables: map[string]string{
			"motherduck_token": cfg["token"].(string),
		},
		Files: map[string]string{
			"connectors/motherduck.yaml": `
type: connector
driver: motherduck
token: "{{ .env.motherduck_token }}"
path: md:rilldata
schema_name: integration_test
`,
		},
	})

	olap, release, err := rt.OLAP(t.Context(), instanceID, "motherduck")
	require.NoError(t, err)
	t.Cleanup(func() { release() })

	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, olap) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, olap) })
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, olap) })
	t.Run("testInformationSchemaPagination", func(t *testing.T) { testInformationSchemaAllPagination(t, olap) })
	t.Run("testInformationSchemaPaginationWithLike", func(t *testing.T) { testInformationSchemaAllPaginationWithLike(t, olap) })
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

	// add this condition to prevent size check for motherduck connector
	if tables[1].DatabaseSchema != "integration_test" {
		require.Greater(t, tables[1].PhysicalSizeBytes, int64(0))
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
	bar, err := olap.InformationSchema().Lookup(ctx, "", "", "bar")
	require.NoError(t, err)
	require.Equal(t, "bar", bar.Name)
	require.Equal(t, 2, len(bar.Schema.Fields))
	require.Equal(t, "bar", bar.Schema.Fields[0].Name)
	require.Equal(t, runtimev1.Type_CODE_STRING, bar.Schema.Fields[0].Type.Code)
	require.Equal(t, "baz", bar.Schema.Fields[1].Name)
	require.Equal(t, runtimev1.Type_CODE_INT32, bar.Schema.Fields[1].Type.Code)
	require.Equal(t, false, bar.View)

	_, err = olap.InformationSchema().Lookup(ctx, "", "", "nonexistent_table")
	require.Equal(t, drivers.ErrNotFound, err)

	table, err := olap.InformationSchema().Lookup(ctx, "", "", "model")
	require.NoError(t, err)
	require.Equal(t, "model", table.Name)
}

func testInformationSchemaAllPagination(t *testing.T, olap drivers.OLAPStore) {
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

func testInformationSchemaAllPaginationWithLike(t *testing.T, olap drivers.OLAPStore) {
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
