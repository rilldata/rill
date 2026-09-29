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

// expectedTables are tables bootstrapped by the Pinot batch quickstart.
// The tests only assert that these are present, not that they are the only ones,
// since newer Pinot versions bootstrap additional tables.
var expectedTables = []string{
	"airlineStats",
	"baseballStats",
	"billing",
	"clickstreamFunnel",
	"dimBaseballTeams",
	"fineFoodReviews",
	"githubComplexTypeEvents",
	"githubEvents",
	"starbucksStores",
	"testUnnest",
}

func TestInformationSchema(t *testing.T) {
	testmode.Expensive(t)
	cfg := testruntime.AcquireConnector(t, "pinot")
	conn, err := drivers.Open("pinot", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, olap) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, olap) })
	t.Run("testInformationSchemaAllPagination", func(t *testing.T) { testInformationSchemaAllPagination(t, olap) })
	t.Run("testInformationSchemaAllPaginationWithLike", func(t *testing.T) { testInformationSchemaAllPaginationWithLike(t, olap) })
	t.Run("testInformationSchemaListDatabaseSchemas", func(t *testing.T) { testInformationSchemaListDatabaseSchemas(t, infoSchema) })
	t.Run("testInformationSchemaListTables", func(t *testing.T) { testInformationSchemaListTables(t, infoSchema) })
	t.Run("testInformationSchemaListTablesPagination", func(t *testing.T) { testInformationSchemaListTablesPagination(t, infoSchema) })
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, olap) })
}

func testInformationSchemaAll(t *testing.T, olap drivers.OLAPStore) {
	tables, _, err := olap.InformationSchema().All(context.Background(), "", 0, "")
	require.NoError(t, err)

	names := make([]string, len(tables))
	for i, table := range tables {
		names[i] = table.Name
	}
	require.Subset(t, names, expectedTables)
	require.IsNonDecreasing(t, names)
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

func testInformationSchemaAllPagination(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	const pageSize = 4

	// Test with page size 0, which returns all tables in one page
	tables, nextToken, err := olap.InformationSchema().All(ctx, "", 0, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	all := make([]string, len(tables))
	for i, table := range tables {
		all[i] = table.Name
	}
	require.Subset(t, all, expectedTables)

	// Test that paging through the results yields the same tables in the same order
	var paged []string
	for token := ""; ; {
		page, next, err := olap.InformationSchema().All(ctx, "", pageSize, token)
		require.NoError(t, err)
		for _, table := range page {
			paged = append(paged, table.Name)
		}
		if next == "" {
			require.LessOrEqual(t, len(page), pageSize)
			break
		}
		require.Equal(t, pageSize, len(page))
		token = next
	}
	require.Equal(t, all, paged)

	// Test with page size larger than total results
	tables, nextToken, err = olap.InformationSchema().All(ctx, "", 1000, "")
	require.NoError(t, err)
	require.Equal(t, len(all), len(tables))
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

	names := make([]string, len(tables))
	for i, table := range tables {
		names[i] = table.Name
	}
	require.Subset(t, names, expectedTables)
	require.IsNonDecreasing(t, names)
}

func testInformationSchemaListTablesPagination(t *testing.T, infoSchema drivers.InformationSchema) {
	ctx := context.Background()
	const pageSize = 4

	// Test with page size 0, which returns all tables in one page
	tables, nextToken, err := infoSchema.ListTables(ctx, "", "default", 0, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	all := make([]string, len(tables))
	for i, table := range tables {
		all[i] = table.Name
	}
	require.Subset(t, all, expectedTables)

	// Test that paging through the results yields the same tables in the same order
	var paged []string
	for token := ""; ; {
		page, next, err := infoSchema.ListTables(ctx, "", "default", pageSize, token)
		require.NoError(t, err)
		for _, table := range page {
			paged = append(paged, table.Name)
		}
		if next == "" {
			require.LessOrEqual(t, len(page), pageSize)
			break
		}
		require.Equal(t, pageSize, len(page))
		token = next
	}
	require.Equal(t, all, paged)

	// Test with page size larger than total results
	tables, nextToken, err = infoSchema.ListTables(ctx, "", "default", 1000, "")
	require.NoError(t, err)
	require.Equal(t, len(all), len(tables))
	require.Empty(t, nextToken)
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
