package athena_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestInformationSchemaListTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	conn, _ := acquireTestAthena(t)
	ctx := t.Context()

	infoSchema, ok := conn.AsInformationSchema()
	require.True(t, ok)

	// awsdatacatalog is the default Athena catalog; integration_test is the test schema.
	// current_schema in Athena sessions defaults to "default", so IsDefaultDatabaseSchema is false.
	tables, _, err := infoSchema.ListTables(ctx, "awsdatacatalog", "integration_test", 0, "")
	require.NoError(t, err)
	require.NotEmpty(t, tables)

	for _, tbl := range tables {
		require.True(t, tbl.IsDefaultDatabase, "table %s: expected IsDefaultDatabase=true", tbl.Name)
		require.False(t, tbl.IsDefaultDatabaseSchema, "table %s: expected IsDefaultDatabaseSchema=false (current_schema=default)", tbl.Name)
	}
}

func TestGetTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	conn, olap := acquireTestAthena(t)
	ctx := t.Context()

	t.Run("get table metadata", func(t *testing.T) {
		infoSchema, ok := conn.AsInformationSchema()
		require.True(t, ok)

		// Test getting metadata for the all_datatypes table
		metadata, err := infoSchema.GetTable(ctx, "awsdatacatalog", "integration_test", "all_datatypes")
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.False(t, metadata.View)
		require.NotEmpty(t, metadata.Schema)

		// Verify some expected columns exist
		_, hasID := metadata.Schema["id"]
		require.True(t, hasID, "Expected 'id' column in table schema")

		_, hasInt32 := metadata.Schema["int32_col"]
		require.True(t, hasInt32, "Expected 'int32_col' column in table schema")

		_, hasFloat := metadata.Schema["float_col"]
		require.True(t, hasFloat, "Expected 'float_col' column in table schema")
	})

	t.Run("get view metadata", func(t *testing.T) {
		infoSchema, ok := conn.AsInformationSchema()
		require.True(t, ok)

		// Create a test view
		err := olap.Exec(ctx, &drivers.Statement{
			Query: "CREATE OR REPLACE VIEW integration_test.test_view AS SELECT id, int32_col FROM integration_test.all_datatypes WHERE id = 1",
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = olap.Exec(ctx, &drivers.Statement{
				Query: "DROP VIEW IF EXISTS integration_test.test_view",
			})
		})

		// Get metadata for the view
		metadata, err := infoSchema.GetTable(ctx, "awsdatacatalog", "integration_test", "test_view")
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.True(t, metadata.View, "Expected test_view to be identified as a view")
		require.NotEmpty(t, metadata.Schema)

		// Verify columns from the view
		_, hasID := metadata.Schema["id"]
		require.True(t, hasID, "Expected 'id' column in view schema")

		_, hasInt32 := metadata.Schema["int32_col"]
		require.True(t, hasInt32, "Expected 'int32_col' column in view schema")
	})
}
