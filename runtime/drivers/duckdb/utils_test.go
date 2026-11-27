package duckdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSourceReaderParquetExtensionCase(t *testing.T) {
	for _, ext := range []string{".parquet", ".PARQUET", ".Parquet"} {
		t.Run(ext, func(t *testing.T) {
			reader, err := sourceReader([]string{"events" + ext}, ext, nil)
			require.NoError(t, err)
			require.Equal(t, "read_parquet(['events"+ext+"'],hive_partitioning=true)", reader)
		})
	}
}
