package metricsview

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

func (e *Executor) executePivot(ctx context.Context, ast *AST, pivot *pivotAST) (*drivers.Result, error) {
	// Build underlying SQL
	underlyingSQL, args, err := ast.SQL()
	if err != nil {
		return nil, err
	}

	// If the dialect supports native pivoting, we can do it as a single query
	if e.olap.Dialect().CanPivot() {
		sql, err := pivot.SQL(ast, underlyingSQL, true)
		if err != nil {
			return nil, err
		}

		res, err := e.olap.Execute(ctx, &drivers.Statement{
			Query:            sql,
			Args:             args,
			Priority:         e.priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to execute native pivot: %w", err)
		}
		return res, nil
	}

	// We know that DuckDB supports pivoting and is always available, so we can use it as a fallback to do the pivot
	duck, release, err := e.rt.OLAP(ctx, e.instanceID, "duckdb")
	if err != nil {
		return nil, fmt.Errorf("failed to acquire duckdb for pivot: %w", err)
	}
	defer release()

	// Execute underlying SQL on the OLAP
	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:            underlyingSQL,
		Args:             args,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	// Apply the pivot cell limit
	res.SetCap(pivot.underlyingRowCap)

	// Transfer data to DuckDB
	// TODO: Implement this

	// Create cleanup function
	cleanup := func() error {
		// TODO: Drop tmp table here
		release()
		return nil
	}

	// Pivot the data in DuckDB
	sql, err := pivot.SQL(ast, "SELECT * FROM <tmp table>", false)
	if err != nil {
		_ = cleanup()
		return nil, err
	}

	res, err = duck.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             nil,
		Priority:         e.priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		_ = cleanup()
		return nil, fmt.Errorf("failed to execute non-native pivot: %w", err)
	}

	res.SetCleanupFunc(cleanup)
	return res, nil
}

func (e *Executor) rewriteQueryForPivot(qry *Query) (*pivotAST, bool, error) {
	// Skip if we're not pivoting
	if qry.PivotOn == nil {
		return nil, false, nil
	}

	// Check pivot fields are dims (not measures)
	for _, f := range qry.PivotOn {
		var found bool
		for _, d := range qry.Dimensions {
			if d.Name == f {
				found = true
				break
			}
		}
		if !found {
			return nil, false, fmt.Errorf("pivot field %q not found in dimensions", f)
		}
	}

	// Check sort fields are non-pivoted dims
	for _, s := range qry.Sort {
		var found bool
		for _, d := range qry.Dimensions {
			if d.Name == s.Name {
				found = true
				break
			}
		}
		if !found {
			return nil, false, fmt.Errorf("sort field %q is not a dimension (pivot queries can only sort on non-pivoted dimensions)", s.Name)
		}

		for _, f := range qry.PivotOn {
			if f == s.Name {
				return nil, false, fmt.Errorf("sort field %q is a pivot field (pivot queries can only sort on non-pivoted dimensions)", s.Name)
			}
		}
	}

	// Build a pivotAST based on fields to apply during and after the pivot (instead of in the underlying query)
	ast := &pivotAST{
		keep:    nil, // Populated below
		on:      qry.PivotOn,
		using:   nil, // Populated below
		orderBy: nil, // Populated below
		limit:   qry.Limit,
		offset:  qry.Offset,
		label:   qry.Label,
		dialect: e.olap.Dialect(),
	}
	for _, d := range qry.Dimensions {
		var found bool
		for _, f := range qry.PivotOn {
			if f == d.Name {
				found = true
				break
			}
		}
		if !found {
			ast.keep = append(ast.keep, d.Name)
		}
	}
	for _, m := range qry.Measures {
		ast.using = append(ast.using, m.Name)
	}
	for _, f := range qry.Sort {
		ast.orderBy = append(ast.orderBy, OrderFieldNode(f))
	}

	// Remove parameters from the underlying query that are now handled in the pivot AST
	qry.PivotOn = nil
	qry.Sort = nil
	qry.Limit = nil
	qry.Offset = nil
	qry.Label = false

	// If we have a cell limit, apply a row limit just above it to the underlying query.
	// This prevents the DB from scanning too much data before we can detect that the query will exceed the cell limit.
	if e.instanceCfg.PivotCellLimit != 0 {
		cols := int64(len(qry.Dimensions) + len(qry.Measures))
		ast.underlyingRowCap = e.instanceCfg.PivotCellLimit / cols

		tmp := ast.underlyingRowCap + 1
		qry.Limit = &tmp
	}

	return ast, true, nil
}

// pivotAST represents config for generating a PIVOT query.
type pivotAST struct {
	keep    []string
	on      []string
	using   []string
	orderBy []OrderFieldNode
	limit   *int64
	offset  *int64

	label            bool
	dialect          drivers.Dialect
	underlyingRowCap int64
}

