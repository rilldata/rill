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
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
)

var supportedOpCodes = map[opcode.Op]any{
	opcode.LogicAnd: nil,
	opcode.LogicOr:  nil,
	opcode.GE:       nil,
	opcode.LE:       nil,
	opcode.EQ:       nil,
	opcode.NE:       nil,
	opcode.LT:       nil,
	opcode.GT:       nil,
	opcode.And:      nil,
	opcode.Or:       nil,
	opcode.Not2:     nil,
	opcode.In:       nil,
	opcode.Like:     nil,
	opcode.IsNull:   nil,
}

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
	p.SetSQLMode(mysql.ModeANSI | mysql.ModeANSIQuotes)
	return &Compiler{
		p:              p,
		controller:     ctrl,
		instanceID:     instanceID,
		userAttributes: userAttributes,
	}
}

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

	s := &walker{
		controller:       c.controller,
		instanceID:       c.instanceID,
		userAttributes:   c.userAttributes,
		dimExprToCol:     make(map[string]string),
		measureExprToCol: make(map[string]string),
	}
	compiledSQL, err := s.walkSelectStmt(ctx, stmt)
	if err != nil {
		return "", "", nil, err
	}

	return compiledSQL, s.connector, s.refs, nil
}

type walker struct {
	controller     *runtime.Controller
	instanceID     string
	userAttributes map[string]any

	metricsView   *runtimev1.MetricsViewV2
	refs          []*runtimev1.ResourceName
	connector     string
	dimsToExpr    map[string]string
	measureToExpr map[string]string

	dimExprToCol     map[string]string
	measureExprToCol map[string]string
}

func (s *walker) walkSelectStmt(ctx context.Context, node *ast.SelectStmt) (string, error) {
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
	fromClause, err := s.walkFromClause(ctx, node.From)
	if err != nil {
		return "", err
	}

	selectList, err := s.walkSelectStmtColumns(ctx, node)
	if err != nil {
		return "", err
	}

	sb.WriteString(selectList)
	sb.WriteString(" FROM ")
	sb.WriteString(fromClause)

	if node.Where != nil {
		where, err := s.walkExprNode(ctx, node.Where, false)
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(where)
	}
	if len(s.measureExprToCol) > 0 {
		if node.GroupBy != nil {
			return "", fmt.Errorf("metrics sql: group by clause is implicitly added when any measure is selected. The implicit group by includes all selected dimensions")
		}

		sb.WriteString(" GROUP BY ")
		dimExprList := maps.Keys(s.dimExprToCol)
		// sort so that consistent group by sql is produced
		slices.Sort(dimExprList)
		for i, expr := range dimExprList {
			if i != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(expr)
		}
	} // todo :: support group by if only dimensions are selected

	if node.Having != nil {
		having, err := s.walkHavingClause(ctx, node.Having, false)
		if err != nil {
			return "", err
		}
		sb.WriteString(" HAVING ")
		sb.WriteString(having)
	}

	if node.OrderBy != nil {
		orderBy, err := s.walkOrderByClause(ctx, node.OrderBy, false)
		if err != nil {
			return "", err
		}
		sb.WriteString(" ORDER BY ")
		sb.WriteString(orderBy)
	}

	if node.Limit != nil {
		limit, err := s.walkLimitClause(ctx, node.Limit, false)
		if err != nil {
			return "", err
		}
		sb.WriteString(" LIMIT ")
		sb.WriteString(limit)
	}
	return sb.String(), nil
}

