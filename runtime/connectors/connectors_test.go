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

	s5 := &Source{
		Name: "s5",
		Properties: map[string]any{
			"number": 0,
			"string": "hello world",
			"nestedMap": map[string]any{
				"nestedMap": map[string]any{
					"string": "value",
					"number": 2,
				},
				"string": "value",
				"number": 1,
			},
		},
	}

	s6 := &Source{
		Name: "s6",
		Properties: map[string]any{
			"number": 0,
			"string": "hello world",
			"nestedMap": map[string]any{
				"number": 1,
				"string": "value",
				"nestedMap": map[string]any{
					"number": 2,
					"string": "value",
				},
			},
		},
	}

	s7 := &Source{
		Name: "s7",
		Properties: map[string]any{
			"number": 0,
			"string": "hello world",
			"nestedMap": map[string]any{
				"number": 1,
				"string": "value",
			},
		},
	}

	// s1 and s2 should be equal
	require.True(t, s1.PropertiesEquals(s2) && s2.PropertiesEquals(s1))

	// s1 should not equal s3 or s4
	require.False(t, s1.PropertiesEquals(s3) || s3.PropertiesEquals(s1))
	require.False(t, s1.PropertiesEquals(s4) || s4.PropertiesEquals(s1))

	// s2 should not equal s3 or s4
	require.False(t, s2.PropertiesEquals(s3) || s3.PropertiesEquals(s2))
	require.False(t, s2.PropertiesEquals(s4) || s4.PropertiesEquals(s2))

	// s5 and s6 should be equal
	require.True(t, s5.PropertiesEquals(s6) && s6.PropertiesEquals(s5))

	// s6 and s7 should not be equal
	require.False(t, s6.PropertiesEquals(s7) || s7.PropertiesEquals(s6))
}
