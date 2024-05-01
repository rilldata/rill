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

	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	duckdbolap "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
)

type MetricsViewAggregation struct {
	MetricsViewName     string                                       `json:"metrics_view,omitempty"`
	Dimensions          []*runtimev1.MetricsViewAggregationDimension `json:"dimensions,omitempty"`
	Measures            []*runtimev1.MetricsViewAggregationMeasure   `json:"measures,omitempty"`
	Sort                []*runtimev1.MetricsViewComparisonSort       `json:"sort,omitempty"`
	TimeRange           *runtimev1.TimeRange                         `json:"time_range,omitempty"`
	ComparisonTimeRange *runtimev1.TimeRange                         `json:"comparison_time_range,omitempty"`
	Where               *runtimev1.Expression                        `json:"where,omitempty"`
	Having              *runtimev1.Expression                        `json:"having,omitempty"`
	Filter              *runtimev1.MetricsViewFilter                 `json:"filter,omitempty"` // Backwards compatibility
	Priority            int32                                        `json:"priority,omitempty"`
	Limit               *int64                                       `json:"limit,omitempty"`
	Offset              int64                                        `json:"offset,omitempty"`
	PivotOn             []string                                     `json:"pivot_on,omitempty"`
	SecurityAttributes  map[string]any                               `json:"security_attributes,omitempty"`

	Exporting bool

	Result *runtimev1.MetricsViewAggregationResponse `json:"-"`

	measuresMeta       map[string]metricsViewMeasureMeta `json:"-"`
	ComparisonMeasures []string
	Exact              bool
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
	// Resolve metrics view
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, q.Dimensions, q.Measures)
	if err != nil {
		return err
	}

	cfg, err := rt.InstanceConfig(ctx, instanceID)
	if err != nil {
		return err
	}

	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid && olap.Dialect() != drivers.DialectClickHouse {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	if mv.TimeDimension == "" && !isTimeRangeNil(q.TimeRange) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", mv)
	}

	// backwards compatibility
	if q.Filter != nil {
		if q.Where != nil {
			return fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	if q.ComparisonTimeRange != nil {
		return q.executeComparisonAggregation(ctx, olap, priority, mv, olap.Dialect(), security)
	}

	if olap.Dialect() == drivers.DialectDuckDB {
		sqlString, args, err := q.buildMetricsAggregationSQL(mv, olap.Dialect(), security, cfg.PivotCellLimit)
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
		return olap.WithConnection(ctx, priority, false, false, func(ctx context.Context, ensuredCtx context.Context, conn *databasesql.Conn) error {
			temporaryTableName := tempName("_for_pivot_")

			err := olap.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("CREATE TEMPORARY TABLE %[1]s AS %[2]s", temporaryTableName, sqlString),
				Args:     args,
				Priority: priority,
			})
			if err != nil {
				return err
			}

			res, err := olap.Execute(ctx, &drivers.Statement{ // a separate query instead of the multi-statement query due to a DuckDB bug
				Query:    fmt.Sprintf("SELECT COUNT(*) FROM %[1]s", temporaryTableName),
				Priority: priority,
			})
			if err != nil {
				return err
			}

			count := 0
			if res.Next() {
				err := res.Scan(&count)
				if err != nil {
					res.Close()
					return err
				}

				if count > int(cfg.PivotCellLimit)/q.cols() {
					res.Close()
					return fmt.Errorf("PIVOT cells count exceeded %d", cfg.PivotCellLimit)
				}
			}
			res.Close()

			defer func() {
				_ = olap.Exec(ensuredCtx, &drivers.Statement{
					Query: `DROP TABLE "` + temporaryTableName + `"`,
				})
			}()

			schema, data, err := olapQuery(ctx, olap, int(q.Priority), q.createPivotSQL(temporaryTableName, mv), nil)
			if err != nil {
				return err
			}

			if q.Limit != nil && *q.Limit > 0 && int64(len(data)) > *q.Limit {
				return fmt.Errorf("Limit exceeded %d", *q.Limit)
			}

			q.Result = &runtimev1.MetricsViewAggregationResponse{
				Schema: schema,
				Data:   data,
			}

			return nil
		})
	}

	sqlString, args, err := q.buildMetricsAggregationSQL(mv, olap.Dialect(), security, cfg.PivotCellLimit)
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

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sqlString,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil
	}
	defer rows.Close()

	return q.pivotDruid(ctx, rows, mv, cfg.PivotCellLimit)
}

