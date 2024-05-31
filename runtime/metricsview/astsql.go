package metricsview

import (
	"fmt"
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

	err := b.writeSelect(a.Root)
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

func (b *sqlBuilder) writeSelect(n *SelectNode) error {
	b.out.WriteString("SELECT ")

	for i, f := range n.DimFields {
		if i > 0 {
			b.out.WriteString(", ")
		}

		expr := f.Expr
		if f.Unnest {
			expr = b.ast.sqlForMember(f.UnnestAlias, f.Name)
		}

		b.out.WriteByte('(')
		b.out.WriteString(expr)
		b.out.WriteString(") as ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
	}

	for i, f := range n.MeasureFields {
		if i > 0 || len(n.DimFields) > 0 {
			b.out.WriteString(", ")
		}

		b.out.WriteByte('(')
		b.out.WriteString(f.Expr)
		b.out.WriteString(") as ")
		b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
	}

	b.out.WriteString(" FROM ")
	if n.FromTable != nil {
		b.out.WriteString(*n.FromTable)

		// Add unnest joins. We only and always apply these against FromPlain (ensuring they are already unnested when referenced in outer SELECTs).
		for _, f := range n.DimFields {
			if !f.Unnest {
				continue
			}

			tblWithAlias, auto, err := b.ast.dialect.LateralUnnest(f.Expr, f.UnnestAlias, f.Name)
			if err != nil {
				return fmt.Errorf("failed to unnest field %q: %w", f.Name, err)
			}

			if auto {
				continue
			}

			b.out.WriteString(", ")
			b.out.WriteString(tblWithAlias)
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
			err := b.writeJoin("LEFT", n.FromSelect, ljs)
			if err != nil {
				return err
			}
		}

		if n.JoinComparisonSelect != nil {
			err := b.writeJoin(n.JoinComparisonType, n.FromSelect, n.JoinComparisonSelect)
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
			b.out.WriteString(b.ast.dialect.EscapeIdentifier(f.Name))
			if f.Desc {
				b.out.WriteString(" DESC")
			}
		}
	}

	if n.Limit != nil {
		b.out.WriteString(" LIMIT ")
		b.out.WriteString(strconv.Itoa(*n.Limit))
	}

	if n.Offset != nil {
		b.out.WriteString(" OFFSET ")
		b.out.WriteString(strconv.Itoa(*n.Offset))
	}

	return nil
}

func (b *sqlBuilder) writeJoin(joinType string, baseSelect, joinSelect *SelectNode) error {
	b.out.WriteByte(' ')
	b.out.WriteString(joinType)
	b.out.WriteString(" JOIN (")
	err := b.writeSelect(joinSelect)
	if err != nil {
		return err
	}
	b.out.WriteString(") ")
	b.out.WriteString(joinSelect.Alias)
	b.out.WriteString(" ON ")
	for i, f := range baseSelect.DimFields {
		if i > 0 {
			b.out.WriteString(" AND ")
		}
		lhs := b.ast.sqlForMember(baseSelect.Alias, f.Name)
		rhs := b.ast.sqlForMember(joinSelect.Alias, f.Name)
		b.out.WriteByte('(')
		b.out.WriteString(lhs)
		b.out.WriteByte('=')
		b.out.WriteString(rhs)
		b.out.WriteString(" OR ")
		b.out.WriteString(lhs)
		b.out.WriteString(" IS NULL AND ")
		b.out.WriteString(rhs)
		b.out.WriteString(" IS NULL")
		b.out.WriteByte(')')
	}
	return nil
}
