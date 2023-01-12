package sources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEmbeddedSourcePositiveTest(t *testing.T) {
	positiveTests := []struct {
		path      string
		connector string
		name      string
	}{
		{"http://server/path/to/AdBids.csv.tgz", "https", "https_http___server_path_to_AdBids_csv_tgz"},
		{"gs://server-name/path/to/AdBids.csv.tgz", "gcs", "gcs_gs___server_name_path_to_AdBids_csv_tgz"},
		{"s3://server-name/path/to/AdBids.csv.tgz", "s3", "s3_s3___server_name_path_to_AdBids_csv_tgz"},
		{"data/AdBids.csv", "local_file", "local_file_data_AdBids_csv"},
		{"/path/to/AdBids", "local_file", "local_file__path_to_AdBids"},
	}

	for _, tt := range positiveTests {
		t.Run(tt.path, func(t *testing.T) {
			s, ok := GetEmbeddedSource(tt.path)
			require.True(t, ok)
			require.NotNil(t, s)
			require.Equal(t, tt.connector, s.Connector)
			require.Equal(t, tt.name, s.Name)
		})
	}
}

func TestGetEmbeddedSourceNegativeTest(t *testing.T) {
	negativeTests := []string{
		"AdBids",
		"Ad_Bids",
		"Ad-Bids",
		"file://data/AdBids.csv",
	}

	for _, tt := range negativeTests {
		t.Run(tt, func(t *testing.T) {
			s, ok := GetEmbeddedSource(tt)
			require.Nil(t, s)
			require.False(t, ok)
		})
	}
}
