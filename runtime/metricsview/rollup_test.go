package metricsview

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestGrainDerivableFrom(t *testing.T) {
	tests := []struct {
		name     string
		query    runtimev1.TimeGrain
		rollup   runtimev1.TimeGrain
		expected bool
	}{
		// Same grain
		{"day from day", runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_DAY, true},
		{"hour from hour", runtimev1.TimeGrain_TIME_GRAIN_HOUR, runtimev1.TimeGrain_TIME_GRAIN_HOUR, true},

		// Finer to coarser on sub-day branch
		{"hour from minute", runtimev1.TimeGrain_TIME_GRAIN_HOUR, runtimev1.TimeGrain_TIME_GRAIN_MINUTE, true},
		{"day from hour", runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_HOUR, true},
		{"day from ms", runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, true},

		// Sub-day/day feeds into week branch
		{"week from day", runtimev1.TimeGrain_TIME_GRAIN_WEEK, runtimev1.TimeGrain_TIME_GRAIN_DAY, true},
		{"week from hour", runtimev1.TimeGrain_TIME_GRAIN_WEEK, runtimev1.TimeGrain_TIME_GRAIN_HOUR, true},

		// Sub-day/day feeds into calendar branch
		{"month from day", runtimev1.TimeGrain_TIME_GRAIN_MONTH, runtimev1.TimeGrain_TIME_GRAIN_DAY, true},
		{"quarter from day", runtimev1.TimeGrain_TIME_GRAIN_QUARTER, runtimev1.TimeGrain_TIME_GRAIN_DAY, true},
		{"year from day", runtimev1.TimeGrain_TIME_GRAIN_YEAR, runtimev1.TimeGrain_TIME_GRAIN_DAY, true},
		{"year from hour", runtimev1.TimeGrain_TIME_GRAIN_YEAR, runtimev1.TimeGrain_TIME_GRAIN_HOUR, true},

		// Calendar branch internal
		{"quarter from month", runtimev1.TimeGrain_TIME_GRAIN_QUARTER, runtimev1.TimeGrain_TIME_GRAIN_MONTH, true},
		{"year from month", runtimev1.TimeGrain_TIME_GRAIN_YEAR, runtimev1.TimeGrain_TIME_GRAIN_MONTH, true},
		{"year from quarter", runtimev1.TimeGrain_TIME_GRAIN_YEAR, runtimev1.TimeGrain_TIME_GRAIN_QUARTER, true},

		// Cross-branch: not derivable
		{"month from week", runtimev1.TimeGrain_TIME_GRAIN_MONTH, runtimev1.TimeGrain_TIME_GRAIN_WEEK, false},
		{"quarter from week", runtimev1.TimeGrain_TIME_GRAIN_QUARTER, runtimev1.TimeGrain_TIME_GRAIN_WEEK, false},
		{"year from week", runtimev1.TimeGrain_TIME_GRAIN_YEAR, runtimev1.TimeGrain_TIME_GRAIN_WEEK, false},

		// Week from month: not derivable (coarser to finer)
		{"week from month", runtimev1.TimeGrain_TIME_GRAIN_WEEK, runtimev1.TimeGrain_TIME_GRAIN_MONTH, false},

		// Coarser to finer: not derivable
		{"day from week", runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_WEEK, false},
		{"hour from day", runtimev1.TimeGrain_TIME_GRAIN_HOUR, runtimev1.TimeGrain_TIME_GRAIN_DAY, false},
		{"minute from hour", runtimev1.TimeGrain_TIME_GRAIN_MINUTE, runtimev1.TimeGrain_TIME_GRAIN_HOUR, false},
		{"day from month", runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_MONTH, false},

		// Unspecified
		{"unspecified query", runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, runtimev1.TimeGrain_TIME_GRAIN_DAY, false},
		{"unspecified rollup", runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GrainDerivableFrom(tt.query, tt.rollup)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestTimeRangeAligned(t *testing.T) {
	utc := time.UTC
	eastern, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name           string
		start          time.Time
		end            time.Time
		grain          runtimev1.TimeGrain
		tz             *time.Location
		firstDayOfWeek uint32
		expected       bool
	}{
		{
			"day aligned UTC",
			time.Date(2024, 1, 1, 0, 0, 0, 0, utc),
			time.Date(2024, 1, 2, 0, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_DAY, utc, 1, true,
		},
		{
			"day not aligned UTC",
			time.Date(2024, 1, 1, 12, 0, 0, 0, utc),
			time.Date(2024, 1, 2, 0, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_DAY, utc, 1, false,
		},
		{
			"hour aligned",
			time.Date(2024, 1, 1, 5, 0, 0, 0, utc),
			time.Date(2024, 1, 1, 10, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_HOUR, utc, 1, true,
		},
		{
			"hour not aligned",
			time.Date(2024, 1, 1, 5, 30, 0, 0, utc),
			time.Date(2024, 1, 1, 10, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_HOUR, utc, 1, false,
		},
		{
			"month aligned",
			time.Date(2024, 1, 1, 0, 0, 0, 0, utc),
			time.Date(2024, 4, 1, 0, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_MONTH, utc, 1, true,
		},
		{
			"month not aligned",
			time.Date(2024, 1, 15, 0, 0, 0, 0, utc),
			time.Date(2024, 4, 1, 0, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_MONTH, utc, 1, false,
		},
		{
			"quarter aligned",
			time.Date(2024, 1, 1, 0, 0, 0, 0, utc),
			time.Date(2024, 7, 1, 0, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_QUARTER, utc, 1, true,
		},
		{
			"quarter not aligned",
			time.Date(2024, 2, 1, 0, 0, 0, 0, utc),
			time.Date(2024, 7, 1, 0, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_QUARTER, utc, 1, false,
		},
		{
			"year aligned",
			time.Date(2024, 1, 1, 0, 0, 0, 0, utc),
			time.Date(2025, 1, 1, 0, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_YEAR, utc, 1, true,
		},
		{
			"week aligned Monday",
			time.Date(2024, 1, 1, 0, 0, 0, 0, utc), // Monday
			time.Date(2024, 1, 8, 0, 0, 0, 0, utc),  // Monday
			runtimev1.TimeGrain_TIME_GRAIN_WEEK, utc, 1, true,
		},
		{
			"week not aligned Monday",
			time.Date(2024, 1, 2, 0, 0, 0, 0, utc), // Tuesday
			time.Date(2024, 1, 8, 0, 0, 0, 0, utc),  // Monday
			runtimev1.TimeGrain_TIME_GRAIN_WEEK, utc, 1, false,
		},
		{
			"week aligned Sunday (fdow=7)",
			time.Date(2023, 12, 31, 0, 0, 0, 0, utc), // Sunday
			time.Date(2024, 1, 7, 0, 0, 0, 0, utc),    // Sunday
			runtimev1.TimeGrain_TIME_GRAIN_WEEK, utc, 7, true,
		},
		{
			"day aligned eastern timezone",
			time.Date(2024, 1, 1, 5, 0, 0, 0, utc), // midnight EST
			time.Date(2024, 1, 2, 5, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_DAY, eastern, 1, true,
		},
		{
			"zero start time",
			time.Time{},
			time.Date(2024, 1, 1, 0, 0, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_DAY, utc, 1, true,
		},
		{
			"second aligned",
			time.Date(2024, 1, 1, 0, 0, 5, 0, utc),
			time.Date(2024, 1, 1, 0, 0, 10, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_SECOND, utc, 1, true,
		},
		{
			"second not aligned",
			time.Date(2024, 1, 1, 0, 0, 5, 500, utc),
			time.Date(2024, 1, 1, 0, 0, 10, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_SECOND, utc, 1, false,
		},
		{
			"minute aligned",
			time.Date(2024, 1, 1, 0, 5, 0, 0, utc),
			time.Date(2024, 1, 1, 0, 10, 0, 0, utc),
			runtimev1.TimeGrain_TIME_GRAIN_MINUTE, utc, 1, true,
		},
		{
			"millisecond aligned",
			time.Date(2024, 1, 1, 0, 0, 0, 5_000_000, utc),
			time.Date(2024, 1, 1, 0, 0, 0, 10_000_000, utc),
			runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, utc, 1, true,
		},
		{
			"millisecond not aligned",
			time.Date(2024, 1, 1, 0, 0, 0, 5_500, utc),
			time.Date(2024, 1, 1, 0, 0, 0, 10_000_000, utc),
			runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, utc, 1, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimeRangeAligned(tt.start, tt.end, tt.grain, tt.tz, tt.firstDayOfWeek)
			require.Equal(t, tt.expected, got)
		})
	}
}
