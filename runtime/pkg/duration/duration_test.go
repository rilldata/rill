package duration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseISO8601(t *testing.T) {
	tests := []struct {
		from     string
		expected Duration
		err      bool
	}{
		{from: "P2W", expected: Duration{Week: 2}},
		{from: "P1Y2WT5M", expected: Duration{Year: 1, Week: 2, Minute: 5}},
		{from: "P1X", err: true},
		{from: "inf", expected: Duration{Inf: true}},
		{from: "Inf", expected: Duration{Inf: true}},
		{from: "infinity", err: true},
	}
	for _, tt := range tests {
		got, err := ParseISO8601(tt.from)
		if tt.err {
			require.Error(t, err)
		} else {
			require.Equal(t, tt.expected, got)
		}
	}
}
