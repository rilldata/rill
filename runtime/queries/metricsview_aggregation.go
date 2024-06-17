package queries

import (
	"context"
	databasesql "database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	duckdbolap "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewAggregation struct {
	MetricsViewName     string                                         `json:"metrics_view,omitempty"`
	Dimensions          []*runtimev1.MetricsViewAggregationDimension   `json:"dimensions,omitempty"`
	Measures            []*runtimev1.MetricsViewAggregationMeasure     `json:"measures,omitempty"`
	Sort                []*runtimev1.MetricsViewAggregationSort        `json:"sort,omitempty"`
	TimeRange           *runtimev1.TimeRange                           `json:"time_range,omitempty"`
	ComparisonTimeRange *runtimev1.TimeRange                           `json:"comparison_time_range,omitempty"`
	Where               *runtimev1.Expression                          `json:"where,omitempty"`
	Having              *runtimev1.Expression                          `json:"having,omitempty"`
	Filter              *runtimev1.MetricsViewFilter                   `json:"filter,omitempty"` // Backwards compatibility
	Priority            int32                                          `json:"priority,omitempty"`
	Limit               *int64                                         `json:"limit,omitempty"`
	Offset              int64                                          `json:"offset,omitempty"`
	PivotOn             []string                                       `json:"pivot_on,omitempty"`
	SecurityAttributes  map[string]any                                 `json:"security_attributes,omitempty"`
	Aliases             []*runtimev1.MetricsViewComparisonMeasureAlias `json:"aliases,omitempty"`
	Exact               bool                                           `json:"exact,omitempty"`

	Exporting    bool                                      `json:"-"`
	Result       *runtimev1.MetricsViewAggregationResponse `json:"-"`
	measuresMeta map[string]metricsViewMeasureMeta         `json:"-"`
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

	// Attempt to route to metricsview executor
	qry, ok, err := q.rewriteToMetricsViewQuery(mv)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}
	if ok {
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

		q.Result = &runtimev1.MetricsViewAggregationResponse{
			Schema: res.Schema,
			Data:   data,
		}
		return nil
	}
	// Falling back to the old implementation

	cfg, err := rt.InstanceConfig(ctx, instanceID)
	if err != nil {
		return err
	}

	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB && olap.Dialect() != drivers.DialectDruid && olap.Dialect() != drivers.DialectClickHouse && olap.Dialect() != drivers.DialectPinot {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	if mv.TimeDimension == "" && !isTimeRangeNil(q.TimeRange) {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", mv)
	}

	if !isTimeRangeNil(q.TimeRange) {
		start, end, err := ResolveTimeRange(q.TimeRange, mv)
		if err != nil {
			return err
		}
		q.TimeRange = &runtimev1.TimeRange{
			Start: timestamppb.New(start.In(time.UTC)),
			End:   timestamppb.New(end.In(time.UTC)),
		}
	}

	// backwards compatibility
	if q.Filter != nil {
		if q.Where != nil {
			return fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	if q.ComparisonTimeRange != nil {
		if isTimeRangeNil(q.ComparisonTimeRange) || isTimeRangeNil(q.TimeRange) {
			return fmt.Errorf("Undefined time range boundaries")
		}

		start, end, err := ResolveTimeRange(q.TimeRange, mv)
		if err != nil {
			return err
		}

		q.TimeRange = &runtimev1.TimeRange{
			Start: timestamppb.New(start.In(time.UTC)),
			End:   timestamppb.New(end.In(time.UTC)),
		}

		start, end, err = ResolveTimeRange(q.ComparisonTimeRange, mv)
		if err != nil {
			return err
		}
		q.ComparisonTimeRange = &runtimev1.TimeRange{
			Start: timestamppb.New(start.In(time.UTC)),
			End:   timestamppb.New(end.In(time.UTC)),
		}

		filterCount := 0
		for _, f := range q.Measures {
			if f.Filter != nil {
				filterCount++
			}
		}

		if filterCount > 1 {
			return fmt.Errorf("multiple measures with filter")
		}

		if filterCount == 1 {
			return q.executeComparisonWithMeasureFilter(ctx, olap, priority, mv, olap.Dialect(), security)
		}

		return q.executeComparisonAggregation(ctx, olap, priority, mv, olap.Dialect(), security, cfg)
	}

	for _, m := range q.Measures {
		switch m.GetCompute().(type) {
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta, *runtimev1.MetricsViewAggregationMeasure_ComparisonValue, *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
			return fmt.Errorf("comaparison measures without comparison time range")
		}
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

	return q.pivotDruid(ctx, rows, mv, cfg.PivotCellLimit, func(temporaryTableName string, mv *runtimev1.MetricsViewSpec) string {
		return q.createPivotSQL(temporaryTableName, mv)
	})
}

func (q *MetricsViewAggregation) executeComparisonWithMeasureFilter(ctx context.Context, olap drivers.OLAPStore, priority int, mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, security *runtime.ResolvedMetricsViewSecurity) error {
	sqlString, args, err := q.buildMeasureFilterComparisonAggregationSQL(mv, dialect, security, false)
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

	return fmt.Errorf("pivot unsupported for the measure filter")
}

func (q *MetricsViewAggregation) executeComparisonAggregation(ctx context.Context, olap drivers.OLAPStore, priority int, mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, security *runtime.ResolvedMetricsViewSecurity, cfg drivers.InstanceConfig) error {
	sqlString, args, err := q.buildMetricsComparisonAggregationSQL(ctx, olap, priority, mv, dialect, security, false)
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

			schema, data, err := olapQuery(ctx, olap, int(q.Priority), q.createComparisonPivotSQL(temporaryTableName, mv), nil)
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

	return q.pivotDruid(ctx, rows, mv, cfg.PivotCellLimit, func(temporaryTableName string, mv *runtimev1.MetricsViewSpec) string {
		return q.createComparisonPivotSQL(temporaryTableName, mv)
	})
}

func (q *MetricsViewAggregation) pivotDruid(ctx context.Context, rows *drivers.Result, mv *runtimev1.MetricsViewSpec, pivotCellLimit int64, pivotSQL func(temporaryTableName string, mv *runtimev1.MetricsViewSpec) string) error {
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
		pivotRows, err := pivotDB.QueryxContext(ctx, pivotSQL(temporaryTableName, mv))
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
	sql := fmt.Sprintf("PIVOT (SELECT %[7]s FROM %[1]s) ON %[2]s USING %[3]s %[4]s %[5]s OFFSET %[6]d",
		temporaryTableName,              // 1
		strings.Join(pivots, ", "),      // 2
		strings.Join(measureCols, ", "), // 3
		orderClause,                     // 4
		limitClause,                     // 5
		q.Offset,                        // 6
		selectList,                      // 7
	)
	return sql
}

func (q *MetricsViewAggregation) createComparisonPivotSQL(temporaryTableName string, mv *runtimev1.MetricsViewSpec) string {
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
					expr = fmt.Sprintf("lower(%s)", safeName(d.Name)) // workaround for DuckDB PIVOT issue
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
		alias := m.Name
		if q.Exporting && aliasesMap[m.Name] != "" {
			alias = aliasesMap[m.Name]
		}
		qalias := safeName(alias)
		measureCols = append(measureCols, fmt.Sprintf("LAST(%s) as %s", qalias, qalias))
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

	selectList := "*"
	if q.Exporting {
		selectList = strings.Join(selectCols, ",")
	}
	sql := fmt.Sprintf("PIVOT (SELECT %[7]s FROM %[1]s) ON %[2]s USING %[3]s %[4]s %[5]s OFFSET %[6]d",
		temporaryTableName,              // 1
		strings.Join(pivots, ", "),      // 2
		strings.Join(measureCols, ", "), // 3
		orderClause,                     // 4
		limitClause,                     // 5
		q.Offset,                        // 6
		selectList,                      // 7
	)
	return sql
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

	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
	if !isTimeRangeNil(q.TimeRange) || q.Where != nil || q.Having != nil {
		filename += "_filtered"
	}

	// Resolve metrics view
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityAttributes, q.Dimensions, q.Measures)
	if err != nil {
		return err
	}

	// Attempt to route to metricsview executor
	qry, ok, err := q.rewriteToMetricsViewQuery(mv)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}
	if ok {
		e, err := metricsview.NewExecutor(ctx, rt, instanceID, mv, security, opts.Priority)
		if err != nil {
			return err
		}
		defer e.Close()

		var format string
		switch opts.Format {
		case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
			format = "csv"
		case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
			format = "xlsx"
		case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
			format = "parquet"
		default:
			return fmt.Errorf("unsupported format: %s", opts.Format.String())
		}

		path, err := e.Export(ctx, qry, nil, format)
		if err != nil {
			return err
		}
		defer func() { _ = os.Remove(path) }()

		err = opts.PreWriteHook(filename)
		if err != nil {
			return err
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
	// Falling back to the old implementation

	err = q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("timeout exceeded")
		}
		return err
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

	whereBuilder := &ExpressionBuilder{
		mv:      mv,
		dialect: dialect,
	}
	if q.Where != nil {
		clause, clauseArgs, err := whereBuilder.buildExpression(q.Where)
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

	var havingClause, extraWhereClause string
	var havingClauseArgs, extraWhereClauseArgs []any
	if q.Having != nil {
		var err error
		// HAVING expression is converted to WHERE expression here
		extraWhereClause, extraWhereClauseArgs, err = whereBuilder.buildExpression(q.Having)
		if err != nil {
			return "", nil, err
		}

		havingBuilder := &ExpressionBuilder{
			mv:      mv,
			dialect: dialect,
			having:  true,
		}
		havingClause, havingClauseArgs, err = havingBuilder.buildExpression(q.Having)
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
			return q.buildMeasureFilterSQL(mv, unnestClauses, selectCols, limitClause, orderClause, havingClause, whereClause, groupClause, args, selectArgs, whereArgs, havingClauseArgs, extraWhereClause, extraWhereClauseArgs, dialect)
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

func originalName(m *runtimev1.MetricsViewAggregationMeasure) string {
	switch t := m.Compute.(type) {
	case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
		return t.ComparisonRatio.Measure
	case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
		return t.ComparisonDelta.Measure
	case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
		return t.ComparisonValue.Measure
	case *runtimev1.MetricsViewAggregationMeasure_Count:
		return m.Name
	case *runtimev1.MetricsViewAggregationMeasure_CountDistinct:
		return m.Name
	default:
		return m.Name
	}
}

func (q *MetricsViewAggregation) buildMeasureFilterComparisonAggregationSQL(mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity, export bool) (string, []any, error) {
	originals := make(map[string]bool, len(q.Measures))
	for _, m := range q.Measures {
		if m.Compute != nil {
			originals[originalName(m)] = true
		}
	}
	if len(originals) > 1 {
		return "", nil, errors.New("more than one original measures specified")
	}

	if len(q.Dimensions) == 0 && len(q.Measures) == 0 {
		return "", nil, errors.New("no dimensions or measures specified")
	}

	if len(q.Measures) == 1 {
		m := &runtimev1.MetricsViewAggregationMeasure{}
		switch t := q.Measures[0].Compute.(type) {
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
			m.Name = t.ComparisonValue.Measure
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
			m.Name = t.ComparisonDelta.Measure
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
			m.Name = t.ComparisonRatio.Measure
		}
		q.Measures = append(q.Measures, m)
	}

	dimByName := make(map[string]*runtimev1.MetricsViewAggregationDimension, len(mv.Dimensions))
	measuresByFinalName := make(map[string]*runtimev1.MetricsViewAggregationMeasure, len(q.Measures))
	for _, d := range q.Dimensions {
		if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED || d.Alias == "" {
			dimByName[d.Name] = d
		} else {
			dimByName[d.Alias] = d
		}
	}
	for _, m := range q.Measures {
		measuresByFinalName[m.Name] = m
	}

	cols := q.cols()
	selectCols := make([]string, 0, cols+1)
	var comparisonSelectCols []string

	finalDims := make([]string, 0, len(q.Dimensions))
	joinConditions := make([]string, 0, len(q.Dimensions))

	unnestClauses := make([]string, 0)
	var selectArgs []any

	err := q.calculateMeasuresMeta()
	if err != nil {
		return "", nil, err
	}

	// Required for t_offset, ie
	// SELECT t_offset, d1, d2, t1, t2, m1, m2
	minTimeGrain := runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
	for _, d := range q.Dimensions {
		if d.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED && d.GetName() == mv.TimeDimension {
			if minTimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED || d.TimeGrain < minTimeGrain {
				minTimeGrain = d.TimeGrain
			}
		}
	}

	// it's required for joining the base and comparison tables
	timeOffsetExpression, err := q.buildOffsetExpression(mv.TimeDimension, minTimeGrain, dialect)
	if err != nil {
		return "", nil, err
	}

	colMap := make(map[string]int, q.cols())

	selectCols = append(selectCols, timeOffsetExpression)
	comparisonSelectCols = append(comparisonSelectCols, timeOffsetExpression)

	joinConditions = append(joinConditions, "base.t_offset = comparison.t_offset")
	var finalComparisonTimeDims []string

	mvDimsByName := make(map[string]*runtimev1.MetricsViewSpec_DimensionV2, len(mv.Dimensions))
	finalDims = append(finalDims, "COALESCE(base.t_offset, comparison.t_offset) AS t_offset")
	var timeDims []string
	var subSelectDimAliases []string
	for _, d := range q.Dimensions {
		// Handle regular dimensions
		if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			dim, err := metricsViewDimension(mv, d.Name)
			if err != nil {
				return "", nil, err
			}
			mvDimsByName[d.Name] = dim
			expr, alias, unnestClause := dialect.DimensionSelectPair(mv.Database, mv.DatabaseSchema, mv.Table, dim)
			dimSel := fmt.Sprintf("%s AS %s", expr, alias)
			selectCols = append(selectCols, dimSel)
			subSelectDimAliases = append(subSelectDimAliases, alias)
			comparisonSelectCols = append(comparisonSelectCols, dimSel)
			finalDims = append(finalDims, fmt.Sprintf("COALESCE(base.%[1]s,comparison.%[1]s) as %[1]s", safeName(dim.Name)))
			if unnestClause != "" {
				unnestClauses = append(unnestClauses, unnestClause)
			}
			colMap[d.Name] = len(selectCols)
			var joinCondition string
			if dialect == drivers.DialectClickHouse {
				joinCondition = fmt.Sprintf("isNotDistinctFrom(base.%[1]s, comparison.%[1]s)", safeName(dim.Name))
			} else {
				joinCondition = fmt.Sprintf("base.%[1]s IS NOT DISTINCT FROM comparison.%[1]s", safeName(dim.Name))
			}
			joinConditions = append(joinConditions, joinCondition)
			continue
		}

		// Handle time dimension
		expr, exprArgs, err := q.buildTimestampExpr(mv, d, dialect)
		if err != nil {
			return "", nil, err
		}
		alias := d.Name
		if d.Alias != "" {
			alias = d.Alias
		}
		timeDimClause := fmt.Sprintf("%s as %s", expr, safeName(alias))

		timeDims = append(timeDims, safeName(alias))
		selectCols = append(selectCols, timeDimClause)
		subSelectDimAliases = append(subSelectDimAliases, safeName(alias))
		colMap[alias] = len(selectCols)
		comparisonSelectCols = append(comparisonSelectCols, timeDimClause)
		finalDims = append(finalDims, fmt.Sprintf("base.%[1]s as %[1]s", safeName(alias)))
		if dialect == drivers.DialectDruid {
			finalComparisonTimeDims = append(finalComparisonTimeDims, fmt.Sprintf("MILLIS_TO_TIMESTAMP(PARSE_LONG(ANY_VALUE(comparison.%[1]s))) as %[2]s", safeName(alias), safeName(alias+"__previous")))
		} else {
			finalComparisonTimeDims = append(finalComparisonTimeDims, fmt.Sprintf("comparison.%[1]s as %[2]s", safeName(alias), safeName(alias+"__previous")))
		}

		selectArgs = append(selectArgs, exprArgs...)
	}

	labelMap := make(map[string]string, len(mv.Measures))
	for _, m := range mv.Measures {
		labelMap[m.Name] = m.Name
		if m.Label != "" {
			labelMap[m.Name] = m.Label
		}
	}

	var subselectMeasureAliases []string
	var subselectComparisonAliases []string

	originalMeasure := OriginalColumnName(q.Measures[0])
	// collect subquery expressions
	for _, m := range q.Measures {
		switch m.Compute.(type) {
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue, *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta, *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
			// nothing
		case *runtimev1.MetricsViewAggregationMeasure_Count:
			selectCols = append(selectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
			subselectMeasureAliases = append(subselectMeasureAliases, safeName(m.Name))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
				subselectComparisonAliases = append(subselectComparisonAliases, safeName(m.Name))
			}
		case *runtimev1.MetricsViewAggregationMeasure_CountDistinct:
			arg := m.GetCountDistinct().GetDimension()
			if arg == "" {
				return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
			}
			selectCols = append(selectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
			subselectMeasureAliases = append(subselectMeasureAliases, safeName(m.Name))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
				subselectComparisonAliases = append(subselectComparisonAliases, safeName(m.Name))
			}
		default:
			expr, err := metricsViewMeasureExpression(mv, m.Name)
			if err != nil {
				return "", nil, err
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
			subselectMeasureAliases = append(subselectMeasureAliases, safeName(m.Name))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
				subselectComparisonAliases = append(subselectComparisonAliases, safeName(m.Name))
			}
		}
	}

	var finalMeasures []*FinalMeasure
	for _, m := range q.Measures {
		var subqueryName, finalName string
		prefix := ""

		var finalMeasure FinalMeasure
		switch m.Compute.(type) {
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
			subqueryName = m.GetComparisonRatio().Measure
			finalName = m.Name
			if dialect == drivers.DialectDruid {
				finalMeasure.expression = fmt.Sprintf("SAFE_DIVIDE(base.%[1]s - comparison.%[1]s, CAST(comparison.%[1]s AS DOUBLE))", safeName(subqueryName))
				finalMeasure.alias = safeName(finalName)
			} else {
				finalMeasure.expression = fmt.Sprintf(
					"(base.%[1]s - comparison.%[1]s)/comparison.%[1]s::DOUBLE",
					safeName(subqueryName))
				finalMeasure.alias = safeName(finalName)
			}
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
			subqueryName = m.GetComparisonDelta().Measure
			finalName = m.Name

			finalMeasure.alias = safeName(finalName)
			finalMeasure.expression = fmt.Sprintf( // non-virtial columns have a label
				"%[2]s(base.%[1]s - comparison.%[1]s)",
				safeName(subqueryName),
				prefix,
			)
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
			subqueryName = m.GetComparisonValue().Measure
			finalName = m.Name

			finalMeasure.alias = safeName(finalName)
			finalMeasure.expression = fmt.Sprintf( // non-virtial columns have a label
				"%[2]s(comparison.%[1]s)",
				safeName(subqueryName),
				prefix,
			)
		case *runtimev1.MetricsViewAggregationMeasure_Count, *runtimev1.MetricsViewAggregationMeasure_CountDistinct:
			subqueryName = m.Name
			finalMeasure.alias = safeName(subqueryName)
			finalMeasure.expression = fmt.Sprintf( // non-virtial columns have a label
				"%[2]s(base.%[1]s)",
				safeName(subqueryName),
				prefix,
			)
		default: // not a virtual (not a generated) column
			subqueryName = m.Name
			finalName = m.Name

			finalMeasure.expression = fmt.Sprintf( // non-virtial columns have a label
				"%[2]s(base.%[1]s)",
				safeName(subqueryName),
				prefix,
			)
			if export {
				finalMeasure.alias = safeName(labelMap[subqueryName])
			} else {
				finalMeasure.alias = safeName(finalName)
			}
		}
		finalMeasures = append(finalMeasures, &finalMeasure)
	}

	baseSelectClause := strings.Join(selectCols, ", ")
	comparisonSelectClause := strings.Join(comparisonSelectCols, ", ")

	baseWhereClause := "1=1"
	comparisonWhereClause := "1=1"

	if mv.TimeDimension == "" {
		return "", nil, fmt.Errorf("metrics view '%s' doesn't have time dimension", q.MetricsViewName)
	}

	td := safeName(mv.TimeDimension)
	if dialect == drivers.DialectDuckDB {
		td = fmt.Sprintf("%s::TIMESTAMP", td)
	}

	whereBuilder := &ExpressionBuilder{
		mv:       mv,
		dialect:  dialect,
		measures: q.Measures,
	}
	whereClause, whereClauseArgs, err := whereBuilder.buildExpression(q.Where)
	if err != nil {
		return "", nil, err
	}

	var baseTimeRangeArgs []any
	trc, err := timeRangeClause(q.TimeRange, mv, td, &baseTimeRangeArgs)
	if err != nil {
		return "", nil, err
	}
	baseWhereClause += trc

	if whereClause != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", whereClause)
	}

	var comparisonTimeRangeArgs []any
	trc, err = timeRangeClause(q.ComparisonTimeRange, mv, td, &comparisonTimeRangeArgs)
	if err != nil {
		return "", nil, err
	}
	comparisonWhereClause += trc

	if whereClause != "" {
		comparisonWhereClause += fmt.Sprintf(" AND (%s)", whereClause)
	}

	if policy != nil && policy.RowFilter != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
		comparisonWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	havingWhereClause := "1=1"
	var havingClauseArgs []any
	if q.Having != nil {
		havingBuilder := &ExpressionBuilder{
			mv:       mv,
			dialect:  dialect,
			measures: q.Measures,
		}
		havingWhereClause, havingClauseArgs, err = havingBuilder.buildExpression(q.Having)
		if err != nil {
			return "", nil, err
		}
	}

	/*
		Dimensions are referenced by index because time dimensions have an issue with aliases in DuckDB
		and other dimensions cannot be referenced by alias in Druid.
	*/
	var sortConstructs []*SortConstruct
	var outerSortConstructs []*SortConstruct
	for _, s := range q.Sort {
		dimClause := false
		var outerClause, subQueryClause string
		if dimByName[s.Name] != nil { // dimension
			outerClause = fmt.Sprintf("%d", colMap[s.Name]-1)
			subQueryClause = fmt.Sprintf("%d", colMap[s.Name])
			dimClause = true
		} else if measuresByFinalName[s.Name] != nil { // measure
			m := measuresByFinalName[s.Name]
			outerClause = safeName(s.Name)
			subQueryClause = OriginalColumnName(m)
		} else {
			return "", nil, fmt.Errorf("no selected dimension or measure '%s' found for sorting", s.Name)
		}

		var ending string
		if s.Desc {
			ending += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			ending += " NULLS LAST"
		}

		if !dimClause {
			sortConstructs = append(sortConstructs, &SortConstruct{
				expression: subQueryClause,
				ending:     ending,
			})
			outerSortConstructs = append(outerSortConstructs, &SortConstruct{
				expression: outerClause,
				ending:     ending,
			})
		} else {
			sortConstructs = append(sortConstructs, &SortConstruct{
				expression: subQueryClause,
				ending:     ending,
				dim:        true,
			})

			outerSortConstructs = append(outerSortConstructs, &SortConstruct{
				expression: outerClause,
				ending:     ending,
				dim:        true,
			})
		}
	}

	outerWhereClause := ""
	if q.Having != nil {
		outerWhereClause += " WHERE " + havingWhereClause
	}

	var args []any
	var sql string

	cw := tempName("casewhere")
	if dialect != drivers.DialectDruid {
		// Using expr was causing issues with query arg expansion in duckdb.
		// Using column name is not possible either since it will take the original column name instead of the aliased column name
		// But using numbered group we can exactly target the correct selected column.
		// Note that the non-timestamp columns also use the numbered group-by for constancy.

		// Inner grouping should include t_offset
		// SELECT t_offset, d1, d2, d3 ... GROUP BY 1, 2, 3, 4 ...
		innerGroupCols := make([]string, 0, len(q.Dimensions)+1)
		innerGroupCols = append(innerGroupCols, "1")
		for i := range q.Dimensions {
			innerGroupCols = append(innerGroupCols, fmt.Sprintf("%d", i+2))
		}

		outerDims := make([]string, 0, len(q.Dimensions))
		for _, d := range q.Dimensions {
			// Handle regular dimensions
			if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				dim := mvDimsByName[d.Name]
				outerDims = append(outerDims, safeName(dim.Name))
			}
		}

		outerDims = append(outerDims, timeDims...)

		measureFilterClause := ""
		var measureFilterArgs []any
		for _, m := range q.Measures {
			if m.Filter != nil {
				whereBuilder := &ExpressionBuilder{
					mv:       mv,
					dialect:  dialect,
					measures: q.Measures,
				}
				measureFilterClause, measureFilterArgs, err = whereBuilder.buildExpression(m.Filter)
				if err != nil {
					return "", nil, err
				}
			}
		}

		limitClause := ""
		subqueryLimitClause := ""
		limit := 0

		if q.Limit != nil {
			limit = int(*q.Limit)
			subQueryLimit := limit * 2 // the amount can be halved due to duplicates
			if q.Offset != 0 {
				subQueryLimit = int(q.Offset) + limit*2
			}
			subqueryLimitClause = fmt.Sprintf(" LIMIT %d", (subQueryLimit))
			limitClause = fmt.Sprintf(" LIMIT %d", limit)
		}

		// base subquery
		if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			args = append(args, q.TimeRange.Start.AsTime())
		}
		args = append(args, selectArgs...)
		args = append(args, measureFilterArgs...)
		args = append(args, baseTimeRangeArgs...)
		args = append(args, whereClauseArgs...)
		// comparison subquery
		if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			args = append(args, q.ComparisonTimeRange.Start.AsTime())
		}
		args = append(args, selectArgs...)
		args = append(args, measureFilterArgs...)
		args = append(args, comparisonTimeRangeArgs...)
		args = append(args, whereClauseArgs...)
		// outer query
		args = append(args, havingClauseArgs...)

		/*
				CASE expression rationale:
				Assume filter is `count(*)/10 FILTER publisher = 'Yahoo'`. FILTER doesn't support arithmetics on aggregation.
			    To mimic FILTER one can produce 2 tables, one with the filter condition and another without it and merge tables aftewards.
				Full table:
					Publisher count(*)
					Microsoft 20
					Yahoo 40

				Filtered table:
					Publisher count(*)
					Yahoo 40

				Result:
					Publisher count(*)
					Microsoft nil
					Yahoo 4

				But those tables should be limited if possible to increase performance. If they're limited the tables should be ordered.
				The full table should contain all the dimension values that are in the filtered table before being limited. That can be
				accomplished by including a virtual column with `CASE <filter condition> AS "casewhere"` expression and ordering by it.

				For example, Google can have multiple domains (news.google.com, google.com, ...):
				┌───────────┬────────────────────┬───────────┬──────────────────┐
				│    pub    │         m1         │ casewhere │      domain      │
				│  varchar  │       double       │   int32   │     varchar      │
				├───────────┼────────────────────┼───────────┼──────────────────┤
				│           │ 2.8476409530549693 │         1 │ news.google.com  │
				│ Google    │ 2.8545245256825496 │         1 │ news.google.com  │
				│ Google    │ 3.0921039628421707 │         0 │ google.com       │

				To get all the aggregations with specified condition grouping and ordering by 'casewhere' splits 'Google' group even when "domain" is unspecified:
				┌───────────┬────────────────────┬───────────┐
				│    pub    │         m1         │ casewhere │
				│  varchar  │       double       │   int32   │
				├───────────┼────────────────────┼───────────┤
				│           │ 2.8476409530549693 │         1 │
				│ Google    │ 2.8545245256825496 │         1 │
				│ Yahoo     │  2.986993492174502 │         0 │
				│ Facebook  │ 2.9896851248642617 │         0 │
				│           │ 3.0256741573033525 │         0 │
				│ Google    │ 3.0921039628421707 │         0 │
		*/
		sql = fmt.Sprintf(`
			SELECT * EXCLUDE(`+cw+`) FROM (
				-- SELECT ... as t_offset, base.d1, ..., base.td1, ..., base.m1, ... , comparison.td1 as td1__previous, ...
				SELECT `+strings.Join(slices.Concat(withPrefixCols("base", outerDims), withCaseAsClauses("coalesce(base."+cw+",comparison."+cw+") = 1", "null", finalMeasures), finalComparisonTimeDims), ",")+`, coalesce(base.`+cw+`,comparison.`+cw+`) as `+cw+` FROM 
					(
						-- removing duplicates
						SELECT t_offset, `+strings.Join(subSelectDimAliases, ",")+","+strings.Join(inFuncCols("FIRST", subselectMeasureAliases), ",")+`, FIRST(`+cw+`) AS `+cw+` FROM (
							-- SELECT t_offset, dim1 as d1, ... timed1 as td1, ..., avg(price) as m1, ... AND d2 = 'a' ...
							SELECT 
								%[1]s, 
								CASE WHEN `+measureFilterClause+` THEN 1 ELSE 0 END `+cw+` 
							FROM %[3]s %[7]s 
							WHERE (%[4]s)
							GROUP BY %[2]s, %[9]d 
							ORDER BY `+strings.Join(convertToNullCaseClauses(cw+" = 1", sortConstructs), ",")+" "+subqueryLimitClause+` 
						)
						GROUP BY `+groupForDims(len(subSelectDimAliases)+1)+`
					) base
				LEFT OUTER JOIN
					(
						-- removing duplicates
						SELECT t_offset, `+strings.Join(subSelectDimAliases, ",")+","+strings.Join(inFuncCols("FIRST", subselectComparisonAliases), ",")+`, FIRST(`+cw+`) AS `+cw+` FROM (
							-- SELECT t_offset, dim1 as d1, ..., timed1 as td1, ... avg(price) as m1, ... AND d2 = 'a' ...
							SELECT 
								`+comparisonSelectClause+`,`+` CASE WHEN `+measureFilterClause+` THEN 1 ELSE 0 END `+cw+`
							FROM %[3]s %[7]s 
							WHERE (%[5]s) 
							GROUP BY %[2]s, %[10]d 
							ORDER BY `+strings.Join(convertToNullCaseClauses(cw+" = 1", sortConstructs), ",")+" "+subqueryLimitClause+`
						)
						GROUP BY `+groupForDims(len(subSelectDimAliases)+1)+`
					) comparison
				ON
						`+strings.Join(append(joinConditions, "base."+cw+"= comparison."+cw), " AND ")+` -- base.t_offset = comparison.t_offset AND ...
			)
				`+outerWhereClause+` -- WHERE d1 = 'having_value'
				ORDER BY `+strings.Join(convertToNullCaseClauses(cw+"= 1", outerSortConstructs), ",")+` -- ORDER BY d1, ...
				`+limitClause+`
				OFFSET %[8]d
			`,
			baseSelectClause,                    // 1
			strings.Join(innerGroupCols, ","),   // 2
			escapeMetricsViewTable(dialect, mv), // 3
			baseWhereClause,                     // 4
			comparisonWhereClause,               // 5
			measureFilterClause,                 // 6
			strings.Join(unnestClauses, ""),     // 7
			q.Offset,                            // 8
			len(selectCols)+1,                   // 9 index of casewhere
			len(comparisonSelectCols)+1,         // 10 index of casewhere
		)
		// ^^^ Outer SELECT rationale:
		// Plain DuckDB query `SELECT publisher, avg(bid_price) AS bid_price ... HAVING bid_price > 0` is ambiguous because alias is the same.
	} else { // Druid measure filter
		limitClause := ""
		subqueryLimitClause := ""
		limit := 0

		if q.Limit != nil {
			limit = int(*q.Limit)
			subQueryLimit := limit * 2
			if q.Offset != 0 {
				subQueryLimit = int(q.Offset) + limit*2
			}
			subqueryLimitClause = fmt.Sprintf(" LIMIT %d", (subQueryLimit))
			limitClause = fmt.Sprintf(" LIMIT %d", limit)
		}

		measureFilterClause := ""
		var measureFilterArgs []any
		for _, m := range q.Measures {
			if m.Filter != nil {
				whereBuilder := &ExpressionBuilder{
					mv:       mv,
					dialect:  dialect,
					measures: q.Measures,
				}
				measureFilterClause, measureFilterArgs, err = whereBuilder.buildExpression(m.Filter)
				if err != nil {
					return "", nil, err
				}
			}
		}

		innerGroupCols := make([]string, 0, len(q.Dimensions)+1)
		outerGroupCols := make([]string, 0, len(q.Dimensions))
		innerGroupCols = append(innerGroupCols, "1")
		for i := range q.Dimensions {
			innerGroupCols = append(innerGroupCols, fmt.Sprintf("%d", i+2))
			outerGroupCols = append(outerGroupCols, fmt.Sprintf("%d", i+1))
		}

		args = args[:0]
		// base subquery
		if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			args = append(args, q.TimeRange.Start.AsTime())
		}
		args = append(args, selectArgs...)
		args = append(args, measureFilterArgs...)
		args = append(args, baseTimeRangeArgs...)
		args = append(args, whereClauseArgs...)

		// comparison subquery
		if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			args = append(args, q.ComparisonTimeRange.Start.AsTime())
		}
		args = append(args, selectArgs...)
		args = append(args, measureFilterArgs...)
		args = append(args, comparisonTimeRangeArgs...)
		args = append(args, whereClauseArgs...)

		// outer query
		args = append(args, havingClauseArgs...)

		if havingWhereClause != "" {
			havingWhereClause = " HAVING " + havingWhereClause
		}
		sql = fmt.Sprintf(`
				-- SELECT COALESCE(base.d1, comparison.d1), ..., base.m1, ..., base.m2 ... 
				SELECT `+strings.Join(slices.Concat(finalDims[1:], toClauses(inFunc("ANY_VALUE", withCase("coalesce(base."+cw+",comparison."+cw+") = 1", "null", finalMeasures))), finalComparisonTimeDims), ",")+` FROM 
					(
						-- 2 additional SELECTs because Druid doesn't support FIRST aggregation
						-- SELECT t_offset, d1, d2 ..., CASE WHEN ARRAY_LENGTH(cw) = 1 OR ARRAY_ORDINAL(cw,1) = 1 THEN ARRAY_ORDINAL(m1,1) ELSE ARRAY_ORDINAL(m1, 2) end m1 FROM (
						SELECT t_offset, `+strings.Join(subSelectDimAliases, ",")+`,`+strings.Join(caseArraySelectCols(cw, []string{originalMeasure, cw}), ",")+` FROM (
							-- SELECT t_offset, d1, d2 ..., array_agg(m1), array_agg(casewhere) FROM ( 
							SELECT t_offset, `+strings.Join(subSelectDimAliases, ",")+","+strings.Join(inFuncCols("ARRAY_AGG", []string{originalMeasure, cw}), ",")+` FROM (
								-- SELECT t_offset, dim1 as d1, dim2 as d2, timed1 as td1, avg(price) as m1, ... 
								SELECT 
									%[1]s,
									CASE WHEN `+measureFilterClause+` THEN 1 ELSE 0 END `+cw+` 
								FROM %[3]s %[6]s 
								WHERE %[4]s 
								GROUP BY %[2]s, %[7]d 
								ORDER BY `+strings.Join(slices.Concat(convertToNullCaseClauses(cw+"= 1", sortConstructs)), ",")+" "+subqueryLimitClause+`
							) GROUP BY t_offset, `+strings.Join(subSelectDimAliases, ",")+`
						)
					) base
				LEFT JOIN
					(
						SELECT t_offset, `+strings.Join(subSelectDimAliases, ",")+`,`+strings.Join(caseArraySelectCols(cw, []string{originalMeasure, cw}), ",")+` FROM (
							-- SELECT t_offset, d1, d2 ..., array_agg(m1), array_agg(casewhere) FROM ( 
							SELECT t_offset, `+strings.Join(subSelectDimAliases, ",")+","+strings.Join(inFuncCols("ARRAY_AGG", []string{originalMeasure, cw}), ",")+` FROM (
								SELECT 
									`+comparisonSelectClause+","+` CASE WHEN `+measureFilterClause+` THEN 1 ELSE 0 END `+cw+`  
								FROM %[3]s %[6]s 
								WHERE %[5]s 
								GROUP BY %[2]s, %[9]d 
								ORDER BY `+strings.Join(slices.Concat(convertToNullCaseClauses(cw+"= 1", sortConstructs)), ",")+" "+subqueryLimitClause+` 
							) GROUP BY t_offset, `+strings.Join(subSelectDimAliases, ",")+`
						)
					) comparison
				ON
				-- base.d1 IS NOT DISTINCT FROM comparison.d1 AND base.d2 IS NOT DISTINCT FROM comparison.d2 AND ...
						`+strings.Join(append(joinConditions, "base."+cw+" = comparison."+cw), " AND ")+` -- base.t_offset = comparison.t_offset AND ...
				GROUP BY `+strings.Join(outerGroupCols, ",")+", coalesce(base."+cw+",comparison."+cw+") "+havingWhereClause+`
				ORDER BY `+strings.Join(slices.Concat(convertToNullCaseClauses("coalesce(base."+cw+",comparison."+cw+") = 1", outerSortConstructs)), ",")+` -- ORDER BY d1, ...
				`+limitClause+`
				OFFSET %[8]d
			`,
			baseSelectClause,                    // 1
			strings.Join(innerGroupCols, ","),   // 2
			escapeMetricsViewTable(dialect, mv), // 3
			baseWhereClause,                     // 4
			comparisonWhereClause,               // 5
			strings.Join(unnestClauses, ""),     // 6
			len(selectCols)+1,                   // 7 index of casewhere
			q.Offset,                            // 8
			len(comparisonSelectCols)+1,         // 9
		)
	}

	return sql, args, nil
}