func (q *MetricsViewAggregation) executeComparisonAggregation(ctx context.Context, olap drivers.OLAPStore, priority int, mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, security *runtime.ResolvedMetricsViewSecurity) error {
	sqlString, args, err := q.buildMetricsComparisonAggregationSQL2(ctx, olap, priority, mv, dialect, security, false)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

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

func (q *MetricsViewAggregation) pivotDruid(ctx context.Context, rows *drivers.Result, mv *runtimev1.MetricsViewSpec, pivotCellLimit int64) error {
	pivotDB, err := sqlx.Connect("duckdb", "")
	if err != nil {
		return err
	}
	defer pivotDB.Close()

	return func() error {
		temporaryTableName := tempName("_for_pivot_")
		createTableSQL, err := duckdbolap.CreateTableQuery(rows.Schema, temporaryTableName)
		if err != nil {
			return err
		}

		_, err = pivotDB.ExecContext(ctx, createTableSQL)
		if err != nil {
			return err
		}
		defer func() {
			_, _ = pivotDB.ExecContext(context.Background(), `DROP TABLE "`+temporaryTableName+`"`)
		}()

		conn, err := pivotDB.Conn(ctx)
		if err != nil {
			return nil
		}
		defer conn.Close()

		err = conn.Raw(func(conn any) error {
			driverCon, ok := conn.(driver.Conn)
			if !ok {
				return fmt.Errorf("cannot obtain driver.Conn")
			}
			appender, err := duckdb.NewAppenderFromConn(driverCon, "", temporaryTableName)
			if err != nil {
				return err
			}
			defer appender.Close()

			batchSize := 10000
			columns, err := rows.Columns()
			if err != nil {
				return err
			}

			scanValues := make([]any, len(columns))
			appendValues := make([]driver.Value, len(columns))
			for i := range scanValues {
				scanValues[i] = new(interface{})
			}
			count := 0
			maxCount := int(pivotCellLimit) / q.cols()

			for rows.Next() {
				err = rows.Scan(scanValues...)
				if err != nil {
					return err
				}
				for i := range columns {
					appendValues[i] = driver.Value(*(scanValues[i].(*interface{})))
				}
				err = appender.AppendRow(appendValues...)
				if err != nil {
					return fmt.Errorf("duckdb append failed: %w", err)
				}
				count++
				if count > maxCount {
					return fmt.Errorf("PIVOT cells count limit exceeded %d", pivotCellLimit)
				}

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
		if rows.Err() != nil {
			return rows.Err()
		}

		ctx, cancelFunc := context.WithTimeout(ctx, defaultExecutionTimeout)
		defer cancelFunc()
		pivotRows, err := pivotDB.QueryxContext(ctx, q.createPivotSQL(temporaryTableName, mv))
		if err != nil {
			return err
		}
		defer pivotRows.Close()

		schema, err := duckdbolap.RowsToSchema(pivotRows)
		if err != nil {
			return err
		}

		data, err := toData(pivotRows, schema)
		if err != nil {
			return err
		}

		if q.Limit != nil && *q.Limit > 0 && int64(len(data)) > *q.Limit {
			return fmt.Errorf("Limit exceeded %d", *q.Limit)
		}

		q.Result = &runtimev1.MetricsViewAggregationResponse{
			Schema: schema,
			Data:   data,
		}

		return nil
	}()
}

func (q *MetricsViewAggregation) createPivotSQL(temporaryTableName string, mv *runtimev1.MetricsViewSpec) string {
	selectCols := make([]string, 0, len(q.Dimensions)+len(q.Measures))
	aliasesMap := make(map[string]string)
	pivotMap := make(map[string]bool)
	for _, p := range q.PivotOn {
		pivotMap[p] = true
	}
	if q.Exporting {
		for _, e := range mv.Measures {
			aliasesMap[e.Name] = e.Name
			if e.Label != "" {
				aliasesMap[e.Name] = e.Label
			}
		}

		for _, e := range mv.Dimensions {
			aliasesMap[e.Name] = e.Name
			if e.Label != "" {
				aliasesMap[e.Name] = e.Label
			}
		}
		for _, e := range q.Dimensions {
			if e.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				aliasesMap[e.Name] = e.Name
				if e.Alias != "" {
					aliasesMap[e.Alias] = e.Alias
				}
			}
		}

		for _, d := range q.Dimensions {
			if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				expr := safeName(d.Name)
				if pivotMap[d.Name] {
					expr = fmt.Sprintf("lower(%s)", safeName(d.Name))
				}
				selectCols = append(selectCols, fmt.Sprintf("%s AS %s", expr, safeName(aliasesMap[d.Name])))
			} else {
				alias := d.Name
				if d.Alias != "" {
					alias = d.Alias
				}
				selectCols = append(selectCols, safeName(alias))
			}
		}
		for _, m := range q.Measures {
			selectCols = append(selectCols, fmt.Sprintf("%s AS %s", safeName(m.Name), safeName(aliasesMap[m.Name])))
		}
	}
	measureCols := make([]string, 0, len(q.Measures))
	for _, m := range q.Measures {
		alias := safeName(m.Name)
		if q.Exporting {
			alias = safeName(aliasesMap[m.Name])
		}
		measureCols = append(measureCols, fmt.Sprintf("LAST(%s) as %s", alias, alias))
	}

	pivots := make([]string, len(q.PivotOn))
	for i, p := range q.PivotOn {
		pivots[i] = p
		if q.Exporting {
			pivots[i] = safeName(aliasesMap[p])
		}
	}

	sortingCriteria := make([]string, 0, len(q.Sort))
	for _, s := range q.Sort {
		sortCriterion := safeName(s.Name)
		if q.Exporting {
			sortCriterion = safeName(aliasesMap[s.Name])
		}

		if s.Desc {
			sortCriterion += " DESC"
		}
		sortCriterion += " NULLS LAST"
		sortingCriteria = append(sortingCriteria, sortCriterion)
	}

	orderClause := ""
	if len(sortingCriteria) > 0 {
		orderClause = "ORDER BY " + strings.Join(sortingCriteria, ", ")
	}

	var limitClause string
	if q.Limit != nil {
		limit := *q.Limit
		if limit == 0 {
			limit = 100
		}
		if q.Exporting && *q.Limit > 0 {
			limit = *q.Limit + 1
		}
		limitClause = fmt.Sprintf("LIMIT %d", limit)
	}

	// PIVOT (SELECT m1 as M1, d1 as D1, d2 as D2)
	// ON D1 USING LAST(M1) as M1
	// ORDER BY D2 LIMIT 10 OFFSET 0
	selectList := "*"
	if q.Exporting {
		selectList = strings.Join(selectCols, ",")
	}
	return fmt.Sprintf("PIVOT (SELECT %[7]s FROM %[1]s) ON %[2]s USING %[3]s %[4]s %[5]s OFFSET %[6]d",
		temporaryTableName,              // 1
		strings.Join(pivots, ", "),      // 2
		strings.Join(measureCols, ", "), // 3
		orderClause,                     // 4
		limitClause,                     // 5
		q.Offset,                        // 6
		selectList,                      // 7
	)
}

func toData(rows *sqlx.Rows, schema *runtimev1.StructType) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := pbutil.ToStruct(rowMap, schema)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (q *MetricsViewAggregation) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	q.Exporting = true
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("timeout exceeded")
		}
		return err
	}

	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
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
		return WriteCSV(meta, q.Result.Data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return WriteXLSX(meta, q.Result.Data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return WriteParquet(meta, q.Result.Data, w)
	}

	return nil
}

func (q *MetricsViewAggregation) cols() int {
	return len(q.Dimensions) + len(q.Measures)
}

