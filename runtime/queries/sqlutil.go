package queries

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

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
	case "YEAR":
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR
	default:
		panic(fmt.Errorf("unconvertable time grain specifier: %v", val))
	}
}
