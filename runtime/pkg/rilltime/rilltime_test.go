package rilltime

import (
	"testing"
	"time"

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
		{"s by s", "2025-03-10T06:32:35Z", "2025-03-10T06:32:36Z", timeutil.TimeGrainSecond},
		{"s~ by s", "2025-03-10T06:32:36Z", "2025-03-10T06:32:37Z", timeutil.TimeGrainSecond},

		{"m", "2025-03-10T06:31:00Z", "2025-03-10T06:32:00Z", timeutil.TimeGrainSecond},
		{"m by s", "2025-03-10T06:31:00Z", "2025-03-10T06:32:00Z", timeutil.TimeGrainSecond},
		{"0m", "2025-03-10T06:31:00Z", "2025-03-10T06:32:00Z", timeutil.TimeGrainSecond},
		{"m~", "2025-03-10T06:32:00Z", "2025-03-10T06:33:00Z", timeutil.TimeGrainSecond},
		// We always ceil by 1st term. So this is the same as `m~`
		{"m~ by s", "2025-03-10T06:32:00Z", "2025-03-10T06:33:00Z", timeutil.TimeGrainSecond},
		{"-1m", "2025-03-10T06:31:00Z", "2025-03-10T06:32:00Z", timeutil.TimeGrainSecond},
		{"<m", "2025-03-10T06:00:00Z", "2025-03-10T06:01:00Z", timeutil.TimeGrainSecond},
		{">m", "2025-03-10T06:59:00Z", "2025-03-10T07:00:00Z", timeutil.TimeGrainSecond},
		{"dTm", "2025-03-10T00:00:00Z", "2025-03-10T06:32:00Z", timeutil.TimeGrainMinute},
		{"dTm~", "2025-03-10T00:00:00Z", "2025-03-10T06:33:00Z", timeutil.TimeGrainMinute},
		{"-1dTm", "2025-03-10T00:00:00Z", "2025-03-10T06:32:00Z", timeutil.TimeGrainMinute},
		{"-3dTm", "2025-03-08T00:00:00Z", "2025-03-10T06:32:00Z", timeutil.TimeGrainMinute},

		{"h by s", "2025-03-10T05:00:00Z", "2025-03-10T06:00:00Z", timeutil.TimeGrainSecond},
		// We always ceil by 1st term. So the end time is till the end of the hour
		{"h~ by s", "2025-03-10T06:00:00Z", "2025-03-10T07:00:00Z", timeutil.TimeGrainSecond},

		{"-2d", "2025-03-08T00:00:00Z", "2025-03-09T00:00:00Z", timeutil.TimeGrainDay},
		{"+2d", "2025-03-12T00:00:00Z", "2025-03-13T00:00:00Z", timeutil.TimeGrainDay},
		{"<2d", "2025-03-10T00:00:00Z", "2025-03-12T00:00:00Z", timeutil.TimeGrainDay},
		{">2d", "2025-03-15T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainDay},

		{"m of -2d", "2025-03-08T06:31:00Z", "2025-03-08T06:32:00Z", timeutil.TimeGrainSecond},
		{"m~ of -2d", "2025-03-08T06:32:00Z", "2025-03-08T06:33:00Z", timeutil.TimeGrainSecond},
		{"<m of -2d", "2025-03-08T00:00:00Z", "2025-03-08T00:01:00Z", timeutil.TimeGrainSecond},
		{">m of -2d", "2025-03-08T23:59:00Z", "2025-03-09T00:00:00Z", timeutil.TimeGrainSecond},
		{"-3dTm of -2d", "2025-03-06T00:00:00Z", "2025-03-08T06:32:00Z", timeutil.TimeGrainMinute},
		{"-3dTm~ of -2d", "2025-03-06T00:00:00Z", "2025-03-08T06:33:00Z", timeutil.TimeGrainMinute},
		{"-3dTm of +2d", "2025-03-10T00:00:00Z", "2025-03-12T06:32:00Z", timeutil.TimeGrainMinute},
		{"-3dTm~ of +2d", "2025-03-10T00:00:00Z", "2025-03-12T06:33:00Z", timeutil.TimeGrainMinute},

		{"m of +2d", "2025-03-12T06:31:00Z", "2025-03-12T06:32:00Z", timeutil.TimeGrainSecond},
		{"m~ of +2d", "2025-03-12T06:32:00Z", "2025-03-12T06:33:00Z", timeutil.TimeGrainSecond},
		{"<m of +2d", "2025-03-12T00:00:00Z", "2025-03-12T00:01:00Z", timeutil.TimeGrainSecond},
		{">m of +2d", "2025-03-12T23:59:00Z", "2025-03-13T00:00:00Z", timeutil.TimeGrainSecond},

		// 1st of March is on a friday so we take the next monday to start.
		{"W1", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"W1 by H", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainHour},
		// `of M` means previous month, so this will be of Feb. Since 1st of feb is on a friday we take the next monday as start.
		{"W1 of M", "2025-02-03T00:00:00Z", "2025-02-10T00:00:00Z", timeutil.TimeGrainDay},
		// `of M~` means to use current month unlike `of M`
		{"W1 of M~", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		// `of 0M` means previous month.
		{"W1 of 0M", "2025-02-03T00:00:00Z", "2025-02-10T00:00:00Z", timeutil.TimeGrainDay},
		// 1st of Jan is on a Wednesday, so include the 2 days from Dec 2024.
		{"W1 of -2M", "2024-12-30T00:00:00Z", "2025-01-06T00:00:00Z", timeutil.TimeGrainDay},
		// 1st of May is on a Thursday, so include the 2 days from Dec 2024.
		{"W1 of +2M", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z", timeutil.TimeGrainDay},
		{"D3 of W1 of -3Y", "2022-01-05T00:00:00Z", "2022-01-06T00:00:00Z", timeutil.TimeGrainHour},
		{"W2 of M11 of +3Y", "2028-11-06T00:00:00Z", "2028-11-13T00:00:00Z", timeutil.TimeGrainDay},
		{"<3m of H2 of -6D of -1M", "2025-02-04T01:00:00Z", "2025-02-04T01:03:00Z", timeutil.TimeGrainMinute},

		{"2025-03-09T09:30:15Z", "2025-03-09T09:30:15Z", "2025-03-09T09:30:16Z", timeutil.TimeGrainMillisecond},
		{"2025-03-09T09:30", "2025-03-09T09:30:00Z", "2025-03-09T09:31:00Z", timeutil.TimeGrainSecond},
		{"2025-03-09T09", "2025-03-09T09:00:00Z", "2025-03-09T10:00:00Z", timeutil.TimeGrainMinute},
		{"2025-03-09", "2025-03-09T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainHour},
		{"2025-03", "2025-03-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainWeek},
		{"2025", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth},
		{"D3 of W1 of 2022", "2022-01-05T00:00:00Z", "2022-01-06T00:00:00Z", timeutil.TimeGrainHour},
		{"<3m of H2 of 2025-02-04", "2025-02-04T01:00:00Z", "2025-02-04T01:03:00Z", timeutil.TimeGrainMinute},

		{"W1 to W3", "2025-03-03T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainDay},
		{"W1 to W3 by W", "2025-03-03T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainWeek},
		{"W1 of -2M to D", "2024-12-30T00:00:00Z", "2025-03-09T00:00:00Z", timeutil.TimeGrainDay},
		{"W1 of -2M to D~", "2024-12-30T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"-4D to -2D", "2025-03-06T00:00:00Z", "2025-03-08T00:00:00Z", timeutil.TimeGrainDay},

		{"inf", "2020-01-01T00:32:36Z", "2025-03-11T06:32:36Z", timeutil.TimeGrainUnspecified},
		{"P2DT10H", "2025-03-08T20:00:00Z", "2025-03-10T06:00:00Z", timeutil.TimeGrainHour},
		{"rill-MTD", "2025-03-01T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"rill-PWC", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"rill-PW", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},

		// Meant to mimic comparison with previous period
		{"-2D to D~", "2025-03-08T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay},
		{"-2D of -3D to D~ of -3D", "2025-03-05T00:00:00Z", "2025-03-07T00:00:00Z", timeutil.TimeGrainDay},

		// Edge cases around week
		{"YTW", "2024-12-30T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainWeek},
		{"YTW of 2022", "2022-01-03T00:00:00Z", "2022-03-07T00:00:00Z", timeutil.TimeGrainWeek},
		{"<3W", "2025-03-03T00:00:00Z", "2025-03-24T00:00:00Z", timeutil.TimeGrainWeek},
		{"6h~ of D3 of W1 of -1M", "2025-02-05T01:00:00Z", "2025-02-05T07:00:00Z", timeutil.TimeGrainHour},
	}

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			rt, err := Parse(testCase.timeRange, ParseOptions{})
			require.NoError(t, err)

			start, end, grain := rt.Eval(EvalOptions{
				Now:        now,
				MinTime:    minTime,
				MaxTime:    maxTime,
				Watermark:  watermark,
				FirstDay:   1,
				FirstMonth: 1,
			})
			require.Equal(t, parseTestTime(t, testCase.start), start)
			require.Equal(t, parseTestTime(t, testCase.end), end)
			require.Equal(t, testCase.grain, grain)
		})
	}
}

func Test_Eval_watermark_on_boundary(t *testing.T) {
	now := parseTestTime(t, "2025-03-12T10:32:36Z")
	minTime := parseTestTime(t, "2020-01-01T00:32:36Z")
	maxTime := parseTestTime(t, "2025-03-06T00:00:00Z")
	watermark := parseTestTime(t, "2025-03-06T00:00:00Z")
	testCases := []struct {
		timeRange string
		start     string
		end       string
		grain     timeutil.TimeGrain
	}{
		{"2D", "2025-03-04T00:00:00Z", "2025-03-06T00:00:00Z", timeutil.TimeGrainDay},
		{"2D~", "2025-03-05T00:00:00Z", "2025-03-07T00:00:00Z", timeutil.TimeGrainDay},
		{"2D of -2D", "2025-03-02T00:00:00Z", "2025-03-04T00:00:00Z", timeutil.TimeGrainDay},
		{"D", "2025-03-05T00:00:00Z", "2025-03-06T00:00:00Z", timeutil.TimeGrainHour},
		{"D~", "2025-03-06T00:00:00Z", "2025-03-07T00:00:00Z", timeutil.TimeGrainHour},

		{"MTD", "2025-03-01T00:00:00Z", "2025-03-06T00:00:00Z", timeutil.TimeGrainDay},
		{"MTD by H", "2025-03-01T00:00:00Z", "2025-03-06T00:00:00Z", timeutil.TimeGrainHour},
		{"MTD of -1M", "2025-02-01T00:00:00Z", "2025-02-06T00:00:00Z", timeutil.TimeGrainDay},
		{"MTD of -1Y", "2024-03-01T00:00:00Z", "2024-03-06T00:00:00Z", timeutil.TimeGrainDay},
		{"MTD~", "2025-03-01T00:00:00Z", "2025-03-07T00:00:00Z", timeutil.TimeGrainDay},
		{"MTD~ by H", "2025-03-01T00:00:00Z", "2025-03-07T00:00:00Z", timeutil.TimeGrainHour},
		{"MTD~ of -1M", "2025-02-01T00:00:00Z", "2025-02-07T00:00:00Z", timeutil.TimeGrainDay},
		{"MTD~ of -1Y", "2024-03-01T00:00:00Z", "2024-03-07T00:00:00Z", timeutil.TimeGrainDay},
	}

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			rt, err := Parse(testCase.timeRange, ParseOptions{})
			require.NoError(t, err)

			start, end, grain := rt.Eval(EvalOptions{
				Now:        now,
				MinTime:    minTime,
				MaxTime:    maxTime,
				Watermark:  watermark,
				FirstDay:   1,
				FirstMonth: 1,
			})
			require.Equal(t, parseTestTime(t, testCase.start), start)
			require.Equal(t, parseTestTime(t, testCase.end), end)
			require.Equal(t, testCase.grain, grain)
		})
	}
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
