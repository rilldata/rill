package rilltime

import (
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"github.com/stretchr/testify/require"
)

var (
	now       = "2025-05-15T10:32:36Z"
	minTime   = "2020-01-01T00:32:36Z"
	maxTime   = "2025-05-14T06:32:36Z"
	watermark = "2025-05-13T06:32:36Z"
)

func TestEval_PreviousAndCurrentCompleteGrain(t *testing.T) {
	testCases := []testCase{
		// Previous complete second
		{"1s as of watermark/s", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-1s to ref as of watermark/s", "2025-05-13T06:32:35Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMillisecond, 1, 1},
		// Last 2 seconds, including current second
		{"2s as of watermark/s+1s", "2025-05-13T06:32:35Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainSecond, 1, 1},
		{"-2s to ref as of watermark/s+1s", "2025-05-13T06:32:35Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainSecond, 1, 1},
		// Last 2 seconds, excluding current second
		{"2s as of watermark/s", "2025-05-13T06:32:34Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainSecond, 1, 1},
		{"-2s to ref as of watermark/s", "2025-05-13T06:32:34Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainSecond, 1, 1},
		// Current complete second
		{"1s as of watermark/s+1s", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"-1s to ref as of watermark/s+1s", "2025-05-13T06:32:36Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"sTD as of watermark/s+1s", "2025-05-13T06:32:37Z", "2025-05-13T06:32:37Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Previous complete minute
		{"1m as of watermark/m", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-1m to ref as of watermark/m", "2025-05-13T06:31:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainSecond, 1, 1},
		// Last 2 minutes, including current minute
		{"2m as of watermark/m+1m", "2025-05-13T06:31:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-2m to ref as of watermark/m+1m", "2025-05-13T06:31:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainMinute, 1, 1},
		// Last 2 minutes, excluding current minute
		{"2m as of watermark/m", "2025-05-13T06:30:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-2m to ref as of watermark/m", "2025-05-13T06:30:00Z", "2025-05-13T06:32:00Z", timeutil.TimeGrainMinute, 1, 1},
		// Current complete minute
		{"1m as of watermark/m+1m", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"-1m to ref as of watermark/m+1m", "2025-05-13T06:32:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"mTD as of watermark/m+1m", "2025-05-13T06:33:00Z", "2025-05-13T06:33:00Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Previous complete hour
		{"1h as of watermark/h", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-1h to ref as of watermark/h", "2025-05-13T05:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		// Last 2 hours, including current hour
		{"2h as of watermark/h+1h", "2025-05-13T05:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2h to ref as of watermark/h+1h", "2025-05-13T05:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// Last 2 hours, excluding current hour
		{"2h as of watermark/h", "2025-05-13T04:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-2h to ref as of watermark/h", "2025-05-13T04:00:00Z", "2025-05-13T06:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// Current complete hour
		{"1h as of watermark/h+1h", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"-1h to ref as of watermark/h+1h", "2025-05-13T06:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"hTD as of watermark/h+1h", "2025-05-13T07:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Previous complete day
		{"1D as of watermark/D", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D to ref as of watermark/D", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		// Last 2 days, including current day
		{"2D as of watermark/D+1D", "2025-05-12T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2D to ref as of watermark/D+1D", "2025-05-12T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Last 2 days, excluding current day
		{"2D as of watermark/D", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-2D to ref as of watermark/D", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Current complete day
		{"1D as of watermark/D+1D", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D to ref as of watermark/D+1D", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"DTD as of watermark/D+1D", "2025-05-14T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Previous complete week
		{"1W as of watermark/W", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1W to ref as of watermark/W", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Last 2 weeks, including current week
		{"2W as of watermark/W+1W", "2025-05-05T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"-2W to ref as of watermark/W+1W", "2025-05-05T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		// Last 2 weeks, excluding current week
		{"2W as of watermark/W", "2025-04-28T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"-2W to ref as of watermark/W", "2025-04-28T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		// Current complete week
		{"1W as of watermark/W+1W", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1W to ref as of watermark/W+1W", "2025-05-12T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"WTD as of watermark/W+1W", "2025-05-19T00:00:00Z", "2025-05-19T00:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Previous complete month
		{"1M as of watermark/M", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1M to ref as of watermark/M", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Last 2 months, including current month
		{"2M as of watermark/M+1M", "2025-04-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2M to ref as of watermark/M+1M", "2025-04-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Last 2 months, excluding current month
		{"2M as of watermark/M", "2025-03-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-2M to ref as of watermark/M", "2025-03-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Current complete month
		{"1M as of watermark/M+1M", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1M to ref as of watermark/M+1M", "2025-05-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"MTD as of watermark/M+1M", "2025-06-01T00:00:00Z", "2025-06-01T00:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Previous complete quarter
		{"1Q as of watermark/Q", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Q to ref as of watermark/Q", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Last 2 quarters, including current quarter
		{"2Q as of watermark/Q+1Q", "2025-01-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		{"-2Q to ref as of watermark/Q+1Q", "2025-01-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		// Last 2 quarters, excluding current quarter
		{"2Q as of watermark/Q", "2024-10-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		{"-2Q to ref as of watermark/Q", "2024-10-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},
		// Current complete quarter
		{"1Q as of watermark/Q+1Q", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Q to ref as of watermark/Q+1Q", "2025-04-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"QTD as of watermark/Q+1Q", "2025-07-01T00:00:00Z", "2025-07-01T00:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Previous complete year
		{"1Y as of watermark/Y", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Y to ref as of watermark/Y", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		// Last 2 years, including current year
		{"2Y as of watermark/Y+1Y", "2024-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainYear, 1, 1},
		{"-2Y to ref as of watermark/Y+1Y", "2024-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainYear, 1, 1},
		// Last 2 years, excluding current year
		{"2Y as of watermark/Y", "2023-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainYear, 1, 1},
		{"-2Y to ref as of watermark/Y", "2023-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainYear, 1, 1},
		// Current complete year
		{"1Y as of watermark/Y+1Y", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-1Y to ref as of watermark/Y+1Y", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"YTD as of watermark/Y+1Y", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMillisecond, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func TestEval_FirstAndLastOfPeriod(t *testing.T) {
	testCases := []testCase{
		// Last 2 secs of last 2 mins
		{"-2m/m-2s to -2m/m as of watermark/m", "2025-05-13T06:29:58Z", "2025-05-13T06:30:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 secs of last 2 mins
		{"-2m/m to -2m/m+2s as of watermark/m", "2025-05-13T06:30:00Z", "2025-05-13T06:30:02Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Sec 2 of last 2 mins
		{"s2 as of -2m/m as of watermark/m", "2025-05-13T06:30:01Z", "2025-05-13T06:30:02Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Last 2 secs of last 2 hrs
		{"-2h/h-2s to -2h/h as of watermark/h", "2025-05-13T03:59:58Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 secs of last 2 hrs
		{"-2h/h to -2h/h+2s as of watermark/h", "2025-05-13T04:00:00Z", "2025-05-13T04:00:02Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Sec 2 of last 2 hrs
		{"s2 as of -2h/h as of watermark/h", "2025-05-13T04:00:01Z", "2025-05-13T04:00:02Z", timeutil.TimeGrainMillisecond, 1, 1},

		// Last 2 mins of last 2 hrs
		{"-2h/h-2m to -2h/h as of watermark/h", "2025-05-13T03:58:00Z", "2025-05-13T04:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 mins of last 2 hrs
		{"-2h/h to -2h/h+2m as of watermark/h", "2025-05-13T04:00:00Z", "2025-05-13T04:02:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Min 2 of last 2 hrs
		{"m2 as of -2h/h as of watermark/h", "2025-05-13T04:01:00Z", "2025-05-13T04:02:00Z", timeutil.TimeGrainSecond, 1, 1},

		// Last 2 mins of last 2 days
		{"-2D/D-2m to -2D/D as of watermark/D", "2025-05-10T23:58:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 mins of last 2 days
		{"-2D/D to -2D/D+2m as of watermark/D", "2025-05-11T00:00:00Z", "2025-05-11T00:02:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Min 2 of last 2 days
		{"m2 as of -2D/D as of watermark/D", "2025-05-11T00:01:00Z", "2025-05-11T00:02:00Z", timeutil.TimeGrainSecond, 1, 1},

		// Last 2 hrs of last 2 days
		{"-2D/D-2h to -2D/D as of watermark/D", "2025-05-10T22:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 hrs of last 2 days
		{"-2D/D to -2D/D+2h as of watermark/D", "2025-05-11T00:00:00Z", "2025-05-11T02:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Hour 2 of last 2 days
		{"h2 as of -2D/D as of watermark/D", "2025-05-11T01:00:00Z", "2025-05-11T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},

		// Last 2 hrs of last 2 weeks
		{"-2W/W-2h to -2W/W as of watermark/W", "2025-04-27T22:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 hrs of last 2 weeks
		{"-2W/W to -2W/W+2h as of watermark/W", "2025-04-28T00:00:00Z", "2025-04-28T02:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Hour 2 of last 2 weeks
		{"h2 as of -2W/W as of watermark/W", "2025-04-28T01:00:00Z", "2025-04-28T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},

		// Last 2 days of last 2 weeks
		{"-2W/W-2D to -2W/W as of watermark/W", "2025-04-26T00:00:00Z", "2025-04-28T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 days of last 2 weeks
		{"-2W/W to -2W/W+2D as of watermark/W", "2025-04-28T00:00:00Z", "2025-04-30T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Day 2 of last 2 weeks
		{"D2 as of -2W/W as of watermark/W", "2025-04-29T00:00:00Z", "2025-04-30T00:00:00Z", timeutil.TimeGrainHour, 1, 1},

		// Last 2 days of last 2 months
		{"-2M/M-2D to -2M/M as of watermark/M", "2025-02-27T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 days of last 2 months
		{"-2M/M to -2M/M+2D as of watermark/M", "2025-03-01T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Day 2 of last 2 months
		{"D2 as of -2M/M as of watermark/M", "2025-03-02T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainHour, 1, 1},

		// Last 2 weeks of last 2 months
		{"-2M/M/W-2W to -2M/M/W as of watermark/M", "2025-02-17T00:00:00Z", "2025-03-03T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 weeks of last 2 months
		{"-2M/M/W to -2M/M/W+2W as of watermark/M", "2025-03-03T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Week 2 of last 2 months
		{"W2 as of -2M/M as of watermark/M", "2025-03-10T00:00:00Z", "2025-03-17T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Last 2 weeks of last 2 quarters
		{"-2Q/Q/W-2W to -2Q/Q/W as of watermark/Q", "2024-09-16T00:00:00Z", "2024-09-30T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 weeks of last 2 quarters
		{"-2Q/Q/W to -2Q/Q/W+2W as of watermark/Q", "2024-09-30T00:00:00Z", "2024-10-14T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Week 2 of last 2 quarters
		{"W2 as of -2Q/Q as of watermark/Q", "2024-10-07T00:00:00Z", "2024-10-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Last 2 months of last 2 quarters
		{"-2Q/Q-2M to -2Q/Q as of watermark/Q", "2024-08-01T00:00:00Z", "2024-10-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 months of last 2 quarters
		{"-2Q/Q to -2Q/Q+2M as of watermark/Q", "2024-10-01T00:00:00Z", "2024-12-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Month 2 of last 2 quarters
		{"M2 as of -2Q/Q as of watermark/Q", "2024-11-01T00:00:00Z", "2024-12-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Last 2 months of last 2 years
		{"-2Y/Y-2M to -2Y/Y as of watermark/Y", "2022-11-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 months of last 2 years
		{"-2Y/Y to -2Y/Y+2M as of watermark/Y", "2023-01-01T00:00:00Z", "2023-03-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Month 2 of last 2 years
		{"M2 as of -2Y/Y as of watermark/Y", "2023-02-01T00:00:00Z", "2023-03-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Last 2 quarters of last 2 years
		{"-2Y/Y-2Q to -2Y/Y as of watermark/Y", "2022-07-01T00:00:00Z", "2023-01-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// First 2 quarters of last 2 years
		{"-2Y/Y to -2Y/Y+2Q as of watermark/Y", "2023-01-01T00:00:00Z", "2023-07-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},
		// Quarter 2 of last 2 years
		{"Q2 as of -2Y/Y as of watermark/Y", "2023-04-01T00:00:00Z", "2023-07-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func TestEval_OrdinalVariations(t *testing.T) {
	testCases := []testCase{
		{"W1", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W1 as of -2M", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Ordinal chaining variations
		{"s57 of m4 of H2 of D4 as of -1M", "2025-04-04T01:03:56Z", "2025-04-04T01:03:57Z", timeutil.TimeGrainMillisecond, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func TestEval_WeekCorrections(t *testing.T) {
	testCases := []testCase{
		// Boundary on Monday, week starts on Monday
		{"W1 as of 2024-07-01T00:00:00Z", "2024-07-01T00:00:00Z", "2024-07-08T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Monday, week starts on Sunday
		{"W1 as of 2024-07-01T00:00:00Z", "2024-06-30T00:00:00Z", "2024-07-07T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Tuesday, week starts on Monday
		{"W1 as of 2025-04-01T00:00:00Z", "2025-03-31T00:00:00Z", "2025-04-07T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Tuesday, week starts on Sunday
		{"W1 as of 2025-04-01T00:00:00Z", "2025-03-30T00:00:00Z", "2025-04-06T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Wednesday, week starts on Monday
		{"W1 as of 2025-01-01T00:00:00Z", "2024-12-30T00:00:00Z", "2025-01-06T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Wednesday, week starts on Sunday
		{"W1 as of 2025-01-01T00:00:00Z", "2024-12-29T00:00:00Z", "2025-01-05T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Thursday, week starts on Monday
		{"W1 as of 2025-05-01T00:00:00Z", "2025-04-28T00:00:00Z", "2025-05-05T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Thursday, week starts on Sunday
		{"W1 as of 2025-05-01T00:00:00Z", "2025-05-04T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Friday, week starts on Monday
		{"W1 as of 2024-11-01T00:00:00Z", "2024-11-04T00:00:00Z", "2024-11-11T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Friday, week starts on Sunday
		{"W1 as of 2024-11-01T00:00:00Z", "2024-11-03T00:00:00Z", "2024-11-10T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Saturday, week starts on Monday
		{"W1 as of 2025-03-01T00:00:00Z", "2025-03-03T00:00:00Z", "2025-03-10T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Saturday, week starts on Sunday
		{"W1 as of 2025-03-01T00:00:00Z", "2025-03-02T00:00:00Z", "2025-03-09T00:00:00Z", timeutil.TimeGrainDay, 7, 1},

		// Boundary on Sunday, week starts on Monday
		{"W1 as of 2024-12-01T00:00:00Z", "2024-12-02T00:00:00Z", "2024-12-09T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Boundary on Sunday, week starts on Sunday
		{"W1 as of 2024-12-01T00:00:00Z", "2024-12-01T00:00:00Z", "2024-12-08T00:00:00Z", timeutil.TimeGrainDay, 7, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func TestEval_ShorthandSyntax(t *testing.T) {
	testCases := []testCase{
		{"7D", "2025-05-08T10:32:36Z", "2025-05-15T10:32:36Z", timeutil.TimeGrainDay, 1, 1},
		{"7D as of now/D", "2025-05-08T00:00:00Z", "2025-05-15T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"7D as of watermark", "2025-05-06T06:32:36Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainDay, 1, 1},
		{"7D as of watermark/H+1H", "2025-05-06T07:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainDay, 1, 1},

		{"MTD", "2025-05-01T00:00:00Z", "2025-05-15T10:32:36Z", timeutil.TimeGrainDay, 1, 1},
		{"MTD as of watermark", "2025-05-01T00:00:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func TestEval_IsoTimeRanges(t *testing.T) {
	testCases := []testCase{
		{"2025-02-20T01:23:45Z to 2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01:23:45Z / 2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01:23:45Z,2025-07-15T02:34:50Z", "2025-02-20T01:23:45Z", "2025-07-15T02:34:50Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01:23:45Z,2025-07-15T02:34:50Z offset -1P", "2024-09-28T00:12:40Z", "2025-02-20T01:23:45Z", timeutil.TimeGrainSecond, 1, 1},

		{"2025-02-20T01:23", "2025-02-20T01:23:00Z", "2025-02-20T01:24:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01:23 offset -1P", "2025-02-20T01:22:00Z", "2025-02-20T01:23:00Z", timeutil.TimeGrainSecond, 1, 1},
		{"2025-02-20T01", "2025-02-20T01:00:00Z", "2025-02-20T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"2025-02-20T01 offset -1P", "2025-02-20T00:00:00Z", "2025-02-20T01:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"2025-02-20", "2025-02-20T00:00:00Z", "2025-02-21T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"2025-02-20 offset -1P", "2025-02-19T00:00:00Z", "2025-02-20T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"2025-02", "2025-02-01T00:00:00Z", "2025-03-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2025-02 offset -1P", "2025-01-01T00:00:00Z", "2025-02-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2025", "2025-01-01T00:00:00Z", "2026-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"2025 offset -1P", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		{"2025-02-20T01:23:45.123Z to 2025-07-15T02:34:50.123456Z", "2025-02-20T01:23:45.123Z", "2025-07-15T02:34:50.123456Z", timeutil.TimeGrainMillisecond, 1, 1},
		{"2025-02-20T01:23:45.123456Z to 2025-07-15T02:34:50.123456789Z", "2025-02-20T01:23:45.123456Z", "2025-07-15T02:34:50.123456789Z", timeutil.TimeGrainMillisecond, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func TestEval_WatermarkOnBoundary(t *testing.T) {
	maxTimeOnBoundary := "2025-07-01T00:00:00Z"   // month and quarter boundary
	watermarkOnBoundary := "2025-05-12T00:00:00Z" // day and week boundary
	testCases := []testCase{
		{"1h as of watermark/h", "2025-05-11T23:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"1h as of watermark/h+1h", "2025-05-12T00:00:00Z", "2025-05-12T01:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"1D as of watermark/h", "2025-05-11T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"1D as of watermark/h+1h", "2025-05-11T01:00:00Z", "2025-05-12T01:00:00Z", timeutil.TimeGrainHour, 1, 1},

		{"-2D/D to ref/D as of watermark/D+1D", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Simulates comparison for the above
		{"-2D/D to ref/D as of -2D as of watermark/D+1D", "2025-05-09T00:00:00Z", "2025-05-11T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		{"-2D/D to ref/D as of watermark/D", "2025-05-10T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		// Simulates comparison for the above
		{"-2D/D to ref/D as of -2D as of watermark/D", "2025-05-08T00:00:00Z", "2025-05-10T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		{"2D as of watermark/D", "2025-05-10T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2D as of watermark/D+1D", "2025-05-11T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		{"2W as of watermark/W", "2025-04-28T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainWeek, 1, 1},
		{"2M as of watermark/M", "2025-03-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"2Q as of watermark/Q", "2024-10-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainQuarter, 1, 1},

		{"H2 as of watermark/D", "2025-05-12T01:00:00Z", "2025-05-12T02:00:00Z", timeutil.TimeGrainMinute, 1, 1},
		{"D2 as of watermark/W", "2025-05-13T00:00:00Z", "2025-05-14T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D2 as of -1W as of watermark/W", "2025-05-06T00:00:00Z", "2025-05-07T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D2 as of -1W as of now/W", "2025-05-06T00:00:00Z", "2025-05-07T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"W2 as of -1M as of latest/M", "2025-06-09T00:00:00Z", "2025-06-16T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W2 as of -1Q as of latest/Q", "2025-04-07T00:00:00Z", "2025-04-14T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W2 as of -1Y as of 2024", "2023-01-09T00:00:00Z", "2023-01-16T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTimeOnBoundary, watermarkOnBoundary, nil)
}

func Test_KatmanduTimezone(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)

	testCases := []testCase{
		{"2D as of watermark/D", "2025-05-10T18:15:00Z", "2025-05-12T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1D/D to ref/D as of watermark", "2025-05-11T18:15:00Z", "2025-05-12T18:15:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D/D to ref/D as of watermark+1D", "2025-05-12T18:15:00Z", "2025-05-13T18:15:00Z", timeutil.TimeGrainHour, 1, 1},
		{"-1D/D to ref as of watermark", "2025-05-11T18:15:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainHour, 1, 1},

		{"W1 as of watermark", "2025-04-27T18:15:00Z", "2025-05-04T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W1 as of watermark tz Asia/Kathmandu", "2025-04-27T18:15:00Z", "2025-05-04T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W1 as of -2M as of watermark", "2025-03-02T18:15:00Z", "2025-03-09T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
		{"W1 as of -1Y as of watermark", "2024-04-28T18:15:00Z", "2024-05-05T18:15:00Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, tz)
}

func Test_TimeNewYorkTimezone(t *testing.T) {
	tz, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	testCases := []testCase{
		// Cases of time moving forward due to daylight savings
		{"D3 of M11 as of 2024", "2024-11-03T04:00:00Z", "2024-11-04T05:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D4 of M11 as of 2024 offset -1P", "2024-11-03T04:00:00Z", "2024-11-04T05:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D3 of M11 as of 2024 offset -1P", "2024-11-02T04:00:00Z", "2024-11-03T04:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"3D as of 2024-11-04", "2024-11-01T04:00:00Z", "2024-11-04T05:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2M as of 2024-12", "2024-10-01T04:00:00Z", "2024-12-01T05:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"2M as of 2024-12 offset -1P", "2024-08-01T04:00:00Z", "2024-10-01T04:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"2M as of 2024-12 offset -2M", "2024-08-01T04:00:00Z", "2024-10-01T04:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		// Cases of time moving backwards due to daylight savings
		{"D10 of M3 as of 2024", "2024-03-10T05:00:00Z", "2024-03-11T04:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D11 of M3 as of 2024 offset -1P", "2024-03-10T05:00:00Z", "2024-03-11T04:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"D10 of M3 as of 2024 offset -1P", "2024-03-09T05:00:00Z", "2024-03-10T05:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"3D as of 2024-03-11", "2024-03-08T05:00:00Z", "2024-03-11T04:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"2M as of 2024-04", "2024-02-01T05:00:00Z", "2024-04-01T04:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"2M as of 2024-04 offset -1P", "2023-12-01T05:00:00Z", "2024-02-01T05:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"2M as of 2024-04 offset -2M", "2023-12-01T05:00:00Z", "2024-02-01T05:00:00Z", timeutil.TimeGrainMonth, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, tz)
}

func TestEval_BackwardsCompatibility(t *testing.T) {
	testCases := []testCase{
		{"rill-TD", "2025-05-13T00:00:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainHour, 1, 1},
		{"rill-WTD", "2025-05-12T00:00:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainHour, 1, 1},
		{"rill-MTD", "2025-05-01T00:00:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-QTD", "2025-04-01T00:00:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-YTD", "2025-01-01T00:00:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMonth, 1, 1},

		{"rill-PDC", "2025-05-12T00:00:00Z", "2025-05-13T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"rill-PWC", "2025-05-05T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-PMC", "2025-04-01T00:00:00Z", "2025-05-01T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"rill-PQC", "2025-01-01T00:00:00Z", "2025-04-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"rill-PYC", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		// `inf` => `earliest to latest+1s`
		{"inf", "2020-01-01T00:32:36Z", "2025-05-14T06:32:37Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"P2DT10H", "2025-05-10T21:00:00Z", "2025-05-13T07:00:00Z", timeutil.TimeGrainHour, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func TestEval_Misc(t *testing.T) {
	testCases := []testCase{
		// Ending on boundary explicitly
		{"watermark/Y to watermark", "2025-01-01T00:00:00Z", "2025-05-13T06:32:36Z", timeutil.TimeGrainMonth, 1, 1},
		{"latest/Y to latest", "2025-01-01T00:00:00Z", "2025-05-14T06:32:36Z", timeutil.TimeGrainMonth, 1, 1},
		// Now is adjusted ref. Since min_grain is unspecified it defaults to millisecond
		{"now/Y to now", "2025-01-01T00:00:00Z", "2025-05-15T10:32:36Z", timeutil.TimeGrainMonth, 1, 1},
		{"watermark to latest", "2025-05-13T06:32:36Z", "2025-05-14T06:32:36Z", timeutil.TimeGrainUnspecified, 1, 1},
		{"watermark-1Y/Y to watermark/Y", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z", timeutil.TimeGrainUnspecified, 1, 1},

		// `as of` without explicit truncate. Should take the higher order for calculating ordinals
		{"D2 as of -2Y as of watermark", "2023-05-02T00:00:00Z", "2023-05-03T00:00:00Z", timeutil.TimeGrainHour, 1, 1},
		{"W2 as of -2Y as of watermark", "2023-05-08T00:00:00Z", "2023-05-15T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Snapping using `/W` does not correct for ISO week boundary.
		{"-1Y/W to -1Y/W+1W as of 2025-05-17T13:43:00Z", "2024-05-13T00:00:00Z", "2024-05-20T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
		{"-1Y/W to -1y/W+1W as of 2025-05-15T13:43:00Z", "2024-05-13T00:00:00Z", "2024-05-20T00:00:00Z", timeutil.TimeGrainDay, 1, 1},

		// Snapping using `/Y/W` will snap by year and corrects for ISO week boundary.
		{"-2Y/Y/W to -1Y/Y/W as of watermark", "2023-01-02T00:00:00Z", "2024-01-01T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},
		{"-0Y/Y/W to ref/W as of watermark", "2024-12-30T00:00:00Z", "2025-05-12T00:00:00Z", timeutil.TimeGrainMonth, 1, 1},

		// The following 2 ranges are different. -5W4M3Q2Y applies together whereas -5W-4M-3Q-2Y applies separately.
		// This can lead to a slightly different start/end times when weeks are involved.
		{"-5W4M3Q2Y to -4W3M2Q1Y as of watermark", "2022-03-09T06:32:36Z", "2023-07-16T06:32:36Z", timeutil.TimeGrainWeek, 1, 1},
		{"-5W-4M-3Q-2Y to -4W-3M-2Q-1Y as of watermark", "2022-03-08T06:32:36Z", "2023-07-15T06:32:36Z", timeutil.TimeGrainMonth, 1, 1},

		{"3W18D23h as of latest-3Y", "2022-04-04T07:32:36Z", "2022-05-14T06:32:36Z", timeutil.TimeGrainWeek, 1, 1},

		{"7D as of latest/D+1D offset -1M", "2025-04-08T00:00:00Z", "2025-04-15T00:00:00Z", timeutil.TimeGrainDay, 1, 1},
	}

	runTests(t, testCases, now, minTime, maxTime, watermark, nil)
}

func TestEval_SyntaxErrors(t *testing.T) {
	testCases := []struct {
		timeRange string
		errorMsg  string
	}{
		{"D2 of -2Y", `unexpected token "-" (expected Ordinal)`},
		{"-4d", `unexpected token "<EOF>" (expected <to> PointInTime)`},
		{"D", `unexpected token "<EOF>" (expected <number>)`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			_, err := Parse(testCase.timeRange, ParseOptions{})
			require.Error(t, err)
			require.ErrorContains(t, err, testCase.errorMsg)
		})
	}
}

type testCase struct {
	timeRange  string
	start      string
	end        string
	grain      timeutil.TimeGrain
	FirstDay   int
	FirstMonth int
}

func runTests(t *testing.T, testCases []testCase, now, minTime, maxTime, watermark string, tz *time.Location) {
	nowTm := parseTestTime(t, now)
	minTimeTm := parseTestTime(t, minTime)
	maxTimeTm := parseTestTime(t, maxTime)
	watermarkTm := parseTestTime(t, watermark)

	for _, testCase := range testCases {
		t.Run(testCase.timeRange, func(t *testing.T) {
			rt, err := Parse(testCase.timeRange, ParseOptions{
				TimeZoneOverride: tz,
			})
			require.NoError(t, err)

			start, end, grain := rt.Eval(EvalOptions{
				Now:        nowTm,
				MinTime:    minTimeTm,
				MaxTime:    maxTimeTm,
				Watermark:  watermarkTm,
				FirstDay:   testCase.FirstDay,
				FirstMonth: testCase.FirstMonth,
			})
			require.Equal(t, parseTestTime(t, testCase.start), start)
			require.Equal(t, parseTestTime(t, testCase.end), end)
			if testCase.grain != timeutil.TimeGrainUnspecified {
				require.Equal(t, testCase.grain, grain)
			}
		})
	}
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