func groupForDims(n int) string {
	s := ""
	for i := range n {
		s += fmt.Sprintf("%d", i+1)
		if i != n-1 && n != 1 {
			s += ","
		}
	}
	return s
}

type FinalMeasure struct {
	expression string
	alias      string
}

type SortConstruct struct {
	expression string
	ending     string
	dim        bool
}

func caseArraySelectCols(condCol string, cols []string) []string {
	cs := make([]string, len(cols))
	for i, c := range cols {
		cs[i] = "CASE WHEN ARRAY_LENGTH(" + condCol + ") = 1 OR ARRAY_ORDINAL(" + condCol + ",1) = 1 THEN ARRAY_ORDINAL(" + c + ",1) ELSE ARRAY_ORDINAL(" + c + ", 2) END " + c
	}
	return cs
}

func convertToNullCaseClauses(cond string, sortConstructs []*SortConstruct) []string {
	cs := make([]string, len(sortConstructs))
	for i, sc := range sortConstructs {
		if sc.dim {
			cs[i] = sc.expression + " " + sc.ending
		} else {
			cs[i] = "CASE WHEN " + cond + " THEN " + sc.expression + " ELSE null END " + sc.ending
		}
	}
	return cs
}

func withCaseAsClauses(cond, output2 string, finalMeasures []*FinalMeasure) []string {
	cs := make([]string, len(finalMeasures))
	for i, fm := range finalMeasures {
		cs[i] = "CASE WHEN " + cond + " THEN " + fm.expression + " ELSE " + output2 + " END AS " + fm.alias
	}
	return cs
}

