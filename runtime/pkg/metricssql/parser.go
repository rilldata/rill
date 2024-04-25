package metricssqlparser

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/parser/opcode"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"

	// need to import parser driver as well
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

var supportedFuncs = map[string]any{
	"date_trunc": nil,
}

type Compiler struct {
	p              *parser.Parser
	controller     *runtime.Controller
	instanceID     string
	userAttributes map[string]any
	priority       int
}

// New returns a compiler and created a tidb parser object.
// The instantiated parser object and thus compiler is not goroutine safe and not lightweight.
// It is better to keep it in a single goroutine, and reuse it if possible.
func New(ctrl *runtime.Controller, instanceID string, userAttributes map[string]any, priority int) *Compiler {
	p := parser.New()
	// Weirdly setting just ModeANSI which is a combination having ModeANSIQuotes doesn't ensure double quotes are used to identify SQL identifiers
	p.SetSQLMode(mysql.ModeANSI | mysql.ModeANSIQuotes)
	return &Compiler{
		p:              p,
		controller:     ctrl,
		instanceID:     instanceID,
		userAttributes: userAttributes,
		priority:       priority,
	}
}

// Compile parses a metrics SQL query and compiles it to a regular SQL query.
// It uses tidb parser(which is a MySQL compliant parser) and transforms over the generated AST to generate regular SQL query.
// We use MySQL's ANSI sql Mode to conform more closely to standard SQL.
//
// Whenever adding transform method over new node type also look at its `Restore` method to get an idea how it can be parsed into a SQL query.
func (c *Compiler) Compile(ctx context.Context, sql string) (string, string, []*runtimev1.ResourceName, error) {
	stmtNodes, _, err := c.p.ParseSQL(sql)
	if err != nil {
		return "", "", nil, err
	}

	if len(stmtNodes) != 1 {
		return "", "", nil, errors.New("metrics sql: expected exactly one SQL statement")
	}

	stmt, ok := stmtNodes[0].(*ast.SelectStmt)
	if !ok {
		return "", "", nil, errors.New("metrics sql: expected a SELECT statement")
	}

	t := &transformer{
		controller:     c.controller,
		instanceID:     c.instanceID,
		userAttributes: c.userAttributes,
		priority:       c.priority,
	}
	compiledSQL, err := t.transformSelectStmt(ctx, stmt)
	if err != nil {
		return "", "", nil, err
	}

	return compiledSQL, t.connector, t.refs, nil
}

type transformer struct {
	controller     *runtime.Controller
	instanceID     string
	userAttributes map[string]any

	metricsView   *runtimev1.MetricsViewV2
	refs          []*runtimev1.ResourceName
	connector     string
	dimsToExpr    map[string]string
	measureToExpr map[string]string
	priority      int
}

func (t *transformer) transformSelectStmt(ctx context.Context, node *ast.SelectStmt) (string, error) {
	if node.WithBeforeBraces {
		return "", fmt.Errorf("metrics sql: WITH clause is not supported")
	}
	if node.IsInBraces {
		return "", fmt.Errorf("metrics sql: sub select is not supported")
	}
	if node.With != nil {
		return "", fmt.Errorf("metrics sql: WITH clause is not supported")
	}

	// parse from clause
	if node.From == nil || node.From.TableRefs == nil {
		return "", fmt.Errorf("metrics sql: need from clause")
	}

	var sb strings.Builder
	sb.WriteString("SELECT ")
	fromClause, err := t.transformFromClause(ctx, node.From)
	if err != nil {
		return "", err
	}

	selectList, groupByClause, err := t.transformSelectStmtColumns(ctx, node)
	if err != nil {
		return "", err
	}

	sb.WriteString(selectList)
	sb.WriteString(" FROM ")
	sb.WriteString(fromClause)

	if node.Where != nil {
		where, err := t.transformExprNode(ctx, node.Where)
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(where.expr)
	}
	if node.GroupBy != nil {
		return "", fmt.Errorf("metrics sql: Explicit group by clause is not supported. Group by clause is implicitly added when both measure and dimensions are selected. The implicit group by includes all selected dimensions")
	}
	if groupByClause != "" {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(groupByClause)
	}

	if node.Having != nil {
		having, err := t.transformHavingClause(ctx, node.Having)
		if err != nil {
			return "", err
		}
		sb.WriteString(" HAVING ")
		sb.WriteString(having)
	}

	if node.OrderBy != nil {
		orderBy, err := t.transformOrderByClause(ctx, node.OrderBy)
		if err != nil {
			return "", err
		}
		sb.WriteString(" ORDER BY ")
		sb.WriteString(orderBy)
	}

	if node.Limit != nil {
		limit, err := t.transformLimitClause(ctx, node.Limit)
		if err != nil {
			return "", err
		}
		sb.WriteString(" LIMIT ")
		sb.WriteString(limit)
	}
	return sb.String(), nil
}

