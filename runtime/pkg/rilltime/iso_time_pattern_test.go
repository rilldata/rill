package rilltime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_isoTimePattern(t *testing.T) {
	positiveTestCases := []struct {
		time     string
		expected AbsoluteTime
	}{
		{
			"2025-03-09T09:30:15Z",
			AbsoluteTime{year: 2025, month: 3, day: 9, hour: 9, minute: 30, second: 15},
		},
		{
			"2025-03-09T09:30",
			AbsoluteTime{year: 2025, month: 3, day: 9, hour: 9, minute: 30},
		},
		{
			"2025-03-09T09",
			AbsoluteTime{year: 2025, month: 3, day: 9, hour: 9},
		},
		{
			"2025-03-09",
			AbsoluteTime{year: 2025, month: 3, day: 9},
		},
		{
			"2025-03",
			AbsoluteTime{year: 2025, month: 3},
		},
		{
			"2025",
			AbsoluteTime{year: 2025},
		},
	}

	for _, testCase := range positiveTestCases {
		t.Run(testCase.time, func(t *testing.T) {
			abs := AbsoluteTime{ISO: testCase.time}
			require.NoError(t, abs.parse())

			abs.ISO = ""
			require.Equal(t, testCase.expected, abs)
		})
	}
}
