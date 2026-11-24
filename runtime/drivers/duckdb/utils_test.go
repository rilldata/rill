package duckdb

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/stretchr/testify/require"
)

func TestSourceReaderParquetExtensionCase(t *testing.T) {
	db, err := sql.Open("duckdb", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	for _, ext := range []string{".parquet", ".PARQUET", ".Parquet"} {
		t.Run(ext, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "events"+ext)
			_, err := db.ExecContext(t.Context(), "COPY (SELECT 42 AS n) TO "+safeSQLString(path)+" (FORMAT PARQUET)")
			require.NoError(t, err)

			reader, err := sourceReader([]string{path}, filepath.Ext(path), nil)
			require.NoError(t, err)
			var n int
			require.NoError(t, db.QueryRowContext(t.Context(), "SELECT n FROM "+reader).Scan(&n))
			require.Equal(t, 42, n)
		})
	}
}
