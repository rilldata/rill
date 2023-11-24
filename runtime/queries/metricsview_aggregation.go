package queries

import (
	"context"
	databasesql "database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	duckdbolap "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/drivers/duckdb/transporter"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
)

var (
	dbOnce    sync.Once
	handle    drivers.Handle
	errHandle error
)

type MetricsViewAggregation struct {
	MetricsViewName    string                                       `json:"metrics_view,omitempty"`
	Dimensions         []*runtimev1.MetricsViewAggregationDimension `json:"dimensions,omitempty"`
	Measures           []*runtimev1.MetricsViewAggregationMeasure   `json:"measures,omitempty"`
	Sort               []*runtimev1.MetricsViewAggregationSort      `json:"sort,omitempty"`
	TimeRange          *runtimev1.TimeRange                         `json:"time_range,omitempty"`
	Filter             *runtimev1.MetricsViewFilter                 `json:"filter,omitempty"`
	Priority           int32                                        `json:"priority,omitempty"`
	Limit              *int64                                       `json:"limit,omitempty"`
	Offset             int64                                        `json:"offset,omitempty"`
	MetricsView        *runtimev1.MetricsViewSpec                   `json:"-"`
	ResolvedMVSecurity *runtime.ResolvedMetricsViewSecurity         `json:"security"`
	PivotOn            []string                                     `json:"pivot_on,omitempty"`

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

	// execute druid query
	sqlString, args, err := q.buildMetricsAggregationSQL(q.MetricsView, olap.Dialect(), q.ResolvedMVSecurity)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	if len(q.PivotOn) == 0 {
		schema, data, err := olapQuery(ctx, olap, priority, sqlString, args)
		if err != nil {
			return err
		}

		q.Result = &runtimev1.MetricsViewAggregationResponse{
			Schema: schema,
			Data:   data,
		}
		return nil
	}

	if olap.Dialect() == drivers.DialectDuckDB {
		return olap.WithConnection(ctx, priority, false, false, func(ctx context.Context, ensuredCtx context.Context, conn *databasesql.Conn) error {
			temporaryTableName := tempName("_for_pivot_")

			err = olap.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("CREATE TEMPORARY TABLE %s AS %s", temporaryTableName, sqlString),
				Args:     args,
				Priority: priority,
			})
			if err != nil {
				return err
			}
			defer func() {
				_ = olap.Exec(ensuredCtx, &drivers.Statement{
					Query: `DROP TABLE "` + temporaryTableName + `"`,
				})
			}()

			return q.pivotOn(ctx, olap, temporaryTableName)
		})
	}
	dbOnce.Do(func() {
		handle, errHandle = duckdbolap.Driver{}.Open(map[string]any{"pool_size": 10}, false, activity.NewNoopClient(), zap.NewNop())
	})
	if errHandle != nil {
		return errHandle
	}

	schema, data, err := olapQuery(ctx, olap, priority, sqlString, args)
	if err != nil {
		return err
	}

	duckDBOLAP, _ := handle.AsOLAP("")
	err = duckDBOLAP.WithConnection(ctx, priority, false, false, func(ctx context.Context, ensuredCtx context.Context, conn *databasesql.Conn) error {
		temporaryTableName := tempName("_for_pivot_")
		createTableSQL, err := transporter.CreateTableQuery(schema, temporaryTableName)
		if err != nil {
			return err
		}

		err = duckDBOLAP.Exec(ctx, &drivers.Statement{
			Query: createTableSQL,
		})
		if err != nil {
			return err
		}
		defer func() {
			_ = duckDBOLAP.Exec(ensuredCtx, &drivers.Statement{
				Query: `DROP TABLE "` + temporaryTableName + `"`,
			})
		}()

		err = transporter.RawConn(conn, func(conn driver.Conn) error {
			appender, err := duckdb.NewAppenderFromConn(conn, "", temporaryTableName)
			if err != nil {
				return err
			}
			defer appender.Close()

			batchSize := 10000
			arr := make([]driver.Value, 0, len(schema.Fields))
			count := 0
			for _, row := range data {
				for _, key := range schema.Fields {
					arr = append(arr, row.Fields[key.Name].AsInterface())
				}
				err = appender.AppendRowArray(arr)
				if err != nil {
					return err
				}
				arr = arr[:0]
				count++
				if count >= batchSize {
					appender.Flush()
					count = 0
				}
			}
			appender.Flush()

			return nil
		})
		if err != nil {
			return err
		}

		return q.pivotOn(ctx, duckDBOLAP, temporaryTableName)
	})
	return err
}