func withCase(cond, output2 string, finalMeasures []*FinalMeasure) []*FinalMeasure {
	cs := make([]*FinalMeasure, len(finalMeasures))
	for i, fm := range finalMeasures {
		cs[i] = &FinalMeasure{
			expression: "CASE WHEN " + cond + " THEN " + fm.expression + " ELSE " + output2 + " END",
			alias:      fm.alias,
		}
	}
	return cs
}

func inFunc(name string, cols []*FinalMeasure) []*FinalMeasure {
	cs := make([]*FinalMeasure, len(cols))
	for i, c := range cols {
		cs[i] = &FinalMeasure{
			expression: fmt.Sprintf("%s(%s)", name, c.expression),
			alias:      c.alias,
		}
	}
	return cs
}

func inFuncCols(name string, cols []string) []string {
	cs := make([]string, len(cols))
	for i, c := range cols {
		cs[i] = fmt.Sprintf("%s(%s) AS %s", name, c, c)
	}
	return cs
}

func toClauses(cols []*FinalMeasure) []string {
	cs := make([]string, len(cols))
	for i, c := range cols {
		cs[i] = fmt.Sprintf("%s AS %s", c.expression, c.alias)
	}
	return cs
}

func withPrefixCols(prefix string, cols []string) []string {
	cs := make([]string, len(cols))
	for i, c := range cols {
		cs[i] = prefix + "." + c
	}
	return cs
}

