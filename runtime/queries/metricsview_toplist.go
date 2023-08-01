package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewToplist struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	DimensionName   string                       `json:"dimension_name,omitempty"`
	MeasureNames    []string                     `json:"measure_names,omitempty"`
	InlineMeasures  []*runtimev1.InlineMeasure   `json:"inline_measures,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Limit           *int64                       `json:"limit,omitempty"`
	Offset          int64                        `json:"offset,omitempty"`
	Sort            []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"`

	Result *runtimev1.MetricsViewToplistResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewToplist{}

func (q *MetricsViewToplist) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewToplist:%s", string(r))
}

func (q *MetricsViewToplist) Deps() []string {
	return []string{q.MetricsViewName}
}

func (q *MetricsViewToplist) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewToplist) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewToplistResponse)
	if !ok {
		return fmt.Errorf("MetricsViewToplist: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewToplist) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	mv, lastUpdateOn, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	policy, err := rt.ResolveMetricsViewPolicy(auth.GetClaims(ctx).Attributes(), instanceID, mv, lastUpdateOn)
	if err != nil {
		return err
	}

	if !checkFieldAccess(q.DimensionName, policy) {
		return status.Error(codes.Unauthenticated, "action not allowed")
	}

	if mv.TimeDimension == "" && (q.TimeStart != nil || q.TimeEnd != nil) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	// Build query
	sql, args, err := q.buildMetricsTopListSQL(mv, olap.Dialect(), policy)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	// Execute
	meta, data, err := metricsQuery(ctx, olap, priority, sql, args)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewToplistResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewToplist) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	mv, lastUpdateOn, err := lookupMetricsView(ctx, rt, instanceID, q.MetricsViewName)
	if err != nil {
		return err
	}

	policy, err := rt.ResolveMetricsViewPolicy(auth.GetClaims(ctx).Attributes(), instanceID, mv, lastUpdateOn)
	if err != nil {
		return err
	}

	if !checkFieldAccess(q.DimensionName, policy) {
		return status.Error(codes.Unauthenticated, "action not allowed")
	}

	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		if opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_CSV || opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET {
			if mv.TimeDimension == "" && (q.TimeStart != nil || q.TimeEnd != nil) {
				return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
			}

			sql, args, err := q.buildMetricsTopListSQL(mv, olap.Dialect(), policy)
			if err != nil {
				return err
			}

			filename := q.generateFilename(mv)
			if err := duckDBCopyExport(ctx, rt, instanceID, w, opts, sql, args, filename, olap, mv, opts.Format); err != nil {
				return err
			}
		} else {
			if err := q.generalExport(ctx, rt, instanceID, w, opts, olap, mv); err != nil {
				return err
			}
		}
	case drivers.DialectDruid:
		if err := q.generalExport(ctx, rt, instanceID, w, opts, olap, mv); err != nil {
			return err
		}
	default:
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	return nil
}

func (q *MetricsViewToplist) generalExport(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions, olap drivers.OLAPStore, mv *runtimev1.MetricsView) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(q.generateFilename(mv))
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

func (q *MetricsViewToplist) generateFilename(mv *runtimev1.MetricsView) string {
	filename := strings.ReplaceAll(mv.Model, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Filter != nil && (len(q.Filter.Include) > 0 || len(q.Filter.Exclude) > 0) {
		filename += "_filtered"
	}
	return filename
}

func (q *MetricsViewToplist) buildMetricsTopListSQL(mv *runtimev1.MetricsView, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewPolicy) (string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", nil, err
	}

	colName, err := metricsViewDimensionToSafeColumn(mv, q.DimensionName)
	if err != nil {
		return "", nil, err
	}

	selectCols := []string{colName}
	for _, m := range ms {
		expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
		selectCols = append(selectCols, expr)
	}

	whereClause := "1=1"
	args := []any{}
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
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, dialect, policy)
		if err != nil {
			return "", nil, err
		}
		whereClause += " " + clause
		args = append(args, clauseArgs...)
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

	sql := fmt.Sprintf("SELECT %s FROM %q WHERE %s GROUP BY %s %s %s OFFSET %d",
		strings.Join(selectCols, ", "),
		mv.Model,
		whereClause,
		colName,
		orderClause,
		limitClause,
		q.Offset,
	)

	return sql, args, nil
}

func checkFieldAccess(field string, policy *runtime.ResolvedMetricsViewPolicy) bool {
	if policy != nil {
		if !policy.HasAccess {
			return false
		}

		if len(policy.Include) > 0 {
			for _, include := range policy.Include {
				if include == field {
					return true
				}
			}
		} else if len(policy.Exclude) > 0 {
			for _, exclude := range policy.Exclude {
				if exclude == field {
					return false
				}
			}
		} else {
			// if no include/exclude is specified, then all fields are allowed
			return true
		}
	}
	return true
}
