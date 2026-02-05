package metricssql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/parser/opcode"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"

	// need to import parser driver as well
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

// Compiler is responsible for parsing metrics SQL queries into metricsview.Query objects.
//
// Internally, it creates and re-uses a TiDB parser object.
// It's not lightweight, so re-use the Compiler when possible.
// However, note that it is not concurrency safe.
type Compiler struct {
	p    *parser.Parser
	opts *CompilerOptions
}

// CompilerOptions provide options for the Compiler.
type CompilerOptions struct {
	// GetMetricsView is a callback to lookup a referenced metrics view.
	// It is required for parsing full queries, but optional when parsing only filters.
	GetMetricsView func(ctx context.Context, name string) (*runtimev1.Resource, error)
	// GetTimestamps is a callback to resolve timestamps for a given time dimension.
	// It is optional, but if not provided, queries that use rilltime expressions will error.
	// TODO: Ideally we should replace this with support for rilltime expressions in *metricsview.Expression itself, so evaluation can be delayed until query execution.
	GetTimestamps func(ctx context.Context, mv *runtimev1.Resource, timeDim string) (metricsview.TimestampsResult, error)
}

// New creates a new Compiler.
func New(opts *CompilerOptions) *Compiler {
	p := parser.New()
	// Weirdly setting just ModeANSI which is a combination having ModeANSIQuotes doesn't ensure double quotes are used to identify SQL identifiers
	p.SetSQLMode(mysql.ModeANSI | mysql.ModeANSIQuotes)

	return &Compiler{
		p:    p,
		opts: opts,
	}
}

// Parse parses a metrics SQL query into a metricview.Query.
// It uses the tidb parser (which is a MySQL compliant parser) and transforms over the generated AST to generate query.
// We use MySQL's ANSI SQL Mode to conform more closely to standard SQL.
//
// Whenever adding transform method over new node type also look at its `Restore` method to get an idea how it can be parsed into a SQL query.
func (c *Compiler) Parse(ctx context.Context, sql string) (*metricsview.Query, error) {
	stmtNodes, _, err := c.p.ParseSQL(sql)
	if err != nil {
		return nil, err
	}

	if len(stmtNodes) != 1 {
		return nil, errors.New("metrics sql: expected exactly one SQL statement")
	}

	stmt, ok := stmtNodes[0].(*ast.SelectStmt)
	if !ok {
		return nil, errors.New("metrics sql: expected a SELECT statement")
	}

	q := &query{
		q:    &metricsview.Query{},
		opts: c.opts,
	}

	// parse from clause
	if err := q.parseFrom(ctx, stmt.From); err != nil {
		return nil, err
	}

	// parse select fields
	if err := q.parseSelect(stmt.Fields); err != nil {
		return nil, err
	}

	// parse where clause
	if stmt.Where != nil {
		expr, err := parseFilter(ctx, stmt.Where, nil, q)
		if err != nil {
			return nil, err
		}
		q.q.Where = expr
	}

	// parse limit clause
	if stmt.Limit != nil {
		if err := q.parseLimit(stmt.Limit); err != nil {
			return nil, err
		}
	}

	// parse order by
	if stmt.OrderBy != nil {
		if err := q.parseOrderBy(stmt.OrderBy); err != nil {
			return nil, err
		}
	}

	// parse having
	if stmt.Having != nil {
		if err := q.parseHaving(ctx, stmt.Having); err != nil {
			return nil, err
		}
	}
	return q.q, nil
}

type query struct {
	q    *metricsview.Query
	opts *CompilerOptions

	// fields available after parsing FROM clause
	metricsView     *runtimev1.Resource
	metricsViewSpec *runtimev1.MetricsViewSpec
	dims            map[string]any
	measures        map[string]any
}