func (q *MetricsViewAggregation) buildMetricsComparisonAggregationSQL(ctx context.Context, olap drivers.OLAPStore, priority int, mv *runtimev1.MetricsViewSpec, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity, export bool) (string, []any, error) {
	if len(q.Dimensions) == 0 && len(q.Measures) == 0 {
		return "", nil, errors.New("no dimensions or measures specified")
	}
	dimByName := make(map[string]*runtimev1.MetricsViewAggregationDimension, len(mv.Dimensions))
	measuresByFinalName := make(map[string]*runtimev1.MetricsViewAggregationMeasure, len(q.Measures))
	for _, d := range q.Dimensions {
		if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED || d.Alias == "" {
			dimByName[d.Name] = d
		} else {
			dimByName[d.Alias] = d
		}
	}
	for _, m := range q.Measures {
		measuresByFinalName[m.Name] = m
	}

	cols := q.cols()
	selectCols := make([]string, 0, cols+1)
	var comparisonSelectCols []string

	finalDims := make([]string, 0, len(q.Dimensions))
	joinConditions := make([]string, 0, len(q.Dimensions))

	unnestClauses := make([]string, 0)
	var selectArgs []any

	err := q.calculateMeasuresMeta()
	if err != nil {
		return "", nil, err
	}

	// Required for t_offset, ie
	// SELECT t_offset, d1, d2, t1, t2, m1, m2
	minTimeGrain := runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
	for _, d := range q.Dimensions {
		if d.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED && d.GetName() == mv.TimeDimension {
			if minTimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED || d.TimeGrain < minTimeGrain {
				minTimeGrain = d.TimeGrain
			}
		}
	}

	// it's required for joining the base and comparison tables
	timeOffsetExpression, err := q.buildOffsetExpression(mv.TimeDimension, minTimeGrain, dialect)
	if err != nil {
		return "", nil, err
	}

	colMap := make(map[string]int, q.cols())

	selectCols = append(selectCols, timeOffsetExpression)
	comparisonSelectCols = append(comparisonSelectCols, timeOffsetExpression)

	joinConditions = append(joinConditions, "base.t_offset = comparison.t_offset")

	// these go last to decrease complexity of indexing columns
	var finalComparisonTimeDims []string
	mvDimsByName := make(map[string]*runtimev1.MetricsViewSpec_DimensionV2, len(mv.Dimensions))
	for _, d := range q.Dimensions {
		// Handle regular dimensions
		if d.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			dim, err := metricsViewDimension(mv, d.Name)
			if err != nil {
				return "", nil, err
			}
			mvDimsByName[d.Name] = dim
			dimSel, unnestClause := dialect.DimensionSelect(mv.Database, mv.DatabaseSchema, mv.Table, dim)
			selectCols = append(selectCols, dimSel)
			comparisonSelectCols = append(comparisonSelectCols, dimSel)
			finalDims = append(finalDims, fmt.Sprintf("COALESCE(base.%[1]s,comparison.%[1]s) as %[1]s", safeName(dim.Name)))
			if unnestClause != "" {
				unnestClauses = append(unnestClauses, unnestClause)
			}
			colMap[d.Name] = len(selectCols)
			var joinCondition string
			if dialect == drivers.DialectClickHouse {
				joinCondition = fmt.Sprintf("isNotDistinctFrom(base.%[1]s, comparison.%[1]s)", safeName(dim.Name))
			} else {
				joinCondition = fmt.Sprintf("base.%[1]s IS NOT DISTINCT FROM comparison.%[1]s", safeName(dim.Name))
			}
			joinConditions = append(joinConditions, joinCondition)
			continue
		}

		// Handle time dimension
		expr, exprArgs, err := q.buildTimestampExpr(mv, d, dialect)
		if err != nil {
			return "", nil, err
		}
		alias := d.Name
		if d.Alias != "" {
			alias = d.Alias
		}
		timeDimClause := fmt.Sprintf("%s as %s", expr, safeName(alias))

		selectCols = append(selectCols, timeDimClause)
		colMap[alias] = len(selectCols)
		comparisonSelectCols = append(comparisonSelectCols, timeDimClause)
		finalDims = append(finalDims, fmt.Sprintf("base.%[1]s as %[1]s", safeName(alias)))
		// workaround for Druid time conversion with aggregates bug
		if dialect == drivers.DialectDruid {
			finalComparisonTimeDims = append(finalComparisonTimeDims, fmt.Sprintf("MILLIS_TO_TIMESTAMP(PARSE_LONG(ANY_VALUE(comparison.%[1]s))) as %[2]s", safeName(alias), safeName(alias+"__previous")))
		} else {
			finalComparisonTimeDims = append(finalComparisonTimeDims, fmt.Sprintf("comparison.%[1]s as %[2]s", safeName(alias), safeName(alias+"__previous")))
		}

		selectArgs = append(selectArgs, exprArgs...)
	}

	labelMap := make(map[string]string, len(mv.Measures))
	for _, m := range mv.Measures {
		labelMap[m.Name] = m.Name
		if m.Label != "" {
			labelMap[m.Name] = m.Label
		}
	}

	// collect subquery expressions
	for _, m := range q.Measures {
		switch m.Compute.(type) {
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue, *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta, *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
			// nothing
		case *runtimev1.MetricsViewAggregationMeasure_Count:
			selectCols = append(selectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("COUNT(*) as %s", safeName(m.Name)))
			}
		case *runtimev1.MetricsViewAggregationMeasure_CountDistinct:
			arg := m.GetCountDistinct().GetDimension()
			if arg == "" {
				return "", nil, fmt.Errorf("builtin measure '%s' expects non-empty string argument, got '%v'", m.BuiltinMeasure.String(), m.BuiltinMeasureArgs[0])
			}
			selectCols = append(selectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("COUNT(DISTINCT %s) as %s", safeName(arg), safeName(m.Name)))
			}
		default:
			expr, err := metricsViewMeasureExpression(mv, m.Name)
			if err != nil {
				return "", nil, err
			}
			selectCols = append(selectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
			if q.measuresMeta[m.Name].expand {
				comparisonSelectCols = append(comparisonSelectCols, fmt.Sprintf("%s as %s", expr, safeName(m.Name)))
			}
		}
	}

	// collect final expressions
	var finalSelectCols []string
	var labelCols []string
	for _, m := range q.Measures {
		var columnsTuple string
		var labelTuple string
		var subqueryName, finalName string
		prefix := ""
		if dialect == drivers.DialectDruid {
			prefix = "ANY_VALUE"
		}

		switch m.Compute.(type) {
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
			subqueryName = m.GetComparisonRatio().Measure
			finalName = m.Name
			if dialect == drivers.DialectDruid {
				columnsTuple = fmt.Sprintf(
					"ANY_VALUE(SAFE_DIVIDE(base.%[1]s - comparison.%[1]s, CAST(comparison.%[1]s AS DOUBLE))) AS %[2]s",
					safeName(subqueryName),
					safeName(finalName),
				)
			} else {
				columnsTuple = fmt.Sprintf(
					"(base.%[1]s - comparison.%[1]s)/comparison.%[1]s::DOUBLE AS %[2]s",
					safeName(subqueryName),
					safeName(finalName),
				)
			}
			labelTuple = columnsTuple
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
			subqueryName = m.GetComparisonDelta().Measure
			finalName = m.Name
			columnsTuple = fmt.Sprintf(
				"%[3]s(base.%[1]s - comparison.%[1]s) AS %[2]s",
				safeName(subqueryName),
				safeName(finalName),
				prefix,
			)
			labelTuple = columnsTuple
		case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
			subqueryName = m.GetComparisonValue().Measure
			finalName = m.Name
			columnsTuple = fmt.Sprintf(
				"%[3]s(comparison.%[1]s) AS %[2]s",
				safeName(subqueryName),
				safeName(finalName),
				prefix,
			)
			labelTuple = columnsTuple
		case *runtimev1.MetricsViewAggregationMeasure_Count, *runtimev1.MetricsViewAggregationMeasure_CountDistinct:
			subqueryName = m.Name
			finalName = m.Name
			columnsTuple = fmt.Sprintf(
				"%[3]s(base.%[1]s) AS %[1]s",
				safeName(subqueryName),
				safeName(finalName),
				prefix,
			)
			labelTuple = columnsTuple
		default: // not a virtual (generated) column
			subqueryName = m.Name
			finalName = m.Name
			columnsTuple = fmt.Sprintf(
				"%[3]s(base.%[1]s) AS %[1]s",
				safeName(subqueryName),
				safeName(finalName),
				prefix,
			)
			labelTuple = fmt.Sprintf( // non-virtial columns have a label
				"%[3]s(base.%[1]s) AS %[1]s",
				safeName(subqueryName),
				safeName(labelMap[subqueryName]),
				prefix,
			)
		}
		finalSelectCols = append(
			finalSelectCols,
			columnsTuple,
		)
		labelCols = append(labelCols, labelTuple)
	}

	baseSelectClause := strings.Join(selectCols, ", ")
	comparisonSelectClause := strings.Join(comparisonSelectCols, ", ")
	finalSelectClause := strings.Join(finalSelectCols, ", ")
	labelSelectClause := strings.Join(labelCols, ", ")
	if export {
		finalSelectClause = labelSelectClause
	}

	baseWhereClause := "1=1"
	comparisonWhereClause := "1=1"

	if mv.TimeDimension == "" {
		return "", nil, fmt.Errorf("metrics view '%s' doesn't have time dimension", q.MetricsViewName)
	}

	td := safeName(mv.TimeDimension)
	if dialect == drivers.DialectDuckDB {
		td = fmt.Sprintf("%s::TIMESTAMP", td)
	}

	whereBuilder := &ExpressionBuilder{
		mv:       mv,
		dialect:  dialect,
		measures: q.Measures,
	}
	whereClause, whereClauseArgs, err := whereBuilder.buildExpression(q.Where)
	if err != nil {
		return "", nil, err
	}

	var baseTimeRangeArgs []any
	trc, err := timeRangeClause(q.TimeRange, mv, td, &baseTimeRangeArgs)
	if err != nil {
		return "", nil, err
	}
	baseWhereClause += trc

	if whereClause != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", whereClause)
	}

	var comparisonTimeRangeArgs []any
	trc, err = timeRangeClause(q.ComparisonTimeRange, mv, td, &comparisonTimeRangeArgs)
	if err != nil {
		return "", nil, err
	}
	comparisonWhereClause += trc

	if whereClause != "" {
		comparisonWhereClause += fmt.Sprintf(" AND (%s)", whereClause)
	}

	if policy != nil && policy.RowFilter != "" {
		baseWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
		comparisonWhereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	havingClause := "1=1"
	var havingClauseArgs []any
	if q.Having != nil {
		havingBuilder := &ExpressionBuilder{
			mv:       mv,
			dialect:  dialect,
			measures: q.Measures,
		}
		havingClause, havingClauseArgs, err = havingBuilder.buildExpression(q.Having)
		if err != nil {
			return "", nil, err
		}
	}

	var orderClauses []string
	var baseOrderClauses []string
	var comparisonOrderClauses []string

	for _, s := range q.Sort {
		var outerClause, subQueryClause string
		if dimByName[s.Name] != nil { // dimension
			outerClause = fmt.Sprintf("%d", colMap[s.Name]-1)
			subQueryClause = fmt.Sprintf("%d", colMap[s.Name])
		} else if measuresByFinalName[s.Name] != nil { // measure
			m := measuresByFinalName[s.Name]
			outerClause = s.Name
			subQueryClause = OriginalColumnName(m)
		} else {
			return "", nil, fmt.Errorf("no selected dimension or measure '%s' found for sorting", s.Name)
		}

		var ending string
		if s.Desc {
			ending += " DESC"
		}
		if dialect == drivers.DialectDuckDB {
			ending += " NULLS LAST"
		}
		outerClause += ending
		subQueryClause += ending
		orderClauses = append(orderClauses, outerClause)
		baseOrderClauses = append(baseOrderClauses, subQueryClause)
		comparisonOrderClauses = append(comparisonOrderClauses, subQueryClause)
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
		limitClause = fmt.Sprintf(" LIMIT %d", *q.Limit)
	}

	baseLimitClause := ""
	comparisonLimitClause := ""

	joinType := "FULL"
	comparisonSort := false
	deltaComparison := false
	for _, s := range q.Sort {
		m := measuresByFinalName[s.Name]
		if measuresByFinalName[s.Name] != nil {
			switch m.Compute.(type) {
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
				comparisonSort = true
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta, *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
				deltaComparison = true
				comparisonSort = true
			}
		}
	}
	if !q.Exact {
		limit := 0
		if q.Limit != nil {
			limit = int(*q.Limit)
		}
		approximationLimit := limit
		if limit != 0 && limit < 100 && deltaComparison {
			approximationLimit = 100
		}
		approximationLimit = (approximationLimit + int(q.Offset)) * 2

		if len(q.Sort) == 0 || !comparisonSort {
			joinType = "LEFT OUTER"
			baseLimitClause = baseSubQueryOrderByClause
			if approximationLimit > 0 {
				baseLimitClause += fmt.Sprintf(" LIMIT %d", approximationLimit)
			}
		} else {
			joinType = "RIGHT OUTER"
			comparisonLimitClause = comparisonSubQueryOrderByClause
			if approximationLimit > 0 {
				comparisonLimitClause += fmt.Sprintf(" LIMIT %d", approximationLimit)
			}
		}
	}

	/*
		Example of the SQL:

		SELECT * from (
				-- SELECT d1, d2, d3, td1, td2, m1, m2 ... , td1__previous, td2__previous
				SELECT COALESCE(base."pub",comparison."pub") as "pub",COALESCE(base."dom",comparison."dom") as "dom",base."timestamp" as "timestamp",base."timestamp_year" as "timestamp_year", base."measure_0" AS "measure_0", comparison."measure_0" AS "measure_0__previous", base."measure_0" - comparison."measure_0" AS "measure_0__delta_abs", (base."measure_0" - comparison."measure_0")/comparison."measure_0"::DOUBLE AS "measure_0__delta_rel", base."measure_1" AS "measure_1", base."m1" AS "m1", comparison."m1" AS "m1__previous", base."m1" - comparison."m1" AS "m1__delta_abs", (base."m1" - comparison."m1")/comparison."m1"::DOUBLE AS "m1__delta_rel" , comparison."timestamp" as "timestamp__previous", comparison."timestamp_year" as "timestamp_year__previous" FROM
					(
						-- SELECT t_offset, d1, d2, d3, td1, td2, m1, m2 ...
						SELECT epoch_ms(date_trunc('DAY', "timestamp")::TIMESTAMP)-epoch_ms(date_trunc('DAY', ?)::TIMESTAMP) as t_offset, ("publisher") as "pub", ("domain") as "dom", date_trunc('DAY', "timestamp"::TIMESTAMP)::TIMESTAMP as "timestamp", date_trunc('YEAR', "timestamp"::TIMESTAMP)::TIMESTAMP as "timestamp_year", count(*) as "measure_0", avg(bid_price) as "measure_1", avg(bid_price) as "m1" FROM "ad_bids"  WHERE 1=1 AND "timestamp"::TIMESTAMP >= ? AND "timestamp"::TIMESTAMP < ? AND ((("publisher") = (?)) OR (("publisher") = (?))) GROUP BY 1,2,3,4,5
					) base
				FULL JOIN
					(
						SELECT epoch_ms(date_trunc('DAY', "timestamp")::TIMESTAMP)-epoch_ms(date_trunc('DAY', ?)::TIMESTAMP) as t_offset, ("publisher") as "pub", ("domain") as "dom", date_trunc('DAY', "timestamp"::TIMESTAMP)::TIMESTAMP as "timestamp", date_trunc('YEAR', "timestamp"::TIMESTAMP)::TIMESTAMP as "timestamp_year", count(*) as "measure_0", avg(bid_price) as "m1" FROM "ad_bids"  WHERE 1=1 AND "timestamp"::TIMESTAMP >= ? AND "timestamp"::TIMESTAMP < ? AND ((("publisher") = (?)) OR (("publisher") = (?))) GROUP BY 1,2,3,4,5
					) comparison
				ON
						base.t_offset = comparison.t_offset AND base."pub" IS NOT DISTINCT FROM comparison."pub" AND base."dom" IS NOT DISTINCT FROM comparison."dom"
				ORDER BY 4 NULLS LAST, 2 NULLS LAST, 3 NULLS LAST, 5 NULLS LAST, 9 NULLS LAST
					LIMIT 1374419126128
				OFFSET
					0
			) WHERE 1=1 AND ("measure_1") > (?)

		Example of arguments:
		[2022-01-01 00:00:00 +0000 UTC 2022-01-01 00:00:00 +0000 UTC 2022-01-03 00:00:00 +0000 UTC Yahoo Google 2022-01-03 00:00:00 +0000 UTC 2022-01-03 00:00:00 +0000 UTC 2022-01-05 00:00:00 +0000 UTC Yahoo Google 0]
	*/

	var args []any
	var sql string
	if dialect != drivers.DialectDruid {
		// base subquery
		if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			args = append(args, q.TimeRange.Start.AsTime())
		}
		args = append(args, selectArgs...)
		args = append(args, baseTimeRangeArgs...)
		args = append(args, whereClauseArgs...)
		// comparison subquery
		if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			args = append(args, q.ComparisonTimeRange.Start.AsTime())
		}
		args = append(args, selectArgs...)
		args = append(args, comparisonTimeRangeArgs...)
		args = append(args, whereClauseArgs...)
		// outer query
		args = append(args, havingClauseArgs...)

		// Using expr was causing issues with query arg expansion in duckdb.
		// Using column name is not possible either since it will take the original column name instead of the aliased column name
		// But using numbered group we can exactly target the correct selected column.
		// Note that the non-timestamp columns also use the numbered group-by for constancy.

		// Inner grouping should include t_offset
		// SELECT t_offset, d1, d2, d3 ... GROUP BY 1, 2, 3, 4 ...
		innerGroupCols := make([]string, 0, len(q.Dimensions)+1)
		innerGroupCols = append(innerGroupCols, "1")
		for i := range q.Dimensions {
			innerGroupCols = append(innerGroupCols, fmt.Sprintf("%d", i+2))
		}

		// these go last to decrease complexity of indexing columns
		finalTimeDimsClause := ""
		if len(finalComparisonTimeDims) > 0 {
			finalTimeDimsClause = fmt.Sprintf(", %s", strings.Join(finalComparisonTimeDims, ", "))
		}

		// measure filter could include the base measure name.
		// this leads to ambiguity whether it applies to the base.measure ot comparison.measure.
		// to keep the clause builder consistent we add an outer query here.
		sql = fmt.Sprintf(`
				SELECT * from (
					-- SELECT base.d1 d1, base.d2 d2, base.timed1 td1, base.m1 m1, comparison.m2 m2 ... , comparison.timed1 td1__previous, ...
					SELECT %[2]s %[18]s FROM 
						(
							-- SELECT t_offset, d1, d2, d3, td1, td2, m1, m2 ...
							SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s GROUP BY %[10]s %[12]s 
						) base
					%[11]s JOIN
						(
							SELECT %[16]s FROM %[3]s %[14]s WHERE %[5]s GROUP BY %[10]s %[13]s 
						) comparison
					ON %[17]s
					%[6]s
				) WHERE 1=1 AND %[15]s 
				%[7]s
				OFFSET %[8]d

			`,
			baseSelectClause, // 1
			strings.Join(slices.Concat(finalDims, []string{finalSelectClause}), ","), // 2
			escapeMetricsViewTable(dialect, mv),                                      // 3
			baseWhereClause,                                                          // 4
			comparisonWhereClause,                                                    // 5
			orderByClause,                                                            // 6
			limitClause,                                                              // 7
			q.Offset,                                                                 // 8
			finalSelectClause,                                                        // 9
			strings.Join(innerGroupCols, ","),                                        // 10
			joinType,                                                                 // 11
			baseLimitClause,                                                          // 12
			comparisonLimitClause,                                                    // 13
			strings.Join(unnestClauses, ""),                                          // 14
			havingClause,                                                             // 15
			comparisonSelectClause,                                                   // 16
			strings.Join(joinConditions, " AND "),                                    // 17
			finalTimeDimsClause,                                                      // 18
		)
		fmt.Println(sql)
	} else {
		if !comparisonSort || len(q.Dimensions) == 0 { // no dimensions means a single row (totals) - no sorting is required
			// SELECT d1, d2, d3 ... GROUP BY 1, 2, 3 ...
			var innerGroupCols []string
			var whereDimConditions []string
			if len(q.Dimensions) > 0 { // an additional request for Druid to prevent full scan
				innerGroupCols = make([]string, 0, len(q.Dimensions))
				for i := range q.Dimensions {
					innerGroupCols = append(innerGroupCols, fmt.Sprintf("%d", i+2))
				}

				nonTimeCols := make([]string, 0, len(selectCols)) // avoid group by time cols
				for i, s := range selectCols[1:] {                // skip t_offset
					if i >= len(q.Dimensions) {
						nonTimeCols = append(nonTimeCols, s)
					} else {
						if q.Dimensions[i].TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
							nonTimeCols = append(nonTimeCols, "1")
						} else {
							nonTimeCols = append(nonTimeCols, s)
						}
					}
				}
				sql = fmt.Sprintf("SELECT %[1]s FROM %[2]s %[3]s WHERE %[4]s GROUP BY %[5]s %[6]s",
					strings.Join(slices.Concat([]string{"1"}, nonTimeCols), ","), // 1
					escapeMetricsViewTable(dialect, mv),                          // 2
					strings.Join(unnestClauses, ""),                              // 3
					baseWhereClause,                                              // 4
					strings.Join(innerGroupCols, ","),                            // 5
					baseLimitClause,                                              // 6
				)

				var druidArgs []any
				druidArgs = append(druidArgs, selectArgs...)
				druidArgs = append(druidArgs, baseTimeRangeArgs...)
				druidArgs = append(druidArgs, whereClauseArgs...)

				_, result, err := olapQuery(ctx, olap, priority, sql, druidArgs)
				if err != nil {
					return "", nil, err
				}

				// without this extra where condition, the join will be a full scan
				for _, row := range result {
					var dimConditions []string
					for _, dim := range q.Dimensions {
						if dim.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
							field := row.Fields[dim.Name]

							// Druid doesn't resolve aliases in where clause
							mvDim := mvDimsByName[dim.Name]

							_, ok := field.GetKind().(*structpb.Value_NullValue)
							if ok {
								dimConditions = append(dimConditions, fmt.Sprintf("%[1]s is null", safeName(mvDim.Column)))
							} else {
								dimConditions = append(dimConditions, fmt.Sprintf("%[1]s = '%[2]s'", safeName(mvDim.Column), field.AsInterface()))
							}
						}
					}
					whereDimConditions = append(whereDimConditions, strings.Join(dimConditions, " AND "))
				}
			}

			innerGroupCols = make([]string, 0, len(q.Dimensions)+1)
			outerGroupCols := make([]string, 0, len(q.Dimensions))
			innerGroupCols = append(innerGroupCols, "1")
			for i := range q.Dimensions {
				innerGroupCols = append(innerGroupCols, fmt.Sprintf("%d", i+2))
				outerGroupCols = append(outerGroupCols, fmt.Sprintf("%d", i+1))
			}
			// base subquery
			if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				args = append(args, q.TimeRange.Start.AsTime())
			}
			args = append(args, selectArgs...)
			args = append(args, baseTimeRangeArgs...)
			args = append(args, whereClauseArgs...)
			// comparison subquery
			if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				args = append(args, q.ComparisonTimeRange.Start.AsTime())
			}
			args = append(args, selectArgs...)
			args = append(args, comparisonTimeRangeArgs...)
			args = append(args, whereClauseArgs...)
			// outer query
			args = append(args, havingClauseArgs...)

			finalTimeDimsClause := ""
			if len(finalComparisonTimeDims) > 0 {
				finalTimeDimsClause = fmt.Sprintf(", %s", strings.Join(finalComparisonTimeDims, ", "))
			}

			whereDimClause := ""
			outerGroupClause := ""
			if len(whereDimConditions) > 0 {
				whereDimClause = fmt.Sprintf(" AND (%s) ", strings.Join(whereDimConditions, " OR "))
			}
			if len(q.Dimensions) > 0 {
				outerGroupClause = " GROUP BY " + strings.Join(outerGroupCols, ",")
			}

			sql = fmt.Sprintf(`
				SELECT * from (
					-- SELECT base.d1 d1, base.d2 d2, base.timed1 td1, base.m1 m1, comparison.m2 m2 ... , comparison.timed1 td1__previous, ...
					SELECT %[2]s %[9]s FROM 
						(
							-- SELECT t_offset, d1, d2, d3, td1, td2, m1, m2 ... 
							SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s GROUP BY %[10]s %[12]s 
						) base
					LEFT JOIN
						(
							SELECT %[16]s FROM %[3]s %[14]s WHERE %[5]s %[18]s GROUP BY %[10]s %[13]s 
						) comparison
					ON
					-- base.d1 IS NOT DISTINCT FROM comparison.d1 AND base.d2 IS NOT DISTINCT FROM comparison.d2 AND ...
							%[17]s
					%[11]s -- GROUP BY ...
					%[6]s -- ORDER BY ...
				) WHERE 1=1 AND %[15]s 
				%[7]s -- LIMIT ...
				OFFSET %[8]d
			`,
				baseSelectClause, // 1
				strings.Join(slices.Concat(finalDims, []string{finalSelectClause}), ","), // 2
				escapeMetricsViewTable(dialect, mv),                                      // 3
				baseWhereClause,                                                          // 4
				comparisonWhereClause,                                                    // 5
				orderByClause,                                                            // 6
				limitClause,                                                              // 7
				q.Offset,                                                                 // 8
				finalTimeDimsClause,                                                      // 9
				strings.Join(innerGroupCols, ","),                                        // 10
				outerGroupClause,                                                         // 11
				baseLimitClause,                                                          // 12
				comparisonLimitClause,                                                    // 13
				strings.Join(unnestClauses, ""),                                          // 14
				havingClause,                                                             // 15
				comparisonSelectClause,                                                   // 16
				strings.Join(joinConditions, " AND "),                                    // 17
				whereDimClause,                                                           // 18
				outerGroupClause,                                                         // 19
			)
		} else {
			limit := 0
			if q.Limit == nil {
				limit = 0
			}
			approximationLimit := limit
			if limit != 0 && limit < 100 {
				approximationLimit = 100
			}

			comparisonLimitClause = comparisonSubQueryOrderByClause
			if approximationLimit > 0 {
				comparisonLimitClause += fmt.Sprintf(" LIMIT %d OFFSET %d", approximationLimit, q.Offset)
			}
			innerGroupCols := make([]string, 0, len(q.Dimensions))
			for i := range q.Dimensions {
				innerGroupCols = append(innerGroupCols, fmt.Sprintf("%d", i+2))
			}
			nonTimeCols := make([]string, 0, len(comparisonSelectCols)) // avoid group by time cols
			for i, s := range comparisonSelectCols[1:] {                // skip t_offset
				if i >= len(q.Dimensions) {
					nonTimeCols = append(nonTimeCols, s)
				} else {
					if q.Dimensions[i].TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
						nonTimeCols = append(nonTimeCols, "1")
					} else {
						nonTimeCols = append(nonTimeCols, s)
					}
				}
			}
			sql = fmt.Sprintf("SELECT %[1]s FROM %[2]s %[3]s WHERE %[4]s GROUP BY %[5]s %[6]s",
				strings.Join(slices.Concat([]string{"1"}, nonTimeCols), ","), // 1
				escapeMetricsViewTable(dialect, mv),                          // 2
				strings.Join(unnestClauses, ""),                              // 3
				comparisonWhereClause,                                        // 4
				strings.Join(innerGroupCols, ","),                            // 5
				comparisonLimitClause,                                        // 6
			)

			var druidArgs []any
			druidArgs = append(druidArgs, selectArgs...)
			druidArgs = append(druidArgs, comparisonTimeRangeArgs...)
			druidArgs = append(druidArgs, whereClauseArgs...)

			_, result, err := olapQuery(ctx, olap, priority, sql, druidArgs)
			if err != nil {
				return "", nil, err
			}

			// without this extra where condition, the join will be a full scan
			var whereDimConditions []string
			for _, row := range result {
				var dimConditions []string
				for _, dim := range q.Dimensions {
					if dim.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
						field := row.Fields[dim.Name]

						// Druid doesn't resolve aliases in where clause
						mvDim := mvDimsByName[dim.Name]

						_, ok := field.GetKind().(*structpb.Value_NullValue)
						if ok {
							dimConditions = append(dimConditions, fmt.Sprintf("%[1]s is null", safeName(mvDim.Column)))
						} else {
							dimConditions = append(dimConditions, fmt.Sprintf("%[1]s = '%[2]s'", safeName(mvDim.Column), field.AsInterface()))
						}
					}
				}
				whereDimConditions = append(whereDimConditions, strings.Join(dimConditions, " AND "))
			}

			innerGroupCols = make([]string, 0, len(q.Dimensions)+1)
			outerGroupCols := make([]string, 0, len(q.Dimensions))
			innerGroupCols = append(innerGroupCols, "1")
			for i := range q.Dimensions {
				innerGroupCols = append(innerGroupCols, fmt.Sprintf("%d", i+2))
				outerGroupCols = append(outerGroupCols, fmt.Sprintf("%d", i+1))
			}

			// base subquery
			if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				args = append(args, q.TimeRange.Start.AsTime())
			}
			args = append(args, selectArgs...)
			args = append(args, baseTimeRangeArgs...)
			args = append(args, whereClauseArgs...)
			// comparison subquery
			if minTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				args = append(args, q.ComparisonTimeRange.Start.AsTime())
			}
			args = append(args, selectArgs...)
			args = append(args, comparisonTimeRangeArgs...)
			args = append(args, whereClauseArgs...)
			// outer query
			args = append(args, havingClauseArgs...)

			finalTimeDimsClause := ""
			if len(finalComparisonTimeDims) > 0 {
				finalTimeDimsClause = fmt.Sprintf(", %s", strings.Join(finalComparisonTimeDims, ", "))
			}

			whereDimClause := ""
			outerGroupClause := ""
			if len(whereDimConditions) > 0 {
				whereDimClause = fmt.Sprintf(" AND (%s) ", strings.Join(whereDimConditions, " OR "))
			}
			if len(q.Dimensions) > 0 {
				outerGroupClause = " GROUP BY " + strings.Join(outerGroupCols, ",")
			}

			sql = fmt.Sprintf(`
				SELECT * from (
					SELECT %[2]s %[9]s FROM 
						(
							SELECT %[1]s FROM %[3]s %[14]s WHERE %[4]s %[18]s GROUP BY %[10]s %[12]s 
						) base
					LEFT JOIN
						(
							SELECT %[16]s FROM %[3]s %[14]s WHERE %[5]s GROUP BY %[10]s %[13]s 
						) comparison
					ON
							%[17]s
					%[11]s
					%[6]s
					%[7]s
					OFFSET
						%[8]d
				) WHERE 1=1 AND %[15]s 
			`,
				baseSelectClause, // 1
				strings.Join(slices.Concat(finalDims, []string{finalSelectClause}), ","), // 2
				escapeMetricsViewTable(dialect, mv),                                      // 3
				baseWhereClause,                                                          // 4
				comparisonWhereClause,                                                    // 5
				orderByClause,                                                            // 6
				limitClause,                                                              // 7
				q.Offset,                                                                 // 8
				finalTimeDimsClause,                                                      // 9
				strings.Join(innerGroupCols, ","),                                        // 10
				outerGroupClause,                                                         // 11
				baseLimitClause,                                                          // 12
				comparisonLimitClause,                                                    // 13
				strings.Join(unnestClauses, ""),                                          // 14
				havingClause,                                                             // 15
				comparisonSelectClause,                                                   // 16
				strings.Join(joinConditions, " AND "),                                    // 17
				whereDimClause,                                                           // 18
			)
		}
	}

	return sql, args, nil
}

