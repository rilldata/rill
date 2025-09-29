package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"google.golang.org/protobuf/types/known/structpb"
)

type MetricsViewSearch struct {
	MetricsViewName string                  `json:"metrics_view_name,omitempty"`
	Dimensions      []string                `json:"dimensions,omitempty"`
	Search          string                  `json:"search,omitempty"`
	TimeRange       *runtimev1.TimeRange    `json:"time_range,omitempty"`
	Where           *runtimev1.Expression   `json:"where,omitempty"`
	Having          *runtimev1.Expression   `json:"having,omitempty"`
	Priority        int32                   `json:"priority,omitempty"`
	Limit           *int64                  `json:"limit,omitempty"`
	SecurityClaims  *runtime.SecurityClaims `json:"security_claims,omitempty"`

	Result *runtimev1.MetricsViewSearchResponse
}

func (q *MetricsViewSearch) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewSearch:%s", string(r))
}

func (q *MetricsViewSearch) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewSearch) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewSearch) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewSearchResponse)
	if !ok {
		return fmt.Errorf("MetricsViewSearch: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewSearch) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, sec, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	exec, err := executor.New(ctx, rt, instanceID, mv.ValidSpec, mv.Streaming, sec, priority)
	if err != nil {
		return err
	}
	defer exec.Close()

	// build a metricsView.SearchQuery
	search := searchQuery(q)
	rows, err := exec.Search(ctx, search, nil)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewSearchResponse{Results: make([]*runtimev1.MetricsViewSearchResponse_SearchResult, len(rows))}
	for i := range rows {
		v, err := structpb.NewValue(rows[i].Value)
		if err != nil {
			return err
		}

		q.Result.Results[i] = &runtimev1.MetricsViewSearchResponse_SearchResult{
			Dimension: rows[i].Dimension,
			Value:     v,
		}
	}
	return nil
}

func (q *MetricsViewSearch) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return nil
}

func searchQuery(q *MetricsViewSearch) *metricsview.SearchQuery {
	search := &metricsview.SearchQuery{
		Dimensions: q.Dimensions,
		Search:     q.Search,
		Limit:      q.Limit,
	}
	if q.Where != nil {
		search.Where = metricsview.NewExpressionFromProto(q.Where)
	}
	if q.Having != nil {
		search.Having = metricsview.NewExpressionFromProto(q.Having)
	}

	if q.TimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.TimeRange.Start != nil {
			res.Start = q.TimeRange.Start.AsTime()
		}
		if q.TimeRange.End != nil {
			res.End = q.TimeRange.End.AsTime()
		}
		res.IsoDuration = q.TimeRange.IsoDuration
		res.IsoOffset = q.TimeRange.IsoOffset
		res.TimeDimension = q.TimeRange.TimeDimension
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.TimeRange.RoundToGrain)
		search.TimeRange = res
	}
	return search
}
