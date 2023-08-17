package queries

import (
	"context"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

type MetricsViewTimeRange struct {
	MetricsViewName string
	Result          *runtimev1.MetricsViewTimeRangeResponse
}

var _ runtime.Query = &MetricsViewTimeRange{}

func (q *MetricsViewTimeRange) Key() string {
	return fmt.Sprintf("MetricsViewTimeRange:%s", q.MetricsViewName)
}

func (q *MetricsViewTimeRange) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewTimeRange) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTimeRange) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTimeRangeResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTimeRange: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTimeRange) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	if mv.TimeDimension == "" {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	ctr := &ColumnTimeRange{
		TableName:  mv.Model,
		ColumnName: mv.TimeDimension,
	}

	err = rt.Query(ctx, instanceID, ctr, priority)
	if err != nil {
		return err
	}
	q.Result = &runtimev1.MetricsViewTimeRangeResponse{
		TimeRangeSummary: ctr.Result,
	}

	return nil
}

func (q *MetricsViewTimeRange) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
