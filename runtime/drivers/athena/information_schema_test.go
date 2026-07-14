package athena_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

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
		metadata, err := infoSchema.Lookup(ctx, "awsdatacatalog", "integration_test", "all_datatypes")
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.False(t, metadata.View)
		require.NotEmpty(t, metadata.Schema)

		// Verify some expected columns exist
		hasID := metadata.Schema.Fields[0].Name == "id"
		require.True(t, hasID, "Expected 'id' column in table schema")

		hasInt32 := metadata.Schema.Fields[2].Name == "int32_col"
		require.True(t, hasInt32, "Expected 'int32_col' column in table schema")

		hasFloat := metadata.Schema.Fields[4].Name == "float_col"
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
		metadata, err := infoSchema.Lookup(ctx, "awsdatacatalog", "integration_test", "test_view")
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.True(t, metadata.View, "Expected test_view to be identified as a view")
		require.NotEmpty(t, metadata.Schema)

		// Verify columns from the view
		hasID := metadata.Schema.Fields[0].Name == "id"
		require.True(t, hasID, "Expected 'id' column in view schema")

		hasInt32 := metadata.Schema.Fields[1].Name == "int32_col"
		require.True(t, hasInt32, "Expected 'int32_col' column in view schema")
	})
}
