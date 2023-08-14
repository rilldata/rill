package queries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewAggregation struct {
	MetricsView              string                       `json:"metrics_view,omitempty"`
	Dimensions               []string                     `json:"dimensions,omitempty"`
	Measures                 []string                     `json:"measures,omitempty"`
	InlineMeasureDefinitions []*runtimev1.InlineMeasure   `json:"inline_measure_definitions,omitempty"`
	TimeStart                *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd                  *timestamppb.Timestamp       `json:"time_end,omitempty"`
	TimeGranularity          runtimev1.TimeGrain          `json:"time_granularity,omitempty"`
	TimeZone                 string                       `json:"time_zone,omitempty"`
	Filter                   *runtimev1.MetricsViewFilter `json:"filter,omitempty"`
	Sort                     []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Priority                 int32                        `json:"priority,omitempty"`
	Limit                    *int64                       `json:"limit,omitempty"`
	Offset                   int64                        `json:"offset,omitempty"`

	Result *runtimev1.MetricsViewAggregationResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewAggregation{}

func (q *MetricsViewAggregation) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewAggregation:%s", string(r))
}

func (q *MetricsViewAggregation) Deps() []string {
	return []string{q.MetricsView}
}

func (q *MetricsViewAggregation) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewAggregation) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewAggregationResponse)
	if !ok {
		return fmt.Errorf("MetricsViewAggregation: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewAggregation) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsView)
	if err != nil {
		return err
	}

	if mv.TimeDimension == "" && (q.TimeStart != nil || q.TimeEnd != nil) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsView)
	}

	// Build query
	sql, args, err := q.buildMetricsAggregationSQL(mv, olap.Dialect())
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	// Execute
	meta, data, err := metricsQuery(ctx, olap, priority, sql, args)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewAggregationResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewAggregation) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	mv, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsView)
	if err != nil {
		return err
	}

	filename := strings.ReplaceAll(mv.Model, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Filter != nil && (len(q.Filter.Include) > 0 || len(q.Filter.Exclude) > 0) {
		filename += "_filtered"
	}

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(filename)
		if err != nil {
			return err
		}
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return writeCSV(q.Result.Meta, q.Result.Data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return writeXLSX(q.Result.Meta, q.Result.Data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return writeParquet(q.Result.Meta, q.Result.Data, w)
	}

	return nil
}

func (q *MetricsViewAggregation) buildMetricsAggregationSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect) (string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasureDefinitions, q.Measures)
	if err != nil {
		return "", nil, err
	}

	if len(q.Dimensions) == 0 && len(ms) == 0 {
		return "", nil, errors.New("no dimensions or measures specified")
	}

	selectCols := make([]string, 0, len(q.Dimensions)+len(ms))
	groupCols := make([]string, 0, len(q.Dimensions))
	args := []any{}

	for _, d := range q.Dimensions {
		if d != mv.TimeDimension {
			col, err := metricsViewDimensionToSafeColumn(mv, d)
			if err != nil {
				return "", nil, err
			}

			selectCols = append(selectCols, col)
			groupCols = append(groupCols, col)
			continue
		}

		// TODO: Handle time dimension
		expr, exprArgs, err := q.buildTimestampExpr(d, dialect)
		if err != nil {
			return "", nil, err
		}
		selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(d)))
		groupCols = append(groupCols, expr)
		args = append(args, exprArgs...)
	}

	for _, m := range ms {
		selectCols = append(selectCols, fmt.Sprintf("%s as %s", m.Expression, safeName(m.Name)))
	}

	groupClause := ""
	if len(groupCols) > 0 {
		groupClause = "GROUP BY " + strings.Join(groupCols, ", ")
	}

	whereClause := ""
	if mv.TimeDimension != "" {
		if q.TimeStart != nil {
			whereClause += fmt.Sprintf(" AND %s >= ?", safeName(mv.TimeDimension))
			args = append(args, q.TimeStart.AsTime())
		}
		if q.TimeEnd != nil {
			whereClause += fmt.Sprintf(" AND %s < ?", safeName(mv.TimeDimension))
			args = append(args, q.TimeEnd.AsTime())
		}
	}
	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, dialect)
		if err != nil {
			return "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
	}
	if len(whereClause) > 0 {
		whereClause = "WHERE 1=1" + whereClause
	}

	sortingCriteria := make([]string, 0, len(q.Sort))
	for _, s := range q.Sort {
		sortCriterion := safeName(s.Name)
		if !s.Ascending {
			sortCriterion += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			sortCriterion += " NULLS LAST"
		}
		sortingCriteria = append(sortingCriteria, sortCriterion)
	}
	orderClause := ""
	if len(sortingCriteria) > 0 {
		orderClause = "ORDER BY " + strings.Join(sortingCriteria, ", ")
	}

	var limitClause string
	if q.Limit != nil {
		if *q.Limit == 0 {
			*q.Limit = 100
		}
		limitClause = fmt.Sprintf("LIMIT %d", *q.Limit)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s %s %s %s %s OFFSET %d",
		strings.Join(selectCols, ", "),
		safeName(mv.Model),
		whereClause,
		groupClause,
		orderClause,
		limitClause,
		q.Offset,
	)

	return sql, args, nil
}

func (q *MetricsViewAggregation) buildTimestampExpr(dim string, dialect drivers.Dialect) (string, []any, error) {
	if q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		return "", nil, fmt.Errorf("querying a timestamp dimension, but time_granularity is not specified")
	}

	switch dialect {
	case drivers.DialectDuckDB:
		if q.TimeZone == "" || q.TimeZone == "UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)", convertToDateTruncSpecifier(q.TimeGranularity), safeName(dim)), nil, nil
		}
		return fmt.Sprintf("timezone(?, date_trunc('%s', timezone(?, %s::TIMESTAMPTZ)))", convertToDateTruncSpecifier(q.TimeGranularity), safeName(dim)), []any{q.TimeZone, q.TimeZone}, nil
	case drivers.DialectDruid:
		if q.TimeZone == "" || q.TimeZone == "UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)", convertToDateTruncSpecifier(q.TimeGranularity), safeName(dim)), nil, nil
		}
		return fmt.Sprintf("time_floor(%s, '%s', null, CAST(? AS VARCHAR)))", safeName(dim), convertToDruidTimeFloorSpecifier(q.TimeGranularity)), []any{q.TimeZone}, nil
	default:
		return "", nil, fmt.Errorf("unsupported dialect %q", dialect)
	}
}