func (q *MetricsViewAggregation) pivotOn(ctx context.Context, olap drivers.OLAPStore, temporaryTableName string) error {
	measureCols := make([]string, 0, len(q.Measures))
	for _, m := range q.Measures {
		sn := safeName(m.Name)
		measureCols = append(measureCols, fmt.Sprintf("LAST(%s) as %s", sn, sn))
	}

	sortingCriteria := make([]string, 0, len(q.Sort))
	for _, s := range q.Sort {
		sortCriterion := safeName(s.Name)
		if s.Desc {
			sortCriterion += " DESC"
		}
		if olap.Dialect() == drivers.DialectDuckDB {
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

	// execute duckdb pivot
	//	PIVOT t ON year USING LAST(ap) ap;
	pivotSQL := fmt.Sprintf("PIVOT %[1]s ON %[2]s USING %[3]s %[4]s %[5]s OFFSET %[6]d",
		temporaryTableName,              // 1
		strings.Join(q.PivotOn, ", "),   // 2
		strings.Join(measureCols, ", "), // 3
		orderClause,                     // 4
		limitClause,                     // 5
		q.Offset,                        // 6
	)
	schema, data, err := olapQuery(ctx, olap, int(q.Priority), pivotSQL, nil)
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
	if !isTimeRangeNil(q.TimeRange) || q.Filter != nil && (len(q.Filter.Include) > 0 || len(q.Filter.Exclude) > 0) {
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

	cols := len(q.Dimensions) + len(q.Measures)
	selectCols := make([]string, 0, cols)

	groupCols := make([]string, 0, len(q.Dimensions))
	args := []any{}
	for _, d := range q.Dimensions {
		// Handle regular dimensions
		if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			col, err := metricsViewDimensionToSafeColumn(mv, d.Name)
			if err != nil {
				return "", nil, err
			}

			selectCols = append(selectCols, fmt.Sprintf("%s as %s", col, safeName(d.Name)))
			groupCols = append(groupCols, col)
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
		sn := safeName(m.Name)
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED:
			expr, err := metricsViewMeasureExpression(mv, m.Name)
			if err != nil {
				return "", nil, err
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, sn))
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
		clause, err := timeRangeClause(q.TimeRange, dialect, timeCol, &args)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
	}
	if q.Filter != nil {
		clause, clauseArgs, err := buildFilterClauseForMetricsViewFilter(mv, q.Filter, dialect, policy)
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
	var sql string
	if q.PivotOn != nil {
		// select m1, m2, d1, d2 from t where d1 = 'a' group by d1, d2
		sql = fmt.Sprintf("SELECT %[1]s FROM %[2]s %[3]s %[4]s",
			strings.Join(selectCols, ", "), // 1
			safeName(mv.Table),             // 2
			whereClause,                    // 3
			groupClause,                    // 4
		)
	} else {
		sql = fmt.Sprintf("SELECT %s FROM %s %s %s %s %s OFFSET %d",
			strings.Join(selectCols, ", "),
			safeName(mv.Table),
			whereClause,
			groupClause,
			orderClause,
			limitClause,
			q.Offset,
		)
	}

	return sql, args, nil
}

func (q *MetricsViewAggregation) buildTimestampExpr(dim *runtimev1.MetricsViewAggregationDimension, dialect drivers.Dialect) (string, []any, error) {
	var colName string
	if dim.Name == q.MetricsView.TimeDimension {
		colName = dim.Name
	} else {
		col, err := metricsViewDimensionToSafeColumn(q.MetricsView, dim.Name)
		if err != nil {
			return "", nil, err
		}
		colName = col
	}

	switch dialect {
	case drivers.DialectDuckDB:
		if dim.TimeZone == "" || dim.TimeZone == "UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)", convertToDateTruncSpecifier(dim.TimeGrain), safeName(colName)), nil, nil
		}
		return fmt.Sprintf("timezone(?, date_trunc('%s', timezone(?, %s::TIMESTAMPTZ)))", convertToDateTruncSpecifier(dim.TimeGrain), safeName(colName)), []any{dim.TimeZone, dim.TimeZone}, nil
	case drivers.DialectDruid:
		if dim.TimeZone == "" || dim.TimeZone == "UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)", convertToDateTruncSpecifier(dim.TimeGrain), safeName(colName)), nil, nil
		}
		return fmt.Sprintf("time_floor(%s, '%s', null, CAST(? AS VARCHAR)))", safeName(colName), convertToDruidTimeFloorSpecifier(dim.TimeGrain)), []any{dim.TimeZone}, nil
	default:
		return "", nil, fmt.Errorf("unsupported dialect %q", dialect)
	}
}
