package sqldialect

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// DruidTimeFloorSpecifier converts a TimeGrain to a Druid time_floor period specifier.
func DruidTimeFloorSpecifier(grain runtimev1.TimeGrain) string {
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "PT0.001S"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "PT1S"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "PT1M"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "PT1H"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "P1D"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "P1W"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "P1M"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "P3M"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "P1Y"
	}
	panic(fmt.Errorf("invalid time grain enum value %d", int(grain)))
}

// TempName generates a unique temporary name with the given prefix.
func TempName(prefix string) string {
	return prefix + strings.ReplaceAll(uuid.New().String(), "-", "")
}

// CheckTypeCompatibility reports whether a struct field's type can be used in SelectInlineResults.
func CheckTypeCompatibility(f *runtimev1.StructType_Field) bool {
	switch f.Type.Code {
	case runtimev1.Type_CODE_STRING,
		runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64,
		runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64,
		runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64,
		runtimev1.Type_CODE_BOOL,
		runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
		return true
	default:
		return false
	}
}
