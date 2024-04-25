// Package duration copied as it is from github.com/senseyeio/duration
package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

type Duration interface {
	Add(t time.Time) time.Time
	Sub(t time.Time) time.Time
	EstimateNative() (time.Duration, bool)
}

// Regexes used by ParseISO8601
var (
	infPattern         = regexp.MustCompile("^(?i)inf$")
	durationPattern    = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
	daxToDateNotations = map[string]timeutil.TimeGrain{
		// Pulled from https://www.daxpatterns.com/standard-time-related-calculations/
		// Add more here once support in UI has been added
		"TD":  timeutil.TimeGrainDay,
		"WTD": timeutil.TimeGrainWeek,
		"MTD": timeutil.TimeGrainMonth,
		"QTD": timeutil.TimeGrainQuarter,
		"YTD": timeutil.TimeGrainYear,
	}
	daxOffsetNotations = map[string]StandardDuration{
		"PP": {},
		"PD": {Day: 1, rillExtension: "PD"},
		"PW": {Week: 1, rillExtension: "PW"},
		"PM": {Month: 1, rillExtension: "PM"},
		"PQ": {Month: 3, rillExtension: "PQ"},
		"PY": {Year: 1, rillExtension: "PY"},
	}
	daxOffsetRangeNotations = map[string]TruncToDateDuration{
		// TODO: add mapping with offset to support these in places where only backend is involved like reports/alerts
		"PDC": {timeutil.TimeGrainDay},
		"PWC": {timeutil.TimeGrainWeek},
		"PMC": {timeutil.TimeGrainMonth},
		"PQC": {timeutil.TimeGrainQuarter},
		"PYC": {timeutil.TimeGrainYear},
	}
)

// ParseISO8601 parses an ISO8601 duration as well as some Rill-specific extensions.
// (Section 3.7 of the standard supposedly allows extensions that do not interfere with the standard.)
// Current extensions are,
// 1. "inf" for representing an unbounded duration of time
// 2. X-To-Date and Previous-X duration supports with a prefix of "rill-" to DAX notations. Pulled from https://www.daxpatterns.com/standard-time-related-calculations/
func ParseISO8601(from string) (Duration, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return InfDuration{}, nil
	}

	if strings.Contains(from, "rill-") {
		// We are using "rill-" as a prefix to DAX notation so that it doesn't interfere with ISO8601 standard.
		// Pulled from https://www.daxpatterns.com/standard-time-related-calculations/
		rillDur := strings.Replace(from, "rill-", "", 1)
		if a, ok := daxToDateNotations[rillDur]; ok {
			return TruncToDateDuration{anchor: a}, nil
		}
		if o, ok := daxOffsetNotations[rillDur]; ok {
			return o, nil
		}
		if o, ok := daxOffsetRangeNotations[rillDur]; ok {
			return o, nil
		}
	}

	// Parse as a regular ISO8601 duration
	if !durationPattern.MatchString(from) {
		return StandardDuration{}, fmt.Errorf("string %q is not a valid ISO 8601 duration", from)
	}

	var d StandardDuration
	match := durationPattern.FindStringSubmatch(from)
	for i, name := range durationPattern.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return StandardDuration{}, err
		}
		switch name {
		case "year":
			d.Year = val
		case "month":
			d.Month = val
		case "week":
			d.Week = val
		case "day":
			d.Day = val
		case "hour":
			d.Hour = val
		case "minute":
			d.Minute = val
		case "second":
			d.Second = val
		default:
			return d, fmt.Errorf("unexpected field %q in duration", name)
		}
	}

	return d, nil
}

// StandardDuration represents an ISO8601 duration with Rill-specific extensions.
// See ParseISO8601 for details.
type StandardDuration struct {
	Year   int
	Month  int
	Week   int
	Day    int
	Hour   int
	Minute int
	Second int

	rillExtension string
}

// Add adds the duration to a timestamp
func (d StandardDuration) Add(t time.Time) time.Time {
	days := 7*d.Week + d.Day
	t = t.AddDate(d.Year, d.Month, days)

	td := time.Duration(d.Second)*time.Second + time.Duration(d.Minute)*time.Minute + time.Duration(d.Hour)*time.Hour
	return t.Add(td)
}

// Sub subtracts the duration from a timestamp
func (d StandardDuration) Sub(t time.Time) time.Time {
	days := 7*d.Week + d.Day
	t = t.AddDate(-d.Year, -d.Month, -days)

	td := time.Duration(d.Second)*time.Second + time.Duration(d.Minute)*time.Minute + time.Duration(d.Hour)*time.Hour
	return t.Add(-td)
}

