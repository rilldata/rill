package metricsview

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid"
)

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

func (q *SearchQuery) executeSearchInDruid(ctx context.Context, olap drivers.OLAPStore, a *AST) ([]SearchResult, error) {
	if a.Root.FromSelect != nil {
		// This means either the dimension uses an unnest or measure filters which are not directly supported by native search.
		// This can be supported in future using query datasource in future if performance turns out to be a concern.
		return nil, errDruidNativeSearchUnimplemented
	}
	var query map[string]interface{}
	if a.Root.Where != nil {
		// NOTE :: this does not work for measure filters.
		// The query planner resolves them to joins instead of filters.
		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:            fmt.Sprintf("EXPLAIN PLAN FOR SELECT 1 FROM %s WHERE %s", *a.Root.FromTable, a.Root.Where.Expr),
			Args:             a.Root.Where.Args,
			DryRun:           false,
			Priority:         0,
			LongRunning:      false,
			ExecutionTimeout: 0,
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
			dims = append(dims, trimQuotes(f.Expr))
		}
	}
	req := druid.NewNativeSearchQueryRequest(trimQuotes(*a.Root.FromTable), q.Search, dims, virtualCols, limit, q.TimeRange.Start, q.TimeRange.End, query) // TODO: timestamps may be nil!

	// Execute the native query
	client, err := druid.NewNativeClient(olap)
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

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

var errDruidNativeSearchUnimplemented = fmt.Errorf("native search is not implemented")
