package clickhouse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileExtToFormatCase(t *testing.T) {
	for _, ext := range []string{".parquet", ".PARQUET", ".Parquet"} {
		format, err := fileExtToFormat(ext)
		require.NoError(t, err, ext)
		require.Equal(t, "Parquet", format, ext)
	}
}
