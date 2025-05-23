package metricsview

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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
}

func (q *Query) AsMap() (map[string]any, error) {
	queryMap := make(map[string]any)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &queryMap,
		DecodeHook: func(from reflect.Type, to reflect.Type, data any) (any, error) {
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
		},
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

type Spine struct {
	Where     *WhereSpine `mapstructure:"where"`
	TimeRange *TimeSpine  `mapstructure:"time"`
}

type WhereSpine struct {
	Expression *Expression `mapstructure:"expr"`
}

type TimeSpine struct {
	Start time.Time `mapstructure:"start"`
	End   time.Time `mapstructure:"end"`
	Grain TimeGrain `mapstructure:"grain"`
}

type Sort struct {
	Name string `mapstructure:"name"`
	Desc bool   `mapstructure:"desc"`
}

type TimeRange struct {
	Start        time.Time `mapstructure:"start"`
	End          time.Time `mapstructure:"end"`
	Expression   string    `mapstructure:"expression"`
	IsoDuration  string    `mapstructure:"iso_duration"`
	IsoOffset    string    `mapstructure:"iso_offset"`
	RoundToGrain TimeGrain `mapstructure:"round_to_grain"`
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

const QueryJSONSchema = `
{
  "type": "object",
  "properties": {
    "metrics_view": {
      "type": "string",
      "description": "The metrics view to query"
    },
    "dimensions": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/Dimension"
      },
      "description": "List of dimensions to include in the query"
    },
    "measures": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/Measure"
      },
      "description": "List of measures to include in the query"
    },
    "pivot_on": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Dimensions to pivot on"
    },
    "spine": {
      "$ref": "#/definitions/Spine",
      "description": "Spine configuration for the query"
    },
    "sort": {
      "type": "array",
      "items": {
        "$ref": "#/definitions/Sort"
      },
      "description": "Sort order for the results"
    },
    "time_range": {
      "$ref": "#/definitions/TimeRange",
      "description": "Time range filter for the query"
    },
    "comparison_time_range": {
      "$ref": "#/definitions/TimeRange",
      "description": "Time range for comparison"
    },
    "where": {
      "$ref": "#/definitions/Expression",
      "description": "WHERE clause expression"
    },
    "having": {
      "$ref": "#/definitions/Expression",
      "description": "HAVING clause expression"
    },
    "limit": {
      "type": "integer",
      "minimum": 0,
      "description": "Maximum number of rows to return"
    },
    "offset": {
      "type": "integer",
      "minimum": 0,
      "description": "Number of rows to skip"
    },
    "time_zone": {
      "type": "string",
      "description": "Time zone for the query"
    },
    "use_display_names": {
      "type": "boolean",
      "description": "Whether to use display names"
    },
    "rows": {
      "type": "boolean",
      "description": "Whether to return raw rows"
    }
  },
  "definitions": {
    "Dimension": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Name of the dimension"
        },
        "compute": {
          "$ref": "#/definitions/DimensionCompute",
          "description": "Compute configuration for the dimension"
        }
      },
      "required": ["name"]
    },
    "DimensionCompute": {
      "type": "object",
      "properties": {
        "time_floor": {
          "$ref": "#/definitions/DimensionComputeTimeFloor"
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
          "$ref": "#/definitions/TimeGrain",
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
          "$ref": "#/definitions/MeasureCompute",
          "description": "Compute configuration for the measure"
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
          "$ref": "#/definitions/MeasureComputeCountDistinct"
        },
        "comparison_value": {
          "$ref": "#/definitions/MeasureComputeComparisonValue"
        },
        "comparison_delta": {
          "$ref": "#/definitions/MeasureComputeComparisonDelta"
        },
        "comparison_ratio": {
          "$ref": "#/definitions/MeasureComputeComparisonRatio"
        },
        "percent_of_total": {
          "$ref": "#/definitions/MeasureComputePercentOfTotal"
        },
        "uri": {
          "$ref": "#/definitions/MeasureComputeURI"
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
          "$ref": "#/definitions/WhereSpine"
        },
        "time": {
          "$ref": "#/definitions/TimeSpine"
        }
      }
    },
    "WhereSpine": {
      "type": "object",
      "properties": {
        "expr": {
          "$ref": "#/definitions/Expression"
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
          "$ref": "#/definitions/TimeGrain",
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
          "description": "Start time"
        },
        "end": {
          "type": "string",
          "format": "date-time",
          "description": "End time"
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
          "$ref": "#/definitions/TimeGrain",
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
          "$ref": "#/definitions/Condition"
        },
        "subquery": {
          "$ref": "#/definitions/Subquery"
        }
      }
    },
    "Condition": {
      "type": "object",
      "properties": {
        "op": {
          "$ref": "#/definitions/Operator",
          "description": "Operator for the condition"
        },
        "exprs": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Expression"
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
          "$ref": "#/definitions/Dimension"
        },
        "measures": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Measure"
          }
        },
        "where": {
          "$ref": "#/definitions/Expression"
        },
        "having": {
          "$ref": "#/definitions/Expression"
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
  },
  "dependencies": {
    "rows": {
      "oneOf": [
        {
          "properties": {
            "rows": {"const": false}
          }
        },
        {
          "properties": {
            "rows": {"const": true},
            "dimensions": {"maxItems": 0},
            "measures": {"maxItems": 0},
            "sort": {"maxItems": 0},
            "comparison_time_range": {"not": {}},
            "having": {"not": {}},
            "pivot_on": {"maxItems": 0}
          }
        }
      ]
    }
  }
}
`
