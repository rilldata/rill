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

func TestGetSourceFromPath(t *testing.T) {
	s := GetSourceFromPath("AdBids")
	require.Nil(t, s)

	s = GetSourceFromPath("http://server/path/to/AdBids.csv.tgz")
	require.NotNil(t, s)
	require.Equal(t, "https", s.Connector)
	require.Equal(t, "https_http___server_path_to_AdBids_csv_tgz", s.Name)

	s = GetSourceFromPath("gs://server-name/path/to/AdBids.csv.tgz")
	require.NotNil(t, s)
	require.Equal(t, "gcs", s.Connector)
	require.Equal(t, "gcs_gs___server_name_path_to_AdBids_csv_tgz", s.Name)

	s = GetSourceFromPath("file://data/AdBids.csv")
	require.NotNil(t, s)
	require.Equal(t, "local_file", s.Connector)
	require.Equal(t, "local_file_data_AdBids_csv", s.Name)

	s = GetSourceFromPath("data/AdBids.csv")
	require.NotNil(t, s)
	require.Equal(t, "local_file", s.Connector)
	require.Equal(t, "local_file_data_AdBids_csv", s.Name)

	s = GetSourceFromPath("/path/to/AdBids")
	require.NotNil(t, s)
	require.Equal(t, "local_file", s.Connector)
	require.Equal(t, "local_file__path_to_AdBids", s.Name)
}