func (q *MetricsViewAggregation) buildMetricsAggregationSQL(mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity, pivotCellLimit int64) (string, []any, error) {
	if len(q.Dimensions) == 0 && len(q.Measures) == 0 {
		return "", nil, errors.New("no dimensions or measures specified")
	}
	filterCount := 0
	for _, f := range q.Measures {
		if f.Filter != nil {
			filterCount++
		}
	}
	if filterCount != 0 && len(q.Measures) > 1 {
		return "", nil, errors.New("multiple measures with filter")
	}
	if filterCount == 1 && len(q.PivotOn) > 0 {
		return "", nil, errors.New("measure filter for pivot-on")
	}

	cols := q.cols()
	selectCols := make([]string, 0, cols)

	groupCols := make([]string, 0, len(q.Dimensions))
	unnestClauses := make([]string, 0)
	var selectArgs []any
	for _, d := range q.Dimensions {
		// Handle regular dimensions
		if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			dim, err := metricsViewDimension(mv, d.Name)
			if err != nil {
				return "", nil, err
			}
			dimSel, unnestClause := dialect.DimensionSelect(mv.Database, mv.DatabaseSchema, mv.Table, dim)
			selectCols = append(selectCols, dimSel)
			if unnestClause != "" {
				unnestClauses = append(unnestClauses, unnestClause)
			}
			groupCols = append(groupCols, fmt.Sprintf("%d", len(selectCols)))
			continue
		}

		// Handle time dimension
		expr, exprArgs, err := q.buildTimestampExpr(mv, d, dialect)
		if err != nil {
			return "", nil, err
		}
		alias := safeName(d.Name)
		if d.Alias != "" {
			alias = safeName(d.Alias)
		}
		selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, alias))
		// Using expr was causing issues with query arg expansion in duckdb.
		// Using column name is not possible either since it will take the original column name instead of the aliased column name
		// But using numbered group we can exactly target the correct selected column.
		// Note that the non-timestamp columns also use the numbered group-by for constancy.
		groupCols = append(groupCols, fmt.Sprintf("%d", len(selectCols)))
		selectArgs = append(selectArgs, exprArgs...)
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
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", "COUNT(*)", sn))
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			if len(m.BuiltinMeasureArgs) != 1 {
				return "", nil, fmt.Errorf("builtin measure '%s' expects 1 argument", m.BuiltinMeasure.String())
			}
			arg := m.BuiltinMeasureArgs[0].GetStringValue()
			if arg == "" {
				return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", fmt.Sprintf("COUNT(DISTINCT %s)", safeName(arg)), sn))
		default:
			return "", nil, fmt.Errorf("unknown builtin measure '%d'", m.BuiltinMeasure)
		}
	}

	groupClause := ""
	if len(groupCols) > 0 {
		groupClause = "GROUP BY " + strings.Join(groupCols, ", ")
	}

	whereClause := ""
	var whereArgs []any
	if mv.TimeDimension != "" {
		timeCol := safeName(mv.TimeDimension)
		if dialect == drivers.DialectDuckDB {
			timeCol = fmt.Sprintf("%s::TIMESTAMP", timeCol)
		}
		clause, err := timeRangeClause(q.TimeRange, mv, timeCol, &whereArgs)
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
		if strings.TrimSpace(clause) != "" {
			whereClause += fmt.Sprintf(" AND (%s)", clause)
		}
		whereArgs = append(whereArgs, clauseArgs...)
	}

	if policy != nil && policy.RowFilter != "" {
		whereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	if whereClause != "" {
		whereClause = "WHERE 1=1" + whereClause
	}

	havingClause := ""
	var havingClauseArgs []any
	if q.Having != nil {
		var err error
		havingClause, havingClauseArgs, err = buildExpression(mv, q.Having, nil, dialect)
		if err != nil {
			return "", nil, err
		}

		if strings.TrimSpace(havingClause) != "" {
			havingClause = "HAVING " + havingClause
		}
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
		limit := *q.Limit
		if limit == 0 {
			limit = 100
		}
		limitClause = fmt.Sprintf("LIMIT %d", limit)
	}

	var args []any
	args = append(args, selectArgs...)
	args = append(args, whereArgs...)
	args = append(args, havingClauseArgs...)

	var sql string
	if len(q.PivotOn) > 0 {
		l := int(pivotCellLimit) / q.cols()
		limitClause = fmt.Sprintf("LIMIT %d", l+1)

		if q.Offset != 0 {
			return "", nil, fmt.Errorf("offset not supported for pivot queries")
		}

		// SELECT m1, m2, d1, d2 FROM t, LATERAL UNNEST(t.d1) tbl(unnested_d1_) WHERE d1 = 'a' GROUP BY d1, d2
		sql = fmt.Sprintf("SELECT %[1]s FROM %[2]s %[3]s %[4]s %[5]s %[6]s %[7]s %[8]s",
			strings.Join(selectCols, ", "),      // 1
			escapeMetricsViewTable(dialect, mv), // 2
			strings.Join(unnestClauses, ""),     // 3
			whereClause,                         // 4
			groupClause,                         // 5
			havingClause,                        // 6
			orderClause,                         // 7
			limitClause,                         // 8
		)
	} else {
		if filterCount == 1 {
			return q.buildMeasureFilterSQL(mv, unnestClauses, selectCols, limitClause, orderClause, havingClause, whereClause, groupClause, args, selectArgs, whereArgs, havingClauseArgs, dialect)
		}
		sql = fmt.Sprintf("SELECT %[1]s FROM %[2]s %[3]s %[4]s %[5]s %[6]s %[7]s %[8]s OFFSET %[9]d",
			strings.Join(selectCols, ", "),      // 1
			escapeMetricsViewTable(dialect, mv), // 2
			strings.Join(unnestClauses, ""),     // 3
			whereClause,                         // 3
			groupClause,                         // 4
			havingClause,                        // 5
			orderClause,                         // 6
			limitClause,                         // 7
			q.Offset,                            // 8
		)
	}

	return sql, args, nil
}