func (s *walker) walkSelectStmtColumns(ctx context.Context, node *ast.SelectStmt) (string, error) {
	if len(node.Fields.Fields) == 0 {
		return "", fmt.Errorf("metrics sql: need to select atleast one dimension or measure")
	}

	var sb strings.Builder
	for i, field := range node.Fields.Fields {
		if i != 0 {
			sb.WriteString(", ")
		}
		if field.WildCard != nil {
			return "", fmt.Errorf("metrics sql: wildcard is not supported")
		}

		expr, err := s.walkExprNode(ctx, field.Expr, true)
		if err != nil {
			return "", err
		}

		sb.WriteString(expr)
		// write alias if any
		if field.AsName.String() != "" { // if explicitly specified in the metrics_sql
			sb.WriteString(" AS ")
			sb.WriteString(field.AsName.String())
		} else {
			var name string
			// selecting a plain dimension or measure adds dimension/measure name as alias
			if col, ok := s.dimExprToCol[expr]; ok {
				name = col
			} else if col, ok := s.measureExprToCol[expr]; ok {
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

func (s *walker) walkFromClause(ctx context.Context, node *ast.TableRefsClause) (string, error) {
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
	mv, err := s.controller.Get(ctx, resource, false)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return "", fmt.Errorf("metrics sql: metrics view `%s` not found", tblName.Name.String())
		}
		return "", err
	}

	s.metricsView = mv.GetMetricsView()
	s.refs = []*runtimev1.ResourceName{resource}
	s.connector = mv.GetMetricsView().Spec.Connector
	return s.fromQueryForMetricsView(mv)
}

func (s *walker) walkHavingClause(ctx context.Context, node *ast.HavingClause, updateSelect bool) (string, error) {
	return s.walkExprNode(ctx, node.Expr, updateSelect)
}

func (s *walker) walkOrderByClause(ctx context.Context, node *ast.OrderByClause, updateSelect bool) (string, error) {
	var sb strings.Builder
	for i, item := range node.Items {
		if i != 0 {
			sb.WriteString(", ")
		}
		expr, err := s.walkExprNode(ctx, item.Expr, updateSelect)
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

func (s *walker) walkLimitClause(ctx context.Context, node *ast.Limit, updateSelect bool) (string, error) {
	count, err := s.walkExprNode(ctx, node.Count, updateSelect)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(count)
	if node.Offset != nil {
		offset, err := s.walkExprNode(ctx, node.Offset, updateSelect)
		if err != nil {
			return "", err
		}
		sb.WriteString(" OFFSET ")
		sb.WriteString(offset)
	}

	return sb.String(), nil
}

func (s *walker) walkExprNode(ctx context.Context, node ast.ExprNode, updateSelect bool) (string, error) {
	switch node := node.(type) {
	case *ast.ColumnNameExpr:
		return s.walkColumnNameExpr(ctx, node, updateSelect)
	case *ast.BinaryOperationExpr:
		return s.walkBinaryOperationExpr(ctx, node, updateSelect)
	case *ast.IsNullExpr:
		return s.walkIsNullOperationExpr(ctx, node, updateSelect)
	case *ast.IsTruthExpr:
		return s.walkIsTruthOperationExpr(ctx, node, updateSelect)
	case *ast.ParenthesesExpr:
		return s.walkParenthesesExpr(ctx, node, updateSelect)
	case *ast.PatternInExpr:
		return s.walkPatternInExpr(ctx, node, updateSelect)
	// case *ast.PatternLikeOrIlikeExpr: // todo :: fix
	// 	return s.walkPatternLikeOrIlikeExpr(ctx, node)
	case *ast.UnaryOperationExpr:
		return s.walkUnaryOperationExpr(ctx, node, updateSelect)
	case ast.ValueExpr:
		return s.walkValueExpr(ctx, node)
	case *ast.FuncCallExpr:
		return s.walkFuncCallExpr(ctx, node, updateSelect)
	default:
		return "", fmt.Errorf("metrics sql: unsupported expression %q", restore(node))
	}
}

func (s *walker) walkColumnNameExpr(_ context.Context, node *ast.ColumnNameExpr, updateSelect bool) (string, error) {
	if node.Name == nil {
		return "", fmt.Errorf("metrics sql: can only have dimension/measure name(s) in select list")
	}
	if node.Name.Schema.String() != "" || node.Name.Table.String() != "" {
		return "", fmt.Errorf("metrics sql: no alias or table reference is supported in column name. Found in `%s`", node.Name.String())
	}

	col := node.Name.Name.O
	if colExpr, ok := s.measureToExpr[col]; ok {
		if updateSelect {
			s.measureExprToCol[colExpr] = restore(node.Name) // makes sure double quotes are added
		}
		return colExpr, nil
	}
	var expr string
	if colExpr, ok := s.dimsToExpr[col]; ok {
		expr = colExpr
	} else if s.metricsView.Spec.TimeDimension == col {
		expr = col
	} else {
		return "", fmt.Errorf("metrics sql: selected column `%s` not found in dimensions/measures in metrics view", col)
	}
	if updateSelect {
		s.dimExprToCol[expr] = restore(node.Name) // makes sure double quotes are added
	}
	return expr, nil
}

func (s *walker) walkBinaryOperationExpr(ctx context.Context, node *ast.BinaryOperationExpr, updateSelect bool) (string, error) {
	left, err := s.walkExprNode(ctx, node.L, updateSelect)
	if err != nil {
		return "", err
	}

	right, err := s.walkExprNode(ctx, node.R, updateSelect)
	if err != nil {
		return "", err
	}

	if _, ok := supportedOpCodes[node.Op]; !ok {
		return "", fmt.Errorf("metrics sql: unsupported operator %q", opToString(node.Op))
	}
	return fmt.Sprintf("%s %s %s", left, opToString(node.Op), right), nil
}

func (s *walker) walkFuncCallExpr(ctx context.Context, node *ast.FuncCallExpr, updateSelect bool) (string, error) {
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
		expr, err := s.walkExprNode(ctx, arg, updateSelect)
		if err != nil {
			return "", err
		}

		sb.WriteString(expr)
	}
	sb.WriteString(")")
	return sb.String(), nil
}

func (s *walker) walkIsNullOperationExpr(ctx context.Context, node *ast.IsNullExpr, updateSelect bool) (string, error) {
	expr, err := s.walkExprNode(ctx, node.Expr, updateSelect)
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

func (s *walker) walkIsTruthOperationExpr(ctx context.Context, n *ast.IsTruthExpr, updateSelect bool) (string, error) {
	expr, err := s.walkExprNode(ctx, n, updateSelect)
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

func (s *walker) walkParenthesesExpr(ctx context.Context, node *ast.ParenthesesExpr, updateSelect bool) (string, error) {
	expr, err := s.walkExprNode(ctx, node.Expr, updateSelect)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s)", expr), nil
}

func (s *walker) walkPatternInExpr(ctx context.Context, node *ast.PatternInExpr, updateSelect bool) (string, error) {
	if node.Sel != nil {
		return "", fmt.Errorf("metrics sql: `IN` is not supported with `SELECT` clause")
	}

	expr, err := s.walkExprNode(ctx, node.Expr, updateSelect)
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
		expr, err = s.walkExprNode(ctx, n, updateSelect)
		if err != nil {
			return "", err
		}
		sb.WriteString(expr)
	}
	sb.WriteRune(')')
	return sb.String(), nil
}

// func (s *stateFlow) walkPatternLikeOrIlikeExpr(ctx context.Context, n *ast.PatternLikeOrIlikeExpr) (string, error) {
// }

func (s *walker) walkUnaryOperationExpr(ctx context.Context, node *ast.UnaryOperationExpr, updateSelect bool) (string, error) {
	expr, err := s.walkExprNode(ctx, node.V, updateSelect)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", opToString(node.Op), expr), nil
}

func (s *walker) walkValueExpr(_ context.Context, node ast.ValueExpr) (string, error) {
	// todo :: fix this
	var sb strings.Builder
	rctx := format.NewRestoreCtx(format.DefaultRestoreFlags|format.RestoreStringWithoutCharset, &sb)
	if err := node.Restore(rctx); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func (s *walker) fromQueryForMetricsView(mv *runtimev1.Resource) (string, error) {
	// Dialect to use for escaping identifiers. Currently hardcoded to DuckDB.
	// TODO: Make dynamic based on actual dialect of the OLAP connector used by the referenced metrics view.
	dialect := drivers.DialectDuckDB

	spec := mv.GetMetricsView().State.ValidSpec
	if spec == nil {
		return "", fmt.Errorf("metrics view %q is not ready for querying, reconcile status: %q", mv.Meta.GetName(), mv.Meta.ReconcileStatus)
	}

	security, err := s.controller.Runtime.ResolveMetricsViewSecurity(s.userAttributes, s.instanceID, spec, mv.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return "", err
	}

	s.measureToExpr = make(map[string]string, len(spec.Measures))
	for _, measure := range spec.Measures {
		s.measureToExpr[measure.Name] = measure.Expression
	}

	s.dimsToExpr = make(map[string]string, len(spec.Dimensions))
	for _, dim := range spec.Dimensions {
		if dim.Expression != "" {
			s.dimsToExpr[dim.Name] = dim.Expression
		} else {
			s.dimsToExpr[dim.Name] = dialect.EscapeIdentifier(dim.Column)
		}
	}

	if security == nil {
		return dialect.EscapeIdentifier(spec.Table), nil
	}

	if !security.Access || security.ExcludeAll {
		return "", fmt.Errorf("access to metrics view %q forbidden", mv.Meta.Name.Name)
	}

	if len(security.Include) != 0 {
		for measure := range s.measureToExpr {
			if !slices.Contains(security.Include, measure) { // measures not part of include clause should not be accessible
				s.measureToExpr[measure] = "null"
			}
		}

		for dimension := range s.dimsToExpr {
			if !slices.Contains(security.Include, dimension) { // dimensions not part of include clause should not be accessible
				s.dimsToExpr[dimension] = "null"
			}
		}
	}

	for _, exclude := range security.Exclude {
		if _, ok := s.dimsToExpr[exclude]; ok {
			s.dimsToExpr[exclude] = "null"
		} else {
			s.measureToExpr[exclude] = "null"
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
