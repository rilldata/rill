package metricsview

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/jsonschemautil"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

type Query struct {
	MetricsView         string      `mapstructure:"metrics_view"`
	Dimensions          []Dimension `mapstructure:"dimensions"`
	Measures            []Measure   `mapstructure:"measures"`
	PivotOn             []string    `mapstructure:"pivot_on"`
	Spine               *Spine      `mapstructure:"spine"`
	Sort                []Sort      `mapstructure:"sort"`
	TimeRange           *TimeRange  `mapstructure:"time_range"`
	ComparisonTimeRange *TimeRange  `mapstructure:"comparison_time_range"`
	Where               *Expression `mapstructure:"where"`
	Having              *Expression `mapstructure:"having"`
	Limit               *int64      `mapstructure:"limit"`
	Offset              *int64      `mapstructure:"offset"`
	TimeZone            string      `mapstructure:"time_zone"`
	UseDisplayNames     bool        `mapstructure:"use_display_names"`
	Rows                bool        `mapstructure:"rows"`
}

type Dimension struct {
	Name    string            `mapstructure:"name"`
	Compute *DimensionCompute `mapstructure:"compute"`
}

type DimensionCompute struct {
	TimeFloor *DimensionComputeTimeFloor `mapstructure:"time_floor"`
}

type DimensionComputeTimeFloor struct {
	Dimension string    `mapstructure:"dimension"`
	Grain     TimeGrain `mapstructure:"grain"`
}

type Measure struct {
	Name    string          `mapstructure:"name"`
	Compute *MeasureCompute `mapstructure:"compute"`
}

type MeasureCompute struct {
	Count           bool                           `mapstructure:"count"`
	CountDistinct   *MeasureComputeCountDistinct   `mapstructure:"count_distinct"`
	ComparisonValue *MeasureComputeComparisonValue `mapstructure:"comparison_value"`
	ComparisonDelta *MeasureComputeComparisonDelta `mapstructure:"comparison_delta"`
	ComparisonRatio *MeasureComputeComparisonRatio `mapstructure:"comparison_ratio"`
	PercentOfTotal  *MeasureComputePercentOfTotal  `mapstructure:"percent_of_total"`
	URI             *MeasureComputeURI             `mapstructure:"uri"`
	ComparisonTime  *MeasureComputeComparisonTime  `mapstructure:"comparison_time"`
}

func (q *Query) AsMap() (map[string]any, error) {
	queryMap := make(map[string]any)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &queryMap,
		DecodeHook: timeDecodeFunc,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(q)
	if err != nil {
		return nil, err
	}
	return queryMap, nil
}

func (q *Query) Validate() error {
	if q.Rows {
		if len(q.Dimensions) > 0 {
			return fmt.Errorf("dimensions not supported when rows is set, all model columns will be returned")
		}
		if len(q.Measures) > 0 {
			return fmt.Errorf("measures not supported when rows is set, all model columns will be returned")
		}
		if len(q.Sort) > 0 {
			return fmt.Errorf("sort not supported when rows is set")
		}
		if q.ComparisonTimeRange != nil {
			return fmt.Errorf("comparison_time_range not supported when rows is set")
		}
		if q.Having != nil {
			return fmt.Errorf("having not supported when rows is set")
		}
		if len(q.PivotOn) > 0 {
			return fmt.Errorf("pivot_on not supported when rows is set")
		}
	}

	if q.TimeRange != nil && q.ComparisonTimeRange != nil && q.TimeRange.TimeDimension != q.ComparisonTimeRange.TimeDimension {
		return fmt.Errorf("time_dimension in time_range and comparison_time_range must match")
	}

	return nil
}

func (m *MeasureCompute) Validate() error {
	n := 0
	if m.Count {
		n++
	}
	if m.CountDistinct != nil {
		n++
	}
	if m.ComparisonValue != nil {
		n++
	}
	if m.ComparisonDelta != nil {
		n++
	}
	if m.ComparisonRatio != nil {
		n++
	}
	if m.PercentOfTotal != nil {
		n++
	}
	if m.URI != nil {
		n++
	}
	if m.ComparisonTime != nil {
		n++
	}
	if n == 0 {
		return fmt.Errorf(`must specify a compute operation`)
	}
	if n > 1 {
		return fmt.Errorf("must specify only one compute operation")
	}
	return nil
}

type MeasureComputeCountDistinct struct {
	Dimension string `mapstructure:"dimension"`
}

type MeasureComputeComparisonValue struct {
	Measure string `mapstructure:"measure"`
}

type MeasureComputeComparisonDelta struct {
	Measure string `mapstructure:"measure"`
}

