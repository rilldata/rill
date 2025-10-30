package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/jsonval"
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
	streaming   bool
	security    *runtime.ResolvedSecurity
	priority    int

	olap        drivers.OLAPStore
	olapRelease func()
	instanceCfg drivers.InstanceConfig

	timestamps map[string]metricsview.TimestampsResult
}

// New creates a new Executor for the provided metrics view.
func New(ctx context.Context, rt *runtime.Runtime, instanceID string, mv *runtimev1.MetricsViewSpec, streaming bool, sec *runtime.ResolvedSecurity, priority int) (*Executor, error) {
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
		streaming:   streaming,
		security:    sec,
		priority:    priority,
		olap:        olap,
		olapRelease: release,
		instanceCfg: instanceCfg,
		timestamps:  make(map[string]metricsview.TimestampsResult),
	}, nil
}

// Close releases the resources held by the Executor.
func (e *Executor) Close() {
	e.olapRelease()
}

// CacheKey returns a cache key based on the executor's metrics view's cache key configuration.
// If ok is false, caching is disabled for the metrics view.
func (e *Executor) CacheKey(ctx context.Context) ([]byte, bool, error) {
	spec := e.metricsView
	// Cache is disabled for metrics views based on external table
	if (spec.CacheEnabled != nil && !*spec.CacheEnabled) || (spec.CacheEnabled == nil && e.streaming) {
		return nil, false, nil
	}

	if spec.CacheKeySql == "" {
		if !e.streaming {
			// for metrics views on rill managed tables, we can cache forever
			// (until the metrics view is refreshed/edited, which always leads to cache invalidations)
			return []byte(""), true, nil
		}
		// watermark is the default cache key for streaming metrics views, use default mv time dimension
		ts, err := e.Timestamps(ctx, "")
		if err != nil {
			return nil, false, err
		}
		return []byte(ts.Watermark.Format(time.RFC3339)), true, nil
	}

	res, err := e.olap.Query(ctx, &drivers.Statement{
		Query:    spec.CacheKeySql,
		Priority: e.priority,
	})
	if err != nil {
		return nil, false, err
	}
	defer res.Close()
	var key any
	if res.Next() {
		if err := res.Scan(&key); err != nil {
			return nil, false, err
		}

		key, err = jsonval.ToValue(key, res.Schema.Fields[0].Type)
		if err != nil {
			return nil, false, err
		}
	}
	if err := res.Err(); err != nil {
		return nil, false, err
	}

	keyBytes, err := json.Marshal(key)
	if err != nil {
		return nil, false, err
	}
	return keyBytes, true, nil
}

// ValidateQuery validates the provided query against the executor's metrics view.
func (e *Executor) ValidateQuery(qry *metricsview.Query) error {
	// TODO: Implement it
	panic("not implemented")
}

// Timestamps queries min, max and watermark for the metrics view.
func (e *Executor) Timestamps(ctx context.Context, timeDim string) (metricsview.TimestampsResult, error) {
	if timeDim == "" {
		timeDim = e.metricsView.TimeDimension
	}

	if res, ok := e.timestamps[timeDim]; ok && !res.Min.IsZero() {
		return res, nil
	}

	timeExpr, err := e.timeColumnOrExpr(timeDim)
	if err != nil {
		return metricsview.TimestampsResult{}, fmt.Errorf("failed to resolve time column or expression: %w", err)
	}
	if timeExpr == "" {
		return metricsview.TimestampsResult{}, fmt.Errorf("no time dimension found in metrics view '%s'", timeDim)
	}

	var res metricsview.TimestampsResult
	switch e.olap.Dialect() {
	case drivers.DialectDuckDB:
		res, err = e.resolveDuckDB(ctx, timeExpr)
	case drivers.DialectClickHouse:
		res, err = e.resolveClickHouse(ctx, timeExpr)
	case drivers.DialectPinot:
		res, err = e.resolvePinot(ctx, timeExpr)
	case drivers.DialectDruid:
		res, err = e.resolveDruid(ctx, timeExpr)
	default:
		return metricsview.TimestampsResult{}, fmt.Errorf("not available for dialect '%s'", e.olap.Dialect())
	}
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}

	res.Now = time.Now()
	e.timestamps[timeDim] = res

	return res, nil
}

// BindQuery allows to set min, max and watermark from a cache.
func (e *Executor) BindQuery(ctx context.Context, qry *metricsview.Query, timestamps metricsview.TimestampsResult) error {
	err := qry.Validate()
	if err != nil {
		return err
	}

	if qry.TimeRange != nil && qry.TimeRange.TimeDimension != "" {
		e.timestamps[qry.TimeRange.TimeDimension] = timestamps
	} else if e.metricsView.TimeDimension != "" {
		e.timestamps[e.metricsView.TimeDimension] = timestamps
	}
	return e.rewriteQueryTimeRanges(ctx, qry, nil)
}

