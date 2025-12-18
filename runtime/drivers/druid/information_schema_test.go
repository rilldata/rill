package druid_test

import (
	"context"
	"testing"

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

	cfg := testruntime.AcquireConnector(t, "druid")
	conn, err := drivers.Open("druid", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	expectedTables := fetchExpectedTables(t, olap)

	t.Run("testInformationSchemaAll", func(t *testing.T) { testInformationSchemaAll(t, olap, expectedTables) })
	t.Run("testInformationSchemaAllLike", func(t *testing.T) { testInformationSchemaAllLike(t, olap, expectedTables) })
	t.Run("testInformationSchemaAllPagination", func(t *testing.T) { testInformationSchemaAllPagination(t, olap, expectedTables) })
	t.Run("testInformationSchemaAllPaginationWithLike", func(t *testing.T) { testInformationSchemaAllPaginationWithLike(t, olap, expectedTables) })
	t.Run("testInformationSchemaLookup", func(t *testing.T) { testInformationSchemaLookup(t, olap, expectedTables) })
	t.Run("testInformationSchemaListDatabaseSchemas", func(t *testing.T) { testInformationSchemaListDatabaseSchemas(t, infoSchema, expectedTables) })
	t.Run("testInformationSchemaListTables", func(t *testing.T) { testInformationSchemaListTables(t, infoSchema, expectedTables) })
	t.Run("testInformationSchemaGetTable", func(t *testing.T) { testInformationSchemaGetTable(t, infoSchema, expectedTables) })
	t.Run("testInformationSchemaListTablesPagination", func(t *testing.T) { testInformationSchemaListTablesPagination(t, infoSchema, expectedTables) })

}

type expectedTable struct {
	Schema string
	Name   string
}

func fetchExpectedTables(t *testing.T, olap drivers.OLAPStore) []expectedTable {
	qry := "SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'druid' ORDER BY TABLE_SCHEMA, TABLE_NAME"
	rows, err := olap.Query(context.Background(), &drivers.Statement{Query: qry})
	require.NoError(t, err)
	defer rows.Close()

	var expected []expectedTable
	for rows.Next() {
		var schema, name string
		err := rows.Scan(&schema, &name)
		require.NoError(t, err)
		expected = append(expected, expectedTable{
			Schema: schema,
			Name:   name,
		})
	}
	require.NotEmpty(t, expected)

	return expected
}

func testInformationSchemaAll(t *testing.T, olap drivers.OLAPStore, expected []expectedTable) {
	ctx := context.Background()

	tables, _, err := olap.InformationSchema().All(ctx, "", 10000, "")
	require.NoError(t, err)
	require.Equal(t, len(expected), len(tables))

	err = olap.InformationSchema().LoadPhysicalSize(ctx, tables)
	require.NoError(t, err)

	// Check tables against expected, preserving order
	for i, tbl := range tables {
		require.Equal(t, expected[i].Name, tbl.Name)
		require.Equal(t, expected[i].Schema, tbl.DatabaseSchema)

		if !tbl.View {
			require.Greater(t, tbl.PhysicalSizeBytes, int64(0), "table %s should have non-zero physical size", tbl.Name)
		}
	}
}

func testInformationSchemaAllLike(t *testing.T, olap drivers.OLAPStore, expected []expectedTable) {
	// Pick the first table whose name starts with "w" from expected tables
	var filteredExpected []expectedTable
	for _, tbl := range expected {
		if len(tbl.Name) > 0 && tbl.Name[0] == 'w' {
			filteredExpected = append(filteredExpected, tbl)
		}
	}
	likePattern := "w%"
	tables, _, err := olap.InformationSchema().All(context.Background(), likePattern, 0, "")
	require.NoError(t, err)

	for i, tbl := range tables {
		require.Equal(t, filteredExpected[i].Name, tbl.Name)
		require.Equal(t, filteredExpected[i].Schema, tbl.DatabaseSchema)
	}
}

func testInformationSchemaAllPagination(t *testing.T, olap drivers.OLAPStore, expected []expectedTable) {
	ctx := context.Background()
	pageSize := 2
	var resultTables []string
	var nextToken string

	for {
		tables, token, err := olap.InformationSchema().All(ctx, "", uint32(pageSize), nextToken)
		require.NoError(t, err)

		// Collect tables in order
		for _, tbl := range tables {
			resultTables = append(resultTables, tbl.Name)
		}

		if token == "" {
			break
		}
		nextToken = token
	}

	// Verify we got all expected tables in the correct order
	require.Equal(t, len(expected), len(resultTables))
	for i, tbl := range expected {
		require.Equal(t, tbl.Name, resultTables[i])
	}
}

func testInformationSchemaAllPaginationWithLike(t *testing.T, olap drivers.OLAPStore, expected []expectedTable) {
	ctx := context.Background()
	pageSize := 2

	var filteredExpected []expectedTable
	for _, tbl := range expected {
		if len(tbl.Name) > 0 && tbl.Name[0] == 'w' {
			filteredExpected = append(filteredExpected, tbl)
		}
	}

	likePattern := "w%"
	var allTables []string
	var nextToken string

	for {
		tables, token, err := olap.InformationSchema().All(ctx, likePattern, uint32(pageSize), nextToken)
		require.NoError(t, err)

		for _, tbl := range tables {
			allTables = append(allTables, tbl.Name)
		}

		if token == "" {
			break
		}
		nextToken = token
	}

	require.Equal(t, len(filteredExpected), len(allTables))
	for i, tbl := range filteredExpected {
		require.Equal(t, tbl.Name, allTables[i])
	}
}

func testInformationSchemaLookup(t *testing.T, olap drivers.OLAPStore, expected []expectedTable) {
	ctx := context.Background()

	require.GreaterOrEqual(t, len(expected), 1, "expected one table for schema lookup test")
	testTable := expected[0].Name
	testSchema := expected[0].Schema

	// Lookup the table
	table, err := olap.InformationSchema().Lookup(ctx, testSchema, "", testTable)
	require.NoError(t, err)
	require.Equal(t, testTable, table.Name)
	require.Equal(t, testSchema, table.DatabaseSchema)

	// Lookup a table that does not exist
	_, err = olap.InformationSchema().Lookup(ctx, "", "", "nonexistent_table")
	require.Equal(t, drivers.ErrNotFound, err)
}

func testInformationSchemaListDatabaseSchemas(t *testing.T, infoSchema drivers.InformationSchema, expected []expectedTable) {
	ctx := context.Background()

	databaseSchemas, _, err := infoSchema.ListDatabaseSchemas(ctx, 10000, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(databaseSchemas))
	require.Equal(t, "", databaseSchemas[0].Database)
	require.Equal(t, "druid", databaseSchemas[0].DatabaseSchema)

	databaseSchemas, _, err = infoSchema.ListDatabaseSchemas(ctx, 0, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(databaseSchemas))
}

func testInformationSchemaListTables(t *testing.T, infoSchema drivers.InformationSchema, expected []expectedTable) {
	ctx := context.Background()

	tables, _, err := infoSchema.ListTables(ctx, "", "druid", 10000, "")
	require.NoError(t, err)
	require.Equal(t, len(expected), len(tables))

	// Check tables against expected, preserving order
	for i, tbl := range tables {
		require.Equal(t, expected[i].Name, tbl.Name)
	}
}

func testInformationSchemaGetTable(t *testing.T, infoSchema drivers.InformationSchema, expected []expectedTable) {
	ctx := context.Background()

	require.GreaterOrEqual(t, len(expected), 1, "expected one table for schema get table test")
	testTable := expected[0].Name

	// Lookup the table
	table, err := infoSchema.GetTable(ctx, "", "druid", testTable)
	require.NoError(t, err)
	require.Greater(t, len(table.Schema), 1)

	table, err = infoSchema.GetTable(ctx, "", "druid", "nonexistent_table")
	require.Equal(t, 0, len(table.Schema))
}

func testInformationSchemaListTablesPagination(t *testing.T, infoSchema drivers.InformationSchema, expected []expectedTable) {
	ctx := context.Background()
	pageSize := 2
	var resultTables []string
	var nextToken string

	for {
		tables, token, err := infoSchema.ListTables(ctx, "", "druid", uint32(pageSize), nextToken)
		require.NoError(t, err)

		// Collect tables in order
		for _, tbl := range tables {
			resultTables = append(resultTables, tbl.Name)
		}

		if token == "" {
			break
		}
		nextToken = token
	}

	// Verify we got all expected tables in the correct order
	require.Equal(t, len(expected), len(resultTables))
	for i, tbl := range expected {
		require.Equal(t, tbl.Name, resultTables[i])
	}
}
