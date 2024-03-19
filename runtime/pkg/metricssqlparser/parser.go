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

type Compiler struct {
	p *parser.Parser
}

func New() *Compiler {
	p := parser.New()
	p.SetSQLMode(mysql.ModeANSI)
	return &Compiler{p: p}
}

func (c *Compiler) Compile(ctrl *runtime.Controller, instanceID, sql string, userAttributes map[string]any) (string, string, []*runtimev1.ResourceName, error) {
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

	s := &walker{controller: ctrl, instanceID: instanceID, userAttributes: userAttributes}
	compiledSQL, err := s.walkSelectStmt(context.Background(), stmt)
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

	dimExpressions     []string
	measureExpressions []string
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
		where, err := s.walkExprNode(ctx, node.Where)
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHERE ")
		sb.WriteString(where)
	}
	if len(s.measureExpressions) > 0 {
		if node.GroupBy != nil {
			return "", fmt.Errorf("metrics sql: group by clause is implicitly added when any measure is selected. The implicit group by includes all selected dimensions")
		}

		sb.WriteString(" GROUP BY ")
		for i, dim := range s.dimExpressions {
			if i != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(dim)
		}
	}

	if node.Having != nil {
		having, err := s.walkHavingClause(ctx, node.Having)
		if err != nil {
			return "", err
		}
		sb.WriteString(" HAVING ")
		sb.WriteString(having)
	}

	if node.OrderBy != nil {
		orderBy, err := s.walkOrderByClause(ctx, node.OrderBy)
		if err != nil {
			return "", err
		}
		sb.WriteString(" ORDER BY ")
		sb.WriteString(orderBy)
	}

	if node.Limit != nil {
		limit, err := s.walkLimitClause(ctx, node.Limit)
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

		colExpr, ok := field.Expr.(*ast.ColumnNameExpr)
		if !ok {
			return "", fmt.Errorf("metrics sql: can only select plain dimension/measures")
		}

		col, expr, typ, err := s.walkColumnNameExpr(ctx, colExpr)
		if err != nil {
			return "", err
		}

		sb.WriteString(expr)
		sb.WriteString(" AS ")
		if field.AsName.String() != "" {
			sb.WriteString(field.AsName.String())
		} else {
			sb.WriteString(col)
		}
		if typ == "MEASURE" {
			s.measureExpressions = append(s.measureExpressions, expr)
		} else {
			s.dimExpressions = append(s.dimExpressions, expr)
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

func (s *walker) walkExprNode(ctx context.Context, node ast.ExprNode) (string, error) {
	switch node := node.(type) {
	case *ast.ColumnNameExpr:
		_, expr, _, err := s.walkColumnNameExpr(ctx, node)
		return expr, err
	case *ast.BinaryOperationExpr:
		return s.walkBinaryOperationExpr(ctx, node)
	case *ast.IsNullExpr:
		return s.walkIsNullOperationExpr(ctx, node)
	case *ast.IsTruthExpr:
		return s.walkIsTruthOperationExpr(ctx, node)
	case *ast.ParenthesesExpr:
		return s.walkParenthesesExpr(ctx, node)
	case *ast.PatternInExpr:
		return s.walkPatternInExpr(ctx, node)
	// case *ast.PatternLikeOrIlikeExpr: // todo :: fix
	// 	return s.walkPatternLikeOrIlikeExpr(ctx, node)
	case *ast.UnaryOperationExpr:
		return s.walkUnaryOperationExpr(ctx, node)
	case ast.ValueExpr:
		return s.walkValueExpr(ctx, node)
	default:
		var sb strings.Builder
		rctx := format.NewRestoreCtx(format.DefaultRestoreFlags, &sb)
		_ = node.Restore(rctx)
		return "", fmt.Errorf("metrics sql: unsupported expression %q", sb.String())
	}
}

func (s *walker) walkColumnNameExpr(_ context.Context, node *ast.ColumnNameExpr) (col, expr, typ string, err error) {
	if node.Name == nil {
		return "", "", "", fmt.Errorf("metrics sql: can only have dimension/measure name(s) in select list")
	}
	if node.Name.Schema.String() != "" || node.Name.Table.String() != "" {
		return "", "", "", fmt.Errorf("metrics sql: no alias or table reference is supported in column name. Found in `%s`", node.Name.String())
	}

	col = node.Name.String()
	if colExpr, ok := s.dimsToExpr[col]; ok {
		expr = colExpr
		typ = "DIMENSION"
	} else if colExpr, ok := s.measureToExpr[col]; ok {
		expr = colExpr
		typ = "MEASURE"
	} else if s.metricsView.Spec.TimeDimension == col {
		expr = col
		typ = "TIMEDIMENSION"
	} else {
		err = fmt.Errorf("metrics sql: selected column `%s` not found in dimensions/measures in metrics view", col)
	}
	return col, expr, typ, err
}

func (s *walker) walkBinaryOperationExpr(ctx context.Context, node *ast.BinaryOperationExpr) (string, error) {
	left, err := s.walkExprNode(ctx, node.L)
	if err != nil {
		return "", err
	}

	right, err := s.walkExprNode(ctx, node.R)
	if err != nil {
		return "", err
	}

	if _, ok := supportedOpCodes[node.Op]; !ok {
		return "", fmt.Errorf("metrics sql: unsupported operator %q", opToString(node.Op))
	}
	return fmt.Sprintf("%s %s %s", left, opToString(node.Op), right), nil
}

func (s *walker) walkIsNullOperationExpr(ctx context.Context, node *ast.IsNullExpr) (string, error) {
	expr, err := s.walkExprNode(ctx, node.Expr)
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

func (s *walker) walkIsTruthOperationExpr(ctx context.Context, n *ast.IsTruthExpr) (string, error) {
	expr, err := s.walkExprNode(ctx, n)
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

func (s *walker) walkParenthesesExpr(ctx context.Context, node *ast.ParenthesesExpr) (string, error) {
	expr, err := s.walkExprNode(ctx, node.Expr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s)", expr), nil
}

func (s *walker) walkPatternInExpr(ctx context.Context, node *ast.PatternInExpr) (string, error) {
	if node.Sel != nil {
		return "", fmt.Errorf("metrics sql: `IN` is not supported with `SELECT` clause")
	}

	expr, err := s.walkExprNode(ctx, node.Expr)
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
		expr, err = s.walkExprNode(ctx, n)
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

func (s *walker) walkUnaryOperationExpr(ctx context.Context, node *ast.UnaryOperationExpr) (string, error) {
	expr, err := s.walkExprNode(ctx, node.V)
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

func (s *walker) walkHavingClause(ctx context.Context, node *ast.HavingClause) (string, error) {
	return s.walkExprNode(ctx, node.Expr)
}

func (s *walker) walkOrderByClause(ctx context.Context, node *ast.OrderByClause) (string, error) {
	var sb strings.Builder
	for i, item := range node.Items {
		if i != 0 {
			sb.WriteString(", ")
		}
		expr, err := s.walkExprNode(ctx, item.Expr)
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

func (s *walker) walkLimitClause(ctx context.Context, node *ast.Limit) (string, error) {
	count, err := s.walkExprNode(ctx, node.Count)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(count)
	if node.Offset != nil {
		offset, err := s.walkExprNode(ctx, node.Offset)
		if err != nil {
			return "", err
		}
		sb.WriteString(" OFFSET ")
		sb.WriteString(offset)
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