func OriginalColumnName(m *runtimev1.MetricsViewAggregationMeasure) string {
	switch v := m.Compute.(type) {
	case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
		return v.ComparisonValue.Measure
	case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
		return v.ComparisonDelta.Measure
	case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
		return v.ComparisonRatio.Measure
	default:
		return m.Name
	}
}

func (q *MetricsViewAggregation) calculateMeasuresMeta() error {
	q.measuresMeta = make(map[string]metricsViewMeasureMeta, len(q.Measures))

	expands := make(map[string]bool, len(q.Measures))
	originalNames := make(map[string]bool, len(q.Measures))
	for _, m := range q.Measures {
		name := OriginalColumnName(m)
		if OriginalColumnName(m) != m.Name {
			expands[name] = true
		} else {
			originalNames[name] = true
		}
	}
	for n := range expands {
		if !originalNames[n] {
			return fmt.Errorf("original measure '%s' should be in the selection list", n)
		}
	}

	for _, m := range q.Measures {
		expand := false
		if expands[OriginalColumnName(m)] {
			expand = true
		}
		q.measuresMeta[m.Name] = metricsViewMeasureMeta{
			expand: expand,
		}
	}

	compare := !isTimeRangeNil(q.ComparisonTimeRange)
	if compare && len(expands) == 0 {
		return fmt.Errorf("no measures to compare")
	}

	return nil
}

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
func (q *MetricsViewAggregation) buildMeasureFilterSQL(mv *runtimev1.MetricsViewSpec, unnestClauses, selectCols []string, limitClause, orderClause, havingClause, whereClause, groupClause string, args, selectArgs, whereArgs, havingClauseArgs []any, extraWhereClause string, extraWhereClauseArgs []any, dialect drivers.Dialect) (string, []any, error) {
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
	builder := &ExpressionBuilder{
		mv:      mv,
		dialect: dialect,
	}
	measureExpression, measureWhereArgs, err := builder.buildExpression(q.Measures[0].Filter)
	if err != nil {
		return "", nil, err
	}

	if whereClause == "" {
		whereClause = "WHERE 1=1"
	}

	measureWhereClause := whereClause + fmt.Sprintf(" AND (%s)", measureExpression)
	if extraWhereClause != "" {
		extraWhereClause = "WHERE " + extraWhereClause
	}
	druidGroupBy := ""
	if dialect == drivers.DialectDruid {
		druidGroupBy = groupClause
	}

	/*
		SQL example:
		SELECT "pub","measure_1" FROM (
			SELECT "ad_bids"."pub", self_join3ba680fe589e49ceabf404f6c6d920e7."measure_1" as "measure_1" FROM (
				SELECT ("publisher") as "pub", COUNT(*) as "measure_1" FROM "ad_bids"  WHERE 1=1 GROUP BY 1
			) "ad_bids"
			LEFT JOIN (
				SELECT ("publisher") as "pub", COUNT(*) as "measure_1" FROM "ad_bids"  WHERE 1=1 AND (("domain") = (?)) GROUP BY 1
			) self_join3ba680fe589e49ceabf404f6c6d920e7
			ON (COALESCE("ad_bids"."pub", 'non_nulle9c9dae4c90746978d24a838c88b9879') = COALESCE(self_join3ba680fe589e49ceabf404f6c6d920e7."pub", 'non_nulle9c9dae4c90746978d24a838c88b9879'))
		)
		ORDER BY "pub" NULLS LAST
		OFFSET 0
	*/

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
		extraWhereClause,                      // 14
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
	args = append(args, extraWhereClauseArgs...)

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
		// The return type of date_trunc('month', ...) is Date so need another TIMESTAMP cast
		return fmt.Sprintf("toTimezone(date_trunc('%s', toTimezone(%s::TIMESTAMP, ?))::TIMESTAMP, ?)", dialect.ConvertToDateTruncSpecifier(dim.TimeGrain), col), []any{dim.TimeZone, dim.TimeZone}, nil
	case drivers.DialectPinot:
		// ToDateTime format truncates millis to secs because we don't support that, for example timeseries api does timestamppb.New(ts) which truncates to seconds
		return fmt.Sprintf("ToDateTime(date_trunc('%s', %s, 'MILLISECONDS', ?), 'yyyy-MM-dd''T''HH:mm:ss''Z''')", dialect.ConvertToDateTruncSpecifier(dim.TimeGrain), col), []any{dim.TimeZone}, nil
	default:
		return "", nil, fmt.Errorf("unsupported dialect %q", dialect)
	}
}

