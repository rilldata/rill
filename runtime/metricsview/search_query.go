package metricsview

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

type searchSQLBuilder struct {
	ast    *AST
	search string
	out    *strings.Builder
	args   []any
}

func searchSQL(ast *AST, search string) (string, []any, error) {
	b := &searchSQLBuilder{
		ast:    ast,
		search: search,
		out:    &strings.Builder{},
	}

	err := b.writeSelect(ast.Root)
	if err != nil {
		return "", nil, err
	}

	return b.out.String(), b.args, nil
}

func (b *searchSQLBuilder) writeSelect(n *SelectNode) error {
	for i, f := range n.DimFields {
		if i > 0 {
			b.out.WriteString(" UNION ALL ")
		}

		b.out.WriteString("SELECT ")
		b.out.WriteByte('(')
		b.out.WriteString(f.Expr)
		b.out.WriteString(") AS value,")
		fmt.Fprintf(b.out, " '%s'", f.Name)
		b.out.WriteString(" AS dimension")

		b.out.WriteString(" FROM ")

		if n.FromTable != nil {
			b.out.WriteString(*n.FromTable)

			// Add unnest joins. We only and always apply these against FromTable (ensuring they are already unnested when referenced in outer SELECTs).
			for _, u := range n.Unnests {
				if u.DimName == f.Name {
					b.out.WriteString(", ")
					b.out.WriteString(u.Expr)
				}
			}
		} else if n.FromSelect != nil {
			if !n.FromSelect.IsCTE {
				b.out.WriteByte('(')
				err := b.writeSelect(n.FromSelect)
				if err != nil {
					return err
				}
				b.out.WriteString(") ")
			}
			b.out.WriteString(n.FromSelect.Alias)

			for _, ljs := range n.LeftJoinSelects {
				err := b.writeJoin("LEFT", n.FromSelect, ljs)
				if err != nil {
					return err
				}
			}
		} else {
			panic("internal: FromTable and FromSelect are both nil")
		}

		b.out.WriteString(" WHERE ")
		b.out.WriteByte('(')
		b.out.WriteString(f.Expr)
		b.out.WriteString(") ILIKE ?")
		b.args = append(b.args, fmt.Sprintf("%%%s%%", b.search))

		if n.TimeWhere != nil && n.TimeWhere.Expr != "" {
			b.out.WriteString(" AND (")
			b.out.WriteString(n.TimeWhere.Expr)
			b.out.WriteString(")")
			b.args = append(b.args, n.TimeWhere.Args...)
		}
		if n.Where != nil && n.Where.Expr != "" {
			b.out.WriteString(" AND (")
			b.out.WriteString(n.Where.Expr)
			b.out.WriteString(")")
			b.args = append(b.args, n.Where.Args...)
		}

		b.out.WriteString(" GROUP BY value")

		if n.Having != nil && n.Having.Expr != "" {
			b.out.WriteString(" HAVING ")
			b.out.WriteString(n.Having.Expr)
			b.args = append(b.args, n.Having.Args...)
		}
	}

	if n.Limit != nil {
		b.out.WriteString(" LIMIT ")
		b.out.WriteString(strconv.FormatInt(*n.Limit, 10))
	}

	if n.Offset != nil {
		b.out.WriteString(" OFFSET ")
		b.out.WriteString(strconv.FormatInt(*n.Offset, 10))
	}

	return nil
}

func (b *searchSQLBuilder) writeJoin(joinType JoinType, baseSelect, joinSelect *SelectNode) error {
	b.out.WriteByte(' ')
	b.out.WriteString(string(joinType))
	b.out.WriteString(" JOIN ")
	if !joinSelect.IsCTE {
		b.out.WriteByte('(')
		err := b.writeSelect(joinSelect)
		if err != nil {
			return err
		}
		b.out.WriteString(") ")
	}
	b.out.WriteString(joinSelect.Alias)

	if len(baseSelect.DimFields) == 0 {
		b.out.WriteString(" ON TRUE")
		return nil
	}

	b.out.WriteString(" ON ")
	for i, f := range baseSelect.DimFields {
		if i > 0 {
			b.out.WriteString(" AND ")
		}
		lhs := b.ast.sqlForMember(baseSelect.Alias, f.Name)
		rhs := b.ast.sqlForMember(joinSelect.Alias, f.Name)
		b.out.WriteByte('(')
		b.out.WriteString(b.ast.dialect.JoinOnExpression(lhs, rhs))
		b.out.WriteByte(')')
	}
	return nil
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
