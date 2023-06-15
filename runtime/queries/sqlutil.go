package queries

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

var ErrExportNotSupported = fmt.Errorf("exporting is not supported")

func quoteName(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

func escapeDoubleQuotes(column string) string {
	return strings.ReplaceAll(column, "\"", "\"\"")
}

func safeName(name string) string {
	if name == "" {
		return name
	}
	return quoteName(escapeDoubleQuotes(name))
}

func tempName(prefix string) string {
	return prefix + strings.ReplaceAll(uuid.New().String(), "-", "")
}

func convertToDateTruncSpecifier(specifier runtimev1.TimeGrain) string {
	switch specifier {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "MILLISECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "SECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "MINUTE"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "HOUR"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "DAY"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "WEEK"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "MONTH"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "QUARTER"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "YEAR"
	}
	panic(fmt.Errorf("unconvertable time grain specifier: %v", specifier))
}

func toTimeGrain(val string) runtimev1.TimeGrain {
	switch strings.ToUpper(val) {
	case "MILLISECOND":
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND
	case "SECOND":
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND
	case "MINUTE":
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE
	case "HOUR":
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR
	case "DAY":
		return runtimev1.TimeGrain_TIME_GRAIN_DAY
	case "WEEK":
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK
	case "MONTH":
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH
	case "QUARTER":
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER
	case "YEAR":
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR
	default:
		panic(fmt.Errorf("unconvertable time grain specifier: %v", val))
	}
}

func addInterval(t time.Time, timeGrain runtimev1.TimeGrain) time.Time {
	switch timeGrain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		t = t.Truncate(time.Millisecond)
		return t.Add(time.Millisecond)
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		t = t.Truncate(time.Second)
		return t.Add(time.Second)
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		t = t.Truncate(time.Minute)
		return t.Add(time.Minute)
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		t = t.Truncate(time.Hour)
		return t.Add(time.Hour)
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		t = t.Truncate(time.Hour * 24)
		return t.Add(time.Hour * 24)
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		t = t.Truncate(time.Hour * 24 * 7)
		return t.Add(time.Hour * 24 * 7)
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		return t.AddDate(0, 1, 0)
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		return t.AddDate(0, 1, 0)
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		t = time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
		return t.AddDate(1, 0, 0)
	default:
		return t
	}
}
