package rilltime

import (
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"github.com/stretchr/testify/require"
)

func Test_EvalFinal(t *testing.T) {
	now := parseTestTime(t, "2025-03-12T10:32:36Z")
	minTime := parseTestTime(t, "2020-01-01T00:32:36Z")
	maxTime := parseTestTime(t, "2025-03-11T06:32:36Z")
	watermark := parseTestTime(t, "2025-03-10T06:32:36Z")
	testCases := []struct {
		timeRange string
		start     string
		end       string
		grain     timeutil.TimeGrain
	}{
		{"-4d^ to -2d^", "2025-03-06T00:00:00Z", "2025-03-08T00:00:00Z", timeutil.TimeGrainDay},
		{"-4d to -2d", "2025-03-06T06:32:36Z", "2025-03-08T06:32:36Z", timeutil.TimeGrainDay},

		{"-3d^ to d^", "2025-03-07T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"-3d/d^ to -0d/d^", "2025-03-07T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},

		{"-2d^ to +1d^", "2025-03-08T00:00:00Z", "2025-03-11T00:00:00Z", timeutil.TimeGrainDay},
		{"-2d/d^ to -0d/d$", "2025-03-08T00:00:00Z", "2025-03-11T00:00:00Z", timeutil.TimeGrainDay},
		{"-2d/d^ to +1d/d^", "2025-03-08T00:00:00Z", "2025-03-11T00:00:00Z", timeutil.TimeGrainDay},

		{"<6h of -1D!", "2025-03-09T00:00:00Z", "2025-03-09T06:00:00Z", timeutil.TimeGrainHour},
		{"-1d^ to -1d^+6h", "2025-03-09T00:00:00Z", "2025-03-09T06:00:00Z", timeutil.TimeGrainHour},
		{"6h starting -1d^", "2025-03-09T00:00:00Z", "2025-03-09T06:00:00Z", timeutil.TimeGrainHour},
		{"-1d/d^ to -1d/d^+6h", "2025-03-09T00:00:00Z", "2025-03-09T06:00:00Z", timeutil.TimeGrainHour},

		{"M^ to d^", "2025-03-01T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"-0M/M^ to -0d/d^", "2025-03-01T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},

		{"-4W^ to -3W^", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
		{"-4W!", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
		{"1W starting -4W^", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
		{"1W ending -3W^", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
		{"-4w/w^ to -3w/w^", "2025-02-10T00:00:00Z", "2025-02-17T00:00:00Z", timeutil.TimeGrainDay},
	}

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			rt, err := ParseFinal(testCase.timeRange, ParseOptions{})
			require.NoError(t, err)

			start, end, grain := rt.Eval(EvalOptions{
				Now:        now,
				MinTime:    minTime,
				MaxTime:    maxTime,
				Watermark:  watermark,
				FirstDay:   1,
				FirstMonth: 1,
			})
			fmt.Println(start, end, grain)
			require.Equal(t, parseTestTime(t, testCase.start), start)
			require.Equal(t, parseTestTime(t, testCase.end), end)
			//require.Equal(t, testCase.grain, grain)
		})
	}
}
