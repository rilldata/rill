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
	// Multiples of a base grain. They are only used to bucket time series.
	// Their numeric values are not the grain order: use Compare to order grains.
	TimeGrain5Minute
	TimeGrain10Minute
	TimeGrain15Minute
	TimeGrain30Minute
	TimeGrain2Hour
	TimeGrain3Hour
	TimeGrain4Hour
	TimeGrain6Hour
	TimeGrain8Hour
	TimeGrain12Hour
)

// multiples maps each multiple grain to its base unit and factor.
// Every factor divides its parent unit (60 minutes, 24 hours), so buckets nest inside hours and days.
var multiples = map[TimeGrain]struct {
	unit TimeGrain
	n    int
}{
	TimeGrain5Minute:  {TimeGrainMinute, 5},
	TimeGrain10Minute: {TimeGrainMinute, 10},
	TimeGrain15Minute: {TimeGrainMinute, 15},
	TimeGrain30Minute: {TimeGrainMinute, 30},
	TimeGrain2Hour:    {TimeGrainHour, 2},
	TimeGrain3Hour:    {TimeGrainHour, 3},
	TimeGrain4Hour:    {TimeGrainHour, 4},
	TimeGrain6Hour:    {TimeGrainHour, 6},
	TimeGrain8Hour:    {TimeGrainHour, 8},
	TimeGrain12Hour:   {TimeGrainHour, 12},
}

var apiToGrain = map[runtimev1.TimeGrain]TimeGrain{
	runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND: TimeGrainMillisecond,
	runtimev1.TimeGrain_TIME_GRAIN_SECOND:      TimeGrainSecond,
	runtimev1.TimeGrain_TIME_GRAIN_MINUTE:      TimeGrainMinute,
	runtimev1.TimeGrain_TIME_GRAIN_HOUR:        TimeGrainHour,
	runtimev1.TimeGrain_TIME_GRAIN_DAY:         TimeGrainDay,
	runtimev1.TimeGrain_TIME_GRAIN_WEEK:        TimeGrainWeek,
	runtimev1.TimeGrain_TIME_GRAIN_MONTH:       TimeGrainMonth,
	runtimev1.TimeGrain_TIME_GRAIN_QUARTER:     TimeGrainQuarter,
	runtimev1.TimeGrain_TIME_GRAIN_YEAR:        TimeGrainYear,
	runtimev1.TimeGrain_TIME_GRAIN_5_MINUTE:    TimeGrain5Minute,
	runtimev1.TimeGrain_TIME_GRAIN_10_MINUTE:   TimeGrain10Minute,
	runtimev1.TimeGrain_TIME_GRAIN_15_MINUTE:   TimeGrain15Minute,
	runtimev1.TimeGrain_TIME_GRAIN_30_MINUTE:   TimeGrain30Minute,
	runtimev1.TimeGrain_TIME_GRAIN_2_HOUR:      TimeGrain2Hour,
	runtimev1.TimeGrain_TIME_GRAIN_3_HOUR:      TimeGrain3Hour,
	runtimev1.TimeGrain_TIME_GRAIN_4_HOUR:      TimeGrain4Hour,
	runtimev1.TimeGrain_TIME_GRAIN_6_HOUR:      TimeGrain6Hour,
	runtimev1.TimeGrain_TIME_GRAIN_8_HOUR:      TimeGrain8Hour,
	runtimev1.TimeGrain_TIME_GRAIN_12_HOUR:     TimeGrain12Hour,
}

var grainToAPI = func() map[TimeGrain]runtimev1.TimeGrain {
	res := make(map[TimeGrain]runtimev1.TimeGrain, len(apiToGrain))
	for k, v := range apiToGrain {
		res[v] = k
	}
	return res
}()

func TimeGrainFromAPI(tg runtimev1.TimeGrain) TimeGrain {
	return apiToGrain[tg]
}

func TimeGrainToAPI(tg TimeGrain) runtimev1.TimeGrain {
	return grainToAPI[tg]
}

// TimeGrainBase returns the base unit of an API grain and the number of those units it spans. Standard grains return themselves and 1.
func TimeGrainBase(tg runtimev1.TimeGrain) (runtimev1.TimeGrain, int) {
	unit, n := TimeGrainFromAPI(tg).Base()
	return TimeGrainToAPI(unit), n
}

// Base returns the unit a grain is a multiple of and the factor. Standard grains return themselves and 1.
func (tg TimeGrain) Base() (TimeGrain, int) {
	if m, ok := multiples[tg]; ok {
		return m.unit, m.n
	}
	return tg, 1
}

// Compare orders grains from fine to coarse by their approximate duration.
// It returns a negative number when tg is finer than other, zero when they are equal, and a positive number when tg is coarser.
func (tg TimeGrain) Compare(other TimeGrain) int {
	a, b := tg.ApproxDuration(), other.ApproxDuration()
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// ApproxDuration returns the nominal length of a grain. Calendar grains use fixed approximations, which is enough to order grains and estimate bin counts.
func (tg TimeGrain) ApproxDuration() time.Duration {
	unit, n := tg.Base()
	var d time.Duration
	switch unit {
	case TimeGrainMillisecond:
		d = time.Millisecond
	case TimeGrainSecond:
		d = time.Second
	case TimeGrainMinute:
		d = time.Minute
	case TimeGrainHour:
		d = time.Hour
	case TimeGrainDay:
		d = 24 * time.Hour
	case TimeGrainWeek:
		d = 7 * 24 * time.Hour
	case TimeGrainMonth:
		d = 30 * 24 * time.Hour
	case TimeGrainQuarter:
		d = 90 * 24 * time.Hour
	case TimeGrainYear:
		d = 365 * 24 * time.Hour
	}
	return d * time.Duration(n)
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

	// Multiples bucket on the local wall clock so buckets align to local midnight, matching the OLAP dialects.
	unit, n := tg.Base()
	if n == 1 {
		return tm
	}
	tm = tm.In(tz)
	switch unit {
	case TimeGrainMinute:
		tm = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute()-tm.Minute()%n, 0, 0, tz)
	case TimeGrainHour:
		tm = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour()-tm.Hour()%n, 0, 0, 0, tz)
	}
	return tm.In(time.UTC)
}

func ApproximateBins(start, end time.Time, tg TimeGrain) int {
	d := tg.ApproxDuration()
	if d == 0 {
		return -1
	}
	return int(end.Sub(start) / d)
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
	default:
		unit, k := tg.Base()
		switch unit {
		case TimeGrainMinute:
			tm = tm.Add(time.Duration(n*k) * time.Minute)
		case TimeGrainHour:
			// Step on the wall clock so bins stay aligned to local midnight across DST changes, like TruncateTime.
			tm = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour()+n*k, tm.Minute(), tm.Second(), tm.Nanosecond(), tz)
		}
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
