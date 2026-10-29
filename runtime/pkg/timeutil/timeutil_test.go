package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTruncateTime(t *testing.T) {
	require.Equal(t, parseTestTime(t, "2019-01-07T04:20:07Z"), TruncateTime(parseTestTime(t, "2019-01-07T04:20:07.29Z"), TimeGrainSecond, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T04:20:00Z"), TruncateTime(parseTestTime(t, "2019-01-07T04:20:07Z"), TimeGrainMinute, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T04:00:00Z"), TruncateTime(parseTestTime(t, "2019-01-07T04:20:01Z"), TimeGrainHour, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T00:00:00Z"), TruncateTime(parseTestTime(t, "2019-01-07T04:20:01Z"), TimeGrainDay, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-10-09T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), TimeGrainWeek, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2019-01-07T01:01:01Z"), TimeGrainMonth, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-04-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2019-05-07T01:01:01Z"), TimeGrainQuarter, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2019-02-07T01:01:01Z"), TimeGrainYear, time.UTC, 1, 1))
}

func TestTruncateTimeNewYork(t *testing.T) {
	tz, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	require.Equal(t, parseTestTime(t, "2023-11-05T05:00:01Z"), TruncateTime(parseTestTime(t, "2023-11-05T05:00:01.2Z"), TimeGrainSecond, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-11-05T05:01:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T05:01:01Z"), TimeGrainMinute, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-11-05T05:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T05:20:01Z"), TimeGrainHour, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-11-05T04:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T05:20:01Z"), TimeGrainDay, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-10-30T04:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T05:20:01Z"), TimeGrainWeek, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-11-01T04:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T05:20:01Z"), TimeGrainMonth, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-10-01T04:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T05:20:01Z"), TimeGrainQuarter, tz, 1, 1))

	require.Equal(t, parseTestTime(t, "2023-11-05T05:00:01Z"), TruncateTime(parseTestTime(t, "2023-11-05T05:00:01.2Z"), TimeGrainSecond, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-11-05T06:01:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T06:01:01Z"), TimeGrainMinute, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-11-05T06:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T06:20:01Z"), TimeGrainHour, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-11-05T04:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T06:20:01Z"), TimeGrainDay, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-10-30T04:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T06:20:01Z"), TimeGrainWeek, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-11-01T04:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T06:20:01Z"), TimeGrainMonth, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-10-01T04:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-05T06:20:01Z"), TimeGrainQuarter, tz, 1, 1))
}

func TestTruncateTime_Kathmandu(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2019-01-07T04:20:07Z"), TruncateTime(parseTestTime(t, "2019-01-07T04:20:07.29Z"), TimeGrainSecond, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T04:20:00Z"), TruncateTime(parseTestTime(t, "2019-01-07T04:20:07Z"), TimeGrainMinute, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-07T04:15:00Z"), TruncateTime(parseTestTime(t, "2019-01-07T04:20:01Z"), TimeGrainHour, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-06T18:15:00Z"), TruncateTime(parseTestTime(t, "2019-01-07T04:20:01Z"), TimeGrainDay, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2023-10-08T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), TimeGrainWeek, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-01-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2019-02-07T01:01:01Z"), TimeGrainMonth, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2019-03-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2019-05-07T01:01:01Z"), TimeGrainQuarter, tz, 1, 1))
	require.Equal(t, parseTestTime(t, "2018-12-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2019-02-07T01:01:01Z"), TimeGrainYear, tz, 1, 1))
}

func TestTruncateTime_UTC_first_day(t *testing.T) {
	tz := time.UTC
	require.Equal(t, parseTestTime(t, "2023-10-08T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), TimeGrainWeek, tz, 7, 1))
	require.Equal(t, parseTestTime(t, "2023-10-10T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), TimeGrainWeek, tz, 2, 1))
	require.Equal(t, parseTestTime(t, "2023-10-10T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-11T04:20:01Z"), TimeGrainWeek, tz, 2, 1))
	require.Equal(t, parseTestTime(t, "2023-10-10T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-10T00:01:01Z"), TimeGrainWeek, tz, 2, 1))
}

func TestTruncateTime_Kathmandu_first_day(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2023-10-07T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), TimeGrainWeek, tz, 7, 1))
	require.Equal(t, parseTestTime(t, "2023-10-09T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-10T04:20:01Z"), TimeGrainWeek, tz, 2, 1))
	require.Equal(t, parseTestTime(t, "2023-10-09T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-11T04:20:01Z"), TimeGrainWeek, tz, 2, 1))
	require.Equal(t, parseTestTime(t, "2023-10-09T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-09T18:16:01Z"), TimeGrainWeek, tz, 2, 1))
}