func (q *MetricsViewAggregation) buildMetricsComparisonAggregationSQL(ctx context.Context, olap drivers.OLAPStore, priority int, mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity, export bool) (string, []any, error) {
	if len(q.Dimensions) == 0 && len(q.Measures) == 0 {
		return "", nil, errors.New("no dimensions or measures specified")
	}

	cols := q.cols()
	selectCols := make([]string, 0, cols)
	var comparisonSelectCols []string

	groupCols := make([]string, 0, len(q.Dimensions))
	finalDims := make([]string, 0, len(q.Dimensions))
	joinCols := make([]string, 0, len(q.Dimensions))

	unnestClauses := make([]string, 0)
	var selectArgs []any

	// for _, m := range q.Measures {
	// q.ComparisonMeasures = append(q.ComparisonMeasures, m.Name)
	// }
	q.calculateMeasuresMeta()

	colMap := make(map[string]int, q.cols())
	onlyDims := make([]string, len(q.Dimensions))
	for _, d := range q.Dimensions {
		// Handle regular dimensions
		if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			dim, err := metricsViewDimension(mv, d.Name)
			if err != nil {
				return "", nil, err
			}
			dimSel, unnestClause := dialect.DimensionSelect(mv.Database, mv.DatabaseSchema, mv.Table, dim)
			selectCols = append(selectCols, dimSel)
			onlyDims = append(onlyDims, dimSel)
			comparisonSelectCols = append(comparisonSelectCols, dimSel)
			finalDims = append(finalDims, fmt.Sprintf("COALESCE(base.%[1]s,comparison.%[1]s) as %[1]s", dimSel))
			if unnestClause != "" {
				unnestClauses = append(unnestClauses, unnestClause)
			}
			groupCols = append(groupCols, fmt.Sprintf("%d", len(selectCols)-1))
			colMap[dimSel] = len(selectCols) - 1
			var joinClause string
			if dialect == drivers.DialectClickHouse {
				joinClause = fmt.Sprintf("isNotDistinctFrom(base.%[1]s, comparison.%[1]s)", dimSel)
			} else {
				joinClause = fmt.Sprintf("base.%[1]s = comparison.%[1]s OR (base.%[1]s is null and comparison.%[1]s is null)", dimSel)
			}
			joinCols = append(joinCols, joinClause)
			continue
		}

		// Handle time dimension
		expr, exprArgs, err := q.buildTimestampExpr(mv, d, dialect)
		if err != nil {
			return "", nil, err
		}
		alias := safeName(d.Name)
		if d.Alias != "" {
			alias = safeName(d.Alias)
		}
		timeDimClause := fmt.Sprintf("%s as %s", expr, alias)

		selectCols = append(selectCols, timeDimClause)
		onlyDims = append(onlyDims, timeDimClause)
		colMap[alias] = len(selectCols) - 1
		comparisonSelectCols = append(comparisonSelectCols, timeDimClause)
		finalDims = append(finalDims, fmt.Sprintf("COALESCE(base.%[1]s,comparison.%[1]s) as %[1]s", alias))

		// Using expr was causing issues with query arg expansion in duckdb.
		// Using column name is not possible either since it will take the original column name instead of the aliased column name
		// But using numbered group we can exactly target the correct selected column.
		// Note that the non-timestamp columns also use the numbered group-by for constancy.
		groupCols = append(groupCols, fmt.Sprintf("%d", len(selectCols)))
		var joinClause string
		if dialect == drivers.DialectClickHouse {
			joinClause = fmt.Sprintf("isNotDistinctFrom(base.%[1]s, comparison.%[1]s)", alias)
		} else {
			joinClause = fmt.Sprintf("base.%[1]s = comparison.%[1]s OR (base.%[1]s is null and comparison.%[1]s is null)", alias)
		}
		joinCols = append(joinCols, joinClause)

		selectArgs = append(selectArgs, exprArgs...)
	}

	// for _, m := range q.Measures {
	// 	sn := safeName(m.Name)
	// 	switch m.BuiltinMeasure {
	// 	case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED:
	// 		expr, err := metricsViewMeasureExpression(mv, m.Name)
	// 		if err != nil {
	// 			return "", nil, err
	// 		}

	// 		selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, sn))
	// 	case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
	// 		selectCols = append(selectCols, fmt.Sprintf("%s as %s", "COUNT(*)", sn))
	// 	case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
	// 		if len(m.BuiltinMeasureArgs) != 1 {
	// 			return "", nil, fmt.Errorf("builtin measure '%s' expects 1 argument", m.BuiltinMeasure.String())
	// 		}
	// 		arg := m.BuiltinMeasureArgs[0].GetStringValue()
	// 		if arg == "" {
	// 			return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
	// 		}
	// 		selectCols = append(selectCols, fmt.Sprintf("%s as %s", fmt.Sprintf("COUNT(DISTINCT %s)", safeName(arg)), sn))
	// 	default:
	// 		return "", nil, fmt.Errorf("unknown builtin measure '%d'", m.BuiltinMeasure)
	// 	}
	// }

	groupClause := ""
	if len(groupCols) > 0 {
		groupClause = strings.Join(groupCols, ", ")
	}

	// dim, err := metricsViewDimension(mv, q.DimensionName)
	// if err != nil {
	// 	return "", nil, err
	// }

	// colName := safeName(dim.Name)

	labelMap := make(map[string]string, len(mv.Measures))
	for _, m := range mv.Measures {
		labelMap[m.Name] = m.Name
		if m.Label != "" {
			labelMap[m.Name] = m.Label
		}
	}

	// dimSel, unnestClause := dialect.DimensionSelect(mv.Database, mv.DatabaseSchema, mv.Table, dim)
	// selectCols = append(selectCols, dimSel)
	// comparisonSelectCols = append(comparisonSelectCols, dimSel)

	for _, m := range q.Measures {
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED:
			expr, err := metricsViewMeasureExpression(mv, m.Name)
			if err != nil {
				return "", nil, err
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
			}
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			selectCols = append(selectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
			}
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			if len(m.BuiltinMeasureArgs) != 1 {
				return "", nil, fmt.Errorf("builtin measure '%s' expects 1 argument", m.BuiltinMeasure.String())
			}
			arg := m.BuiltinMeasureArgs[0].GetStringValue()
			if arg == "" {
				return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
			}
			selectCols = append(selectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
			}
		default:
			return "", nil, fmt.Errorf("unknown builtin measure '%d'", m.BuiltinMeasure)
		}
	}

	var finalSelectCols []string
	var labelCols []string
	for _, m := range q.Measures {
		var columnsTuple string
		var labelTuple string
		if dialect != drivers.DialectDruid {
			if q.measuresMeta[m.Name].expand {
				columnsTuple = fmt.Sprintf(
					"base.%[1]s AS %[1]s, comparison.%[1]s AS %[2]s, base.%[1]s - comparison.%[1]s AS %[3]s, (base.%[1]s - comparison.%[1]s)/comparison.%[1]s::DOUBLE AS %[4]s",
					safeName(m.Name),
					safeName(m.Name+"__previous"),
					safeName(m.Name+"__delta_abs"),
					safeName(m.Name+"__delta_rel"),
				)
				labelTuple = fmt.Sprintf(
					"base.%[1]s AS %[5]s, comparison.%[1]s AS %[2]s, base.%[1]s - comparison.%[1]s AS %[3]s, (base.%[1]s - comparison.%[1]s)/comparison.%[1]s::DOUBLE AS %[4]s",
					safeName(m.Name),
					safeName(labelMap[m.Name]+" (prev)"),
					safeName(labelMap[m.Name]+" (Δ)"),
					safeName(labelMap[m.Name]+" (Δ%)"),
					safeName(labelMap[m.Name]),
				)
			} else {
				columnsTuple = fmt.Sprintf("base.%[1]s AS %[1]s", safeName(m.Name))
				labelTuple = fmt.Sprintf("base.%[1]s AS %[2]s", safeName(m.Name), safeName(labelMap[m.Name]))
			}
		} else {
			if q.measuresMeta[m.Name].expand {
				columnsTuple = fmt.Sprintf(
					"ANY_VALUE(base.%[1]s) AS %[1]s, ANY_VALUE(comparison.%[1]s) AS %[2]s, ANY_VALUE(base.%[1]s - comparison.%[1]s) AS %[3]s, ANY_VALUE(SAFE_DIVIDE(base.%[1]s - comparison.%[1]s, CAST(comparison.%[1]s AS DOUBLE))) AS %[4]s",
					safeName(m.Name),
					safeName(m.Name+"__previous"),
					safeName(m.Name+"__delta_abs"),
					safeName(m.Name+"__delta_rel"),
				)
				labelTuple = fmt.Sprintf(
					"ANY_VALUE(base.%[1]s) AS %[2]s, ANY_VALUE(comparison.%[1]s) AS %[3]s, ANY_VALUE(base.%[1]s - comparison.%[1]s) AS %[4]s, ANY_VALUE(SAFE_DIVIDE(base.%[1]s - comparison.%[1]s, CAST(comparison.%[1]s AS DOUBLE))) AS %[5]s",
					safeName(m.Name),
					safeName(labelMap[m.Name]),
					safeName(labelMap[m.Name]+" (prev)"),
					safeName(labelMap[m.Name]+" (Δ)"),
					safeName(labelMap[m.Name]+" (Δ%)"),
				)
			} else {
				columnsTuple = fmt.Sprintf("ANY_VALUE(base.%[1]s) AS %[1]s", safeName(m.Name))
				labelTuple = fmt.Sprintf("ANY_VALUE(base.%[1]s) AS %[2]s", safeName(m.Name), safeName(labelMap[m.Name]))
				// todo for right join
			}
		}
		finalSelectCols = append(
			finalSelectCols,
			columnsTuple,
		)
		labelCols = append(labelCols, labelTuple)
	}

	subSelectClause := strings.Join(selectCols, ", ")
	subComparisonSelectClause := strings.Join(comparisonSelectCols, ", ")
	finalSelectClause := strings.Join(finalSelectCols, ", ")
	labelSelectClause := strings.Join(labelCols, ", ")
	if export {
		finalSelectClause = labelSelectClause
	}

	baseWhereClause := "1=1"
	comparisonWhereClause := "1=1"

	var args []any
	if mv.TimeDimension == "" {
		return "", nil, fmt.Errorf("metrics view '%s' doesn't have time dimension", q.MetricsViewName)
	}

	td := safeName(mv.TimeDimension)
	if dialect == drivers.DialectDuckDB {
		td = fmt.Sprintf("%s::TIMESTAMP", td)
	}

	whereClause, whereClauseArgs, err := buildExpression(mv, q.Where, nil, dialect)
	if err != nil {
		return "", nil, err
	}

	trc, err := timeRangeClause(q.TimeRange, mv, td, &args)
	if err != nil {
		return "", nil, err
	}
	baseWhereClause += trc

	if whereClause != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", whereClause)
		args = append(args, whereClauseArgs...)
	}

	trc, err = timeRangeClause(q.ComparisonTimeRange, mv, td, &args)
	if err != nil {
		return "", nil, err
	}
	comparisonWhereClause += trc

	if whereClause != "" {
		comparisonWhereClause += fmt.Sprintf(" AND (%s)", whereClause)
		args = append(args, whereClauseArgs...)
	}

	if policy != nil && policy.RowFilter != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
		comparisonWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	havingClause := "1=1"
	if q.Having != nil {
		var havingClauseArgs []any
		havingClause, havingClauseArgs, err = buildExpression(mv, q.Having, nil, dialect)
		if err != nil {
			return "", nil, err
		}
		args = append(args, havingClauseArgs...)
	}

	var orderClauses []string
	var baseOrderClauses []string
	var comparisonOrderClauses []string

	for _, s := range q.Sort {
		if s.SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_UNSPECIFIED {
			clause := fmt.Sprintf("%d", colMap[s.Name])
			subQueryClause := clause
			var ending string
			if s.Desc {
				ending += " DESC"
			}
			if dialect == drivers.DialectDuckDB {
				ending += " NULLS LAST"
			}
			clause += ending
			subQueryClause += ending
			orderClauses = append(orderClauses, clause)
			baseOrderClauses = append(baseOrderClauses, subQueryClause)
			comparisonOrderClauses = append(comparisonOrderClauses, subQueryClause)
			continue
		}
		measureMeta, ok := q.measuresMeta[s.Name]
		if !ok {
			return "", nil, fmt.Errorf("metrics view '%s' doesn't contain '%s' sort column", q.MetricsViewName, s.Name)
		}

		var pos int
		switch s.SortType {
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE:
			pos = measureMeta.outerIndex
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE:
			pos = 1 + measureMeta.outerIndex
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA:
			pos = 2 + measureMeta.outerIndex
		case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA:
			pos = 3 + measureMeta.outerIndex
		default:
			return "", nil, fmt.Errorf("undefined sort type for measure %s", s.Name)
		}
		orderClause := fmt.Sprint(pos)
		baseOrderClause := fmt.Sprint(measureMeta.innerIndex)
		comparisonOrderClause := fmt.Sprint(measureMeta.comparisonInnerIndex)
		ending := ""
		if s.Desc {
			ending += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			ending += " NULLS LAST"
		}
		orderClause += ending
		baseOrderClause += ending
		orderClauses = append(orderClauses, orderClause)
		baseOrderClauses = append(baseOrderClauses, baseOrderClause)
		comparisonOrderClauses = append(comparisonOrderClauses, comparisonOrderClause)
	}

	orderByClause := ""
	baseSubQueryOrderByClause := ""
	comparisonSubQueryOrderByClause := ""

	if len(orderClauses) > 0 {
		orderByClause = "ORDER BY " + strings.Join(orderClauses, ", ")
		baseSubQueryOrderByClause = "ORDER BY " + strings.Join(baseOrderClauses, ", ")
		comparisonSubQueryOrderByClause = "ORDER BY " + strings.Join(comparisonOrderClauses, ", ")

	}

	limitClause := ""
	if q.Limit != nil && *q.Limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	baseLimitClause := ""
	comparisonLimitClause := ""

	joinType := "FULL"
	if !q.Exact {
		deltaComparison := q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA ||
			q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA

		limit := 0
		if q.Limit != nil {
			limit = int(*q.Limit)
		}
		approximationLimit := int(limit)
		if limit != 0 && limit < 100 && deltaComparison {
			approximationLimit = 100
		}

		if q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE || deltaComparison {
			joinType = "LEFT OUTER"
			baseLimitClause = baseSubQueryOrderByClause
			if approximationLimit > 0 {
				baseLimitClause += fmt.Sprintf(" LIMIT %d OFFSET %d", approximationLimit, q.Offset)
			}
		} else if q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE {
			joinType = "RIGHT OUTER"
			comparisonLimitClause = comparisonSubQueryOrderByClause
			if approximationLimit > 0 {
				comparisonLimitClause += fmt.Sprintf(" LIMIT %d OFFSET %d", approximationLimit, q.Offset)
			}
		}
	}

	/*
		Example of the SQL:

		SELECT COALESCE(base."domain", comparison."domain") AS "dom", base."measure_1", comparison."measure_1" AS "measure_1__previous", base."measure_1" - comparison."measure_1" AS "measure_1__delta_abs", (base."measure_1" - comparison."measure_1")/comparison."measure_1"::DOUBLE AS "measure_1__delta_rel" FROM
			(
				SELECT "domain", avg(bid_price) as "measure_1" FROM "ad_bids" WHERE 1=1 AND "timestamp" >= ? AND "timestamp" < ? GROUP BY "domain" ORDER BY true, 1 NULLS LAST LIMIT 100
			) base
		LEFT OUTER JOIN
			(
				SELECT "domain", avg(bid_price) as "measure_1" FROM "ad_bids" WHERE 1=1 AND "timestamp" >= ? AND "timestamp" < ? GROUP BY "domain"
			) comparison
		ON
				base."domain" = comparison."domain" OR (base."domain" is null and comparison."domain" is null)
		ORDER BY
			true, 1 NULLS LAST
		LIMIT 10
		OFFSET 0
	*/

	// finalDimName := safeName(q.DimensionName)
	// if export && dim.Label != "" {
	// 	finalDimName = safeName(dim.Label)
	// }
	var sql string
	if dialect != drivers.DialectDruid {
		// if dialect == drivers.DialectClickHouse {
		// 	joinOnClause = fmt.Sprintf("isNotDistinctFrom(base.%[1]s, comparison.%[1]s)", colName)
		// } else {
		// 	joinOnClause = fmt.Sprintf("base.%[1]s = comparison.%[1]s OR (base.%[1]s is null and comparison.%[1]s is null)", colName)
		// }
		// measure filter could include the base measure name.
		// this leads to ambiguity whether it applies to the base.measure ot comparison.measure.
		// to keep the clause builder consistent we add an outer query here.
		sql = fmt.Sprintf(`
				SELECT * from (
					SELECT %[2]s, %[9]s FROM 
						(
							SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s GROUP BY %[10]s %[12]s 
						) base
					%[11]s JOIN
						(
							SELECT %[16]s FROM %[3]s %[14]s WHERE %[5]s GROUP BY %[10]s %[13]s 
						) comparison
					ON
							%[17]s
					%[6]s
					%[7]s
					OFFSET
						%[8]d
				) WHERE 1=1 AND %[15]s 
			`,
			subSelectClause,                     // 1
			strings.Join(finalDims, ","),        // 2
			escapeMetricsViewTable(dialect, mv), // 3
			baseWhereClause,                     // 4
			comparisonWhereClause,               // 5
			orderByClause,                       // 6
			limitClause,                         // 7
			q.Offset,                            // 8
			finalSelectClause,                   // 9
			groupClause,                         // 10
			joinType,                            // 11
			baseLimitClause,                     // 12
			comparisonLimitClause,               // 13
			unnestClause,                        // 14
			havingClause,                        // 15
			subComparisonSelectClause,           // 16
			strings.Join(joinCols, " AND "),     // 17
			// groupClause,                         // 18
		)
	} else {
		/* else if dialect == drivers.DialectClickHouse {
				leftSubQueryAlias := "base"
		rightSubQueryAlias := "comparison"
		leftWhereClause := baseWhereClause
		rightWhereClause := comparisonWhereClause

		if q.Sort[0].SortType == runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE {
			leftSubQueryAlias = "comparison"
			rightSubQueryAlias = "base"
			leftWhereClause = comparisonWhereClause
			rightWhereClause = baseWhereClause
		}

		twiceTheLimitClause := ""
		if q.Exact {
			if q.Limit > 0 {
				twiceTheLimitClause = fmt.Sprintf(" LIMIT %d", q.Limit*2)
			} else if q.Limit == 0 {
				twiceTheLimitClause = fmt.Sprintf(" LIMIT %d", 100_000) // use Druid limit
			}
		}

		sql = fmt.Sprintf(`
				WITH %[11]s AS (
					SELECT %[1]s FROM %[3]s WHERE %[4]s GROUP BY %[18]s %[13]s %[10]s OFFSET %[8]d
				), %[12]s AS (
					SELECT %[17]s FROM %[3]s WHERE %[5]s AND %[16]s IN (SELECT * FROM %[11]s) GROUP BY %[18]s %[10]s
				)
				SELECT %[11]s.%[2]s AS %[14]s, %[9]s FROM %[11]s LEFT JOIN %[12]s ON base.%[2]s = comparison.%[2]s
				GROUP BY 1
				HAVING %[15]s
				%[6]s
				%[7]s
				OFFSET %[8]d
			`,
			subSelectClause,                     // 1
			strings.Join(finalDims, ","),        // 2
			escapeMetricsViewTable(dialect, mv), // 3
			leftWhereClause,                     // 4
			rightWhereClause,                    // 5
			orderByClause,                       // 6
			limitClause,                         // 7
			q.Offset,                            // 8
			finalSelectClause,                   // 9
			twiceTheLimitClause,                 // 10
			leftSubQueryAlias,                   // 11
			rightSubQueryAlias,                  // 12
			baseSubQueryOrderByClause,           // 13
			finalDimName,                        // 14
			havingClause,                        // 15
			dialect.MetricsViewDimensionExpression(dim), // 16
			subComparisonSelectClause,                   // 17
			groupClause,                                 // 18
		)
		*/
		/*
			Example of the SQL query with expression based dimension:

				WITH base AS (
				  SELECT (replace("channel", 'a', 'b')) as "b",
					count(*) as "total_records", sum("added") as "sum"
					FROM "wikipedia"
					WHERE 1=1 AND "__time" >= '2016-06-27T02:00:00.000Z' AND "__time" < '2016-06-27T03:00:00.000Z'
					GROUP BY 1 -- Druid does not support group by aliases
					ORDER BY 2 DESC
					LIMIT 500 OFFSET 0
				), comparison AS (
				  SELECT (replace("channel", 'a', 'b')) as "c",
					count(*) as "total_records"
					FROM "wikipedia"
					WHERE 1=1 AND "__time" >= '2016-06-27T01:00:00.000Z' AND "__time" < '2016-06-27T02:00:00.000Z'
					AND replace("channel", 'a', 'b') IN (SELECT "b" FROM base)
					GROUP BY 1 -- Druid does not support group by aliases
					LIMIT 500
				)
				SELECT base."b" AS "channel",
					ANY_VALUE(base."total_records") AS "total_records",
					ANY_VALUE(comparison."total_records") AS "total_records__previous",
					ANY_VALUE(base."total_records" - comparison."total_records") AS "total_records__delta_abs",
					ANY_VALUE(SAFE_DIVIDE(base."total_records" - comparison."total_records", CAST(comparison."total_records" AS DOUBLE))) AS "total_records__delta_rel",
					ANY_VALUE(base."sum") AS "sum",
				FROM base LEFT JOIN comparison ON base."b" = comparison."c"
				GROUP BY 1 -- Druid does not support group by aliases
				HAVING 1=1
				ORDER BY 2 DESC -- order by without group by is not supported by Druid
				 LIMIT 250
				OFFSET 0

			Apache Druid requires that one part of the JOIN fits in memory, that can be achieved by pushing down the limit clause to a subquery (works only if the sorting is based entirely on a single subquery result)
		*/
		if q.Sort[0].SortType != runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE {
			sql = fmt.Sprintf("SELECT %[1]s FROM %[2]s %[3]s WHERE %[4]s GROUP BY %[5]s %[6]s",
				onlyDims,                            // 1
				escapeMetricsViewTable(dialect, mv), // 2
				unnestClause,                        // 3
				baseWhereClause,                     // 4
				groupClause,                         // 5
				baseLimitClause,                     // 6
			)

			var druidArgs []any
			druidArgs = append(druidArgs, selectArgs...)
			druidArgs = append(druidArgs, whereClauseArgs...)

			_, result, err := olapQuery(ctx, olap, priority, sql, druidArgs)
			if err != nil {
				return "", nil, err
			}

			var innerWhereConditions []string
			for _, row := range result {
				var innerWhereConditions0 []string
				for k, field := range row.Fields {
					innerWhereConditions0 = append(innerWhereConditions0, fmt.Sprintf("%[1]s = '%[2]s'", k, field.AsInterface()))
				}
				innerWhereConditions = append(innerWhereConditions, strings.Join(innerWhereConditions0, " AND "))
			}

			sql = fmt.Sprintf(`
				SELECT * from (
					SELECT %[2]s, %[9]s FROM 
						(
							SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s GROUP BY %[10]s %[12]s 
						) base
					LEFT JOIN
						(
							SELECT %[16]s FROM %[3]s %[14]s WHERE %[5]s AND (%[18]s) GROUP BY %[10]s %[13]s 
						) comparison
					ON
							%[17]s
					GROUP BY %[10]s
					%[6]s
					%[7]s
					OFFSET
						%[8]d
				) WHERE 1=1 AND %[15]s 
			`,
				subSelectClause,                     // 1
				strings.Join(finalDims, ","),        // 2
				escapeMetricsViewTable(dialect, mv), // 3
				baseWhereClause,                     // 4
				comparisonWhereClause,               // 5
				orderByClause,                       // 6
				limitClause,                         // 7
				q.Offset,                            // 8
				finalSelectClause,                   // 9
				groupClause,                         // 10
				joinType,                            // 11
				baseLimitClause,                     // 12
				comparisonLimitClause,               // 13
				unnestClause,                        // 14
				havingClause,                        // 15
				subComparisonSelectClause,           // 16
				strings.Join(joinCols, " AND "),     // 17
				strings.Join(innerWhereConditions, " OR "), // 18
			)
		} else {
			limit := 0
			if q.Limit == nil {
				limit = 0
			}
			approximationLimit := int(limit)
			if limit != 0 && limit < 100 {
				approximationLimit = 100
			}

			comparisonLimitClause = comparisonSubQueryOrderByClause
			if approximationLimit > 0 {
				comparisonLimitClause += fmt.Sprintf(" LIMIT %d OFFSET %d", approximationLimit, q.Offset)
			}
			sql = fmt.Sprintf("SELECT %[1]s FROM %[2]s %[3]s WHERE %[4]s GROUP BY %[5]s %[6]s",
				onlyDims,                            // 1
				escapeMetricsViewTable(dialect, mv), // 2
				unnestClause,                        // 3
				comparisonWhereClause,               // 4
				groupClause,                         // 5
				comparisonLimitClause,               // 6
			)

			var druidArgs []any
			druidArgs = append(druidArgs, selectArgs...)
			druidArgs = append(druidArgs, whereClauseArgs...)

			_, result, err := olapQuery(ctx, olap, priority, sql, druidArgs)
			if err != nil {
				return "", nil, err
			}

			var innerWhereConditions []string
			for _, row := range result {
				var innerWhereConditions0 []string
				for k, field := range row.Fields {
					innerWhereConditions0 = append(innerWhereConditions0, fmt.Sprintf("%[1]s = '%[2]s'", k, field.AsInterface()))
				}
				innerWhereConditions = append(innerWhereConditions, strings.Join(innerWhereConditions0, " AND "))
			}

			sql = fmt.Sprintf(`
				SELECT * from (
					SELECT %[2]s, %[9]s FROM 
						(
							SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s AND (%[18]s) GROUP BY %[10]s %[12]s 
						) base
					LEFT JOIN
						(
							SELECT %[16]s FROM %[3]s %[14]s WHERE %[5]s GROUP BY %[10]s %[13]s 
						) comparison
					ON
							%[17]s
					GROUP BY %[10]s
					%[6]s
					%[7]s
					OFFSET
						%[8]d
				) WHERE 1=1 AND %[15]s 
			`,
				subSelectClause,                     // 1
				strings.Join(finalDims, ","),        // 2
				escapeMetricsViewTable(dialect, mv), // 3
				baseWhereClause,                     // 4
				comparisonWhereClause,               // 5
				orderByClause,                       // 6
				limitClause,                         // 7
				q.Offset,                            // 8
				finalSelectClause,                   // 9
				groupClause,                         // 10
				joinType,                            // 11
				baseLimitClause,                     // 12
				comparisonLimitClause,               // 13
				unnestClause,                        // 14
				havingClause,                        // 15
				subComparisonSelectClause,           // 16
				strings.Join(joinCols, " AND "),     // 17
				strings.Join(innerWhereConditions, " OR "), // 18
			)

		}
	}

	return sql, args, nil
}

