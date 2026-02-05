package duckdb_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestExecute(t *testing.T) {
	tempDir := t.TempDir()
	duckDB, err := drivers.Open("duckdb", "default", map[string]any{}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer duckDB.Close()

	olap, ok := duckDB.AsOLAP("")
	require.True(t, ok)

	// Create test table with all types
	result, err := olap.Query(context.Background(), &drivers.Statement{
		Query: `CREATE TABLE all_types (
			id INTEGER PRIMARY KEY,
			small_int SMALLINT,
			normal_int INTEGER,
			big_int BIGINT,
			huge_int HUGEINT,
			numeric_val NUMERIC(10,2),
			decimal_val DECIMAL(10,2),
			real_val REAL,
			double_val DOUBLE,
			float_val FLOAT,
			bool_val BOOLEAN,
			uuid_val UUID,
			char_val CHAR(10),
			varchar_val VARCHAR,
			text_val TEXT,
			blob_val BLOB,
			date_val DATE,
			time_val TIME,
			timestamp_val TIMESTAMP,
			timestamptz_val TIMESTAMPTZ,
			interval_val INTERVAL,
			json_val JSON,
			array_val INTEGER[],
			struct_val STRUCT(x INTEGER, y TEXT)
		);
		
		INSERT INTO all_types VALUES (
			1, 10, 100, 1000, 10000, 123.45, 987.65, 1.23, 4.56, 7.89, TRUE, 
			'550e8400-e29b-41d4-a716-446655440000', 'A', 'Hello', 'Sample Text', 
			'68656c6c6f', '2025-03-20', '12:34:56', '2025-03-20 12:34:56', '2025-03-20 12:34:56 UTC', 
			INTERVAL '1 day', '{"key": "value"}', ARRAY[1,2,3], ROW(42, 'struct text')
		);
		
		INSERT INTO all_types VALUES (
			2, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 
			NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 
			NULL, NULL, NULL, NULL
		);`,
	})
	require.NoError(t, err)
	require.NoError(t, result.Close())

	fileHandle, err := drivers.Open("file", "default", map[string]any{}, storage.MustNew(tempDir, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	defer fileHandle.Close()

	opts := &drivers.ModelExecutorOptions{
		InputHandle:     duckDB,
		InputConnector:  "duckdb",
		OutputHandle:    fileHandle,
		OutputConnector: "file",
		Env: &drivers.ModelEnv{
			AllowHostAccess: false,
			StageChanges:    true,
		},
		PreliminaryOutputProperties: map[string]any{
			"format": "csv",
		},
	}

	me, err := duckDB.AsModelExecutor("default", opts)
	require.NoError(t, err)

	t.Run("test_csv_export", func(t *testing.T) {
		outPath := filepath.Join(tempDir, "out.csv")
		execOpts := &drivers.ModelExecuteOptions{
			ModelExecutorOptions: opts,
			InputProperties: map[string]any{
				"sql": "SELECT * FROM all_types",
			},
			OutputProperties: map[string]any{
				"path":   outPath,
				"format": "csv",
			},
		}

		result, err := me.Execute(context.Background(), execOpts)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify file exists and has content
		stat, err := os.Stat(outPath)
		require.NoError(t, err)
		require.True(t, stat.Size() > 0)

		// Read back and verify contents
		compareResult, err := olap.Query(context.Background(), &drivers.Statement{
			Query: fmt.Sprintf(`
				WITH 
				actual AS (SELECT * FROM read_csv_auto('%s')),
				expected AS (SELECT * FROM read_csv_auto('testdata/expected_export/expected_output.csv')),
				comparison AS (
					SELECT COUNT(*) as mismatch_count
					FROM (
						SELECT * FROM actual
						EXCEPT
						SELECT * FROM expected
					)
				)
				SELECT CASE 
					WHEN mismatch_count = 0 THEN 'MATCH'
					ELSE 'MISMATCH: ' || mismatch_count || ' rows differ'
				END as result
				FROM comparison;
			`, outPath),
		})
		require.NoError(t, err)
		var comparisonResult string
		require.True(t, compareResult.Next())
		require.NoError(t, compareResult.Scan(&comparisonResult))
		require.Equal(t, "MATCH", comparisonResult, "Exported CSV data does not match expected data")
		require.NoError(t, compareResult.Close())
	})
}
