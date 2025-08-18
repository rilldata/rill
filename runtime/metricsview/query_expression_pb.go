package metricsview

import (
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewExpressionFromProto(expr *runtimev1.Expression) *Expression {
	if expr == nil {
		return nil
	}

	res := &Expression{}

	switch e := expr.Expression.(type) {
	case *runtimev1.Expression_Ident:
		res.Name = e.Ident
	case *runtimev1.Expression_Val:
		res.Value = e.Val.AsInterface()
	case *runtimev1.Expression_Cond:
		var op Operator
		switch e.Cond.Op {
		case runtimev1.Operation_OPERATION_UNSPECIFIED:
			op = OperatorUnspecified
		case runtimev1.Operation_OPERATION_EQ:
			op = OperatorEq
		case runtimev1.Operation_OPERATION_NEQ:
			op = OperatorNeq
		case runtimev1.Operation_OPERATION_LT:
			op = OperatorLt
		case runtimev1.Operation_OPERATION_LTE:
			op = OperatorLte
		case runtimev1.Operation_OPERATION_GT:
			op = OperatorGt
		case runtimev1.Operation_OPERATION_GTE:
			op = OperatorGte
		case runtimev1.Operation_OPERATION_OR:
			op = OperatorOr
		case runtimev1.Operation_OPERATION_AND:
			op = OperatorAnd
		case runtimev1.Operation_OPERATION_IN:
			op = OperatorIn
		case runtimev1.Operation_OPERATION_NIN:
			op = OperatorNin
		case runtimev1.Operation_OPERATION_LIKE:
			op = OperatorIlike
		case runtimev1.Operation_OPERATION_NLIKE:
			op = OperatorNilike
		}

		exprs := make([]*Expression, 0, len(e.Cond.Exprs))
		for _, e := range e.Cond.Exprs {
			exprs = append(exprs, NewExpressionFromProto(e))
		}

		res.Condition = &Condition{
			Operator:    op,
			Expressions: exprs,
		}
	case *runtimev1.Expression_Subquery:
		measures := make([]Measure, 0, len(e.Subquery.Measures))
		for _, m := range e.Subquery.Measures {
			measures = append(measures, Measure{Name: m})
		}

		res.Subquery = &Subquery{
			Dimension: Dimension{Name: e.Subquery.Dimension},
			Measures:  measures,
			Where:     NewExpressionFromProto(e.Subquery.Where),
			Having:    NewExpressionFromProto(e.Subquery.Having),
		}
	}

	return res
}

func ExpressionToProto(expr *Expression) *runtimev1.Expression {
	if expr == nil {
		return nil
	}

	res := &runtimev1.Expression{}
	if expr.Name != "" {
		res.Expression = &runtimev1.Expression_Ident{Ident: expr.Name}
	} else if expr.Value != nil {
		val, err := structpb.NewValue(expr.Value)
		if err != nil {
			// If we can't convert the value, return nil
			return nil
		}

		res.Expression = &runtimev1.Expression_Val{Val: val}
	} else if expr.Condition != nil {
		var op runtimev1.Operation
		switch expr.Condition.Operator {
		case OperatorUnspecified:
			op = runtimev1.Operation_OPERATION_UNSPECIFIED
		case OperatorEq:
			op = runtimev1.Operation_OPERATION_EQ
		case OperatorNeq:
			op = runtimev1.Operation_OPERATION_NEQ
		case OperatorLt:
			op = runtimev1.Operation_OPERATION_LT
		case OperatorLte:
			op = runtimev1.Operation_OPERATION_LTE
		case OperatorGt:
			op = runtimev1.Operation_OPERATION_GT
		case OperatorGte:
			op = runtimev1.Operation_OPERATION_GTE
		case OperatorOr:
			op = runtimev1.Operation_OPERATION_OR
		case OperatorAnd:
			op = runtimev1.Operation_OPERATION_AND
		case OperatorIn:
			op = runtimev1.Operation_OPERATION_IN
		case OperatorNin:
			op = runtimev1.Operation_OPERATION_NIN
		case OperatorIlike:
			op = runtimev1.Operation_OPERATION_LIKE
		case OperatorNilike:
			op = runtimev1.Operation_OPERATION_NLIKE
		default:
			panic(fmt.Sprintf("unknown operator %q", expr.Condition.Operator))
		}
		exprs := make([]*runtimev1.Expression, 0, len(expr.Condition.Expressions))
		for _, e := range expr.Condition.Expressions {
			protoExpr := ExpressionToProto(e)
			if protoExpr != nil {
				exprs = append(exprs, protoExpr)
			}
		}

		res.Expression = &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    op,
				Exprs: exprs,
			},
		}
	} else if expr.Subquery != nil {
		measures := make([]string, 0, len(expr.Subquery.Measures))
		for _, m := range expr.Subquery.Measures {
			measures = append(measures, m.Name)
		}

		res.Expression = &runtimev1.Expression_Subquery{
			Subquery: &runtimev1.Subquery{
				Dimension: expr.Subquery.Dimension.Name,
				Measures:  measures,
				Where:     ExpressionToProto(expr.Subquery.Where),
				Having:    ExpressionToProto(expr.Subquery.Having),
			},
		}
	}

	return res
}
