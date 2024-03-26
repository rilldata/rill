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
	"golang.org/x/exp/maps"

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
}

// New returns a compiler and created a tidb parser object.
// The instantiated parser object and thus compiler is not goroutine safe and not lightweight.
// It is better to keep it in a single goroutine, and reuse it if possible.
func New(ctrl *runtime.Controller, instanceID string, userAttributes map[string]any) *Compiler {
	p := parser.New()
	// Weirdly setting just ModeANSI which is a combination having ModeANSIQuotes doesn't ensure double quotes are used to identify SQL identifiers
	p.SetSQLMode(mysql.ModeANSI | mysql.ModeANSIQuotes)
	return &Compiler{
		p:              p,
		controller:     ctrl,
		instanceID:     instanceID,
		userAttributes: userAttributes,
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
		controller:       c.controller,
		instanceID:       c.instanceID,
		userAttributes:   c.userAttributes,
		dimExprToCol:     make(map[string]string),
		measureExprToCol: make(map[string]string),
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

	// resolved dimension expression to dimension name
	dimExprToCol map[string]string
	// resolved measure expression to measure name
	measureExprToCol map[string]string
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
	if node.From.TableRefs == nil {
		return "", fmt.Errorf("metrics sql: need from clause")
	}

	var sb strings.Builder
	sb.WriteString("SELECT ")
	fromClause, err := t.transformFromClause(ctx, node.From)
	if err != nil {
		return "", err
	}

	selectList, err := t.transformSelectStmtColumns(ctx, node)
	if err != nil {
		return "", err
	}

	sb.WriteString(selectList)
	sb.WriteString(" FROM ")
	sb.WriteString(fromClause)

	if node.Where != nil {
		where, err := t.transformExprNode(ctx, node.Where, false)
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(where)
	}
	if len(t.measureExprToCol) > 0 && len(t.dimExprToCol) > 0 {
		if node.GroupBy != nil {
			return "", fmt.Errorf("metrics sql: group by clause is implicitly added when any measure is selected. The implicit group by includes all selected dimensions")
		}

		sb.WriteString(" GROUP BY ")
		dimExprList := maps.Keys(t.dimExprToCol)
		// sort so that consistent group by sql is produced
		slices.Sort(dimExprList)
		for i, expr := range dimExprList {
			if i != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(expr)
		}
	}

	if node.Having != nil {
		having, err := t.transformHavingClause(ctx, node.Having, false)
		if err != nil {
			return "", err
		}
		sb.WriteString(" HAVING ")
		sb.WriteString(having)
	}

	if node.OrderBy != nil {
		orderBy, err := t.transformOrderByClause(ctx, node.OrderBy, false)
		if err != nil {
			return "", err
		}
		sb.WriteString(" ORDER BY ")
		sb.WriteString(orderBy)
	}

	if node.Limit != nil {
		limit, err := t.transformLimitClause(ctx, node.Limit, false)
		if err != nil {
			return "", err
		}
		sb.WriteString(" LIMIT ")
		sb.WriteString(limit)
	}
	return sb.String(), nil
}

func (t *transformer) transformSelectStmtColumns(ctx context.Context, node *ast.SelectStmt) (string, error) {
	if len(node.Fields.Fields) == 0 {
		return "", fmt.Errorf("metrics sql: need to select atleast one dimension or measure")
	}

	var sb strings.Builder
	for i, field := range node.Fields.Fields {
		if i != 0 {
			sb.WriteString(", ")
		} else if node.Distinct {
			sb.WriteString("DISTINCT ")
		}
		if field.WildCard != nil {
			return "", fmt.Errorf("metrics sql: wildcard is not supported")
		}

		expr, err := t.transformExprNode(ctx, field.Expr, true)
		if err != nil {
			return "", err
		}

		sb.WriteString(expr)
		// write alias if any
		if field.AsName.String() != "" { // if explicitly specified in the metrics_sql
			sb.WriteString(" AS \"")
			sb.WriteString(field.AsName.String())
			sb.WriteString("\"")
		} else {
			var name string
			// selecting a plain dimension or measure adds dimension/measure name as alias
			if col, ok := t.dimExprToCol[expr]; ok {
				name = col
			} else if col, ok := t.measureExprToCol[expr]; ok {
				name = col
			} else {
				continue
			}
			sb.WriteString(" AS ")
			sb.WriteString(name)
		}
	}
	return sb.String(), nil
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

func (t *transformer) transformHavingClause(ctx context.Context, node *ast.HavingClause, updateSelect bool) (string, error) {
	return t.transformExprNode(ctx, node.Expr, updateSelect)
}

func (t *transformer) transformOrderByClause(ctx context.Context, node *ast.OrderByClause, updateSelect bool) (string, error) {
	var sb strings.Builder
	for i, item := range node.Items {
		if i != 0 {
			sb.WriteString(", ")
		}
		expr, err := t.transformExprNode(ctx, item.Expr, updateSelect)
		if err != nil {
			return "", err
		}
		if item.Desc {
			sb.WriteString(expr + " DESC")
		} else {
			sb.WriteString(expr + " ASC")
		}
	}
	return sb.String(), nil
}

func (t *transformer) transformLimitClause(ctx context.Context, node *ast.Limit, updateSelect bool) (string, error) {
	count, err := t.transformExprNode(ctx, node.Count, updateSelect)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(count)
	if node.Offset != nil {
		offset, err := t.transformExprNode(ctx, node.Offset, updateSelect)
		if err != nil {
			return "", err
		}
		sb.WriteString(" OFFSET ")
		sb.WriteString(offset)
	}

	return sb.String(), nil
}

func (t *transformer) transformExprNode(ctx context.Context, node ast.ExprNode, updateSelect bool) (string, error) {
	switch node := node.(type) {
	case *ast.ColumnNameExpr:
		return t.transformColumnNameExpr(ctx, node, updateSelect)
	case *ast.BinaryOperationExpr:
		return t.transformBinaryOperationExpr(ctx, node, updateSelect)
	case *ast.IsNullExpr:
		return t.transformIsNullOperationExpr(ctx, node, updateSelect)
	case *ast.IsTruthExpr:
		return t.transformIsTruthOperationExpr(ctx, node, updateSelect)
	case *ast.ParenthesesExpr:
		return t.transformParenthesesExpr(ctx, node, updateSelect)
	case *ast.PatternInExpr:
		return t.transformPatternInExpr(ctx, node, updateSelect)
	case *ast.PatternLikeOrIlikeExpr:
		return t.transformPatternLikeOrIlikeExpr(ctx, node)
	case *ast.UnaryOperationExpr:
		return t.transformUnaryOperationExpr(ctx, node, updateSelect)
	case ast.ValueExpr:
		return t.transformValueExpr(ctx, node)
	case *ast.FuncCallExpr:
		return t.transformFuncCallExpr(ctx, node, updateSelect)
	default:
		return "", fmt.Errorf("metrics sql: unsupported expression %q", restore(node))
	}
}

func (t *transformer) transformColumnNameExpr(_ context.Context, node *ast.ColumnNameExpr, updateSelect bool) (string, error) {
	if node.Name == nil {
		return "", fmt.Errorf("metrics sql: can only have dimension/measure name(s) in select list")
	}
	if node.Name.Schema.String() != "" || node.Name.Table.String() != "" {
		return "", fmt.Errorf("metrics sql: no alias or table reference is supported in column name. Found in `%s`", node.Name.String())
	}

	col := node.Name.Name.O
	if colExpr, ok := t.measureToExpr[col]; ok {
		if updateSelect {
			t.measureExprToCol[colExpr] = restore(node.Name) // makes sure double quotes are added
		}
		return colExpr, nil
	}
	var expr string
	if colExpr, ok := t.dimsToExpr[col]; ok {
		expr = colExpr
	} else if t.metricsView.Spec.TimeDimension == col {
		expr = col
	} else {
		return "", fmt.Errorf("metrics sql: selected column `%s` not found in dimensions/measures in metrics view", col)
	}
	if updateSelect {
		t.dimExprToCol[expr] = restore(node.Name) // makes sure double quotes are added
	}
	return expr, nil
}

func (t *transformer) transformBinaryOperationExpr(ctx context.Context, node *ast.BinaryOperationExpr, updateSelect bool) (string, error) {
	left, err := t.transformExprNode(ctx, node.L, updateSelect)
	if err != nil {
		return "", err
	}

	right, err := t.transformExprNode(ctx, node.R, updateSelect)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s %s", left, opToString(node.Op), right), nil
}

func (t *transformer) transformFuncCallExpr(ctx context.Context, node *ast.FuncCallExpr, updateSelect bool) (string, error) {
	fncName := node.FnName
	if _, ok := supportedFuncs[fncName.L]; !ok {
		return "", fmt.Errorf("metrics sql: unsupported function %v", fncName.O)
	}

	var sb strings.Builder
	sb.WriteString(fncName.O)
	sb.WriteString("(")
	// keeping it generic and not doing any arg validation for now, can be added later in future if required
	for i, arg := range node.Args {
		if i != 0 {
			sb.WriteString(", ")
		}
		expr, err := t.transformExprNode(ctx, arg, updateSelect)
		if err != nil {
			return "", err
		}

		sb.WriteString(expr)
	}
	sb.WriteString(")")
	return sb.String(), nil
}

func (t *transformer) transformIsNullOperationExpr(ctx context.Context, node *ast.IsNullExpr, updateSelect bool) (string, error) {
	expr, err := t.transformExprNode(ctx, node.Expr, updateSelect)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(expr)
	if node.Not {
		sb.WriteString(" IS NOT NULL")
	} else {
		sb.WriteString(" IS NULL")
	}
	return sb.String(), nil
}

func (t *transformer) transformIsTruthOperationExpr(ctx context.Context, n *ast.IsTruthExpr, updateSelect bool) (string, error) {
	expr, err := t.transformExprNode(ctx, n, updateSelect)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(expr)
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
	return sb.String(), nil
}

func (t *transformer) transformParenthesesExpr(ctx context.Context, node *ast.ParenthesesExpr, updateSelect bool) (string, error) {
	expr, err := t.transformExprNode(ctx, node.Expr, updateSelect)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s)", expr), nil
}

func (t *transformer) transformPatternInExpr(ctx context.Context, node *ast.PatternInExpr, updateSelect bool) (string, error) {
	if node.Sel != nil {
		return "", fmt.Errorf("metrics sql: sub_query is not supported")
	}

	expr, err := t.transformExprNode(ctx, node.Expr, updateSelect)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(expr)
	if node.Not {
		sb.WriteString(" NOT IN( ")
	} else {
		sb.WriteString(" IN( ")
	}
	for i, n := range node.List {
		if i > 0 {
			sb.WriteString(", ")
		}
		expr, err = t.transformExprNode(ctx, n, updateSelect)
		if err != nil {
			return "", err
		}
		sb.WriteString(expr)
	}
	sb.WriteRune(')')
	return sb.String(), nil
}

func (t *transformer) transformPatternLikeOrIlikeExpr(ctx context.Context, n *ast.PatternLikeOrIlikeExpr) (string, error) {
	if string(n.Escape) != "\\" {
		// druid supports it, duckdb and clickhouse do not
		return "", fmt.Errorf("metrics sql: `ESCAPE` is not supported")
	}

	expr, err := t.transformExprNode(ctx, n.Expr, false)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(expr)
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

	patternExpr, err := t.transformExprNode(ctx, n.Pattern, false)
	if err != nil {
		return "", err
	}
	sb.WriteString(patternExpr)
	return sb.String(), nil
}

func (t *transformer) transformUnaryOperationExpr(ctx context.Context, node *ast.UnaryOperationExpr, updateSelect bool) (string, error) {
	expr, err := t.transformExprNode(ctx, node.V, updateSelect)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", opToString(node.Op), expr), nil
}

func (t *transformer) transformValueExpr(_ context.Context, node ast.ValueExpr) (string, error) {
	var sb strings.Builder
	rctx := format.NewRestoreCtx(format.DefaultRestoreFlags|format.RestoreStringWithoutCharset, &sb)
	if err := node.Restore(rctx); err != nil {
		return "", err
	}

	return sb.String(), nil
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
