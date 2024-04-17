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
	MetricsViewName    string                                       `json:"metrics_view,omitempty"`
	Dimensions         []*runtimev1.MetricsViewAggregationDimension `json:"dimensions,omitempty"`
	Measures           []*runtimev1.MetricsViewAggregationMeasure   `json:"measures,omitempty"`
	Sort               []*runtimev1.MetricsViewAggregationSort      `json:"sort,omitempty"`
	TimeRange          *runtimev1.TimeRange                         `json:"time_range,omitempty"`
	Where              *runtimev1.Expression                        `json:"where,omitempty"`
	Having             *runtimev1.Expression                        `json:"having,omitempty"`
	Filter             *runtimev1.MetricsViewFilter                 `json:"filter,omitempty"` // Backwards compatibility
	Priority           int32                                        `json:"priority,omitempty"`
	Limit              *int64                                       `json:"limit,omitempty"`
	Offset             int64                                        `json:"offset,omitempty"`
	PivotOn            []string                                     `json:"pivot_on,omitempty"`
	SecurityAttributes map[string]any                               `json:"security_attributes,omitempty"`

	Exporting bool

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

	// Build query
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

	if olap.Dialect() == drivers.DialectDuckDB {
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
					return fmt.Errorf("duckdb append failed: %v")
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
			dimSel, unnestClause := dimensionSelect(mv.Database, mv.DatabaseSchema, mv.Table, dim, dialect)
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
		if filterCount == 1 {
			return q.buildMeasureFilterSQL(mv, unnestClauses, selectCols, limitClause, orderClause, havingClause, whereClause, groupClause, args, selectArgs, whereArgs, havingClauseArgs, dialect)
		}
		sql = fmt.Sprintf("SELECT %s FROM %s %s %s %s %s %s %s OFFSET %d",
			strings.Join(selectCols, ", "),
			escapeMetricsViewTable(dialect, mv),
			strings.Join(unnestClauses, ""),
			whereClause,
			groupClause,
			havingClause,
			orderClause,
			limitClause,
			q.Offset,
		)
	}

	return sql, args, nil
}

func (q *MetricsViewAggregation) buildMeasureFilterSQL(mv *runtimev1.MetricsViewSpec, unnestClauses, selectCols []string, limitClause, orderClause, havingClause, whereClause, groupClause string, args, selectArgs, whereArgs, havingClauseArgs []any, dialect drivers.Dialect) (string, []any, error) {
	joinConditions := make([]string, 0, len(q.Dimensions))
	selfJoinCols := make([]string, 0, len(q.Dimensions)+1)
	finalProjection := make([]string, 0, len(q.Dimensions)+1)

	selfJoinTableAlias := tempName("self_join")
	nonNullValue := tempName("non_null")
	for _, d := range q.Dimensions {
		joinConditions = append(joinConditions, fmt.Sprintf("COALESCE(%[1]s.%[2]s, '%[4]s') = COALESCE(%[3]s.%[2]s, '%[4]s')", escapeMetricsViewTable(dialect, mv), safeName(d.Name), selfJoinTableAlias, nonNullValue))
		selfJoinCols = append(selfJoinCols, fmt.Sprintf("%s.%s", escapeMetricsViewTable(dialect, mv), safeName(d.Name)))
		finalProjection = append(finalProjection, fmt.Sprintf("%[1]s", safeName(d.Name)))
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
		col = metricsViewDimensionExpression(d)
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
