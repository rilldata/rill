package duration

import (
	"testing"

	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"github.com/stretchr/testify/require"
)

func TestParseISO8601(t *testing.T) {
	tests := []struct {
		from     string
		expected Duration
		err      bool
	}{
		{from: "P2W", expected: StandardDuration{Week: 2}},
		{from: "P1Y2WT5M", expected: StandardDuration{Year: 1, Week: 2, Minute: 5}},
		{from: "P1X", err: true},
		{from: "inf", expected: InfDuration{}},
		{from: "Inf", expected: InfDuration{}},
		{from: "infinity", err: true},
		{from: "rill-TD", expected: TruncToDateDuration{timeutil.TimeGrainDay}},
		{from: "TD", err: true},
		{from: "rill-PM", expected: StandardDuration{Month: 1}},
		{from: "PM", err: true},
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
