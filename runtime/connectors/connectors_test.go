package connectors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPropertiesEquals(t *testing.T) {
	s1 := &Source{
		Name:       "s1",
		Properties: map[string]any{"a": 100, "b": "hello world"},
	}

	s2 := &Source{
		Name:       "s2",
		Properties: map[string]any{"a": 100, "b": "hello world"},
	}

	s3 := &Source{
		Name:       "s3",
		Properties: map[string]any{"a": 101, "b": "hello world"},
	}

	s4 := &Source{
		Name:       "s4",
		Properties: map[string]any{"a": 100, "c": "hello world"},
	}

	// s1 and s2 should be equal
	require.True(t, s1.PropertiesEquals(s2) && s2.PropertiesEquals(s1))

	// s1 should not equal s3 or s4
	require.False(t, s1.PropertiesEquals(s3) || s3.PropertiesEquals(s1))
	require.False(t, s1.PropertiesEquals(s4) || s4.PropertiesEquals(s1))

	// s2 should not equal s3 or s4
	require.False(t, s2.PropertiesEquals(s3) || s3.PropertiesEquals(s2))
	require.False(t, s2.PropertiesEquals(s4) || s4.PropertiesEquals(s2))
}

func TestGetSourceFromPathPositiveTest(t *testing.T) {
	positiveTests := []struct {
		path      string
		connector string
		name      string
	}{
		{"http://server/path/to/AdBids.csv.tgz", "https", "https_http___server_path_to_AdBids_csv_tgz"},
		{"s3://server-name/path/to/AdBids.csv.tgz", "s3", "s3_s3___server_name_path_to_AdBids_csv_tgz"},
		{"data/AdBids.csv", "local_file", "local_file_data_AdBids_csv"},
		{"/path/to/AdBids", "local_file", "local_file__path_to_AdBids"},
	}

	for _, tt := range positiveTests {
		t.Run(tt.path, func(t *testing.T) {
			s := GetSourceFromPath(tt.path)
			require.NotNil(t, s)
			require.Equal(t, tt.connector, s.Connector)
			require.Equal(t, tt.name, s.Name)
		})
	}
}

func TestGetSourceFromPathNegativeTest(t *testing.T) {
	negativeTests := []string{
		"AdBids",
		"Ad_Bids",
		"Ad-Bids",
		"gs://server-name/path/to/AdBids.csv.tgz",
		"file://data/AdBids.csv",
	}

	for _, tt := range negativeTests {
		t.Run(tt, func(t *testing.T) {
			s := GetSourceFromPath(tt)
			require.Nil(t, s)
		})
	}
}