func (q *MetricsViewAggregation) buildOffsetExpression(col string, timeGrain runtimev1.TimeGrain, dialect drivers.Dialect) (string, error) {
	if timeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		return "0 as t_offset", nil
	}

	timeCol, err := q.trancationExpression(safeName(col), timeGrain, dialect)
	if err != nil {
		return "", err
	}
	start, _ := q.trancationExpression("?", timeGrain, dialect)

	var timeOffsetColumn string
	if dialect == drivers.DialectDuckDB {
		timeOffsetColumn = fmt.Sprintf("epoch_ms(%s)-epoch_ms(%s) as t_offset", timeCol, start)
	} else if dialect == drivers.DialectDruid {
		timeOffsetColumn = fmt.Sprintf("timestamp_to_millis(%s)-timestamp_to_millis(%s) as t_offset", timeCol, start)
	} else if dialect == drivers.DialectClickHouse {
		timeOffsetColumn = fmt.Sprintf("toUnixTimestamp(%s)-toUnixTimestamp(%s) as t_offset", timeCol, start)
	} else {
		return "", fmt.Errorf("unsupported dialect %q", dialect)
	}
	return timeOffsetColumn, nil
}

func (q *MetricsViewAggregation) trancationExpression(s string, timeGrain runtimev1.TimeGrain, dialect drivers.Dialect) (string, error) {
	switch dialect {
	case drivers.DialectDuckDB:
		return fmt.Sprintf("date_trunc('%s', %s)::TIMESTAMP", dialect.ConvertToDateTruncSpecifier(timeGrain), s), nil
	case drivers.DialectDruid:
		return fmt.Sprintf("date_trunc('%s', CAST(%s AS TIMESTAMP))", dialect.ConvertToDateTruncSpecifier(timeGrain), s), nil
	case drivers.DialectClickHouse:
		return fmt.Sprintf("date_trunc('%s', %s)", dialect.ConvertToDateTruncSpecifier(timeGrain), s), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", dialect)
	}
}