func (q *MetricsViewAggregation) calculateMeasuresMeta() error {
	compare := !isTimeRangeNil(q.ComparisonTimeRange)

	if !compare && len(q.ComparisonMeasures) > 0 {
		return fmt.Errorf("comparison measures are provided but comparison time range is not")
	}

	if len(q.ComparisonMeasures) == 0 && compare {
		// backwards compatibility
		q.ComparisonMeasures = make([]string, len(q.Measures))
		for i, m := range q.Measures {
			q.ComparisonMeasures[i] = m.Name
		}
	}

	q.measuresMeta = make(map[string]metricsViewMeasureMeta, len(q.Measures))

	// compare m2 -> expand m2
	// order by d2, m1 base -> left join
	// base subquery: SELECT d1, d2, m1, m2 from t order by 2, 4
	// comp subquery: SELECT d1, d2, m2 from t

	// compare m2 -> expand m2
	// order by d2, m2 comp -> right join
	// base subquery: SELECT d1, d2, m1, m2 from t
	// comp subquery: SELECT d1, d2, m2 from t order by 2, 3

	// compare m2 -> expand m2
	// order by d2, m2 delta -> left join
	// base subquery: SELECT d1, d2, m1, m2 from t order by 2, 4
	// comp subquery: SELECT d1, d2, m2 from t
	inner := len(q.Dimensions) + 1
	outer := len(q.Dimensions) + 1
	comparisonInnerIndex := len(q.Dimensions) + 1
	for _, m := range q.Measures {
		expand := false
		for _, cm := range q.ComparisonMeasures {
			if m.Name == cm {
				expand = true
				break
			}
		}
		q.measuresMeta[m.Name] = metricsViewMeasureMeta{
			innerIndex:           inner,
			outerIndex:           outer,
			comparisonInnerIndex: comparisonInnerIndex,
			expand:               expand,
		}
		if expand {
			outer += 4
			comparisonInnerIndex++
		} else {
			outer++
		}
		inner++
	}

	// check all comparison measures are present in the measures list
	for _, cm := range q.ComparisonMeasures {
		if _, ok := q.measuresMeta[cm]; !ok {
			return fmt.Errorf("comparison measure '%s' is not present in the measures list", cm)
		}
	}

	err := validateSort(q.Sort, q.measuresMeta, compare)
	if err != nil {
		return err
	}

	// err = validateMeasureAliases(q.Aliases, q.measuresMeta, compare)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (q *MetricsViewAggregation) buildMeasureFilterSQL(mv *runtimev1.MetricsViewSpec, unnestClauses, selectCols []string, limitClause, orderClause, havingClause, whereClause, groupClause string, args, selectArgs, whereArgs, havingClauseArgs []any, dialect drivers.Dialect) (string, []any, error) {
	/*
		Example:
		SELECT d1, d2, d3, m1 FROM (
			SELECT t.d1, t.d2, t.d3, t2.m1 (
				SELECT t.d1, t.d2, t.d3, t2.m1 FROM (
					SELECT d1, d2, d3, m1 FROM t WHERE ...  GROUP BY d1, d2, d3 HAVING m1 > 10 ) t
				) t
				LEFT JOIN (
					SELECT d1, d2, d3, m1 FROM t WHERE ... AND (d4 = 'Safari') GROUP BY d1, d2, d3 HAVING m1 > 10
				)  t2 ON (COALESCE(t.d1, 'val') = COALESCE(t2.d1, 'val') and COALESCE(t.d2, 'val') = COALESCE(t2.d2, 'val') and ...
			)
		)
		WHERE m1 > 10 -- mimicing FILTER behavior for empty sets produced by HAVING
		GROUP BY d1, d2, d3 -- GROUP BY is required for Apache Druid
		ORDER BY ...
		LIMIT 100
		OFFSET 0

		This JOIN mirrors functionality of SELECT d1, d2, d3, m1 FILTER (WHERE d4 = 'Safari') FROM t WHERE... GROUP BY d1, d2, d3
		bacause FILTER cannot be applied for arbitrary measure, ie sum(a)/1000
	*/

	joinConditions := make([]string, 0, len(q.Dimensions))
	selfJoinCols := make([]string, 0, len(q.Dimensions)+1)
	finalProjection := make([]string, 0, len(q.Dimensions)+1)

	selfJoinTableAlias := tempName("self_join")
	nonNullValue := tempName("non_null")
	for _, d := range q.Dimensions {
		name := d.Name
		if d.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED && d.Alias != "" {
			name = d.Alias
		}
		joinConditions = append(joinConditions, fmt.Sprintf("COALESCE(%[1]s.%[2]s, '%[4]s') = COALESCE(%[3]s.%[2]s, '%[4]s')", escapeMetricsViewTable(dialect, mv), safeName(name), selfJoinTableAlias, nonNullValue))
		selfJoinCols = append(selfJoinCols, fmt.Sprintf("%s.%s", escapeMetricsViewTable(dialect, mv), safeName(name)))
		finalProjection = append(finalProjection, fmt.Sprintf("%[1]s", safeName(name)))
	}
	if dialect == drivers.DialectDruid { // Apache Druid cannot order without timestamp or GROUP BY
		finalProjection = append(finalProjection, fmt.Sprintf("ANY_VALUE(%[1]s) as %[1]s", safeName(q.Measures[0].Name)))
	} else {
		finalProjection = append(finalProjection, fmt.Sprintf("%[1]s", safeName(q.Measures[0].Name)))
	}
	selfJoinCols = append(selfJoinCols, fmt.Sprintf("%[1]s.%[2]s as %[3]s", selfJoinTableAlias, safeName(q.Measures[0].Name), safeName(q.Measures[0].Name)))

	measureExpression, measureWhereArgs, err := buildExpression(mv, q.Measures[0].Filter, nil, dialect)
	if err != nil {
		return "", nil, err
	}

	if whereClause == "" {
		whereClause = "WHERE 1=1"
	}

	measureWhereClause := whereClause + fmt.Sprintf(" AND (%s)", measureExpression)
	var extraWhere string
	var extraWhereArgs []any
	if q.Having != nil {
		extraWhere, extraWhereArgs, err = buildExpression(mv, q.Having, nil, dialect)
		if err != nil {
			return "", nil, err
		}
		extraWhere = "WHERE " + extraWhere
	}
	druidGroupBy := ""
	if dialect == drivers.DialectDruid {
		druidGroupBy = groupClause
	}

	sql := fmt.Sprintf(`
					SELECT %[16]s FROM (
						SELECT %[1]s FROM (
							SELECT %[10]s FROM %[2]s %[3]s %[4]s %[5]s %[6]s 
						) %[2]s 
						LEFT JOIN (
							SELECT %[10]s FROM %[2]s %[3]s %[9]s %[5]s %[6]s
						) %[7]s 
						ON (%[8]s)
					)
					%[14]s
					%[15]s
					%[13]s 
					%[11]s  
					OFFSET %[12]d
				`,
		strings.Join(selfJoinCols, ", "),      // 1
		escapeMetricsViewTable(dialect, mv),   // 2
		strings.Join(unnestClauses, ""),       // 3
		whereClause,                           // 4
		groupClause,                           // 5
		havingClause,                          // 6
		selfJoinTableAlias,                    // 7
		strings.Join(joinConditions, " AND "), // 8
		measureWhereClause,                    // 9
		strings.Join(selectCols, ", "),        // 10
		limitClause,                           // 11
		q.Offset,                              // 12
		orderClause,                           // 13
		extraWhere,                            // 14
		druidGroupBy,                          // 15
		strings.Join(finalProjection, ","),    // 16
	)

	args = args[:0]
	args = append(args, selectArgs...)
	args = append(args, whereArgs...)
	args = append(args, havingClauseArgs...)
	args = append(args, whereArgs...)
	args = append(args, measureWhereArgs...)
	args = append(args, havingClauseArgs...)
	args = append(args, extraWhereArgs...)

	return sql, args, nil
}

