package queries

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestResolveTimeRange(t *testing.T) {
	cases := []struct {
		title      string
		tr         *runtimev1.TimeRange
		start, end string
	}{
		{
			"day light savings start US/Canada",
			&runtimev1.TimeRange{End: timeToPB("2023-03-12T12:00:00Z"), IsoDuration: "PT4H", TimeZone: "America/Los_Angeles"},
			"2023-03-12T08:00:00Z",
			"2023-03-12T12:00:00Z",
		},
		{
			"day light savings end US/Canada",
			&runtimev1.TimeRange{Start: timeToPB("2023-11-05T08:00:00.000Z"), IsoDuration: "PT4H", TimeZone: "America/Los_Angeles"},
			"2023-11-05T08:00:00Z",
			"2023-11-05T12:00:00Z",
		},
		{
			"going through feb",
			&runtimev1.TimeRange{Start: timeToPB("2023-01-05T00:00:00Z"), IsoDuration: "P1M"},
			"2023-01-05T00:00:00Z",
			"2023-02-05T00:00:00Z",
		},
		{
			"month-to-date",
			&runtimev1.TimeRange{End: timeToPB("2023-01-10T00:00:00Z"), IsoDuration: "rill-MTD"},
			"2023-01-01T00:00:00Z",
			"2023-01-10T00:00:00Z",
		},
		{
			"month-to-date in Kathmandu",
			&runtimev1.TimeRange{End: timeToPB("2023-01-10T00:00:00Z"), IsoDuration: "rill-MTD", TimeZone: "Asia/Kathmandu"},
			"2022-12-31T18:15:00Z", // since we truncate to beginning of year this is correct
			"2023-01-10T00:00:00Z",
		},
		{
			"previous month",
			&runtimev1.TimeRange{End: timeToPB("2023-01-10T00:00:00Z"), IsoDuration: "rill-PM"},
			"2022-12-10T00:00:00Z",
			"2023-01-10T00:00:00Z",
		},
		{
			"previous month in Kathmandu",
			&runtimev1.TimeRange{End: timeToPB("2023-01-10T00:00:00Z"), IsoDuration: "rill-PM", TimeZone: "Asia/Kathmandu"},
			"2022-12-10T00:00:00Z", // there is no truncation so this -1 month exactly
			"2023-01-10T00:00:00Z",
		},
		{
			"previous month offset",
			&runtimev1.TimeRange{End: timeToPB("2023-01-10T00:00:00Z"), IsoDuration: "P1M", IsoOffset: "rill-PM"},
			"2022-11-10T00:00:00Z",
			"2022-12-10T00:00:00Z",
		},
		{
			"previous month offset in Kathmandu",
			&runtimev1.TimeRange{End: timeToPB("2023-01-10T00:00:00Z"), IsoDuration: "P1M", IsoOffset: "rill-PM", TimeZone: "Asia/Kathmandu"},
			"2022-11-10T00:00:00Z",
			"2022-12-10T00:00:00Z",
		},
		{
			// Simulates UI filling in duration, offset and round to grain instead of sending rill-PMC (previous month complete)
			"previous complete month",
			&runtimev1.TimeRange{End: timeToPB("2023-01-10T00:00:00Z"), IsoDuration: "P1M", RoundToGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH},
			"2022-12-01T00:00:00Z",
			"2023-01-01T00:00:00Z",
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			start, end, err := ResolveTimeRange(tc.tr, &runtimev1.MetricsViewSpec{
				FirstDayOfWeek:   1,
				FirstMonthOfYear: 1,
			})
			require.NoError(t, err)
			require.Equal(t, parseTestTime(t, tc.start), start.UTC())
			require.Equal(t, parseTestTime(t, tc.end), end.UTC())
		})
	}
}

func timeToPB(t string) *timestamppb.Timestamp {
	ts, _ := time.Parse(time.RFC3339, t)
	return timestamppb.New(ts)
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}
