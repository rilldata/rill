package clickhouse

import (
	"strconv"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// ParseIntervalToMillis parses a ClickHouse INTERVAL string into milliseconds.
// ClickHouse currently returns INTERVALs as strings in the format "1 Month", "2 Minutes", etc.
// This function follows our current policy of treating months as 30 days when converting to milliseconds.
func ParseIntervalToMillis(s string) (int64, bool) {
	s1, s2, ok := strings.Cut(s, " ")
	if !ok {
		return 0, false
	}

	units, err := strconv.ParseInt(s1, 10, 64)
	if err != nil {
		return 0, false
	}

	switch s2 {
	case "Nanosecond", "Nanoseconds":
		return int64(float64(units) / 1_000_000), true
	case "Microsecond", "Microseconds":
		return int64(float64(units) / 1_000), true
	case "Millisecond", "Milliseconds":
		return units * 1, true
	case "Second", "Seconds":
		return units * 1000, true
	case "Minute", "Minutes":
		return units * 60 * 1000, true
	case "Hour", "Hours":
		return units * 60 * 60 * 1000, true
	case "Day", "Days":
		return units * 24 * 60 * 60 * 1000, true
	case "Month", "Months":
		return units * 30 * 24 * 60 * 60 * 1000, true
	case "Quarter", "Quarters":
		return units * 3 * 30 * 24 * 60 * 60 * 1000, true
	case "Year", "Years":
		return units * 365 * 24 * 60 * 60 * 1000, true
	default:
		return 0, false
	}
}

func safeSQLName(name string) string {
	return drivers.DialectClickHouse.EscapeIdentifier(name)
}