// EstimateNative estimates the duration as a native Go duration.
func (d StandardDuration) EstimateNative() (time.Duration, bool) {
	res := time.Duration(d.Second)*time.Second + time.Duration(d.Minute)*time.Minute + time.Duration(d.Hour)*time.Hour
	res += time.Duration(d.Day) * 24 * time.Hour
	res += time.Duration(d.Week*7) * 24 * time.Hour
	res += time.Duration(d.Month*30) * 24 * time.Hour
	res += time.Duration(d.Year*365) * 24 * time.Hour
	return res, true
}

// Truncate truncates a timestamp to the duration. It keeps the timezone of the timestamp.
// It only partially supports durations with multiple components (e.g. P1Y3M): it will truncate by the first component only.
// TODO: Is there a well-defined way to truncate by multiple components?
// TODO: Can this be merged with timeutil.TruncateTime? timeutil returns UTC values, whereas this returns in the same tz as the input time.
func (d StandardDuration) Truncate(t time.Time, firstDayOfWeek, firstMonthOfYear int) time.Time {
	if d.Year != 0 {
		n := t.Year()
		if int(t.Month()) < firstMonthOfYear {
			n--
		}
		n -= n % d.Year
		return time.Date(n, time.Month(firstMonthOfYear), 1, 0, 0, 0, 0, t.Location())
	}
	if d.Month != 0 {
		n := int(t.Month()) - 1 // Month is 1-indexed
		n -= n % d.Month
		n++ // Correct back for 1-indexing
		return time.Date(t.Year(), time.Month(n), 1, 0, 0, 0, 0, t.Location())
	}
	if d.Week != 0 {
		if firstDayOfWeek < 1 {
			firstDayOfWeek = 1
		}
		if firstDayOfWeek > 7 {
			firstDayOfWeek = 7
		}

		weekday := int(t.Weekday())
		if weekday == 0 { // We treat Sunday as 7, not 0
			weekday = 7
		}

		daysToSubstract := weekday - firstDayOfWeek
		if daysToSubstract < 0 {
			daysToSubstract = 7 + daysToSubstract
		}

		_, weeksToSubtract := t.AddDate(0, 0, -daysToSubstract).ISOWeek()
		weeksToSubtract-- // ISOWeek is 1-indexed
		weeksToSubtract %= d.Week

		daysToSubstract += weeksToSubtract * 7

		return time.Date(t.Year(), t.Month(), t.Day()-daysToSubstract, 0, 0, 0, 0, t.Location())
	}
	if d.Day != 0 {
		n := t.Day() - 1 // Day is 1-indexed
		n -= n % d.Day
		n++ // Correct back for 1-indexing
		return time.Date(t.Year(), t.Month(), n, 0, 0, 0, 0, t.Location())
	}
	if d.Hour != 0 {
		n := t.Hour() % d.Hour
		return t.Truncate(time.Hour).Add(-time.Duration(n) * time.Hour)
	}
	if d.Minute != 0 {
		n := t.Minute() % d.Minute
		return t.Truncate(time.Minute).Add(-time.Duration(n) * time.Minute)
	}
	if d.Second != 0 {
		n := t.Second() % d.Second
		return t.Truncate(time.Second).Add(-time.Duration(n) * time.Second)
	}
	return t
}

func (d StandardDuration) EndTime(t time.Time) time.Time {
	if d.rillExtension == "" {
		return t
	}
	switch d.rillExtension {
	case "PD":
		return d.Sub(t).AddDate(0, 0, 1)
	case "PW":
		return d.Sub(t).AddDate(0, 0, 7)
	case "PM":
		return d.Sub(t).AddDate(0, 1, 0)
	case "PQ":
		return d.Sub(t).AddDate(0, 3, 0)
	case "PY":
		return d.Sub(t).AddDate(1, 0, 0)
	default:
		return t
	}
}

type InfDuration struct{}

func (d InfDuration) Add(t time.Time) time.Time {
	return time.Time{}
}

func (d InfDuration) Sub(t time.Time) time.Time {
	return time.Time{}
}

func (d InfDuration) EstimateNative() (time.Duration, bool) {
	return 0, false
}

type TruncToDateDuration struct {
	anchor timeutil.TimeGrain
}

func (d TruncToDateDuration) Add(t time.Time) time.Time {
	return time.Time{}
}

func (d TruncToDateDuration) Sub(t time.Time) time.Time {
	return timeutil.TruncateTime(t, d.anchor, t.Location(), 1, 1) // TODO: get first day and month
}

func (d TruncToDateDuration) SubWithUnit(t time.Time, unit int) time.Time {
	if unit <= 0 {
		return t
	}
	for i := 1; i <= unit; i++ {
		t = timeutil.TruncateTime(t, d.anchor, t.Location(), 1, 1) // TODO: get first day and month
		t = t.AddDate(0, 0, -1)
	}
	return t
}

func (d TruncToDateDuration) EstimateNative() (time.Duration, bool) {
	return 0, false
}
