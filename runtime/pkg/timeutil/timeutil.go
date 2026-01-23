package timeutil

import (
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"

	// Load IANA time zone data
	_ "time/tzdata"
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

func TimeGrainFromAPI(tg runtimev1.TimeGrain) TimeGrain {
	switch tg {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return TimeGrainMillisecond
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return TimeGrainSecond
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return TimeGrainMinute
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return TimeGrainHour
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return TimeGrainDay
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return TimeGrainWeek
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return TimeGrainMonth
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return TimeGrainQuarter
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return TimeGrainYear
	}
	return TimeGrainUnspecified
}

func TimeGrainToAPI(tg TimeGrain) runtimev1.TimeGrain {
	switch tg {
	case TimeGrainUnspecified:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
	case TimeGrainMillisecond:
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND
	case TimeGrainSecond:
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND
	case TimeGrainMinute:
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE
	case TimeGrainHour:
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR
	case TimeGrainDay:
		return runtimev1.TimeGrain_TIME_GRAIN_DAY
	case TimeGrainWeek:
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK
	case TimeGrainMonth:
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH
	case TimeGrainQuarter:
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER
	case TimeGrainYear:
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR
	}
	return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
}

func TruncateTime(tm time.Time, tg TimeGrain, tz *time.Location, firstDay, firstMonth int) time.Time {
	switch tg {
	case TimeGrainUnspecified:
		return tm
	case TimeGrainMillisecond:
		return tm.Truncate(time.Millisecond)
	case TimeGrainSecond:
		return tm.Truncate(time.Second)
	case TimeGrainMinute:
		return tm.Truncate(time.Minute)
	case TimeGrainHour:
		previousTimestamp := tm.Add(-time.Hour)      // DST check, ie in NewYork 1:00am can be equal 2:00am
		previousTimestamp = previousTimestamp.In(tz) // if it happens then converting back to UTC loses the hour
		tm = tm.In(tz)
		tm = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), 0, 0, 0, tz)
		utc := tm.In(time.UTC)
		if previousTimestamp.Hour() == tm.Hour() {
			return utc.Add(time.Hour)
		}
		return utc
	case TimeGrainDay:
		tm = tm.In(tz)
		tm = time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tz)
		return tm.In(time.UTC)
	case TimeGrainWeek:
		tm = tm.In(tz)
		weekday := int(tm.Weekday())
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
		tm = time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tz)
		tm = tm.AddDate(0, 0, daysToSubtract)
		return tm.In(time.UTC)
	case TimeGrainMonth:
		tm = tm.In(tz)
		tm = time.Date(tm.Year(), tm.Month(), 1, 0, 0, 0, 0, tz)
		tm = tm.In(time.UTC)
		return tm
	case TimeGrainQuarter:
		if firstMonth < 1 {
			firstMonth = 1
		}
		if firstMonth > 12 {
			firstMonth = 12
		}
		tm = tm.In(tz)
		monthsToSubtract := (3 + int(tm.Month()) - firstMonth%3) % 3
		tm = time.Date(tm.Year(), tm.Month(), 1, 0, 0, 0, 0, tz)
		tm = tm.AddDate(0, -monthsToSubtract, 0)
		return tm.In(time.UTC)
	case TimeGrainYear:
		if firstMonth < 1 {
			firstMonth = 1
		}
		if firstMonth > 12 {
			firstMonth = 12
		}

		tm = tm.In(tz)
		year := tm.Year()
		if int(tm.Month()) < firstMonth {
			year = tm.Year() - 1
		}

		tm = time.Date(year, time.Month(firstMonth), 1, 0, 0, 0, 0, tz)
		return tm.In(time.UTC)
	}

	return tm
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

func OffsetTime(tm time.Time, tg TimeGrain, n int, tz *time.Location) time.Time {
	tm = tm.In(tz)

	switch tg {
	case TimeGrainUnspecified:
	case TimeGrainMillisecond:
		tm = tm.Add(time.Duration(n) * time.Millisecond)
	case TimeGrainSecond:
		tm = tm.Add(time.Duration(n) * time.Second)
	case TimeGrainMinute:
		tm = tm.Add(time.Duration(n) * time.Minute)
	case TimeGrainHour:
		tm = tm.Add(time.Duration(n) * time.Hour)
	case TimeGrainDay:
		tm = tm.AddDate(0, 0, n)
	case TimeGrainWeek:
		tm = tm.AddDate(0, 0, n*7)
	case TimeGrainMonth, TimeGrainQuarter, TimeGrainYear:
		// Offset with correction for different days in months

		yearOffset := 0
		monthOffset := 0
		switch tg {
		case TimeGrainMonth:
			monthOffset = n
		case TimeGrainQuarter:
			monthOffset = n * 3
		case TimeGrainYear:
			yearOffset = n
		default:
			// Won't happen since there is an outer switch
		}

		// `tm` offset as if it were the 1st day of month. Day is applied below based on max days in the month.
		offsetFirstDay := time.Date(tm.Year(), tm.Month(), 1, tm.Hour(), tm.Minute(), tm.Second(), tm.Nanosecond(), tm.Location()).AddDate(yearOffset, monthOffset, 0)

		// Get the max days possible for the month in the year.
		maxDays := daysInMonth(offsetFirstDay.Year(), int(offsetFirstDay.Month()))
		// Take the min of max-days or day from `tm`
		tm = offsetFirstDay.AddDate(0, 0, min(maxDays-1, tm.Day()-1))
	}

	return tm.In(time.UTC)
}

var daysForMonths = []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

func daysInMonth(year, month int) int {
	if month == 2 {
		isLeapYear := year%4 == 0 && (year%100 != 0 || year%400 == 0)
		if isLeapYear {
			return 29
		}
		return 28
	}
	return daysForMonths[month-1]
}
