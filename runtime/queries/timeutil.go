package queries

import (
	"time"
	// Load IANA time zone data
	_ "time/tzdata"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

func TruncateTime(start time.Time, tg runtimev1.TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	switch tg {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return start.Truncate(time.Millisecond)
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return start.Truncate(time.Second)
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return start.Truncate(time.Minute)
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, tz)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, tz)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		start = start.In(tz)
		weekday := int(start.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if firstDay < 1 {
			firstDay = 1
		}
		if firstDay > 7 {
			firstDay = 7
		}

		daysToSubtract := -(weekday - firstDay)
		if weekday < firstDay {
			daysToSubtract = -7 + daysToSubtract
		}
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, tz)
		start = start.AddDate(0, 0, daysToSubtract)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		start = start.In(time.UTC)
		return start
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		monthsToSubtract := 1 - int(start.Month())%3 // todo first month of year
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		start = start.AddDate(0, monthsToSubtract, 0)
		return start.In(time.UTC)
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		start = start.In(tz)
		year := start.Year()
		if int(start.Month()) < firstMonth {
			year = start.Year() - 1
		}

		start = time.Date(year, time.Month(firstMonth), 1, 0, 0, 0, 0, tz)
		return start.In(time.UTC)
	}

	return start
}

func CeilTime(start time.Time, tg runtimev1.TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	switch tg {
	case runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED:
		return start
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		start = start.Add(time.Millisecond)
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		start = start.Add(time.Second)
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		start = start.Add(time.Minute)
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		start = start.Add(time.Hour)
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		start = start.AddDate(0, 0, 1)
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		start = start.AddDate(0, 0, 7)
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		start = start.AddDate(0, 1, 0)
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		start = start.AddDate(0, 3, 0)
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		start = start.AddDate(1, 0, 0)
	}
	return TruncateTime(start, tg, tz, firstDay, firstMonth)
}
