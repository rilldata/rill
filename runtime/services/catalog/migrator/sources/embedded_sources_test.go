package sources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEmbeddedSourcePositiveTest(t *testing.T) {
	positiveTests := []struct {
		path      string
		connector string
		name      string
	}{
		{"http://server/path/to/AdBids.csv.tgz", "https", "a5306e6985ea2b666c2e3668c9474246c"},
		{"gs://server-name/path/to/AdBids.csv.tgz", "gcs", "a421299d258dfe2d33eb7aad12fc09355"},
		{"s3://server-name/path/to/AdBids.csv.tgz", "s3", "ad8c587ab74eb39ea61155a61e3484f65"},
		{"s3://server-name/path/**/*AdBids[0-9].csv.tgz", "s3", "a9d3c3ffaca10ecb34dfd8369e5524f43"},
		{"data/AdBids.csv", "local_file", "a5679a659bbebf0ea9bf47a382e380b7b"},
		{"/path/to/AdBids", "local_file", "af93d462b56dd94a9e1ff648f8c10603c"},
	}

	for _, tt := range positiveTests {
		t.Run(tt.path, func(t *testing.T) {
			s, ok := ParseEmbeddedSource(tt.path)
			require.True(t, ok)
			require.NotNil(t, s)
			require.Equal(t, tt.connector, s.Connector)
			require.Equal(t, tt.name, s.Name)
		})
	}
}

func TestParseEmbeddedSourceNegativeTest(t *testing.T) {
	negativeTests := []string{
		"AdBids",
		"Ad_Bids",
		"Ad-Bids",
		"file://data/AdBids.csv",
	}

	for _, tt := range negativeTests {
		t.Run(tt, func(t *testing.T) {
			s, ok := ParseEmbeddedSource(tt)
			require.Nil(t, s)
			require.False(t, ok)
		})
	}
}