// SQL generates a query that outputs a pivoted table based on the pivot config and data in the underlying query.
// As a convenience, it accepts the underlying AST and SQL separately, which enables pivoting in a different connector than the one that runs the actual underlying query.
func (a *pivotAST) SQL(underlyingAST *AST, underlyingSQL string, checkCap bool) (string, error) {
	if !a.dialect.CanPivot() {
		return "", fmt.Errorf("pivot queries not supported for dialect %q", a.dialect.String())
	}

	b := &strings.Builder{}

	// Since we query the underlying data and do the pivot in a single query, we need to be creative to enforce the pivot cell limit.
	// We leverage CTEs and DuckDB's ERROR function to enforce the limit.
	// This is pretty DuckDB-specific, but that's also currently the only OLAP we use that supports pivoting.
	// The query looks something like:
	//
	//   WITH t1 AS (<underlyingSQL>),
	//   t2 AS (SELECT * FROM t1 WHERE IF(EXISTS (SELECT COUNT(*) AS count FROM t1 HAVING count > <limit>), ERROR('pivot query exceeds limit'), TRUE))
	//   PIVOT t2 ON ...
	if checkCap {
		t1, err := randomString("t1", 8)
		if err != nil {
			return "", fmt.Errorf("failed to generate random alias: %w", err)
		}
		t2, err := randomString("t2", 8)
		if err != nil {
			return "", fmt.Errorf("failed to generate random alias: %w", err)
		}

		b.WriteString("WITH ")
		b.WriteString(t1)
		b.WriteString(" AS (")
		b.WriteString(underlyingSQL)
		b.WriteString("), ")
		b.WriteString(t2)
		b.WriteString(" AS (SELECT * FROM ")
		b.WriteString(t1)
		b.WriteString(" WHERE IF(EXISTS (SELECT COUNT(*) AS count FROM ")
		b.WriteString(t1)
		b.WriteString(" HAVING count > ")
		b.WriteString(strconv.FormatInt(a.underlyingRowCap, 10))
		b.WriteString("), ERROR('pivot query exceeds limit of ")
		b.WriteString(strconv.FormatInt(a.underlyingRowCap, 10))
		b.WriteString(" cells'), TRUE)) ")

		underlyingSQL = fmt.Sprintf("SELECT * FROM %s", t2)
	}

	// If we need to label some fields (in practice, this will be non-pivoted dims during exports),
	// we emit a query like: SELECT d1 AS "L1", d2 AS "L2", * EXCLUDE (d1, d2) FROM (PIVOT ...)
	wrapWithLabels := a.label && len(a.keep) > 0
	if wrapWithLabels {
		b.WriteString("SELECT ")
		for _, fn := range a.keep {
			f, ok := findField(fn, underlyingAST.Root.DimFields)
			if !ok {
				return "", fmt.Errorf("pivot keep dimension %q not found in underlying query", fn)
			}

			b.WriteString(a.dialect.EscapeIdentifier(f.Name))
			if f.Label != "" {
				b.WriteString(" AS ")
				b.WriteString(a.dialect.EscapeIdentifier(f.Label))
			}
			b.WriteString(", ")
		}

		b.WriteString("* EXCLUDE (")
		for i, fn := range a.keep {
			f, ok := findField(fn, underlyingAST.Root.DimFields)
			if !ok {
				return "", fmt.Errorf("pivot keep dimension %q not found in underlying query", fn)
			}

			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.dialect.EscapeIdentifier(f.Name))
		}

		b.WriteString(") FROM (")
	}

	// Build a PIVOT query like: PIVOT (<underlyingSQL>) ON <dimensions> USING <measures> ORDER BY <sort> LIMIT <limit> OFFSET <offset>
	b.WriteString("PIVOT (")
	b.WriteString(underlyingSQL)
	b.WriteString(") ON ")
	for i, fn := range a.on {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(a.dialect.EscapeIdentifier(fn))
	}

	if len(a.using) > 0 {
		b.WriteString(" USING ")
		for i, fn := range a.using {
			f, ok := findField(fn, underlyingAST.Root.MeasureFields)
			if !ok {
				return "", fmt.Errorf("pivot using measure %q not found in underlying query", fn)
			}

			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString("ANY_VALUE(")
			b.WriteString(a.dialect.EscapeIdentifier(fn))
			b.WriteString(")")
			b.WriteString(" AS ")
			if a.label && f.Label != "" {
				b.WriteString(a.dialect.EscapeIdentifier(f.Label))
			} else {
				b.WriteString(a.dialect.EscapeIdentifier(f.Name))
			}
		}
	}

	if len(a.orderBy) > 0 {
		b.WriteString(" ORDER BY ")
		for i, f := range a.orderBy {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(a.dialect.OrderByExpression(f.Name, f.Desc))
		}
	}

	if a.limit != nil {
		b.WriteString(" LIMIT ")
		b.WriteString(strconv.FormatInt(*a.limit, 10))
	}

	if a.offset != nil {
		b.WriteString(" OFFSET ")
		b.WriteString(strconv.FormatInt(*a.offset, 10))
	}

	if wrapWithLabels {
		b.WriteString(")")
	}

	return b.String(), nil
}

func randomString(prefix string, n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}

func findField(n string, fs []FieldNode) (FieldNode, bool) {
	for _, f := range fs {
		if f.Name == n {
			return f, true
		}
	}
	return FieldNode{}, false
}
