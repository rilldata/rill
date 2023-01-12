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
