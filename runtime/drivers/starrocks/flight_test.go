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

// testFlightScanWithPointers verifies that flightRows.Scan correctly writes values
// through pointers, matching the database/sql convention used by SelectInlineResults.
func testFlightScanWithPointers(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT 42 AS id, 'hello' AS name, 3.14 AS price",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	// Scan with pointer targets (the pattern used by SelectInlineResults)
	var id, name, price any
	err = res.Scan(&id, &name, &price)
	require.NoError(t, err)

	require.NotNil(t, id, "id should not be nil after Scan")
	require.NotNil(t, name, "name should not be nil after Scan")
	require.NotNil(t, price, "price should not be nil after Scan")

	// Verify values are written through pointers
	require.Equal(t, "hello", name)
}

// testFlightScanNullValues verifies that Scan correctly handles NULL values via pointers.
func testFlightScanNullValues(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT NULL AS nullable_col, 'present' AS non_null_col",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	var nullable, nonNull any
	err = res.Scan(&nullable, &nonNull)
	require.NoError(t, err)

	require.Nil(t, nullable, "NULL column should be nil after Scan")
	require.Equal(t, "present", nonNull)
}

// testFlightScanColumnMismatch verifies that Scan returns an error for wrong column count.
func testFlightScanColumnMismatch(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT 1 AS a, 2 AS b, 3 AS c",
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	// Too few targets
	var a, b any
	err = res.Scan(&a, &b)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected 3 columns")
}

// testFlightParameterFallback verifies that Flight SQL connection falls back to MySQL
// when parameterized queries are used.
func testFlightParameterFallback(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()

	res, err := olap.Query(ctx, &drivers.Statement{
		Query: "SELECT * FROM test_db.all_types WHERE id = ?",
		Args:  []any{1},
	})
	require.NoError(t, err)
	defer res.Close()

	require.True(t, res.Next())

	row := make(map[string]any)
	err = res.MapScan(row)
	require.NoError(t, err)
	require.NotNil(t, row["id"])
}

// TestParseFlightLocation tests parsing of Flight endpoint Location URIs.
func TestParseFlightLocation(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr string
	}{
		{
			name: "grpc+tcp with port",
			uri:  "grpc+tcp://10.0.0.1:8419",
			want: "10.0.0.1:8419",
		},
		{
			name: "grpc+tcp without port",
			uri:  "grpc+tcp://10.0.0.1",
			want: "10.0.0.1",
		},
		{
			name: "grpc scheme",
			uri:  "grpc://host.example.com:9419",
			want: "host.example.com:9419",
		},
		{
			name: "IPv6 with port",
			uri:  "grpc+tcp://[::1]:8419",
			want: "::1:8419",
		},
		{
			name: "IPv6 without port",
			uri:  "grpc+tcp://[::1]",
			want: "::1",
		},
		{
			name:    "empty URI",
			uri:     "",
			wantErr: "empty location URI",
		},
		{
			name:    "no host",
			uri:     "grpc+tcp:///path",
			wantErr: "no host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlightLocation(tt.uri)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
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
