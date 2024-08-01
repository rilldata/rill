package metricsview

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/druid"
)

type SearchQuery struct {
	MetricsView string     `mapstructure:"metrics_view"`
	Dimensions  []string   `mapstructure:"dimensions"`
	Search      string     `mapstructure:"search"`
	TimeRange   *TimeRange `mapstructure:"time_range"`
}

type SearchResult struct {
	Dimension string
	Value     any
}

func (s *SearchQuery) searchSQL(a *AST) (string, []interface{}, error) {
	// TODO add validation of AST
	var b strings.Builder
	var args []any

	for i, f := range a.Root.DimFields {
		if i > 0 {
			b.WriteString(" UNION ALL ")
		}

		b.WriteString("SELECT ")
		b.WriteByte('(')
		b.WriteString(f.Expr)
		b.WriteString(") AS value,")
		b.WriteString(fmt.Sprintf(" '%s'", f.Name))
		b.WriteString(" AS dimension")

		b.WriteString(" FROM ")
		b.WriteString(*a.Root.FromTable)
		for _, u := range a.Root.Unnests {
			if u.DimName == f.Name {
				b.WriteString(", ")
				b.WriteString(u.Expr)
			}
		}

		b.WriteString(" WHERE ")
		b.WriteByte('(')
		b.WriteString(f.Expr)
		b.WriteString(") ILIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", s.Search))

		if a.Root.Where != nil && a.Root.Where.Expr != "" {
			b.WriteString(" AND (")
			b.WriteString(a.Root.Where.Expr)
			b.WriteByte(')')
			args = append(args, a.Root.Where.Args...)
		}

		if a.Root.TimeWhere != nil && a.Root.TimeWhere.Expr != "" {
			b.WriteString(" AND (")
			b.WriteString(a.Root.TimeWhere.Expr)
			b.WriteByte(')')
			args = append(args, a.Root.TimeWhere.Args...)
		}

		b.WriteString(" GROUP BY value")
	}

	return b.String(), args, nil
}

var druidSQLDSN = regexp.MustCompile(`/v2/sql/?`)

func (q *SearchQuery) executeSearchInDruid(ctx context.Context, olap drivers.OLAPStore, a *AST) ([]SearchResult, error) {
	var query map[string]interface{}
	if a.Root.Where != nil {
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

		if !rows.Next() {
			return nil, fmt.Errorf("failed to parse filter")
		}

		var planRaw string
		var resRaw string
		var attrRaw string
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
	dsn, err := druid.GetDSN(olap.(drivers.Handle).Config())
	if err != nil {
		return nil, err
	}
	if dsn == "" {
		return nil, fmt.Errorf("druid connector config not found in instance")
	}

	nq := druid.NewNativeQuery(druidSQLDSN.ReplaceAllString(dsn, "/v2/"))
	req := druid.NewNativeSearchQueryRequest(trimQuotes(*a.Root.FromTable), q.Search, q.Dimensions, q.TimeRange.Start, q.TimeRange.End, query) // TODO: timestamps may be nil!
	var res druid.NativeSearchQueryResponse
	err = nq.Do(ctx, req, &res, req.Context.QueryID)
	if err != nil {
		return nil, err
	}
	fmt.Printf("native query response %v\n", res)

	result := make([]SearchResult, len(res))
	for _, re := range res {
		for _, r := range re.Result {
			result = append(result, SearchResult{
				Dimension: r.Dimension,
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
