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

		{"-4Y^ to -1M^", "2021-01-01T00:00:00Z", "2025-02-01T00:00:00Z", timeutil.TimeGrainMonth},

		{"Y^ to now", "2025-01-01T00:00:00Z", "2025-03-12T10:32:36.001Z", timeutil.TimeGrainSecond},
		{"-1Y!", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainSecond},
		{"W1 of Y", "2024-12-30T00:00:00Z", "2025-01-06T00:00:00Z", timeutil.TimeGrainSecond},
		{"W1 of -1M^ to -1M$", "2025-02-03T00:00:00Z", "2025-02-10T00:00:00Z", timeutil.TimeGrainSecond},
		{"-2d^ to d$ as of -1Q", "2024-12-08T00:00:00Z", "2024-12-11T00:00:00Z", timeutil.TimeGrainSecond},

		{"<6H of D25 as of -3M", "2024-12-25T00:00:00Z", "2024-12-25T06:00:00Z", timeutil.TimeGrainSecond},
		{"6h starting D25^ as of -3M", "2024-12-25T00:00:00Z", "2024-12-25T06:00:00Z", timeutil.TimeGrainSecond},

		{"-4d^ to now", "2025-03-06T00:00:00Z", "2025-03-12T10:32:36.001Z", timeutil.TimeGrainSecond},
		{"M/MW^ to M/MW^+3W", "2025-03-03T00:00:00Z", "2025-03-24T00:00:00Z", timeutil.TimeGrainSecond},
		{"3W starting M^", "2025-03-03T00:00:00Z", "2025-03-24T00:00:00Z", timeutil.TimeGrainSecond},
		{"1Y starting Y^", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainSecond},

		{">7h of -1d!", "2025-03-09T17:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainSecond},
		{"H4 as of -1d", "2025-03-09T03:00:00Z", "2025-03-09T04:00:00Z", timeutil.TimeGrainSecond},
		{"H4 of -1d!", "2025-03-09T03:00:00Z", "2025-03-09T04:00:00Z", timeutil.TimeGrainSecond},
		{"3d ending -1Q/d$", "2024-12-08T00:00:00Z", "2024-12-11T00:00:00Z", timeutil.TimeGrainSecond},

		{"y/yw^ to w^", "2024-12-30T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainSecond},
		{"m30 of H12 of D5 of >1W of Q3", "2025-09-26T11:29:00Z", "2025-09-26T11:30:00Z", timeutil.TimeGrainSecond},
		{"m30 of H12 of D5 of >1W of Q3 as of -2Y", "2023-09-29T11:29:00Z", "2023-09-29T11:30:00Z", timeutil.TimeGrainSecond},
		{"m30 of H12 of D5 of >1W of Q3 as of -4Y", "2021-10-01T11:29:00Z", "2021-10-01T11:30:00Z", timeutil.TimeGrainSecond},
		{"m30 of H12 of D5 of >1W of Q3 as of -5Y", "2020-09-25T11:29:00Z", "2020-09-25T11:30:00Z", timeutil.TimeGrainSecond},
		{"-5W4M3Q2Y to -4W3M2Q1Y", "2022-01-06T06:32:36Z", "2023-05-13T06:32:36Z", timeutil.TimeGrainSecond},
		{"-5W-4M-3Q-2Y to -4W-3M-2Q-1Y", "2022-01-06T06:32:36Z", "2023-05-13T06:32:36Z", timeutil.TimeGrainSecond},
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
