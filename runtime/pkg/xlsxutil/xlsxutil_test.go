package xlsxutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUniqueNames(t *testing.T) {
	tests := []struct {
		name  string
		names []string
		want  []string
	}{
		{"unique", []string{"Country", "Revenue"}, []string{"Country", "Revenue"}},
		{"case", []string{"Country", "country"}, []string{"Country", "country (2)"}},
		// The repeats skip the suffixes of names that are already in use, which keep their names.
		{"repeats and a real suffix", []string{"X", "x", "X", "X (2)"}, []string{"X", "x (3)", "X (4)", "X (2)"}},
		{"empty", []string{"", "Column 1"}, []string{"Column 1", "Column 1 (2)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, UniqueNames(tt.names))
		})
	}
}