// Schema returns a schema for the metrics view's dimensions and measures.
func (e *Executor) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	if !e.security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	// Build a query that selects all dimensions and measures
	qry := &metricsview.Query{}

	if e.metricsView.TimeDimension != "" {
		qry.Dimensions = append(qry.Dimensions, metricsview.Dimension{
			Name: e.metricsView.TimeDimension,
			Compute: &metricsview.DimensionCompute{
				TimeFloor: &metricsview.DimensionComputeTimeFloor{
					Dimension: e.metricsView.TimeDimension,
					Grain:     metricsview.TimeGrainDay,
				},
			},
		})
	}

	for _, d := range e.metricsView.Dimensions {
		if e.security.CanAccessField(d.Name) {
			if e.metricsView.TimeDimension == d.Name {
				// Skip the time dimension if it is already added
				continue
			}
			qry.Dimensions = append(qry.Dimensions, metricsview.Dimension{Name: d.Name})
		}
	}

	for _, m := range e.metricsView.Measures {
		if e.security.CanAccessField(m.Name) {
			qry.Measures = append(qry.Measures, metricsview.Measure{Name: m.Name})
		}
	}

	// Setting both base and comparison time ranges in case there are time_comparison measures.
	if e.metricsView.TimeDimension != "" {
		now := time.Now()
		qry.TimeRange = &metricsview.TimeRange{
			Start: now.Add(-time.Second),
			End:   now,
		}
		qry.ComparisonTimeRange = &metricsview.TimeRange{
			Start: now.Add(-2 * time.Second),
			End:   now.Add(-time.Second),
		}
	}

	// Importantly, limit to 0 rows
	zero := int64(0)
	qry.Limit = &zero

	// Execute the query to get the schema
	ast, err := metricsview.NewAST(e.metricsView, e.security, qry, e.olap.Dialect())
	if err != nil {
		return nil, err
	}

	sql, args, err := ast.SQL()
	if err != nil {
		return nil, err
	}

	schema, err := e.olap.QuerySchema(ctx, sql, args)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

