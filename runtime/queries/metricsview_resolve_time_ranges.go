package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTimeRanges struct {
	MetricsViewName string                  `json:"metrics_view_name,omitempty"`
	Expressions     []string                `json:"expressions,omitempty"`
	SecurityClaims  *runtime.SecurityClaims `json:"security_claims,omitempty"`
	Priority        int32                   `json:"priority,omitempty"`

	Result *runtimev1.MetricsViewTimeRangesResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewTimeRanges{}

func (q *MetricsViewTimeRanges) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewTimeRanges:%s", string(r))
}

func (q *MetricsViewTimeRanges) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewTimeRanges) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTimeRanges) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTimeRangesResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTimeRanges: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTimeRanges) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	// Resolve metrics view
	mv, _, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	tsRes, err := ResolveTimestampResult(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims, priority)
	if err != nil {
		return err
	}

	// to keep results consistent
	now := time.Now()

	timeRanges := make([]*runtimev1.TimeRange, len(q.Expressions))
	for i, tr := range q.Expressions {
		rillTime, err := rilltime.Parse(tr)
		if err != nil {
			return fmt.Errorf("error parsing time range %s: %w", tr, err)
		}

		start, end, err := rillTime.Eval(rilltime.EvalOptions{
			Now:        now,
			MinTime:    tsRes.Min,
			MaxTime:    tsRes.Max,
			Watermark:  tsRes.Watermark,
			FirstDay:   int(mv.ValidSpec.FirstDayOfWeek),
			FirstMonth: int(mv.ValidSpec.FirstMonthOfYear),
		})
		if err != nil {
			return err
		}

		timeRanges[i] = &runtimev1.TimeRange{
			Start: timestamppb.New(start),
			End:   timestamppb.New(end),
			// for a reference
			Expression: tr,
		}
	}

	q.Result = &runtimev1.MetricsViewTimeRangesResponse{
		TimeRanges: timeRanges,
	}

	return nil
}

func (q *MetricsViewTimeRanges) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return nil
}
