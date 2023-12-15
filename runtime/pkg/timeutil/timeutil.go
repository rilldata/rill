package timeutil

import (
	"time"

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