type MeasureComputeComparisonRatio struct {
	Measure string `mapstructure:"measure"`
}

type MeasureComputePercentOfTotal struct {
	Measure string   `mapstructure:"measure"`
	Total   *float64 `mapstructure:"total"`
}

type MeasureComputeURI struct {
	Dimension string `mapstructure:"dimension"`
}

type MeasureComputeComparisonTime struct {
	Dimension string `mapstructure:"dimension"`
}

type Spine struct {
	Where     *WhereSpine `mapstructure:"where"`
	TimeRange *TimeSpine  `mapstructure:"time"`
}

type WhereSpine struct {
	Expression *Expression `mapstructure:"expr"`
}

type TimeSpine struct {
	Start         time.Time `mapstructure:"start"`
	End           time.Time `mapstructure:"end"`
	Grain         TimeGrain `mapstructure:"grain"`
	TimeDimension string    `mapstructure:"time_dimension"` // optional time dimension to use for time-based operations, if not specified, the default time dimension in the metrics view is used
}

type Sort struct {
	Name string `mapstructure:"name"`
	Desc bool   `mapstructure:"desc"`
}

type TimeRange struct {
	Start         time.Time `mapstructure:"start"`
	End           time.Time `mapstructure:"end"`
	Expression    string    `mapstructure:"expression"`
	IsoDuration   string    `mapstructure:"iso_duration"`
	IsoOffset     string    `mapstructure:"iso_offset"`
	RoundToGrain  TimeGrain `mapstructure:"round_to_grain"`
	TimeDimension string    `mapstructure:"time_dimension"` // optional time dimension to use for time-based operations, if not specified, the default time dimension in the metrics view is used
}

func (tr *TimeRange) IsZero() bool {
	return tr.Start.IsZero() && tr.End.IsZero() && tr.Expression == "" && tr.IsoDuration == "" && tr.IsoOffset == "" && tr.RoundToGrain == TimeGrainUnspecified
}

type Expression struct {
	Name      string     `mapstructure:"name"`
	Value     any        `mapstructure:"val"`
	Condition *Condition `mapstructure:"cond"`
	Subquery  *Subquery  `mapstructure:"subquery"`
}

type Condition struct {
	Operator    Operator      `mapstructure:"op"`
	Expressions []*Expression `mapstructure:"exprs"`
}

type Subquery struct {
	Dimension Dimension   `mapstructure:"dimension"`
	Measures  []Measure   `mapstructure:"measures"`
	Where     *Expression `mapstructure:"where"`
	Having    *Expression `mapstructure:"having"`
}

type Operator string

const (
	OperatorUnspecified Operator = ""
	OperatorEq          Operator = "eq"
	OperatorNeq         Operator = "neq"
	OperatorLt          Operator = "lt"
	OperatorLte         Operator = "lte"
	OperatorGt          Operator = "gt"
	OperatorGte         Operator = "gte"
	OperatorIn          Operator = "in"
	OperatorNin         Operator = "nin"
	OperatorIlike       Operator = "ilike"
	OperatorNilike      Operator = "nilike"
	OperatorOr          Operator = "or"
	OperatorAnd         Operator = "and"
)

func (o Operator) Valid() bool {
	switch o {
	case OperatorEq, OperatorNeq, OperatorLt, OperatorLte, OperatorGt, OperatorGte, OperatorIn, OperatorNin, OperatorIlike, OperatorNilike, OperatorOr, OperatorAnd:
		return true
	}
	return false
}

type TimeGrain string

const (
	TimeGrainUnspecified TimeGrain = ""
	TimeGrainMillisecond TimeGrain = "millisecond"
	TimeGrainSecond      TimeGrain = "second"
	TimeGrainMinute      TimeGrain = "minute"
	TimeGrainHour        TimeGrain = "hour"
	TimeGrainDay         TimeGrain = "day"
	TimeGrainWeek        TimeGrain = "week"
	TimeGrainMonth       TimeGrain = "month"
	TimeGrainQuarter     TimeGrain = "quarter"
	TimeGrainYear        TimeGrain = "year"
)

func (t TimeGrain) Valid() bool {
	switch t {
	case TimeGrainUnspecified, TimeGrainMillisecond, TimeGrainSecond, TimeGrainMinute, TimeGrainHour, TimeGrainDay, TimeGrainWeek, TimeGrainMonth, TimeGrainQuarter, TimeGrainYear:
		return true
	}
	return false
}

