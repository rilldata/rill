package server

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func safeTimeStr(t *timestamppb.Timestamp) string {
	if t == nil {
		return ""
	}
	return t.AsTime().String()
}

func filterCount(m *runtimev1.MetricsViewFilter) int {
	if m == nil {
		return 0
	}
	return len(m.Include) + len(m.Exclude)
}

func marshalInlineMeasure(ms []*runtimev1.InlineMeasure) []string {
	if len(ms) == 0 {
		return make([]string, 0)
	}

	names := make([]string, len(ms))
	for i := 0; i < len(ms); i++ {
		names[i] = ms[i].Name
	}
	return nil
}

func marshalMetricsViewComparisonSort(ms []*runtimev1.MetricsViewComparisonSort) []string {
	if len(ms) == 0 {
		return make([]string, 0)
	}

	names := make([]string, len(ms))
	for i := 0; i < len(ms); i++ {
		names[i] = ms[i].MeasureName
	}
	return nil
}

func marshalMetricsViewSort(ms []*runtimev1.MetricsViewSort) []string {
	if len(ms) == 0 {
		return make([]string, 0)
	}

	names := make([]string, len(ms))
	for i := 0; i < len(ms); i++ {
		names[i] = ms[i].Name
	}
	return nil
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
