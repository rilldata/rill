// Package duration copied as it is from github.com/senseyeio/duration
package duration

import (
	"fmt"
	"regexp"
	"strconv"
)

// Duration represents an ISO8601 duration with Rill-specific extensions.
// See ParseISO8601 for details.
type Duration struct {
	// If Inf is true, the other components should be ignored
	Inf bool
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

// Regexes used by ParseISO8601
var (
	infPattern      = regexp.MustCompile("^(?i)inf$")
	durationPattern = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
)

// ParseISO8601 parses an ISO8601 duration as well as some Rill-specific extensions.
// (Section 3.7 of the standard supposedly allows extensions that do not interfere with the standard.)
// The only current extension is "inf", for representing an unbounded duration of time.
func ParseISO8601(from string) (Duration, error) {
	// Try parsing for "inf"
	if infPattern.MatchString(from) {
		return Duration{Inf: true}, nil
	}

	// Parse as a regular ISO8601 duration
	if !durationPattern.MatchString(from) {
		return Duration{}, fmt.Errorf("string %q is not a valid ISO 8601 duration", from)
	}

	var d Duration
	match := durationPattern.FindStringSubmatch(from)
	for i, name := range durationPattern.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return Duration{}, err
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
