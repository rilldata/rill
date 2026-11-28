package duration

import (
	"testing"
	"time"

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
		{from: "rill-PM", expected: StandardDuration{Month: 1, rillExtension: "PM"}},
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

func TestTruncate(t *testing.T) {
	halfHourZone := time.FixedZone("IST", 5*3600+1800)
	quarterHourZone := time.FixedZone("NPT", 5*3600+2700)
	copenhagen := must(time.LoadLocation("Europe/Copenhagen"))

	// Test case structure
	type testCase struct {
		name             string
		duration         StandardDuration
		inputTime        time.Time
		firstDayOfWeek   int
		firstMonthOfYear int
		expectedTime     time.Time
	}

	// Define your test cases
	testCases := []testCase{
		{
			name:             "Truncate to start of year",
			duration:         StandardDuration{Year: 1},
			inputTime:        time.Date(2022, 5, 15, 10, 30, 45, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:             "Truncate to start of month",
			duration:         StandardDuration{Month: 1},
			inputTime:        time.Date(2022, 5, 15, 10, 30, 45, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:             "Truncate to start of week",
			duration:         StandardDuration{Week: 2},
			inputTime:        time.Date(2022, 5, 15, 10, 30, 45, 0, time.UTC), // 15th May 2022 is a Sunday
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 5, 9, 0, 0, 0, 0, time.UTC), // 9th May 2022 is a Monday
		},
		{
			name:             "Truncate to start of same day",
			duration:         StandardDuration{Day: 1},
			inputTime:        time.Date(2022, 7, 15, 10, 30, 45, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:             "Truncate to same start of day",
			duration:         StandardDuration{Day: 1},
			inputTime:        time.Date(2022, 7, 15, 0, 0, 0, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:             "Truncate to start of same hour",
			duration:         StandardDuration{Hour: 2},
			inputTime:        time.Date(2022, 7, 15, 11, 30, 45, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:             "Truncate to same start of hour",
			duration:         StandardDuration{Hour: 2},
			inputTime:        time.Date(2022, 7, 15, 11, 30, 0, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			name:             "Truncate to start of hour with spring DST",
			duration:         StandardDuration{Hour: 1},
			inputTime:        time.Date(2023, 3, 26, 3, 15, 0, 0, must(time.LoadLocation("Europe/Copenhagen"))),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2023, 3, 26, 3, 0, 0, 0, must(time.LoadLocation("Europe/Copenhagen"))),
		},
		{
			name:             "Truncate to start of hour with fall DST (first occurrence, CEST)",
			duration:         StandardDuration{Hour: 1},
			inputTime:        time.Date(2023, 10, 29, 0, 15, 0, 0, time.UTC).In(copenhagen),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2023, 10, 29, 0, 0, 0, 0, time.UTC).In(copenhagen),
		},
		{
			name:             "Truncate to start of hour with fall DST (second occurrence, CET)",
			duration:         StandardDuration{Hour: 1},
			inputTime:        time.Date(2023, 10, 29, 1, 15, 0, 0, time.UTC).In(copenhagen),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2023, 10, 29, 1, 0, 0, 0, time.UTC).In(copenhagen),
		},
		{
			name:             "Truncate to start of same minute",
			duration:         StandardDuration{Minute: 1},
			inputTime:        time.Date(2022, 7, 15, 10, 30, 45, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:             "Truncate to same start of minute",
			duration:         StandardDuration{Minute: 1},
			inputTime:        time.Date(2022, 7, 15, 10, 30, 0, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:             "Truncate to start of same second",
			duration:         StandardDuration{Second: 1},
			inputTime:        time.Date(2022, 7, 15, 10, 30, 45, 500, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 10, 30, 45, 0, time.UTC),
		},
		{
			name:             "Truncate to same start of second",
			duration:         StandardDuration{Second: 1},
			inputTime:        time.Date(2022, 7, 15, 10, 30, 45, 0, time.UTC),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 10, 30, 45, 0, time.UTC),
		},
		{
			name:             "Truncate to start of week with Sunday as first day",
			duration:         StandardDuration{Week: 1},
			inputTime:        time.Date(2022, 7, 15, 10, 30, 45, 0, time.UTC), // 15th July 2022 is a Friday
			firstDayOfWeek:   7,                                               // Sunday
			firstMonthOfYear: 1,                                               // January
			expectedTime:     time.Date(2022, 7, 10, 0, 0, 0, 0, time.UTC),    // 10th July 2022 is a Sunday
		},
		{
			name:             "Truncate to start of year with February as first month",
			duration:         StandardDuration{Year: 1},
			inputTime:        time.Date(2022, 3, 15, 10, 30, 45, 0, time.UTC),
			firstDayOfWeek:   1, // Monday
			firstMonthOfYear: 2, // February
			expectedTime:     time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			// t.Truncate(time.Hour) aligns to UTC hour boundaries, which are at :30 in a +05:30 zone.
			name:             "Truncate to start of hour in half-hour offset zone",
			duration:         StandardDuration{Hour: 2},
			inputTime:        time.Date(2022, 7, 15, 10, 45, 0, 0, halfHourZone),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 10, 0, 0, 0, halfHourZone),
		},
		{
			// The UTC truncation can even land in the previous local hour in a +05:45 zone.
			name:             "Truncate to start of hour in quarter-hour offset zone",
			duration:         StandardDuration{Hour: 2},
			inputTime:        time.Date(2022, 7, 15, 10, 15, 0, 0, quarterHourZone),
			firstDayOfWeek:   1,
			firstMonthOfYear: 1,
			expectedTime:     time.Date(2022, 7, 15, 10, 0, 0, 0, quarterHourZone),
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualTime := tc.duration.Truncate(tc.inputTime, tc.firstDayOfWeek, tc.firstMonthOfYear)
			require.Equal(t, tc.expectedTime, actualTime)
		})
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
