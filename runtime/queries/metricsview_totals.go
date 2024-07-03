package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTotals struct {
	MetricsViewName    string                       `json:"metrics_view_name,omitempty"`
	MeasureNames       []string                     `json:"measure_names,omitempty"`
	TimeStart          *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd            *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Where              *runtimev1.Expression        `json:"where,omitempty"`
	Filter             *runtimev1.MetricsViewFilter `json:"filter,omitempty"` // backwards compatibility
	SecurityAttributes map[string]any               `json:"security_attributes,omitempty"`
	SecurityRules      []*runtimev1.SecurityRule    `json:"security_policy,omitempty"`

	Result *runtimev1.MetricsViewTotalsResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewTotals{}

func (q *MetricsViewTotals) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewTotals:%s", r)
}

func (q *MetricsViewTotals) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewTotals) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTotals) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTotalsResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTotals: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTotals) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, q.SecurityRules)
	if err != nil {
		return err
	}

	qry, err := q.rewriteToMetricsViewQuery(false)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := metricsview.NewExecutor(ctx, rt, instanceID, mv, security, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	res, _, err := e.Query(ctx, qry, nil)
	if err != nil {
		return err
	}
	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("no data returned")
	}

	q.Result = &runtimev1.MetricsViewTotalsResponse{
		Meta: structTypeToMetricsViewColumn(res.Schema),
		Data: data[0],
	}

	return nil
}

func (q *MetricsViewTotals) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}

func (q *MetricsViewTotals) rewriteToMetricsViewQuery(exporting bool) (*metricsview.Query, error) {
	qry := &metricsview.Query{MetricsView: q.MetricsViewName}

	for _, m := range q.MeasureNames {
		qry.Measures = append(qry.Measures, metricsview.Measure{Name: m})
	}

	if q.TimeStart != nil || q.TimeEnd != nil {
		res := &metricsview.TimeRange{}
		if q.TimeStart != nil {
			res.Start = q.TimeStart.AsTime()
		}
		if q.TimeEnd != nil {
			res.End = q.TimeEnd.AsTime()
		}
		qry.TimeRange = res
	}

	if q.Filter != nil { // Backwards compatibility
		if q.Where != nil {
			return nil, fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	if q.Where != nil {
		qry.Where = metricsview.NewExpressionFromProto(q.Where)
	}

	qry.Label = exporting

	return qry, nil
}
