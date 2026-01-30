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