func (t *transformer) transformSelectStmtColumns(ctx context.Context, node *ast.SelectStmt) (string, string, error) {
	if len(node.Fields.Fields) == 0 {
		return "", "", fmt.Errorf("metrics sql: need to select atleast one dimension or measure")
	}

	var sb strings.Builder
	var groupByList []string
	var hasMeasure bool
	for i, field := range node.Fields.Fields {
		if i != 0 {
			sb.WriteString(", ")
		} else if node.Distinct {
			sb.WriteString("DISTINCT ")
		}
		if field.WildCard != nil {
			return "", "", fmt.Errorf("metrics sql: wildcard is not supported")
		}

		res, err := t.transformExprNode(ctx, field.Expr)
		if err != nil {
			return "", "", err
		}

		sb.WriteString(res.expr)
		// write alias if any
		if field.AsName.String() != "" { // if explicitly specified in the metrics_sql
			sb.WriteString(" AS \"")
			sb.WriteString(field.AsName.String())
			sb.WriteString("\"")
		} else if _, ok := field.Expr.(*ast.ColumnNameExpr); ok { // plain dimension or measure
			sb.WriteString(" AS ")
			sb.WriteString(res.columns[0])
		}
		if len(res.types) > 0 {
			for i := 1; i < len(res.types); i++ {
				if res.types[i] != res.types[0] {
					return "", "", fmt.Errorf("metrics sql: operations combining measure and dimension is not supported in select field: %v", restore(field))
				}
			}

			if res.types[0] == "DIMENSION" {
				groupByList = append(groupByList, res.expr)
			} else {
				hasMeasure = true
			}
		}
	}
	if hasMeasure && len(groupByList) > 0 {
		slices.Sort(groupByList)
		return sb.String(), strings.Join(groupByList, ", "), nil
	}
	return sb.String(), "", nil
}

func (t *transformer) transformFromClause(ctx context.Context, node *ast.TableRefsClause) (string, error) {
	n := node.TableRefs
	if n == nil || n.Left == nil {
		return "", fmt.Errorf("metrics sql: need `FROM metrics_view` clause")
	}

	tblSrc, ok := n.Left.(*ast.TableSource)
	if !ok {
		// if left is not a table source, then it must be a join
		return "", fmt.Errorf("metrics sql: join is not supported")
	}

	tblName, ok := tblSrc.Source.(*ast.TableName)
	if !ok {
		return "", fmt.Errorf("metrics sql: only FROM `metrics_view` is supported")
	}

	resource := &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: tblName.Name.String()}
	mv, err := t.controller.Get(ctx, resource, false)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return "", fmt.Errorf("metrics sql: metrics view `%s` not found", tblName.Name.String())
		}
		return "", err
	}

	t.metricsView = mv.GetMetricsView()
	t.refs = []*runtimev1.ResourceName{resource}
	t.connector = mv.GetMetricsView().Spec.Connector
	return t.fromQueryForMetricsView(ctx, mv)
}

func (t *transformer) transformHavingClause(ctx context.Context, node *ast.HavingClause) (string, error) {
	res, err := t.transformExprNode(ctx, node.Expr)
	if err != nil {
		return "", err
	}
	return res.expr, nil
}

func (t *transformer) transformOrderByClause(ctx context.Context, node *ast.OrderByClause) (string, error) {
	var sb strings.Builder
	for i, item := range node.Items {
		if i != 0 {
			sb.WriteString(", ")
		}
		expr, err := t.transformExprNode(ctx, item.Expr)
		if err != nil {
			return "", err
		}
		if item.Desc {
			sb.WriteString(expr.expr + " DESC")
		} else {
			sb.WriteString(expr.expr + " ASC")
		}
	}
	return sb.String(), nil
}

