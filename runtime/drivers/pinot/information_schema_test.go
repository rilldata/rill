package pinot

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInformationSchema(t *testing.T) {
	testmode.Expensive(t)
	cfg := testruntime.AcquireConnector(t, "pinot")
	conn, err := drivers.Open("pinot", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, olap) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, olap) })
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, olap) })
	t.Run("testInformationSchemaAllPagination", func(t *testing.T) { testInformationSchemaAllPagination(t, olap) })
	t.Run("testInformationSchemaAllPaginationWithLike", func(t *testing.T) { testInformationSchemaAllPaginationWithLike(t, olap) })
	t.Run("testInformationSchemaListDatabaseSchemas", func(t *testing.T) { testInformationSchemaListDatabaseSchemas(t, infoSchema) })
	t.Run("testInformationSchemaListTables", func(t *testing.T) { testInformationSchemaListTables(t, infoSchema) })
	t.Run("testInformationSchemaGetTable", func(t *testing.T) { testInformationSchemaGetTable(t, infoSchema) })
	t.Run("testInformationSchemaListTablesPagination", func(t *testing.T) { testInformationSchemaListTablesPagination(t, infoSchema) })

}

func testInformationSchemaAll(t *testing.T, olap drivers.OLAPStore) {
	tables, _, err := olap.InformationSchema().All(context.Background(), "", 0, "")
	require.NoError(t, err)
	require.Equal(t, 10, len(tables))

	require.Equal(t, "airlineStats", tables[0].Name)
	require.Equal(t, "baseballStats", tables[1].Name)
	require.Equal(t, "billing", tables[2].Name)
	require.Equal(t, "clickstreamFunnel", tables[3].Name)
	require.Equal(t, "dimBaseballTeams", tables[4].Name)
	require.Equal(t, "fineFoodReviews", tables[5].Name)
	require.Equal(t, "githubComplexTypeEvents", tables[6].Name)
	require.Equal(t, "githubEvents", tables[7].Name)
	require.Equal(t, "starbucksStores", tables[8].Name)
	require.Equal(t, "testUnnest", tables[9].Name)
}

