package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDuckdbSQL(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		defaultToJSON bool
		wantContains  string
	}{
		{"parquet", "s3://bucket/data.parquet", false, "read_parquet"},
		{"csv", "s3://bucket/data.csv", false, "read_csv"},
		{"tsv", "s3://bucket/data.tsv", false, "read_csv"},
		{"json", "s3://bucket/data.json", false, "read_json"},
		{"ndjson", "s3://bucket/data.ndjson", false, "read_json"},
		{"compound parquet.gz", "s3://bucket/data.v1.parquet.gz", false, "read_parquet"},
		{"unknown default generic", "s3://bucket/data.xyz", false, "select * from 's3://bucket/data.xyz'"},
		{"unknown default json", "https://example.com/api", true, "read_json"},
		{"txt", "s3://bucket/data.txt", false, "read_csv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := duckdbSQL(tt.path, tt.defaultToJSON)
			require.Contains(t, query, tt.wantContains)
			require.Contains(t, query, tt.path)
		})
	}
}

func TestMatchesExtFixedFalsePositive(t *testing.T) {
	// This was a bug: "parquet-archive/readme.txt" should NOT match ".parquet"
	require.False(t, matchesExt("s3://bucket/parquet-archive/readme.txt", ".parquet"),
		"path with 'parquet' in directory name should not match .parquet extension")

	// But actual parquet files should match
	require.True(t, matchesExt("s3://bucket/data.parquet", ".parquet"))
	require.True(t, matchesExt("s3://bucket/data.v1.parquet.gz", ".parquet"))
}

func TestMatchesExtCompound(t *testing.T) {
	require.True(t, matchesExt("data.csv.gz", ".csv"))
	require.True(t, matchesExt("data.ndjson.gz", ".ndjson"))
	require.False(t, matchesExt("data.xyz", ".csv", ".json"))
}

func TestClickhouseHeaders(t *testing.T) {
	// With headers: produces headers() syntax
	props := []ProcessedProp{
		{Key: "headers", Value: "\n  Authorization: \"Bearer {{ .env.connector.https.authorization }}\"\n  X-API-Key: \"{{ .env.connector.https.x_api_key }}\"", Quoted: false},
	}
	result := clickhouseHeaders(props)
	require.Contains(t, result, "headers(")
	require.Contains(t, result, "'Authorization'='Bearer {{ .env.connector.https.authorization }}'")
	require.Contains(t, result, "'X-API-Key'='{{ .env.connector.https.x_api_key }}'")

	// No headers prop: returns empty
	noHeaders := []ProcessedProp{
		{Key: "path", Value: "https://example.com", Quoted: true},
	}
	require.Equal(t, "", clickhouseHeaders(noHeaders))

	// Nil/wrong type: returns empty
	require.Equal(t, "", clickhouseHeaders(nil))
	require.Equal(t, "", clickhouseHeaders("not a slice"))
}
