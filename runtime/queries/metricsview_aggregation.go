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
)

type MetricsViewAggregation struct {
	MetricsViewName    string                                       `json:"metrics_view,omitempty"`
	Dimensions         []*runtimev1.MetricsViewAggregationDimension `json:"dimensions,omitempty"`
	Measures           []*runtimev1.MetricsViewAggregationMeasure   `json:"measures,omitempty"`
	Sort               []*runtimev1.MetricsViewAggregationSort      `json:"sort,omitempty"`
	TimeRange          *runtimev1.TimeRange                         `json:"time_range,omitempty"`
	Where              *runtimev1.Expression                        `json:"where,omitempty"`
	Having             *runtimev1.Expression                        `json:"having,omitempty"`
	Priority           int32                                        `json:"priority,omitempty"`
	Limit              *int64                                       `json:"limit,omitempty"`
	Offset             int64                                        `json:"offset,omitempty"`
	MetricsView        *runtimev1.MetricsViewSpec                   `json:"-"`
	ResolvedMVSecurity *runtime.ResolvedMetricsViewSecurity         `json:"security"`

	// backwards compatibility
	Filter *runtimev1.MetricsViewFilter `json:"filter,omitempty"`

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

func (q *MetricsViewAggregation) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
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
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	if q.MetricsView.TimeDimension == "" && !isTimeRangeNil(q.TimeRange) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsView)
	}

	// backwards compatibility
	if q.Filter != nil {
		if q.Where != nil {
			return fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	// Build query
	sql, args, err := q.buildMetricsAggregationSQL(q.MetricsView, olap.Dialect(), q.ResolvedMVSecurity)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	// Execute
	schema, data, err := olapQuery(ctx, olap, priority, sql, args)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewAggregationResponse{
		Schema: schema,
		Data:   data,
	}

	return nil
}

func (q *MetricsViewAggregation) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	filename := strings.ReplaceAll(q.MetricsView.Table, `"`, `_`)
	if !isTimeRangeNil(q.TimeRange) || q.Where != nil || q.Having != nil {
		filename += "_filtered"
	}

	meta := structTypeToMetricsViewColumn(q.Result.Schema)

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
		return writeCSV(meta, q.Result.Data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return writeXLSX(meta, q.Result.Data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return writeParquet(meta, q.Result.Data, w)
	}

	return nil
}

func (q *MetricsViewAggregation) buildMetricsAggregationSQL(mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity) (string, []any, error) {
	if len(q.Dimensions) == 0 && len(q.Measures) == 0 {
		return "", nil, errors.New("no dimensions or measures specified")
	}

	selectCols := make([]string, 0, len(q.Dimensions)+len(q.Measures))
	groupCols := make([]string, 0, len(q.Dimensions))
	unnestClauses := make([]string, 0)
	args := []any{}

	for _, d := range q.Dimensions {
		// Handle regular dimensions
		if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			dim, err := metricsViewDimension(mv, d.Name)
			if err != nil {
				return "", nil, err
			}
			dimSel, unnestClause := dimensionSelect(mv, dim, dialect)
			selectCols = append(selectCols, dimSel)
			if unnestClause != "" {
				unnestClauses = append(unnestClauses, unnestClause)
			}
			groupCols = append(groupCols, safeName(d.Name))
			continue
		}

		// Handle time dimension
		expr, exprArgs, err := q.buildTimestampExpr(d, dialect)
		if err != nil {
			return "", nil, err
		}
		selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(d.Name)))
		groupCols = append(groupCols, expr)
		args = append(args, exprArgs...)
	}

	for _, m := range q.Measures {
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED:
			expr, err := metricsViewMeasureExpression(mv, m.Name)
			if err != nil {
				return "", nil, err
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			selectCols = append(selectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			if len(m.BuiltinMeasureArgs) != 1 {
				return "", nil, fmt.Errorf("builtin measure '%s' expects 1 argument", m.BuiltinMeasure.String())
			}
			arg := m.BuiltinMeasureArgs[0].GetStringValue()
			if arg == "" {
				return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
			}
			selectCols = append(selectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
		default:
			return "", nil, fmt.Errorf("unknown builtin measure '%d'", m.BuiltinMeasure)
		}
	}

	groupClause := ""
	if len(groupCols) > 0 {
		groupClause = "GROUP BY " + strings.Join(groupCols, ", ")
	}

	whereClause := ""
	if mv.TimeDimension != "" {
		timeCol := safeName(mv.TimeDimension)
		clause, err := timeRangeClause(q.TimeRange, mv, dialect, timeCol, &args)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
	}
	if q.Where != nil {
		clause, clauseArgs, err := buildExpression(mv, q.Where, nil, dialect)
		if err != nil {
			return "", nil, err
		}
		whereClause += " AND " + clause
		args = append(args, clauseArgs...)
	}
	if len(whereClause) > 0 {
		whereClause = "WHERE 1=1" + whereClause
	}

	havingClause := ""
	if q.Having != nil {
		var havingClauseArgs []any
		var err error
		havingClause, havingClauseArgs, err = buildExpression(mv, q.Having, nil, dialect)
		if err != nil {
			return "", nil, err
		}
		havingClause = "HAVING " + havingClause
		args = append(args, havingClauseArgs...)
	}

	sortingCriteria := make([]string, 0, len(q.Sort))
	for _, s := range q.Sort {
		sortCriterion := safeName(s.Name)
		if s.Desc {
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

	sql := fmt.Sprintf("SELECT %s FROM %s %s %s %s %s %s %s OFFSET %d",
		strings.Join(selectCols, ", "),
		safeName(mv.Table),
		strings.Join(unnestClauses, ""),
		whereClause,
		groupClause,
		havingClause,
		orderClause,
		limitClause,
		q.Offset,
	)

	return sql, args, nil
}

func (q *MetricsViewAggregation) buildTimestampExpr(dim *runtimev1.MetricsViewAggregationDimension, dialect drivers.Dialect) (string, []any, error) {
	var col string
	if dim.Name == q.MetricsView.TimeDimension {
		col = safeName(dim.Name)
	} else {
		d, err := metricsViewDimension(q.MetricsView, dim.Name)
		if err != nil {
			return "", nil, err
		}
		if d.Expression != "" {
			// TODO: we should add support for this in a future PR
			return "", nil, fmt.Errorf("expression dimension not supported as time column")
		}
		col = metricsViewDimensionExpression(d)
	}

	switch dialect {
	case drivers.DialectDuckDB:
		if dim.TimeZone == "" || dim.TimeZone == "UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)", convertToDateTruncSpecifier(dim.TimeGrain), col), nil, nil
		}
		return fmt.Sprintf("timezone(?, date_trunc('%s', timezone(?, %s::TIMESTAMPTZ)))", convertToDateTruncSpecifier(dim.TimeGrain), col), []any{dim.TimeZone, dim.TimeZone}, nil
	case drivers.DialectDruid:
		if dim.TimeZone == "" || dim.TimeZone == "UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)", convertToDateTruncSpecifier(dim.TimeGrain), col), nil, nil
		}
		return fmt.Sprintf("time_floor(%s, '%s', null, CAST(? AS VARCHAR)))", col, convertToDruidTimeFloorSpecifier(dim.TimeGrain)), []any{dim.TimeZone}, nil
	default:
		return "", nil, fmt.Errorf("unsupported dialect %q", dialect)
	}
}