func (q *MetricsViewAggregation) buildTimestampExpr(mv *runtimev1.MetricsViewSpec, dim *runtimev1.MetricsViewAggregationDimension, dialect drivers.Dialect) (string, []any, error) {
	var col string
	if dim.Name == mv.TimeDimension {
		col = safeName(dim.Name)
		if dialect == drivers.DialectDuckDB {
			col = fmt.Sprintf("%s::TIMESTAMP", col)
		}
	} else {
		d, err := metricsViewDimension(mv, dim.Name)
		if err != nil {
			return "", nil, err
		}
		if d.Expression != "" {
			// TODO: we should add support for this in a future PR
			return "", nil, fmt.Errorf("expression dimension not supported as time column")
		}
		col = dialect.MetricsViewDimensionExpression(d)
	}

	switch dialect {
	case drivers.DialectDuckDB:
		if dim.TimeZone == "" || dim.TimeZone == "UTC" || dim.TimeZone == "Etc/UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)::TIMESTAMP", dialect.ConvertToDateTruncSpecifier(dim.TimeGrain), col), nil, nil
		}
		return fmt.Sprintf("timezone(?, date_trunc('%s', timezone(?, %s::TIMESTAMPTZ)))::TIMESTAMP", dialect.ConvertToDateTruncSpecifier(dim.TimeGrain), col), []any{dim.TimeZone, dim.TimeZone}, nil
	case drivers.DialectDruid:
		if dim.TimeZone == "" || dim.TimeZone == "UTC" || dim.TimeZone == "Etc/UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)", dialect.ConvertToDateTruncSpecifier(dim.TimeGrain), col), nil, nil
		}
		return fmt.Sprintf("time_floor(%s, '%s', null, CAST(? AS VARCHAR))", col, convertToDruidTimeFloorSpecifier(dim.TimeGrain)), []any{dim.TimeZone}, nil
	case drivers.DialectClickHouse:
		if dim.TimeZone == "" || dim.TimeZone == "UTC" || dim.TimeZone == "Etc/UTC" {
			return fmt.Sprintf("date_trunc('%s', %s)", dialect.ConvertToDateTruncSpecifier(dim.TimeGrain), col), nil, nil
		}
		return fmt.Sprintf("toTimezone(date_trunc('%s', toTimezone(%s::TIMESTAMP, ?)), ?)", dialect.ConvertToDateTruncSpecifier(dim.TimeGrain), col), []any{dim.TimeZone, dim.TimeZone}, nil
	default:
		return "", nil, fmt.Errorf("unsupported dialect %q", dialect)
	}
}
