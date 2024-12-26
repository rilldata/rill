package metricsview

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid"
)

const (
	defaultInteractiveTimeout = time.Minute * 3
	defaultExportTimeout      = time.Minute * 5
	defaultPivotExportTimeout = time.Minute * 5
)

// Executor is capable of executing queries and other operations against a metrics view.
type Executor struct {
	rt          *runtime.Runtime
	instanceID  string
	metricsView *runtimev1.MetricsViewSpec
	security    *runtime.ResolvedSecurity
	priority    int

	olap        drivers.OLAPStore
	olapRelease func()
	instanceCfg drivers.InstanceConfig

	watermark time.Time
}

// NewExecutor creates a new Executor for the provided metrics view.
func NewExecutor(ctx context.Context, rt *runtime.Runtime, instanceID string, mv *runtimev1.MetricsViewSpec, sec *runtime.ResolvedSecurity, priority int) (*Executor, error) {
	olap, release, err := rt.OLAP(ctx, instanceID, mv.Connector)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connector for metrics view: %w", err)
	}

	instanceCfg, err := rt.InstanceConfig(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	return &Executor{
		rt:          rt,
		instanceID:  instanceID,
		metricsView: mv,
		security:    sec,
		priority:    priority,
		olap:        olap,
		olapRelease: release,
		instanceCfg: instanceCfg,
	}, nil
}

// Close releases the resources held by the Executor.
func (e *Executor) Close() {
	e.olapRelease()
}

// Cacheable returns whether the result of running the given query is cacheable.
func (e *Executor) Cacheable(qry *Query) bool {
	// TODO: Get from OLAP instead of hardcoding
	return e.olap.Dialect() == drivers.DialectDuckDB
}

// ValidateQuery validates the provided query against the executor's metrics view.
func (e *Executor) ValidateQuery(qry *Query) error {
	// TODO: Implement it
	panic("not implemented")
}

// Watermark returns the current watermark of the metrics view.
// If the watermark resolves to null, it defaults to the current time.
func (e *Executor) Watermark(ctx context.Context) (time.Time, error) {
	return e.loadWatermark(ctx, nil)
}

// Schema returns a schema for the metrics view's dimensions and measures.
func (e *Executor) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	if !e.security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	// Build a query that selects all dimensions and measures
	qry := &Query{}

	if e.metricsView.TimeDimension != "" {
		qry.Dimensions = append(qry.Dimensions, Dimension{
			Name: e.metricsView.TimeDimension,
			Compute: &DimensionCompute{
				TimeFloor: &DimensionComputeTimeFloor{
					Dimension: e.metricsView.TimeDimension,
					Grain:     TimeGrainDay,
				},
			},
		})
	}

	for _, d := range e.metricsView.Dimensions {
		if e.security.CanAccessField(d.Name) {
			qry.Dimensions = append(qry.Dimensions, Dimension{Name: d.Name})
		}
	}

	for _, m := range e.metricsView.Measures {
		if e.security.CanAccessField(m.Name) {
			qry.Measures = append(qry.Measures, Measure{Name: m.Name})
		}
	}

	// Setting both base and comparison time ranges in case there are time_comparison measures.
	if e.metricsView.TimeDimension != "" {
		now := time.Now()
		qry.TimeRange = &TimeRange{
			Start: now.Add(-time.Second),
			End:   now,
		}
		qry.ComparisonTimeRange = &TimeRange{
			Start: now.Add(-2 * time.Second),
			End:   now.Add(-time.Second),
		}
	}

	// Importantly, limit to 0 rows
	zero := int64(0)
	qry.Limit = &zero

	// Execute the query to get the schema
	ast, err := NewAST(e.metricsView, e.security, qry, e.olap.Dialect())
	if err != nil {
		return nil, err
	}

	sql, args, err := ast.SQL()
	if err != nil {
		return nil, err
	}

	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         e.priority,
		ExecutionTimeout: defaultInteractiveTimeout,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return res.Schema, nil
}