func (q *MetricsViewAggregation) rewriteToMetricsViewQuery(mv *runtimev1.MetricsViewSpec) (*metricsview.Query, bool, error) {
	// Time offset-based comparison joins not supported yet
	if q.ComparisonTimeRange != nil && !isTimeRangeNil(q.ComparisonTimeRange) {
		for _, d := range q.Dimensions {
			if d.Name == mv.TimeDimension {
				return nil, false, nil
			}
		}
	}

	qry := &metricsview.Query{MetricsView: q.MetricsViewName}

	for _, d := range q.Dimensions {
		res := metricsview.Dimension{Name: d.Name}
		if d.Alias != "" {
			res.Name = d.Alias
		}
		if d.TimeZone != "" {
			qry.TimeZone = d.TimeZone
		}
		if d.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			res.Compute = &metricsview.DimensionCompute{
				TimeFloor: &metricsview.DimensionComputeTimeFloor{
					Dimension: d.Name,
					Grain:     metricsview.TimeGrainFromProto(d.TimeGrain),
				},
			}
		}
		qry.Dimensions = append(qry.Dimensions, res)
	}

	for _, m := range q.Measures {
		res := metricsview.Measure{Name: m.Name}
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			res.Compute = &metricsview.MeasureCompute{Count: true}
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			res.Compute = &metricsview.MeasureCompute{CountDistinct: &metricsview.MeasureComputeCountDistinct{
				Dimension: m.BuiltinMeasureArgs[0].GetStringValue(),
			}}
		}

		// Measure filters not supported yet
		if m.Filter != nil {
			return nil, false, nil
		}

		if m.Compute != nil {
			switch c := m.Compute.(type) {
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
				res.Compute = &metricsview.MeasureCompute{ComparisonValue: &metricsview.MeasureComputeComparisonValue{
					Measure: c.ComparisonValue.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
				res.Compute = &metricsview.MeasureCompute{ComparisonDelta: &metricsview.MeasureComputeComparisonDelta{
					Measure: c.ComparisonDelta.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
				res.Compute = &metricsview.MeasureCompute{ComparisonRatio: &metricsview.MeasureComputeComparisonRatio{
					Measure: c.ComparisonRatio.Measure,
				}}
			}
		}

		qry.Measures = append(qry.Measures, res)
	}

	qry.PivotOn = q.PivotOn

	for _, s := range q.Sort {
		qry.Sort = append(qry.Sort, metricsview.Sort{
			Name: s.Name,
			Desc: s.Desc,
		})
	}

	if q.TimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.TimeRange.Start != nil {
			res.Start = q.TimeRange.Start.AsTime()
		}
		if q.TimeRange.End != nil {
			res.End = q.TimeRange.End.AsTime()
		}
		res.IsoDuration = q.TimeRange.IsoDuration
		res.IsoOffset = q.TimeRange.IsoOffset
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.TimeRange.RoundToGrain)
		qry.TimeRange = res
	}

	if q.ComparisonTimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.ComparisonTimeRange.Start != nil {
			res.Start = q.ComparisonTimeRange.Start.AsTime()
		}
		if q.ComparisonTimeRange.End != nil {
			res.End = q.ComparisonTimeRange.End.AsTime()
		}
		res.IsoDuration = q.ComparisonTimeRange.IsoDuration
		res.IsoOffset = q.ComparisonTimeRange.IsoOffset
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.ComparisonTimeRange.RoundToGrain)
		qry.ComparisonTimeRange = res
	}

	if q.Filter != nil { // backwards backwards compatibility
		if q.Where != nil {
			return nil, false, fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	if q.Where != nil {
		qry.Where = rewriteToMetricsResolverExpression(q.Where)
	}

	if q.Having != nil {
		qry.Having = rewriteToMetricsResolverExpression(q.Having)
	}

	if q.Limit != nil {
		if *q.Limit == 0 {
			tmp := int64(100)
			q.Limit = &tmp
		}
		qry.Limit = q.Limit
	}

	if q.Offset != 0 {
		qry.Offset = &q.Offset
	}

	if len(q.PivotOn) > 0 {
		qry.PivotOn = q.PivotOn
	}

	qry.Label = q.Exporting

	return qry, true, nil
}
