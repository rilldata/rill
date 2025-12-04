package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewRows struct {
	MetricsViewName    string                       `json:"metrics_view_name,omitempty"`
	TimeStart          *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd            *timestamppb.Timestamp       `json:"time_end,omitempty"`
	TimeGranularity    runtimev1.TimeGrain          `json:"time_granularity,omitempty"`
	Where              *runtimev1.Expression        `json:"where,omitempty"`
	Sort               []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Limit              *int64                       `json:"limit,omitempty"`
	Offset             int64                        `json:"offset,omitempty"`
	TimeZone           string                       `json:"time_zone,omitempty"`
	MetricsView        *runtimev1.MetricsViewSpec   `json:"-"`
	ResolvedMVSecurity *runtime.ResolvedSecurity    `json:"security"`
	Streaming          bool                         `json:"streaming,omitempty"`
	TimeDimension      string                       `json:"time_dimension,omitempty"` // if empty, the default time dimension in mv is used

	// backwards compatibility
	Filter *runtimev1.MetricsViewFilter `json:"filter,omitempty"`

	Result *runtimev1.MetricsViewRowsResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewRows{}

func (q *MetricsViewRows) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewRows:%s", r)
}

func (q *MetricsViewRows) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewRows) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewRows) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewRowsResponse)
	if !ok {
		return fmt.Errorf("MetricsViewRows: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewRows) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	if (q.MetricsView.TimeDimension == "" && q.TimeDimension == "") && (q.TimeStart != nil || q.TimeEnd != nil) {
		return fmt.Errorf("no time dimension specified for metrics view '%s' and time range provided", q.MetricsViewName)
	}

	qry, err := q.rewriteToMetricsViewQuery()
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := executor.New(ctx, rt, instanceID, q.MetricsView, q.Streaming, q.ResolvedMVSecurity, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	res, err := e.Query(ctx, qry, nil)
	if err != nil {
		return err
	}
	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return err
	}

	meta := structTypeToMetricsViewColumn(res.Schema)

	q.Result = &runtimev1.MetricsViewRowsResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewRows) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	if (q.MetricsView.TimeDimension == "" && q.TimeDimension == "") && (q.TimeStart != nil || q.TimeEnd != nil) {
		return fmt.Errorf("no time dimension specified for metrics view '%s' and time range provided", q.MetricsViewName)
	}

	qry, err := q.rewriteToMetricsViewQuery()
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}
	qry.Rows = true

	e, err := executor.New(ctx, rt, instanceID, q.MetricsView, q.Streaming, q.ResolvedMVSecurity, opts.Priority)
	if err != nil {
		return err
	}
	defer e.Close()

	var format drivers.FileFormat
	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		format = drivers.FileFormatCSV
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		format = drivers.FileFormatXLSX
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		format = drivers.FileFormatParquet
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format.String())
	}

	path, err := e.Export(ctx, qry, nil, format, nil)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()

	filename := q.generateFilename(q.MetricsView)
	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(filename)
		if err != nil {
			return err
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	return nil
}

func (q *MetricsViewRows) rewriteToMetricsViewQuery() (*metricsview.Query, error) {
	if q.TimeGranularity != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		return nil, fmt.Errorf("time_granularity is not supported in metrics view rows query")
	}

	qry := &metricsview.Query{MetricsView: q.MetricsViewName}

	res := &metricsview.TimeRange{}
	if q.TimeStart != nil {
		res.Start = q.TimeStart.AsTime()
	}
	if q.TimeEnd != nil {
		res.End = q.TimeEnd.AsTime()
	}
	res.TimeDimension = q.TimeDimension
	qry.TimeRange = res

	qry.Limit = q.Limit

	if qry.Limit != nil && *qry.Limit == 0 {
		*qry.Limit = 100
	}

	if q.Offset != 0 {
		qry.Offset = &q.Offset
	}

	for _, s := range q.Sort {
		qry.Sort = append(qry.Sort, metricsview.Sort{
			Name: s.Name,
			Desc: !s.Ascending,
		})
	}

	if q.Filter != nil { // Backwards compatibility
		if q.Where != nil {
			return nil, fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	var err error
	qry.Where, err = metricViewExpression(q.Where, "")
	if err != nil {
		return nil, fmt.Errorf("error converting where clause: %w", err)
	}

	qry.TimeZone = q.TimeZone
	qry.Rows = true

	return qry, nil
}
