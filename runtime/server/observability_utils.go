package server

import (
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func safeTimeStr(t *timestamppb.Timestamp) string {
	if t == nil {
		return ""
	}
	return t.AsTime().String()
}

func filterCount(m *runtimev1.Expression) int {
	if m == nil {
		return 0
	}
	c := 0
	switch e := m.Expression.(type) {
	case *runtimev1.Expression_Ident:
		c++
	case *runtimev1.Expression_Cond:
		for _, expr := range e.Cond.Exprs {
			c += filterCount(expr)
		}
	}
	return c
}

func marshalTimeRange(m *runtimev1.TimeRange) string {
	if m == nil {
		return ""
	}
	var tr strings.Builder
	if m.Start != nil {
		tr.WriteString(safeTimeStr(m.Start))
		tr.WriteByte('-')
		tr.WriteString(safeTimeStr(m.End))
	}
	if m.IsoDuration != "" {
		if tr.Len() > 0 {
			tr.WriteByte('-')
		}
		tr.WriteString(m.IsoDuration)
	}
	if m.IsoOffset != "" {
		if tr.Len() > 0 {
			tr.WriteByte('-')
		}
		tr.WriteString(m.IsoOffset)
	}

	tr.WriteByte('-')
	tr.WriteString(m.RoundToGrain.String())

	if m.TimeZone != "" {
		if tr.Len() > 0 {
			tr.WriteByte('-')
		}
		tr.WriteString(m.TimeZone)
	}
	if m.Expression != "" {
		if tr.Len() > 0 {
			tr.WriteByte('-')
		}
		tr.WriteString(m.Expression)
	}
	if m.TimeDimension != "" {
		if tr.Len() > 0 {
			tr.WriteByte('-')
		}
		tr.WriteString(m.TimeDimension)
	}

	return tr.String()
}

func marshalMetricsViewAggregationDimension(ms []*runtimev1.MetricsViewAggregationDimension) []string {
	if len(ms) == 0 {
		return make([]string, 0)
	}

	names := make([]string, len(ms))
	for i := 0; i < len(ms); i++ {
		names[i] = ms[i].Name
	}
	return names
}

func marshalMetricsViewAggregationMeasures(ms []*runtimev1.MetricsViewAggregationMeasure) []string {
	if len(ms) == 0 {
		return make([]string, 0)
	}

	names := make([]string, len(ms))
	for i := 0; i < len(ms); i++ {
		names[i] = ms[i].Name
	}
	return names
}

func marshalMetricsViewAggregationSort(ms []*runtimev1.MetricsViewAggregationSort) []string {
	if len(ms) == 0 {
		return make([]string, 0)
	}

	names := make([]string, len(ms))
	for i := 0; i < len(ms); i++ {
		names[i] = ms[i].Name
	}
	return names
}

func marshalMetricsViewComparisonSort(ms []*runtimev1.MetricsViewComparisonSort) []string {
	if len(ms) == 0 {
		return make([]string, 0)
	}

	names := make([]string, len(ms))
	for i := 0; i < len(ms); i++ {
		names[i] = ms[i].Name
	}
	return names
}

func marshalMetricsViewSort(ms []*runtimev1.MetricsViewSort) []string {
	if len(ms) == 0 {
		return make([]string, 0)
	}

	names := make([]string, len(ms))
	for i := 0; i < len(ms); i++ {
		names[i] = ms[i].Name
	}
	return names
}

func marshalColumnTimeSeriesRequestBasicMeasure(m []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure) []string {
	if len(m) == 0 {
		return make([]string, 0)
	}
	ids := make([]string, len(m))
	for i := 0; i < len(m); i++ {
		ids[i] = m[i].Id
	}
	return ids
}
