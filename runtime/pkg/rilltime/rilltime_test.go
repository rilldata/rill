package rilltime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Resolve(t *testing.T) {
	now := parseTestTime(t, "2024-08-09T10:32:36Z")
	testCases := []struct {
		timeRange string
		start     string
		end       string
	}{
		{`m : |s|`, "2024-08-09T10:32:00Z", "2024-08-09T10:32:36Z"},
		{`-5m : |m|`, "2024-08-09T10:27:00Z", "2024-08-09T10:33:00Z"},
		{`-5m, 0m : |m|`, "2024-08-09T10:27:00Z", "2024-08-09T10:32:00Z"},
		{`h : m`, "2024-08-09T10:00:00Z", "2024-08-09T10:33:00Z"},
		{`-7d, 0d : |h|`, "2024-08-02T00:00:00Z", "2024-08-09T00:00:00Z"},
		{`-6d, now : |h|`, "2024-08-03T00:00:00Z", "2024-08-10T00:00:00Z"},
		{`-6d, now : h`, "2024-08-03T00:00:00Z", "2024-08-11T00:00:00Z"},
	}

	for _, tc := range testCases {
		t.Run(tc.timeRange, func(t *testing.T) {
			rt, err := Parse(tc.timeRange)
			require.NoError(t, err)

			start, end, err := rt.Resolve(ResolverContext{
				Now:        now,
				MinTime:    now.AddDate(-1, 0, 0),
				MaxTime:    now,
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