func (t *transformer) transformLimitClause(ctx context.Context, node *ast.Limit) (string, error) {
	count, err := t.transformExprNode(ctx, node.Count)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(count.expr)
	if node.Offset != nil {
		offset, err := t.transformExprNode(ctx, node.Offset)
		if err != nil {
			return "", err
		}
		sb.WriteString(" OFFSET ")
		sb.WriteString(offset.expr)
	}

	return sb.String(), nil
}

func (t *transformer) transformExprNode(ctx context.Context, node ast.ExprNode) (exprResult, error) {
	switch node := node.(type) {
	case *ast.ColumnNameExpr:
		return t.transformColumnNameExpr(ctx, node)
	case *ast.BinaryOperationExpr:
		return t.transformBinaryOperationExpr(ctx, node)
	case *ast.IsNullExpr:
		return t.transformIsNullOperationExpr(ctx, node)
	case *ast.IsTruthExpr:
		return t.transformIsTruthOperationExpr(ctx, node)
	case *ast.ParenthesesExpr:
		return t.transformParenthesesExpr(ctx, node)
	case *ast.PatternInExpr:
		return t.transformPatternInExpr(ctx, node)
	case *ast.PatternLikeOrIlikeExpr:
		return t.transformPatternLikeOrIlikeExpr(ctx, node)
	case *ast.UnaryOperationExpr:
		return t.transformUnaryOperationExpr(ctx, node)
	case ast.ValueExpr:
		return t.transformValueExpr(ctx, node)
	case *ast.FuncCallExpr:
		return t.transformFuncCallExpr(ctx, node)
	default:
		return exprResult{}, fmt.Errorf("metrics sql: unsupported expression %q", restore(node))
	}
}

func (t *transformer) transformColumnNameExpr(_ context.Context, node *ast.ColumnNameExpr) (exprResult, error) {
	if node.Name == nil {
		return exprResult{}, fmt.Errorf("metrics sql: can only have dimension/measure name(s) in select list")
	}
	if node.Name.Schema.String() != "" || node.Name.Table.String() != "" {
		return exprResult{}, fmt.Errorf("metrics sql: no alias or table reference is supported in column name. Found in `%s`", node.Name.String())
	}

	col := node.Name.Name.O
	if colExpr, ok := t.measureToExpr[col]; ok {
		return exprResult{expr: colExpr, columns: []string{restore(node.Name)}, types: []string{"MEASURE"}}, nil
	}
	var expr string
	if colExpr, ok := t.dimsToExpr[col]; ok {
		expr = colExpr
	} else if t.metricsView.Spec.TimeDimension == col {
		expr = col
	} else {
		return exprResult{}, fmt.Errorf("metrics sql: selected column `%s` not found in dimensions/measures in metrics view", col)
	}
	return exprResult{expr: expr, columns: []string{restore(node.Name)}, types: []string{"DIMENSION"}}, nil
}

func (t *transformer) transformBinaryOperationExpr(ctx context.Context, node *ast.BinaryOperationExpr) (exprResult, error) {
	left, err := t.transformExprNode(ctx, node.L)
	if err != nil {
		return exprResult{}, err
	}

	right, err := t.transformExprNode(ctx, node.R)
	if err != nil {
		return exprResult{}, err
	}

	var cols []string
	cols = append(cols, left.columns...)
	cols = append(cols, right.columns...)

	var types []string
	types = append(types, left.types...)
	types = append(types, right.types...)
	return exprResult{expr: fmt.Sprintf("%s %s %s", left.expr, opToString(node.Op), right.expr), columns: cols, types: types}, nil
}