func (q *query) parseFrom(ctx context.Context, node *ast.TableRefsClause) error {
	n := node.TableRefs
	if n == nil || n.Left == nil {
		return fmt.Errorf("metrics sql: need `FROM metrics_view` clause")
	}

	tblSrc, ok := n.Left.(*ast.TableSource)
	if !ok {
		// if left is not a table source, then it must be a join
		return fmt.Errorf("metrics sql: join is not supported")
	}

	tblName, ok := tblSrc.Source.(*ast.TableName)
	if !ok {
		return fmt.Errorf("metrics sql: only FROM `metrics_view` is supported")
	}

	if q.opts == nil || q.opts.GetMetricsView == nil {
		return fmt.Errorf("metrics sql: must provide the GetMetricsView option to the compiler")
	}
	mv, err := q.opts.GetMetricsView(ctx, tblName.Name.String())
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return fmt.Errorf("metrics sql: metrics view `%s` not found", tblName.Name.String())
		}
		return err
	}
	q.metricsView = mv
	q.q.MetricsView = mv.Meta.Name.Name

	spec := mv.GetMetricsView().State.ValidSpec
	if spec == nil {
		return fmt.Errorf("metrics view %q is not valid", mv.Meta.Name.Name)
	}
	q.metricsViewSpec = spec

	q.measures = make(map[string]any, len(spec.Measures))
	for _, measure := range spec.Measures {
		q.measures[measure.Name] = nil
	}

	q.dims = make(map[string]any, len(spec.Dimensions))
	for _, dim := range spec.Dimensions {
		q.dims[dim.Name] = nil
	}
	return nil
}

func (q *query) parseSelect(node *ast.FieldList) error {
	for _, field := range node.Fields {
		switch v := field.Expr.(type) {
		case *ast.ColumnNameExpr:
			// TODO no alias handling for column names
			col, typ, err := q.parseColumnNameExpr(v)
			if err != nil {
				return err
			}
			if typ == "DIMENSION" {
				q.q.Dimensions = append(q.q.Dimensions, metricsview.Dimension{Name: col})
			} else {
				q.q.Measures = append(q.q.Measures, metricsview.Measure{Name: col})
			}
		case *ast.FuncCallExpr:
			alias := field.AsName.String()
			res, err := q.parseFuncCallExpr(v)
			if err != nil {
				return err
			}
			if alias == "" {
				alias = restore(v)
			}
			q.q.Dimensions = append(q.q.Dimensions, metricsview.Dimension{
				Name:    alias,
				Compute: res,
			})
		default:
			return fmt.Errorf("metrics sql: unsupported expression in select field")
		}
	}
	return nil
}

func (q *query) parseLimit(node *ast.Limit) error {
	limitVal, err := parseValueExpr(node.Count)
	if err != nil {
		return err
	}
	limit, ok := limitVal.(int)
	if !ok {
		return fmt.Errorf("metrics sql: expected int for limit, got %T", limitVal)
	}
	limit64 := int64(limit)
	q.q.Limit = &limit64

	if node.Offset != nil {
		offsetVal, err := parseValueExpr(node.Offset)
		if err != nil {
			return err
		}
		offset, ok := offsetVal.(int)
		if !ok {
			return fmt.Errorf("metrics sql: expected int for offset, got %T", offsetVal)
		}
		offset64 := int64(offset)
		q.q.Offset = &offset64
	}
	return nil
}

func (q *query) parseOrderBy(node *ast.OrderByClause) error {
	for _, item := range node.Items {
		col, err := parseColumnNameExpr(item.Expr)
		if err != nil {
			return err
		}
		q.q.Sort = append(q.q.Sort, metricsview.Sort{
			Name: col,
			Desc: item.Desc,
		})
	}
	return nil
}

func (q *query) parseHaving(ctx context.Context, node *ast.HavingClause) error {
	expr, err := parseFilter(ctx, node.Expr, nil, q)
	if err != nil {
		return err
	}
	q.q.Having = expr
	return nil
}

func (q *query) parseColumnNameExpr(in ast.Node) (string, string, error) {
	node, ok := in.(*ast.ColumnNameExpr)
	if !ok {
		return "", "", fmt.Errorf("metrics sql: expected column name expression")
	}
	if node.Name == nil {
		return "", "", fmt.Errorf("metrics sql: can only have dimension/measure name(s) in select list")
	}
	if node.Name.Schema.String() != "" || node.Name.Table.String() != "" {
		return "", "", fmt.Errorf("metrics sql: no alias or table reference is supported in column name. Found in `%s`", node.Name.String())
	}

	col := node.Name.Name.O
	if _, ok := q.dims[col]; ok {
		return col, "DIMENSION", nil
	} else if _, ok := q.measures[col]; ok {
		return col, "MEASURE", nil
	} else if q.metricsViewSpec.TimeDimension == col {
		return col, "DIMENSION", nil
	}
	return "", "", fmt.Errorf("metrics sql: selected column `%s` not found in dimensions/measures in metrics view", col)
}

