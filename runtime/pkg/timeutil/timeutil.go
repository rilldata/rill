package timeutil

import (
	"time"
	// Load IANA time zone data
	_ "time/tzdata"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// TimeGrain is extension of std time package with Week and Quarter added
type TimeGrain int

const (
	TimeGrainUnspecified TimeGrain = iota
	TimeGrainMillisecond
	TimeGrainSecond
	TimeGrainMinute
	TimeGrainHour
	TimeGrainDay
	TimeGrainWeek
	TimeGrainMonth
	TimeGrainQuarter
	TimeGrainYear
)

func TruncateTime(start time.Time, tg TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	switch tg {
	case TimeGrainUnspecified:
		return start
	case TimeGrainMillisecond:
		return start.Truncate(time.Millisecond)
	case TimeGrainSecond:
		return start.Truncate(time.Second)
	case TimeGrainMinute:
		return start.Truncate(time.Minute)
	case TimeGrainHour:
		previousTimestamp := start.Add(-time.Hour)   // DST check, ie in NewYork 1:00am can be equal 2:00am
		previousTimestamp = previousTimestamp.In(tz) // if it happens then converting back to UTC loses the hour
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), 0, 0, 0, tz)
		utc := start.In(time.UTC)
		if previousTimestamp.Hour() == start.Hour() {
			return utc.Add(time.Hour)
		}
		return utc
	case TimeGrainDay:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, tz)
		return start.In(time.UTC)
	case TimeGrainWeek:
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
	case TimeGrainMonth:
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		start = start.In(time.UTC)
		return start
	case TimeGrainQuarter:
		monthsToSubtract := (3 + int(start.Month()) - firstMonth%3) % 3
		start = start.In(tz)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, tz)
		start = start.AddDate(0, -monthsToSubtract, 0)
		return start.In(time.UTC)
	case TimeGrainYear:
		if firstMonth < 1 {
			firstMonth = 1
		}
		if firstMonth > 12 {
			firstMonth = 12
		}

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

func CeilTime(start time.Time, tg TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	truncated := TruncateTime(start, tg, tz, firstDay, firstMonth)
	if start.Equal(truncated) {
		return start
	}

	switch tg {
	case TimeGrainUnspecified, TimeGrainMillisecond:
		return start
	case TimeGrainSecond:
		start = start.Add(time.Second)
	case TimeGrainMinute:
		start = start.Add(time.Minute)
	case TimeGrainHour:
		start = start.Add(time.Hour)
	case TimeGrainDay:
		start = start.AddDate(0, 0, 1)
	case TimeGrainWeek:
		start = start.AddDate(0, 0, 7)
	case TimeGrainMonth:
		start = start.AddDate(0, 1, 0)
	case TimeGrainQuarter:
		start = start.AddDate(0, 3, 0)
	case TimeGrainYear:
		start = start.AddDate(1, 0, 0)
	}

	return TruncateTime(start, tg, tz, firstDay, firstMonth)
}

func ApproximateBins(start, end time.Time, tg TimeGrain) int {
	switch tg {
	case TimeGrainUnspecified:
		return -1
	case TimeGrainMillisecond:
		return int(end.Sub(start) / time.Millisecond)
	case TimeGrainSecond:
		return int(end.Sub(start) / time.Second)
	case TimeGrainMinute:
		return int(end.Sub(start) / time.Minute)
	case TimeGrainHour:
		return int(end.Sub(start) / time.Hour)
	case TimeGrainDay:
		return int(end.Sub(start) / (24 * time.Hour))
	case TimeGrainWeek:
		return int(end.Sub(start) / (7 * 24 * time.Hour))
	case TimeGrainMonth:
		return int(end.Sub(start) / (30 * 24 * time.Hour))
	case TimeGrainQuarter:
		return int(end.Sub(start) / (90 * 24 * time.Hour))
	case TimeGrainYear:
		return int(end.Sub(start) / (365 * 24 * time.Hour))
	}

	return -1
}

func AddTimeProto(to time.Time, tg runtimev1.TimeGrain, count int) time.Time {
	switch tg {
	case runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED:
		return to
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return to.Add(time.Duration(count) * time.Millisecond)
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return to.Add(time.Duration(count) * time.Second)
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return to.Add(time.Duration(count) * time.Minute)
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return to.Add(time.Duration(count) * time.Hour)
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return to.AddDate(0, 0, count)
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return to.AddDate(0, 0, count*7)
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return to.AddDate(0, count, 0)
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return to.AddDate(0, count*3, 0)
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return to.AddDate(count, 0, 0)
	}

	return to
}

func OffsetTime(tm time.Time, tg TimeGrain, n int) time.Time {
	switch tg {
	case TimeGrainUnspecified:
		return tm
	case TimeGrainMillisecond:
		return tm.Add(time.Duration(n) * time.Millisecond)
	case TimeGrainSecond:
		return tm.Add(time.Duration(n) * time.Second)
	case TimeGrainMinute:
		return tm.Add(time.Duration(n) * time.Minute)
	case TimeGrainHour:
		return tm.Add(time.Duration(n) * time.Hour)
	case TimeGrainDay:
		return tm.AddDate(0, 0, n)
	case TimeGrainWeek:
		return tm.AddDate(0, 0, n*7)
	case TimeGrainMonth:
		return tm.AddDate(0, n, 0)
	case TimeGrainQuarter:
		return tm.AddDate(0, n*3, 0)
	case TimeGrainYear:
		return tm.AddDate(n, 0, 0)
	}

	return tm
}
