package metricsresolver

import (
	"fmt"
	"strconv"
	"strings"
)

func (a *AST) SQL() (string, []any, error) {
	var args []any
	b := &strings.Builder{}
	err := a.writeSQLForMetricsSelect(a.Root, b, &args)
	if err != nil {
		return "", nil, err
	}

	return b.String(), args, nil
}

func (a *AST) writeSQLForMetricsSelect(n *MetricsSelect, b *strings.Builder, args *[]any) error {
	b.WriteString("SELECT ")

	for i, f := range n.DimFields {
		if i > 0 {
			b.WriteString(", ")
		}

		expr := f.Expr
		if f.Unnest {
			expr = a.expressionForMember(f.UnnestAlias, f.Name)
		}

		b.WriteByte('(')
		b.WriteString(expr)
		b.WriteString(") as ")
		b.WriteString(a.dialect.EscapeIdentifier(f.Name))
	}

	for i, f := range n.MeasureFields {
		if i > 0 || len(n.DimFields) > 0 {
			b.WriteString(", ")
		}

		b.WriteByte('(')
		b.WriteString(f.Expr)
		b.WriteString(") as ")
		b.WriteString(a.dialect.EscapeIdentifier(f.Name))
	}

	b.WriteString(" FROM ")
	if n.FromPlain != nil {
		b.WriteString(n.FromPlain.From)

		// Add unnest joins. We only and always apply these against FromPlain (ensuring they are already unnested when referenced in outer SELECTs).
		for _, f := range n.DimFields {
			if !f.Unnest {
				continue
			}

			tblWithAlias, auto, err := a.dialect.LateralUnnest(f.Expr, f.UnnestAlias, f.Name)
			if err != nil {
				return fmt.Errorf("failed to unnest field %q: %w", f.Name, err)
			}

			if auto {
				continue
			}

			b.WriteString(", ")
			b.WriteString(tblWithAlias)
		}
	} else if n.FromSelect != nil {
		b.WriteByte('(')
		err := a.writeSQLForMetricsSelect(n.FromSelect, b, args)
		if err != nil {
			return err
		}
		b.WriteString(") ")
		b.WriteString(n.FromSelect.Alias)

		for _, ljs := range n.LeftJoinSelects {
			b.WriteString(" LEFT JOIN (")
			err := a.writeSQLForMetricsSelect(ljs, b, args)
			if err != nil {
				return err
			}
			b.WriteString(") ")
			b.WriteString(ljs.Alias)
			b.WriteString(" ON ")
			for i, f := range n.FromSelect.DimFields {
				if i > 0 {
					b.WriteString(" AND ")
				}
				b.WriteString(a.expressionForMember(n.FromSelect.Alias, f.Name))
				b.WriteByte('=')
				b.WriteString(a.expressionForMember(ljs.Alias, f.Name))
			}
		}

		if n.JoinComparisonSelect != nil {
			b.WriteByte(' ')
			b.WriteString(n.JoinComparisonType)
			b.WriteString(" JOIN (")
			err := a.writeSQLForMetricsSelect(n.JoinComparisonSelect, b, args)
			if err != nil {
				return err
			}
			b.WriteString(") ")
			b.WriteString(n.JoinComparisonSelect.Alias)
			b.WriteString(" ON ")
			for i, f := range n.FromSelect.DimFields {
				if i > 0 {
					b.WriteString(" AND ")
				}
				b.WriteString(a.expressionForMember(n.FromSelect.Alias, f.Name))
				b.WriteByte('=')
				b.WriteString(a.expressionForMember(n.JoinComparisonSelect.Alias, f.Name))
			}
		}
	} else {
		panic("internal: FromPlain and FromSelect are both nil")
	}

	var wroteWhere bool
	if n.FromPlain != nil && n.FromPlain.Where != nil && n.FromPlain.Where.Expr != "" {
		wroteWhere = true
		b.WriteString(" WHERE (")
		b.WriteString(n.FromPlain.Where.Expr)
		b.WriteString(")")
		*args = append(*args, n.FromPlain.Where.Args...)
	}
	if n.Where != nil && n.Where.Expr != "" {
		if wroteWhere {
			b.WriteString(" AND (")
		} else {
			b.WriteString(" WHERE (")
		}
		b.WriteString(n.Where.Expr)
		b.WriteString(")")
		*args = append(*args, n.Where.Args...)
	}

	if n.Group {
		var baseAlias string
		if n.FromSelect != nil {
			baseAlias = n.FromSelect.Alias
		}

		b.WriteString(" GROUP BY ")
		for i, f := range n.DimFields {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.expressionForMember(baseAlias, f.Name))
		}
	}

	if n.Having != nil && n.Having.Expr != "" {
		b.WriteString(" HAVING ")
		b.WriteString(n.Having.Expr)
		*args = append(*args, n.Having.Args...)
	}

	if len(n.OrderBy) > 0 {
		b.WriteString(" ORDER BY ")
		for i, f := range n.OrderBy {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.dialect.EscapeIdentifier(f.Name))
			if f.Desc {
				b.WriteString(" DESC")
			}
		}
	}

	if n.Limit != nil {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.Itoa(*n.Limit))
	}

	if n.Offset != nil {
		b.WriteString(" OFFSET ")
		b.WriteString(strconv.Itoa(*n.Offset))
	}

	return nil
}
