package clickhouse

import (
	"strconv"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// ParseIntervalToMillis parses a ClickHouse INTERVAL string into milliseconds.
// ClickHouse currently returns INTERVALs as strings in the format "1 Month", "2 Minutes", etc.
// This function follows our current policy of treating months as 30 days when converting to milliseconds.
func ParseIntervalToMillis(s string) (int, bool) {
	s1, s2, ok := strings.Cut(s, " ")
	if !ok {
		return 0, false
	}

	units, err := strconv.Atoi(s1)
	if err != nil {
		return 0, false
	}

	var multiplier int
	switch s2 {
	case "Second", "Seconds":
		multiplier = 1000
	case "Minute", "Minutes":
		multiplier = 60 * 1000
	case "Hour", "Hours":
		multiplier = 60 * 60 * 1000
	case "Day", "Days":
		multiplier = 24 * 60 * 60 * 1000
	case "Month", "Months":
		multiplier = 30 * 24 * 60 * 60 * 1000
	case "Year", "Years":
		multiplier = 365 * 24 * 60 * 60 * 1000
	default:
		return 0, false
	}

	return units * multiplier, true
}

func safeSQLName(name string) string {
	return drivers.DialectClickHouse.EscapeIdentifier(name)
}
