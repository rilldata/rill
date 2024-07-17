package metricsview

import (
	"strconv"
	"strings"
)

// SQL builds a SQL query from the AST.
// It returns the query and query arguments to be passed to the database driver.
func (a *AST) SQL() (string, []any, error) {
	b := &sqlBuilder{
		ast: a,
		out: &strings.Builder{},
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
	if len(n.CTEs) > 0 {
		b.out.WriteString("WITH ")
		for _, cte := range n.CTEs {
			b.out.WriteString(cte.Alias)
			b.out.WriteString(" AS (")
			err := b.writeSelect(cte)
			if err != nil {
				return err
			}
			b.out.WriteString(") ")
		}
	}

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

		// Add unnest joins. We only and always apply these against FromPlain (ensuring they are already unnested when referenced in outer SELECTs).
		for _, u := range n.Unnests {
			if n.JoinComparisonSelect != nil {
				b.out.WriteString("JOIN ")
				b.out.WriteString(u)
				b.out.WriteString(" ON TRUE")
			} else {
				b.out.WriteString(", ")
				b.out.WriteString(u)
			}
		}

		if n.JoinComparisonSelect != nil {
			err := b.writeJoin(n.JoinComparisonType, n.CTEs[0], n.JoinComparisonSelect, nil)
			if err != nil {
				return err
			}
		}
	} else if n.FromSelect != nil {
		b.out.WriteByte('(')
		err := b.writeSelect(n.FromSelect)
		if err != nil {
			return err
		}
		b.out.WriteString(") ")
		b.out.WriteString(n.FromSelect.Alias)

		for _, ljs := range n.LeftJoinSelects {
			err := b.writeJoin("LEFT", n.FromSelect, ljs, nil)
			if err != nil {
				return err
			}
		}

		if n.SpineSelect != nil {
			err := b.writeJoin("RIGHT", n.FromSelect, n.SpineSelect, nil)
			if err != nil {
				return err
			}
		}

		if n.JoinComparisonSelect != nil {
			err := b.writeJoin(n.JoinComparisonType, n.FromSelect, n.JoinComparisonSelect, nil)
			if err != nil {
				return err
			}
		}

		if n.JoinComparisonTable != nil {
			err := b.writeJoin(n.JoinComparisonType, n.FromSelect, nil, n.JoinComparisonTable)
			if err != nil {
				return err
			}
		}
	} else {
		panic("internal: FromPlain and FromSelect are both nil")
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
			b.out.WriteString(" WHERE (")
		}
		b.out.WriteString(n.Where.Expr)
		b.out.WriteString(")")
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

func (b *sqlBuilder) writeJoin(joinType JoinType, baseSelect, joinSelect *SelectNode, joinTable *string) error {
	b.out.WriteByte(' ')
	b.out.WriteString(string(joinType))
	var joinSelectAlias string
	if joinSelect != nil {
		joinSelectAlias = joinSelect.Alias
		b.out.WriteString(" JOIN (")
		err := b.writeSelect(joinSelect)
		if err != nil {
			return err
		}
		b.out.WriteString(") ")
		b.out.WriteString(joinSelectAlias)
	} else if joinTable != nil {
		joinSelectAlias = *joinTable
		b.out.WriteString(" JOIN ")
		b.out.WriteString(joinSelectAlias)
	} else {
		panic("internal: joinSelect and joinTable are both nil")
	}

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
		rhs := b.ast.sqlForMember(joinSelectAlias, f.Name)
		b.out.WriteByte('(')
		b.out.WriteString(b.ast.dialect.JoinOnExpression(lhs, rhs))
		b.out.WriteByte(')')
	}
	return nil
}