func testInformationSchemaAllLike(t *testing.T, olap drivers.OLAPStore) {
	tables, _, err := olap.InformationSchema().All(context.Background(), "%tarbucks%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "starbucksStores", tables[0].Name)

	tables, _, err = olap.InformationSchema().All(context.Background(), "%starbucksStores%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Equal(t, "starbucksStores", tables[0].Name)

	tables, _, err = olap.InformationSchema().All(context.Background(), "%nonexistent_table%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 0, len(tables))
}

func testInformationSchemaLookup(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	starbucksStores, err := olap.InformationSchema().Lookup(ctx, "", "", "starbucksStores")
	require.NoError(t, err)
	require.Equal(t, "starbucksStores", starbucksStores.Name)

	require.Equal(t, 5, len(starbucksStores.Schema.Fields))
	require.Equal(t, "starbucksStores", starbucksStores.Name)
	require.Equal(t, "lon", starbucksStores.Schema.Fields[0].Name)
	require.Equal(t, runtimev1.Type_CODE_FLOAT32, starbucksStores.Schema.Fields[0].Type.Code)
	require.Equal(t, "lat", starbucksStores.Schema.Fields[1].Name)
	require.Equal(t, runtimev1.Type_CODE_FLOAT32, starbucksStores.Schema.Fields[1].Type.Code)
	require.Equal(t, "name", starbucksStores.Schema.Fields[2].Name)
	require.Equal(t, runtimev1.Type_CODE_STRING, starbucksStores.Schema.Fields[2].Type.Code)
	require.Equal(t, "address", starbucksStores.Schema.Fields[3].Name)
	require.Equal(t, runtimev1.Type_CODE_STRING, starbucksStores.Schema.Fields[3].Type.Code)
	require.Equal(t, "location_st_point", starbucksStores.Schema.Fields[4].Name)
	require.Equal(t, runtimev1.Type_CODE_BYTES, starbucksStores.Schema.Fields[4].Type.Code)
	require.Equal(t, false, starbucksStores.View)

	_, err = olap.InformationSchema().Lookup(ctx, "", "", "nonexistent_table")
	require.ErrorContains(t, err, "unexpected status code: 404")
}

func testInformationSchemaAllPagination(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	pageSize := 4

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
	require.Equal(t, 10, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = olap.InformationSchema().All(ctx, "", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 10, len(tables))
	require.Empty(t, nextToken)
}

func testInformationSchemaAllPaginationWithLike(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	pageSize := 1
	// Test first page
	tables1, nextToken1, err := olap.InformationSchema().All(ctx, "b%", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := olap.InformationSchema().All(ctx, "b%", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.Empty(t, nextToken2)

	// Test with page size 0
	tables, nextToken, err := olap.InformationSchema().All(ctx, "b%", 0, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = olap.InformationSchema().All(ctx, "b%", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))
	require.Empty(t, nextToken)
}

func testInformationSchemaListDatabaseSchemas(t *testing.T, infoSchema drivers.InformationSchema) {
	databaseSchemas, _, err := infoSchema.ListDatabaseSchemas(context.Background(), 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(databaseSchemas))

	require.Equal(t, "", databaseSchemas[0].Database)
	require.Equal(t, "default", databaseSchemas[0].DatabaseSchema)
}

func testInformationSchemaListTables(t *testing.T, infoSchema drivers.InformationSchema) {
	tables, _, err := infoSchema.ListTables(context.Background(), "", "default", 0, "")
	require.NoError(t, err)
	require.Equal(t, 10, len(tables))

	require.Equal(t, "airlineStats", tables[0].Name)
	require.Equal(t, "baseballStats", tables[1].Name)
	require.Equal(t, "billing", tables[2].Name)
	require.Equal(t, "clickstreamFunnel", tables[3].Name)
	require.Equal(t, "dimBaseballTeams", tables[4].Name)
	require.Equal(t, "fineFoodReviews", tables[5].Name)
	require.Equal(t, "githubComplexTypeEvents", tables[6].Name)
	require.Equal(t, "githubEvents", tables[7].Name)
	require.Equal(t, "starbucksStores", tables[8].Name)
	require.Equal(t, "testUnnest", tables[9].Name)
}

func testInformationSchemaGetTable(t *testing.T, infoSchema drivers.InformationSchema) {
	ctx := context.Background()
	starbucksStores, err := infoSchema.GetTable(ctx, "", "default", "starbucksStores")
	require.NoError(t, err)

	require.Equal(t, 5, len(starbucksStores.Schema))
	require.Equal(t, "FLOAT32", starbucksStores.Schema["lon"])
	require.Equal(t, "FLOAT32", starbucksStores.Schema["lat"])
	require.Equal(t, "STRING", starbucksStores.Schema["name"])
	require.Equal(t, "STRING", starbucksStores.Schema["address"])
	require.Equal(t, "BYTES", starbucksStores.Schema["location_st_point"])
	require.Equal(t, false, starbucksStores.View)

	_, err = infoSchema.GetTable(ctx, "", "default", "nonexistent_table")
	require.ErrorContains(t, err, "unexpected status code: 404")
}

func testInformationSchemaListTablesPagination(t *testing.T, infoSchema drivers.InformationSchema) {
	ctx := context.Background()
	pageSize := 4

	// Test first page
	tables1, nextToken1, err := infoSchema.ListTables(ctx, "", "default", uint32(pageSize), "")
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables1))
	require.NotEmpty(t, nextToken1)

	// Test second page
	tables2, nextToken2, err := infoSchema.ListTables(ctx, "", "default", uint32(pageSize), nextToken1)
	require.NoError(t, err)
	require.Equal(t, pageSize, len(tables2))
	require.NotEmpty(t, nextToken2)

	// Test third page
	tables3, nextToken3, err := infoSchema.ListTables(ctx, "", "default", uint32(pageSize), nextToken2)
	require.NoError(t, err)
	require.Equal(t, 2, len(tables3))
	require.Empty(t, nextToken3)

	// Test with page size 0
	tables, nextToken, err := infoSchema.ListTables(ctx, "", "default", 0, "")
	require.NoError(t, err)
	require.Equal(t, 10, len(tables))
	require.Empty(t, nextToken)

	// Test with page size larger than total results
	tables, nextToken, err = infoSchema.ListTables(ctx, "", "default", 1000, "")
	require.NoError(t, err)
	require.Equal(t, 10, len(tables))
	require.Empty(t, nextToken)
}
