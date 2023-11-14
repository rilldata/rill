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
		"PD": {Day: 1},
		"PW": {Week: 1},
		"PM": {Month: 1},
		"PQ": {Month: 3},
		"PY": {Year: 1},
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
	// Date component
	Year  int
	Month int
	Week  int
	Day   int
	// Time Component
	Hour   int
	Minute int
	Second int
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

type InfDuration struct{}

func (d InfDuration) Add(t time.Time) time.Time {
	return time.Time{}
}

func (d InfDuration) Sub(t time.Time) time.Time {
	return time.Time{}
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