// Query executes the provided query against the metrics view.
func (e *Executor) Query(ctx context.Context, qry *metricsview.Query, executionTime *time.Time) (*drivers.Result, error) {
	if !e.security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	err := qry.Validate()
	if err != nil {
		return nil, err
	}

	// preserve the original limit, required in 2 phase comparison
	ogLimit := qry.Limit

	rowsCap, err := e.rewriteQueryEnforceCaps(qry)
	if err != nil {
		return nil, err
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

	ast, err := metricsview.NewAST(e.metricsView, e.security, qry, e.olap.Dialect())
	if err != nil {
		return nil, err
	}

	ok, err := e.rewriteTwoPhaseComparisons(ctx, qry, ast, ogLimit)
	if err != nil {
		return nil, err
	} // TODO if !ok then can log a warning that two phase comparison is not possible with a reason

	e.rewriteApproxComparisons(ast, ok)

	if err := e.rewriteLimitsIntoSubqueries(ast); err != nil {
		return nil, err
	}

	if err := e.rewriteDruidGroups(ast); err != nil {
		return nil, err
	}

	err = e.wrapClickhouseComputedTimeDim(ast)
	if err != nil {
		return nil, err
	}

	var res *drivers.Result
	if !pivoting {
		sql, args, err := ast.SQL()
		if err != nil {
			return nil, err
		}

		res, err = e.olap.Query(ctx, &drivers.Statement{
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
		path, err := e.executePivotExport(ctx, ast, pivotAST, "parquet", nil)
		if err != nil {
			return nil, err
		}

		// Use DuckDB to read the Parquet file into a *drivers.Result
		res, err = duck.Query(ctx, &drivers.Statement{
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
func (e *Executor) Export(ctx context.Context, qry *metricsview.Query, executionTime *time.Time, format drivers.FileFormat, headers []string) (string, error) {
	if !e.security.CanAccess() {
		return "", runtime.ErrForbidden
	}

	err := qry.Validate()
	if err != nil {
		return "", err
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

	ast, err := metricsview.NewAST(e.metricsView, e.security, qry, e.olap.Dialect())
	if err != nil {
		return "", err
	}

	e.rewriteApproxComparisons(ast, false)

	if err := e.rewriteLimitsIntoSubqueries(ast); err != nil {
		return "", err
	}

	if err := e.rewriteDruidGroups(ast); err != nil {
		return "", err
	}

	err = e.wrapClickhouseComputedTimeDim(ast)
	if err != nil {
		return "", err
	}

	if pivoting {
		return e.executePivotExport(ctx, ast, pivotAST, format, headers)
	}

	sql, args, err := ast.SQL()
	if err != nil {
		return "", err
	}

	return e.executeExport(ctx, format, e.metricsView.Connector, map[string]any{
		"sql":  sql,
		"args": args,
	}, headers)
}

// Search executes the provided query against the metrics view.
func (e *Executor) Search(ctx context.Context, qry *metricsview.SearchQuery, executionTime *time.Time) ([]metricsview.SearchResult, error) {
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
		q := &metricsview.Query{
			MetricsView:         qry.MetricsView,
			Dimensions:          []metricsview.Dimension{{Name: d}},
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
			Rows:                false,
		} //exhaustruct:enforce
		q.Where = whereExprForSearch(qry.Where, d, qry.Search)

		if err := e.rewriteQueryTimeRanges(ctx, q, executionTime); err != nil {
			return nil, err
		}

		rowsCap, err = e.rewriteQueryEnforceCaps(q)
		if err != nil {
			return nil, err
		}

		ast, err := metricsview.NewAST(e.metricsView, e.security, q, e.olap.Dialect())
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

	res, err := e.olap.Query(ctx, &drivers.Statement{
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
	searchResult := make([]metricsview.SearchResult, 0)
	for res.Next() {
		var row metricsview.SearchResult
		if err := res.Scan(&row.Dimension, &row.Value); err != nil {
			return nil, err
		}
		searchResult = append(searchResult, row)
	}
	err = res.Err()
	if err != nil {
		return nil, err
	}
	return searchResult, nil
}

// BindAnnotationsQuery allows setting min, max and watermark from a cache for an AnnotationsQuery
func (e *Executor) BindAnnotationsQuery(ctx context.Context, qry *metricsview.AnnotationsQuery, timestamps metricsview.TimestampsResult) error {
	if qry.TimeRange != nil && qry.TimeRange.TimeDimension != "" {
		e.timestamps[qry.TimeRange.TimeDimension] = timestamps
	} else if e.metricsView.TimeDimension != "" {
		e.timestamps[e.metricsView.TimeDimension] = timestamps
	}

	tz, err := time.LoadLocation(qry.TimeZone)
	if err != nil {
		return err
	}

	err = e.resolveTimeRange(ctx, qry.TimeRange, tz, nil)
	if err != nil {
		return err
	}

	return nil
}

func (e *Executor) Annotations(ctx context.Context, qry *metricsview.AnnotationsQuery) ([]map[string]any, error) {
	reqMeasures := qry.Measures
	if len(reqMeasures) == 0 {
		for _, mes := range e.metricsView.Measures {
			reqMeasures = append(reqMeasures, mes.Name)
		}
	}

	rows := make([]map[string]any, 0)

	for _, ann := range e.metricsView.Annotations {
		annMeasures := make([]string, 0)

		// Collect measures that are requested.
		for _, measure := range ann.Measures {
			if slices.Contains(reqMeasures, measure) {
				annMeasures = append(annMeasures, measure)
			}
		}

		// If none of the measures in the annotation are requested, skip the annotation.
		if len(annMeasures) == 0 {
			continue
		}

		rowsForAnn, err := e.executeAnnotationsQuery(ctx, qry, ann, annMeasures)
		if err != nil {
			return nil, err
		}

		rows = append(rows, rowsForAnn...)
	}

	return rows, nil
}

func (e *Executor) executeSearchInDruid(ctx context.Context, qry *metricsview.SearchQuery, executionTime *time.Time) ([]metricsview.SearchResult, error) {
	if qry.TimeRange == nil {
		return nil, errDruidNativeSearchUnimplemented
	}
	dimensions := make([]metricsview.Dimension, len(qry.Dimensions))
	for i, d := range qry.Dimensions {
		dimensions[i] = metricsview.Dimension{Name: d}
	}
	q := &metricsview.Query{
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
		Rows:                false,
	} //exhaustruct:enforce

	if err := e.rewriteQueryTimeRanges(ctx, q, executionTime); err != nil {
		return nil, err
	}

	a, err := metricsview.NewAST(e.metricsView, e.security, q, e.olap.Dialect())
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
		rows, err := e.olap.Query(ctx, &drivers.Statement{
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
		dim, err := a.LookupDimension(f.Name, true)
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
	req := druid.NewNativeSearchQueryRequest(e.metricsView.Table, qry.Search, dims, virtualCols, limit, a.Query.TimeRange.Start, a.Query.TimeRange.End, query)

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
	result := make([]metricsview.SearchResult, 0)
	for _, re := range res {
		for _, r := range re.Result {
			result = append(result, metricsview.SearchResult{
				Dimension: strings.TrimSuffix(r.Dimension, "_virtual_native"),
				Value:     r.Value,
			})
		}
	}
	return result, nil
}

// timeColumnOrExpr returns the time column or expression to use for the metrics view. ues time column if provided, otherwise fall back to the metrics view TimeDimension.
func (e *Executor) timeColumnOrExpr(timeDim string) (string, error) {
	// figure out the time column or expression to use from the dimension list
	for _, dim := range e.metricsView.Dimensions {
		if dim.Name == timeDim {
			expr, err := e.olap.Dialect().MetricsViewDimensionExpression(dim)
			if err != nil {
				return "", fmt.Errorf("failed to get time dimension expression for '%s': %w", timeDim, err)
			}
			return expr, nil
		}
	}
	return e.olap.Dialect().EscapeIdentifier(timeDim), nil // fallback to the time dimension if not found in dimensions
}

func (e *Executor) executeAnnotationsQuery(ctx context.Context, qry *metricsview.AnnotationsQuery, annotation *runtimev1.MetricsViewSpec_Annotation, forMeasures []string) ([]map[string]any, error) {
	// Acquire olap connection for the annotation's table's connector
	olap, release, err := e.rt.OLAP(ctx, e.instanceID, annotation.Connector)
	if err != nil {
		return nil, err
	}
	defer release()
	dialect := olap.Dialect()

	// Only call resolveTimeRange is either start/end was not provided.
	// This avoids executing Timestamps without caching if it was not bound already.
	if qry.TimeRange.Start.IsZero() || qry.TimeRange.End.IsZero() {
		tz, err := time.LoadLocation(qry.TimeZone)
		if err != nil {
			return nil, err
		}

		err = e.resolveTimeRange(ctx, qry.TimeRange, tz, nil)
		if err != nil {
			return nil, err
		}
	}

	start := qry.TimeRange.Start.Format(time.RFC3339)
	end := qry.TimeRange.End.Format(time.RFC3339)

	b := &strings.Builder{}

	b.WriteString("SELECT *")
	if annotation.HasDuration {
		// Convert the string grain to an integer so that it is easy to calculate "greater than or equal to".
		b.WriteString(`,(CASE
  WHEN duration = 'millisecond' THEN 1
  WHEN duration = 'second' THEN 2
  WHEN duration = 'minute' THEN 3
  WHEN duration = 'hour' THEN 4
  WHEN duration = 'day' THEN 5
  WHEN duration = 'week' THEN 6
  WHEN duration = 'month' THEN 7
  WHEN duration = 'quarter' THEN 8
  WHEN duration = 'year' THEN 9
  ELSE 0
END) as __rill_time_grain`)
	}

	b.WriteString(" FROM ")
	b.WriteString(dialect.EscapeTable(annotation.Database, annotation.DatabaseSchema, annotation.Table))

	b.WriteString(" WHERE ")

	b.WriteString("time >= ? AND time < ?")
	args := []any{start, end}

	if annotation.HasTimeEnd {
		b.WriteString(" AND time_end >= ? AND time_end < ?")
		args = append(args, start, end)
	}

	if annotation.HasDuration && qry.TimeGrain != metricsview.TimeGrainUnspecified {
		b.WriteString(" AND (__rill_time_grain == 0 OR __rill_time_grain <= ?)")
		args = append(args, int(qry.TimeGrain.ToTimeutil()))
	}

	b.WriteString(" ORDER BY time")

	if qry.Limit != nil {
		b.WriteString(" LIMIT ?")
		args = append(args, *qry.Limit)
	}

	if qry.Offset != nil {
		b.WriteString(" OFFSET ?")
		args = append(args, *qry.Offset)
	}

	res, err := olap.Query(ctx, &drivers.Statement{
		Query:    b.String(),
		Args:     args,
		Priority: 0,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	rows := make([]map[string]any, 0)

	for res.Next() {
		row := make(map[string]any)
		if err := res.MapScan(row); err != nil {
			return nil, err
		}

		// Fill in the for_measures field. Used which annotations apply to which measures.
		row["for_measures"] = forMeasures

		rows = append(rows, row)
	}

	err = res.Err()
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func whereExprForSearch(where *metricsview.Expression, dimension, search string) *metricsview.Expression {
	if where == nil {
		return &metricsview.Expression{
			Condition: &metricsview.Condition{
				Operator: metricsview.OperatorIlike,
				Expressions: []*metricsview.Expression{
					{Name: dimension},
					{Value: fmt.Sprintf("%%%s%%", search)},
				},
			},
		}
	}
	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: metricsview.OperatorAnd,
			Expressions: []*metricsview.Expression{
				{
					Condition: &metricsview.Condition{
						Operator: metricsview.OperatorIlike,
						Expressions: []*metricsview.Expression{
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
