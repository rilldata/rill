package rillv1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", ""},
		{"foo", "Foo"},
		{"foo_bar", "Foo Bar"},
		{"foo-bar", "Foo Bar"},
		{"_foo", "_foo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := ToDisplayName(test.name)
			require.Equal(t, test.expected, actual)
		})
	}
}
