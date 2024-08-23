package metricsview

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SQL builds a SQL query from the AST.
// It returns the query and query arguments to be passed to the database driver.
func (a *AST) SQL() (string, []any, error) {
	b := &sqlBuilder{
		ast: a,
		out: &strings.Builder{},
	}

	if len(a.CTEs) > 0 {
		b.out.WriteString("WITH ")
		for i, cte := range a.CTEs {
			if i > 0 {
				b.out.WriteString(", ")
			}
			b.out.WriteString(cte.Alias)
			b.out.WriteString(" AS (")
			err := b.writeSelect(cte)
			if err != nil {
				return "", nil, err
			}
			b.out.WriteString(") ")
		}
	}

	var err error
	if a.query.Label {
		err = b.writeSelectWithLabels(a.Root)
	} else {
		err = b.writeSelect(a.Root)
	}
	if err != nil {
		return "", nil, err
	}

	return b.out.String(), b.args, nil
}

type sqlBuilder struct {
	ast  *AST
	out  *strings.Builder
	args []any
}

func (b *sqlBuilder) writeSelectWithLabels(n *SelectNode) error {
	b.out.WriteString("SELECT ")

	for i, f := range n.DimFields {
		label := f.Label
		if label == "" {
			label = f.Name
		}

		if i > 0 {
			b.out.WriteString(", ")
		}
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
		b.out.WriteString(" AS ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(label))
	}

	for i, f := range n.MeasureFields {
		label := f.Label
		if label == "" {
			label = f.Name
		}

		if i > 0 || len(n.DimFields) > 0 {
			b.out.WriteString(", ")
		}
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
		b.out.WriteString(" AS ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(label))
	}

	b.out.WriteString(" FROM (")
	err := b.writeSelect(n)
	if err != nil {
		return err
	}
	b.out.WriteString(")")

	return nil
}

func (b *sqlBuilder) writeSelect(n *SelectNode) error {
	b.out.WriteString("SELECT ")

	for i, f := range n.DimFields {
		if i > 0 {
			b.out.WriteString(", ")
		}

		b.out.WriteByte('(')
		b.out.WriteString(f.Expr)
		b.out.WriteString(") AS ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
	}

	for i, f := range n.MeasureFields {
		if i > 0 || len(n.DimFields) > 0 {
			b.out.WriteString(", ")
		}

		b.out.WriteByte('(')
		b.out.WriteString(f.Expr)
		b.out.WriteString(") AS ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
	}

	b.out.WriteString(" FROM ")
	if n.FromTable != nil {
		b.out.WriteString(*n.FromTable)

		// Add unnest joins. We only and always apply these against FromTable (ensuring they are already unnested when referenced in outer SELECTs).
		for _, u := range n.Unnests {
			b.out.WriteString(", ")
			b.out.WriteString(u)
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
			err := b.writeJoin("LEFT", n.FromSelect, ljs, false)
			if err != nil {
				return err
			}
		}

		if n.SpineSelect != nil {
			err := b.writeJoin("RIGHT", n.FromSelect, n.SpineSelect, false)
			if err != nil {
				return err
			}
		}
		if n.JoinComparisonSelect != nil {
			err := b.writeJoin(n.JoinComparisonType, n.FromSelect, n.JoinComparisonSelect, true)
			if err != nil {
				return err
			}
		}
	} else {
		panic("internal: FromTable and FromSelect are both nil")
	}

	var wroteWhere bool
	if n.TimeWhere != nil && n.TimeWhere.Expr != "" {
		wroteWhere = true
		b.out.WriteString(" WHERE (")
		b.out.WriteString(n.TimeWhere.Expr)
		b.out.WriteString(")")
		b.args = append(b.args, n.TimeWhere.Args...)
	}
	if n.Where != nil && n.Where.Expr != "" {
		if wroteWhere {
			b.out.WriteString(" AND (")
		} else {
			b.out.WriteString(" WHERE ")
		}
		b.out.WriteString(n.Where.Expr)
		if wroteWhere {
			b.out.WriteString(")")
		}
		b.args = append(b.args, n.Where.Args...)
	}

	if n.Group && len(n.DimFields) > 0 {
		b.out.WriteString(" GROUP BY ")
		for i := range n.DimFields {
			if i > 0 {
				b.out.WriteString(", ")
			}
			b.out.WriteString(strconv.Itoa(i + 1))
		}
	}

	if n.Having != nil && n.Having.Expr != "" {
		b.out.WriteString(" HAVING ")
		b.out.WriteString(n.Having.Expr)
		b.args = append(b.args, n.Having.Args...)
	}

	if len(n.OrderBy) > 0 {
		b.out.WriteString(" ORDER BY ")
		for i, f := range n.OrderBy {
			if i > 0 {
				b.out.WriteString(", ")
			}
			b.out.WriteString(b.ast.dialect.OrderByExpression(f.Name, f.Desc))
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

func (a *AST) interval(g, mg TimeGrain) (string, error) {
	var start1 time.Time
	var start2 time.Time
	if a.query.TimeRange == nil {
		return "", fmt.Errorf("no time range for the offset")
	}
	if a.query.TimeRange.Start.IsZero() {
		return "", fmt.Errorf("no start time for the offset")
	}
	start1 = a.query.TimeRange.Start
	if a.query.ComparisonTimeRange == nil {
		return "", fmt.Errorf("no comparison time range for the offset")
	}
	if a.query.ComparisonTimeRange.Start.IsZero() {
		return "", fmt.Errorf("no start time for the comparison time range")
	}
	start2 = a.query.ComparisonTimeRange.Start
	if g == TimeGrainUnspecified {
		g = TimeGrainMillisecond // todo millis won't work for druid
		return a.dialect.DateDiff(string(g), start1, start2)
	} else if g == mg {
		return a.dialect.DateDiff(string(g), start1, start2)
	}
	// g > mg -> zero diff
	return "0", nil
}

func (b *sqlBuilder) writeJoin(joinType JoinType, baseSelect, joinSelect *SelectNode, comp bool) error {
	b.out.WriteByte(' ')
	b.out.WriteString(string(joinType))
	b.out.WriteString(" JOIN ")
	// If the join select is a CTE, then just add the CTE alias otherwise add the full select query
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
		if comp && f.Time {
			intv, err := b.ast.interval(f.TimeGrain, f.MinGrain)
			if err != nil {
				return err
			}

			if f.TimeGrain == TimeGrainUnspecified {
				return fmt.Errorf("unspecified time grain")
			}

			// example: base.ts IS NOT DISTINCT FROM comparison.ts - INTERVAL (DATEDIFF(...)) SECONDS
			rhs = fmt.Sprintf("(%s - INTERVAL (%s) %s)", rhs, intv, string(f.TimeGrain))
		}
		b.out.WriteByte('(')
		b.out.WriteString(b.ast.dialect.JoinOnExpression(lhs, rhs))
		b.out.WriteByte(')')
	}
	return nil
}