func (q *query) parseFuncCallExpr(node *ast.FuncCallExpr) (*metricsview.DimensionCompute, error) {
	fncName := node.FnName
	if fncName.L != "date_trunc" {
		return nil, fmt.Errorf("metrics sql: function `%s` not supported in select field", fncName.L)
	}

	// example date_trunc(MONTH, COL)
	if len(node.Args) != 2 {
		return nil, fmt.Errorf("metrics sql: expected 2 arguments in date_trunc function")
	}
	grainVal, err := parseValueExpr(node.Args[0]) // handling of MONTH
	if err != nil {
		return nil, err
	}
	grain, ok := grainVal.(string)
	if !ok {
		return nil, fmt.Errorf("metrics sql: expected string for grain, got %T", grainVal)
	}

	col, typ, err := q.parseColumnNameExpr(node.Args[1]) // handling of col
	if err != nil {
		return nil, err
	}
	if typ != "DIMENSION" {
		return nil, fmt.Errorf("metrics sql: expected dimension in date_trunc function")
	}

	return &metricsview.DimensionCompute{
		TimeFloor: &metricsview.DimensionComputeTimeFloor{
			Dimension: col,
			Grain:     metricsview.TimeGrain(strings.ToLower(grain)),
		},
	}, nil
}

func restore(node ast.Node) string {
	var sb strings.Builder
	rctx := format.NewRestoreCtx(format.RestoreStringSingleQuotes|format.RestoreKeyWordUppercase|format.RestoreNameDoubleQuotes|format.RestoreStringWithoutCharset, &sb)
	_ = node.Restore(rctx)
	return sb.String()
}

func add(t time.Time, unit string, amount int) (time.Time, error) {
	switch strings.ToLower(unit) {
	case "second":
		return t.Add(time.Duration(amount) * time.Second), nil
	case "minute":
		return t.Add(time.Duration(amount) * time.Minute), nil
	case "hour":
		return t.Add(time.Duration(amount) * time.Hour), nil
	case "day":
		return t.AddDate(0, 0, amount), nil
	case "week":
		return t.AddDate(0, 0, 7*amount), nil
	case "month":
		return t.AddDate(0, amount, 0), nil
	case "year":
		return t.AddDate(amount, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid time unit %q", unit)
	}
}

func sub(t time.Time, unit string, amount int) (time.Time, error) {
	switch strings.ToLower(unit) {
	case "second":
		return t.Add(-time.Duration(amount) * time.Second), nil
	case "minute":
		return t.Add(-time.Duration(amount) * time.Minute), nil
	case "hour":
		return t.Add(-time.Duration(amount) * time.Hour), nil
	case "day":
		return t.AddDate(0, 0, -amount), nil
	case "week":
		return t.AddDate(0, 0, -7*amount), nil
	case "month":
		return t.AddDate(0, -amount, 0), nil
	case "year":
		return t.AddDate(-amount, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid time unit %q", unit)
	}
}

func operator(op opcode.Op) metricsview.Operator {
	switch op {
	case opcode.LT:
		return metricsview.OperatorLt
	case opcode.LE:
		return metricsview.OperatorLte
	case opcode.GT:
		return metricsview.OperatorGt
	case opcode.GE:
		return metricsview.OperatorGte
	case opcode.EQ:
		return metricsview.OperatorEq
	case opcode.NE:
		return metricsview.OperatorNeq
	case opcode.In:
		return metricsview.OperatorIn
	case opcode.Like:
		return metricsview.OperatorIlike
	case opcode.Or:
		return metricsview.OperatorOr
	case opcode.And:
		return metricsview.OperatorAnd
	default:
		// let the underlying ast parser through errors
		return metricsview.Operator(op.String())
	}
}