func (t *transformer) transformFuncCallExpr(ctx context.Context, node *ast.FuncCallExpr) (exprResult, error) {
	fncName := node.FnName
	switch fncName.L {
	case "time_range_start":
		return t.transformTimeRangeStart(ctx, node)
	case "time_range_end":
		return t.transformTimeRangeEnd(ctx, node)
	case "time_range":
		return t.transformTimeRange(ctx, node)
	}

	// generic functions that do not require translation
	if _, ok := supportedFuncs[fncName.L]; !ok {
		return exprResult{}, fmt.Errorf("metrics sql: unsupported function %v", fncName.O)
	}

	var sb strings.Builder
	sb.WriteString(fncName.O)
	sb.WriteString("(")
	// keeping it generic and not doing any arg validation for now, can be added later in future if required
	var cols, types []string
	for i, arg := range node.Args {
		if i != 0 {
			sb.WriteString(", ")
		}
		expr, err := t.transformExprNode(ctx, arg)
		if err != nil {
			return exprResult{}, err
		}
		cols = append(cols, expr.columns...)
		types = append(types, expr.types...)
		sb.WriteString(expr.expr)
	}
	sb.WriteString(")")
	return exprResult{expr: sb.String(), columns: cols, types: types}, nil
}

func (t *transformer) transformIsNullOperationExpr(ctx context.Context, node *ast.IsNullExpr) (exprResult, error) {
	expr, err := t.transformExprNode(ctx, node.Expr)
	if err != nil {
		return exprResult{}, err
	}

	var sb strings.Builder
	sb.WriteString(expr.expr)
	if node.Not {
		sb.WriteString(" IS NOT NULL")
	} else {
		sb.WriteString(" IS NULL")
	}
	return exprResult{expr: sb.String(), columns: expr.columns, types: expr.types}, nil
}

func (t *transformer) transformIsTruthOperationExpr(ctx context.Context, n *ast.IsTruthExpr) (exprResult, error) {
	expr, err := t.transformExprNode(ctx, n)
	if err != nil {
		return exprResult{}, err
	}

	var sb strings.Builder
	sb.WriteString(expr.expr)
	if n.Not {
		sb.WriteString(" IS NOT")
	} else {
		sb.WriteString(" IS")
	}
	if n.True > 0 {
		sb.WriteString(" TRUE")
	} else {
		sb.WriteString(" FALSE")
	}
	return exprResult{expr: sb.String(), columns: expr.columns, types: expr.types}, nil
}

func (t *transformer) transformParenthesesExpr(ctx context.Context, node *ast.ParenthesesExpr) (exprResult, error) {
	expr, err := t.transformExprNode(ctx, node.Expr)
	if err != nil {
		return exprResult{}, err
	}
	return exprResult{expr: fmt.Sprintf("(%s)", expr), columns: expr.columns, types: expr.types}, nil
}

func (t *transformer) transformPatternInExpr(ctx context.Context, node *ast.PatternInExpr) (exprResult, error) {
	if node.Sel != nil {
		return exprResult{}, fmt.Errorf("metrics sql: sub_query is not supported")
	}

	expr, err := t.transformExprNode(ctx, node.Expr)
	if err != nil {
		return exprResult{}, err
	}

	var sb strings.Builder
	sb.WriteString(expr.expr)
	var cols, types []string
	cols = append(cols, expr.columns...)
	types = append(types, expr.types...)
	if node.Not {
		sb.WriteString(" NOT IN( ")
	} else {
		sb.WriteString(" IN( ")
	}
	for i, n := range node.List {
		if i > 0 {
			sb.WriteString(", ")
		}
		expr, err = t.transformExprNode(ctx, n)
		if err != nil {
			return exprResult{}, err
		}
		sb.WriteString(expr.expr)
		cols = append(cols, expr.columns...)
		types = append(types, expr.types...)
	}
	sb.WriteRune(')')
	return exprResult{expr: sb.String(), columns: cols, types: types}, nil
}

func (t *transformer) transformPatternLikeOrIlikeExpr(ctx context.Context, n *ast.PatternLikeOrIlikeExpr) (exprResult, error) {
	if string(n.Escape) != "\\" {
		// druid supports it, duckdb and clickhouse do not
		return exprResult{}, fmt.Errorf("metrics sql: `ESCAPE` is not supported")
	}

	expr, err := t.transformExprNode(ctx, n.Expr)
	if err != nil {
		return exprResult{}, err
	}

	var sb strings.Builder
	sb.WriteString(expr.expr)
	if n.IsLike {
		if n.Not {
			sb.WriteString(" NOT LIKE ")
		} else {
			sb.WriteString(" LIKE ")
		}
	} else {
		if n.Not {
			sb.WriteString(" NOT ILIKE ")
		} else {
			sb.WriteString(" ILIKE ")
		}
	}

	patternExpr, err := t.transformExprNode(ctx, n.Pattern)
	if err != nil {
		return exprResult{}, err
	}
	sb.WriteString(patternExpr.expr)
	var cols, types []string
	cols = append(cols, expr.columns...)
	cols = append(cols, patternExpr.columns...)
	types = append(types, expr.types...)
	types = append(types, patternExpr.types...)
	return exprResult{expr: sb.String(), columns: cols, types: types}, nil
}

