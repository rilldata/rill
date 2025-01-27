package rilltime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Resolve(t *testing.T) {
	now := parseTestTime(t, "2024-08-16T10:32:36Z")
	minTime := parseTestTime(t, "2024-01-01T00:32:36Z")
	maxTime := parseTestTime(t, "2024-08-06T06:32:36Z")
	watermark := parseTestTime(t, "2024-08-05T06:32:36Z")
	testCases := []struct {
		timeRange string
		start     string
		end       string
	}{
		// Earliest = 2023-08-09T10:32:36Z, Latest = 2024-08-06T06:32:36Z, = Now = 2024-08-09T10:32:36Z
		{`m : |s|`, "2024-08-05T06:32:00Z", "2024-08-05T06:32:36Z"},
		{`m : s`, "2024-08-05T06:32:00Z", "2024-08-05T06:32:37Z"},
		{`-5m : |m|`, "2024-08-05T06:27:00Z", "2024-08-05T06:32:00Z"},
		{`-5m, 0m : |m|`, "2024-08-05T06:27:00Z", "2024-08-05T06:32:00Z"},
		{`h : m`, "2024-08-05T06:00:00Z", "2024-08-05T06:33:00Z"},
		{`-7d, 0d : |h|`, "2024-07-29T00:00:00Z", "2024-08-05T00:00:00Z"},
		{`-7d, now/d : |h|`, "2024-08-09T00:00:00Z", "2024-08-16T00:00:00Z"},
		{`-6d, now : |h|`, "2024-08-10T00:00:00Z", "2024-08-16T10:00:00Z"},
		{`-6d, now : h`, "2024-08-10T00:00:00Z", "2024-08-16T11:00:00Z"},

		{`0Y, now @ {UTC}`, "2024-01-01T00:00:00Z", "2024-08-16T10:32:37Z"},
		{`0Y, latest @ {UTC}`, "2024-01-01T00:00:00Z", "2024-08-06T06:32:37Z"},
		{`0Y, watermark @ {UTC}`, "2024-01-01T00:00:00Z", "2024-08-05T06:32:37Z"},
		{`0y, watermark @ {UTC}`, "2024-01-01T00:00:00Z", "2024-08-05T06:32:37Z"},

		{`0d : h`, "2024-08-05T00:00:00Z", "2024-08-05T07:00:00Z"},
		{`0d : h @ -1d`, "2024-08-04T00:00:00Z", "2024-08-04T07:00:00Z"},
		{`0d : h @ now`, "2024-08-16T00:00:00Z", "2024-08-16T11:00:00Z"},

		{`-7d, -5d : h`, "2024-07-29T00:00:00Z", "2024-07-31T00:00:00Z"},
		{`watermark-7d, watermark-5d : h`, "2024-07-29T00:00:00Z", "2024-07-31T00:00:00Z"},
		{`-2d, now/d : h @ -5d`, "2024-08-09T00:00:00Z", "2024-08-11T00:00:00Z"},
		{`-2d, now/d @ -5d`, "2024-08-09T00:00:00Z", "2024-08-11T00:00:00Z"},
		{`-7d, -5d @ now`, "2024-08-09T00:00:00Z", "2024-08-11T00:00:00Z"},

		{`watermark-7D, watermark : h`, "2024-07-29T00:00:00Z", "2024-08-05T07:00:00Z"},

		{`-7d, now/d : h @ {Asia/Kathmandu}`, "2024-08-08T18:15:00Z", "2024-08-15T18:15:00Z"},
		{`-7d, now/d : |h| @ {Asia/Kathmandu}`, "2024-08-08T18:15:00Z", "2024-08-15T18:15:00Z"},
		{`-7d, now/d : |h| @ -5d {Asia/Kathmandu}`, "2024-08-03T18:15:00Z", "2024-08-10T18:15:00Z"},

		{`-7d, latest/d : |h|`, "2024-07-30T00:00:00Z", "2024-08-06T00:00:00Z"},
		{`-6d, latest : |h|`, "2024-07-31T00:00:00Z", "2024-08-06T06:00:00Z"},
		{`-6d, latest : h`, "2024-07-31T00:00:00Z", "2024-08-06T07:00:00Z"},
		{`2024-03-01-7d, 2024-03-01`, "2024-02-23T00:00:00Z", "2024-03-01T00:00:00Z"},
		{`2024-03-01-7d, 2024-03-01 @-2d`, "2024-02-21T00:00:00Z", "2024-02-28T00:00:00Z"},

		{`2024-01-01, latest : h`, "2024-01-01T00:00:00Z", "2024-08-06T07:00:00Z"},
		{`2024-01-01 12:00, latest : h`, "2024-01-01T12:00:00Z", "2024-08-06T07:00:00Z"},

		{`2024-01-01+5d, latest : h`, "2024-01-06T00:00:00Z", "2024-08-06T07:00:00Z"},
		{`-7W+5d, latest : h`, "2024-06-22T00:00:00Z", "2024-08-06T07:00:00Z"},
		{`-7w+5D, latest : h`, "2024-06-22T00:00:00Z", "2024-08-06T07:00:00Z"},
		{`-7W+8d, latest : h`, "2024-06-25T00:00:00Z", "2024-08-06T07:00:00Z"},

		{`0W, 0W+1D`, "2024-08-05T00:00:00Z", "2024-08-06T00:00:00Z"},
		{`watermark/W, watermark/W+1D`, "2024-08-05T00:00:00Z", "2024-08-06T00:00:00Z"},
		{`0W, 0W+1D @ latest`, "2024-08-05T00:00:00Z", "2024-08-06T00:00:00Z"},
		{`0W, 0W+1D @ now`, "2024-08-12T00:00:00Z", "2024-08-13T00:00:00Z"},
		{`now/W, now/W+1D`, "2024-08-12T00:00:00Z", "2024-08-13T00:00:00Z"},

		{"P2DT10H", "2024-08-03T20:00:00Z", "2024-08-06T07:32:36Z"},
		{"rill-MTD", "2024-08-01T00:00:00Z", "2024-08-06T06:32:37Z"},
		{"rill-PW", "2024-07-29T00:00:00Z", "2024-08-05T00:00:00Z"},
	}

	for _, tc := range testCases {
		t.Run(tc.timeRange, func(t *testing.T) {
			rillTime, err := Parse(tc.timeRange, ParseOptions{})
			require.NoError(t, err)

			start, end, err := rillTime.Eval(EvalOptions{
				Now:        now,
				MinTime:    minTime,
				MaxTime:    maxTime,
				Watermark:  watermark,
				FirstDay:   1,
				FirstMonth: 1,
			})
			require.NoError(t, err)
			require.Equal(t, parseTestTime(t, tc.start), start)
			require.Equal(t, parseTestTime(t, tc.end), end)
		})
	}
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