// Query executes the provided query against the metrics view.
func (e *Executor) Query(ctx context.Context, qry *Query, executionTime *time.Time) (*drivers.Result, error) {
	if !e.security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	rowsCap, err := e.rewriteQueryEnforceCaps(qry)
	if err != nil {
		return nil, err
	}

	if qry.ComparisonTimeRange != nil && len(qry.Dimensions) == 1 && len(qry.Sort) == 1 {
		// TODO  need to handle sorting by dimension ?
		// perform two phase query - first for base time range and then for comparison time range and then merge the results

		// Build a query for the base time range
		baseQry := &Query{
			MetricsView:         qry.MetricsView,
			Dimensions:          qry.Dimensions,
			Measures:            qry.Measures,
			PivotOn:             qry.PivotOn,
			Spine:               qry.Spine,
			Sort:                qry.Sort,
			TimeRange:           qry.TimeRange,
			ComparisonTimeRange: nil,
			Where:               qry.Where,
			Having:              qry.Having,
			Limit:               qry.Limit,
			Offset:              qry.Offset,
			TimeZone:            qry.TimeZone,
			UseDisplayNames:     false,
		}

		for i, m := range baseQry.Measures {
			if m.Compute != nil {
				if m.Compute.ComparisonValue != nil || m.Compute.ComparisonDelta != nil || m.Compute.ComparisonRatio != nil {
					baseQry.Measures[i].Compute.Constant = true
				}
			}
		}

		// Execute the query for the base time range
		baseRes, err := e.Query(ctx, baseQry, executionTime)
		if err != nil {
			return nil, err
		}
		defer baseRes.Close()

		// Build a query for the comparison time range, use results from the base time range to add as filter in the comparison time range query
		compQry := &Query{
			MetricsView:         qry.MetricsView,
			Dimensions:          qry.Dimensions,
			Measures:            qry.Measures,
			PivotOn:             qry.PivotOn,
			Spine:               qry.Spine,
			Sort:                qry.Sort,
			TimeRange:           qry.ComparisonTimeRange,
			ComparisonTimeRange: nil,
			Where:               qry.Where,
			Having:              qry.Having,
			Limit:               nil,
			Offset:              nil,
			TimeZone:            qry.TimeZone,
			UseDisplayNames:     false,
		}

		for i, m := range compQry.Measures {
			if m.Compute != nil {
				if m.Compute.ComparisonValue != nil || m.Compute.ComparisonDelta != nil || m.Compute.ComparisonRatio != nil {
					baseQry.Measures[i].Compute.Constant = true
				}
			}
		}

		// Extract the dimension values returned from the inner query.
		baseAgg := make(map[any]ComparisonMeasures)
		//var dimVals []any
		//var measureVals []any
		for baseRes.Next() {
			var dim, measure, prev, delta, deltaP any
			if err := baseRes.Scan(&dim, &measure, &prev, &delta, &deltaP); err != nil {
				return nil, fmt.Errorf("two phase comparison: base query failed to scan value: %w", err)
			}
			//dimVals = append(dimVals, dim)
			//measureVals = append(measureVals, measure)
			if dim == nil {
				baseAgg[nil] = ComparisonMeasures{base: measure, prev: prev, delta: delta, deltaP: deltaP}
			} else {
				d := dim.(string)
				baseAgg[d] = ComparisonMeasures{base: measure, prev: prev, delta: delta, deltaP: deltaP}
			}
		}

		// Add the dimensions values as a "<dim> IN (<vals...>)" expression in the outer query's WHERE clause.
		var inExpr *Expression
		if len(baseAgg) == 0 {
			inExpr = &Expression{
				Value: false,
			}
		} else {
			// if any dim value is nil add condition with eq operator with nil value
			var vals []any
			foundNil := false
			for k := range baseAgg {
				if k == nil {
					foundNil = true
				} else {
					vals = append(vals, k)
				}
			}
			inExpr = &Expression{
				Condition: &Condition{
					Operator: OperatorIn,
					Expressions: []*Expression{
						{Name: qry.Dimensions[0].Name},
						{Value: vals},
					},
				},
			}
			if foundNil {
				inExpr = &Expression{
					Condition: &Condition{
						Operator: OperatorOr,
						Expressions: []*Expression{
							inExpr,
							{
								Condition: &Condition{
									Operator: OperatorEq,
									Expressions: []*Expression{
										{Name: qry.Dimensions[0].Name},
										{Value: nil},
									},
								},
							},
						},
					},
				}
			}
		}

		if compQry.Where == nil {
			compQry.Where = inExpr
		} else {
			compQry.Where = &Expression{
				Condition: &Condition{
					Operator: OperatorAnd,
					Expressions: []*Expression{
						compQry.Where,
						inExpr,
					},
				},
			}
		}

		// Execute the query for the comparison time range
		compRes, err := e.Query(ctx, compQry, executionTime)
		if err != nil {
			return nil, err
		}
		defer compRes.Close()

		compAgg := make(map[any]ComparisonMeasures)
		for compRes.Next() {
			var dim, measure, prev, delta, deltaP any
			if err := compRes.Scan(&dim, &measure, &prev, &delta, &deltaP); err != nil {
				return nil, fmt.Errorf("two phase comparison: base query failed to scan value: %w", err)
			}
			//dimVals = append(dimVals, dim)
			//measureVals = append(measureVals, measure)
			if dim == nil {
				compAgg[nil] = ComparisonMeasures{base: measure, prev: prev, delta: delta, deltaP: deltaP}
			} else {
				d := dim.(string)
				compAgg[d] = ComparisonMeasures{base: measure, prev: prev, delta: delta, deltaP: deltaP}
			}
		}

		// create select query with inlined base and comparison measures
		finalQry := ""
		measureName := qry.Measures[0].Name
		for dim, baseMeasures := range baseAgg {
			compMeasures := compAgg[dim]
			if compMeasures.base == nil {
				compMeasures.base = 0
			}
			if dim != nil {
				finalQry += fmt.Sprintf("SELECT '%[1]s' AS %[2]s, %[3]v AS %[4]s, %[5]v AS %[4]s_prev, %[3]v-%[5]v AS %[4]s_delta, (%[3]v-%[5]v)/%[3]v AS %[4]s_perc UNION ALL ", dim, qry.Dimensions[0].Name, baseMeasures.base, measureName, compMeasures.base)
			} else {
				finalQry += fmt.Sprintf("SELECT NULL AS %[1]s, %[2]v AS %[3]s, %[4]v AS %[3]s_prev, %[2]v-%[4]v AS %[3]s_delta, (%[2]v-%[4]v)/%[2]v AS %[3]s_perc UNION ALL ", qry.Dimensions[0].Name, baseMeasures.base, measureName, compMeasures.base)
			}
		}

		// remove last UNION ALL
		finalQry = finalQry[:len(finalQry)-10]

		// TODO: union all with order by - query planning failing on druid
		// add order by clause
		/*finalQry += fmt.Sprintf(" ORDER BY %s", qry.Sort[0].Name)
		if qry.Sort[0].Desc {
			finalQry += " DESC"
		} else {
			finalQry += " ASC"
		}*/

		// Execute the final query
		res, err := e.olap.Execute(ctx, &drivers.Statement{
			Query:    finalQry,
			Priority: e.priority,
		})
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	pivotAST, pivoting, err := e.rewriteQueryForPivot(qry)
	if err != nil {
		return nil, err
	}

	if err := e.rewriteQueryTimeRanges(ctx, qry, executionTime); err != nil {
		return nil, err
	}

	if err := e.rewritePercentOfTotals(ctx, qry); err != nil {
		return nil, err
	}

	if err := e.rewriteQueryDruidExactify(ctx, qry); err != nil {
		return nil, err
	}

	ast, err := NewAST(e.metricsView, e.security, qry, e.olap.Dialect())
	if err != nil {
		return nil, err
	}

	e.rewriteApproxComparisons(ast)

	if err := e.rewriteLimitsIntoSubqueries(ast); err != nil {
		return nil, err
	}

	if err := e.rewriteDruidGroups(ast); err != nil {
		return nil, err
	}

	var res *drivers.Result
	if !pivoting {
		sql, args, err := ast.SQL()
		if err != nil {
			return nil, err
		}

		res, err = e.olap.Execute(ctx, &drivers.Statement{
			Query:            sql,
			Args:             args,
			Priority:         e.priority,
			ExecutionTimeout: defaultInteractiveTimeout,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Since pivots are mainly used for exports, we just do an inefficient shim that runs a pivoted export to a temporary Parquet file, and then reads the file into a *drivers.Result using DuckDB.
		// (An efficient interactive pivot implementation would look quite different from the export-based implementation, so is not worth it at this point.)

		// If e.olap is a DuckDB, use it directly. Else open a "duckdb" handle (which is always available, even for instances where DuckDB is not the main OLAP connector).
		var duck drivers.OLAPStore
		var releaseDuck func()
		if e.olap.Dialect() == drivers.DialectDuckDB {
			duck = e.olap
		} else {
			handle, release, err := e.rt.AcquireHandle(ctx, e.instanceID, "duckdb")
			if err != nil {
				return nil, fmt.Errorf("failed to acquire DuckDB for serving pivot: %w", err)
			}

			var ok bool
			duck, ok = handle.AsOLAP(e.instanceID)
			if !ok {
				release()
				return nil, fmt.Errorf(`connector "duckdb" is not an OLAP store`)
			}
			releaseDuck = release
		}

		// Execute the pivot export
		path, err := e.executePivotExport(ctx, ast, pivotAST, "parquet")
		if err != nil {
			return nil, err
		}

		// Use DuckDB to read the Parquet file into a *drivers.Result
		res, err = duck.Execute(ctx, &drivers.Statement{
			Query:            fmt.Sprintf("SELECT * FROM '%s'", path),
			Priority:         e.priority,
			ExecutionTimeout: defaultInteractiveTimeout,
		})
		if err != nil {
			_ = os.Remove(path)
			return nil, err
		}
		res.SetCleanupFunc(func() error {
			if releaseDuck != nil {
				releaseDuck()
			}
			_ = os.Remove(path)
			return nil
		})
	}

	if rowsCap > 0 {
		res.SetCap(rowsCap)
	}

	return res, nil
}

// Export executes and exports the provided query against the metrics view.
// It returns a path to a temporary file containing the export. The caller is responsible for cleaning up the file.
func (e *Executor) Export(ctx context.Context, qry *Query, executionTime *time.Time, format drivers.FileFormat) (string, error) {
	if !e.security.CanAccess() {
		return "", runtime.ErrForbidden
	}

	pivotAST, pivoting, err := e.rewriteQueryForPivot(qry)
	if err != nil {
		return "", err
	}

	if err := e.rewriteQueryTimeRanges(ctx, qry, executionTime); err != nil {
		return "", err
	}

	if err := e.rewritePercentOfTotals(ctx, qry); err != nil {
		return "", err
	}

	if err := e.rewriteQueryDruidExactify(ctx, qry); err != nil {
		return "", err
	}

	ast, err := NewAST(e.metricsView, e.security, qry, e.olap.Dialect())
	if err != nil {
		return "", err
	}

	e.rewriteApproxComparisons(ast)

	if err := e.rewriteLimitsIntoSubqueries(ast); err != nil {
		return "", err
	}

	if err := e.rewriteDruidGroups(ast); err != nil {
		return "", err
	}

	if pivoting {
		return e.executePivotExport(ctx, ast, pivotAST, format)
	}

	sql, args, err := ast.SQL()
	if err != nil {
		return "", err
	}

	return e.executeExport(ctx, format, e.metricsView.Connector, map[string]any{
		"sql":  sql,
		"args": args,
	})
}

type ComparisonMeasures struct {
	base   any
	prev   any
	delta  any
	deltaP any
}

type SearchQuery struct {
	MetricsView string      `mapstructure:"metrics_view"`
	Dimensions  []string    `mapstructure:"dimensions"`
	Search      string      `mapstructure:"search"`
	Where       *Expression `mapstructure:"where"`
	Having      *Expression `mapstructure:"having"`
	TimeRange   *TimeRange  `mapstructure:"time_range"`
	Limit       *int64      `mapstructure:"limit"`
}

type SearchResult struct {
	Dimension string
	Value     any
}

// SearchQuery executes the provided query against the metrics view.
func (e *Executor) Search(ctx context.Context, qry *SearchQuery, executionTime *time.Time) ([]SearchResult, error) {
	if !e.security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	// Generate a metricsview.Query and build a AST
	// This is a hacky implementation since both metricsview.Query and AST are designed for aggregate queries.
	// TODO :: Refactor the code and extract common functionality from metricsview.Query and AST and write SearchQuery to underlying SQL/Native druid query directly.

	if e.olap.Dialect() == drivers.DialectDruid {
		// native search
		res, err := e.executeSearchInDruid(ctx, qry, executionTime)
		if err == nil || !errors.Is(err, errDruidNativeSearchUnimplemented) {
			return res, err
		}
	}

	var (
		finalSQL  strings.Builder
		finalArgs []any
		rowsCap   int64
		err       error
	)
	for i, d := range qry.Dimensions {
		if i > 0 {
			finalSQL.WriteString(" UNION ALL ")
		}
		q := &Query{
			MetricsView:         qry.MetricsView,
			Dimensions:          []Dimension{{Name: d}},
			Measures:            nil,
			PivotOn:             nil,
			Spine:               nil,
			Sort:                nil,
			TimeRange:           qry.TimeRange,
			ComparisonTimeRange: nil,
			Where:               nil,
			Having:              qry.Having,
			Limit:               qry.Limit,
			Offset:              nil,
			TimeZone:            "",
			UseDisplayNames:     false,
		} //exhaustruct:enforce
		q.Where = whereExprForSearch(qry.Where, d, qry.Search)

		if err := e.rewriteQueryTimeRanges(ctx, q, executionTime); err != nil {
			return nil, err
		}

		rowsCap, err = e.rewriteQueryEnforceCaps(q)
		if err != nil {
			return nil, err
		}

		ast, err := NewAST(e.metricsView, e.security, q, e.olap.Dialect())
		if err != nil {
			return nil, err
		}

		if err := e.rewriteLimitsIntoSubqueries(ast); err != nil {
			return nil, err
		}

		sql, args, err := ast.SQL()
		if err != nil {
			return nil, err
		}
		finalSQL.WriteString(fmt.Sprintf("SELECT %s AS dimension, %s AS value FROM (%s)", e.olap.Dialect().EscapeStringValue(d), e.olap.Dialect().EscapeIdentifier(d), sql))
		finalArgs = append(finalArgs, args...)
	}

	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:            finalSQL.String(),
		Args:             finalArgs,
		Priority:         e.priority,
		ExecutionTimeout: defaultInteractiveTimeout,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if rowsCap > 0 {
		res.SetCap(rowsCap)
	}
	searchResult := make([]SearchResult, 0)
	for res.Next() {
		var row SearchResult
		if err := res.Scan(&row.Dimension, &row.Value); err != nil {
			return nil, err
		}
		searchResult = append(searchResult, row)
	}
	if res.Err() != nil {
		return nil, res.Err()
	}
	return searchResult, nil
}

func (e *Executor) executeSearchInDruid(ctx context.Context, qry *SearchQuery, executionTime *time.Time) ([]SearchResult, error) {
	if qry.TimeRange == nil {
		return nil, errDruidNativeSearchUnimplemented
	}
	dimensions := make([]Dimension, len(qry.Dimensions))
	for i, d := range qry.Dimensions {
		dimensions[i] = Dimension{Name: d}
	}
	q := &Query{
		MetricsView:         qry.MetricsView,
		Dimensions:          dimensions,
		Measures:            nil,
		PivotOn:             nil,
		Spine:               nil,
		Sort:                nil,
		TimeRange:           qry.TimeRange,
		ComparisonTimeRange: nil,
		Where:               qry.Where,
		Having:              qry.Having,
		Limit:               qry.Limit,
		Offset:              nil,
		TimeZone:            "",
		UseDisplayNames:     false,
	} //exhaustruct:enforce

	if err := e.rewriteQueryTimeRanges(ctx, q, executionTime); err != nil {
		return nil, err
	}

	a, err := NewAST(e.metricsView, e.security, q, e.olap.Dialect())
	if err != nil {
		return nil, err
	}

	if err := e.rewriteLimitsIntoSubqueries(a); err != nil {
		return nil, err
	}

	if a.Root.FromSelect != nil {
		// This means either the dimension uses an unnest or measure filters which are not directly supported by native search.
		// This can be supported in future using query datasource in future if performance turns out to be a concern.
		return nil, errDruidNativeSearchUnimplemented
	}
	var query map[string]interface{}
	if a.Root.Where != nil {
		// NOTE :: this does not work for measure filters.
		// The query planner resolves them to joins instead of filters.
		rows, err := e.olap.Execute(ctx, &drivers.Statement{
			Query:            fmt.Sprintf("EXPLAIN PLAN FOR SELECT 1 FROM %s WHERE %s", *a.Root.FromTable, a.Root.Where.Expr),
			Args:             a.Root.Where.Args,
			Priority:         e.priority,
			ExecutionTimeout: defaultInteractiveTimeout,
		})
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, fmt.Errorf("failed to parse filter")
		}

		var (
			planRaw string
			resRaw  string
			attrRaw string
		)
		err = rows.Scan(&planRaw, &resRaw, &attrRaw)
		if err != nil {
			return nil, err
		}

		var plan []druid.QueryPlan
		err = json.Unmarshal([]byte(planRaw), &plan)
		if err != nil {
			return nil, err
		}

		if len(plan) == 0 {
			return nil, fmt.Errorf("failed to parse policy filter")
		}
		if plan[0].Query.Filter == nil {
			// if we failed to parse a filter we return and run UNION query.
			// this can happen when the row filter is complex
			// TODO: iterate over this and integrate more parts like joins and subfilter in policy filter
			return nil, errDruidNativeSearchUnimplemented
		}
		query = *plan[0].Query.Filter
	}

	// Build a native query
	limit := 100
	if a.Root.Limit != nil {
		limit = int(*a.Root.Limit)
	}
	dims := make([]string, 0)
	virtualCols := make([]druid.NativeVirtualColumns, 0)
	for _, f := range a.Root.DimFields {
		dim, err := a.lookupDimension(f.Name, true)
		if err != nil {
			return nil, err
		}
		// if the dimension is a expression we need a virtual column that can be scanned in SearchDimensions
		if dim.Expression != "" {
			virtualCols = append(virtualCols, druid.NativeVirtualColumns{
				Type:       "expression",
				Name:       fmt.Sprintf("%v_virtual_native", f.Name), // The name of the virtual column should not clash with actual column
				Expression: dim.Expression,
			})
			dims = append(dims, fmt.Sprintf("%v_virtual_native", f.Name))
		} else {
			dims = append(dims, f.Name)
		}
	}
	req := druid.NewNativeSearchQueryRequest(e.metricsView.Table, qry.Search, dims, virtualCols, limit, a.query.TimeRange.Start, a.query.TimeRange.End, query)

	// Execute the native query
	client, err := druid.NewNativeClient(e.olap)
	if err != nil {
		return nil, err
	}
	res, err := client.Search(ctx, &req)
	if err != nil {
		return nil, err
	}

	// Convert the response to a SearchResult
	result := make([]SearchResult, 0)
	for _, re := range res {
		for _, r := range re.Result {
			result = append(result, SearchResult{
				Dimension: strings.TrimSuffix(r.Dimension, "_virtual_native"),
				Value:     r.Value,
			})
		}
	}
	return result, nil
}

func whereExprForSearch(where *Expression, dimension, search string) *Expression {
	if where == nil {
		return &Expression{
			Condition: &Condition{
				Operator: OperatorIlike,
				Expressions: []*Expression{
					{Name: dimension},
					{Value: fmt.Sprintf("%%%s%%", search)},
				},
			},
		}
	}
	return &Expression{
		Condition: &Condition{
			Operator: OperatorAnd,
			Expressions: []*Expression{
				{
					Condition: &Condition{
						Operator: OperatorIlike,
						Expressions: []*Expression{
							{Name: dimension},
							{Value: fmt.Sprintf("%%%s%%", search)},
						},
					},
				},
				where,
			},
		},
	}
}

var errDruidNativeSearchUnimplemented = fmt.Errorf("native search is not implemented")
