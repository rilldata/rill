package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTimeSeries struct {
	MetricsViewName    string                               `json:"metrics_view_name,omitempty"`
	MeasureNames       []string                             `json:"measure_names,omitempty"`
	InlineMeasures     []*runtimev1.InlineMeasure           `json:"inline_measures,omitempty"`
	TimeStart          *timestamppb.Timestamp               `json:"time_start,omitempty"`
	TimeEnd            *timestamppb.Timestamp               `json:"time_end,omitempty"`
	Limit              int64                                `json:"limit,omitempty"`
	Offset             int64                                `json:"offset,omitempty"`
	Sort               []*runtimev1.MetricsViewSort         `json:"sort,omitempty"`
	Filter             *runtimev1.MetricsViewFilter         `json:"filter,omitempty"`
	TimeGranularity    runtimev1.TimeGrain                  `json:"time_granularity,omitempty"`
	TimeZone           string                               `json:"time_zone,omitempty"`
	MetricsView        *runtimev1.MetricsView               `json:"-"`
	ResolvedMVSecurity *runtime.ResolvedMetricsViewSecurity `json:"security"`

	Result *runtimev1.MetricsViewTimeSeriesResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewTimeSeries{}

func (q *MetricsViewTimeSeries) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewTimeSeries:%s", r)
}

func (q *MetricsViewTimeSeries) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewTimeSeries) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTimeSeries) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTimeSeriesResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTimeSeries: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTimeSeries) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	if q.MetricsView.TimeDimension == "" {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	obj, err := rt.GetCatalogEntry(ctx, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	mv := obj.GetMetricsView()

	sql, tsAlias, args, err := q.buildMetricsTimeseriesSQL(olap, mv, q.ResolvedMVSecurity)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	// Omit the time value from the result schema
	schema := rows.Schema
	if schema != nil {
		for i, f := range schema.Fields {
			if f.Name == tsAlias {
				schema.Fields = slices.Delete(schema.Fields, i, i+1)
				break
			}
		}
	}

	var data []*runtimev1.TimeSeriesValue
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return err
		}

		var t time.Time
		switch v := rowMap[tsAlias].(type) {
		case time.Time:
			t = v
		default:
			panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))
		}
		delete(rowMap, tsAlias)

		records, err := pbutil.ToStruct(rowMap, schema)
		if err != nil {
			return err
		}

		data = append(data, &runtimev1.TimeSeriesValue{
			Ts:      timestamppb.New(t),
			Records: records,
		})
	}

	meta := structTypeToMetricsViewColumn(rows.Schema)

	q.Result = &runtimev1.MetricsViewTimeSeriesResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewTimeSeries) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	obj, err := rt.GetCatalogEntry(ctx, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	mv := obj.GetMetricsView()

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(q.generateFilename(mv))
		if err != nil {
			return err
		}
	}

	tmp := make([]*structpb.Struct, 0, len(q.Result.Data))
	meta := append([]*runtimev1.MetricsViewColumn{{
		Name: mv.TimeDimension,
	}}, q.Result.Meta...)
	for _, dt := range q.Result.Data {
		dt.Records.Fields[mv.TimeDimension] = structpb.NewStringValue(dt.Ts.AsTime().Format(time.RFC3339Nano))
		tmp = append(tmp, dt.Records)
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return writeCSV(meta, tmp, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return writeXLSX(meta, tmp, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return writeParquet(meta, tmp, w)
	}

	return nil
}

func (q *MetricsViewTimeSeries) generateFilename(mv *runtimev1.MetricsView) string {
	filename := strings.ReplaceAll(mv.Model, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Filter != nil && (len(q.Filter.Include) > 0 || len(q.Filter.Exclude) > 0) {
		filename += "_filtered"
	}
	return filename
}

func (q *MetricsViewTimeSeries) buildMetricsTimeseriesSQL(olap drivers.OLAPStore, mv *runtimev1.MetricsView, policy *runtime.ResolvedMetricsViewSecurity) (string, string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", "", nil, err
	}

	selectCols := []string{}
	for _, m := range ms {
		expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
		selectCols = append(selectCols, expr)
	}

	whereClause := "1=1"
	args := []any{}
	if q.TimeStart != nil {
		whereClause += fmt.Sprintf(" AND %s >= ?", safeName(mv.TimeDimension))
		args = append(args, q.TimeStart.AsTime())
	}
	if q.TimeEnd != nil {
		whereClause += fmt.Sprintf(" AND %s < ?", safeName(mv.TimeDimension))
		args = append(args, q.TimeEnd.AsTime())
	}

	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, drivers.DialectDruid, policy)
		if err != nil {
			return "", "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}

	tsAlias := tempName("_ts_")
	timezone := "UTC"
	if q.TimeZone != "" {
		timezone = q.TimeZone
	}
	args = append([]any{timezone, timezone}, args...)

	var sql string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		sql = q.buildDuckDBSQL(args, mv, tsAlias, selectCols, whereClause)
	case drivers.DialectDruid:
		sql = q.buildDruidSQL(args, mv, tsAlias, selectCols, whereClause)
	default:
		return "", "", nil, fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	return sql, tsAlias, args, nil
}

func (q *MetricsViewTimeSeries) buildDruidSQL(args []any, mv *runtimev1.MetricsView, tsAlias string, selectCols []string, whereClause string) string {
	tsSpecifier := convertToDruidTimeFloorSpecifier(q.TimeGranularity)

	sql := fmt.Sprintf(
		`SELECT time_floor(%s, '%s', null, CAST(? AS VARCHAR)) AS %s, %s FROM %q WHERE %s GROUP BY 1 ORDER BY 1`,
		safeName(mv.TimeDimension),
		tsSpecifier,
		tsAlias,
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
	)

	return sql
}

func (q *MetricsViewTimeSeries) buildDuckDBSQL(args []any, mv *runtimev1.MetricsView, tsAlias string, selectCols []string, whereClause string) string {
	dateTruncSpecifier := convertToDateTruncSpecifier(q.TimeGranularity)
	sql := fmt.Sprintf(
		`SELECT timezone(?, date_trunc('%s', timezone(?, %s::TIMESTAMPTZ))) as %s, %s FROM %q WHERE %s GROUP BY 1 ORDER BY 1`,
		dateTruncSpecifier,
		safeName(mv.TimeDimension),
		tsAlias,
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
	)

	return sql
}
