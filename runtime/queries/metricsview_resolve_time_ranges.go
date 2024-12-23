package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewResolveTimeRanges struct {
	MetricsViewName string                  `json:"metrics_view_name,omitempty"`
	MinTime         time.Time               `json:"min_time,omitempty"`
	RillTimes       []string                `json:"rill_times,omitempty"`
	SecurityClaims  *runtime.SecurityClaims `json:"security_claims,omitempty"`

	Result *runtimev1.MetricsViewResolveTimeRangesResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewResolveTimeRanges{}

func (q *MetricsViewResolveTimeRanges) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewResolveTimeRanges:%s", string(r))
}

func (q *MetricsViewResolveTimeRanges) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewResolveTimeRanges) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewResolveTimeRanges) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewResolveTimeRangesResponse)
	if !ok {
		return fmt.Errorf("MetricsViewResolveTimeRanges: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewResolveTimeRanges) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	// Resolve metrics view
	mv, sec, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	e, err := metricsview.NewExecutor(ctx, rt, instanceID, mv, sec, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	watermark, err := e.Watermark(ctx)
	if err != nil {
		return err
	}

	// to keep results consistent
	now := time.Now()

	ranges := make([]*runtimev1.TimeRange, len(q.RillTimes))
	for i, tr := range q.RillTimes {
		rt, err := rilltime.Parse(tr)
		if err != nil {
			return fmt.Errorf("error parsing time range %s: %w", tr, err)
		}

		start, end, err := rt.Resolve(rilltime.ResolverContext{
			Now:        now,
			MinTime:    q.MinTime,
			MaxTime:    watermark,
			FirstDay:   int(mv.FirstDayOfWeek),
			FirstMonth: int(mv.FirstMonthOfYear),
		})
		if err != nil {
			return err
		}

		ranges[i] = &runtimev1.TimeRange{
			Start: timestamppb.New(start),
			End:   timestamppb.New(end),
			// for a reference
			RillTime: tr,
		}
	}

	q.Result = &runtimev1.MetricsViewResolveTimeRangesResponse{
		Ranges: ranges,
	}

	return nil
}

func (q *MetricsViewResolveTimeRanges) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return nil
}
