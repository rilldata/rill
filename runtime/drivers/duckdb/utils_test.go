package duckdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDuckDBPropertiesBackwardsCompatibility(t *testing.T) {
	props := map[string]any{
		"format":            "csv",
		"csv.delimiter":     "|",
		"hive_partitioning": true,
	}

	cfg, err := parseFileSourceProperties(props)
	require.NoError(t, err)

	require.Equal(t, cfg, &fileSourceProperties{
		Format: "csv",
		DuckDB: map[string]any{
			"delim":             "'|'",
			"hive_partitioning": true,
		},
	})
}
