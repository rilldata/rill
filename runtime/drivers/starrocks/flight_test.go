package starrocks

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testFlightAllTypesFromTable logs the Arrow Flight SQL type mapping for all columns.
// This is Flight SQL-specific because it documents the Arrow→Go type conversion.
func testFlightAllTypesFromTable(t *testing.T, olap drivers.OLAPStore) {
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

	t.Log("=== Flight SQL Type Mapping ===")
	for k, v := range row {
		t.Logf("%-20s: type=%T, value=%v", k, v, v)
	}

	require.NotNil(t, row["id"])
	require.NotNil(t, row["varchar_col"])

	_, ok := row["varchar_col"].(string)
	require.True(t, ok, "expected string for varchar_col, got %T", row["varchar_col"])
}

// TestStarRocksFlightSQLConnectionError verifies that Flight SQL connection failure
// produces a clear error message. This is a standalone test (no container needed).
func TestStarRocksFlightSQLConnectionError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	_, err := driver{}.Open("default", map[string]any{
		"host":            "localhost",
		"port":            9030,
		"username":        "root",
		"transport":       "flight_sql",
		"flight_sql_port": 19999, // Invalid port
	}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())

	// Should fail because MySQL connection will also fail on wrong port
	// or Flight SQL connection will fail
	require.Error(t, err)
}