func (t TimeGrain) ToTimeutil() timeutil.TimeGrain {
	switch t {
	case TimeGrainUnspecified:
		return timeutil.TimeGrainUnspecified
	case TimeGrainMillisecond:
		return timeutil.TimeGrainMillisecond
	case TimeGrainSecond:
		return timeutil.TimeGrainSecond
	case TimeGrainMinute:
		return timeutil.TimeGrainMinute
	case TimeGrainHour:
		return timeutil.TimeGrainHour
	case TimeGrainDay:
		return timeutil.TimeGrainDay
	case TimeGrainWeek:
		return timeutil.TimeGrainWeek
	case TimeGrainMonth:
		return timeutil.TimeGrainMonth
	case TimeGrainQuarter:
		return timeutil.TimeGrainQuarter
	case TimeGrainYear:
		return timeutil.TimeGrainYear
	default:
		panic(fmt.Errorf("invalid time grain %q", t))
	}
}

func (t TimeGrain) ToProto() runtimev1.TimeGrain {
	switch t {
	case TimeGrainUnspecified:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
	case TimeGrainMillisecond:
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND
	case TimeGrainSecond:
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND
	case TimeGrainMinute:
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE
	case TimeGrainHour:
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR
	case TimeGrainDay:
		return runtimev1.TimeGrain_TIME_GRAIN_DAY
	case TimeGrainWeek:
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK
	case TimeGrainMonth:
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH
	case TimeGrainQuarter:
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER
	case TimeGrainYear:
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR
	default:
		panic(fmt.Errorf("invalid time grain %q", t))
	}
}

func TimeGrainFromProto(t runtimev1.TimeGrain) TimeGrain {
	switch t {
	case runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED:
		return TimeGrainUnspecified
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return TimeGrainMillisecond
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return TimeGrainSecond
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return TimeGrainMinute
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return TimeGrainHour
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return TimeGrainDay
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return TimeGrainWeek
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return TimeGrainMonth
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return TimeGrainQuarter
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return TimeGrainYear
	default:
		panic(fmt.Errorf("invalid time grain %q", t))
	}
}

// AnalyzeQueryFields returns a list of all fields (dimensions and measures) that are part of the query.
func AnalyzeQueryFields(q *Query) []string {
	// Extract accessible fields from the query
	fieldsMap := make(map[string]struct{})
	// Add dimensions
	for _, dim := range q.Dimensions {
		fieldsMap[getDimensionName(dim)] = struct{}{}
	}
	// Add measures
	for _, meas := range q.Measures {
		fieldsMap[getMeasureName(meas)] = struct{}{}
	}
	// Add time dimension if present
	if q.TimeRange != nil && q.TimeRange.TimeDimension != "" {
		fieldsMap[q.TimeRange.TimeDimension] = struct{}{}
	}

	exprFields := AnalyzeExpressionFields(q.Where)
	for _, f := range exprFields {
		fieldsMap[f] = struct{}{}
	}

	var fields []string
	for f := range fieldsMap {
		fields = append(fields, f)
	}

	return fields
}

func getDimensionName(dim Dimension) string {
	if dim.Compute == nil {
		return dim.Name
	}

	if dim.Compute.TimeFloor != nil {
		return dim.Compute.TimeFloor.Dimension
	}

	panic("could not find dimension name")
}

func getMeasureName(m Measure) string {
	if m.Compute == nil {
		return m.Name
	}
	switch {
	case m.Compute.Count:
		return "" // skip
	case m.Compute.CountDistinct != nil:
		return m.Compute.CountDistinct.Dimension
	case m.Compute.ComparisonValue != nil: // although comparison cases can be skipped as base fields would have already been added but adding for switch completeness as it will deduped
		return m.Compute.ComparisonValue.Measure
	case m.Compute.ComparisonDelta != nil:
		return m.Compute.ComparisonDelta.Measure
	case m.Compute.ComparisonRatio != nil:
		return m.Compute.ComparisonRatio.Measure
	case m.Compute.PercentOfTotal != nil:
		return m.Compute.PercentOfTotal.Measure
	case m.Compute.URI != nil:
		return m.Compute.URI.Dimension
	case m.Compute.ComparisonTime != nil:
		return m.Compute.ComparisonTime.Dimension
	default:
		panic("could not find measure name")
	}
}

var timeDecodeFunc mapstructure.DecodeHookFunc = func(from reflect.Type, to reflect.Type, data any) (any, error) {
	if from == reflect.TypeOf(&time.Time{}) {
		t, ok := data.(*time.Time)
		if !ok {
			return nil, fmt.Errorf("expected *time.Time, got %T", data)
		}
		return map[string]any{
			"t": t.Format(time.RFC3339Nano),
		}, nil
	}
	return data, nil
}

