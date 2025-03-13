package rilltime

import (
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"github.com/stretchr/testify/require"
)

func Test_Eval(t *testing.T) {
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
		{"m", "2025-03-10T06:31:00Z", "2025-03-10T06:32:00Z", timeutil.TimeGrainSecond},
		{"m~", "2025-03-10T06:32:00Z", "2025-03-10T06:32:36Z", timeutil.TimeGrainSecond},
		{"<m", "2025-03-10T06:00:00Z", "2025-03-10T06:01:00Z", timeutil.TimeGrainSecond},
		{">m", "2025-03-10T06:59:00Z", "2025-03-10T07:00:00Z", timeutil.TimeGrainSecond},

		{"-2d", "2025-03-08T00:00:00Z", "2025-03-09T00:00:00Z", timeutil.TimeGrainHour},
		{"+2d", "2025-03-12T00:00:00Z", "2025-03-13T00:00:00Z", timeutil.TimeGrainHour},
		{"<2d", "2025-03-10T00:00:00Z", "2025-03-12T00:00:00Z", timeutil.TimeGrainHour},
		{">2d", "2025-03-15T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainHour},

		{"m of -2d", "2025-03-08T06:31:00Z", "2025-03-08T06:32:00Z", timeutil.TimeGrainSecond},
		{"m~ of -2d", "2025-03-08T06:32:00Z", "2025-03-08T06:32:36Z", timeutil.TimeGrainSecond},
		{"<m of -2d", "2025-03-08T00:00:00Z", "2025-03-08T00:01:00Z", timeutil.TimeGrainSecond},
		{">m of -2d", "2025-03-08T23:59:00Z", "2025-03-09T00:00:00Z", timeutil.TimeGrainSecond},

		{"m of +2d", "2025-03-12T06:31:00Z", "2025-03-12T06:32:00Z", timeutil.TimeGrainSecond},
		{"m~ of +2d", "2025-03-12T06:32:00Z", "2025-03-12T06:32:36Z", timeutil.TimeGrainSecond},
		{"<m of +2d", "2025-03-12T00:00:00Z", "2025-03-12T00:01:00Z", timeutil.TimeGrainSecond},
		{">m of +2d", "2025-03-12T23:59:00Z", "2025-03-13T00:00:00Z", timeutil.TimeGrainSecond},

		{"W1", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"W1 of -2M", "2024-12-30T00:00:00Z", "2025-01-06T00:00:00Z", timeutil.TimeGrainDay},
		{"D3 of W1 of -3Y", "2022-01-05T00:00:00Z", "2022-01-06T00:00:00Z", timeutil.TimeGrainHour},
		{"W2 of M11 of +3Y", "2028-11-06T00:00:00Z", "2028-11-13T00:00:00Z", timeutil.TimeGrainDay},
		{"<3m of H2 of -6D of -1M", "2025-02-04T01:00:00Z", "2025-02-04T01:03:00Z", timeutil.TimeGrainSecond},
	}

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			rt, err := ParseV2(testCase.timeRange, ParseOptions{})
			require.NoError(t, err)

			start, end, grain := rt.Eval(EvalOptions{
				Now:        now,
				MinTime:    minTime,
				MaxTime:    maxTime,
				Watermark:  watermark,
				FirstDay:   1,
				FirstMonth: 1,
			})
			fmt.Println(testCase.timeRange, start, end)
			require.Equal(t, parseTestTime(t, testCase.start), start)
			require.Equal(t, parseTestTime(t, testCase.end), end)
			require.Equal(t, testCase.grain, grain)
		})
	}
}