func TestTruncateTime_UTC_first_month(t *testing.T) {
	tz := time.UTC
	require.Equal(t, parseTestTime(t, "2023-08-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 2))
	require.Equal(t, parseTestTime(t, "2023-11-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 5))
	require.Equal(t, parseTestTime(t, "2023-09-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2023-09-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-11-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 6))
	require.Equal(t, parseTestTime(t, "2022-12-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-02-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2022-12-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-02-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 6))

	// With firstMonth=0 (not set), should behave like firstMonth=1: Q2 starts in April
	require.Equal(t, parseTestTime(t, "2023-04-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-04-15T00:20:00Z"), TimeGrainQuarter, tz, 1, 0))
	require.Equal(t, parseTestTime(t, "2023-04-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-05-15T00:20:00Z"), TimeGrainQuarter, tz, 1, 0))
	require.Equal(t, parseTestTime(t, "2023-01-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-01-15T00:20:00Z"), TimeGrainQuarter, tz, 1, 0))

	require.Equal(t, parseTestTime(t, "2023-02-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), TimeGrainYear, tz, 2, 2))
	require.Equal(t, parseTestTime(t, "2023-03-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), TimeGrainYear, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2023-03-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-03-01T00:20:00Z"), TimeGrainYear, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2022-12-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), TimeGrainYear, tz, 2, 12))
	require.Equal(t, parseTestTime(t, "2023-01-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-01-01T00:20:00Z"), TimeGrainYear, tz, 2, 1))

	// Invalid firstMonth
	require.Equal(t, parseTestTime(t, "2023-01-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-01-01T00:20:00Z"), TimeGrainYear, tz, 2, 0))
	require.Equal(t, parseTestTime(t, "2022-12-01T00:00:00Z"), TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), TimeGrainYear, tz, 2, 13))
}

func TestTruncateTime_Kathmandu_first_month(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2023-07-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 2))
	require.Equal(t, parseTestTime(t, "2023-10-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-11-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 5))
	require.Equal(t, parseTestTime(t, "2023-08-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2023-08-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-11-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 6))
	require.Equal(t, parseTestTime(t, "2022-11-30T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-02-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2022-11-30T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-02-01T00:20:00Z"), TimeGrainQuarter, tz, 2, 6))

	require.Equal(t, parseTestTime(t, "2023-01-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-02T00:20:00Z"), TimeGrainYear, tz, 2, 2))
	require.Equal(t, parseTestTime(t, "2023-02-28T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-02T00:20:00Z"), TimeGrainYear, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2023-02-28T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-03-02T00:20:00Z"), TimeGrainYear, tz, 2, 3))
	require.Equal(t, parseTestTime(t, "2022-11-30T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-10-02T00:20:00Z"), TimeGrainYear, tz, 2, 12))
	require.Equal(t, parseTestTime(t, "2022-12-31T18:15:00Z"), TruncateTime(parseTestTime(t, "2023-01-02T00:20:00Z"), TimeGrainYear, tz, 2, 1))
}

func TestOffsetTime(t *testing.T) {
	// Offset to February in a non-leap year
	require.Equal(t, parseTestTime(t, "2025-02-28T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-03-31T10:10:00Z"), TimeGrainMonth, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-02-28T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-05-31T10:10:00Z"), TimeGrainQuarter, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-02-28T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-01-31T10:10:00Z"), TimeGrainMonth, 1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-02-28T10:10:00Z"), OffsetTime(parseTestTime(t, "2024-11-30T10:10:00Z"), TimeGrainQuarter, 1, time.UTC))

	// Offset to February in a leap year
	require.Equal(t, parseTestTime(t, "2024-02-29T10:10:00Z"), OffsetTime(parseTestTime(t, "2024-03-31T10:10:00Z"), TimeGrainMonth, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2024-02-29T10:10:00Z"), OffsetTime(parseTestTime(t, "2024-05-31T10:10:00Z"), TimeGrainQuarter, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2024-02-29T10:10:00Z"), OffsetTime(parseTestTime(t, "2024-01-31T10:10:00Z"), TimeGrainMonth, 1, time.UTC))
	require.Equal(t, parseTestTime(t, "2024-02-29T10:10:00Z"), OffsetTime(parseTestTime(t, "2023-11-30T10:10:00Z"), TimeGrainQuarter, 1, time.UTC))

	// Offset to April (30 days)
	require.Equal(t, parseTestTime(t, "2025-04-30T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-05-31T10:10:00Z"), TimeGrainMonth, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-04-30T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-07-31T10:10:00Z"), TimeGrainQuarter, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-04-30T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-03-31T10:10:00Z"), TimeGrainMonth, 1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-04-30T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-01-31T10:10:00Z"), TimeGrainQuarter, 1, time.UTC))

	// Offset by a year from feb in leap year to non-leap year
	require.Equal(t, parseTestTime(t, "2023-02-28T10:10:00Z"), OffsetTime(parseTestTime(t, "2024-02-29T10:10:00Z"), TimeGrainYear, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-02-28T10:10:00Z"), OffsetTime(parseTestTime(t, "2024-02-29T10:10:00Z"), TimeGrainYear, 1, time.UTC))

	// Offset within max days
	require.Equal(t, parseTestTime(t, "2025-02-20T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-03-20T10:10:00Z"), TimeGrainMonth, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-02-20T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-05-20T10:10:00Z"), TimeGrainQuarter, -1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-02-20T10:10:00Z"), OffsetTime(parseTestTime(t, "2025-01-20T10:10:00Z"), TimeGrainMonth, 1, time.UTC))
	require.Equal(t, parseTestTime(t, "2025-02-20T10:10:00Z"), OffsetTime(parseTestTime(t, "2024-11-20T10:10:00Z"), TimeGrainQuarter, 1, time.UTC))

	tz, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	// Offset through daylight savings
	require.Equal(t, parseTestTime(t, "2024-11-02T04:00:00Z"), OffsetTime(parseTestTime(t, "2024-11-04T05:00:00Z"), TimeGrainDay, -2, tz))
	require.Equal(t, parseTestTime(t, "2024-03-11T04:00:00Z"), OffsetTime(parseTestTime(t, "2024-03-09T05:00:00Z"), TimeGrainDay, 2, tz))
}

func parseTestTime(tst *testing.T, t string) time.Time {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return ts
}

func TestTruncateTimeMultiples(t *testing.T) {
	// UTC: minute multiples floor within the hour, hour multiples within the day.
	require.Equal(t, parseTestTime(t, "2024-03-10T04:20:00Z"), TruncateTime(parseTestTime(t, "2024-03-10T04:27:33Z"), TimeGrain10Minute, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2024-03-10T04:15:00Z"), TruncateTime(parseTestTime(t, "2024-03-10T04:27:33Z"), TimeGrain15Minute, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2024-03-10T04:00:00Z"), TruncateTime(parseTestTime(t, "2024-03-10T04:27:33Z"), TimeGrain30Minute, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2024-03-10T00:00:00Z"), TruncateTime(parseTestTime(t, "2024-03-10T04:27:33Z"), TimeGrain6Hour, time.UTC, 1, 1))
	require.Equal(t, parseTestTime(t, "2024-03-10T12:00:00Z"), TruncateTime(parseTestTime(t, "2024-03-10T17:59:59Z"), TimeGrain12Hour, time.UTC, 1, 1))

	// Asia/Kolkata (+05:30): hour multiples align to local midnight (18:30Z), not to the UTC epoch.
	kolkata, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2024-03-09T18:30:00Z"), TruncateTime(parseTestTime(t, "2024-03-09T23:00:00Z"), TimeGrain6Hour, kolkata, 1, 1))
	require.Equal(t, parseTestTime(t, "2024-03-10T00:30:00Z"), TruncateTime(parseTestTime(t, "2024-03-10T05:00:00Z"), TimeGrain6Hour, kolkata, 1, 1))
	require.Equal(t, parseTestTime(t, "2024-03-10T04:50:00Z"), TruncateTime(parseTestTime(t, "2024-03-10T04:57:00Z"), TimeGrain10Minute, kolkata, 1, 1))

	// Asia/Kathmandu (+05:45): minute multiples floor on the local wall clock, so 10:07 local becomes 10:00 local (04:15Z).
	kathmandu, err := time.LoadLocation("Asia/Kathmandu")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2024-03-10T04:15:00Z"), TruncateTime(parseTestTime(t, "2024-03-10T04:22:00Z"), TimeGrain10Minute, kathmandu, 1, 1))

	// America/New_York across the fall-back change (2024-11-03): 6h buckets stay aligned to local midnight (04:00Z EDT) and the
	// second bucket starts at 06:00 EST (11:00Z), seven real hours later.
	newYork, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2024-11-03T04:00:00Z"), TruncateTime(parseTestTime(t, "2024-11-03T10:30:00Z"), TimeGrain6Hour, newYork, 1, 1))
	require.Equal(t, parseTestTime(t, "2024-11-03T11:00:00Z"), TruncateTime(parseTestTime(t, "2024-11-03T11:30:00Z"), TimeGrain6Hour, newYork, 1, 1))
}

func TestOffsetTimeMultiples(t *testing.T) {
	require.Equal(t, parseTestTime(t, "2024-03-10T04:30:00Z"), OffsetTime(parseTestTime(t, "2024-03-10T04:00:00Z"), TimeGrain10Minute, 3, time.UTC))
	require.Equal(t, parseTestTime(t, "2024-03-10T12:00:00Z"), OffsetTime(parseTestTime(t, "2024-03-10T00:00:00Z"), TimeGrain6Hour, 2, time.UTC))

	// Stepping 6h bins through the New York fall-back change keeps them on local midnight boundaries.
	newYork, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	require.Equal(t, parseTestTime(t, "2024-11-03T11:00:00Z"), OffsetTime(parseTestTime(t, "2024-11-03T04:00:00Z"), TimeGrain6Hour, 1, newYork))
	require.Equal(t, parseTestTime(t, "2024-11-04T05:00:00Z"), OffsetTime(parseTestTime(t, "2024-11-03T04:00:00Z"), TimeGrain6Hour, 4, newYork))
}

func TestGrainBaseAndCompare(t *testing.T) {
	unit, n := TimeGrain10Minute.Base()
	require.Equal(t, TimeGrainMinute, unit)
	require.Equal(t, 10, n)
	unit, n = TimeGrainDay.Base()
	require.Equal(t, TimeGrainDay, unit)
	require.Equal(t, 1, n)

	require.Negative(t, TimeGrainMinute.Compare(TimeGrain5Minute))
	require.Negative(t, TimeGrain30Minute.Compare(TimeGrainHour))
	require.Positive(t, TimeGrain12Hour.Compare(TimeGrainHour))
	require.Negative(t, TimeGrain12Hour.Compare(TimeGrainDay))
	require.Zero(t, TimeGrain6Hour.Compare(TimeGrain6Hour))

	require.Equal(t, 6, ApproximateBins(parseTestTime(t, "2024-03-10T00:00:00Z"), parseTestTime(t, "2024-03-10T01:00:00Z"), TimeGrain10Minute))
	require.Equal(t, 4, ApproximateBins(parseTestTime(t, "2024-03-10T00:00:00Z"), parseTestTime(t, "2024-03-11T00:00:00Z"), TimeGrain6Hour))

	for api, tg := range apiToGrain {
		require.Equal(t, api, TimeGrainToAPI(tg))
	}
}