func (t *transformer) transformUnaryOperationExpr(ctx context.Context, node *ast.UnaryOperationExpr) (exprResult, error) {
	expr, err := t.transformExprNode(ctx, node.V)
	if err != nil {
		return exprResult{}, err
	}

	return exprResult{expr: fmt.Sprintf("%s%s", opToString(node.Op), expr.expr), columns: expr.columns, types: expr.types}, nil
}

func (t *transformer) transformValueExpr(_ context.Context, node ast.ValueExpr) (exprResult, error) {
	var sb strings.Builder
	rctx := format.NewRestoreCtx(format.DefaultRestoreFlags|format.RestoreStringWithoutCharset, &sb)
	if err := node.Restore(rctx); err != nil {
		return exprResult{}, err
	}

	return exprResult{expr: sb.String()}, nil
}

func (t *transformer) fromQueryForMetricsView(ctx context.Context, mv *runtimev1.Resource) (string, error) {
	spec := mv.GetMetricsView().State.ValidSpec
	if spec == nil {
		return "", fmt.Errorf("metrics view %q is not ready for querying, reconcile status: %q", mv.Meta.GetName(), mv.Meta.ReconcileStatus)
	}

	olap, release, err := t.controller.Runtime.OLAP(ctx, t.instanceID, spec.Connector)
	if err != nil {
		return "", err
	}
	defer release()
	dialect := olap.Dialect()

	security, err := t.controller.Runtime.ResolveMetricsViewSecurity(t.userAttributes, t.instanceID, spec, mv.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return "", err
	}

	t.measureToExpr = make(map[string]string, len(spec.Measures))
	for _, measure := range spec.Measures {
		t.measureToExpr[measure.Name] = measure.Expression
	}

	t.dimsToExpr = make(map[string]string, len(spec.Dimensions))
	for _, dim := range spec.Dimensions {
		if dim.Expression != "" {
			t.dimsToExpr[dim.Name] = dim.Expression
		} else {
			t.dimsToExpr[dim.Name] = dialect.EscapeIdentifier(dim.Column)
		}
	}

	if security == nil {
		return dialect.EscapeIdentifier(spec.Table), nil
	}

	if !security.Access || security.ExcludeAll {
		return "", fmt.Errorf("access to metrics view %q forbidden", mv.Meta.Name.Name)
	}

	if len(security.Include) != 0 {
		for measure := range t.measureToExpr {
			if !slices.Contains(security.Include, measure) { // measures not part of include clause should not be accessible
				t.measureToExpr[measure] = "null"
			}
		}

		for dimension := range t.dimsToExpr {
			if !slices.Contains(security.Include, dimension) { // dimensions not part of include clause should not be accessible
				t.dimsToExpr[dimension] = "null"
			}
		}
	}

	for _, exclude := range security.Exclude {
		if _, ok := t.dimsToExpr[exclude]; ok {
			t.dimsToExpr[exclude] = "null"
		} else {
			t.measureToExpr[exclude] = "null"
		}
	}

	sql := "SELECT * FROM " + dialect.EscapeIdentifier(spec.Table)
	if security.RowFilter != "" {
		sql += " WHERE " + security.RowFilter
	}
	return fmt.Sprintf("(%s)", sql), nil
}

func opToString(op opcode.Op) string {
	var sb strings.Builder
	op.Format(&sb)
	return sb.String()
}

func restore(node ast.Node) string {
	var sb strings.Builder
	rctx := format.NewRestoreCtx(format.RestoreStringSingleQuotes|format.RestoreKeyWordUppercase|format.RestoreNameDoubleQuotes|format.RestoreStringWithoutCharset, &sb)
	_ = node.Restore(rctx)
	return sb.String()
}

type exprResult struct {
	expr string
	// underlying dimension/measure name
	// can be empty for constants
	columns []string
	// typ DIMENSION/MEASURE
	// can be empty for constants
	types []string
}