const QueryJSONSchema = `
{
  "type": "object",
  "properties": {
    "metrics_view": {
      "type": "string",
      "description": "The metrics view to query."
    },
    "dimensions": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/Dimension"
      },
      "description": "List of dimensions to include in the query. The result will be grouped by these."
    },
    "measures": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/Measure"
      },
      "description": "List of measures to include in the query. These will be aggregated based on the dimensions."
    },
    "pivot_on": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Optional dimensions to pivot on. The provided dimensions must be present in the query. If not provided, the query will return a flat result set. Note that pivoting can have poor performance on large result sets."
    },
    "spine": {
      "$ref": "#/$defs/Spine",
      "description": "Optionally configure a 'spine' of dimension values that should be present in the result regardless of whether they have data. This is for example useful for generating a time series with zero values for missing dates."
    },
    "sort": {
      "type": "array",
      "items": {
        "$ref": "#/$defs/Sort"
      },
      "description": "Sort order for the results."
    },
    "time_range": {
      "$ref": "#/$defs/TimeRange",
      "description": "Time range filter for the query. Time ranges are inclusive of start time and exclusive of end time. Note that for large datasets, querying shorter and/or more recent time ranges has significant performance benefits."
    },
    "comparison_time_range": {
      "$ref": "#/$defs/TimeRange",
      "description": "Time range filter to use for comparison measures."
    },
    "where": {
      "$ref": "#/$defs/Expression",
      "description": "Optional expression for filtering the underlying data before aggregation. This is the recommended way to filter data."
    },
    "having": {
      "$ref": "#/$defs/Expression",
      "description": "Optional expression for filtering the results after aggregation. This is useful for filtering based on the aggregated measure values."
    },
    "limit": {
      "type": "integer",
      "minimum": 0,
      "description": "Maximum number of rows to return. It is required for interactive queries."
    },
    "offset": {
      "type": "integer",
      "minimum": 0,
      "description": "Optional offset for the query results. This is useful for pagination together with 'limit'."
    },
    "time_zone": {
      "type": "string",
      "description": "Optional time zone for time_floor operations and dynamic time ranges. Defaults to UTC."
    },
    "use_display_names": {
      "type": "boolean",
      "description": "Optional flag to return results using display names for dimensions and measures instead of their unique names. Defaults to false."
    },
    "rows": {
      "type": "boolean",
      "description": "Optional flag to return the underlying rows instead of aggregated results. This is useful for debugging or exploring the data. Setting it to true is incompatible with the following options: dimensions, measures, sort, comparison_time_range, having, pivot_on."
    }
  },
  "$defs": {
    "Dimension": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Name of the dimension"
        },
        "compute": {
          "$ref": "#/$defs/DimensionCompute",
          "description": "Optionally configure a derived dimension, such as a time floor."
        }
      },
      "required": ["name"]
    },
    "DimensionCompute": {
      "type": "object",
      "properties": {
        "time_floor": {
          "$ref": "#/$defs/DimensionComputeTimeFloor"
        }
      }
    },
    "DimensionComputeTimeFloor": {
      "type": "object",
      "properties": {
        "dimension": {
          "type": "string",
          "description": "Dimension to apply time floor to"
        },
        "grain": {
          "$ref": "#/$defs/TimeGrain",
          "description": "Time grain for flooring"
        }
      },
      "required": ["dimension", "grain"]
    },
    "Measure": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Name of the measure"
        },
        "compute": {
          "$ref": "#/$defs/MeasureCompute",
          "description": "Optionally configure a derived measure, such as a comparison."
        }
      },
      "required": ["name"]
    },
    "MeasureCompute": {
      "type": "object",
      "properties": {
        "count": {
          "type": "boolean",
          "description": "Whether to compute count"
        },
        "count_distinct": {
          "$ref": "#/$defs/MeasureComputeCountDistinct"
        },
        "comparison_value": {
          "$ref": "#/$defs/MeasureComputeComparisonValue"
        },
        "comparison_delta": {
          "$ref": "#/$defs/MeasureComputeComparisonDelta"
        },
        "comparison_ratio": {
          "$ref": "#/$defs/MeasureComputeComparisonRatio"
        },
        "percent_of_total": {
          "$ref": "#/$defs/MeasureComputePercentOfTotal"
        },
        "uri": {
          "$ref": "#/$defs/MeasureComputeURI"
        }
      },
      "oneOf": [
        {"required": ["count"]},
        {"required": ["count_distinct"]},
        {"required": ["comparison_value"]},
        {"required": ["comparison_delta"]},
        {"required": ["comparison_ratio"]},
        {"required": ["percent_of_total"]},
        {"required": ["uri"]}
      ]
    },
    "MeasureComputeCountDistinct": {
      "type": "object",
      "properties": {
        "dimension": {
          "type": "string",
          "description": "Dimension to count distinct values for"
        }
      },
      "required": ["dimension"]
    },
    "MeasureComputeComparisonValue": {
      "type": "object",
      "properties": {
        "measure": {
          "type": "string",
          "description": "Measure to compare"
        }
      },
      "required": ["measure"]
    },
    "MeasureComputeComparisonDelta": {
      "type": "object",
      "properties": {
        "measure": {
          "type": "string",
          "description": "Measure to compute delta for"
        }
      },
      "required": ["measure"]
    },
    "MeasureComputeComparisonRatio": {
      "type": "object",
      "properties": {
        "measure": {
          "type": "string",
          "description": "Measure to compute ratio for"
        }
      },
      "required": ["measure"]
    },
    "MeasureComputePercentOfTotal": {
      "type": "object",
      "properties": {
        "measure": {
          "type": "string",
          "description": "Measure to compute percentage for"
        },
        "total": {
          "type": "number",
          "description": "Total value to use for percentage calculation"
        }
      },
      "required": ["measure"]
    },
    "MeasureComputeURI": {
      "type": "object",
      "properties": {
        "dimension": {
          "type": "string",
          "description": "Dimension to generate URI for"
        }
      },
      "required": ["dimension"]
    },
    "Spine": {
      "type": "object",
      "properties": {
        "where": {
          "$ref": "#/$defs/WhereSpine"
        },
        "time": {
          "$ref": "#/$defs/TimeSpine"
        }
      }
    },
    "WhereSpine": {
      "type": "object",
      "properties": {
        "expr": {
          "$ref": "#/$defs/Expression"
        }
      }
    },
    "TimeSpine": {
      "type": "object",
      "properties": {
        "start": {
          "type": "string",
          "format": "date-time",
          "description": "Start time"
        },
        "end": {
          "type": "string",
          "format": "date-time",
          "description": "End time"
        },
        "grain": {
          "$ref": "#/$defs/TimeGrain",
          "description": "Time grain for the spine"
        }
      },
      "required": ["start", "end", "grain"]
    },
    "Sort": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Field name to sort by"
        },
        "desc": {
          "type": "boolean",
          "description": "Whether to sort in descending order"
        }
      },
      "required": ["name"]
    },
    "TimeRange": {
      "type": "object",
      "properties": {
        "start": {
          "type": "string",
          "format": "date-time",
          "description": "Start time (inclusive)"
        },
        "end": {
          "type": "string",
          "format": "date-time",
          "description": "End time (exclusive)"
        },
        "expression": {
          "type": "string",
          "description": "Time range expression"
        },
        "iso_duration": {
          "type": "string",
          "description": "ISO 8601 duration"
        },
        "iso_offset": {
          "type": "string",
          "description": "ISO 8601 offset"
        },
        "round_to_grain": {
          "$ref": "#/$defs/TimeGrain",
          "description": "Time grain to round to"
        }
      }
    },
    "Expression": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Expression name"
        },
        "val": {
          "description": "Expression value"
        },
        "cond": {
          "$ref": "#/$defs/Condition"
        },
        "subquery": {
          "$ref": "#/$defs/Subquery"
        }
      }
    },
    "Condition": {
      "type": "object",
      "properties": {
        "op": {
          "$ref": "#/$defs/Operator",
          "description": "Operator for the condition"
        },
        "exprs": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/Expression"
          },
          "description": "Expressions in the condition"
        }
      },
      "required": ["op"]
    },
    "Subquery": {
      "type": "object",
      "properties": {
        "dimension": {
          "$ref": "#/$defs/Dimension"
        },
        "measures": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/Measure"
          }
        },
        "where": {
          "$ref": "#/$defs/Expression"
        },
        "having": {
          "$ref": "#/$defs/Expression"
        }
      },
      "required": ["dimension", "measures"]
    },
    "Operator": {
      "type": "string",
      "enum": [
        "",
        "eq",
        "neq",
        "lt",
        "lte",
        "gt",
        "gte",
        "in",
        "nin",
        "ilike",
        "nilike",
        "or",
        "and"
      ],
      "description": "Comparison or logical operator"
    },
    "TimeGrain": {
      "type": "string",
      "enum": [
        "",
        "millisecond",
        "second",
        "minute",
        "hour",
        "day",
        "week",
        "month",
        "quarter",
        "year"
      ],
      "description": "Time granularity"
    }
  }
}
`

var ExpressionJSONSchema = jsonschemautil.MustExtractDefAsSchema(QueryJSONSchema, "Expression")
