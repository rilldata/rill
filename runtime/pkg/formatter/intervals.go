package formatter

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	xnumber "golang.org/x/text/number"
)

// Constants for time calculations
const (
	msPerSec   = 1000
	msPerMin   = 60 * msPerSec
	msPerHour  = 60 * msPerMin
	msPerDay   = 24 * msPerHour
	msPerMonth = 30 * msPerDay
	msPerYear  = 365 * msPerDay
)

type timeunit string

const (
	tuMs    timeunit = "ms"
	tuSec   timeunit = "s"
	tuMin   timeunit = "m"
	tuHour  timeunit = "h"
	tuDay   timeunit = "d"
	tuMonth timeunit = "mon"
	tuYear  timeunit = "y"
)

type intervalFormatter struct{}

func newIntervalFormatter() *intervalFormatter {
	return &intervalFormatter{}
}

func (f *intervalFormatter) StringFormat(x any) (string, error) {
	if v, ok := asNumber[float64](x); ok {
		return formatMsInterval(v), nil
	}
	return "", fmt.Errorf("not a number: %v", x)
}

// formatMsInterval formats a millisecond value into a compact human-readable time interval.
// The strategy is to:
// - show two digits of precision
// - prefer to show two integer digits in a smaller unit
// - if that is not possible, show a floating point number in a larger unit with one digit of precision (e.g. 1.2 days)
// see https://www.notion.so/rilldata/Support-display-of-intervals-and-formatting-of-intervals-e-g-25-days-in-dashboardsal-data-t-8720522eded648f58f35421ebc28ee2f
func formatMsInterval(ms float64) string {
	var neg string
	if ms < 0 {
		ms = -ms
		neg = "-"
	}

	// Helper function to format number based on unit and check digits
	format := func(value float64, tu timeunit) string {
		p := message.NewPrinter(language.English)
		var formatter xnumber.Formatter
		// https://github.com/golang/go/issues/31254
		if value < 1 {
			formatter = xnumber.Decimal(
				value,
				xnumber.MaxFractionDigits(2),
				xnumber.MinFractionDigits(1),
			)
		} else if value < 10 {
			formatter = xnumber.Decimal(
				value,
				xnumber.MaxFractionDigits(1),
				xnumber.MinFractionDigits(0),
				xnumber.Precision(2),
			)
		} else {
			formatter = xnumber.Decimal(
				value,
				xnumber.Scale(0),
			)
		}
		valueStr := p.Sprint(formatter)
		return fmt.Sprintf("%s%s %s", neg, valueStr, tu)
	}

	switch {
	case ms == 0:
		return fmt.Sprintf("0 %s", tuSec)
	case ms < 1:
		return fmt.Sprintf("~0 %s", tuSec)
	case ms < 100:
		return format(ms, "ms")
	case ms < 90*msPerSec:
		return format(ms/msPerSec, "s")
	case ms < 90*msPerMin:
		return format(ms/msPerMin, "m")
	case ms < 72*msPerHour:
		return format(ms/msPerHour, "h")
	case ms < 90*msPerDay:
		return format(ms/msPerDay, "d")
	case ms < 18*msPerMonth:
		return format(ms/msPerMonth, "mon")
	case ms < 100*msPerYear:
		return format(ms/msPerYear, "y")
	default:
		if neg == "-" {
			return fmt.Sprintf("< -100 %s", tuYear)
		}
		return fmt.Sprintf(">100 %s", tuYear)
	}
}

type intervalExpFormatter struct{}

func newIntervalExpFormatter() *intervalExpFormatter {
	return &intervalExpFormatter{}
}

func (f *intervalExpFormatter) StringFormat(x any) (string, error) {
	if v, ok := asNumber[float64](x); ok {
		return formatMsToDuckDbIntervalString(v, "short"), nil
	}
	return "", fmt.Errorf("not a number: %v", x)
}

// formatMsToDuckDbIntervalString formats a millisecond value into an expanded interval string
// that will be parsable by a duckdb INTERVAL constructor.
// The hour+min+sec portion will use whichever is shorter between the `HH:MM:SS.xxx`
// format and a sparse format like `2h 4s` for the HMS part of the interval.
func formatMsToDuckDbIntervalString(ms float64, style string) string {
	neg := ""
	if ms < 0 {
		ms = -ms
		neg = "-"
	}

	if ms == 0 {
		return fmt.Sprintf("0%s", tuSec)
	}

	if ms < 1 {
		return fmt.Sprintf("~0%s", tuSec)
	}

	var result string

	intMs := int64(ms)
	years := intMs / msPerYear
	months := (intMs % msPerYear) / msPerMonth
	days := (intMs % msPerMonth) / msPerDay
	hours := (intMs % msPerDay) / msPerHour
	minutes := (intMs % msPerHour) / msPerMin
	seconds := (intMs % msPerMin) / msPerSec
	milliseconds := intMs % msPerSec

	timeParts := []struct {
		value int64
		unit  timeunit
	}{
		{years, tuYear},
		{months, tuMonth},
		{days, tuDay},
	}

	for _, part := range timeParts {
		if part.value > 0 {
			result += fmt.Sprintf("%s%d%s ", neg, part.value, part.unit)
		}
	}

	if hours == 0 && minutes == 0 && seconds == 0 && milliseconds == 0 {
		return strings.TrimSpace(result)
	}

	switch style {
	case "units":
		return result + formatUnitsHMS(hours, minutes, seconds, milliseconds, neg)
	case "colon":
		return result + formatColonHMS(hours, minutes, seconds, neg)
	default:
		return result + formatShortHMS(hours, minutes, seconds, milliseconds, neg)
	}
}

func formatUnitsHMS(h, m, s, ms int64, neg string) string {
	parts := []struct {
		value int64
		unit  timeunit
	}{
		{h, tuHour},
		{m, tuMin},
		{s, tuSec},
		{ms, tuMs},
	}

	var result string
	for _, part := range parts {
		if part.value > 0 {
			result += fmt.Sprintf("%s%d%s ", neg, part.value, part.unit)
		}
	}
	return strings.TrimSpace(result)
}

func formatColonHMS(h, m, s int64, neg string) string {
	return fmt.Sprintf("%s%02d:%02d:%02d", neg, h, m, s)
}

func formatShortHMS(h, m, s, ms int64, neg string) string {
	secondsWithMs := float64(s) + float64(ms)/1000.0
	colonFormatted := formatColonHMS(h, m, int64(secondsWithMs), neg)
	unitsFormatted := formatUnitsHMS(h, m, s, ms, neg)
	if len(colonFormatted) < len(unitsFormatted) {
		return colonFormatted
	}
	return unitsFormatted
}
