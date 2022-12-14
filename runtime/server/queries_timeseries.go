package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"google.golang.org/protobuf/types/known/structpb"
)

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getExpressionColumnsFromMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += measure.Expression + " as " + measure.SqlName
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getCoalesceStatementsMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += fmt.Sprintf(`COALESCE(series.%s, 0) as %s`, measure.SqlName, measure.SqlName)
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

func EscapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func getFilterFromDimensionValuesFilter(
	dimensionValues []*runtimev1.MetricsViewDimensionValue,
	prefix string,
	dimensionJoiner string,
) string {
	if len(dimensionValues) == 0 {
		return ""
	}
	var result string
	conditions := make([]string, 3)
	if len(dimensionValues) > 0 {
		result += " ( "
	}
	for i, dv := range dimensionValues {
		escapedName := EscapeSingleQuotes(dv.Name)
		var nulls bool
		var notNulls bool
		for _, iv := range dv.In {
			if _, ok := iv.Kind.(*structpb.Value_NullValue); ok {
				nulls = true
			} else {
				notNulls = true
			}
		}
		conditions = conditions[:0]
		if notNulls {
			var inClause = escapedName + " " + prefix + " IN ("
			for j, iv := range dv.In {
				if _, ok := iv.Kind.(*structpb.Value_NullValue); !ok {
					inClause += "'" + EscapeSingleQuotes(iv.GetStringValue()) + "'"
					if j < len(dv.In)-1 {
						inClause += ", "
					}
				}
			}
			inClause += ")"
			conditions = append(conditions, inClause)
		}
		if nulls {
			var nullClause = escapedName + " IS " + prefix + " NULL"
			conditions = append(conditions, nullClause)
		}
		if len(dv.Like) > 0 {
			var likeClause string
			for j, lv := range dv.Like {
				if lv.GetKind() == nil {
					continue
				}
				likeClause += escapedName + " " + prefix + " ILIKE '" + EscapeSingleQuotes(lv.GetStringValue()) + "'"
				if j < len(dv.Like)-1 {
					likeClause += " OR "
				}
			}
			conditions = append(conditions, likeClause)
		}
		result += strings.Join(conditions, " "+dimensionJoiner+" ")
		if i < len(dimensionValues)-1 {
			result += ") AND ("
		}
	}
	result += " ) "

	return result
}

func getFilterFromMetricsViewFilters(filters *runtimev1.MetricsViewRequestFilter) string {
	includeFilters := getFilterFromDimensionValuesFilter(filters.Include, "", "OR")
	excludeFilters := getFilterFromDimensionValuesFilter(filters.Exclude, "NOT", "AND")
	if includeFilters != "" && excludeFilters != "" {
		return " ( " + includeFilters + ") AND (" + excludeFilters + ")"
	} else if includeFilters != "" {
		return includeFilters
	} else if excludeFilters != "" {
		return excludeFilters
	} else {
		return ""
	}
}

func getFallbackMeasureName(index int, sqlName string) string {
	if sqlName == "" {
		s := fmt.Sprintf("measure_%d", index)
		return s
	} else {
		return sqlName
	}
}

var countName = "count"

func normaliseMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure {
	if len(measures) == 0 {
		return []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    countName,
				Id:         "",
			},
		}
	}
	for i, measure := range measures {
		measure.SqlName = getFallbackMeasureName(i, measure.SqlName)
	}
	return measures
}

// Metrics/Timeseries APIs
func (s *Server) EstimateRollupInterval(ctx context.Context, request *runtimev1.EstimateRollupIntervalRequest) (*runtimev1.EstimateRollupIntervalResponse, error) {
	q := &queries.RollupInterval{
		TableName:  request.TableName,
		ColumnName: request.ColumnName,
	}
	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func (s *Server) GenerateTimeSeries(ctx context.Context, request *runtimev1.GenerateTimeSeriesRequest) (*runtimev1.GenerateTimeSeriesResponse, error) {
	q := &queries.ColumnTimeseries{
		TableName:           request.TableName,
		TimestampColumnName: request.TimestampColumnName,
		Measures:            request.Measures,
		Filters:             request.Filters,
		TimeRange:           request.TimeRange,
		Pixels:              request.Pixels,
		SampleSize:          request.SampleSize,
	}
	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateTimeSeriesResponse{
		Rollup: q.Result,
	}, nil
}
